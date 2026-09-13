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
