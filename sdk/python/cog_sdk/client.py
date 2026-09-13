"""
COG SDK HTTP Client

A lightweight client for interacting with the CanadaOpportunityGraph REST API.
"""

import json
import time
from datetime import datetime
from typing import Any, Dict, List, Optional, Union
from urllib import error as url_error
from urllib.parse import urlencode
from urllib.request import Request, urlopen

from .exceptions import (
    COGError,
    COGNotFoundError,
    COGServerError,
    COGValidationError,
    COGRateLimitError,
)
from .models import (
    CapitalCategory,
    CapitalItem,
    CapitalProgram,
    ConfidenceLevel,
    Evidence,
    Event,
    IndigenousBusiness,
    IntelligenceStatus,
    LifecycleStage,
    Opportunity,
    Procurement,
    ProgramMatch,
    Project,
    ProjectScore,
    RadarStats,
    Relationship,
    Sector,
    Signal,
    StackingEvaluation,
)


def _parse_datetime(value: Optional[str]) -> Optional[datetime]:
    if value is None or value == "":
        return None
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00"))
    except (ValueError, TypeError):
        return None


def _parse_date(value: Optional[str]) -> Optional[datetime]:
    return _parse_datetime(value)


def _to_enum(value: Optional[str], enum_cls) -> Optional:
    if value is None:
        return None
    try:
        return enum_cls(value)
    except ValueError:
        return value


def _deserialize_evidence(data: Dict[str, Any]) -> Evidence:
    return Evidence(
        id=data["id"],
        source_url=data.get("source_url", ""),
        publisher=data.get("publisher", ""),
        source_tier=data.get("source_tier", 1),
        retrieval_timestamp=_parse_datetime(data.get("retrieval_timestamp")) or datetime.utcnow(),
        publication_date=_parse_date(data.get("publication_date")),
        effective_date=_parse_date(data.get("effective_date")),
        confidence=_to_enum(data.get("confidence"), ConfidenceLevel) or ConfidenceLevel.UNKNOWN,
        extraction_method=data.get("extraction_method", ""),
        content_hash=data.get("content_hash", ""),
        hash_scope=data.get("hash_scope"),
        source_class=data.get("source_class"),
        source_record_id=data.get("source_record_id"),
        locator=data.get("locator"),
        pipeline_version=data.get("pipeline_version"),
        parser_version=data.get("parser_version"),
        raw_snippet=data.get("raw_snippet"),
    )


def _deserialize_entity(data: Dict[str, Any]) -> Optional["Project"]:
    from .models import Entity
    return Entity(
        id=data.get("id", ""),
        slug=data.get("slug", ""),
        legal_name=data.get("legal_name", ""),
        common_name=data.get("common_name", ""),
        aliases=data.get("aliases", []),
        entity_type=data.get("entity_type", ""),
        jurisdiction=data.get("jurisdiction", ""),
        website=data.get("website"),
        identifiers=data.get("identifiers", {}),
        description=data.get("description"),
        ai_sovereignty=data.get("ai_sovereignty"),
        evidence_id=data.get("evidence_id"),
        evidence=_deserialize_evidence(data["evidence"]) if data.get("evidence") else None,
        metadata=data.get("metadata", {}),
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
        updated_at=_parse_datetime(data.get("updated_at")) or datetime.utcnow(),
    )


def _deserialize_project(data: Dict[str, Any]) -> Project:
    ev = data.get("evidence")
    return Project(
        id=data["id"],
        slug=data["slug"],
        name=data["name"],
        summary=data.get("summary", ""),
        sector=_to_enum(data.get("sector"), Sector) or Sector.INDUSTRIAL_MFG,
        subsector=data.get("subsector", ""),
        province=data.get("province", ""),
        location_name=data.get("location_name", ""),
        latitude=data.get("latitude"),
        longitude=data.get("longitude"),
        current_stage=_to_enum(data.get("current_stage"), LifecycleStage) or LifecycleStage.UNKNOWN,
        capex_cad=data.get("capex_cad", 0),
        capex_status=_to_enum(data.get("capex_status"), ConfidenceLevel) or ConfidenceLevel.UNKNOWN,
        proponent_id=data.get("proponent_id"),
        proponent=_deserialize_entity(data["proponent"]) if data.get("proponent") else None,
        confidence=_to_enum(data.get("confidence"), ConfidenceLevel) or ConfidenceLevel.UNKNOWN,
        evidence_ids=data.get("evidence_ids", []),
        external_ids=data.get("external_ids", {}),
        is_synthetic=data.get("is_synthetic", False),
        last_meaningful_update=_parse_datetime(data.get("last_meaningful_update")) or datetime.utcnow(),
        scores=data.get("scores", {}),
        score_details=[_deserialize_score(s) for s in data.get("score_details", [])],
        metadata=data.get("metadata", {}),
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
        updated_at=_parse_datetime(data.get("updated_at")) or datetime.utcnow(),
    )


