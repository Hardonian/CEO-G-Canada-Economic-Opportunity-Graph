package forecast

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

var (
	ErrInvalidInput    = errors.New("forecast: invalid input")
	ErrTemporalLeakage = errors.New("forecast: project snapshot is newer than as-of time")
)

var scenarioIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type normalizedRequest struct {
	asOf      time.Time
	horizons  []int
	scenarios []Scenario
}

type normalizedContext struct {
	project       *domain.Project
	events        []*domain.Event
	capital       []*domain.CapitalItem
	relationships []*domain.Relationship
	procurements  []*domain.Procurement
	opportunities []*domain.Opportunity
	signals       []*domain.Signal
	evidence      map[string]*domain.Evidence
	referencedIDs []string
	warnings      []string
}

type observations struct {
	committedCAD       int64
	conditionalCAD     int64
	announcedCAD       int64
	financedRatio      float64
	financingEvidence  []string
	regulatory         bool
	regulatoryEvidence []string
	indigenous         bool
	indigenousEvidence []string
	offtake            bool
	offtakeEvidence    []string
	activeProcurement  int
	momentum           float64
	momentumEvidence   []string
	delaySignal        bool
}

// Evaluate produces an immutable, deterministic forecast. It performs no
// network calls and uses Request.AsOf rather than wall-clock time.
func Evaluate(ctx Context, request Request) (*Report, error) {
	req, err := normalizeRequest(request)
	if err != nil {
		return nil, err
	}
	nctx, err := normalizeContext(ctx, req.asOf)
	if err != nil {
		return nil, err
	}

	quality := evaluateDataQuality(nctx, req.asOf)
	provenance := buildProvenance(nctx)
	inputHash, err := calculateInputHash(nctx, req)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot hash canonical input: %v", ErrInvalidInput, err)
	}
	obs, err := collectObservations(nctx, req.asOf)
	if err != nil {
		return nil, err
	}

	scenarioForecasts := make([]ScenarioForecast, 0, len(req.scenarios))
	for _, scenario := range req.scenarios {
		result, buildErr := evaluateScenario(nctx, req, quality, obs, scenario)
		if buildErr != nil {
			return nil, buildErr
		}
		scenarioForecasts = append(scenarioForecasts, result)
	}

	status := statusForQuality(quality)
	report := &Report{
		ID:                 "forecast_" + inputHash[:24],
		ProjectID:          nctx.project.ID,
		ProjectName:        nctx.project.Name,
		MethodologyVersion: MethodologyVersion,
		AsOf:               req.asOf,
		CalculatedAt:       req.asOf,
		Status:             status,
		Confidence:         quality.Confidence,
		InputHash:          inputHash,
		DataQuality:        quality,
		Scenarios:          scenarioForecasts,
		Provenance:         provenance,
		Warnings:           append([]string(nil), nctx.warnings...),
		Limitations: []string{
			"Completion percentages are deterministic planning likelihood indices, not statistically calibrated probabilities, forecasts from a trained ML model, or investment advice.",
			"Milestone dates are conditional schedules derived from stage defaults and scenario inputs; project-specific schedules should replace defaults when available.",
			"Funding gaps compare scenario-adjusted capex with evidenced exact committed, closed, or disbursed capital; they do not model eligible costs, cash-flow timing, returns, or financing terms.",
			"Employment, GDP, tax revenue, emissions, capacity, and security outcomes are not estimated because those causal inputs are absent from the current domain model.",
			"Scenario values are caller-controlled stress assumptions and must not be represented as sourced facts.",
		},
		Assurance: SovereignAssurance{
			ProcessingMode:          "LOCAL_DETERMINISTIC_RULES",
			ExternalNetworkRequired: false,
			ExternalTelemetry:       false,
			RawEvidenceEmitted:      false,
			Deterministic:           true,
			Statement:               "These are package execution properties, not a certification of the host, deployment, data classification, or legal compliance.",
		},
	}
	return report, nil
}

func normalizeRequest(request Request) (normalizedRequest, error) {
	if request.AsOf.IsZero() {
		return normalizedRequest{}, fmt.Errorf("%w: as_of is required", ErrInvalidInput)
	}
	asOf := request.AsOf.UTC()
	maxHorizon := request.MaxHorizonMonths
	if maxHorizon == 0 {
		maxHorizon = DefaultMaxHorizonMonths
	}
	if maxHorizon < 1 || maxHorizon > DefaultMaxHorizonMonths {
		return normalizedRequest{}, fmt.Errorf("%w: max_horizon_months must be between 1 and %d", ErrInvalidInput, DefaultMaxHorizonMonths)
	}

	horizons := append([]int(nil), request.HorizonsMonths...)
	if len(horizons) == 0 {
		horizons = []int{12, 36, 60}
	}
	if len(horizons) > 24 {
		return normalizedRequest{}, fmt.Errorf("%w: at most 24 horizons are allowed", ErrInvalidInput)
	}
	seenHorizon := make(map[int]struct{}, len(horizons))
	uniqueHorizons := make([]int, 0, len(horizons))
	for _, horizon := range horizons {
		if horizon < 1 || horizon > maxHorizon {
			return normalizedRequest{}, fmt.Errorf("%w: horizon %d must be between 1 and %d months", ErrInvalidInput, horizon, maxHorizon)
		}
		if _, exists := seenHorizon[horizon]; !exists {
			seenHorizon[horizon] = struct{}{}
			uniqueHorizons = append(uniqueHorizons, horizon)
		}
	}
	sort.Ints(uniqueHorizons)

	scenarios := append([]Scenario(nil), request.Scenarios...)
	if len(scenarios) == 0 {
		scenarios = []Scenario{BaselineScenario()}
	}
	if len(scenarios) > DefaultMaxScenarios {
		return normalizedRequest{}, fmt.Errorf("%w: at most %d scenarios are allowed", ErrInvalidInput, DefaultMaxScenarios)
	}
	seenScenario := make(map[string]struct{}, len(scenarios))
	for i := range scenarios {
		if err := validateScenario(scenarios[i]); err != nil {
			return normalizedRequest{}, err
		}
		if _, exists := seenScenario[scenarios[i].ID]; exists {
			return normalizedRequest{}, fmt.Errorf("%w: duplicate scenario id %q", ErrInvalidInput, scenarios[i].ID)
		}
		seenScenario[scenarios[i].ID] = struct{}{}
	}
	sort.Slice(scenarios, func(i, j int) bool { return scenarios[i].ID < scenarios[j].ID })
	return normalizedRequest{asOf: asOf, horizons: uniqueHorizons, scenarios: scenarios}, nil
}

func validateScenario(s Scenario) error {
	if !scenarioIDPattern.MatchString(s.ID) {
		return fmt.Errorf("%w: scenario id %q must be 1-64 safe identifier characters", ErrInvalidInput, s.ID)
	}
	if strings.TrimSpace(s.Name) == "" || len(s.Name) > 200 {
		return fmt.Errorf("%w: scenario %q name must be 1-200 characters", ErrInvalidInput, s.ID)
	}
	if len(s.Description) > 2_000 {
		return fmt.Errorf("%w: scenario %q description exceeds 2000 characters", ErrInvalidInput, s.ID)
	}
	values := []struct {
		name    string
		value   float64
		minimum float64
		maximum float64
	}{
		{"financing_availability_delta", s.FinancingAvailabilityDelta, -1, 1},
		{"capex_escalation_pct", s.CapexEscalationPct, -50, 200},
		{"demand_delta", s.DemandDelta, -1, 1},
		{"policy_support_delta", s.PolicySupportDelta, -1, 1},
		{"supply_chain_stress", s.SupplyChainStress, 0, 1},
	}
	for _, value := range values {
		if math.IsNaN(value.value) || math.IsInf(value.value, 0) || value.value < value.minimum || value.value > value.maximum {
			return fmt.Errorf("%w: scenario %q %s must be between %g and %g", ErrInvalidInput, s.ID, value.name, value.minimum, value.maximum)
		}
	}
	if s.RegulatoryDelayMonths < -24 || s.RegulatoryDelayMonths > 120 {
		return fmt.Errorf("%w: scenario %q regulatory_delay_months must be between -24 and 120", ErrInvalidInput, s.ID)
	}
	if s.ConstructionDelayMonths < -24 || s.ConstructionDelayMonths > 120 {
		return fmt.Errorf("%w: scenario %q construction_delay_months must be between -24 and 120", ErrInvalidInput, s.ID)
	}
	return nil
}

