// Package forecast provides deterministic, explainable project planning
// forecasts. It deliberately does not claim to be a trained or calibrated
// machine-learning model: completion percentages are planning likelihood
// indices produced from documented rules, supplied scenarios, and evidence
// available at a caller-controlled as-of time.
package forecast

import (
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const (
	// MethodologyVersion changes whenever a rule, weight, or default assumption
	// changes in a way that can alter an output.
	MethodologyVersion = "deterministic-scenario-v1.0"

	DefaultMaxHorizonMonths = 240
	DefaultMaxScenarios     = 16
	DefaultMaxRecords       = 10_000
	DefaultMaxProjects      = 1_000
)

// ConfidenceRating describes support for a forecast, not project quality.
type ConfidenceRating string

const (
	ConfidenceHigh         ConfidenceRating = "HIGH"
	ConfidenceModerate     ConfidenceRating = "MODERATE"
	ConfidenceLow          ConfidenceRating = "LOW"
	ConfidenceInsufficient ConfidenceRating = "INSUFFICIENT"
)

// AssumptionSource distinguishes observed inputs from methodology defaults and
// caller-authored scenario choices.
type AssumptionSource string

const (
	AssumptionObserved    AssumptionSource = "OBSERVED"
	AssumptionMethodology AssumptionSource = "METHODOLOGY_DEFAULT"
	AssumptionScenario    AssumptionSource = "SCENARIO_INPUT"
)

// Context is the complete project-local input boundary. Evaluate validates
// project ownership and ignores records that were not knowable at AsOf.
type Context struct {
	Project       *domain.Project
	Events        []*domain.Event
	CapitalItems  []*domain.CapitalItem
	Relationships []*domain.Relationship
	Procurements  []*domain.Procurement
	Opportunities []*domain.Opportunity
	Signals       []*domain.Signal
	Evidence      []*domain.Evidence
}

// Request controls the point-in-time forecast and scenario set. Empty horizons
// default to 12, 36, and 60 months. Empty scenarios default to BaselineScenario.
type Request struct {
	AsOf             time.Time  `json:"as_of"`
	HorizonsMonths   []int      `json:"horizons_months,omitempty"`
	Scenarios        []Scenario `json:"scenarios,omitempty"`
	MaxHorizonMonths int        `json:"max_horizon_months,omitempty"`
}

// Scenario contains bounded stress-test inputs. Deltas are expressed as
// fractions: 0.20 means +20 percentage points/intensity, not 20x.
type Scenario struct {
	ID                         string  `json:"id"`
	Name                       string  `json:"name"`
	Description                string  `json:"description,omitempty"`
	FinancingAvailabilityDelta float64 `json:"financing_availability_delta"` // -1 to +1
	RegulatoryDelayMonths      int     `json:"regulatory_delay_months"`      // -24 to +120
	ConstructionDelayMonths    int     `json:"construction_delay_months"`    // -24 to +120
	CapexEscalationPct         float64 `json:"capex_escalation_pct"`         // -50 to +200
	DemandDelta                float64 `json:"demand_delta"`                 // -1 to +1
	PolicySupportDelta         float64 `json:"policy_support_delta"`         // -1 to +1
	SupplyChainStress          float64 `json:"supply_chain_stress"`          // 0 to 1
}

// BaselineScenario returns a neutral scenario. Neutral means no caller-applied
// stress; observed project conditions and methodology defaults still apply.
func BaselineScenario() Scenario {
	return Scenario{
		ID:          "baseline",
		Name:        "Baseline",
		Description: "Observed conditions with no caller-applied stress adjustment.",
	}
}

// StandardScenarios returns transparent generic planning stresses. They are
// not predictions and are never applied unless the caller supplies them.
func StandardScenarios() []Scenario {
	return []Scenario{
		BaselineScenario(),
		{
			ID:                         "delivery_stress",
			Name:                       "Delivery stress",
			Description:                "Generic financing, approval, delivery, and cost stress test.",
			FinancingAvailabilityDelta: -0.25,
			RegulatoryDelayMonths:      12,
			ConstructionDelayMonths:    9,
			CapexEscalationPct:         20,
			DemandDelta:                -0.10,
			PolicySupportDelta:         -0.15,
			SupplyChainStress:          0.60,
		},
		{
			ID:                         "accelerated_delivery",
			Name:                       "Accelerated delivery",
			Description:                "Generic coordinated-delivery stress test; it is not an endorsed outcome.",
			FinancingAvailabilityDelta: 0.20,
			RegulatoryDelayMonths:      -6,
			ConstructionDelayMonths:    -3,
			DemandDelta:                0.15,
			PolicySupportDelta:         0.20,
			SupplyChainStress:          0.10,
		},
	}
}

// Report is a point-in-time, reproducible project forecast.
type Report struct {
	ID                 string                    `json:"id"`
	ProjectID          string                    `json:"project_id"`
	ProjectName        string                    `json:"project_name"`
	MethodologyVersion string                    `json:"methodology_version"`
	AsOf               time.Time                 `json:"as_of"`
	CalculatedAt       time.Time                 `json:"calculated_at"`
	Status             domain.IntelligenceStatus `json:"status"`
	Confidence         ConfidenceRating          `json:"confidence"`
	InputHash          string                    `json:"input_hash"`
	DataQuality        DataQuality               `json:"data_quality"`
	Scenarios          []ScenarioForecast        `json:"scenarios"`
	Provenance         []EvidenceReference       `json:"provenance"`
	Warnings           []string                  `json:"warnings,omitempty"`
	Limitations        []string                  `json:"limitations"`
	Assurance          SovereignAssurance        `json:"sovereign_assurance"`
}

// DataQuality makes forecast support and missing information independent from
// the project outcome.
type DataQuality struct {
	Score                 float64          `json:"score"`
	Confidence            ConfidenceRating `json:"confidence"`
	CoveragePct           float64          `json:"coverage_pct"`
	EvidenceQualityPct    float64          `json:"evidence_quality_pct"`
	EvidenceFreshnessPct  float64          `json:"evidence_freshness_pct"`
	NewestEvidenceAgeDays *int             `json:"newest_evidence_age_days,omitempty"`
	ReferencedEvidence    int              `json:"referenced_evidence"`
	MissingEvidenceRefs   int              `json:"missing_evidence_refs"`
	ConflictedEvidence    int              `json:"conflicted_evidence"`
	RecordCounts          RecordCounts     `json:"record_counts"`
	CriticalGaps          []string         `json:"critical_gaps,omitempty"`
}

type RecordCounts struct {
	Events        int `json:"events"`
	CapitalItems  int `json:"capital_items"`
	Relationships int `json:"relationships"`
	Procurements  int `json:"procurements"`
	Opportunities int `json:"opportunities"`
	Signals       int `json:"signals"`
}

type ScenarioForecast struct {
	Scenario    Scenario            `json:"scenario"`
	Assumptions []Assumption        `json:"assumptions"`
	Capital     CapitalOutlook      `json:"capital_outlook"`
	Milestones  []MilestoneForecast `json:"milestones"`
	Drivers     []Driver            `json:"drivers"`
	WatchItems  []WatchItem         `json:"watch_items,omitempty"`
}

// Assumption is an auditable numeric or textual model input.
type Assumption struct {
	Parameter    string           `json:"parameter"`
	NumericValue *float64         `json:"numeric_value,omitempty"`
	TextValue    string           `json:"text_value,omitempty"`
	Unit         string           `json:"unit,omitempty"`
	Source       AssumptionSource `json:"source"`
	Rationale    string           `json:"rationale"`
	EvidenceIDs  []string         `json:"evidence_ids,omitempty"`
}

type MoneyRange struct {
	Low  int64 `json:"low"`
	Base int64 `json:"base"`
	High int64 `json:"high"`
}

type CapitalOutlook struct {
	ReportedCapexCAD          int64      `json:"reported_capex_cad"`
	ScenarioAdjustedCapexCAD  MoneyRange `json:"scenario_adjusted_capex_cad"`
	EvidencedCommittedCAD     int64      `json:"evidenced_committed_cad"`
	ConditionallyCommittedCAD int64      `json:"conditionally_committed_cad"`
	AnnouncedOrProposedCAD    int64      `json:"announced_or_proposed_cad"`
	IndicativeFundingGapCAD   MoneyRange `json:"indicative_funding_gap_cad"`
	ConfirmedProcurementCAD   int64      `json:"confirmed_procurement_cad"`
	ConfirmedOpportunityCAD   int64      `json:"confirmed_opportunity_cad"`
	DerivedOpportunityCAD     int64      `json:"derived_opportunity_cad"`
	CapexUncertaintyPct       float64    `json:"capex_uncertainty_pct"`
}

type DateRange struct {
	Earliest *time.Time `json:"earliest,omitempty"`
	Base     *time.Time `json:"base,omitempty"`
	Latest   *time.Time `json:"latest,omitempty"`
}

type LikelihoodRange struct {
	Low  float64 `json:"low"`
	Base float64 `json:"base"`
	High float64 `json:"high"`
}

type HorizonLikelihood struct {
	HorizonMonths                int             `json:"horizon_months"`
	HorizonDate                  time.Time       `json:"horizon_date"`
	CompletionLikelihoodIndexPct LikelihoodRange `json:"completion_likelihood_index_pct"`
}

// MilestoneForecast is conditional on the project continuing. Its likelihood
// index is a transparent planning score, not a statistically calibrated
// probability or investment recommendation.
type MilestoneForecast struct {
	Stage               domain.LifecycleStage `json:"stage"`
	ConditionalSchedule DateRange             `json:"conditional_schedule"`
	ByHorizon           []HorizonLikelihood   `json:"by_horizon"`
	Basis               string                `json:"basis"`
}

type Driver struct {
	Code         string   `json:"code"`
	Label        string   `json:"label"`
	ImpactPoints float64  `json:"impact_points"`
	Explanation  string   `json:"explanation"`
	EvidenceIDs  []string `json:"evidence_ids,omitempty"`
}

type WatchItem struct {
	Code     string `json:"code"`
	Severity string `json:"severity"` // INFO, MEDIUM, HIGH
	Message  string `json:"message"`
}

// EvidenceReference intentionally excludes raw snippets. It is sufficient to
// audit the source without replicating potentially sensitive source text.
type EvidenceReference struct {
	EvidenceID    string                 `json:"evidence_id"`
	Publisher     string                 `json:"publisher"`
	SourceURL     string                 `json:"source_url,omitempty"`
	SourceTier    domain.SourceTier      `json:"source_tier"`
	Confidence    domain.ConfidenceLevel `json:"confidence"`
	RetrievalTime time.Time              `json:"retrieval_time"`
	EffectiveDate *time.Time             `json:"effective_date,omitempty"`
	ContentHash   string                 `json:"content_hash,omitempty"`
}

// SovereignAssurance describes execution properties; it is not a security
// certification or a representation about the caller's deployment environment.
type SovereignAssurance struct {
	ProcessingMode          string `json:"processing_mode"`
	ExternalNetworkRequired bool   `json:"external_network_required"`
	ExternalTelemetry       bool   `json:"external_telemetry"`
	RawEvidenceEmitted      bool   `json:"raw_evidence_emitted"`
	Deterministic           bool   `json:"deterministic"`
	Statement               string `json:"statement"`
}

// PortfolioReport aggregates project reports without hiding their individual
// confidence, provenance, or limitations.
type PortfolioReport struct {
	ID                 string                    `json:"id"`
	MethodologyVersion string                    `json:"methodology_version"`
	AsOf               time.Time                 `json:"as_of"`
	CalculatedAt       time.Time                 `json:"calculated_at"`
	Status             domain.IntelligenceStatus `json:"status"`
	Confidence         ConfidenceRating          `json:"confidence"`
	InputHash          string                    `json:"input_hash"`
	ProjectCount       int                       `json:"project_count"`
	DataQuality        PortfolioDataQuality      `json:"data_quality"`
	ScenarioAggregates []PortfolioScenario       `json:"scenario_aggregates"`
	Concentrations     []Concentration           `json:"concentrations"`
	ProjectForecasts   []*Report                 `json:"project_forecasts"`
	WatchItems         []WatchItem               `json:"watch_items,omitempty"`
	Warnings           []string                  `json:"warnings,omitempty"`
	Limitations        []string                  `json:"limitations"`
	Assurance          SovereignAssurance        `json:"sovereign_assurance"`
}

type PortfolioDataQuality struct {
	AverageScore         float64 `json:"average_score"`
	HighProjects         int     `json:"high_projects"`
	ModerateProjects     int     `json:"moderate_projects"`
	LowProjects          int     `json:"low_projects"`
	InsufficientProjects int     `json:"insufficient_projects"`
	UnknownCapexProjects int     `json:"unknown_capex_projects"`
	StaleProjects        int     `json:"stale_projects"`
}

type PortfolioScenario struct {
	ScenarioID                string                      `json:"scenario_id"`
	ScenarioName              string                      `json:"scenario_name"`
	ScenarioAdjustedCapexCAD  MoneyRange                  `json:"scenario_adjusted_capex_cad"`
	EvidencedCommittedCAD     int64                       `json:"evidenced_committed_cad"`
	ConditionallyCommittedCAD int64                       `json:"conditionally_committed_cad"`
	IndicativeFundingGapCAD   MoneyRange                  `json:"indicative_funding_gap_cad"`
	ConfirmedProcurementCAD   int64                       `json:"confirmed_procurement_cad"`
	ConfirmedOpportunityCAD   int64                       `json:"confirmed_opportunity_cad"`
	DerivedOpportunityCAD     int64                       `json:"derived_opportunity_cad"`
	MilestoneSummaries        []PortfolioMilestoneSummary `json:"milestone_summaries"`
}

// PortfolioMilestoneSummary aggregates planning indices. The 50-point count is
// a screening threshold, not a count of projects guaranteed to complete.
type PortfolioMilestoneSummary struct {
	Stage                      domain.LifecycleStage `json:"stage"`
	HorizonMonths              int                   `json:"horizon_months"`
	HorizonDate                time.Time             `json:"horizon_date"`
	ProjectsAtOrAbove50Index   int                   `json:"projects_at_or_above_50_index"`
	MeanBaseLikelihoodIndexPct float64               `json:"mean_base_likelihood_index_pct"`
	CapexWeightedBaseIndexPct  *float64              `json:"capex_weighted_base_index_pct,omitempty"`
}

// Concentration reports shares of known reported capex. Unknown capex is
// excluded from the denominator and reported separately in DataQuality.
type Concentration struct {
	Dimension          string  `json:"dimension"` // province or sector
	Value              string  `json:"value"`
	ProjectCount       int     `json:"project_count"`
	ReportedCapexCAD   int64   `json:"reported_capex_cad"`
	KnownCapexSharePct float64 `json:"known_capex_share_pct"`
}
