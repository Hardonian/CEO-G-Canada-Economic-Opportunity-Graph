package domain

import (
	"encoding/json"
	"time"
)

// VisibilityClass is the redistribution boundary attached to every source and
// source-derived record. Restricted values are intentionally not ordered: a
// caller must make an explicit allow decision rather than compare severities.
type VisibilityClass string

const (
	VisibilityPublic               VisibilityClass = "PUBLIC"
	VisibilityPublicAttribution    VisibilityClass = "PUBLIC_WITH_ATTRIBUTION"
	VisibilityLicensedPrivate      VisibilityClass = "LICENSED_PRIVATE"
	VisibilityUserPrivate          VisibilityClass = "USER_PRIVATE"
	VisibilityInternalRestricted   VisibilityClass = "INTERNAL_RESTRICTED"
)

func (v VisibilityClass) Valid() bool {
	switch v {
	case VisibilityPublic, VisibilityPublicAttribution, VisibilityLicensedPrivate,
		VisibilityUserPrivate, VisibilityInternalRestricted:
		return true
	default:
		return false
	}
}

func (v VisibilityClass) Public() bool {
	return v == VisibilityPublic || v == VisibilityPublicAttribution
}

type PublicationState string

const (
	PublicationPublicCanonical PublicationState = "PUBLIC_CANONICAL"
	PublicationPrivateOnly     PublicationState = "PRIVATE_ONLY"
	PublicationUnverified      PublicationState = "UNVERIFIED"
	PublicationConflicted      PublicationState = "CONFLICTED"
)

// Claim is an immutable assertion. Canonical values are projections over a set
// of claims; SourceID/EvidenceID never collapse to a single URL on the project.
type Claim struct {
	ID                     string                 `json:"claim_id"`
	SubjectID              string                 `json:"subject"`
	SubjectType            string                 `json:"subject_type"`
	Predicate              string                 `json:"predicate"`
	Value                  json.RawMessage        `json:"value"`
	Unit                   string                 `json:"unit,omitempty"`
	Currency               string                 `json:"currency,omitempty"`
	EffectiveAt            *time.Time             `json:"effective_at,omitempty"`
	ObservedAt             time.Time              `json:"observed_at"`
	SourceID               string                 `json:"source_id"`
	EvidenceID             string                 `json:"evidence_id"`
	SourceVisibility       VisibilityClass        `json:"source_visibility"`
	Status                 ConfidenceLevel        `json:"claim_status"`
	Confidence             float64                `json:"confidence"`
	AuthorityClass         string                 `json:"authority_class"`
	Publishable            bool                   `json:"publishable"`
	PublicationState       PublicationState       `json:"publication_state"`
	CorroborationCount     int                    `json:"corroboration_count"`
	IndependentSourceCount int                    `json:"independent_source_count"`
	Supersedes             []string               `json:"supersedes,omitempty"`
	ConflictsWith          []string               `json:"conflicts_with,omitempty"`
	ExtractorVersion       string                 `json:"extractor_version"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
}

type AmountType string

const (
	AmountExact        AmountType = "EXACT"
	AmountApproximate  AmountType = "APPROXIMATE"
	AmountRange        AmountType = "RANGE"
	AmountMinimum      AmountType = "MINIMUM"
	AmountMaximum      AmountType = "MAXIMUM"
	AmountNotAvailable AmountType = "NOT_AVAILABLE"
)

// MonetaryAmount preserves the reported currency and semantics. CAD conversion
// is a separate, fully sourced observation and never overwrites the original.
type MonetaryAmount struct {
	Amount        *int64     `json:"amount,omitempty"`
	Minimum       *int64     `json:"minimum,omitempty"`
	Maximum       *int64     `json:"maximum,omitempty"`
	Currency      string     `json:"currency,omitempty"`
	AmountType    AmountType `json:"amount_type"`
	OriginalText  string     `json:"original_text,omitempty"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	EvidenceID    string     `json:"evidence_id,omitempty"`
	Converted     *FXAmount  `json:"converted,omitempty"`
}