func normalizeContext(ctx Context, asOf time.Time) (normalizedContext, error) {
	if ctx.Project == nil {
		return normalizedContext{}, fmt.Errorf("%w: project is required", ErrInvalidInput)
	}
	if strings.TrimSpace(ctx.Project.ID) == "" {
		return normalizedContext{}, fmt.Errorf("%w: project id is required", ErrInvalidInput)
	}
	if len(ctx.Project.ID) > 256 || len(ctx.Project.Name) > 1_000 {
		return normalizedContext{}, fmt.Errorf("%w: project identity fields exceed safe length limits", ErrInvalidInput)
	}
	if !validLifecycleStage(ctx.Project.CurrentStage) {
		return normalizedContext{}, fmt.Errorf("%w: unsupported lifecycle stage %q", ErrInvalidInput, ctx.Project.CurrentStage)
	}
	if ctx.Project.CapexCAD < 0 {
		return normalizedContext{}, fmt.Errorf("%w: project capex cannot be negative", ErrInvalidInput)
	}
	if (!ctx.Project.UpdatedAt.IsZero() && ctx.Project.UpdatedAt.After(asOf)) ||
		(!ctx.Project.LastMeaningfulUpdate.IsZero() && ctx.Project.LastMeaningfulUpdate.After(asOf)) {
		return normalizedContext{}, fmt.Errorf("%w: project %q", ErrTemporalLeakage, ctx.Project.ID)
	}
	for name, size := range map[string]int{
		"events": len(ctx.Events), "capital_items": len(ctx.CapitalItems), "relationships": len(ctx.Relationships),
		"procurements": len(ctx.Procurements), "opportunities": len(ctx.Opportunities), "signals": len(ctx.Signals), "evidence": len(ctx.Evidence),
	} {
		if size > DefaultMaxRecords {
			return normalizedContext{}, fmt.Errorf("%w: %s exceeds %d records", ErrInvalidInput, name, DefaultMaxRecords)
		}
	}

	n := normalizedContext{project: ctx.Project, evidence: make(map[string]*domain.Evidence)}
	warnings := make(map[string]struct{})
	warn := func(value string) { warnings[value] = struct{}{} }
	seenIDs := make(map[string]map[string]struct{})
	checkID := func(kind, id string) error {
		if seenIDs[kind] == nil {
			seenIDs[kind] = make(map[string]struct{})
		}
		if _, exists := seenIDs[kind][id]; exists {
			return fmt.Errorf("%w: duplicate %s id %q", ErrInvalidInput, kind, id)
		}
		seenIDs[kind][id] = struct{}{}
		return nil
	}

	allEvidence := append([]*domain.Evidence(nil), ctx.Evidence...)
	for i, event := range ctx.Events {
		if event == nil {
			return normalizedContext{}, fmt.Errorf("%w: events[%d] is nil", ErrInvalidInput, i)
		}
		if err := validateOwnedRecord("event", event.ID, event.ProjectID, ctx.Project.ID); err != nil {
			return normalizedContext{}, err
		}
		if err := checkID("event", event.ID); err != nil {
			return normalizedContext{}, err
		}
		if !event.CreatedAt.IsZero() && event.CreatedAt.After(asOf) {
			warn("future_event_ignored")
			if event.Evidence != nil {
				allEvidence = append(allEvidence, event.Evidence)
			}
			continue
		}
		if event.EventDate.IsZero() {
			warn("event_missing_event_date")
		} else if event.EventDate.After(asOf) {
			warn("future_event_ignored")
			if event.Evidence != nil {
				allEvidence = append(allEvidence, event.Evidence)
			}
			continue
		}
		if event.CreatedAt.IsZero() {
			warn("event_missing_observation_time")
		}
		if err := validateEvidenceLink(event.EvidenceID, event.Evidence); err != nil {
			return normalizedContext{}, err
		}
		if event.Evidence != nil {
			allEvidence = append(allEvidence, event.Evidence)
		}
		n.events = append(n.events, event)
	}
	for i, item := range ctx.CapitalItems {
		if item == nil {
			return normalizedContext{}, fmt.Errorf("%w: capital_items[%d] is nil", ErrInvalidInput, i)
		}
		if err := validateOwnedRecord("capital item", item.ID, item.ProjectID, ctx.Project.ID); err != nil {
			return normalizedContext{}, err
		}
		if err := checkID("capital item", item.ID); err != nil {
			return normalizedContext{}, err
		}
		if item.AmountCAD < 0 {
			return normalizedContext{}, fmt.Errorf("%w: capital item %q has negative amount", ErrInvalidInput, item.ID)
		}
		if !item.CreatedAt.IsZero() && item.CreatedAt.After(asOf) {
			warn("future_capital_item_ignored")
			if item.Evidence != nil {
				allEvidence = append(allEvidence, item.Evidence)
			}
			continue
		}
		if item.CreatedAt.IsZero() {
			warn("capital_item_missing_observation_time")
		}
		if err := validateEvidenceLink(item.EvidenceID, item.Evidence); err != nil {
			return normalizedContext{}, err
		}
		if item.Evidence != nil {
			allEvidence = append(allEvidence, item.Evidence)
		}
		n.capital = append(n.capital, item)
	}
	for i, relationship := range ctx.Relationships {
		if relationship == nil {
			return normalizedContext{}, fmt.Errorf("%w: relationships[%d] is nil", ErrInvalidInput, i)
		}
		if err := validateOwnedRecord("relationship", relationship.ID, relationship.ProjectID, ctx.Project.ID); err != nil {
			return normalizedContext{}, err
		}
		if err := checkID("relationship", relationship.ID); err != nil {
			return normalizedContext{}, err
		}
		if !relationship.CreatedAt.IsZero() && relationship.CreatedAt.After(asOf) {
			warn("future_relationship_ignored")
			if relationship.Evidence != nil {
				allEvidence = append(allEvidence, relationship.Evidence)
			}
			continue
		}
		if relationship.CreatedAt.IsZero() {
			warn("relationship_missing_observation_time")
		}
		if relationship.ValidFrom != nil && relationship.ValidFrom.After(asOf) {
			warn("future_relationship_ignored")
			continue
		}
		if relationship.ValidTo != nil && relationship.ValidTo.Before(asOf) {
			continue
		}
		if err := validateEvidenceLink(relationship.EvidenceID, relationship.Evidence); err != nil {
			return normalizedContext{}, err
		}
		if relationship.Evidence != nil {
			allEvidence = append(allEvidence, relationship.Evidence)
		}
		n.relationships = append(n.relationships, relationship)
	}
	for i, procurement := range ctx.Procurements {
		if procurement == nil {
			return normalizedContext{}, fmt.Errorf("%w: procurements[%d] is nil", ErrInvalidInput, i)
		}
		if err := validateOwnedRecord("procurement", procurement.ID, procurement.ProjectID, ctx.Project.ID); err != nil {
			return normalizedContext{}, err
		}
		if err := checkID("procurement", procurement.ID); err != nil {
			return normalizedContext{}, err
		}
		if procurement.EstimatedCAD < 0 {
			return normalizedContext{}, fmt.Errorf("%w: procurement %q has negative amount", ErrInvalidInput, procurement.ID)
		}
		if !procurement.CreatedAt.IsZero() && procurement.CreatedAt.After(asOf) {
			warn("future_procurement_ignored")
			if procurement.Evidence != nil {
				allEvidence = append(allEvidence, procurement.Evidence)
			}
			continue
		}
		if procurement.CreatedAt.IsZero() {
			warn("procurement_missing_observation_time")
		}
		if err := validateEvidenceLink(procurement.EvidenceID, procurement.Evidence); err != nil {
			return normalizedContext{}, err
		}
		if procurement.Evidence != nil {
			allEvidence = append(allEvidence, procurement.Evidence)
		}
		n.procurements = append(n.procurements, procurement)
	}
	for i, opportunity := range ctx.Opportunities {
		if opportunity == nil {
			return normalizedContext{}, fmt.Errorf("%w: opportunities[%d] is nil", ErrInvalidInput, i)
		}
		if err := validateOwnedRecord("opportunity", opportunity.ID, opportunity.ProjectID, ctx.Project.ID); err != nil {
			return normalizedContext{}, err
		}
		if err := checkID("opportunity", opportunity.ID); err != nil {
			return normalizedContext{}, err
		}
		if opportunity.EstimatedCAD < 0 {
			return normalizedContext{}, fmt.Errorf("%w: opportunity %q has negative amount", ErrInvalidInput, opportunity.ID)
		}
		if !opportunity.CreatedAt.IsZero() && opportunity.CreatedAt.After(asOf) {
			warn("future_opportunity_ignored")
			continue
		}
		if opportunity.CreatedAt.IsZero() {
			warn("opportunity_missing_observation_time")
		}
		n.opportunities = append(n.opportunities, opportunity)
	}
	for i, signal := range ctx.Signals {
		if signal == nil {
			return normalizedContext{}, fmt.Errorf("%w: signals[%d] is nil", ErrInvalidInput, i)
		}
		if err := validateOwnedRecord("signal", signal.ID, signal.ProjectID, ctx.Project.ID); err != nil {
			return normalizedContext{}, err
		}
		if err := checkID("signal", signal.ID); err != nil {
			return normalizedContext{}, err
		}
		if math.IsNaN(signal.Magnitude) || math.IsInf(signal.Magnitude, 0) || signal.Magnitude < 0 || signal.Magnitude > 1 ||
			math.IsNaN(signal.Confidence) || math.IsInf(signal.Confidence, 0) || signal.Confidence < 0 || signal.Confidence > 1 {
			return normalizedContext{}, fmt.Errorf("%w: signal %q magnitude and confidence must be within 0..1", ErrInvalidInput, signal.ID)
		}
		if !signal.Timestamp.IsZero() && signal.Timestamp.After(asOf) {
			warn("future_signal_ignored")
			continue
		}
		if signal.Timestamp.IsZero() {
			warn("signal_missing_observation_time")
		}
		n.signals = append(n.signals, signal)
	}

	seenEvidenceDigests := make(map[string]string)
	for i, evidence := range allEvidence {
		if evidence == nil {
			return normalizedContext{}, fmt.Errorf("%w: evidence[%d] is nil", ErrInvalidInput, i)
		}
		if strings.TrimSpace(evidence.ID) == "" {
			return normalizedContext{}, fmt.Errorf("%w: evidence id is required", ErrInvalidInput)
		}
		if err := validateEvidenceBoundary(evidence); err != nil {
			return normalizedContext{}, err
		}
		digest := evidenceDigest(evidence)
		if existingDigest, exists := seenEvidenceDigests[evidence.ID]; exists {
			if existingDigest != digest {
				return normalizedContext{}, fmt.Errorf("%w: conflicting duplicate evidence id %q", ErrInvalidInput, evidence.ID)
			}
			continue
		}
		seenEvidenceDigests[evidence.ID] = digest
		if !evidence.RetrievalTimestamp.IsZero() && evidence.RetrievalTimestamp.After(asOf) {
			warn("future_retrieved_evidence_ignored")
			continue
		}
		if evidence.RetrievalTimestamp.IsZero() {
			warn("evidence_missing_retrieval_time")
		}
		n.evidence[evidence.ID] = evidence
	}

	references := make(map[string]struct{})
	for _, id := range ctx.Project.EvidenceIDs {
		if id != "" {
			references[id] = struct{}{}
		}
	}
	for _, event := range n.events {
		addReference(references, event.EvidenceID, event.Evidence)
	}
	for _, item := range n.capital {
		addReference(references, item.EvidenceID, item.Evidence)
	}
	for _, relationship := range n.relationships {
		addReference(references, relationship.EvidenceID, relationship.Evidence)
	}
	for _, procurement := range n.procurements {
		addReference(references, procurement.EvidenceID, procurement.Evidence)
	}
	for _, signal := range n.signals {
		if signal.EvidenceID != "" {
			references[signal.EvidenceID] = struct{}{}
		}
	}
	for id := range references {
		n.referencedIDs = append(n.referencedIDs, id)
	}
	sort.Strings(n.referencedIDs)

	sort.Slice(n.events, func(i, j int) bool { return n.events[i].ID < n.events[j].ID })
	sort.Slice(n.capital, func(i, j int) bool { return n.capital[i].ID < n.capital[j].ID })
	sort.Slice(n.relationships, func(i, j int) bool { return n.relationships[i].ID < n.relationships[j].ID })
	sort.Slice(n.procurements, func(i, j int) bool { return n.procurements[i].ID < n.procurements[j].ID })
	sort.Slice(n.opportunities, func(i, j int) bool { return n.opportunities[i].ID < n.opportunities[j].ID })
	sort.Slice(n.signals, func(i, j int) bool { return n.signals[i].ID < n.signals[j].ID })
	for warning := range warnings {
		n.warnings = append(n.warnings, warning)
	}
	sort.Strings(n.warnings)
	return n, nil
}

