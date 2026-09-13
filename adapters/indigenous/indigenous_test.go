package indigenous

import (
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const iscFixture = `{
  "source": "Indigenous Services Canada Business Directory",
  "source_url": "https://www.sac-isc.gc.ca/eng/1100100032800/1558373938872",
  "retrieved_at": "2026-09-13T00:00:00Z",
  "effective_at": "2026-09-13T00:00:00Z",
  "dataset_vintage": "2026-Q3",
  "businesses": [
    {
      "business_id": "ISC-2024-001",
      "business_name": "Nihtaw Consulting Ltd.",
      "legal_name": "Nihtaw Consulting Ltd.",
      "operating_name": "Nihtaw Consulting",
      "indigenous_group": "First Nation",
      "community_name": "Treaty 6 Territory",
      "province": "AB",
      "city": "Edmonton",
      "postal_code": "T5H 2W9",
      "naics_code": "541620",
      "naics_description": "Environmental Consulting Services",
      "business_type": "Corporation",
      "ownership_percent": 60,
      "certification_date": "2024-03-15",
      "expiry_date": "2027-03-15",
      "status": "Active",
      "contact_email": "info@nihtaw.com",
      "contact_phone": "780-555-0100",
      "website": "https://www.nihtaw.com",
      "address": "100 King's University Drive, Edmonton, AB",
      "description": "Indigenous-owned environmental and Indigenous relations consulting firm.",
      "capabilities": ["environmental_assessment", "indigenous_relations", "consultation"],
      "metadata": {
        "status": "Active",
        "ownership_percent": 60
      }
    },
    {
      "business_id": "ISC-2024-002",
      "business_name": "Mino Bimaadiziwin Construction",
      "legal_name": "Mino Bimaadiziwin Construction Ltd.",
      "operating_name": "Mino Bimaadiziwin Construction",
      "indigenous_group": "First Nation",
      "community_name": "Treaty 3 Territory",
      "province": "ON",
      "city": "Thunder Bay",
      "postal_code": "P7B 3A6",
      "naics_code": "23",
      "naics_description": "Construction",
      "business_type": "Corporation",
      "ownership_percent": 100,
      "certification_date": "2024-06-01",
      "expiry_date": "2027-06-01",
      "status": "Active",
      "contact_email": "info@minobimaadiziwin.com",
      "contact_phone": "807-555-0123",
      "website": "https://www.minobimaadiziwin.com",
      "address": "1250 Memorial University Drive, Thunder Bay, ON",
      "description": "Indigenous-owned construction company.",
      "capabilities": ["site_preparation", "civil_works", "road_construction"],
      "metadata": {
        "status": "Active",
        "ownership_percent": 100
      }
    },
    {
      "business_id": "ISC-2024-003",
      "business_name": "Duplicate Biz",
      "legal_name": "Duplicate Biz",
      "operating_name": "",
      "indigenous_group": "Métis",
      "community_name": "Saskatoon",
      "province": "SK",
      "city": "Saskatoon",
      "postal_code": "S7K 4B8",
      "naics_code": "541330",
      "naics_description": "Engineering Services",
      "business_type": "Corporation",
      "ownership_percent": 80,
      "certification_date": "2024-08-20",
      "expiry_date": "2027-08-20",
      "status": "Active",
      "contact_email": "",
      "contact_phone": "",
      "website": "",
      "address": "",
      "description": "Duplicate entry for dedup testing.",
      "capabilities": ["engineering", "project_management"],
      "metadata": {
        "status": "Active",
        "ownership_percent": 80
      }
    }
  ]
}`

func TestParseIndigenousFixtureProducesEntities(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	result, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Entities) != 3 {
		t.Fatalf("entities = %d, want 3", len(result.Entities))
	}
	if len(result.Evidence) != 3 {
		t.Fatalf("evidence = %d, want 3", len(result.Evidence))
	}
}

func TestParseIndigenousEntityNormalization(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	result, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	entity := result.Entities[0]
	if entity.EntityType != "IndigenousBusiness" {
		t.Errorf("entity_type = %q, want IndigenousBusiness", entity.EntityType)
	}
	if entity.Jurisdiction != "CA:AB" {
		t.Errorf("jurisdiction = %q, want CA:AB", entity.Jurisdiction)
	}
	if entity.Identifiers["isc_business_id"] != "ISC-2024-001" {
		t.Errorf("isc_business_id = %q, want ISC-2024-001", entity.Identifiers["isc_business_id"])
	}
	if entity.Identifiers["naics_code"] != "541620" {
		t.Errorf("naics_code = %q, want 541620", entity.Identifiers["naics_code"])
	}
}

