package reconciliation

import (
	"context"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestReconcileRecords_Empty(t *testing.T) {
	report := ReconcileRecords([]*JurisdictionRecord{})
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.TotalRecords != 0 {
		t.Fatalf("expected 0 total records, got %d", report.TotalRecords)
	}
	if report.Summary.Merged != 0 || report.Summary.Linked != 0 || report.Summary.Conflicts != 0 {
		t.Fatalf("expected all-zero summary for empty input, got %+v", report.Summary)
	}
}

func TestReconcileRecords_AllLinked(t *testing.T) {
	// Federal + Provincial records that describe the same project.
	fed := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionFederal, Title: "Nuclear Reactor Build",
		Location: "Ontario", Stage: domain.StageFeasibility, CapexCAD: 1_000_000_000,
	}
	prov := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionProvincial, Title: "Nuclear Reactor Build",
		Location: "Ontario", Stage: domain.StageFeasibility, CapexCAD: 1_000_000_000,
	}
	report := ReconcileRecords([]*JurisdictionRecord{fed, prov})
	if report.TotalRecords != 2 {
		t.Fatalf("expected 2 total records, got %d", report.TotalRecords)
	}
	if report.Summary.Linked != 1 {
		t.Fatalf("expected 1 linked match, got %d (summary=%+v)", report.Summary.Linked, report.Summary)
	}
	if report.Summary.Merged != 0 || report.Summary.Conflicts != 0 {
		t.Fatalf("expected no merges/conflicts, got %+v", report.Summary)
	}
}

func TestReconcileRecords_MultiJurisdictionMerge(t *testing.T) {
	fed := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionFederal, Title: "Cameco Uranium Facility",
		Location: "Saskatchewan", Stage: domain.StageFEED, CapexCAD: 2_000_000_000,
	}
	prov := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionProvincial, Title: "Cameco Uranium Facility",
		Location: "Saskatchewan", Stage: domain.StageFEED, CapexCAD: 2_000_000_000,
	}
	mun := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionMunicipal, Title: "Cameco Uranium Facility",
		Location: "Saskatchewan", Stage: domain.StageFEED, CapexCAD: 2_000_000_000,
	}
	report := ReconcileRecords([]*JurisdictionRecord{fed, prov, mun})
	if report.Summary.Merged != 1 {
		t.Fatalf("expected 1 merge, got %d (summary=%+v)", report.Summary.Merged, report.Summary)
	}
}

func TestReconcileRecords_ConflictingStages(t *testing.T) {
	fed := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionFederal, Title: "Hydro Dam Expansion",
		Location: "Manitoba", Stage: domain.StageConstruction, CapexCAD: 5_000_000_000,
	}
	prov := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionProvincial, Title: "Hydro Dam Expansion",
		Location: "Manitoba", Stage: domain.StageFeasibility, CapexCAD: 5_000_000_000,
	}
	report := ReconcileRecords([]*JurisdictionRecord{fed, prov})
	if report.Summary.Conflicts < 1 {
		t.Fatalf("expected at least 1 conflict for stage mismatch, got %d (summary=%+v)",
			report.Summary.Conflicts, report.Summary)
	}
}

func TestReconcileRecords_ConflictingCAPEX(t *testing.T) {
	fed := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionFederal, Title: "LNG Terminal",
		Location: "BC", Stage: domain.StageFEED, CapexCAD: 1_000_000_000,
	}
	prov := &JurisdictionRecord{
		ProjectID: "p1", Level: JurisdictionProvincial, Title: "LNG Terminal",
		Location: "BC", Stage: domain.StageFEED, CapexCAD: 50_000_000,
	}
	report := ReconcileRecords([]*JurisdictionRecord{fed, prov})
	if report.Summary.Conflicts < 1 {
		t.Fatalf("expected at least 1 conflict for CAPEX discrepancy, got %d (summary=%+v)",
			report.Summary.Conflicts, report.Summary)
	}
}

func TestReconcileRecords_UnmatchedProvincial(t *testing.T) {
	prov := &JurisdictionRecord{
		ProjectID: "orphan-1", Level: JurisdictionProvincial, Title: "Orphan Project",
		Location: "Ontario", Stage: domain.StageAnnounced, CapexCAD: 100_000_000,
	}
	report := ReconcileRecords([]*JurisdictionRecord{prov})
	if report.Summary.NoMatch < 1 {
		t.Fatalf("expected at least 1 no_match for orphan provincial, got %d (summary=%+v)",
			report.Summary.NoMatch, report.Summary)
	}
	if len(report.UnmatchedProvincial) != 1 {
		t.Fatalf("expected 1 unmatched provincial, got %d", len(report.UnmatchedProvincial))
	}
}

func TestReconcile_WithStore(t *testing.T) {
	store := database.NewMemoryStore()
	now := time.Now().UTC()
	proj := &domain.Project{
		ID: "p1", Name: "P1", Sector: domain.SectorNuclearEnergy, Province: "ON",
		CurrentStage: domain.StageFeasibility, CapexCAD: 1_000_000_000, UpdatedAt: now,
	}
	ev := &domain.Evidence{
		ID: "ev-1", ContentHash: "abc", SourceURL: "https://example.com", Publisher: "T",
		SourceTier: domain.SourceTier1, RetrievalTimestamp: now,
		Confidence: domain.ConfidenceVerified, ExtractionMethod: "test",
		Visibility: domain.VisibilityPublic, Publishable: true,
	}
	if err := store.SaveProject(context.Background(), proj); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvidence(context.Background(), ev); err != nil {
		t.Fatal(err)
	}
	report := Reconcile(context.Background(), store)
	if report.TotalRecords != 1 {
		t.Fatalf("expected 1 total record (project only), got %d", report.TotalRecords)
	}
}