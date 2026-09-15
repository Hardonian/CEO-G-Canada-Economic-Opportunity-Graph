package domain

// StrategicTheme classifies projects according to Canada's national economic
// strategy and priority investment themes. A project may carry multiple themes.
type StrategicTheme string

const (
	ThemeAIInfrastructure       StrategicTheme = "AI_INFRASTRUCTURE"
	ThemeArctic                 StrategicTheme = "ARCTIC"
	ThemeDefence                StrategicTheme = "DEFENCE"
	ThemeCriticalMinerals       StrategicTheme = "CRITICAL_MINERALS"
	ThemeEnergySecurity         StrategicTheme = "ENERGY_SECURITY"
	ThemeTradeDiversification   StrategicTheme = "TRADE_DIVERSIFICATION"
	ThemeSovereignCompute       StrategicTheme = "SOVEREIGN_COMPUTE"
	ThemeNuclear                StrategicTheme = "NUCLEAR"
	ThemeIndigenousOwnership    StrategicTheme = "INDIGENOUS_OWNERSHIP"
	ThemeBatterySupplyChain     StrategicTheme = "BATTERY_SUPPLY_CHAIN"
	ThemeSemiconductors         StrategicTheme = "SEMICONDUCTORS"
	ThemePorts                  StrategicTheme = "PORTS"
	ThemeGridModernization      StrategicTheme = "GRID_MODERNIZATION"
	ThemeAdvancedManufacturing  StrategicTheme = "ADVANCED_MANUFACTURING"
	ThemeCleanHydrogen          StrategicTheme = "CLEAN_HYDROGEN"
	ThemeCarbonCapture          StrategicTheme = "CARBON_CAPTURE"
	ThemeFoodSecurity           StrategicTheme = "FOOD_SECURITY"
	ThemeBioeconomy             StrategicTheme = "BIOECONOMY"
	ThemeHousing                StrategicTheme = "HOUSING"
	ThemeTransportCorridor      StrategicTheme = "TRANSPORT_CORRIDOR"
)

// ValidStrategicThemes returns all known strategic themes.
var ValidStrategicThemes = []StrategicTheme{
	ThemeAIInfrastructure, ThemeArctic, ThemeDefence, ThemeCriticalMinerals,
	ThemeEnergySecurity, ThemeTradeDiversification, ThemeSovereignCompute,
	ThemeNuclear, ThemeIndigenousOwnership, ThemeBatterySupplyChain,
	ThemeSemiconductors, ThemePorts, ThemeGridModernization,
	ThemeAdvancedManufacturing, ThemeCleanHydrogen, ThemeCarbonCapture,
	ThemeFoodSecurity, ThemeBioeconomy, ThemeHousing, ThemeTransportCorridor,
}

// TechnologyMaturity classifies the commercial readiness of a project's core
// technology. Values loosely map to Technology Readiness Levels but use the
// investment-oriented vocabulary that capital allocators actually use.
type TechnologyMaturity string

const (
	TechMaturityResearch      TechnologyMaturity = "RESEARCH"       // TRL 1-3
	TechMaturityDevelopment   TechnologyMaturity = "DEVELOPMENT"    // TRL 4-5
	TechMaturityPilot         TechnologyMaturity = "PILOT"          // TRL 6
	TechMaturityDemonstration TechnologyMaturity = "DEMONSTRATION"  // TRL 7
	TechMaturityFOAK          TechnologyMaturity = "FOAK"           // First-of-a-kind
	TechMaturityNOAK          TechnologyMaturity = "NOAK"           // Nth-of-a-kind
	TechMaturityCommercial    TechnologyMaturity = "COMMERCIAL"     // TRL 9, proven at scale
	TechMaturityProven        TechnologyMaturity = "PROVEN"         // Widespread, de-risked
)

// ProjectFinanceType classifies the capital deployment mode.
type ProjectFinanceType string

const (
	FinanceTypeGreenfield     ProjectFinanceType = "GREENFIELD"
	FinanceTypeBrownfield     ProjectFinanceType = "BROWNFIELD"
	FinanceTypeExpansion      ProjectFinanceType = "EXPANSION"
	FinanceTypeRestart        ProjectFinanceType = "RESTART"
	FinanceTypeRedevelopment  ProjectFinanceType = "REDEVELOPMENT"
	FinanceTypeRepowering     ProjectFinanceType = "REPOWERING"
	FinanceTypeTechScaleUp    ProjectFinanceType = "TECH_SCALE_UP"
	FinanceTypeAcquisition    ProjectFinanceType = "ACQUISITION"
)

// RevenueModel classifies the expected revenue structure which directly
// impacts financability and the set of eligible capital instruments.
type RevenueModel string

const (
	RevenueOfftakeBacked     RevenueModel = "OFFTAKE_BACKED"
	RevenueRegulated         RevenueModel = "REGULATED"
	RevenueMerchant          RevenueModel = "MERCHANT"
	RevenueConcession        RevenueModel = "CONCESSION"
	RevenueTolling           RevenueModel = "TOLLING"
	RevenueAvailabilityPay   RevenueModel = "AVAILABILITY_PAYMENT"
	RevenueHybrid            RevenueModel = "HYBRID"
	RevenueGovernmentFunded  RevenueModel = "GOVERNMENT_FUNDED"
	RevenueSubscription      RevenueModel = "SUBSCRIPTION"
	RevenueUnknown           RevenueModel = "UNKNOWN"
)
