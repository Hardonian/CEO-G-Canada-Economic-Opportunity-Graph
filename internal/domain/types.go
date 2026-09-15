package domain

import "strings"

// SourceTier represents the authority level of a data source.
type SourceTier int

const (
	SourceTier1 SourceTier = 1 // Primary / Authoritative (statutory registries, gazettes)
	SourceTier2 SourceTier = 2 // Corporate Issuers & Regulatory Filings (SEDAR+, TSX)
	SourceTier3 SourceTier = 3 // Credible News Organizations (Globe and Mail, Bloomberg, Reuters)
	SourceTier4 SourceTier = 4 // Secondary Industry Research (trade associations, consultancies)
)

// ConfidenceLevel represents the verification status of a claim.
type ConfidenceLevel string

const (
	ConfidenceVerified   ConfidenceLevel = "VERIFIED"   // Corroborated by Tier 1 authoritative publisher
	ConfidenceSupported  ConfidenceLevel = "SUPPORTED"  // Corroborated by Tier 2 or multiple Tier 3 sources
	ConfidenceReported   ConfidenceLevel = "REPORTED"   // Stated by a single source without independent verification
	ConfidenceInferred   ConfidenceLevel = "INFERRED"   // Derived algorithmically via dependency ontologies
	ConfidenceUnknown    ConfidenceLevel = "UNKNOWN"    // Data point is unobserved or not disclosed
	ConfidenceConflict   ConfidenceLevel = "CONFLICTED" // Multiple active sources assert irreconcilable facts
	ConfidenceStale      ConfidenceLevel = "STALE"      // Information has exceeded its review horizon
	ConfidenceRetracted  ConfidenceLevel = "RETRACTED"  // Source publication or claim formally withdrawn
	// ConfidenceConflicted is retained as a compatibility alias for older
	// callers and models.
	ConfidenceConflicted = ConfidenceConflict
)

// Sector represents a Canadian economic infrastructure sector.
type Sector string

const (
	SectorCriticalMinerals     Sector = "Critical Minerals"
	SectorNuclearEnergy        Sector = "Nuclear & Clean Power"
	SectorCleanEnergy          Sector = "Clean Energy & Grid"
	SectorAICompute            Sector = "AI Compute & Data Centres"
	SectorDefenceArctic        Sector = "Defence & Arctic"
	SectorTransportation       Sector = "Transportation & Ports"
	SectorIndustrialMfg        Sector = "Industrial & Manufacturing"
	SectorHousingEnabling      Sector = "Housing-Enabling Infrastructure"
	SectorMiningMetals         Sector = "Mining & Metals"
	SectorEnergyFuels          Sector = "Energy & Fuels"
	SectorForestryBioeconomy   Sector = "Forestry & Bioeconomy"
	// Compatibility aliases for earlier domain versions.
	SectorHousingInfra = SectorHousingEnabling
	SectorManufacturing = SectorIndustrialMfg
)

// LifecycleStage represents the current stage of a capital project.
type LifecycleStage string

const (
	StageUnknown           LifecycleStage = "UNKNOWN"
	StageDiscovered        LifecycleStage = "DISCOVERED"
	StageConcept           LifecycleStage = "CONCEPT"
	StagePreDevelopment    LifecycleStage = "PRE_DEVELOPMENT"
	StageAnnounced         LifecycleStage = "ANNOUNCED"
	StageReferred          LifecycleStage = "REFERRED"
	StageEarlyDevelopment  LifecycleStage = "EARLY_DEVELOPMENT"
	StageFeasibility       LifecycleStage = "FEASIBILITY"
	StagePreFEED           LifecycleStage = "PRE_FEED"
	StageFEED              LifecycleStage = "FEED"
	StageDetailedEngineering LifecycleStage = "DETAILED_ENGINEERING"
	StageFinancing         LifecycleStage = "FINANCING"
	StageEnvironmentalReview LifecycleStage = "ENVIRONMENTAL_REVIEW"
	StagePermitting        LifecycleStage = "PERMITTING"
	StageProcurement       LifecycleStage = "PROCUREMENT"
	StageFIDLikely         LifecycleStage = "FID_LIKELY"
	StageFID               LifecycleStage = "FID"
	StageConstructionReady LifecycleStage = "CONSTRUCTION_READY"
	StageConstruction      LifecycleStage = "CONSTRUCTION"
	StageCommissioning     LifecycleStage = "COMMISSIONING"
	StageOperating         LifecycleStage = "OPERATING"
	StageExpansion         LifecycleStage = "EXPANSION"
	StageDelayed           LifecycleStage = "DELAYED"
	StagePaused            LifecycleStage = "PAUSED"
	StageCancelled         LifecycleStage = "CANCELLED"
)

