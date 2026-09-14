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
