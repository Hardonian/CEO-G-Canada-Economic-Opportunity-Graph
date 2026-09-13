package capitalstack

import (
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestCanonicalProgramsNotEmpty(t *testing.T) {
	programs := CanonicalPrograms()
	if len(programs) == 0 {
		t.Fatal("CanonicalPrograms() returned no programs")
	}
}

func TestCanonicalProgramsCoversAllSectors(t *testing.T) {
	programs := CanonicalPrograms()
	seen := make(map[domain.Sector]bool)
	for _, p := range programs {
		for _, s := range p.SectorEligibility {
			seen[s] = true
		}
	}
	expected := []domain.Sector{
		domain.SectorCriticalMinerals,
		domain.SectorNuclearEnergy,
		domain.SectorCleanEnergy,
		domain.SectorAICompute,
		domain.SectorTransportation,
		domain.SectorIndustrialMfg,
		domain.SectorHousingEnabling,
		domain.SectorEnergyFuels,
		domain.SectorForestryBioeconomy,
	}
	for _, s := range expected {
		if !seen[s] {
			t.Errorf("sector %q not covered by any program", s)
		}
	}
}

func TestCanonicalProgramsHasFederalAndProvincial(t *testing.T) {
	programs := CanonicalPrograms()
	federal := 0
	provincial := 0
	for _, p := range programs {
		if p.Jurisdiction == "Federal" {
			federal++
		} else {
			provincial++
		}
	}
	if federal == 0 {
		t.Error("expected federal programs")
	}
	if provincial == 0 {
		t.Error("expected provincial programs")
	}
}

func TestEvaluateProjectStackCleanEnergy(t *testing.T) {
	project := &domain.Project{
		ID:           "test-clean-energy",
		Name:         "Test Solar Farm",
		Sector:       domain.SectorCleanEnergy,
		Province:     "AB",
		CurrentStage: domain.StageConstruction,
		CapexCAD:     100_000_000,
	}
	result := EvaluateProjectStack(project)
	if len(result.MatchedPrograms) == 0 {
		t.Fatal("expected at least one matched program")
	}
	foundCleanTechITC := false
	for _, m := range result.MatchedPrograms {
		if m.Program.ID == "clean_tech_itc" {
			foundCleanTechITC = true
			if m.Classification != MatchLikely {
				t.Errorf("clean_tech_itc classification = %q, want %q (project is in construction stage)", m.Classification, MatchLikely)
			}
			expectedValue := int64(float64(100_000_000) * 0.30)
			if m.EstimatedValueCAD != expectedValue {
				t.Errorf("estimated value = %d, want %d", m.EstimatedValueCAD, expectedValue)
			}
		}
	}
	if !foundCleanTechITC {
		t.Error("expected clean_tech_itc in matched programs")
	}
}

func TestEvaluateProjectStackMutualExclusivity(t *testing.T) {
	project := &domain.Project{
		ID:           "test-mutual-exclusion",
		Name:         "Clean Hydrogen Plant",
		Sector:       domain.SectorCleanEnergy,
		Province:     "Federal",
		CurrentStage: domain.StageFID,
		CapexCAD:     500_000_000,
	}
	result := EvaluateProjectStack(project)
	hasConflict := false
	for _, conflict := range result.StackingConflicts {
		if conflict != "" {
			hasConflict = true
		}
	}
	if !hasConflict {
		t.Error("expected stacking conflict for mutually exclusive programs (clean_tech_itc + clean_hydrogen_itc)")
	}
}

func TestEvaluateProjectStackHighCapex(t *testing.T) {
	project := &domain.Project{
		ID:           "test-high-capex",
		Name:         "Major Infrastructure",
		Sector:       domain.SectorHousingEnabling,
		Province:     "ON",
		CurrentStage: domain.StageConstruction,
		CapexCAD:     500_000_000,
	}
	result := EvaluateProjectStack(project)
	if result.EffectiveFundingPct <= 0 {
		t.Errorf("effective_funding_pct = %f, expected positive value", result.EffectiveFundingPct)
	}
	if result.TotalPotentialCAD <= 0 {
		t.Error("expected non-zero total potential")
	}
	if result.Disclaimer == "" {
		t.Error("expected disclaimer to be set")
	}
}

func TestEvaluateProjectStackNoSectorMatch(t *testing.T) {
	project := &domain.Project{
		ID:           "test-no-match",
		Name:         "Unrelated Project",
		Sector:       "Nonexistent Sector",
		Province:     "XX",
		CurrentStage: domain.StageAnnounced,
		CapexCAD:     10_000_000,
	}
	result := EvaluateProjectStack(project)
	if len(result.MatchedPrograms) != 0 {
		t.Errorf("expected 0 matched programs for unmapped sector, got %d", len(result.MatchedPrograms))
	}
}

func TestProgramMatchLaborRequirement(t *testing.T) {
	programs := CanonicalPrograms()
	for _, p := range programs {
		if p == nil {
			t.Fatal("nil program in CanonicalPrograms")
		}
		if p.ID == "" {
			t.Error("program has empty ID")
		}
		if p.Name == "" {
			t.Errorf("program %s has empty name", p.ID)
		}
		if p.Administrator == "" {
			t.Errorf("program %s has empty administrator", p.ID)
		}
		if p.StatutoryReference == "" {
			t.Errorf("program %s has empty statutory reference", p.ID)
		}
		if p.Jurisdiction == "" {
			t.Errorf("program %s has empty jurisdiction", p.ID)
		}
		if p.MaxSupportRatePct <= 0 || p.MaxSupportRatePct > 100 {
			t.Errorf("program %s has unrealistic support rate %.1f", p.ID, p.MaxSupportRatePct)
		}
	}
}

func TestEvaluateProjectStackCapexZero(t *testing.T) {
	project := &domain.Project{
		ID:           "test-zero-capex",
		Name:         "Zero Capex",
		Sector:       domain.SectorCleanEnergy,
		Province:     "BC",
		CurrentStage: domain.StageAnnounced,
		CapexCAD:     0,
	}
	result := EvaluateProjectStack(project)
	if result.EffectiveFundingPct != 0 {
		t.Errorf("effective_funding_pct = %f, want 0 for zero capex", result.EffectiveFundingPct)
	}
}
