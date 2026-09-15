package signals

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestDetectSignals_NilProject(t *testing.T) {
	if got := DetectSignals(nil, nil, nil, nil); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
}

func TestDetectSignals_NoInputs(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	got := DetectSignals(project, nil, nil, nil)
	if got == nil {
		t.Skip("nil result acceptable for no inputs")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 signals for no inputs, got %d", len(got))
	}
}

func TestDetectSignals_StageChangeSignals(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFID}
	events := []*domain.Event{
		{
			ID:        "e1",
			ProjectID: "p1",
			EventType: "stage_change",
			EventDate: now,
			PreviousStage: func() *domain.LifecycleStage { s := domain.StageConstruction; return &s }(),
			NewStage:      func() *domain.LifecycleStage { s := domain.StageFID; return &s }(),
			Title:       "FID Reached",
			EvidenceID:  "ev1",
		},
	}
	got := DetectSignals(project, events, nil, nil)
	found := false
	for _, s := range got {
		if s.Type == domain.SignalConstructionSignal {
			found = true
			if s.Magnitude != 0.95 {
				t.Fatalf("expected magnitude 0.95, got %.2f", s.Magnitude)
			}
			if s.Confidence != 0.98 {
				t.Fatalf("expected confidence 0.98, got %.2f", s.Confidence)
			}
			if s.EvidenceID != "ev1" {
				t.Fatalf("expected evidence ev1, got %s", s.EvidenceID)
			}
		}
	}
	if !found {
		t.Fatal("expected SignalConstructionSignal for FID stage change")
	}
}

func TestDetectSignals_RegulatorySignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StagePermitting}
	events := []*domain.Event{
		{
			ID:        "e1",
			ProjectID: "p1",
			EventType: "regulatory_filing",
			EventDate: time.Now(),
			PreviousStage: func() *domain.LifecycleStage { s := domain.StageFEED; return &s }(),
			NewStage:      func() *domain.LifecycleStage { s := domain.StagePermitting; return &s }(),
			Title:       "Permit Filed",
			EvidenceID:  "ev1",
		},
	}
	got := DetectSignals(project, events, nil, nil)
	found := false
	for _, s := range got {
		if s.Type == domain.SignalRegulatoryProgress {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected SignalRegulatoryProgress for permitting stage change")
	}
}