type FXAmount struct {
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Rate       float64   `json:"fx_rate"`
	Source     string    `json:"fx_source"`
	Date       time.Time `json:"fx_date"`
}

type ProjectPhase struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	Name           string          `json:"name"`
	Sequence       int             `json:"sequence"`
	OriginalStage  string          `json:"original_stage,omitempty"`
	NormalizedStages []LifecycleStage `json:"normalized_stages,omitempty"`
	EvidenceIDs    []string        `json:"evidence_ids,omitempty"`
	Visibility     VisibilityClass `json:"visibility"`
	Publishable    bool            `json:"publishable"`
	CreatedAt      time.Time       `json:"created_at"`
}

type CapitalNeedType string

const (
	NeedEquity                CapitalNeedType = "EQUITY"
	NeedDebt                  CapitalNeedType = "DEBT"
	NeedSeniorDebt            CapitalNeedType = "SENIOR_DEBT"
	NeedSubordinatedDebt      CapitalNeedType = "SUBORDINATED_DEBT"
	NeedProjectFinance        CapitalNeedType = "PROJECT_FINANCE"
	NeedInfrastructureEquity  CapitalNeedType = "INFRASTRUCTURE_EQUITY"
	NeedPrivateCredit         CapitalNeedType = "PRIVATE_CREDIT"
	NeedJointVenture          CapitalNeedType = "JOINT_VENTURE"
	NeedStrategicInvestment   CapitalNeedType = "STRATEGIC_INVESTMENT"
	NeedGovernmentSupport     CapitalNeedType = "GOVERNMENT_SUPPORT"
	NeedGrant                 CapitalNeedType = "GRANT"
	NeedLoanGuarantee         CapitalNeedType = "LOAN_GUARANTEE"
	NeedExportCredit          CapitalNeedType = "EXPORT_CREDIT"
	NeedIndigenousEquity      CapitalNeedType = "INDIGENOUS_EQUITY"
	NeedPensionCapital        CapitalNeedType = "PENSION_CAPITAL"
	NeedSovereignCapital      CapitalNeedType = "SOVEREIGN_CAPITAL"
	NeedOfftake               CapitalNeedType = "OFFTAKE"
	NeedAnchorTenant          CapitalNeedType = "ANCHOR_TENANT"
	NeedPrepayment            CapitalNeedType = "PREPAYMENT"
	NeedStreaming             CapitalNeedType = "STREAMING"
	NeedRoyalty               CapitalNeedType = "ROYALTY"
	NeedConcession            CapitalNeedType = "CONCESSION"
	NeedLease                 CapitalNeedType = "LEASE"
	NeedTolling               CapitalNeedType = "TOLLING"
	NeedServiceAgreement      CapitalNeedType = "SERVICE_AGREEMENT"
	NeedFirmTransportation    CapitalNeedType = "FIRM_TRANSPORTATION"
	NeedCommercialPartnership CapitalNeedType = "COMMERCIAL_PARTNERSHIP"
	NeedTechnologyPartnership CapitalNeedType = "TECHNOLOGY_PARTNERSHIP"
	NeedDevelopmentCapital    CapitalNeedType = "DEVELOPMENT_CAPITAL"
)

type CounterpartyType string

