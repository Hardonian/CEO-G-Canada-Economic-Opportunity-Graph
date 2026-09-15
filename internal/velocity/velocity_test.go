package velocity

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestCalculateCapitalVelocity_NilProject(t *testing.T) {
	now := time.Now()
	if got := CalculateCapitalVelocity(CapitalVelocityInputs{Project: nil}, now); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
}

func TestCalculateCapitalVelocity_ZeroInputs(t *testing.T) {
	now := time.Now()
	got := CalculateCapitalVelocity(CapitalVelocityInputs{
		Project: &domain.Project{ID: "p1", CapexCAD: 1_000_000_000},
	}, now)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Score != 0 {
		t.Fatalf("expected score 0 for no activity, got %.2f", got.Score)
	}
	if got.Type != VelocityCapital {
		t.Fatalf("wrong type: %s", got.Type)
	}
	if got.MethodologyVersion != MethodologyVersion {
		t.Fatalf("wrong methodology: %s", got.MethodologyVersion)
	}
}

func TestCalculateCapitalVelocity_RecentCommitments(t *testing.T) {
	now := time.Now()
	recent := now.AddDate(0, 0, -30)
	items := []*domain.CapitalItem{
		{ID: "c1", ProjectID: "p1", Status: domain.CapitalCommitted, AmountCAD: 500_000_000, CreatedAt: recent, EvidenceID: "e1"},
	}
	got := CalculateCapitalVelocity(CapitalVelocityInputs{
		Project:      &domain.Project{ID: "p1", CapexCAD: 1_000_000_000},
		CapitalItems: items,
	}, now)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Score <= 0 {
		t.Fatalf("expected positive score for recent commitment, got %.2f", got.Score)
	}
	if got.Factors["recent_commitments"] == 0 {
		t.Fatal("recent_commitments factor should be non-zero")
	}
}

func TestCalculateCapitalVelocity_GapClosure(t *testing.T) {
	now := time.Now()
	items := []*domain.CapitalItem{
		{ID: "c1", ProjectID: "p1", Status: domain.CapitalCommitted, AmountCAD: 800_000_000},
	}
	needs := []*domain.CapitalNeed{{Types: []domain.CapitalNeedType{domain.NeedEquity}}}
	got := CalculateCapitalVelocity(CapitalVelocityInputs{
		Project:      &domain.Project{ID: "p1", CapexCAD: 1_000_000_000},
		CapitalItems: items,
		CapitalNeeds: needs,
	}, now)
	if got == nil || got.Factors["gap_closure"] == 0 {
		t.Fatal("gap_closure factor should be non-zero for 80% committed")
	}
	if got.Factors["gap_closure"] < 70 || got.Factors["gap_closure"] > 90 {
		t.Fatalf("gap_closure expected ~80, got %.1f", got.Factors["gap_closure"])
	}
}

func TestCalculateCapitalVelocity_Trend(t *testing.T) {
	now := time.Now()
	recent := now.AddDate(0, 0, -30)
	items := []*domain.CapitalItem{
		{ID: "c1", ProjectID: "p1", Status: domain.CapitalCommitted, AmountCAD: 500_000_000, CreatedAt: recent},
	}
	events := []*domain.Event{
		{ID: "e1", ProjectID: "p1", EventType: "financing_announced", EventDate: recent, EvidenceID: "e1"},
	}
	got := CalculateCapitalVelocity(CapitalVelocityInputs{
		Project:      &domain.Project{ID: "p1", CapexCAD: 1_000_000_000},
		CapitalItems: items,
		Events:       events,
	}, now)
	if got.Trend == "" {
		t.Fatal("trend should be set")
	}
}

func TestCalculateExecutionVelocity_NilProject(t *testing.T) {
	now := time.Now()
	if got := CalculateExecutionVelocity(ExecutionVelocityInputs{Project: nil}, now); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
}

