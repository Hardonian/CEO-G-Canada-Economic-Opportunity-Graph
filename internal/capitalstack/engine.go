package capitalstack

import (
	"fmt"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// MatchClassification indicates structured qualification confidence.
type MatchClassification string

const (
	MatchLikely   MatchClassification = "likely_match"
	MatchPossible MatchClassification = "possible_match"
	MatchUnlikely MatchClassification = "unlikely_match"
	MatchReview   MatchClassification = "requires_review"
)

// CapitalProgram defines a Canadian federal or provincial funding or tax program.
type CapitalProgram struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Administrator       string              `json:"administrator"` // CRA, NRCan, CIB, CGF, ISED, provincial
	ProgramType         string              `json:"program_type"` // refundable_tax_credit, concessionary_loan, grant, equity
	SectorEligibility   []domain.Sector     `json:"sector_eligibility"`
	MaxSupportRatePct   float64             `json:"max_support_rate_pct"` // e.g. 30% for Clean Tech ITC
	LaborConditionsReq  bool                `json:"labor_conditions_req"` // Prevailing wage and apprenticeship rules
	MutuallyExclusive   []string            `json:"mutually_exclusive"`   // Program IDs that cannot be claimed on identical capital assets
	StackingCapPct      float64             `json:"stacking_cap_pct"`     // Total government assistance ceiling
	Summary             string              `json:"summary"`
	StatutoryReference  string              `json:"statutory_reference"`
	Jurisdiction        string              `json:"jurisdiction"` // Federal, ON, QC, BC, AB
}

// StackingEvaluation evaluates stacking feasibility and net eligible incentives.
type StackingEvaluation struct {
	ProjectID          string              `json:"project_id"`
	ProjectCapexCAD    int64               `json:"project_capex_cad"`
	MatchedPrograms    []*ProgramMatch     `json:"matched_programs"`
	TotalPotentialCAD  int64               `json:"total_potential_cad"`
	StackingConflicts  []string            `json:"stacking_conflicts"`
	EffectiveFundingPct float64            `json:"effective_funding_pct"`
	Disclaimer         string              `json:"disclaimer"`
}

// ProgramMatch details program eligibility for a project.
type ProgramMatch struct {
	Program            *CapitalProgram     `json:"program"`
	Classification     MatchClassification `json:"classification"`
	EstimatedValueCAD  int64               `json:"estimated_value_cad"`
	LaborRequirement   string              `json:"labor_requirement"`
	Rationale          string              `json:"rationale"`
}

