package matching

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestCalculateArchetypeFit_InfrastructureFundScoring(t *testing.T) {
	now := time.Now()
	t.Run("infra sector gets high sector fit", func(t *testing.T) {
		ctx := FitContext{
			Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000},
			CapitalNeeds: []*domain.CapitalNeed{{Types: []domain.CapitalNeedType{domain.NeedInfrastructureEquity}}},
		}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyInfrastructureFund, now)
		if got.Factors["sector_fit"] != 80 {
			t.Fatalf("expected sector_fit=80 for infra sector, got %.1f", got.Factors["sector_fit"])
		}
	})
	t.Run("non-infra sector gets low sector fit", func(t *testing.T) {
		ctx := FitContext{
			Project: &domain.Project{ID: "p1", Sector: domain.SectorHousingEnabling, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000},
		}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyInfrastructureFund, now)
		if got.Factors["sector_fit"] != 30 {
			t.Fatalf("expected sector_fit=30 for non-infra sector, got %.1f", got.Factors["sector_fit"])
		}
	})
	t.Run("ECA gets fixed score", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyECA, now)
		if got.Factors["sector_fit"] != 60 {
			t.Fatalf("expected sector_fit=60 for ECA, got %.1f", got.Factors["sector_fit"])
		}
	})
}

func TestCalculateArchetypeFit_StageFit(t *testing.T) {
	now := time.Now()
	t.Run("PE prefers development stage", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyPrivateEquity, now)
		if got.Factors["stage_fit"] != 90 {
			t.Fatalf("expected stage_fit=90 for FEED with PE, got %.1f", got.Factors["stage_fit"])
		}
	})
	t.Run("Infra fund prefers construction", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageConstruction, CapexCAD: 1_000_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyInfrastructureFund, now)
		if got.Factors["stage_fit"] != 90 {
			t.Fatalf("expected stage_fit=90 for Construction with Infra, got %.1f", got.Factors["stage_fit"])
		}
	})
	t.Run("Bank prefers post-FID", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFID, CapexCAD: 1_000_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyBank, now)
		if got.Factors["stage_fit"] != 90 {
			t.Fatalf("expected stage_fit=90 for FID with Bank, got %.1f", got.Factors["stage_fit"])
		}
	})
}

func TestCalculateArchetypeFit_InstrumentFit(t *testing.T) {
	now := time.Now()
	t.Run("Bank with debt need", func(t *testing.T) {
		ctx := FitContext{
			Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000},
			CapitalNeeds: []*domain.CapitalNeed{{Types: []domain.CapitalNeedType{domain.NeedSeniorDebt}}},
		}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyBank, now)
		if got.Factors["instrument_fit"] != 90 {
			t.Fatalf("expected instrument_fit=90 for senior debt with bank, got %.1f", got.Factors["instrument_fit"])
		}
	})
	t.Run("Unknown archetype gets default", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyGovernment, now)
		if got.Factors["instrument_fit"] != 50 {
			t.Fatalf("expected instrument_fit=50 for unknown archetype, got %.1f", got.Factors["instrument_fit"])
		}
	})
}

func TestCalculateArchetypeFit_ScaleFit(t *testing.T) {
	now := time.Now()
	t.Run("Infra fund large scale", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyInfrastructureFund, now)
		if got.Factors["scale_fit"] != 90 {
			t.Fatalf("expected scale_fit=90 for $1B with infra fund, got %.1f", got.Factors["scale_fit"])
		}
	})
	t.Run("PE sweet spot", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 500_000_000}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyPrivateEquity, now)
		if got.Factors["scale_fit"] != 80 {
			t.Fatalf("expected scale_fit=80 for $500M with PE, got %.1f", got.Factors["scale_fit"])
		}
	})
	t.Run("Zero capex unknown scale", func(t *testing.T) {
		ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 0}}
		got := CalculateArchetypeFit(ctx, domain.CounterpartyInfrastructureFund, now)
		if got.Factors["scale_fit"] != 30 {
			t.Fatalf("expected scale_fit=30 for zero capex, got %.1f", got.Factors["scale_fit"])
		}
	})
}

func TestCalculateInvestorFit_TypeAlignment(t *testing.T) {
	now := time.Now()
	ctx := FitContext{
		Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000},
		CapitalNeeds: []*domain.CapitalNeed{{Types: []domain.CapitalNeedType{domain.NeedInfrastructureEquity}}},
	}
	profile := &domain.InvestorProfile{
		EntityID:      "e1",
		InvestorTypes: []domain.CounterpartyType{domain.CounterpartyInfrastructureFund, domain.CounterpartyPrivateEquity},
	}
	got := CalculateInvestorFit(ctx, profile, now)
	if got.Factors["type_alignment"] == 0 {
		t.Fatal("type_alignment should be non-zero when profile has matching archetype")
	}
}

func TestCalculateInvestorFit_GeographyAlignment(t *testing.T) {
	now := time.Now()
	ctx := FitContext{Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, Province: "ON"}}
	t.Run("matching geography", func(t *testing.T) {
		profile := &domain.InvestorProfile{EntityID: "e1", TargetGeographies: []string{"ON"}}
		got := CalculateInvestorFit(ctx, profile, now)
		if got.Factors["geography_alignment"] != 100 {
			t.Fatalf("expected geography_alignment=100, got %.1f", got.Factors["geography_alignment"])
		}
	})
	t.Run("non-matching geography", func(t *testing.T) {
		profile := &domain.InvestorProfile{EntityID: "e1", TargetGeographies: []string{"BC"}}
		got := CalculateInvestorFit(ctx, profile, now)
		if got.Factors["geography_alignment"] != 0 {
			t.Fatalf("expected geography_alignment=0, got %.1f", got.Factors["geography_alignment"])
		}
	})
}
