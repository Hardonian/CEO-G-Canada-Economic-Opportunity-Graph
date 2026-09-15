package indigenouslinker

import (
	"context"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestLinkResult_EmptyStore(t *testing.T) {
	store := database.NewMemoryStore()
	result := Link(context.Background(), store)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.RelationshipsCreated != 0 {
		t.Errorf("expected 0 relationships, got %d", result.RelationshipsCreated)
	}
}

func TestLinkResult_NoIndigenousBusinesses(t *testing.T) {
	store := database.NewMemoryStore()
	ctx := context.Background()
	proj := &domain.Project{
		ID: "p1", Name: "Test", Sector: domain.SectorCleanEnergy, Province: "ON",
		CurrentStage: domain.StageConcept, CapexCAD: 100_000_000, UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveProject(ctx, proj); err != nil {
		t.Fatal(err)
	}
	result := Link(ctx, store)
	if result.ProjectsScanned != 1 {
		t.Errorf("expected 1 project scanned, got %d", result.ProjectsScanned)
	}
	if result.RelationshipsCreated != 0 {
		t.Errorf("expected 0 relationships (no indigenous businesses), got %d", result.RelationshipsCreated)
	}
}

func TestLinkResult_WithIndigenousBusiness(t *testing.T) {
	store := database.NewMemoryStore()
	ctx := context.Background()
	proj := &domain.Project{
		ID: "p1", Name: "Test Project", Sector: domain.SectorNuclearEnergy, Province: "ON",
		CurrentStage: domain.StageFEED, CapexCAD: 2_000_000_000, UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveProject(ctx, proj); err != nil {
		t.Fatal(err)
	}
	biz := &domain.Entity{
		ID:           "ib-1",
	 CommonName:   "Test Indigenous Business",
	 LegalName:    "Test Indigenous Business Ltd",
	 EntityType:   "IndigenousBusiness",
	 Jurisdiction: "ON",
	 Identifiers: map[string]string{
			"naics_code": "221113",
		},
		UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveEntity(ctx, biz); err != nil {
		t.Fatal(err)
	}
	result := Link(ctx, store)
	if result.ProjectsScanned != 1 {
		t.Errorf("expected 1 project scanned, got %d", result.ProjectsScanned)
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestLinkResult_Idempotent(t *testing.T) {
	store := database.NewMemoryStore()
	ctx := context.Background()
	proj := &domain.Project{
		ID: "p1", Name: "Test Project", Sector: domain.SectorNuclearEnergy, Province: "ON",
		CurrentStage: domain.StageFEED, CapexCAD: 2_000_000_000, UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveProject(ctx, proj); err != nil {
		t.Fatal(err)
	}
	biz := &domain.Entity{
		ID:           "ib-1",
	 CommonName:   "Test Indigenous Business",
	 LegalName:    "Test Indigenous Business Ltd",
	 EntityType:   "IndigenousBusiness",
	 Jurisdiction: "ON",
	 Identifiers: map[string]string{
			"naics_code": "221113",
		},
		UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveEntity(ctx, biz); err != nil {
		t.Fatal(err)
	}
	result1 := Link(ctx, store)
	result2 := Link(ctx, store)
	if result1.RelationshipsCreated != result2.RelationshipsCreated {
		t.Errorf("idempotency violated: first=%d, second=%d",
			result1.RelationshipsCreated, result2.RelationshipsCreated)
	}
}

func TestLinkResult_ProcurementScanned(t *testing.T) {
	store := database.NewMemoryStore()
	ctx := context.Background()
	proj := &domain.Project{
		ID: "p1", Name: "Test Project", Sector: domain.SectorCleanEnergy, Province: "ON",
		CurrentStage: domain.StageConcept, CapexCAD: 100_000_000, UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveProject(ctx, proj); err != nil {
		t.Fatal(err)
	}
	pr := &domain.Procurement{
		ID:         "proc-1",
		ProjectID:  "p1",
		Stage:      "OPEN",
		SourceURL:  "https://example.com/proc-1",
		EvidenceID: "ev-1",
	}
	if err := store.SaveProcurement(ctx, pr); err != nil {
		t.Fatal(err)
	}
	biz := &domain.Entity{
		ID:           "ib-1",
	 CommonName:   "Test Indigenous Business",
	 LegalName:    "Test Indigenous Business Ltd",
	 EntityType:   "IndigenousBusiness",
	 Jurisdiction: "ON",
	 Identifiers: map[string]string{
			"naics_code": "221113",
		},
		UpdatedAt: time.Now().UTC(),
	}
	if err := store.SaveEntity(ctx, biz); err != nil {
		t.Fatal(err)
	}
	result := Link(ctx, store)
	if result.ProcurementsScanned != 1 {
		t.Errorf("expected 1 procurement scanned, got %d", result.ProcurementsScanned)
	}
}

func TestLinkResult_Errors(t *testing.T) {
	result := Link(context.Background(), nil)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Errors) == 0 {
		t.Error("expected at least one error for nil store")
	}
}

func TestLinkResult_Fields(t *testing.T) {
	r := &LinkResult{
		ProjectsScanned:      10,
		ProcurementsScanned:  5,
		RelationshipsCreated: 3,
		Errors:               []string{"test error"},
	}
	if r.ProjectsScanned != 10 {
		t.Error("ProjectsScanned field not set")
	}
	if r.ProcurementsScanned != 5 {
		t.Error("ProcurementsScanned field not set")
	}
	if r.RelationshipsCreated != 3 {
		t.Error("RelationshipsCreated field not set")
	}
	if len(r.Errors) != 1 || r.Errors[0] != "test error" {
		t.Error("Errors field not set")
	}
}