func validateOwnedRecord(kind, id, projectID, expectedProjectID string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%w: %s id is required", ErrInvalidInput, kind)
	}
	if len(id) > 256 || len(projectID) > 256 {
		return fmt.Errorf("%w: %s identity fields exceed safe length limits", ErrInvalidInput, kind)
	}
	if projectID != expectedProjectID {
		return fmt.Errorf("%w: %s %q belongs to project %q, expected %q", ErrInvalidInput, kind, id, projectID, expectedProjectID)
	}
	return nil
}

func validateEvidenceBoundary(evidence *domain.Evidence) error {
	if len(evidence.ID) > 256 || len(evidence.Publisher) > 512 || len(evidence.SourceURL) > 4_096 || len(evidence.ContentHash) > 512 {
		return fmt.Errorf("%w: evidence %q fields exceed safe length limits", ErrInvalidInput, evidence.ID)
	}
	if evidence.SourceURL == "" {
		return nil
	}
	parsed, err := url.Parse(evidence.SourceURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
		return fmt.Errorf("%w: evidence %q source_url must be an http(s) URL without credentials", ErrInvalidInput, evidence.ID)
	}
	return nil
}

func validateEvidenceLink(id string, embedded *domain.Evidence) error {
	if embedded != nil && id != "" && embedded.ID != id {
		return fmt.Errorf("%w: evidence link %q conflicts with embedded evidence %q", ErrInvalidInput, id, embedded.ID)
	}
	return nil
}

func addReference(target map[string]struct{}, id string, embedded *domain.Evidence) {
	if id != "" {
		target[id] = struct{}{}
		return
	}
	if embedded != nil && embedded.ID != "" {
		target[embedded.ID] = struct{}{}
	}
}

