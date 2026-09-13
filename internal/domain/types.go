package domain

// LifecycleStage represents the official lifecycle progression of a Canadian major project.
type LifecycleStage string

const (
	StageUnknown             LifecycleStage = "UNKNOWN"
	StageDiscovered          LifecycleStage = "DISCOVERED"
	StageAnnounced           LifecycleStage = "ANNOUNCED"
	StageReferred            LifecycleStage = "REFERRED"
	StageEarlyDevelopment    LifecycleStage = "EARLY_DEVELOPMENT"
	StageFeasibility         LifecycleStage = "FEASIBILITY"
	StageFinancing           LifecycleStage = "FINANCING"
	StageEnvironmentalReview LifecycleStage = "ENVIRONMENTAL_REVIEW"
	StagePermitting          LifecycleStage = "PERMITTING"
	StageProcurement         LifecycleStage = "PROCUREMENT"
	StageFIDLikely           LifecycleStage = "FID_LIKELY"
	StageFID                 LifecycleStage = "FID"
	StageConstruction        LifecycleStage = "CONSTRUCTION"
	StageCommissioning       LifecycleStage = "COMMISSIONING"
	StageOperating           LifecycleStage = "OPERATING"
	StageDelayed             LifecycleStage = "DELAYED"
	StagePaused              LifecycleStage = "PAUSED"
	StageCancelled           LifecycleStage = "CANCELLED"
)

// ValidLifecycleStages returns all supported stages in chronological expectation.
func ValidLifecycleStages() []LifecycleStage {
	return []LifecycleStage{
		StageUnknown,
		StageDiscovered,
		StageAnnounced,
		StageReferred,
		StageEarlyDevelopment,
		StageFeasibility,
		StageFinancing,
		StageEnvironmentalReview,
		StagePermitting,
		StageProcurement,
		StageFIDLikely,
		StageFID,
		StageConstruction,
		StageCommissioning,
		StageOperating,
		StageDelayed,
		StagePaused,
		StageCancelled,
	}
}

// ConfidenceLevel defines factual verifiability.
type ConfidenceLevel string

const (
	ConfidenceVerified   ConfidenceLevel = "VERIFIED"
	ConfidenceSupported  ConfidenceLevel = "SUPPORTED"
	ConfidenceReported   ConfidenceLevel = "REPORTED"
	ConfidenceInferred   ConfidenceLevel = "INFERRED"
	ConfidenceConflicted ConfidenceLevel = "CONFLICTED"
	ConfidenceUnknown    ConfidenceLevel = "UNKNOWN"
	ConfidenceStale      ConfidenceLevel = "STALE"
	ConfidenceRetracted  ConfidenceLevel = "RETRACTED"
)

// IntelligenceStatus makes degraded and partial results explicit at API and
// product boundaries.
type IntelligenceStatus string

const (
	StatusHealthy     IntelligenceStatus = "HEALTHY"
	StatusStale       IntelligenceStatus = "STALE"
	StatusPartial     IntelligenceStatus = "PARTIAL"
	StatusDegraded    IntelligenceStatus = "DEGRADED"
	StatusUnavailable IntelligenceStatus = "UNAVAILABLE"
)

// SourceTier classifies publisher reliability.
type SourceTier int

const (
	SourceTier1 SourceTier = 1 // Primary/Authoritative: Government, Regulators (IAAC, CER, CanadaBuys)
	SourceTier2 SourceTier = 2 // Corporate Issuers: SEDAR+, official annual reports, regulatory filings
	SourceTier3 SourceTier = 3 // Credible News: Bloomberg, Globe and Mail, Reuters, CBC
	SourceTier4 SourceTier = 4 // Secondary Industry Research
)

// Sector defines key Canadian strategic sectors.
type Sector string

const (
	SectorCriticalMinerals   Sector = "Critical Minerals"
	SectorNuclearEnergy      Sector = "Nuclear & Clean Power"
	SectorCleanEnergy        Sector = "Clean Energy & Grid"
	SectorAICompute          Sector = "AI Compute & Data Centres"
	SectorDefenceArctic      Sector = "Defence & Arctic"
	SectorTransportation     Sector = "Transportation & Ports"
	SectorIndustrial         Sector = "Industrial & Manufacturing"
	SectorHousingInfra       Sector = "Housing-Enabling Infrastructure"
	SectorMiningMetals       Sector = "Mining & Metals"
	SectorEnergyFuels        Sector = "Energy & Fuels"
	SectorForestryBioeconomy Sector = "Forestry & Bioeconomy"
)

// RequirementClass differentiates inferred dependencies from confirmed procurement.
type RequirementClass string

const (
	RequirementConfirmed   RequirementClass = "CONFIRMED"
	RequirementDerived     RequirementClass = "DERIVED"
	RequirementSpeculative RequirementClass = "SPECULATIVE"
)

// CapitalStatus distinguishes capital progression.
type CapitalStatus string

const (
	CapitalAnnounced              CapitalStatus = "announced"
	CapitalProposed               CapitalStatus = "proposed"
	CapitalConditionallyCommitted CapitalStatus = "conditionally_committed"
	CapitalCommitted              CapitalStatus = "committed"
	CapitalClosed                 CapitalStatus = "closed"
	CapitalDisbursed              CapitalStatus = "disbursed"
	CapitalEstimated              CapitalStatus = "estimated"
)

// CapitalCategory specifies the nature of capital.
type CapitalCategory string

const (
	CapitalCategoryEquity             CapitalCategory = "equity"
	CapitalCategoryDebt               CapitalCategory = "debt"
	CapitalCategoryGrant              CapitalCategory = "grant"
	CapitalCategoryTaxIncentive       CapitalCategory = "tax_incentive"
	CapitalCategoryLoanGuarantee      CapitalCategory = "loan_guarantee"
	CapitalCategoryCIB                CapitalCategory = "cib"
	CapitalCategoryCGF                CapitalCategory = "cgf"
	CapitalCategoryIndigenousLoanGuar CapitalCategory = "indigenous_loan_guarantee"
	CapitalCategoryPension            CapitalCategory = "pension"
	CapitalCategoryStrategicCorporate CapitalCategory = "strategic_corporate"
)

// SignalType indicates momentum and inflection events.
type SignalType string

const (
	SignalCapitalAcceleration     SignalType = "CAPITAL_ACCELERATION"
	SignalRegulatoryProgress      SignalType = "REGULATORY_PROGRESS"
	SignalProcurementAcceleration SignalType = "PROCUREMENT_ACCELERATION"
	SignalFinancingAcceleration   SignalType = "FINANCING_ACCELERATION"
	SignalConstructionSignal      SignalType = "CONSTRUCTION_SIGNAL"
	SignalPoliticalSupportGain    SignalType = "POLITICAL_SUPPORT_GAIN"
	SignalPoliticalSupportLoss    SignalType = "POLITICAL_SUPPORT_LOSS"
	SignalTimelineSlip            SignalType = "TIMELINE_SLIP"
	SignalCapexIncrease           SignalType = "CAPEX_INCREASE"
	SignalCapexDecrease           SignalType = "CAPEX_DECREASE"
	SignalProjectDelay            SignalType = "PROJECT_DELAY"
	SignalIndigenousPartnership   SignalType = "INDIGENOUS_PARTNERSHIP_SIGNAL"
	SignalOfftakeSecured          SignalType = "OFFTAKE_SIGNAL"
	SignalSupplierContract        SignalType = "SUPPLIER_SIGNAL"
)
