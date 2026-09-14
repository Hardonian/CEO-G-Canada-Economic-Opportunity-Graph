package cegs_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestCEGSValidation(t *testing.T) {
	// Valid Project JSON
	validProj := map[string]interface{}{
		"cegs":           "0.1",
		"id":             "cegs:project:ca:on:darlington-smr",
		"type":           "project",
		"canonical_name": "Darlington Small Modular Reactor (SMR) Project",
		"jurisdiction":   "CA:ON",
		"sector":         "Nuclear & Clean Power",
		"stage":          "CONSTRUCTION",
		"capex": map[string]interface{}{
			"amount":      3400000000,
			"currency":    "CAD",
			"amount_type": "reported",
		},
		"location": map[string]interface{}{
			"name":      "Clarington",
			"province":  "ON",
			"latitude":  43.86,
			"longitude": -78.71,
		},
		"provenance": []string{"cegs:evidence:ca:iaac-registry-darlington"},
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	raw, err := json.Marshal(validProj)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	report, err := cegs.Validate(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Valid {
		t.Errorf("expected valid, got errors: %v", report.Errors)
	}
	if report.ConformanceLevel != "CEGS Provenance" {
		t.Errorf("expected 'CEGS Provenance', got '%s'", report.ConformanceLevel)
	}

	// Invalid Project: missing sector and stage
	invalidProj := map[string]interface{}{
		"cegs":           "0.1",
		"id":             "cegs:project:ca:test",
		"type":           "project",
		"canonical_name": "Test Project",
		"jurisdiction":   "CA",
		"created_at":     time.Now().Format(time.RFC3339),
		"updated_at":     time.Now().Format(time.RFC3339),
	}
	rawInv, _ := json.Marshal(invalidProj)
	repInv, _ := cegs.Validate(rawInv)
	if repInv.Valid {
		t.Errorf("expected invalid for missing sector and capex, but marked valid")
	}
}

func TestCEGSInspectionAndDiff(t *testing.T) {
	oldDoc := map[string]interface{}{
		"cegs":           "0.1",
		"id":             "cegs:project:ca:on:darlington-smr",
		"type":           "project",
		"canonical_name": "Darlington SMR Project",
		"stage":          "PERMITTING",
		"capex": map[string]interface{}{
			"amount":   2800000000,
			"currency": "CAD",
		},
		"source_status": "REPORTED",
		"provenance":    []string{"cegs:evidence:ca:initial-news"},
	}

	newDoc := map[string]interface{}{
		"cegs":           "0.1",
		"id":             "cegs:project:ca:on:darlington-smr",
		"type":           "project",
		"canonical_name": "Darlington Small Modular Reactor (SMR) Project",
		"stage":          "CONSTRUCTION",
		"capex": map[string]interface{}{
			"amount":   3400000000,
			"currency": "CAD",
		},
		"source_status": "VERIFIED",
		"provenance":    []string{"cegs:evidence:ca:initial-news", "cegs:evidence:ca:cib-agreement"},
	}

	oldBytes, _ := json.Marshal(oldDoc)
	newBytes, _ := json.Marshal(newDoc)

	// Test Inspect
	insp, err := cegs.Inspect(newBytes)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}
	if insp.Stage != "CONSTRUCTION" {
		t.Errorf("expected stage CONSTRUCTION, got %s", insp.Stage)
	}
	if insp.EvidenceCount != 2 {
		t.Errorf("expected evidence count 2, got %d", insp.EvidenceCount)
	}

	// Test Diff
	diff, err := cegs.Diff(oldBytes, newBytes)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !diff.HasChanges {
		t.Fatalf("expected semantic changes, found none")
	}

	foundStageChange := false
	foundCapexChange := false
	for _, ch := range diff.Changes {
		if ch.Code == "PROJECT_STAGE_CHANGED" {
			foundStageChange = true
		}
		if ch.Code == "CAPEX_CHANGED" {
			foundCapexChange = true
		}
	}
	if !foundStageChange {
		t.Errorf("expected PROJECT_STAGE_CHANGED in diff")
	}
	if !foundCapexChange {
		t.Errorf("expected CAPEX_CHANGED in diff")
	}
}

func TestDomainToCEGSRoundtrip(t *testing.T) {
	proj := &domain.Project{
		ID:           "proj-123",
		Slug:         "ontario-grid-battery",
		Name:         "Oneida Energy Storage Project",
		Summary:      "250 MW / 1,000 MWh utility-scale battery energy storage system.",
		Sector:       domain.SectorCleanEnergy,
		Subsector:    "Battery Storage",
		Province:     "ON",
		LocationName: "Jarvis, Haldimand County",
		Latitude:     42.88,
		Longitude:    -80.05,
		CurrentStage: domain.StageConstruction,
		CapexCAD:     450000000,
		CapexStatus:  domain.ConfidenceReported,
		Confidence:   domain.ConfidenceVerified,
		ProponentID:  "entity-123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	cegsProj := cegs.ToCEGSProject(proj, []string{"cegs:evidence:ca:iaac-filing"})
	if cegsProj.ID != "cegs:project:ca:on:proj-123" {
		t.Errorf("expected stable ID 'cegs:project:ca:on:proj-123', got '%s'", cegsProj.ID)
	}
	if cegsProj.Capex.Amount != 450000000 {
		t.Errorf("expected capex amount 450000000, got %d", cegsProj.Capex.Amount)
	}
	if len(cegsProj.Proponents) != 1 || cegsProj.Proponents[0] != "cegs:org:ca:entity-123" {
		t.Fatalf("project proponent reference is not canonical: %#v", cegsProj.Proponents)
	}
}

func TestConvertersPreserveCanonicalReferencesWithoutMutatingDomain(t *testing.T) {
	now := time.Now().UTC()
	evidenceIDs := []string{"evidence-1"}
	entity := &domain.Entity{ID: "entity-1", Slug: "shared-name", CommonName: "Shared Name", LegalName: "Shared Name Inc.", EvidenceID: "evidence-1", CreatedAt: now, UpdatedAt: now}
	project := &domain.Project{ID: "project-1", Province: "ON", Name: "Project", ProponentID: entity.ID, EvidenceIDs: evidenceIDs, CreatedAt: now, UpdatedAt: now}
	event := &domain.Event{ID: "event-1", ProjectID: project.ID, EventType: "stage_change", EventDate: now, Title: "Changed", Description: "Changed", EvidenceID: "evidence-1", CreatedAt: now}

	exportedProject := cegs.ToCEGSProject(project, evidenceIDs)
	exportedEntity := cegs.ToCEGSOrganization(entity)
	exportedEvent := cegs.ToCEGSEvent(event, project)

	if evidenceIDs[0] != "evidence-1" || project.EvidenceIDs[0] != "evidence-1" {
		t.Fatalf("project conversion mutated domain evidence IDs: %#v", evidenceIDs)
	}
	if exportedProject.Proponents[0] != exportedEntity.ID {
		t.Fatalf("orphaned proponent reference: project=%q organization=%q", exportedProject.Proponents[0], exportedEntity.ID)
	}
	if exportedEvent.Subject != exportedProject.ID {
		t.Fatalf("orphaned event subject: event=%q project=%q", exportedEvent.Subject, exportedProject.ID)
	}
	if exportedEvent.Evidence[0] != exportedProject.Provenance[0] {
		t.Fatalf("inconsistent evidence reference: event=%q project=%q", exportedEvent.Evidence[0], exportedProject.Provenance[0])
	}
}

func TestEvidenceExportPreservesVersionedDataLineage(t *testing.T) {
	evidence := &domain.Evidence{
		ID:                 "evidence-42",
		SourceURL:          "https://example.gc.ca/dataset/42",
		Publisher:          "Example public authority",
		SourceTier:         domain.SourceTier1,
		ContentHash:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SourceID:           "source-42",
		SourceVersionID:    "version-7",
		SourceRecordID:     "row-9",
		Locator:            "row=9",
		ParserVersion:      "csv-v2",
		MappingVersion:     "project-v3",
		PipelineVersion:    "ingestion-v4",
		Confidence:         domain.ConfidenceVerified,
		ExtractionMethod:   "deterministic_adapter",
		RetrievalTimestamp: time.Now().UTC(),
	}
	exported := cegs.ToCEGSEvidence(evidence)
	lineage, ok := exported.Extensions["ca.opengraph.public_data_lineage"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing lineage extension: %#v", exported.Extensions)
	}
	if lineage["source_version_id"] != "version-7" || lineage["mapping_version"] != "project-v3" {
		t.Fatalf("lineage was not preserved: %#v", lineage)
	}
}

func TestSpecExamples(t *testing.T) {
	exampleFiles := []string{
		"../../spec/cegs/examples/project.json",
		"../../spec/cegs/examples/organization.json",
		"../../spec/cegs/examples/event.json",
		"../../spec/cegs/examples/relationship.json",
		"../../spec/cegs/examples/evidence.json",
		"../../spec/cegs/examples/dataset.json",
		"../../spec/cegs/examples/source.json",
	}

	for _, file := range exampleFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}
		rep, err := cegs.Validate(data)
		if err != nil {
			t.Fatalf("error validating %s: %v", file, err)
		}
		if !rep.Valid {
			t.Errorf("file %s failed validation: %v", file, rep.Errors)
		}
	}
}