func TestParseIndigenousStableIDs(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	first, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	if first.Entities[0].ID != second.Entities[0].ID {
		t.Error("stable source input produced unstable entity IDs")
	}
	if first.Evidence[0].ID != second.Evidence[0].ID {
		t.Error("stable source input produced unstable evidence IDs")
	}
}

func TestParseIndigenousEmptyDocument(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	if _, err := adapter.Parse([]byte("")); err == nil {
		t.Fatal("expected error for empty document")
	}
}

func TestParseIndigenousDuplicateBusinessIDs(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	result, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	if result.Evidence[2].SourceRecordID != "ISC-2024-003" {
		t.Errorf("third entity should be ISC-2024-003, got %q", result.Evidence[2].SourceRecordID)
	}
}

func TestCrossReferenceProcurementNAICS(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	result, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	businesses := result.Entities

	procurement := &domain.Procurement{
		ID:            "P-001",
		TenderID:      "T-001",
		Title:         "Environmental Remediation Contract",
		Buyer:         "Environment and Climate Change Canada",
		BuyerType:     "Federal",
		SourceURL:     "https://example.com",
		Categories:    []string{"environmental"},
		EvidenceID:    "E-001",
		CreatedAt:     result.Evidence[0].RetrievalTimestamp,
		Metadata:      map[string]interface{}{"naics_code": "5416", "province": "AB"},
	}
	matches := CrossReferenceProcurement(businesses, procurement)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match for AB environmental procurement, got %d", len(matches))
	}
	for _, m := range matches {
		if m.EntityType != "IndigenousBusiness" {
			t.Errorf("matched entity type = %q, want IndigenousBusiness", m.EntityType)
		}
		if m.Jurisdiction != "CA:AB" {
			t.Errorf("matched jurisdiction = %q, want CA:AB", m.Jurisdiction)
		}
	}
}

func TestCrossReferenceProcurementRegion(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	result, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	businesses := result.Entities

	procurement := &domain.Procurement{
		ID:     "P-002",
		TenderID: "T-002",
		Title:  "Construction Project in Ontario",
		Buyer:  "Province of Ontario",
		Categories: []string{"construction"},
		EvidenceID: "E-002",
		CreatedAt:  result.Evidence[0].RetrievalTimestamp,
	}
	procurement.Metadata = map[string]interface{}{"province": "ON"}

	matches := CrossReferenceProcurement(businesses, procurement)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match for ON region, got %d", len(matches))
	}
	if matches[0].Jurisdiction != "CA:ON" {
		t.Errorf("matched jurisdiction = %q, want CA:ON", matches[0].Jurisdiction)
	}
}

func TestParseIndigenousExpiredBusiness(t *testing.T) {
	adapter := NewIndigenousAdapter("unused.json")
	result, err := adapter.Parse([]byte(iscFixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range result.Entities {
		if e.Metadata["status"] == "Expired" {
			if isBusinessActive(e) {
				t.Error("expired business should not be active")
			}
		}
	}
}

func TestRegionMatchIndigenous(t *testing.T) {
	tests := []struct {
		proc string
		biz  string
		want bool
	}{
		{"ON", "ON", true},
		{"ON", "BC", false},
		{"ONTARIO", "QC", false},
		{"QUEBEC", "QC", true},
	}
	for _, tt := range tests {
		if got := regionMatch(tt.proc, tt.biz); got != tt.want {
			t.Errorf("regionMatch(%q, %q) = %v, want %v", tt.proc, tt.biz, got, tt.want)
		}
	}
}

func TestNaicsMatchIndigenous(t *testing.T) {
	tests := []struct {
		proc string
		biz  string
		want bool
	}{
		{"5416", "541620", true},
		{"23", "541620", false},
		{"541620", "541620", true},
	}
	for _, tt := range tests {
		if got := naicsMatch(tt.proc, tt.biz); got != tt.want {
			t.Errorf("naicsMatch(%q, %q) = %v, want %v", tt.proc, tt.biz, got, tt.want)
		}
	}
}

func TestProvinceAdjacentIndigenous(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"ON", "ON", true},
		{"AB", "BC", true},
		{"ON", "BC", false},
		{"SK", "MB", true},
	}
	for _, tt := range tests {
		if got := provinceAdjacent(tt.a, tt.b); got != tt.want {
			t.Errorf("provinceAdjacent(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