def _deserialize_score(data: Dict[str, Any]) -> ProjectScore:
    return ProjectScore(
        id=data.get("id", ""),
        project_id=data.get("project_id", ""),
        score_type=data.get("score_type", ""),
        score_value=data.get("score_value", 0.0),
        score_version=data.get("score_version", ""),
        factors=data.get("factors", {}),
        unknown_factors=data.get("unknown_factors", []),
        coverage=data.get("coverage", 0.0),
        confidence=_to_enum(data.get("confidence"), ConfidenceLevel) or ConfidenceLevel.UNKNOWN,
        input_hash=data.get("input_hash", ""),
        previous_value=data.get("previous_value"),
        movement=data.get("movement"),
        movement_reasons=data.get("movement_reasons", []),
        explanation=data.get("explanation", ""),
        calculated_at=_parse_datetime(data.get("calculated_at")) or datetime.utcnow(),
    )


def _deserialize_event(data: Dict[str, Any]) -> Event:
    return Event(
        id=data.get("id", ""),
        project_id=data.get("project_id", ""),
        event_type=data.get("event_type", ""),
        event_date=_parse_datetime(data.get("event_date")) or datetime.utcnow(),
        previous_stage=_to_enum(data.get("previous_stage"), LifecycleStage),
        new_stage=_to_enum(data.get("new_stage"), LifecycleStage),
        title=data.get("title", ""),
        description=data.get("description", ""),
        evidence_id=data.get("evidence_id"),
        evidence=_deserialize_evidence(data["evidence"]) if data.get("evidence") else None,
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
    )


def _deserialize_relationship(data: Dict[str, Any]) -> Relationship:
    return Relationship(
        id=data.get("id", ""),
        project_id=data.get("project_id", ""),
        source_entity_id=data.get("source_entity_id", ""),
        target_entity_id=data.get("target_entity_id", ""),
        source_entity=_deserialize_entity(data["source_entity"]) if data.get("source_entity") else None,
        target_entity=_deserialize_entity(data["target_entity"]) if data.get("target_entity") else None,
        relation_type=data.get("relation_type", ""),
        confidence=_to_enum(data.get("confidence"), ConfidenceLevel) or ConfidenceLevel.UNKNOWN,
        evidence_id=data.get("evidence_id"),
        evidence=_deserialize_evidence(data["evidence"]) if data.get("evidence") else None,
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
        valid_from=_parse_datetime(data.get("valid_from")),
        valid_to=_parse_datetime(data.get("valid_to")),
    )


def _deserialize_capital_item(data: Dict[str, Any]) -> CapitalItem:
    return CapitalItem(
        id=data.get("id", ""),
        project_id=data.get("project_id", ""),
        category=_to_enum(data.get("category"), CapitalCategory),
        status=data.get("status", ""),
        amount_cad=data.get("amount_cad", 0),
        amount_type=data.get("amount_type", "exact"),
        provider_entity_id=data.get("provider_entity_id"),
        provider_name=data.get("provider_name", ""),
        notes=data.get("notes"),
        evidence_id=data.get("evidence_id"),
        evidence=_deserialize_evidence(data["evidence"]) if data.get("evidence") else None,
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
    )


