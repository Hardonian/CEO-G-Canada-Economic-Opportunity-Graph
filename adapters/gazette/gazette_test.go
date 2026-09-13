package gazette

import (
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const onFixture = `{
  "source": "Ontario Gazette",
  "source_url": "https://www.ontario.ca/page/ontario-gazette",
  "retrieved_at": "2026-09-13T00:00:00Z",
  "effective_at": "2026-09-13T00:00:00Z",
  "dataset_vintage": "2026-Q3",
  "notices": [
    {
      "notice_id": "OG-2026-09-001",
      "title": "Environmental Assessment Registration for the Ring of Fire Clean Energy Transmission Line",
      "content": "The proponent, NorthStar Energy Corporation, has filed a preliminary screening report for a 450-kilometre 500-kV transmission line to connect Ring of Fire mining projects to the Ontario grid. Construction is estimated at C$2.8 billion.",
      "publish_date": "2026-09-10",
      "source_url": "https://www.ontario.ca/page/ontario-gazette-notice-2026-09-001",
      "category": "Environmental Assessment",
      "proponent_name": "NorthStar Energy Corporation",
      "project_name": "Ring of Fire Clean Energy Transmission Line",
      "location": "Thunder Bay",
      "amount_text": "C$2.8 billion",
      "metadata": {"naics_code": "221114"}
    },
    {
      "notice_id": "OG-2026-09-005",
      "title": "Annual Report on Provincial Parks Operations",
      "content": "This notice provides the annual summary of operating revenues and expenditures for Ontario's provincial parks system for the fiscal year 2025-26.",
      "publish_date": "2026-09-08",
      "source_url": "https://www.ontario.ca/page/ontario-gazette-notice-2026-09-005",
      "category": "Administrative",
      "metadata": {"naics_code": "924110"}
    }
  ]
}`

func TestParseGazetteFixtureProducesProjects(t *testing.T) {
	adapter := NewGazetteAdapter("ON", "unused.json")
	result, err := adapter.Parse([]byte(onFixture))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(result.Projects) != 1 {
		t.Fatalf("projects = %d, want 1 (administrative notice should be filtered)", len(result.Projects))
	}
	proj := result.Projects[0]
	if proj.Province != "ON" {
		t.Errorf("province = %q, want ON", proj.Province)
	}
	if proj.Name != "Ring of Fire Clean Energy Transmission Line" {
		t.Errorf("name = %q, want project name", proj.Name)
	}
	if proj.Sector != domain.SectorCleanEnergy {
		t.Errorf("sector = %q, want %q", proj.Sector, domain.SectorCleanEnergy)
	}
	if proj.CapexCAD != 2_800_000_000 {
		t.Errorf("capex = %d, want 2800000000", proj.CapexCAD)
	}
	if proj.CurrentStage != domain.StageEnvironmentalReview {
		t.Errorf("stage = %q, want %q", proj.CurrentStage, domain.StageEnvironmentalReview)
	}
	if proj.Proponent == nil || proj.Proponent.CommonName != "NorthStar Energy Corporation" {
		t.Errorf("proponent = %v", proj.Proponent)
	}
	if len(result.Evidence) != 1 {
		t.Errorf("evidence = %d, want 1", len(result.Evidence))
	}
	if len(result.Entities) != 1 {
		t.Errorf("entities = %d, want 1", len(result.Entities))
	}
}

func TestParseGazetteStableIDs(t *testing.T) {
	adapter := NewGazetteAdapter("ON", "unused.json")
	first, err := adapter.Parse([]byte(onFixture))
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Parse([]byte(onFixture))
	if err != nil {
		t.Fatal(err)
	}
	if first.Projects[0].ID != second.Projects[0].ID {
		t.Error("stable source input produced unstable project IDs")
	}
	if first.Evidence[0].ID != second.Evidence[0].ID {
		t.Error("stable source input produced unstable evidence IDs")
	}
}

func TestParseGazetteEmptyDocument(t *testing.T) {
	adapter := NewGazetteAdapter("ON", "unused.json")
	if _, err := adapter.Parse([]byte("")); err == nil {
		t.Fatal("expected error for empty document")
	}
}

func TestParseGazetteNoNotices(t *testing.T) {
	adapter := NewGazetteAdapter("ON", "unused.json")
	fixture := `{"source":"test","notices":[]}`
	if _, err := adapter.Parse([]byte(fixture)); err == nil {
		t.Fatal("expected error for empty notices")
	}
}

func TestParseGazetteCapexParsing(t *testing.T) {
	adapter := NewGazetteAdapter("ON", "unused.json")
	result, err := adapter.Parse([]byte(onFixture))
	if err != nil {
		t.Fatal(err)
	}
	if result.Projects[0].CapexCAD != 2_800_000_000 {
		t.Errorf("capex = %d, want 2800000000", result.Projects[0].CapexCAD)
	}
	if result.Projects[0].CapexStatus != domain.ConfidenceReported {
		t.Errorf("capex_status = %q, want %q", result.Projects[0].CapexStatus, domain.ConfidenceReported)
	}
}

func TestGazettePublisher(t *testing.T) {
	tests := []struct {
		province string
		want     string
	}{
		{"ON", "Ontario Gazette"},
		{"QC", "Gazette officielle du Québec"},
		{"BC", "BC Gazette"},
		{"CA", "Canada Gazette"},
		{"XX", "XX Gazette"},
	}
	for _, tt := range tests {
		if got := gazettePublisher(tt.province); got != tt.want {
			t.Errorf("gazettePublisher(%q) = %q, want %q", tt.province, got, tt.want)
		}
	}
}
