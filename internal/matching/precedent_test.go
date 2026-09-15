package matching

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestFindDealPrecedents_EmptyInputs(t *testing.T) {
	if got := FindDealPrecedents(nil, nil, 5); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
	if got := FindDealPrecedents(&domain.Project{}, nil, 5); got != nil {
		t.Fatalf("expected nil for no profiles, got %v", got)
	}
	if got := FindDealPrecedents(&domain.Project{}, []*domain.InvestorProfile{{}}, 0); got != nil {
		t.Fatalf("expected nil for topN=0, got %v", got)
	}
}

func TestFindDealPrecedents_TopNLimits(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, Province: "ON", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}
	profile := &domain.InvestorProfile{
		EntityID: "e1",
		PublicDealHistory: []domain.DealPrecedent{
			{DealID: "d1", Sector: domain.SectorCleanEnergy, Province: "ON", AmountCAD: 1_000_000_000, Stage: domain.StageFEED, Year: 2024},
			{DealID: "d2", Sector: domain.SectorCleanEnergy, Province: "ON", AmountCAD: 500_000_000, Stage: domain.StageFEED, Year: 2024},
			{DealID: "d3", Sector: domain.SectorCleanEnergy, Province: "ON", AmountCAD: 250_000_000, Stage: domain.StageFEED, Year: 2024},
		},
	}
	got := FindDealPrecedents(project, []*domain.InvestorProfile{profile}, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 results for topN=2, got %d", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].SimilarityPct < got[i].SimilarityPct {
			t.Fatalf("results not sorted by similarity descending: %v", got)
		}
	}
}

func TestFindDealPrecedents_MatchesAllDimensions(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, Province: "ON", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}
	deal := domain.DealPrecedent{
		Sector:     domain.SectorCleanEnergy,
		Province:   "ON",
		Stage:      domain.StageFEED,
		AmountCAD:  1_000_000_000,
		Year:       2024,
	}
	profile := &domain.InvestorProfile{EntityID: "e1", PublicDealHistory: []domain.DealPrecedent{deal}}
	got := FindDealPrecedents(project, []*domain.InvestorProfile{profile}, 5)
	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if got[0].SimilarityPct != 100 {
		t.Fatalf("expected 100%% similarity for all-dimension match, got %.1f", got[0].SimilarityPct)
	}
	expected := []string{"sector", "province", "stage", "scale", "recent"}
	if len(got[0].MatchedOn) != len(expected) {
		t.Fatalf("expected %d matched dimensions, got %d: %v", len(expected), len(got[0].MatchedOn), got[0].MatchedOn)
	}
}

func TestFindDealPrecedents_LowSimilarityFiltered(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, Province: "ON", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}
	deal := domain.DealPrecedent{
		Sector:    domain.SectorMiningMetals,
		Province:  "BC",
		Stage:     domain.StageOperating,
		AmountCAD: 100,
		Year:      2020,
	}
	profile := &domain.InvestorProfile{EntityID: "e1", PublicDealHistory: []domain.DealPrecedent{deal}}
	got := FindDealPrecedents(project, []*domain.InvestorProfile{profile}, 5)
	if len(got) != 0 {
		t.Fatalf("expected 0 results for low similarity, got %d", len(got))
	}
}

func TestDealSimilarity_SectorWeight(t *testing.T) {
	project := &domain.Project{Sector: domain.SectorCleanEnergy}
	deal := domain.DealPrecedent{Sector: domain.SectorCleanEnergy}
	score, matched := dealSimilarity(project, deal)
	if score < 35 {
		t.Fatalf("sector match should contribute at least 35 points, got %.1f", score)
	}
	found := false
	for _, m := range matched {
		if m == "sector" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("sector should be in matched dimensions")
	}
}

func TestCalculateArchetypeFit_NilInputs(t *testing.T) {
	now := time.Now()
	if got := CalculateArchetypeFit(FitContext{}, domain.CounterpartyInfrastructureFund, now); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
}

