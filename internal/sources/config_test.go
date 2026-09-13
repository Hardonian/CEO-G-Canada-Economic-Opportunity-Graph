package sources

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
)

func TestConfigurationImportsHierarchyWithoutActivatingSources(t *testing.T) {
	input := `{
  "schema_version":"1",
  "publishers":[{"id":"publisher:test","name":"Test Authority","authority_tier":1}],
  "sources":[
    {"id":"source:resource","publisher_id":"publisher:test","parent_source_id":"source:catalog","name":"Resource","canonical_url":"https://data.example.ca/resource.csv","source_kind":"RESOURCE","source_family":"CSV","authority_tier":1,"lifecycle_status":"CLASSIFIED","health_status":"UNKNOWN","terms_status":"UNKNOWN","robots_status":"NOT_APPLICABLE"},
    {"id":"source:catalog","publisher_id":"publisher:test","name":"Catalog","canonical_url":"https://data.example.ca/api","source_kind":"CATALOG","source_family":"CKAN","authority_tier":1,"lifecycle_status":"APPROVED","health_status":"UNKNOWN","terms_status":"ALLOWED","robots_status":"NOT_APPLICABLE"}
  ]
}`
	config, err := DecodeConfiguration(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	store := database.NewMemoryStore()
	when := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	if err := ApplyConfiguration(context.Background(), store, config, when); err != nil {
		t.Fatal(err)
	}
	resource, err := store.GetSource(context.Background(), "source:resource")
	if err != nil {
		t.Fatal(err)
	}
	if resource.ParentSourceID != "source:catalog" || resource.CreatedAt != when {
		t.Fatalf("unexpected imported resource: %#v", resource)
	}
	coverage, err := store.GetSourceCoverage(context.Background(), when)
	if err != nil {
		t.Fatal(err)
	}
	if coverage.TotalSources != 2 || coverage.Active != 0 {
		t.Fatalf("registration was conflated with activation: %#v", coverage)
	}
}

func TestConfigurationRejectsUnknownFields(t *testing.T) {
	_, err := DecodeConfiguration(strings.NewReader(`{"schema_version":"1","publishers":[],"sources":[],"secret":"no"}`))
	if err == nil {
		t.Fatal("unknown configuration field was accepted")
	}
}
