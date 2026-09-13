"""
COG SDK Data Models
"""

from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional, List, Dict, Any
from enum import Enum


class LifecycleStage(str, Enum):
    UNKNOWN = "UNKNOWN"
    DISCOVERED = "DISCOVERED"
    ANNOUNCED = "ANNOUNCED"
    REFERRED = "REFERRED"
    EARLY_DEVELOPMENT = "EARLY_DEVELOPMENT"
    FEASIBILITY = "FEASIBILITY"
    FINANCING = "FINANCING"
    ENVIRONMENTAL_REVIEW = "ENVIRONMENTAL_REVIEW"
    PERMITTING = "PERMITTING"
    PROCUREMENT = "PROCUREMENT"
    FID_LIKELY = "FID_LIKELY"
    FID = "FID"
    CONSTRUCTION = "CONSTRUCTION"
    COMMISSIONING = "COMMISSIONING"
    OPERATING = "OPERATING"
    DELAYED = "DELAYED"
    PAUSED = "PAUSED"
    CANCELLED = "CANCELLED"


class ConfidenceLevel(str, Enum):
    VERIFIED = "VERIFIED"
    SUPPORTED = "SUPPORTED"
    REPORTED = "REPORTED"
    INFERRED = "INFERRED"
    CONFLICTED = "CONFLICTED"
    UNKNOWN = "UNKNOWN"
    STALE = "STALE"
    RETRACTED = "RETRACTED"


class Sector(str, Enum):
    CRITICAL_MINERALS = "Critical Minerals"
    NUCLEAR_ENERGY = "Nuclear & Clean Power"
    CLEAN_ENERGY = "Clean Energy & Grid"
    AI_COMPUTE = "AI Compute & Data Centres"
    DEFENCE_ARCTIC = "Defence & Arctic"
    TRANSPORTATION = "Transportation & Ports"
    INDUSTRIAL_MFG = "Industrial & Manufacturing"
    HOUSING_ENABLING = "Housing-Enabling Infrastructure"
    MINING_METALS = "Mining & Metals"
    ENERGY_FUELS = "Energy & Fuels"
    FORESTRY_BIOECONOMY = "Forestry & Bioeconomy"


class CapitalCategory(str, Enum):
    EQUITY = "equity"
    DEBT = "debt"
    GRANT = "grant"
    TAX_INCENTIVE = "tax_incentive"
    LOAN_GUARANTEE = "loan_guarantee"
    CIB = "cib"
    CGF = "cgf"
    INDIGENOUS_LOAN_GUARANTEE = "indigenous_loan_guarantee"
    PENSION = "pension"
    STRATEGIC_CORPORATE = "strategic_corporate"


class IntelligenceStatus(str, Enum):
    HEALTHY = "HEALTHY"
    PARTIAL = "PARTIAL"
    DEGRADED = "DEGRADED"
    STALE = "STALE"
    UNAVAILABLE = "UNAVAILABLE"


@dataclass
class Evidence:
    id: str
    source_url: str
    publisher: str
    source_tier: int
    retrieval_timestamp: datetime
    publication_date: Optional[datetime] = None
    effective_date: Optional[datetime] = None
    confidence: ConfidenceLevel = ConfidenceLevel.UNKNOWN
    extraction_method: str = ""
    content_hash: str = ""
    hash_scope: Optional[str] = None
    source_class: Optional[str] = None
    source_record_id: Optional[str] = None
    locator: Optional[str] = None
    pipeline_version: Optional[str] = None
    parser_version: Optional[str] = None
    raw_snippet: Optional[str] = None


@dataclass
class Entity:
    id: str
    slug: str
    legal_name: str
    common_name: str
    aliases: List[str] = field(default_factory=list)
    entity_type: str = ""
    jurisdiction: str = ""
    website: Optional[str] = None
    identifiers: Dict[str, str] = field(default_factory=dict)
    description: Optional[str] = None
    ai_sovereignty: Optional[Dict[str, Any]] = None
    evidence_id: Optional[str] = None
    evidence: Optional[Evidence] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Project:
    id: str
    slug: str
    name: str
    summary: str
    sector: Sector
    subsector: str
    province: str
    location_name: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    current_stage: LifecycleStage = LifecycleStage.UNKNOWN
    capex_cad: int = 0
    capex_status: ConfidenceLevel = ConfidenceLevel.UNKNOWN
    proponent_id: Optional[str] = None
    proponent: Optional[Entity] = None
    confidence: ConfidenceLevel = ConfidenceLevel.UNKNOWN
    evidence_ids: List[str] = field(default_factory=list)
    external_ids: Dict[str, str] = field(default_factory=dict)
    is_synthetic: bool = False
    last_meaningful_update: datetime = field(default_factory=datetime.utcnow)
    scores: Dict[str, float] = field(default_factory=dict)
    score_details: List["ProjectScore"] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Event:
    id: str
    project_id: str
    event_type: str
    event_date: datetime
    previous_stage: Optional[LifecycleStage] = None
    new_stage: Optional[LifecycleStage] = None
    title: str = ""
    description: str = ""
    evidence_id: Optional[str] = None
    evidence: Optional[Evidence] = None
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Relationship:
    id: str
    project_id: str
    source_entity_id: str
    target_entity_id: str
    source_entity: Optional[Entity] = None
    target_entity: Optional[Entity] = None
    relation_type: str = ""
    confidence: ConfidenceLevel = ConfidenceLevel.UNKNOWN
    evidence_id: Optional[str] = None
    evidence: Optional[Evidence] = None
    created_at: datetime = field(default_factory=datetime.utcnow)
    valid_from: Optional[datetime] = None
    valid_to: Optional[datetime] = None