const (
	CounterpartyInfrastructureFund CounterpartyType = "INFRASTRUCTURE_FUND"
	CounterpartyPensionFund        CounterpartyType = "PENSION_FUND"
	CounterpartyPrivateEquity      CounterpartyType = "PRIVATE_EQUITY"
	CounterpartyBank               CounterpartyType = "BANK"
	CounterpartyPrivateCredit      CounterpartyType = "PRIVATE_CREDIT"
	CounterpartyECA                CounterpartyType = "EXPORT_CREDIT_AGENCY"
	CounterpartySovereignFund      CounterpartyType = "SOVEREIGN_WEALTH_FUND"
	CounterpartyStrategicCorporate CounterpartyType = "STRATEGIC_CORPORATE"
	CounterpartyEPC                CounterpartyType = "EPC_CONTRACTOR"
	CounterpartyOEM                CounterpartyType = "OEM"
	CounterpartyOperator           CounterpartyType = "OPERATOR"
	CounterpartyOfftaker           CounterpartyType = "OFFTAKER"
	CounterpartyUtility            CounterpartyType = "UTILITY"
	CounterpartyAnchorTenant       CounterpartyType = "ANCHOR_TENANT"
	CounterpartyLogisticsOperator  CounterpartyType = "LOGISTICS_OPERATOR"
	CounterpartyTerminalOperator   CounterpartyType = "TERMINAL_OPERATOR"
	CounterpartyTechnologyProvider CounterpartyType = "TECHNOLOGY_PROVIDER"
	CounterpartyIndigenousPartner  CounterpartyType = "INDIGENOUS_PARTNER"
	CounterpartyGovernment         CounterpartyType = "GOVERNMENT"
	CounterpartyJVPartner          CounterpartyType = "JOINT_VENTURE_PARTNER"
)

type CapitalRequirement struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	PhaseID     string          `json:"phase_id,omitempty"`
	Purpose     string          `json:"purpose,omitempty"`
	Amount      MonetaryAmount  `json:"amount"`
	Status      ConfidenceLevel `json:"status"`
	EvidenceIDs []string        `json:"evidence_ids,omitempty"`
	Visibility  VisibilityClass `json:"visibility"`
	Publishable bool            `json:"publishable"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CapitalNeed struct {
	ID                   string             `json:"id"`
	ProjectID            string             `json:"project_id"`
	PhaseID              string             `json:"phase_id,omitempty"`
	Types                []CapitalNeedType  `json:"types"`
	Counterparties       []CounterpartyType `json:"counterparties,omitempty"`
	Amount               *MonetaryAmount    `json:"amount,omitempty"`
	Status               CapitalStatus      `json:"status"`
	OriginalLanguage     string             `json:"original_language,omitempty"`
	EvidenceIDs          []string           `json:"evidence_ids"`
	Visibility           VisibilityClass    `json:"visibility"`
	Publishable          bool               `json:"publishable"`
	PublicationState     PublicationState   `json:"publication_state"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

type MilestoneType string

const (
	MilestoneLandSecured          MilestoneType = "LAND_SECURED"
	MilestoneSiteSelected         MilestoneType = "SITE_SELECTED"
	MilestoneFeasibilityComplete  MilestoneType = "FEASIBILITY_COMPLETE"
	MilestoneFEEDComplete         MilestoneType = "FEED_COMPLETE"
	MilestoneEnvironmentalApproval MilestoneType = "ENVIRONMENTAL_APPROVAL"
	MilestonePermitReceived       MilestoneType = "PERMIT_RECEIVED"
	MilestoneGridConnection       MilestoneType = "GRID_CONNECTION"
	MilestoneOfftakeSigned        MilestoneType = "OFFTAKE_SIGNED"
	MilestoneFinancingCommitted   MilestoneType = "FINANCING_COMMITTED"
	MilestoneFinancialClose       MilestoneType = "FINANCIAL_CLOSE"
	MilestoneFID                  MilestoneType = "FID"
	MilestoneEPCAwarded           MilestoneType = "EPC_AWARDED"
	MilestoneConstructionStarted  MilestoneType = "CONSTRUCTION_STARTED"
	MilestoneCommercialOperation  MilestoneType = "COMMERCIAL_OPERATION"
)

type MilestoneStatus string

const (
	MilestoneUnknown   MilestoneStatus = "UNKNOWN"
	MilestonePlanned   MilestoneStatus = "PLANNED"
	MilestoneInProgress MilestoneStatus = "IN_PROGRESS"
	MilestoneComplete  MilestoneStatus = "COMPLETE"
	MilestoneDelayed   MilestoneStatus = "DELAYED"
	MilestoneCancelled MilestoneStatus = "CANCELLED"
)