// CanonicalPrograms returns major Canadian capital stack programs.
func CanonicalPrograms() []*CapitalProgram {
	return []*CapitalProgram{
		// FEDERAL PROGRAMS
		{
			ID:                 "clean_tech_itc",
			Name:               "Clean Technology Investment Tax Credit (Clean Tech ITC)",
			Administrator:      "Canada Revenue Agency (CRA) / Finance Canada",
			ProgramType:        "refundable_tax_credit",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorNuclearEnergy},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: true,
			MutuallyExclusive:  []string{"clean_hydrogen_itc", "ccus_itc", "clean_electricity_itc"},
			StackingCapPct:     100.0,
			Summary:            "Refundable 30% tax credit on capital cost of eligible clean energy generation, storage, and zero-emission industrial equipment.",
			StatutoryReference: "Income Tax Act, Section 127.45",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "clean_electricity_itc",
			Name:               "Clean Electricity Investment Tax Credit",
			Administrator:      "Canada Revenue Agency (CRA) / NRCan",
			ProgramType:        "refundable_tax_credit",
			SectorEligibility:  []domain.Sector{domain.SectorNuclearEnergy, domain.SectorCleanEnergy},
			MaxSupportRatePct:  15.0,
			LaborConditionsReq: true,
			MutuallyExclusive:  []string{"clean_tech_itc"},
			StackingCapPct:     100.0,
			Summary:            "Refundable 15% credit for publicly owned utilities, Crown corporations, and private proponents investing in clean generation, SMRs, and inter-provincial transmission.",
			StatutoryReference: "Federal Budget 2023 / Income Tax Act",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "clean_hydrogen_itc",
			Name:               "Clean Hydrogen Investment Tax Credit",
			Administrator:      "Canada Revenue Agency (CRA) / Finance Canada",
			ProgramType:        "refundable_tax_credit",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorIndustrial},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: true,
			MutuallyExclusive:  []string{"clean_tech_itc", "ccus_itc"},
			StackingCapPct:     100.0,
			Summary:            "Refundable 15-40% tax credit for clean hydrogen production projects based on carbon intensity.",
			StatutoryReference: "Income Tax Act, Section 127.46",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "ccus_itc",
			Name:               "Carbon Capture, Utilization, and Storage Investment Tax Credit",
			Administrator:      "Canada Revenue Agency (CRA) / Finance Canada",
			ProgramType:        "refundable_tax_credit",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorEnergyFuels, domain.SectorCleanEnergy},
			MaxSupportRatePct:  60.0,
			LaborConditionsReq: true,
			MutuallyExclusive:  []string{"clean_tech_itc", "clean_hydrogen_itc"},
			StackingCapPct:     100.0,
			Summary:            "Refundable tax credit for eligible CCUS equipment: 60% for capture, 50% for transport, 37.5% for storage/use.",
			StatutoryReference: "Income Tax Act, Section 127.44",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "critical_minerals_itc",
			Name:               "Critical Mineral Exploration & Processing Tax Credit (CMITC)",
			Administrator:      "Canada Revenue Agency (CRA)",
			ProgramType:        "refundable_tax_credit",
			SectorEligibility:  []domain.Sector{domain.SectorCriticalMinerals},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "30% investment tax credit for eligible critical mineral exploration and processing assets targeting Canada's 34 prioritized minerals.",
			StatutoryReference: "Income Tax Act, Section 127(9)",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "cib_clean_power",
			Name:               "Canada Infrastructure Bank (CIB) Clean Power Concessionary Financing",
			Administrator:      "Canada Infrastructure Bank",
			ProgramType:        "concessionary_loan",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorNuclearEnergy, domain.SectorTransportation},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: true,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Patient, below-market long-term debt and Indigenous equity participation loans for large-scale transmission, renewables, and district energy.",
			StatutoryReference: "Canada Infrastructure Bank Act",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "cgf_carbon_contracts",
			Name:               "Canada Growth Fund (CGF) Carbon Contracts for Difference",
			Administrator:      "Public Sector Pension Investment Board (PSP) / CGF Inc.",
			ProgramType:        "concessionary_loan",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorCleanEnergy},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Direct equity investment, off-take guarantees, and carbon price certainty contracts (CCfDs) de-risking private industrial decarbonization.",
			StatutoryReference: "Budget Implementation Act 2023",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "sif_net_zero",
			Name:               "Strategic Innovation Fund (SIF) - Net Zero Accelerator",
			Administrator:      "Innovation, Science and Economic Development Canada (ISED)",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorAICompute, domain.SectorCriticalMinerals},
			MaxSupportRatePct:  25.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Federal grants and conditionally repayable contributions for mega-scale industrial decarbonization, battery manufacturing, and critical mineral facilities.",
			StatutoryReference: "ISED SIF Guidelines",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "ira_transit",
			Name:               "Investing in Canada Infrastructure Program - Public Transit",
			Administrator:      "Infrastructure Canada",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorTransportation, domain.SectorHousingInfra},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     100.0,
			Summary:            "Federal cost-sharing for public transit infrastructure including zero-emission buses, rail, and active transportation.",
			StatutoryReference: "Investing in Canada Plan",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "ira_green",
			Name:               "Investing in Canada Infrastructure Program - Green Infrastructure",
			Administrator:      "Infrastructure Canada",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorHousingInfra, domain.SectorForestryBioeconomy},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     100.0,
			Summary:            "Federal cost-sharing for climate adaptation, disaster mitigation, and environmental quality infrastructure.",
			StatutoryReference: "Investing in Canada Plan",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "bdc_cleantech",
			Name:               "BDC Cleantech & Sustainability Financing",
			Administrator:      "Business Development Bank of Canada",
			ProgramType:        "concessionary_loan",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorIndustrial, domain.SectorCriticalMinerals},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     85.0,
			Summary:            "Flexible financing for cleantech companies including venture debt, growth capital, and sustainability-linked loans.",
			StatutoryReference: "BDC Act",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "edc_cleantech",
			Name:               "Export Development Canada (EDC) Cleantech Financing",
			Administrator:      "Export Development Canada",
			ProgramType:        "concessionary_loan",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorCriticalMinerals, domain.SectorIndustrial},
			MaxSupportRatePct:  35.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     85.0,
			Summary:            "Financing and insurance solutions for Canadian cleantech exporters and international project developers.",
			StatutoryReference: "Export Development Act",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "indigenous_loan_guarantee",
			Name:               "Indigenous Loan Guarantee Program",
			Administrator:      "NRCan / CIB / Finance Canada",
			ProgramType:        "loan_guarantee",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorNuclearEnergy, domain.SectorCriticalMinerals, domain.SectorTransportation},
			MaxSupportRatePct:  80.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     100.0,
			Summary:            "Federal loan guarantees enabling Indigenous equity ownership in major resource and infrastructure projects.",
			StatutoryReference: "Budget 2024 / NRCan Mandate",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "cdpq_infra",
			Name:               "CDPQ Infra / Caisse de dépôt et placement du Québec Infrastructure",
			Administrator:      "Caisse de dépôt et placement du Québec",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorTransportation, domain.SectorAICompute, domain.SectorNuclearEnergy},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Major institutional investor in Quebec infrastructure; direct equity and co-investment in large-scale projects.",
			StatutoryReference: "Act respecting the Caisse de dépôt et placement du Québec",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "cpib_infra",
			Name:               "CPP Investments Infrastructure",
			Administrator:      "CPP Investments",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorTransportation, domain.SectorNuclearEnergy, domain.SectorAICompute},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "National pension fund infrastructure equity investments; long-term capital for essential infrastructure.",
			StatutoryReference: "Canada Pension Plan Investment Board Act",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "psp_infra",
			Name:               "PSP Investments Infrastructure",
			Administrator:      "Public Sector Pension Investment Board",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorTransportation, domain.SectorIndustrial},
			MaxSupportRatePct:  35.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Pension fund infrastructure platform investing in essential assets across Canada and globally.",
			StatutoryReference: "Public Sector Pension Investment Board Act",
			Jurisdiction:       "Federal",
		},
		{
			ID:                 "omers_infra",
			Name:               "OMERS Infrastructure",
			Administrator:      "OMERS",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorTransportation, domain.SectorNuclearEnergy},
			MaxSupportRatePct:  35.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Municipal pension fund infrastructure investments; active in Canadian energy transition and transport.",
			StatutoryReference: "OMERS Act",
			Jurisdiction:       "Federal",
		},
		
		// ALBERTA PROVINCIAL PROGRAMS
		{
			ID:                 "era_emissions_reduction",
			Name:               "Emissions Reduction Alberta (ERA) - Technology Innovation Funding",
			Administrator:      "Emissions Reduction Alberta",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorIndustrial, domain.SectorEnergyFuels, domain.SectorCriticalMinerals},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Grant funding for pilot, demonstration, and first-of-kind deployment of emissions reduction technologies in Alberta.",
			StatutoryReference: "Climate Change and Emissions Management Act",
			Jurisdiction:       "AB",
		},
		{
			ID:                 "era_industrial_efficiency",
			Name:               "ERA - Industrial Energy Efficiency Program",
			Administrator:      "Emissions Reduction Alberta",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorManufacturing, domain.SectorCriticalMinerals},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Funding for energy efficiency retrofits, process optimization, and waste heat recovery in Alberta industrial facilities.",
			StatutoryReference: "Climate Change and Emissions Management Act",
			Jurisdiction:       "AB",
		},
		{
			ID:                 "alberta_petrochemical_incentive",
			Name:               "Alberta Petrochemical Incentive Program (APIP)",
			Administrator:      "Alberta Ministry of Energy and Minerals",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorEnergyFuels, domain.SectorIndustrial},
			MaxSupportRatePct:  12.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "12% grant on eligible capital costs for new or expanded petrochemical manufacturing facilities in Alberta.",
			StatutoryReference: "Alberta Petrochemicals Incentive Act",
			Jurisdiction:       "AB",
		},
		{
			ID:                 "invest_alberta",
			Name:               "Invest Alberta - Strategic Investment Fund",
			Administrator:      "Invest Alberta Corporation",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCriticalMinerals, domain.SectorCleanEnergy, domain.SectorAICompute, domain.SectorIndustrial},
			MaxSupportRatePct:  20.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Direct equity investments and strategic partnerships for major projects advancing Alberta's economic diversification.",
			StatutoryReference: "Invest Alberta Corporation Act",
			Jurisdiction:       "AB",
		},
		{
			ID:                 "alberta_innovation_voucher",
			Name:               "Alberta Innovates - Innovation Voucher Program",
			Administrator:      "Alberta Innovates",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorIndustrial, domain.SectorCriticalMinerals, domain.SectorAICompute},
			MaxSupportRatePct:  75.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Vouchers up to $75k for Alberta SMEs to access research expertise and facilities for technology development.",
			StatutoryReference: "Alberta Research and Innovation Act",
			Jurisdiction:       "AB",
		},
		{
			ID:                 "alberta_carbon_trunk_line",
			Name:               "Alberta Carbon Trunk Line (ACTL) Integration",
			Administrator:      "Alberta Ministry of Energy and Minerals / Enhance Energy",
			ProgramType:        "concessionary_loan",
			SectorEligibility:  []domain.Sector{domain.SectorEnergyFuels, domain.SectorIndustrial, domain.SectorCleanEnergy},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Access to the world's largest CO2 pipeline for CCUS projects; enables carbon capture integration for industrial emitters.",
			StatutoryReference: "Alberta Carbon Competitiveness Incentive Regulation",
			Jurisdiction:       "AB",
		},
		
		// QUEBEC PROVINCIAL PROGRAMS
		{
			ID:                 "investissement_quebec_credit",
			Name:               "Investissement Québec - Tax Credit for Investments",
			Administrator:      "Investissement Québec / Revenu Québec",
			ProgramType:        "refundable_tax_credit",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorCleanEnergy, domain.SectorCriticalMinerals, domain.SectorAICompute},
			MaxSupportRatePct:  20.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Refundable tax credit of up to 20% for eligible capital investments in manufacturing and processing in Quebec.",
			StatutoryReference: "Taxation Act (Quebec), Section 1029.8.36.0.3.38",
			Jurisdiction:       "QC",
		},
		{
			ID:                 "investissement_quebec_green",
			Name:               "Investissement Québec - Green Fund / Fonds vert",
			Administrator:      "Investissement Québec / Ministère de l'Environnement",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorTransportation, domain.SectorForestryBioeconomy, domain.SectorIndustrial},
			MaxSupportRatePct:  40.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Funding from Quebec's Green Fund for GHG reduction, climate adaptation, and circular economy projects.",
			StatutoryReference: "Act respecting the Ministère du Développement durable, de l'Environnement et des Parcs",
			Jurisdiction:       "QC",
		},
		{
			ID:                 "iq_innovation",
			Name:               "Investissement Québec - Innovation Program",
			Administrator:      "Investissement Québec",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorAICompute, domain.SectorCriticalMinerals, domain.SectorCleanEnergy, domain.SectorIndustrial},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Financial assistance for R&D, innovation, and technology commercialization projects in Quebec.",
			StatutoryReference: "Act respecting Investissement Québec",
			Jurisdiction:       "QC",
		},
		{
			ID:                 "hydro_quebec_partnerships",
			Name:               "Hydro-Québec Partnerships & Procurement",
			Administrator:      "Hydro-Québec",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorNuclearEnergy, domain.SectorTransportation},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: true,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Equity partnerships and long-term power purchase agreements for renewable generation, storage, and transmission in Quebec.",
			StatutoryReference: "Hydro-Québec Act",
			Jurisdiction:       "QC",
		},
		{
			ID:                 "quebec_critical_minerals",
			Name:               "Québec Critical and Strategic Minerals Strategy",
			Administrator:      "Ministère de l'Énergie et des Ressources naturelles",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCriticalMinerals},
			MaxSupportRatePct:  35.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Financial support for exploration, processing, and recycling of critical minerals in Quebec; part of 2023-2025 action plan.",
			StatutoryReference: "Québec Critical and Strategic Minerals Strategy 2023-2025",
			Jurisdiction:       "QC",
		},
		{
			ID:                 "quebec_battery_chain",
			Name:               "Québec Battery Value Chain Program",
			Administrator:      "Investissement Québec / Propulsion Québec",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCriticalMinerals, domain.SectorCleanEnergy, domain.SectorIndustrial},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Support for battery materials, cell manufacturing, and recycling projects across the Quebec battery value chain.",
			StatutoryReference: "Québec Battery Value Chain Strategy",
			Jurisdiction:       "QC",
		},
		
		// BRITISH COLUMBIA PROVINCIAL PROGRAMS
		{
			ID:                 "inbc_investment",
			Name:               "InBC Investment Corp. - Strategic Investments",
			Administrator:      "InBC Investment Corp.",
			ProgramType:        "equity",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorCriticalMinerals, domain.SectorAICompute, domain.SectorIndustrial, domain.SectorForestryBioeconomy},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Provincial strategic investment fund providing patient capital for scalable BC companies in clean tech, critical minerals, and advanced manufacturing.",
			StatutoryReference: "InBC Investment Corp. Act",
			Jurisdiction:       "BC",
		},
		{
			ID:                 "bc_cleanbc_industry",
			Name:               "CleanBC Industry Fund",
			Administrator:      "BC Ministry of Environment and Climate Change Strategy",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorCleanEnergy, domain.SectorCriticalMinerals, domain.SectorEnergyFuels},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Funding for large-scale industrial emissions reduction projects in BC; supports electrification, fuel switching, and CCUS.",
			StatutoryReference: "Climate Change Accountability Act",
			Jurisdiction:       "BC",
		},
		{
			ID:                 "bc_cleanbc_communities",
			Name:               "CleanBC Communities Fund",
			Administrator:      "BC Ministry of Municipal Affairs",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCleanEnergy, domain.SectorTransportation, domain.SectorHousingInfra, domain.SectorForestryBioeconomy},
			MaxSupportRatePct:  73.3,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Federal-provincial cost-sharing (up to 73.3%) for community infrastructure reducing GHG emissions.",
			StatutoryReference: "CleanBC Plan / Investing in Canada Infrastructure Program",
			Jurisdiction:       "BC",
		},
		{
			ID:                 "bc_hydro_electrification",
			Name:               "BC Hydro - Electrification Programs",
			Administrator:      "BC Hydro",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorIndustrial, domain.SectorTransportation, domain.SectorCleanEnergy, domain.SectorHousingInfra},
			MaxSupportRatePct:  100.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     80.0,
			Summary:            "Incentives for industrial and commercial electrification, heat pumps, EV charging, and energy efficiency in BC Hydro service territory.",
			StatutoryReference: "Hydro and Power Authority Act / CleanBC",
			Jurisdiction:       "BC",
		},
		{
			ID:                 "bc_critical_minerals",
			Name:               "BC Critical Minerals Strategy Funding",
			Administrator:      "BC Ministry of Energy, Mines and Low Carbon Innovation",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorCriticalMinerals},
			MaxSupportRatePct:  30.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Funding for critical mineral exploration, processing, and supply chain development under BC's Critical Minerals Strategy.",
			StatutoryReference: "BC Critical Minerals Strategy 2024",
			Jurisdiction:       "BC",
		},
		{
			ID:                 "bc_forest_innovation",
			Name:               "BC Forest Innovation Investment",
			Administrator:      "Forestry Innovation Investment Ltd.",
			ProgramType:        "grant",
			SectorEligibility:  []domain.Sector{domain.SectorForestryBioeconomy, domain.SectorCleanEnergy},
			MaxSupportRatePct:  50.0,
			LaborConditionsReq: false,
			MutuallyExclusive:  []string{},
			StackingCapPct:     75.0,
			Summary:            "Support for advanced wood products, mass timber, biofuels, and bioproducts manufacturing in British Columbia.",
			StatutoryReference: "Forestry Innovation Investment Act",
			Jurisdiction:       "BC",
		},
	}
}