@dataclass
class ProjectScore:
    id: str
    project_id: str
    score_type: str
    score_value: float
    score_version: str
    factors: Dict[str, float] = field(default_factory=dict)
    unknown_factors: List[str] = field(default_factory=list)
    coverage: float = 0.0
    confidence: ConfidenceLevel = ConfidenceLevel.UNKNOWN
    input_hash: str = ""
    previous_value: Optional[float] = None
    movement: Optional[float] = None
    movement_reasons: List[str] = field(default_factory=list)
    explanation: str = ""
    calculated_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class CapitalItem:
    id: str
    project_id: str
    category: CapitalCategory
    status: str
    amount_cad: int
    amount_type: str = "exact"
    provider_entity_id: Optional[str] = None
    provider_name: str = ""
    notes: Optional[str] = None
    evidence_id: Optional[str] = None
    evidence: Optional[Evidence] = None
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Procurement:
    id: str
    tender_id: str
    project_id: Optional[str] = None
    project_name: Optional[str] = None
    title: str = ""
    stage: str = ""
    closing_date: Optional[datetime] = None
    estimated_cad: Optional[int] = None
    buyer: str = ""
    buyer_type: str = ""
    source_url: str = ""
    categories: List[str] = field(default_factory=list)
    requirement_class: str = ""
    evidence_id: Optional[str] = None
    evidence: Optional[Evidence] = None
    metadata: Dict[str, Any] = field(default_factory=dict)
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Opportunity:
    id: str
    project_id: str
    project_name: str
    title: str
    sector: Sector
    requirement_class: str
    category: str
    estimated_cad: Optional[int] = None
    estimate_status: ConfidenceLevel = ConfidenceLevel.UNKNOWN
    description: str = ""
    trigger_milestone: str = ""
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class Signal:
    id: str
    project_id: str
    project_name: str
    type: str
    timestamp: datetime
    magnitude: float
    confidence: float
    previous_state: Optional[str] = None
    new_state: Optional[str] = None
    description: str = ""
    evidence_id: Optional[str] = None


@dataclass
class RadarStats:
    total_projects: int
    total_capex_cad: int
    capital_moving_week_cad: int
    accelerating_projects_count: int
    stalled_projects_count: int
    active_procurements_count: int
    unknown_capex_projects: int
    data_status: IntelligenceStatus
    sector_breakdown: Dict[str, int]
    province_breakdown: Dict[str, int]
    generated_at: datetime


@dataclass
class CapitalProgram:
    id: str
    name: str
    administrator: str
    program_type: str
    sector_eligibility: List[Sector]
    max_support_rate_pct: float
    labor_conditions_req: bool
    mutually_exclusive: List[str]
    stacking_cap_pct: float
    summary: str
    statutory_reference: str
    jurisdiction: str


@dataclass
class ProgramMatch:
    program: CapitalProgram
    classification: str
    estimated_value_cad: int
    labor_requirement: str
    rationale: str


@dataclass
class StackingEvaluation:
    project_id: str
    project_capex_cad: int
    matched_programs: List[ProgramMatch]
    total_potential_cad: int
    stacking_conflicts: List[str]
    effective_funding_pct: float
    disclaimer: str


@dataclass
class ReconciliationMatch:
    federal_record: Optional[Dict[str, Any]] = None
    provincial_record: Optional[Dict[str, Any]] = None
    municipal_record: Optional[Dict[str, Any]] = None
    indigenous_record: Optional[Dict[str, Any]] = None
    action: str = ""
    confidence: str = ""
    rationale: str = ""
    conflicts: List[str] = field(default_factory=list)
    merged_project_id: Optional[str] = None


@dataclass
class IndigenousBusiness:
    business_id: str
    business_name: str
    legal_name: str
    operating_name: Optional[str] = None
    indigenous_group: str = ""
    community_name: str = ""
    province: str = ""
    city: str = ""
    postal_code: str = ""
    naics_code: str = ""
    naics_description: str = ""
    business_type: str = ""
    ownership_percent: int = 100
    certification_date: Optional[datetime] = None
    expiry_date: Optional[datetime] = None
    status: str = "Active"
    contact_email: Optional[str] = None
    contact_phone: Optional[str] = None
    website: Optional[str] = None
    address: Optional[str] = None
    description: Optional[str] = None
    capabilities: List[str] = field(default_factory=list)


@dataclass
class MerkleLeaf:
    index: int
    evidence_id: str
    content_hash: str
    source_url: Optional[str] = None
    timestamp: datetime = field(default_factory=datetime.utcnow)


@dataclass
class MerkleNode:
    hash: str
    left_hash: Optional[str] = None
    right_hash: Optional[str] = None
    leaf_index: Optional[int] = None
    is_leaf: bool = False
    parent_hash: Optional[str] = None


@dataclass
class MerkleTree:
    root_hash: str
    leaf_count: int
    level_count: int
    generated_at: datetime
    leaves: List[MerkleLeaf]
    levels: List[List[MerkleNode]]


@dataclass
class MerkleProof:
    methodology_version: str
    root_hash: str
    leaf: MerkleLeaf
    path: List[MerkleNode]
    verified: bool = False
    verified_at: datetime = field(default_factory=datetime.utcnow)