func TestCalculateArchetypeFit_InfrastructureFund(t *testing.T) {
	now := time.Now()
	ctx := FitContext{
		Project: &domain.Project{
			ID:            "p1",
			Sector:        domain.SectorCleanEnergy,
			CurrentStage:  domain.StageConstruction,
			CapexCAD:      1_000_000_000,
		},
		CapitalNeeds: []*domain.CapitalNeed{{Types: []domain.CapitalNeedType{domain.NeedInfrastructureEquity}}},
	}
	got := CalculateArchetypeFit(ctx, domain.CounterpartyInfrastructureFund, now)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.ProjectID != "p1" {
		t.Fatalf("expected project ID p1, got %s", got.ProjectID)
	}
	if got.Archetype != domain.CounterpartyInfrastructureFund {
		t.Fatalf("wrong archetype: %s", got.Archetype)
	}
	if got.Score <= 0 || got.Score > 100 {
		t.Fatalf("score out of range: %.2f", got.Score)
	}
	if got.Methodology != FitMethodologyVersion {
		t.Fatalf("wrong methodology: %s", got.Methodology)
	}
	if got.CalculatedAt.IsZero() {
		t.Fatal("calculated_at should be set")
	}
	if _, ok := got.Factors["need_match"]; !ok {
		t.Fatal("need_match factor missing")
	}
	if _, ok := got.Factors["sector_fit"]; !ok {
		t.Fatal("sector_fit factor missing")
	}
	if _, ok := got.Factors["stage_fit"]; !ok {
		t.Fatal("stage_fit factor missing")
	}
	if _, ok := got.Factors["instrument_fit"]; !ok {
		t.Fatal("instrument_fit factor missing")
	}
	if _, ok := got.Factors["scale_fit"]; !ok {
		t.Fatal("scale_fit factor missing")
	}
}

func TestCalculateInvestorFit_NilInputs(t *testing.T) {
	now := time.Now()
	if got := CalculateInvestorFit(FitContext{}, nil, now); got != nil {
		t.Fatalf("expected nil for nil profile, got %v", got)
	}
	if got := CalculateInvestorFit(FitContext{Project: &domain.Project{ID: "p1"}}, nil, now); got != nil {
		t.Fatalf("expected nil for nil profile with project, got %v", got)
	}
}

func TestCalculateInvestorFit_ScoreRange(t *testing.T) {
	now := time.Now()
	ctx := FitContext{
		Project: &domain.Project{
			ID:            "p1",
			Sector:        domain.SectorCleanEnergy,
			CurrentStage:  domain.StageConstruction,
			CapexCAD:      1_000_000_000,
		},
		CapitalNeeds: []*domain.CapitalNeed{{Types: []domain.CapitalNeedType{domain.NeedInfrastructureEquity}}},
	}
	profile := &domain.InvestorProfile{
		EntityID:            "e1",
		InvestorTypes:       []domain.CounterpartyType{domain.CounterpartyInfrastructureFund},
		TargetSectors:       []domain.Sector{domain.SectorCleanEnergy},
		TargetGeographies:   []string{"ON"},
		MinTicketCAD:        100_000_000,
		MaxTicketCAD:        2_000_000_000,
		PreferredStages:     []domain.LifecycleStage{domain.StageConstruction, domain.StageOperating},
	}
	got := CalculateInvestorFit(ctx, profile, now)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Score < 0 || got.Score > 100 {
		t.Fatalf("score out of range: %.2f", got.Score)
	}
	if got.EntityID != "e1" {
		t.Fatalf("wrong entity ID: %s", got.EntityID)
	}
}

func TestCalculateInvestorFit_TicketSizeExceedsMax(t *testing.T) {
	now := time.Now()
	ctx := FitContext{
		Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 5_000_000_000},
	}
	profile := &domain.InvestorProfile{EntityID: "e1", MaxTicketCAD: 2_000_000_000}
	got := CalculateInvestorFit(ctx, profile, now)
	if got == nil || got.Score != 0 {
		t.Fatalf("expected score 0 when project exceeds max ticket, got %v", got)
	}
}

func TestCalculateInvestorFit_AllFactorsPresent(t *testing.T) {
	now := time.Now()
	ctx := FitContext{
		Project: &domain.Project{ID: "p1", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED, CapexCAD: 500_000_000},
	}
	profile := &domain.InvestorProfile{
		EntityID:          "e1",
		InvestorTypes:     []domain.CounterpartyType{domain.CounterpartyPrivateEquity},
		TargetSectors:     []domain.Sector{domain.SectorCleanEnergy},
		TargetGeographies: []string{"ON"},
		MinTicketCAD:      100_000_000,
		MaxTicketCAD:      1_000_000_000,
		PreferredStages:   []domain.LifecycleStage{domain.StageFeasibility, domain.StagePreFEED, domain.StageFEED},
	}
	got := CalculateInvestorFit(ctx, profile, now)
	required := []string{"type_alignment", "sector_alignment", "geography_alignment", "ticket_alignment", "stage_alignment"}
	for _, f := range required {
		if _, ok := got.Factors[f]; !ok {
			t.Fatalf("missing factor %s", f)
		}
	}
}