// ValidLifecycleStages returns all valid lifecycle stages for validation.
var ValidLifecycleStages = []LifecycleStage{
	StageUnknown, StageDiscovered, StageConcept, StagePreDevelopment, StageAnnounced, StageReferred,
	StageEarlyDevelopment, StageFeasibility, StagePreFEED, StageFEED, StageDetailedEngineering, StageFinancing,
	StageEnvironmentalReview, StagePermitting, StageProcurement,
	StageFIDLikely, StageFID, StageConstructionReady, StageConstruction, StageCommissioning,
	StageOperating, StageExpansion, StageDelayed, StagePaused, StageCancelled,
}

// CapitalCategory represents the type of capital event.
type CapitalCategory string

const (
	CapitalCategoryCIB           CapitalCategory = "CIB"           // Canada Infrastructure Bank
	CapitalCategoryCGF           CapitalCategory = "CGF"           // Canada Growth Fund
	CapitalCategoryEquity        CapitalCategory = "EQUITY"
	CapitalCategoryDebt          CapitalCategory = "DEBT"
	CapitalCategoryGrant         CapitalCategory = "GRANT"
	CapitalCategoryLoan          CapitalCategory = "LOAN"
	CapitalCategoryTaxCredit     CapitalCategory = "TAX_CREDIT"
	CapitalCategoryProcurement   CapitalCategory = "PROCUREMENT"
	CapitalCategoryPrivateEquity CapitalCategory = "PRIVATE_EQUITY"
	CapitalCategoryLoanGuarantee CapitalCategory = "LOAN_GUARANTEE"
)

// CapitalStatus represents the status of a capital item.
type CapitalStatus string

const (
	CapitalRumoured               CapitalStatus = "RUMOURED"
	CapitalSeeking                CapitalStatus = "SEEKING"
	CapitalProposed               CapitalStatus = "PROPOSED"
	CapitalAnnounced              CapitalStatus = "ANNOUNCED"
	CapitalCommitted              CapitalStatus = "COMMITTED"
	CapitalConditionallyCommitted CapitalStatus = "CONDITIONALLY_COMMITTED"
	CapitalClosed                 CapitalStatus = "CLOSED"
	CapitalSigned                 CapitalStatus = "SIGNED"
	CapitalDisbursed              CapitalStatus = "DISBURSED"
	CapitalWithdrawn              CapitalStatus = "WITHDRAWN"
	CapitalCancelled              CapitalStatus = "CANCELLED"
)

// RequirementClass represents the confidence class of a procurement or opportunity.
type RequirementClass string

const (
	RequirementConfirmed   RequirementClass = "CONFIRMED"
	RequirementDerived     RequirementClass = "DERIVED"
	RequirementSpeculative RequirementClass = "SPECULATIVE"
)

// SignalType represents the type of economic signal.
type SignalType string

const (
	SignalConstructionSignal       SignalType = "CONSTRUCTION_SIGNAL"
	SignalRegulatoryProgress       SignalType = "REGULATORY_PROGRESS"
	SignalIndigenousPartnership    SignalType = "INDIGENOUS_PARTNERSHIP"
	SignalFinancingAcceleration    SignalType = "FINANCING_ACCELERATION"
	SignalProcurementAcceleration  SignalType = "PROCUREMENT_ACCELERATION"
	SignalTimelineSlip             SignalType = "TIMELINE_SLIP"
	SignalProjectDelay             SignalType = "PROJECT_DELAY"
	SignalPoliticalSupportLoss     SignalType = "POLITICAL_SUPPORT_LOSS"
	SignalCapexIncrease            SignalType = "CAPEX_INCREASE"
)

// IntelligenceStatus represents the overall data quality status.
type IntelligenceStatus string

const (
	StatusHealthy     IntelligenceStatus = "HEALTHY"
	StatusPartial     IntelligenceStatus = "PARTIAL"
	StatusDegraded    IntelligenceStatus = "DEGRADED"
	StatusStale       IntelligenceStatus = "STALE"
	StatusUnavailable IntelligenceStatus = "UNAVAILABLE"
)

// NormalizeLookupName produces a case-insensitive, whitespace-normalized
// key for entity and project name indexes.
func NormalizeLookupName(value string) string {
	if value == "" {
		return ""
	}
	b := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c >= 'A' && c <= 'Z' {
			b = append(b, c+'a'-'A')
		} else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if len(b) > 0 && b[len(b)-1] != ' ' {
				b = append(b, ' ')
			}
		} else {
			b = append(b, c)
		}
	}
	return strings.TrimSpace(string(b))
}