func evidenceDigest(evidence *domain.Evidence) string {
	value := struct {
		ID, URL, Publisher, Confidence, ContentHash, SourceRecordID, Locator string
		Tier                                                                 domain.SourceTier
		Retrieval                                                            time.Time
		Publication, Effective                                               *time.Time
	}{evidence.ID, evidence.SourceURL, evidence.Publisher, string(evidence.Confidence), evidence.ContentHash, evidence.SourceRecordID, evidence.Locator,
		evidence.SourceTier, evidence.RetrievalTimestamp.UTC(), utcPtr(evidence.PublicationDate), utcPtr(evidence.EffectiveDate)}
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func evaluateDataQuality(ctx normalizedContext, asOf time.Time) DataQuality {
	coverage := 0.0
	gaps := make([]string, 0)
	if ctx.project.CurrentStage != "" && ctx.project.CurrentStage != domain.StageUnknown {
		coverage += 15
	} else {
		gaps = append(gaps, "current_stage")
	}
	if ctx.project.CapexCAD > 0 {
		coverage += 15
	} else {
		gaps = append(gaps, "capex")
	}
	if isKnownConfidence(ctx.project.CapexStatus) {
		coverage += 5
	} else {
		gaps = append(gaps, "capex_confidence")
	}
	if !ctx.project.LastMeaningfulUpdate.IsZero() || !ctx.project.UpdatedAt.IsZero() {
		coverage += 10
	} else {
		gaps = append(gaps, "project_update_time")
	}
	if ctx.project.ProponentID != "" {
		coverage += 10
	} else {
		gaps = append(gaps, "proponent")
	}
	if ctx.project.LocationName != "" || (ctx.project.Latitude != 0 || ctx.project.Longitude != 0) {
		coverage += 5
	} else {
		gaps = append(gaps, "location")
	}
	if ctx.project.Sector != "" {
		coverage += 5
	} else {
		gaps = append(gaps, "sector")
	}
	if len(ctx.events) > 0 {
		coverage += 10
	} else {
		gaps = append(gaps, "historical_event_series")
	}
	if len(ctx.capital) > 0 {
		coverage += 10
	} else {
		gaps = append(gaps, "financing_evidence")
	}

	resolved := make([]*domain.Evidence, 0, len(ctx.referencedIDs))
	missing := 0
	for _, id := range ctx.referencedIDs {
		if item, ok := ctx.evidence[id]; ok {
			resolved = append(resolved, item)
		} else {
			missing++
		}
	}
	if len(ctx.referencedIDs) > 0 {
		coverage += 15 * float64(len(resolved)) / float64(len(ctx.referencedIDs))
	} else {
		gaps = append(gaps, "referenced_evidence")
	}
	if missing > 0 {
		gaps = append(gaps, "unresolved_evidence_references")
	}

	evidenceQuality, freshness := 0.0, 0.0
	conflicted := 0
	newestAge := (*int)(nil)
	for _, evidence := range resolved {
		evidenceQuality += (tierScore(evidence.SourceTier) + confidenceScore(evidence.Confidence)) / 2
		if evidence.Confidence == domain.ConfidenceConflict {
			conflicted++
		}
		date := evidence.RetrievalTimestamp
		if evidence.PublicationDate != nil && !evidence.PublicationDate.After(asOf) {
			date = *evidence.PublicationDate
		}
		if evidence.EffectiveDate != nil && !evidence.EffectiveDate.After(asOf) {
			date = *evidence.EffectiveDate
		}
		ageDays := 10_000
		if !date.IsZero() {
			ageDays = int(asOf.Sub(date.UTC()).Hours() / 24)
			if ageDays < 0 {
				ageDays = 0
			}
		}
		if newestAge == nil || ageDays < *newestAge {
			value := ageDays
			newestAge = &value
		}
		freshness += freshnessScore(ageDays)
	}
	if len(resolved) > 0 {
		evidenceQuality /= float64(len(resolved))
		freshness /= float64(len(resolved))
	}
	if conflicted > 0 {
		gaps = append(gaps, "conflicted_evidence")
	}
	coverage = round1(coverage)
	evidenceQuality = round1(evidenceQuality)
	freshness = round1(freshness)
	score := round1(coverage*0.55 + evidenceQuality*0.30 + freshness*0.15)
	if conflicted > 0 {
		score = round1(math.Max(0, score-math.Min(30, float64(conflicted)*15)))
	}
	confidence := confidenceForScore(score)
	sort.Strings(gaps)
	return DataQuality{
		Score: score, Confidence: confidence, CoveragePct: coverage, EvidenceQualityPct: evidenceQuality,
		EvidenceFreshnessPct: freshness, NewestEvidenceAgeDays: newestAge, ReferencedEvidence: len(resolved),
		MissingEvidenceRefs: missing, ConflictedEvidence: conflicted, CriticalGaps: gaps,
		RecordCounts: RecordCounts{Events: len(ctx.events), CapitalItems: len(ctx.capital), Relationships: len(ctx.relationships),
			Procurements: len(ctx.procurements), Opportunities: len(ctx.opportunities), Signals: len(ctx.signals)},
	}
}

func buildProvenance(ctx normalizedContext) []EvidenceReference {
	result := make([]EvidenceReference, 0, len(ctx.referencedIDs))
	for _, id := range ctx.referencedIDs {
		evidence, ok := ctx.evidence[id]
		if !ok {
			continue
		}
		result = append(result, EvidenceReference{EvidenceID: evidence.ID, Publisher: evidence.Publisher, SourceURL: evidence.SourceURL,
			SourceTier: evidence.SourceTier, Confidence: evidence.Confidence, RetrievalTime: evidence.RetrievalTimestamp.UTC(),
			EffectiveDate: utcPtr(evidence.EffectiveDate), ContentHash: evidence.ContentHash})
	}
	return result
}

func collectObservations(ctx normalizedContext, asOf time.Time) (observations, error) {
	var result observations
	for _, item := range ctx.capital {
		if !strings.EqualFold(item.AmountType, "exact") {
			continue
		}
		var destination *int64
		switch item.Status {
		case domain.CapitalCommitted, domain.CapitalClosed, domain.CapitalDisbursed:
			destination = &result.committedCAD
		case domain.CapitalConditionallyCommitted:
			destination = &result.conditionalCAD
		case domain.CapitalAnnounced, domain.CapitalProposed:
			destination = &result.announcedCAD
		}
		if destination != nil {
			value, err := safeAdd(*destination, item.AmountCAD)
			if err != nil {
				return observations{}, fmt.Errorf("%w: capital totals overflow", ErrInvalidInput)
			}
			*destination = value
			if item.Status == domain.CapitalCommitted || item.Status == domain.CapitalClosed || item.Status == domain.CapitalDisbursed {
				result.financingEvidence = appendIf(result.financingEvidence, item.EvidenceID)
			}
		}
	}
	if ctx.project.CapexCAD > 0 {
		result.financedRatio = clamp(float64(result.committedCAD)/float64(ctx.project.CapexCAD), 0, 1.5)
	}

	for _, event := range ctx.events {
		switch event.EventType {
		case "regulatory.impact_assessment_filing", "regulatory_filing", "terms_of_reference", "regulatory.impact_assessment_decision_issued", "environmental_approval", "ministerial_decision", "regulatory.hold_point_removed":
			result.regulatory = true
			result.regulatoryEvidence = appendIf(result.regulatoryEvidence, event.EvidenceID)
		case "indigenous_agreement", "impact_benefit_agreement":
			result.indigenous = true
			result.indigenousEvidence = appendIf(result.indigenousEvidence, event.EvidenceID)
		case "offtake_agreement", "power_purchase_agreement":
			result.offtake = true
			result.offtakeEvidence = appendIf(result.offtakeEvidence, event.EvidenceID)
		case "project_delay", "timeline_slip":
			result.delaySignal = true
		}
		if event.NewStage != nil && (*event.NewStage == domain.StageDelayed || *event.NewStage == domain.StagePaused) {
			result.delaySignal = true
		}
	}
	for _, relationship := range ctx.relationships {
		switch relationship.RelationType {
		case "indigenous_partner", "participates_in", "partners_with":
			result.indigenous = true
			result.indigenousEvidence = appendIf(result.indigenousEvidence, relationship.EvidenceID)
		case "offtakes_from", "offtaker":
			result.offtake = true
			result.offtakeEvidence = appendIf(result.offtakeEvidence, relationship.EvidenceID)
		}
	}
	for _, procurement := range ctx.procurements {
		stage := strings.ToLower(strings.TrimSpace(procurement.Stage))
		if stage == "cancelled" || stage == "closed" || stage == "expired" {
			continue
		}
		if procurement.ClosingDate == nil || !procurement.ClosingDate.Before(asOf) || stage == "awarded" {
			result.activeProcurement++
		}
	}
	for _, signal := range ctx.signals {
		ageDays := 365.0
		if !signal.Timestamp.IsZero() {
			ageDays = math.Max(0, asOf.Sub(signal.Timestamp.UTC()).Hours()/24)
		}
		decay := 1 / (1 + ageDays/90)
		direction := 1.0
		switch signal.Type {
		case domain.SignalTimelineSlip, domain.SignalProjectDelay, domain.SignalPoliticalSupportLoss, domain.SignalCapexIncrease:
			direction = -1
		}
		result.momentum += direction * signal.Magnitude * signal.Confidence * decay
		result.momentumEvidence = appendIf(result.momentumEvidence, signal.EvidenceID)
	}
	result.momentum = clamp(result.momentum, -1, 1)
	if result.delaySignal {
		result.momentum = clamp(result.momentum-0.25, -1, 1)
	}
	result.financingEvidence = sortedUnique(result.financingEvidence)
	result.regulatoryEvidence = sortedUnique(result.regulatoryEvidence)
	result.indigenousEvidence = sortedUnique(result.indigenousEvidence)
	result.offtakeEvidence = sortedUnique(result.offtakeEvidence)
	result.momentumEvidence = sortedUnique(result.momentumEvidence)
	return result, nil
}

func evaluateScenario(ctx normalizedContext, req normalizedRequest, quality DataQuality, obs observations, scenario Scenario) (ScenarioForecast, error) {
	capital, err := capitalOutlook(ctx, scenario, obs)
	if err != nil {
		return ScenarioForecast{}, err
	}
	ceiling, drivers := viabilityCeiling(ctx, scenario, obs)
	uncertainty := scheduleUncertainty(quality.Confidence, ctx.project.CurrentStage)
	ceilingSpread := map[ConfidenceRating]float64{ConfidenceHigh: 5, ConfidenceModerate: 10, ConfidenceLow: 18, ConfidenceInsufficient: 25}[quality.Confidence]

	assumptions := scenarioAssumptions(ctx, scenario, obs, ceiling, uncertainty)
	targets := []domain.LifecycleStage{domain.StageFID, domain.StageConstruction, domain.StageOperating}
	milestones := make([]MilestoneForecast, 0, len(targets))
	for _, target := range targets {
		milestones = append(milestones, forecastMilestone(ctx.project.CurrentStage, target, req, scenario, obs, ceiling, ceilingSpread, uncertainty))
	}
	watch := buildWatchItems(ctx, quality, capital, scenario)
	return ScenarioForecast{Scenario: scenario, Assumptions: assumptions, Capital: capital, Milestones: milestones, Drivers: drivers, WatchItems: watch}, nil
}

func viabilityCeiling(ctx normalizedContext, scenario Scenario, obs observations) (float64, []Driver) {
	stageValue := stageCeiling(ctx.project.CurrentStage)
	drivers := []Driver{{Code: "stage_maturity", Label: "Lifecycle maturity", ImpactPoints: round1(stageValue - 50), Explanation: "Methodology baseline relative to a 50-point announced-project reference."}}
	ceiling := stageValue
	if ctx.project.CapexCAD > 0 {
		impact := clamp(12*(obs.financedRatio-0.25), -3, 12)
		ceiling += impact
		drivers = append(drivers, Driver{Code: "evidenced_financing", Label: "Evidenced financing", ImpactPoints: round1(impact), Explanation: "Exact committed, closed, or disbursed capital relative to reported capex.", EvidenceIDs: obs.financingEvidence})
	}
	if obs.regulatory {
		ceiling += 4
		drivers = append(drivers, Driver{Code: "regulatory_evidence", Label: "Regulatory progression", ImpactPoints: 4, Explanation: "A recognized regulatory filing, decision, or hold-point event is present.", EvidenceIDs: obs.regulatoryEvidence})
	}
	if obs.indigenous {
		ceiling += 4
		drivers = append(drivers, Driver{Code: "indigenous_partnership", Label: "Indigenous partnership", ImpactPoints: 4, Explanation: "A typed partnership relationship or agreement event is present; legal sufficiency is not inferred.", EvidenceIDs: obs.indigenousEvidence})
	}
	if obs.offtake {
		ceiling += 4
		drivers = append(drivers, Driver{Code: "offtake", Label: "Commercial offtake", ImpactPoints: 4, Explanation: "A typed offtake relationship or agreement event is present.", EvidenceIDs: obs.offtakeEvidence})
	}
	if obs.activeProcurement > 0 {
		impact := math.Min(4, float64(obs.activeProcurement))
		ceiling += impact
		drivers = append(drivers, Driver{Code: "active_procurement", Label: "Active procurement", ImpactPoints: impact, Explanation: fmt.Sprintf("%d active or awarded procurement records are present.", obs.activeProcurement)})
	}
	if obs.momentum != 0 {
		impact := obs.momentum * 8
		ceiling += impact
		drivers = append(drivers, Driver{Code: "recent_momentum", Label: "Time-decayed signal momentum", ImpactPoints: round1(impact), Explanation: "Signals are weighted by supplied confidence and deterministic 90-day hyperbolic decay.", EvidenceIDs: obs.momentumEvidence})
	}
	scenarioDrivers := []Driver{
		{Code: "scenario_financing", Label: "Scenario financing", ImpactPoints: round1(scenario.FinancingAvailabilityDelta * 10), Explanation: "Caller-supplied financing availability stress."},
		{Code: "scenario_policy", Label: "Scenario policy support", ImpactPoints: round1(scenario.PolicySupportDelta * 8), Explanation: "Caller-supplied policy support stress."},
		{Code: "scenario_demand", Label: "Scenario demand", ImpactPoints: round1(scenario.DemandDelta * 5), Explanation: "Caller-supplied demand stress."},
		{Code: "scenario_supply_chain", Label: "Scenario supply chain", ImpactPoints: round1(-scenario.SupplyChainStress * 10), Explanation: "Caller-supplied supply-chain stress."},
	}
	for _, driver := range scenarioDrivers {
		if driver.ImpactPoints != 0 {
			ceiling += driver.ImpactPoints
			drivers = append(drivers, driver)
		}
	}
	if ctx.project.CurrentStage == domain.StageCancelled {
		ceiling = 0
	}
	ceiling = round1(clamp(ceiling, 0, 100))
	sort.SliceStable(drivers, func(i, j int) bool {
		left, right := math.Abs(drivers[i].ImpactPoints), math.Abs(drivers[j].ImpactPoints)
		if left == right {
			return drivers[i].Code < drivers[j].Code
		}
		return left > right
	})
	return ceiling, drivers
}

func forecastMilestone(current, target domain.LifecycleStage, req normalizedRequest, scenario Scenario, obs observations, ceiling, ceilingSpread, uncertainty float64) MilestoneForecast {
	result := MilestoneForecast{Stage: target, Basis: "Conditional stage-duration rule adjusted by the explicit scenario; likelihood is capped by an explainable continuation index."}
	if current == domain.StageCancelled {
		result.Basis = "The project snapshot is CANCELLED; no continuation schedule is asserted."
		for _, horizon := range req.horizons {
			result.ByHorizon = append(result.ByHorizon, HorizonLikelihood{HorizonMonths: horizon, HorizonDate: addCalendarMonths(req.asOf, horizon), CompletionLikelihoodIndexPct: LikelihoodRange{}})
		}
		return result
	}
	if hasReached(current, target) {
		date := req.asOf
		result.ConditionalSchedule = DateRange{Earliest: &date, Base: &date, Latest: &date}
		result.Basis = "The current lifecycle snapshot has already reached this milestone; the actual historical milestone date is not inferred."
		for _, horizon := range req.horizons {
			result.ByHorizon = append(result.ByHorizon, HorizonLikelihood{HorizonMonths: horizon, HorizonDate: addCalendarMonths(req.asOf, horizon), CompletionLikelihoodIndexPct: LikelihoodRange{Low: 100, Base: 100, High: 100}})
		}
		return result
	}

	baseMonths := defaultMonthsTo(current, target)
	additional := 0
	if !hasReached(current, domain.StageFID) {
		additional += scenario.RegulatoryDelayMonths
	}
	if target == domain.StageOperating && !hasReached(current, domain.StageOperating) {
		additional += scenario.ConstructionDelayMonths
	}
	baseMonths = math.Max(1, baseMonths+float64(additional))
	factor := scheduleFactor(current, scenario, obs.momentum)
	baseMonths = math.Max(1, math.Round(baseMonths*factor))
	earliestMonths := math.Max(1, math.Floor(baseMonths*(1-uncertainty)))
	latestMonths := math.Max(baseMonths, math.Ceil(baseMonths*(1+uncertainty)))
	earliestDate, baseDate, latestDate := addCalendarMonths(req.asOf, int(earliestMonths)), addCalendarMonths(req.asOf, int(baseMonths)), addCalendarMonths(req.asOf, int(latestMonths))
	result.ConditionalSchedule = DateRange{Earliest: &earliestDate, Base: &baseDate, Latest: &latestDate}
	lowCeiling, highCeiling := clamp(ceiling-ceilingSpread, 0, 100), clamp(ceiling+ceilingSpread, 0, 100)
	ramp := math.Max(3, baseMonths*0.25)
	for _, horizon := range req.horizons {
		low := completionCurve(float64(horizon), latestMonths, ramp) * lowCeiling / 100
		base := completionCurve(float64(horizon), baseMonths, ramp) * ceiling / 100
		high := completionCurve(float64(horizon), earliestMonths, ramp) * highCeiling / 100
		result.ByHorizon = append(result.ByHorizon, HorizonLikelihood{HorizonMonths: horizon, HorizonDate: addCalendarMonths(req.asOf, horizon),
			CompletionLikelihoodIndexPct: orderedLikelihood(low, base, high)})
	}
	return result
}

func scenarioAssumptions(ctx normalizedContext, scenario Scenario, obs observations, ceiling, uncertainty float64) []Assumption {
	items := []Assumption{
		textAssumption("current_stage", string(ctx.project.CurrentStage), AssumptionObserved, "Current project snapshot; no hidden stage is inferred.", nil),
		numericAssumption("financed_share", round1(obs.financedRatio*100), "percent_of_reported_capex", AssumptionObserved, "Only exact committed, closed, or disbursed capital is counted.", obs.financingEvidence),
		numericAssumption("momentum", round1(obs.momentum), "index_-1_to_1", AssumptionObserved, "Time-decayed supplied signals; no text sentiment inference is performed.", obs.momentumEvidence),
		numericAssumption("stage_default_months_to_fid", defaultMonthsTo(ctx.project.CurrentStage, domain.StageFID), "months", AssumptionMethodology, "Generic stage-duration prior used only until a project-specific schedule is supplied.", nil),
		numericAssumption("stage_default_months_to_construction", defaultMonthsTo(ctx.project.CurrentStage, domain.StageConstruction), "months", AssumptionMethodology, "Generic stage-duration prior used only until a project-specific schedule is supplied.", nil),
		numericAssumption("stage_default_months_to_operation", defaultMonthsTo(ctx.project.CurrentStage, domain.StageOperating), "months", AssumptionMethodology, "Generic stage-duration prior used only until a project-specific schedule is supplied.", nil),
		numericAssumption("schedule_multiplier", round1(scheduleFactor(ctx.project.CurrentStage, scenario, obs.momentum)), "multiplier", AssumptionMethodology, "Deterministic combination of momentum and the listed scenario stresses, bounded to 0.5..2.5.", nil),
		numericAssumption("schedule_uncertainty", round1(uncertainty*100), "percent_each_side", AssumptionMethodology, "Interval width expands when evidence support or lifecycle specificity is low.", nil),
		numericAssumption("capex_range_uncertainty", round1(capexUncertainty(ctx.project.CapexStatus)*100), "percent_each_side", AssumptionMethodology, "Range width is selected from the reported capex confidence classification.", nil),
		numericAssumption("continuation_reference", 50, "planning_index_points", AssumptionMethodology, "Reference from which the lifecycle maturity driver is expressed.", nil),
		numericAssumption("continuation_ceiling", ceiling, "planning_index_percent", AssumptionMethodology, "Stage baseline plus listed observed and scenario drivers; not an empirical probability.", nil),
		numericAssumption("financing_availability_delta", scenario.FinancingAvailabilityDelta, "fraction", AssumptionScenario, "Caller-supplied scenario stress.", nil),
		numericAssumption("regulatory_delay", float64(scenario.RegulatoryDelayMonths), "months", AssumptionScenario, "Caller-supplied schedule adjustment before FID.", nil),
		numericAssumption("construction_delay", float64(scenario.ConstructionDelayMonths), "months", AssumptionScenario, "Caller-supplied schedule adjustment for operation.", nil),
		numericAssumption("capex_escalation", scenario.CapexEscalationPct, "percent", AssumptionScenario, "Caller-supplied nominal capex stress; no inflation series is inferred.", nil),
		numericAssumption("demand_delta", scenario.DemandDelta, "fraction", AssumptionScenario, "Caller-supplied scenario stress.", nil),
		numericAssumption("policy_support_delta", scenario.PolicySupportDelta, "fraction", AssumptionScenario, "Caller-supplied scenario stress.", nil),
		numericAssumption("supply_chain_stress", scenario.SupplyChainStress, "index_0_to_1", AssumptionScenario, "Caller-supplied scenario stress.", nil),
	}
	return items
}

func capitalOutlook(ctx normalizedContext, scenario Scenario, obs observations) (CapitalOutlook, error) {
	uncertainty := capexUncertainty(ctx.project.CapexStatus)
	base, err := safeScale(ctx.project.CapexCAD, 1+scenario.CapexEscalationPct/100)
	if err != nil {
		return CapitalOutlook{}, fmt.Errorf("%w: scenario %q adjusted capex overflows CAD range", ErrInvalidInput, scenario.ID)
	}
	low, err := safeScale(base, 1-uncertainty)
	if err != nil {
		return CapitalOutlook{}, err
	}
	high, err := safeScale(base, 1+uncertainty)
	if err != nil {
		return CapitalOutlook{}, err
	}
	confirmedProcurement, confirmedOpportunity, derivedOpportunity := int64(0), int64(0), int64(0)
	for _, item := range ctx.procurements {
		if item.RequirementClass == domain.RequirementConfirmed && item.EstimatedCAD > 0 {
			confirmedProcurement, err = safeAdd(confirmedProcurement, item.EstimatedCAD)
			if err != nil {
				return CapitalOutlook{}, fmt.Errorf("%w: procurement totals overflow", ErrInvalidInput)
			}
		}
	}
	for _, item := range ctx.opportunities {
		if item.EstimatedCAD <= 0 {
			continue
		}
		switch item.RequirementClass {
		case domain.RequirementConfirmed:
			confirmedOpportunity, err = safeAdd(confirmedOpportunity, item.EstimatedCAD)
		case domain.RequirementDerived:
			derivedOpportunity, err = safeAdd(derivedOpportunity, item.EstimatedCAD)
		}
		if err != nil {
			return CapitalOutlook{}, fmt.Errorf("%w: opportunity totals overflow", ErrInvalidInput)
		}
	}
	return CapitalOutlook{
		ReportedCapexCAD: ctx.project.CapexCAD, ScenarioAdjustedCapexCAD: MoneyRange{Low: low, Base: base, High: high},
		EvidencedCommittedCAD: obs.committedCAD, ConditionallyCommittedCAD: obs.conditionalCAD, AnnouncedOrProposedCAD: obs.announcedCAD,
		IndicativeFundingGapCAD: MoneyRange{Low: positiveDifference(low, obs.committedCAD), Base: positiveDifference(base, obs.committedCAD), High: positiveDifference(high, obs.committedCAD)},
		ConfirmedProcurementCAD: confirmedProcurement, ConfirmedOpportunityCAD: confirmedOpportunity, DerivedOpportunityCAD: derivedOpportunity,
		CapexUncertaintyPct: round1(uncertainty * 100),
	}, nil
}

func buildWatchItems(ctx normalizedContext, quality DataQuality, capital CapitalOutlook, scenario Scenario) []WatchItem {
	var result []WatchItem
	if ctx.project.CapexCAD == 0 {
		result = append(result, WatchItem{Code: "CAPEX_UNKNOWN", Severity: "HIGH", Message: "A funding requirement cannot be quantified until reported capex is available."})
	}
	if capital.IndicativeFundingGapCAD.Base > 0 {
		result = append(result, WatchItem{Code: "INDICATIVE_FUNDING_GAP", Severity: "MEDIUM", Message: "Scenario-adjusted capex exceeds evidenced exact committed, closed, and disbursed capital."})
	}
	if capital.EvidencedCommittedCAD > capital.ScenarioAdjustedCapexCAD.High && capital.ScenarioAdjustedCapexCAD.High > 0 {
		result = append(result, WatchItem{Code: "COMMITMENTS_EXCEED_CAPEX_RANGE", Severity: "HIGH", Message: "Evidenced commitments exceed the scenario capex range; scope, units, duplication, and capital classifications require review."})
	}
	if quality.Confidence == ConfidenceLow || quality.Confidence == ConfidenceInsufficient {
		result = append(result, WatchItem{Code: "LOW_FORECAST_SUPPORT", Severity: "HIGH", Message: "Use scenario outputs as a data-collection aid until the listed critical gaps are resolved."})
	}
	if quality.MissingEvidenceRefs > 0 {
		result = append(result, WatchItem{Code: "MISSING_EVIDENCE", Severity: "HIGH", Message: "One or more referenced evidence records were not supplied or were not knowable at the as-of time."})
	}
	if quality.ConflictedEvidence > 0 {
		result = append(result, WatchItem{Code: "CONFLICTED_EVIDENCE", Severity: "HIGH", Message: "Conflicted evidence is present and materially reduces decision confidence."})
	}
	switch ctx.project.CurrentStage {
	case domain.StageDelayed, domain.StagePaused:
		result = append(result, WatchItem{Code: "PROJECT_NOT_ACTIVE", Severity: "HIGH", Message: "The current delayed or paused state materially widens schedule uncertainty."})
	case domain.StageCancelled:
		result = append(result, WatchItem{Code: "PROJECT_CANCELLED", Severity: "INFO", Message: "No future milestone schedule is asserted for a cancelled project."})
	}
	if scenario.RegulatoryDelayMonths != 0 || scenario.ConstructionDelayMonths != 0 || scenario.SupplyChainStress != 0 {
		result = append(result, WatchItem{Code: "SCENARIO_SCHEDULE_STRESS", Severity: "INFO", Message: "Conditional dates include caller-authored schedule stress assumptions."})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Code < result[j].Code })
	return result
}

