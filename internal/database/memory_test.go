package database

import (
	"context"
	"errors"
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestSaveProjectMergesComplementarySourceFacts(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	first := &domain.Project{
		ID: "project-1", Name: "Source inventory project",
		Latitude: 48.4, Longitude: -89.2,
		CapexCAD: 500_000_000, CapexStatus: domain.ConfidenceReported,
		EvidenceIDs: []string{"evidence-inventory"},
		ExternalIDs: map[string]string{"nrcan_mpi": "1234"},
		Metadata:    map[string]interface{}{"dataset_vintage": "2025-2035"},
	}
	if err := store.SaveProject(ctx, first); err != nil {
		t.Fatal(err)
	}
	curated := &domain.Project{
		ID: "project-1", Name: "Reviewed canonical project",
		CapexStatus: domain.ConfidenceUnknown,
		EvidenceIDs: []string{"evidence-curated"},
		ExternalIDs: map[string]string{"iaac_registry": "9876"},
		Metadata:    map[string]interface{}{"review_status": "reviewed"},
	}
	if err := store.SaveProject(ctx, curated); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetProject(ctx, "project-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Reviewed canonical project" {
		t.Fatalf("canonical name = %q", got.Name)
	}
	if got.Latitude != 48.4 || got.Longitude != -89.2 {
		t.Fatalf("coordinates were erased: %f,%f", got.Latitude, got.Longitude)
	}
	if got.CapexCAD != 500_000_000 || got.CapexStatus != domain.ConfidenceReported {
		t.Fatalf("reported capex was erased: %d %s", got.CapexCAD, got.CapexStatus)
	}
	if len(got.EvidenceIDs) != 2 || got.ExternalIDs["nrcan_mpi"] != "1234" || got.ExternalIDs["iaac_registry"] != "9876" {
		t.Fatalf("source lineage was not merged: %#v", got)
	}
	if got.Metadata["dataset_vintage"] != "2025-2035" || got.Metadata["review_status"] != "reviewed" {
		t.Fatalf("metadata was not merged: %#v", got.Metadata)
	}
}

func TestSaveProjectRejectsAmbiguousSlug(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	if err := store.SaveProject(ctx, &domain.Project{ID: "project-1", Slug: "same"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveProject(ctx, &domain.Project{ID: "project-2", Slug: "same"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("ambiguous slug error = %v, want ErrConflict", err)
	}
}

func TestSaveEventRequiresExistingProjectAndEvidence(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()
	event := &domain.Event{ID: "event-1", ProjectID: "project-1", EvidenceID: "evidence-1"}
	if err := store.SaveEvent(ctx, event); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing project error = %v, want ErrNotFound", err)
	}
	if err := store.SaveProject(ctx, &domain.Project{ID: "project-1", Slug: "project"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvent(ctx, event); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing evidence error = %v, want ErrNotFound", err)
	}
	if err := store.SaveEvidence(ctx, &domain.Evidence{ID: "evidence-1"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
}

// ─── secondary index consistency ──────────────────────────────────────────────

func TestEventsByProjectIndex_Consistency(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	store.SaveProject(ctx, &domain.Project{ID: "proj-A", Slug: "proj-a"})
	store.SaveProject(ctx, &domain.Project{ID: "proj-B", Slug: "proj-b"})
	store.SaveEvidence(ctx, &domain.Evidence{ID: "ev-1"})
	store.SaveEvidence(ctx, &domain.Evidence{ID: "ev-2"})
	store.SaveEvidence(ctx, &domain.Evidence{ID: "ev-3"})

	store.SaveEvent(ctx, &domain.Event{ID: "event-1", ProjectID: "proj-A", EvidenceID: "ev-1", Title: "Alpha"})
	store.SaveEvent(ctx, &domain.Event{ID: "event-2", ProjectID: "proj-A", EvidenceID: "ev-2", Title: "Beta"})
	store.SaveEvent(ctx, &domain.Event{ID: "event-3", ProjectID: "proj-B", EvidenceID: "ev-3", Title: "Gamma"})

	eventsA, err := store.ListEventsByProject(ctx, "proj-A")
	if err != nil {
		t.Fatal(err)
	}
	if len(eventsA) != 2 {
		t.Errorf("proj-A events: got %d, want 2", len(eventsA))
	}

	eventsB, err := store.ListEventsByProject(ctx, "proj-B")
	if err != nil {
		t.Fatal(err)
	}
	if len(eventsB) != 1 {
		t.Errorf("proj-B events: got %d, want 1", len(eventsB))
	}

	eventsC, err := store.ListEventsByProject(ctx, "proj-MISSING")
	if err != nil {
		t.Fatal(err)
	}
	if len(eventsC) != 0 {
		t.Errorf("missing project events: got %d, want 0", len(eventsC))
	}
}

func TestSignalsByProjectIndex_Consistency(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	store.SaveSignal(ctx, &domain.Signal{ID: "sig-1", ProjectID: "proj-X", Type: domain.SignalConstructionSignal, Magnitude: 0.8})
	store.SaveSignal(ctx, &domain.Signal{ID: "sig-2", ProjectID: "proj-X", Type: domain.SignalRegulatoryProgress, Magnitude: 0.6})
	store.SaveSignal(ctx, &domain.Signal{ID: "sig-3", ProjectID: "proj-Y", Type: domain.SignalFinancingAcceleration, Magnitude: 0.9})

	// Rebuild index from scratch.
	store.RebuildSignalIndex()

	// Verify signal index is internally consistent by checking map sizes.
	store.mu.RLock()
	xIDs := store.signalsByProject["proj-X"]
	yIDs := store.signalsByProject["proj-Y"]
	store.mu.RUnlock()

	if len(xIDs) != 2 {
		t.Errorf("proj-X signal IDs: got %d, want 2", len(xIDs))
	}
	if len(yIDs) != 1 {
		t.Errorf("proj-Y signal IDs: got %d, want 1", len(yIDs))
	}
}

func TestCapexCacheDirty(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	store.SaveProject(ctx, &domain.Project{ID: "proj-1", CapexCAD: 100_000_000, CapexStatus: domain.ConfidenceReported})
	stats1, _ := store.GetRadarStats(ctx)
	if stats1.TotalCapexCAD != 100_000_000 {
		t.Errorf("capex = %d, want 100000000", stats1.TotalCapexCAD)
	}

	// Save another project — capex cache should be marked dirty.
	store.SaveProject(ctx, &domain.Project{ID: "proj-2", CapexCAD: 50_000_000, CapexStatus: domain.ConfidenceVerified})
	stats2, _ := store.GetRadarStats(ctx)
	if stats2.TotalCapexCAD != 150_000_000 {
		t.Errorf("capex = %d, want 150000000", stats2.TotalCapexCAD)
	}
}

// ─── benchmarks ───────────────────────────────────────────────────────────────

func BenchmarkListEventsByProject(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()

	store.SaveProject(ctx, &domain.Project{ID: "proj-bench", Slug: "bench"})
	store.SaveEvidence(ctx, &domain.Evidence{ID: "ev-bench"})

	for i := 0; i < 500; i++ {
		store.SaveEvent(ctx, &domain.Event{
			ID:         "event-" + string(rune(i)) + "-" + string(rune(i/256)),
			ProjectID:  "proj-bench",
			EvidenceID: "ev-bench",
			Title:      "bench event",
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.ListEventsByProject(ctx, "proj-bench")
	}
}