def _deserialize_opportunity(data: Dict[str, Any]) -> Opportunity:
    return Opportunity(
        id=data.get("id", ""),
        project_id=data.get("project_id", ""),
        project_name=data.get("project_name", ""),
        title=data.get("title", ""),
        sector=_to_enum(data.get("sector"), Sector) or Sector.INDUSTRIAL_MFG,
        requirement_class=data.get("requirement_class", ""),
        category=data.get("category", ""),
        estimated_cad=data.get("estimated_cad"),
        estimate_status=_to_enum(data.get("estimate_status"), ConfidenceLevel) or ConfidenceLevel.UNKNOWN,
        description=data.get("description", ""),
        trigger_milestone=data.get("trigger_milestone", ""),
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
    )


def _deserialize_procurement(data: Dict[str, Any]) -> Procurement:
    return Procurement(
        id=data.get("id", ""),
        tender_id=data.get("tender_id", ""),
        project_id=data.get("project_id"),
        project_name=data.get("project_name"),
        title=data.get("title", ""),
        stage=data.get("stage", ""),
        closing_date=_parse_date(data.get("closing_date")),
        estimated_cad=data.get("estimated_cad"),
        buyer=data.get("buyer", ""),
        buyer_type=data.get("buyer_type", ""),
        source_url=data.get("source_url", ""),
        categories=data.get("categories", []),
        requirement_class=data.get("requirement_class", ""),
        evidence_id=data.get("evidence_id"),
        evidence=_deserialize_evidence(data["evidence"]) if data.get("evidence") else None,
        metadata=data.get("metadata", {}),
        created_at=_parse_datetime(data.get("created_at")) or datetime.utcnow(),
    )


def _deserialize_signal(data: Dict[str, Any]) -> Signal:
    return Signal(
        id=data.get("id", ""),
        project_id=data.get("project_id", ""),
        project_name=data.get("project_name", ""),
        type=data.get("type", ""),
        timestamp=_parse_datetime(data.get("timestamp")) or datetime.utcnow(),
        magnitude=data.get("magnitude", 0.0),
        confidence=data.get("confidence", 0.0),
        previous_state=data.get("previous_state"),
        new_state=data.get("new_state"),
        description=data.get("description", ""),
        evidence_id=data.get("evidence_id"),
    )


def _deserialize_capital_program(data: Dict[str, Any]) -> CapitalProgram:
    return CapitalProgram(
        id=data["id"],
        name=data["name"],
        administrator=data.get("administrator", ""),
        program_type=data.get("program_type", ""),
        sector_eligibility=[_to_enum(s, Sector) for s in data.get("sector_eligibility", [])],
        max_support_rate_pct=data.get("max_support_rate_pct", 0.0),
        labor_conditions_req=data.get("labor_conditions_req", False),
        mutually_exclusive=data.get("mutually_exclusive", []),
        stacking_cap_pct=data.get("stacking_cap_pct", 0.0),
        summary=data.get("summary", ""),
        statutory_reference=data.get("statutory_reference", ""),
        jurisdiction=data.get("jurisdiction", ""),
    )


def _deserialize_program_match(data: Dict[str, Any]) -> ProgramMatch:
    return ProgramMatch(
        program=_deserialize_capital_program(data["program"]),
        classification=data.get("classification", ""),
        estimated_value_cad=data.get("estimated_value_cad", 0),
        labor_requirement=data.get("labor_requirement", ""),
        rationale=data.get("rationale", ""),
    )


def _deserialize_stacking_evaluation(data: Dict[str, Any]) -> StackingEvaluation:
    return StackingEvaluation(
        project_id=data.get("project_id", ""),
        project_capex_cad=data.get("project_capex_cad", 0),
        matched_programs=[_deserialize_program_match(m) for m in data.get("matched_programs", [])],
        total_potential_cad=data.get("total_potential_cad", 0),
        stacking_conflicts=data.get("stacking_conflicts", []),
        effective_funding_pct=data.get("effective_funding_pct", 0.0),
        disclaimer=data.get("disclaimer", ""),
    )