func calculateInputHash(ctx normalizedContext, req normalizedRequest) (string, error) {
	type projectInput struct {
		ID, Name                        string
		Sector                          domain.Sector
		Province, Location, ProponentID string
		Stage                           domain.LifecycleStage
		Capex                           int64
		CapexStatus, Confidence         domain.ConfidenceLevel
		EvidenceIDs                     []string
		LastUpdate, UpdatedAt           time.Time
	}
	type eventInput struct {
		ID, Type, EvidenceID string
		Date, Created        time.Time
		Previous, Next       string
	}
	type capitalInput struct {
		ID, Category, Status, AmountType, EvidenceID string
		Amount                                       int64
		Created                                      time.Time
	}
	type relationshipInput struct {
		ID, Type, Confidence, EvidenceID string
		Created                          time.Time
		ValidFrom, ValidTo               *time.Time
	}
	type procurementInput struct {
		ID, Stage, Class, EvidenceID string
		Amount                       int64
		Created                      time.Time
		Closing                      *time.Time
	}
	type opportunityInput struct {
		ID, Class, Category, EstimateStatus string
		Amount                              int64
		Created                             time.Time
	}
	type signalInput struct {
		ID, Type, EvidenceID  string
		Timestamp             time.Time
		Magnitude, Confidence float64
	}
	p := ctx.project
	value := struct {
		Version       string
		AsOf          time.Time
		Horizons      []int
		Scenarios     []Scenario
		Project       projectInput
		Events        []eventInput
		Capital       []capitalInput
		Relationships []relationshipInput
		Procurements  []procurementInput
		Opportunities []opportunityInput
		Signals       []signalInput
		Evidence      []EvidenceReference
		Warnings      []string
	}{Version: MethodologyVersion, AsOf: req.asOf, Horizons: req.horizons, Scenarios: req.scenarios,
		Project:  projectInput{p.ID, p.Name, p.Sector, p.Province, p.LocationName, p.ProponentID, p.CurrentStage, p.CapexCAD, p.CapexStatus, p.Confidence, sortedUnique(p.EvidenceIDs), p.LastMeaningfulUpdate.UTC(), p.UpdatedAt.UTC()},
		Evidence: buildProvenance(ctx), Warnings: ctx.warnings}
	for _, item := range ctx.events {
		previous, next := "", ""
		if item.PreviousStage != nil {
			previous = string(*item.PreviousStage)
		}
		if item.NewStage != nil {
			next = string(*item.NewStage)
		}
		value.Events = append(value.Events, eventInput{item.ID, item.EventType, item.EvidenceID, item.EventDate.UTC(), item.CreatedAt.UTC(), previous, next})
	}
	for _, item := range ctx.capital {
		value.Capital = append(value.Capital, capitalInput{item.ID, string(item.Category), string(item.Status), item.AmountType, item.EvidenceID, item.AmountCAD, item.CreatedAt.UTC()})
	}
	for _, item := range ctx.relationships {
		value.Relationships = append(value.Relationships, relationshipInput{item.ID, item.RelationType, string(item.Confidence), item.EvidenceID, item.CreatedAt.UTC(), utcPtr(item.ValidFrom), utcPtr(item.ValidTo)})
	}
	for _, item := range ctx.procurements {
		value.Procurements = append(value.Procurements, procurementInput{item.ID, item.Stage, string(item.RequirementClass), item.EvidenceID, item.EstimatedCAD, item.CreatedAt.UTC(), utcPtr(item.ClosingDate)})
	}
	for _, item := range ctx.opportunities {
		value.Opportunities = append(value.Opportunities, opportunityInput{item.ID, string(item.RequirementClass), item.Category, string(item.EstimateStatus), item.EstimatedCAD, item.CreatedAt.UTC()})
	}
	for _, item := range ctx.signals {
		value.Signals = append(value.Signals, signalInput{item.ID, string(item.Type), item.EvidenceID, item.Timestamp.UTC(), item.Magnitude, item.Confidence})
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func defaultMonthsTo(current, target domain.LifecycleStage) float64 {
	type schedule struct{ fid, construction, operating float64 }
	values := map[domain.LifecycleStage]schedule{
		domain.StageUnknown: {60, 66, 102}, domain.StageDiscovered: {48, 54, 90}, domain.StageAnnounced: {42, 48, 84},
		domain.StageReferred: {36, 42, 78}, domain.StageEarlyDevelopment: {36, 42, 78}, domain.StageFeasibility: {30, 36, 72},
		domain.StageFinancing: {18, 24, 60}, domain.StageEnvironmentalReview: {24, 30, 66}, domain.StagePermitting: {18, 24, 60},
		domain.StageProcurement: {12, 18, 54}, domain.StageFIDLikely: {6, 12, 48}, domain.StageFID: {0, 6, 42},
		domain.StageConstruction: {0, 0, 36}, domain.StageCommissioning: {0, 0, 6}, domain.StageOperating: {0, 0, 0},
		domain.StageDelayed: {48, 54, 90}, domain.StagePaused: {54, 60, 96},
	}
	value, ok := values[current]
	if !ok {
		value = values[domain.StageUnknown]
	}
	switch target {
	case domain.StageFID:
		return value.fid
	case domain.StageConstruction:
		return value.construction
	default:
		return value.operating
	}
}

func stageCeiling(stage domain.LifecycleStage) float64 {
	switch stage {
	case domain.StageDiscovered:
		return 42
	case domain.StageAnnounced:
		return 50
	case domain.StageReferred, domain.StageEarlyDevelopment:
		return 56
	case domain.StageFeasibility:
		return 62
	case domain.StageFinancing:
		return 68
	case domain.StageEnvironmentalReview:
		return 67
	case domain.StagePermitting:
		return 75
	case domain.StageProcurement:
		return 80
	case domain.StageFIDLikely:
		return 86
	case domain.StageFID:
		return 94
	case domain.StageConstruction:
		return 97
	case domain.StageCommissioning:
		return 99
	case domain.StageOperating:
		return 100
	case domain.StageDelayed:
		return 42
	case domain.StagePaused:
		return 32
	case domain.StageCancelled:
		return 0
	default:
		return 30
	}
}

func validLifecycleStage(stage domain.LifecycleStage) bool {
	if stage == "" {
		return true
	}
	for _, candidate := range domain.ValidLifecycleStages {
		if stage == candidate {
			return true
		}
	}
	return false
}

func hasReached(current, target domain.LifecycleStage) bool {
	order := map[domain.LifecycleStage]int{domain.StageUnknown: 0, domain.StageDiscovered: 1, domain.StageAnnounced: 2, domain.StageReferred: 3,
		domain.StageEarlyDevelopment: 4, domain.StageFeasibility: 5, domain.StageEnvironmentalReview: 6, domain.StageFinancing: 6,
		domain.StagePermitting: 7, domain.StageProcurement: 8, domain.StageFIDLikely: 9, domain.StageFID: 10,
		domain.StageConstruction: 11, domain.StageCommissioning: 12, domain.StageOperating: 13}
	currentOrder, currentOK := order[current]
	targetOrder, targetOK := order[target]
	return currentOK && targetOK && currentOrder >= targetOrder
}

func scheduleUncertainty(confidence ConfidenceRating, stage domain.LifecycleStage) float64 {
	value := map[ConfidenceRating]float64{ConfidenceHigh: 0.18, ConfidenceModerate: 0.28, ConfidenceLow: 0.42, ConfidenceInsufficient: 0.60}[confidence]
	if stage == domain.StageUnknown || stage == domain.StageDelayed || stage == domain.StagePaused {
		value += 0.10
	}
	return clamp(value, 0.15, 0.75)
}

func scheduleFactor(stage domain.LifecycleStage, scenario Scenario, momentum float64) float64 {
	factor := 1 + scenario.SupplyChainStress*0.30 - scenario.FinancingAvailabilityDelta*0.12 - scenario.PolicySupportDelta*0.08 - scenario.DemandDelta*0.05
	factor *= 1 - momentum*0.12
	if stage == domain.StageDelayed {
		factor += 0.20
	}
	if stage == domain.StagePaused {
		factor += 0.35
	}
	return clamp(factor, 0.50, 2.50)
}

func capexUncertainty(confidence domain.ConfidenceLevel) float64 {
	switch confidence {
	case domain.ConfidenceVerified:
		return 0.05
	case domain.ConfidenceSupported:
		return 0.10
	case domain.ConfidenceReported:
		return 0.20
	case domain.ConfidenceInferred:
		return 0.35
	case domain.ConfidenceStale:
		return 0.40
	case domain.ConfidenceConflict:
		return 0.50
	default:
		return 0.50
	}
}

func completionCurve(horizon, expected, ramp float64) float64 {
	if expected <= 0 {
		return 100
	}
	return clamp(50+50*(horizon-expected)/math.Max(1, ramp), 0, 100)
}

func orderedLikelihood(low, base, high float64) LikelihoodRange {
	low = clamp(low, 0, 100)
	base = clamp(base, low, 100)
	high = clamp(high, base, 100)
	return LikelihoodRange{Low: round1(low), Base: round1(base), High: round1(high)}
}

func addCalendarMonths(value time.Time, months int) time.Time {
	value = value.UTC()
	year, month, day := value.Date()
	first := time.Date(year, month, 1, value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), time.UTC).AddDate(0, months, 0)
	lastDay := time.Date(first.Year(), first.Month()+1, 0, value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(first.Year(), first.Month(), day, value.Hour(), value.Minute(), value.Second(), value.Nanosecond(), time.UTC)
}

func statusForQuality(quality DataQuality) domain.IntelligenceStatus {
	if quality.ReferencedEvidence > 0 && quality.EvidenceFreshnessPct < 30 {
		return domain.StatusStale
	}
	switch quality.Confidence {
	case ConfidenceHigh:
		if len(quality.CriticalGaps) == 0 {
			return domain.StatusHealthy
		}
		return domain.StatusPartial
	case ConfidenceModerate:
		return domain.StatusPartial
	case ConfidenceLow:
		return domain.StatusPartial
	default:
		return domain.StatusDegraded
	}
}

func confidenceForScore(score float64) ConfidenceRating {
	switch {
	case score >= 80:
		return ConfidenceHigh
	case score >= 60:
		return ConfidenceModerate
	case score >= 35:
		return ConfidenceLow
	default:
		return ConfidenceInsufficient
	}
}

func tierScore(tier domain.SourceTier) float64 {
	switch tier {
	case domain.SourceTier1:
		return 100
	case domain.SourceTier2:
		return 80
	case domain.SourceTier3:
		return 60
	case domain.SourceTier4:
		return 40
	default:
		return 20
	}
}
func confidenceScore(value domain.ConfidenceLevel) float64 {
	switch value {
	case domain.ConfidenceVerified:
		return 100
	case domain.ConfidenceSupported:
		return 80
	case domain.ConfidenceReported:
		return 60
	case domain.ConfidenceInferred:
		return 35
	case domain.ConfidenceStale:
		return 20
	case domain.ConfidenceConflict:
		return 10
	case domain.ConfidenceRetracted:
		return 0
	default:
		return 20
	}
}
func freshnessScore(ageDays int) float64 {
	switch {
	case ageDays <= 90:
		return 100
	case ageDays <= 365:
		return 70
	case ageDays <= 730:
		return 40
	default:
		return 10
	}
}
func isKnownConfidence(value domain.ConfidenceLevel) bool {
	return value == domain.ConfidenceVerified || value == domain.ConfidenceSupported || value == domain.ConfidenceReported
}

func numericAssumption(parameter string, value float64, unit string, source AssumptionSource, rationale string, evidenceIDs []string) Assumption {
	copyValue := value
	return Assumption{Parameter: parameter, NumericValue: &copyValue, Unit: unit, Source: source, Rationale: rationale, EvidenceIDs: sortedUnique(evidenceIDs)}
}
func textAssumption(parameter, value string, source AssumptionSource, rationale string, evidenceIDs []string) Assumption {
	return Assumption{Parameter: parameter, TextValue: value, Source: source, Rationale: rationale, EvidenceIDs: sortedUnique(evidenceIDs)}
}
func appendIf(values []string, value string) []string {
	if value == "" {
		return values
	}
	return append(values, value)
}
func positiveDifference(left, right int64) int64 {
	if left <= right {
		return 0
	}
	return left - right
}
func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
func round1(value float64) float64 { return math.Round(value*10) / 10 }
func safeAdd(left, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, errors.New("integer overflow")
	}
	return left + right, nil
}
func safeScale(value int64, factor float64) (int64, error) {
	result := float64(value) * factor
	if math.IsNaN(result) || math.IsInf(result, 0) || result < 0 || result >= math.Exp2(63) {
		return 0, errors.New("money range overflow")
	}
	return int64(math.Round(result)), nil
}
func sortedUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
func utcPtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}