func TestCalculateExecutionVelocity_ZeroMilestones(t *testing.T) {
	now := time.Now()
	got := CalculateExecutionVelocity(ExecutionVelocityInputs{
		Project: &domain.Project{ID: "p1", CurrentStage: domain.StageFEED},
	}, now)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Score != 8 {
		t.Fatalf("expected score 8 for FEED stage (stage_advancement=40 * 0.20), got %.2f", got.Score)
	}
}

func TestCalculateExecutionVelocity_MilestoneScoring(t *testing.T) {
	now := time.Now()
	milestones := []*domain.Milestone{
		{ID: "m1", ProjectID: "p1", Type: domain.MilestoneEPCAwarded, Status: domain.MilestoneComplete, EvidenceIDs: []string{"e1"}},
		{ID: "m2", ProjectID: "p1", Type: domain.MilestoneConstructionStarted, Status: domain.MilestoneComplete, EvidenceIDs: []string{"e2"}},
		{ID: "m3", ProjectID: "p1", Type: domain.MilestoneCommercialOperation, Status: domain.MilestoneInProgress, EvidenceIDs: []string{"e3"}},
	}
	events := []*domain.Event{
		{ID: "e4", ProjectID: "p1", EventType: "regulatory_filing", EventDate: now.AddDate(0, 0, -60)},
	}
	got := CalculateExecutionVelocity(ExecutionVelocityInputs{
		Project:    &domain.Project{ID: "p1", CurrentStage: domain.StageConstruction},
		Milestones: milestones,
		Events:     events,
	}, now)
	if got == nil || got.Score == 0 {
		t.Fatalf("expected non-zero score, got %v", got)
	}
	if got.Factors["milestone_completion"] == 0 {
		t.Fatal("milestone_completion factor should be non-zero")
	}
	if got.Factors["execution_progress"] == 0 {
		t.Fatal("execution_progress factor should be non-zero")
	}
}

func TestCalculateExecutionVelocity_StageAdvancement(t *testing.T) {
	now := time.Now()
	got := CalculateExecutionVelocity(ExecutionVelocityInputs{
		Project: &domain.Project{ID: "p1", CurrentStage: domain.StageOperating},
	}, now)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Factors["stage_advancement"] != 100 {
		t.Fatalf("expected stage_advancement=100 for Operating, got %.1f", got.Factors["stage_advancement"])
	}
}

func TestCalculateExecutionVelocity_Trend(t *testing.T) {
	now := time.Now()
	milestones := []*domain.Milestone{
		{ID: "m1", ProjectID: "p1", Type: domain.MilestoneEPCAwarded, Status: domain.MilestoneComplete},
	}
	events := []*domain.Event{
		{ID: "e1", ProjectID: "p1", EventType: "stage_change", EventDate: now.AddDate(0, 0, -30)},
	}
	got := CalculateExecutionVelocity(ExecutionVelocityInputs{
		Project:    &domain.Project{ID: "p1", CurrentStage: domain.StageConstruction},
		Milestones: milestones,
		Events:     events,
	}, now)
	if got.Trend == "" {
		t.Fatal("trend should be set")
	}
}

func TestVelocityScore_Structure(t *testing.T) {
	now := time.Now()
	got := CalculateCapitalVelocity(CapitalVelocityInputs{
		Project: &domain.Project{ID: "p1", CapexCAD: 1_000_000_000},
		CapitalItems: []*domain.CapitalItem{
			{ID: "c1", ProjectID: "p1", Status: domain.CapitalCommitted, AmountCAD: 500_000_000, CreatedAt: now.AddDate(0, 0, -30)},
		},
	}, now)
	if got.InputHash == "" {
		t.Fatal("input_hash should be set")
	}
	if got.CalculatedAt.IsZero() {
		t.Fatal("calculated_at should be set")
	}
	if got.Factors == nil {
		t.Fatal("factors should not be nil")
	}
	if got.FactorEvidence == nil {
		t.Fatal("factor_evidence should not be nil")
	}
	if got.Explanation == nil {
		t.Fatal("explanation should not be nil")
	}
}