// EvaluateProjectStack analyzes a project's eligibility and models program interaction constraints.
func EvaluateProjectStack(p *domain.Project) *StackingEvaluation {
	programs := CanonicalPrograms()
	var matches []*ProgramMatch
	var totalEst int64
	var conflicts []string

	claimedCategories := make(map[string]string) // program ID -> category

	for _, prog := range programs {
		// Check sector eligibility
		sectorMatch := false
		for _, s := range prog.SectorEligibility {
			if s == p.Sector {
				sectorMatch = true
				break
			}
		}

		if !sectorMatch {
			continue
		}

		classification := MatchPossible
		rationale := fmt.Sprintf("Project sector (%s) aligns with statutory requirements.", p.Sector)

		if p.CurrentStage == domain.StageFID || p.CurrentStage == domain.StageConstruction || p.CurrentStage == domain.StagePermitting {
			classification = MatchLikely
			rationale += " Project is in mature procurement or construction stage with quantifiable capital deployment."
		}

		// Calculate gross incentive
		gross := int64(float64(p.CapexCAD) * (prog.MaxSupportRatePct / 100.0))

		// Check mutual exclusivity with existing matched programs
		for _, prev := range matches {
			for _, mut := range prog.MutuallyExclusive {
				if prev.Program.ID == mut {
					conflictMsg := fmt.Sprintf("Stacking constraint: %s cannot be claimed alongside %s on identical capital assets.", prog.Name, prev.Program.Name)
					conflicts = append(conflicts, conflictMsg)
					classification = MatchReview
					rationale += fmt.Sprintf(" Warning: Potential mutual exclusivity conflict with %s.", prev.Program.Name)
				}
			}
		}

		laborReq := "Standard compliance"
		if prog.LaborConditionsReq {
			laborReq = "Mandatory: Prevailing wage rates and minimum 10% apprentice labor hours required to claim top-tier rate."
		}

		match := &ProgramMatch{
			Program:           prog,
			Classification:    classification,
			EstimatedValueCAD: gross,
			LaborRequirement:  laborReq,
			Rationale:         rationale,
		}

		matches = append(matches, match)
		claimedCategories[prog.ID] = prog.ProgramType
		totalEst += gross
	}

	effectiveFundingPct := 0.0
	if p.CapexCAD > 0 {
		effectiveFundingPct = (float64(totalEst) / float64(p.CapexCAD)) * 100.0
	}

	return &StackingEvaluation{
		ProjectID:           p.ID,
		ProjectCapexCAD:     p.CapexCAD,
		MatchedPrograms:     matches,
		TotalPotentialCAD:   totalEst,
		StackingConflicts:   conflicts,
		EffectiveFundingPct: effectiveFundingPct,
		Disclaimer:          "LEGAL NOTICE: This analysis is provided for informational and preliminary planning purposes only. It does not constitute legal, financial, or tax advice. Program rules, stacking limits, and statutory eligibility must be formally confirmed with the Canada Revenue Agency or the relevant program administrator.",
	}
}