def _deserialize_radar_stats(data: Dict[str, Any]) -> RadarStats:
    return RadarStats(
        total_projects=data.get("total_projects", 0),
        total_capex_cad=data.get("total_capex_cad", 0),
        capital_moving_week_cad=data.get("capital_moving_week_cad", 0),
        accelerating_projects_count=data.get("accelerating_projects_count", 0),
        stalled_projects_count=data.get("stalled_projects_count", 0),
        active_procurements_count=data.get("active_procurements_count", 0),
        unknown_capex_projects=data.get("unknown_capex_projects", 0),
        data_status=_to_enum(data.get("data_status"), IntelligenceStatus),
        sector_breakdown=data.get("sector_breakdown", {}),
        province_breakdown=data.get("province_breakdown", {}),
        generated_at=_parse_datetime(data.get("generated_at")) or datetime.utcnow(),
    )


class COGClient:
    """
    Client for the CanadaOpportunityGraph REST API.

    Usage:
        client = COGClient("https://api.canadaopportunitygraph.ca")
        projects = client.get_projects(limit=50)
    """

    def __init__(
        self,
        base_url: str,
        api_key: Optional[str] = None,
        timeout: int = 30,
    ):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.timeout = timeout

    def _headers(self) -> Dict[str, str]:
        headers = {"Accept": "application/json"}
        if self.api_key:
            headers["X-API-Key"] = self.api_key
        return headers

    def _request(
        self,
        path: str,
        params: Optional[Dict[str, Any]] = None,
    ) -> Any:
        url = self.base_url + path
        if params:
            query = urlencode({k: v for k, v in params.items() if v is not None})
            if query:
                url += "?" + query

        req = Request(url, headers=self._headers(), method="GET")
        try:
            with urlopen(req, timeout=self.timeout) as resp:
                body = resp.read().decode("utf-8")
                if not body:
                    return None
                return json.loads(body)
        except url_error.HTTPError as e:
            body = e.read().decode("utf-8", errors="replace")
            try:
                err = json.loads(body)
            except (ValueError, TypeError):
                err = {"message": body}
            if e.code == 404:
                raise COGNotFoundError(err.get("message", "Resource not found"))
            elif e.code == 400:
                raise COGValidationError(err.get("message", "Validation error"))
            elif e.code == 429:
                retry = e.headers.get("Retry-After") if e.headers else None
                raise COGRateLimitError(
                    err.get("message", "Rate limit exceeded"),
                    retry_after=int(retry) if retry else None,
                )
            elif e.code >= 500:
                raise COGServerError(err.get("message", "Server error"), status_code=e.code)
            else:
                raise COGError(err.get("message", "Request failed"), status_code=e.code)
        except url_error.URLError as e:
            raise COGError(f"Connection error: {e.reason}")

    def health(self) -> Dict[str, Any]:
        """Check API health."""
        return self._request("/health")

    def ready(self) -> Dict[str, Any]:
        """Check if the API data store is ready."""
        return self._request("/ready")

    def get_projects(
        self,
        limit: int = 50,
        offset: int = 0,
        sector: Optional[str] = None,
        province: Optional[str] = None,
        stage: Optional[str] = None,
        min_capex: Optional[int] = None,
        q: Optional[str] = None,
        sort_by: Optional[str] = None,
        sort_dir: Optional[str] = None,
    ) -> Dict[str, Any]:
        """
        List projects with optional filters.

        Returns a dict with 'projects' (List[Project]), 'total', 'limit', 'offset'.
        """
        resp = self._request("/api/v1/projects", params={
            "limit": limit,
            "offset": offset,
            "sector": sector,
            "province": province,
            "stage": stage,
            "min_capex": min_capex,
            "q": q,
            "sort_by": sort_by,
            "sort_dir": sort_dir,
        })
        if resp is None:
            return {"projects": [], "total": 0, "limit": limit, "offset": offset}
        resp["projects"] = [_deserialize_project(p) for p in resp.get("projects", [])]
        return resp

    def get_project(self, project_id: str) -> Dict[str, Any]:
        """
        Get a single project with scores, events, relationships, capital items,
        and opportunities.
        """
        resp = self._request(f"/api/v1/projects/{project_id}")
        if resp is None:
            return {}
        if "project" in resp:
            resp["project"] = _deserialize_project(resp["project"])
        if "scores" in resp:
            resp["scores"] = [_deserialize_score(s) for s in resp["scores"]]
        if "events" in resp:
            resp["events"] = [_deserialize_event(e) for e in resp["events"]]
        if "relationships" in resp:
            resp["relationships"] = [_deserialize_relationship(r) for r in resp["relationships"]]
        if "capital_items" in resp:
            resp["capital_items"] = [_deserialize_capital_item(c) for c in resp["capital_items"]]
        if "opportunities" in resp:
            resp["opportunities"] = [_deserialize_opportunity(o) for o in resp["opportunities"]]
        return resp

    def get_project_events(self, project_id: str) -> List[Event]:
        """List events for a project."""
        resp = self._request(f"/api/v1/projects/{project_id}/events")
        if resp is None:
            return []
        return [_deserialize_event(e) for e in resp]

    def get_project_scores(self, project_id: str) -> List[ProjectScore]:
        """Get latest scores for a project."""
        resp = self._request(f"/api/v1/projects/{project_id}/scores")
        if resp is None:
            return []
        return [_deserialize_score(s) for s in resp]

    def get_project_score_history(
        self, project_id: str, score_type: str = ""
    ) -> Dict[str, Any]:
        """Get score history for a project."""
        params = {"type": score_type} if score_type else None
        resp = self._request(f"/api/v1/projects/{project_id}/scores/history", params=params)
        if resp is None:
            return {"project_id": project_id, "history": []}
        resp["history"] = [_deserialize_score(s) for s in resp.get("history", [])]
        return resp

    def get_project_provenance(self, project_id: str) -> Dict[str, Any]:
        """Get provenance information for a project."""
        return self._request(f"/api/v1/projects/{project_id}/provenance")

    def get_project_trust(self, project_id: str) -> Dict[str, Any]:
        """Get trust evaluation for a project."""
        return self._request(f"/api/v1/projects/{project_id}/trust")

    def get_procurements(self, limit: int = 50, offset: int = 0) -> Dict[str, Any]:
        """
        List procurements.

        Returns a dict with 'procurements' (List[Procurement]), 'limit', 'offset'.
        """
        resp = self._request("/api/v1/procurements", params={"limit": limit, "offset": offset})
        if resp is None:
            return {"procurements": [], "limit": limit, "offset": offset}
        resp["procurements"] = [_deserialize_procurement(p) for p in resp.get("procurements", [])]
        return resp

    def get_signals(self) -> List[Signal]:
        """List recent signals."""
        resp = self._request("/api/v1/signals")
        if resp is None:
            return []
        return [_deserialize_signal(s) for s in resp]

    def search(self, query: str, limit: int = 20) -> Dict[str, Any]:
        """
        Search projects by keyword.

        Returns a dict with 'query', 'total', 'projects'.
        """
        resp = self._request("/api/v1/search", params={"q": query, "limit": limit})
        if resp is None:
            return {"query": query, "total": 0, "projects": []}
        resp["projects"] = [_deserialize_project(p) for p in resp.get("projects", [])]
        return resp

    def get_capital_stack(self) -> Dict[str, Any]:
        """
        Get the capital stack analysis: canonical programs and disclaimer.
        """
        return self._request("/api/v1/capital/stack")

    def get_radar(self) -> Dict[str, Any]:
        """Get radar dashboard data: stats, accelerating projects, recent signals/events."""
        return self._request("/api/v1/radar")

    def get_rankings(self, dimension: str, limit: int = 50) -> List[Any]:
        """Get project rankings by a scoring dimension (buildability, investability, etc.)."""
        resp = self._request(f"/api/v1/rankings/{dimension}", params={"limit": limit})
        if resp is None:
            return []
        return resp

    def export_project(self, project_id: str) -> Dict[str, Any]:
        """Export a project bundle in CEGS format."""
        return self._request(f"/api/v1/export/project/{project_id}")

    def get_cegs_project(self, project_id: str) -> Dict[str, Any]:
        """Get a project in CEGS (Canada Economic Graph Schema) format."""
        return self._request(f"/api/v1/cegs/projects/{project_id}")

    def export_cegs(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Export the full CEGS dataset with optional filters."""
        return self._request("/api/v1/cegs/export", params=params or {})