type Milestone struct {
	ID           string          `json:"id"`
	ProjectID    string          `json:"project_id"`
	PhaseID      string          `json:"phase_id,omitempty"`
	Type         MilestoneType   `json:"type"`
	Status       MilestoneStatus `json:"status"`
	TargetDate   *time.Time      `json:"target_date,omitempty"`
	ActualDate   *time.Time      `json:"actual_date,omitempty"`
	EvidenceIDs  []string        `json:"evidence_ids"`
	Confidence   ConfidenceLevel `json:"confidence"`
	Visibility   VisibilityClass `json:"visibility"`
	Publishable  bool            `json:"publishable"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type ReadinessVector struct {
	Engineering    float64 `json:"engineering"`
	Regulatory     float64 `json:"regulatory"`
	Financing      float64 `json:"financing"`
	Commercial     float64 `json:"commercial"`
	Site           float64 `json:"site"`
	Infrastructure float64 `json:"infrastructure"`
	Offtake        float64 `json:"offtake"`
	Execution      float64 `json:"execution"`
}

type ReadinessAssessment struct {
	ID                 string              `json:"id"`
	ProjectID          string              `json:"project_id"`
	MethodologyVersion string              `json:"methodology_version"`
	Vector             ReadinessVector     `json:"vector"`
	InvestmentReadiness float64             `json:"investment_readiness"`
	Coverage           float64             `json:"coverage"`
	Factors            map[string]float64  `json:"factors"`
	FactorEvidence     map[string][]string `json:"factor_evidence"`
	UnknownFactors     []string            `json:"unknown_factors,omitempty"`
	Explanation        []string            `json:"explanation"`
	InputHash          string              `json:"input_hash"`
	CalculatedAt       time.Time           `json:"calculated_at"`
}

type CapitalGapState string

const (
	CapitalGapKnown      CapitalGapState = "KNOWN"
	CapitalGapPartial    CapitalGapState = "PARTIAL"
	CapitalGapUnknown    CapitalGapState = "UNKNOWN"
	CapitalGapConflicted CapitalGapState = "CONFLICTED"
)

type CapitalGap struct {
	ProjectID       string          `json:"project_id"`
	State           CapitalGapState `json:"state"`
	TotalRequirement *MonetaryAmount `json:"total_requirement,omitempty"`
	CompatibleCommitted *MonetaryAmount `json:"compatible_committed,omitempty"`
	PotentialGap    *MonetaryAmount `json:"potential_gap,omitempty"`
	Confidence      ConfidenceLevel `json:"confidence"`
	Explanation     []string        `json:"explanation"`
}

type AuditAction string

const (
	AuditPrivateSourceIngested AuditAction = "PRIVATE_SOURCE_INGESTED"
	AuditClaimPromoted         AuditAction = "CLAIM_PROMOTED"
	AuditClaimRejected         AuditAction = "CLAIM_REJECTED"
	AuditManualCorrection      AuditAction = "MANUAL_CORRECTION"
	AuditVisibilityChanged     AuditAction = "VISIBILITY_CHANGED"
	AuditPermissionChanged     AuditAction = "SOURCE_PERMISSION_CHANGED"
)

type AuditEntry struct {
	ID           string                 `json:"id"`
	Action       AuditAction            `json:"action"`
	Actor        string                 `json:"actor"`
	SubjectID    string                 `json:"subject_id"`
	Reason       string                 `json:"reason"`
	FromVisibility VisibilityClass      `json:"from_visibility,omitempty"`
	ToVisibility VisibilityClass        `json:"to_visibility,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	OccurredAt   time.Time              `json:"occurred_at"`
}

// CandidateProject is deliberately separate from Project: restricted source
// extraction cannot create a public canonical project as a side effect.
type CandidateProject struct {
	ID             string          `json:"id"`
	SourceID       string          `json:"source_id"`
	Name           string          `json:"name"`
	Proponent      string          `json:"proponent,omitempty"`
	Location       string          `json:"location,omitempty"`
	OriginalStage  string          `json:"original_stage,omitempty"`
	Visibility     VisibilityClass `json:"visibility"`
	ClaimIDs       []string        `json:"claim_ids"`
	CreatedAt      time.Time       `json:"created_at"`
}