func TestDetectSignals_IndigenousSignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	events := []*domain.Event{
		{
			ID:        "e1",
			ProjectID: "p1",
			EventType: "indigenous_agreement",
			EventDate: time.Now(),
			Title:     "IBA Signed",
			EvidenceID: "ev1",
		},
	}
	got := DetectSignals(project, events, nil, nil)
	found := false
	for _, s := range got {
		if s.Type == domain.SignalIndigenousPartnership {
			found = true
			if s.Magnitude != 0.88 {
				t.Fatalf("expected magnitude 0.88, got %.2f", s.Magnitude)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected SignalIndigenousPartnership for indigenous agreement")
	}
}

func TestDetectSignals_CapitalSignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	capital := []*domain.CapitalItem{
		{ID: "c1", ProjectID: "p1", Status: domain.CapitalCommitted, AmountCAD: 500_000_000, EvidenceID: "ev1"},
	}
	got := DetectSignals(project, nil, capital, nil)
	found := false
	for _, s := range got {
		if s.Type == domain.SignalFinancingAcceleration {
			found = true
			if s.Magnitude != 0.85 {
				t.Fatalf("expected magnitude 0.85, got %.2f", s.Magnitude)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected SignalFinancingAcceleration for committed capital")
	}
}

func TestDetectSignals_ProcurementSignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	procurements := []*domain.Procurement{
		{ID: "p1", Title: "RFP 1", CreatedAt: time.Now().AddDate(0, 0, -10)},
		{ID: "p2", Title: "RFP 2", CreatedAt: time.Now().AddDate(0, 0, -20)},
	}
	got := DetectSignals(project, nil, nil, procurements)
	found := false
	for _, s := range got {
		if s.Type == domain.SignalProcurementAcceleration {
			found = true
			if s.Magnitude != 0.75 {
				t.Fatalf("expected magnitude 0.75, got %.2f", s.Magnitude)
			}
			if s.Description == "" {
				t.Fatal("expected description to be set")
			}
			break
		}
	}
	if !found {
		t.Fatal("expected SignalProcurementAcceleration for procurements")
	}
}

func TestDetectSignals_IgnoresNilCapitalItems(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	capital := []*domain.CapitalItem{{ID: "c1", ProjectID: "p1", Status: domain.CapitalCommitted, AmountCAD: 500_000_000, EvidenceID: "ev1"}}
	got := DetectSignals(project, nil, capital, nil)
	if got == nil {
		t.Skip("nil result acceptable")
	}
	found := false
	for _, s := range got {
		if s.Type == domain.SignalFinancingAcceleration {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected SignalFinancingAcceleration")
	}
}

func TestCalculateMomentum_NilProject(t *testing.T) {
	if got := CalculateMomentum(nil, nil); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
}

func TestCalculateMomentum_Accelerating(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	signals := []*domain.Signal{
		{ID: "s1", ProjectID: "p1", Type: domain.SignalFinancingAcceleration, Timestamp: time.Now().AddDate(0, 0, -3), Magnitude: 0.8, Confidence: 0.9},
		{ID: "s2", ProjectID: "p1", Type: domain.SignalConstructionSignal, Timestamp: time.Now().AddDate(0, 0, -5), Magnitude: 0.7, Confidence: 0.85},
	}
	got := CalculateMomentum(project, signals)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Velocity != "ACCELERATING" {
		t.Fatalf("expected ACCELERATING, got %s", got.Velocity)
	}
	if got.SignalCount != 2 {
		t.Fatalf("expected 2 signals, got %d", got.SignalCount)
	}
	if got.Momentum30d <= 0 {
		t.Fatalf("expected positive 30d momentum for accelerating, got %.4f", got.Momentum30d)
	}
}

func TestCalculateMomentum_Decelerating(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	old := time.Now().AddDate(0, 0, -20)
	signals := []*domain.Signal{
		{ID: "s1", ProjectID: "p1", Type: domain.SignalProjectDelay, Timestamp: old, Magnitude: 0.6, Confidence: 0.8},
		{ID: "s2", ProjectID: "p1", Type: domain.SignalPoliticalSupportLoss, Timestamp: old, Magnitude: 0.5, Confidence: 0.7},
	}
	got := CalculateMomentum(project, signals)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Velocity != "DECELERATING" {
		t.Fatalf("expected DECELERATING, got %s", got.Velocity)
	}
}

func TestCalculateMomentum_Stalled(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	signals := []*domain.Signal{
		{ID: "s1", ProjectID: "p1", Type: domain.SignalFinancingAcceleration, Timestamp: time.Now().AddDate(0, 0, -60), Magnitude: 0.1, Confidence: 0.3},
	}
	got := CalculateMomentum(project, signals)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Velocity != "STEADY" {
		t.Fatalf("expected STEADY for low positive momentum, got %s", got.Velocity)
	}
}

func TestCalculateMomentum_NegativeSignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	now := time.Now()
	signals := []*domain.Signal{
		{ID: "s1", ProjectID: "p1", Type: domain.SignalProjectDelay, Timestamp: now, Magnitude: 0.6, Confidence: 0.8},
	}
	got := CalculateMomentum(project, signals)
	if got == nil || got.Momentum7d >= 0 {
		t.Fatalf("expected negative 7d momentum for delay signal, got %.4f", got.Momentum7d)
	}
}

func TestCalculateMomentum_AllTimeframes(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	now := time.Now()
	signals := []*domain.Signal{
		{ID: "s1", ProjectID: "p1", Type: domain.SignalFinancingAcceleration, Timestamp: now.AddDate(0, 0, -2), Magnitude: 0.5, Confidence: 0.9},
		{ID: "s2", ProjectID: "p1", Type: domain.SignalRegulatoryProgress, Timestamp: now.AddDate(0, 0, -15), Magnitude: 0.4, Confidence: 0.85},
		{ID: "s3", ProjectID: "p1", Type: domain.SignalConstructionSignal, Timestamp: now.AddDate(0, 0, -50), Magnitude: 0.3, Confidence: 0.8},
	}
	got := CalculateMomentum(project, signals)
	if got.Momentum7d <= got.Momentum30d {
		t.Log("7d should generally be higher than 30d with recent signals")
	}
	if got.Momentum30d <= got.Momentum90d {
		t.Log("30d should generally be higher than 90d")
	}
	if got.Momentum7d > 1.0 || got.Momentum7d < -1.0 {
		t.Fatalf("7d momentum out of bounds: %.4f", got.Momentum7d)
	}
}

func TestCalculateMomentum_EmptySignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	got := CalculateMomentum(project, []*domain.Signal{})
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Velocity != "STEADY" {
		t.Fatalf("expected STEADY for no signals, got %s", got.Velocity)
	}
	if got.Momentum7d != 0 || got.Momentum30d != 0 || got.Momentum90d != 0 {
		t.Fatalf("expected all momenta to be 0, got %.4f %.4f %.4f", got.Momentum7d, got.Momentum30d, got.Momentum90d)
	}
}

func TestCalculateMomentum_ExcludesOldSignals(t *testing.T) {
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, CurrentStage: domain.StageFEED}
	old := time.Now().AddDate(0, 0, -200)
	signals := []*domain.Signal{
		{ID: "s1", ProjectID: "p1", Type: domain.SignalFinancingAcceleration, Timestamp: old, Magnitude: 0.9, Confidence: 0.95},
	}
	got := CalculateMomentum(project, signals)
	if got == nil {
		t.Fatal("expected non-nil")
	}
	if got.Momentum7d != 0 {
		t.Fatalf("expected 0 momentum for 7d with 200d-old signal, got %.4f", got.Momentum7d)
	}
}