// OpportunityFunnelState tracks the lifecycle of an opportunity from initial
// discovery through to financing close or death. This is the pipeline state
// machine that transforms the graph from a static database into an active
// investment intelligence surface.
type OpportunityFunnelState string

const (
	FunnelDiscovered         OpportunityFunnelState = "DISCOVERED"
	FunnelQualifying         OpportunityFunnelState = "QUALIFYING"
	FunnelResearching        OpportunityFunnelState = "RESEARCHING"
	FunnelCorroborated       OpportunityFunnelState = "CORROBORATED"
	FunnelActiveOpportunity  OpportunityFunnelState = "ACTIVE_OPPORTUNITY"
	FunnelFinancingProgress  OpportunityFunnelState = "FINANCING_IN_PROGRESS"
	FunnelClosed             OpportunityFunnelState = "CLOSED"
	FunnelStale              OpportunityFunnelState = "STALE"
	FunnelDead               OpportunityFunnelState = "DEAD"
)

// FIDStatus classifies the Final Investment Decision state.
type FIDStatus string

const (
	FIDNotApplicable FIDStatus = "NOT_APPLICABLE"
	FIDUnknown       FIDStatus = "UNKNOWN"
	FIDTarget        FIDStatus = "TARGET"
	FIDExpected      FIDStatus = "EXPECTED"
	FIDAchieved      FIDStatus = "ACHIEVED"
	FIDDelayed       FIDStatus = "DELAYED"
	FIDDeferred      FIDStatus = "DEFERRED"
	FIDNegative      FIDStatus = "NEGATIVE" // project decided NOT to proceed
)

// FIDIntelligence tracks the Final Investment Decision trajectory. FID is the
// single most important inflection point for capital deployment: before FID,
// the project is speculative; after FID, capital flows start.
type FIDIntelligence struct {
	ProjectID     string          `json:"project_id"`
	Status        FIDStatus       `json:"status"`
	TargetDate    *time.Time      `json:"target_date,omitempty"`
	EarliestDate  *time.Time      `json:"earliest_date,omitempty"`
	LatestDate    *time.Time      `json:"latest_date,omitempty"`
	ActualDate    *time.Time      `json:"actual_date,omitempty"`
	Confidence    ConfidenceLevel `json:"confidence"`
	Revisions     []FIDRevision   `json:"revisions,omitempty"`
	EvidenceIDs   []string        `json:"evidence_ids,omitempty"`
	Visibility    VisibilityClass `json:"visibility"`
	Publishable   bool            `json:"publishable"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// FIDRevision records a change in FID expectations, creating an auditable
// history of schedule evolution.
type FIDRevision struct {
	PreviousTarget *time.Time      `json:"previous_target,omitempty"`
	NewTarget      *time.Time      `json:"new_target,omitempty"`
	PreviousStatus FIDStatus       `json:"previous_status"`
	NewStatus      FIDStatus       `json:"new_status"`
	Reason         string          `json:"reason,omitempty"`
	EvidenceID     string          `json:"evidence_id,omitempty"`
	ObservedAt     time.Time       `json:"observed_at"`
}

// InvestorProfile extends Entity with investment-specific metadata for
// counterparty matching. This never stores personal contact data or
// subscription preferences — it records public sector mandates, deal
// history, and instrument preferences that are discoverable from public
// disclosures and regulatory filings.
type InvestorProfile struct {
	EntityID              string             `json:"entity_id"`
	InvestorTypes         []CounterpartyType `json:"investor_types"`
	TargetSectors         []Sector           `json:"target_sectors,omitempty"`
	TargetGeographies     []string           `json:"target_geographies,omitempty"` // province codes
	MinTicketCAD          int64              `json:"min_ticket_cad,omitempty"`
	MaxTicketCAD          int64              `json:"max_ticket_cad,omitempty"`
	PreferredInstruments  []CapitalNeedType  `json:"preferred_instruments,omitempty"`
	PreferredStages       []LifecycleStage   `json:"preferred_stages,omitempty"`
	CanadianExposureCAD   int64              `json:"canadian_exposure_cad,omitempty"`
	ActiveInvestments     int                `json:"active_investments,omitempty"`
	PublicDealHistory     []DealPrecedent    `json:"public_deal_history,omitempty"`
	EvidenceIDs           []string           `json:"evidence_ids,omitempty"`
	Visibility            VisibilityClass    `json:"visibility"`
	Publishable           bool               `json:"publishable"`
	UpdatedAt             time.Time          `json:"updated_at"`
}

// DealPrecedent records a public historical transaction for deal-matching.
type DealPrecedent struct {
	DealID          string          `json:"deal_id"`
	ProjectName     string          `json:"project_name"`
	Sector          Sector          `json:"sector"`
	Province        string          `json:"province,omitempty"`
	AmountCAD       int64           `json:"amount_cad,omitempty"`
	Instrument      CapitalNeedType `json:"instrument,omitempty"`
	Stage           LifecycleStage  `json:"stage,omitempty"`
	Year            int             `json:"year,omitempty"`
	EvidenceID      string          `json:"evidence_id,omitempty"`
}

// RequirementType classifies the infrastructure or service dependency of a
// project. These produce the dependency graph that drives opportunity
// propagation: a mine needs power, which needs transmission, which needs
// an environmental assessment, each of which is an opportunity.
type RequirementType string

const (
	RequirePower         RequirementType = "POWER"
	RequireTransmission  RequirementType = "TRANSMISSION"
	RequireFiber         RequirementType = "FIBER"
	RequireRoad          RequirementType = "ROAD"
	RequireRail          RequirementType = "RAIL"
	RequirePort          RequirementType = "PORT"
	RequireWater         RequirementType = "WATER"
	RequireNaturalGas    RequirementType = "NATURAL_GAS"
	RequireHydrogen      RequirementType = "HYDROGEN"
	RequireWastewater    RequirementType = "WASTEWATER"
	RequireAirport       RequirementType = "AIRPORT"
	RequireHousing       RequirementType = "WORKFORCE_HOUSING"
	RequireCooling       RequirementType = "COOLING"
	RequireInterconnect  RequirementType = "INTERCONNECTION"
	RequireStorage       RequirementType = "STORAGE"
	RequireProcessing    RequirementType = "PROCESSING"
	RequireLogistics     RequirementType = "LOGISTICS"
)

// RequirementConfidence classifies how the requirement was established.
type RequirementConfidence string

const (
	RequirementStated    RequirementConfidence = "STATED"
	RequirementInferred  RequirementConfidence = "INFERRED"
	RequirementOntology  RequirementConfidence = "ONTOLOGY_DERIVED"
)

// ProjectRequirement models a typed infrastructure or service dependency.
// These are the edges in the dependency graph that drive opportunity
// propagation: if a project requires POWER, that requirement propagates
// as a confirmed or inferred opportunity for power generation or
// transmission in the same geography.
type ProjectRequirement struct {
	ID             string                `json:"id"`
	ProjectID      string                `json:"project_id"`
	Type           RequirementType       `json:"type"`
	Description    string                `json:"description,omitempty"`
	CapacityNeeded *CapacityMetric       `json:"capacity_needed,omitempty"`
	Confidence     RequirementConfidence `json:"confidence"`
	SatisfiedBy    string                `json:"satisfied_by,omitempty"` // project ID of satisfying project
	EvidenceIDs    []string              `json:"evidence_ids,omitempty"`
	Visibility     VisibilityClass       `json:"visibility"`
	Publishable    bool                  `json:"publishable"`
	CreatedAt      time.Time             `json:"created_at"`
}

