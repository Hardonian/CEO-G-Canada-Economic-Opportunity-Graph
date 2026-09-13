package nrcan_major_projects

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const testSnapshot = `{
  "source": "Natural Resources Canada Major Projects Inventory",
  "source_url": "https://open.canada.ca/data/en/dataset/f5f2db55-31e4-42fb-8c73-23e1c44de9b2",
  "dataset_vintage": "2025-2035",
  "effective_at": "2026-04-20T00:00:00Z",
  "retrieved_at": "2026-09-13T00:00:00Z",
  "license": "Open Government Licence - Canada",
  "features": [
    {
      "attributes": {
        "id": "0010",
        "company": "Aeolis / Boralex",
        "project_name": "Hackney Hills Wind Park",
        "province": "British Columbia",
        "location": "Fort St. John",
        "capital_cost": "400.00",
        "capital_cost_range": "250 - 500",
        "sector": "Energy",
        "status": "Planned",
        "clean_technology": "Yes",
        "clean_technology_type": "Wind",
        "OBJECTID": 2
      },
      "geometry": {"x": -120.8475, "y": 56.2492}
    },
    {
      "attributes": {
        "id": "0021",
        "company": "Champion Iron Limited",
        "project_name": "Kamistiatusset (Kami)",
        "province": "Newfoundland and Labrador",
        "location": "Labrador West",
        "capital_cost": "3,864.00",
        "capital_cost_range": "2,500 - 5,000",
        "sector": "Mining",
        "status": "Under Construction",
        "clean_technology": "No",
        "clean_technology_type": "N/A",
        "OBJECTID": 3
      },
      "geometry": {"x": -67.0269, "y": 52.8240}
    }
  ]
}`

func TestParseSnapshotProducesSourceGroundedProjects(t *testing.T) {
	adapter := NewNRCanAdapter("unused.json")
	result, err := adapter.Parse([]byte(testSnapshot))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := len(result.Projects); got != 2 {
		t.Fatalf("projects = %d, want 2", got)
	}
	wind := result.Projects[0]
	if wind.Province != "BC" || wind.Sector != domain.SectorCleanEnergy || wind.CurrentStage != domain.StageAnnounced {
		t.Fatalf("wind normalization = province %q sector %q stage %q", wind.Province, wind.Sector, wind.CurrentStage)
	}
	if wind.CapexCAD != 400_000_000 || wind.CapexStatus != domain.ConfidenceReported {
		t.Fatalf("wind capex = %d (%s)", wind.CapexCAD, wind.CapexStatus)
	}
	if wind.Confidence != domain.ConfidenceReported || len(wind.EvidenceIDs) != 1 {
		t.Fatalf("wind provenance not explicitly reported: %#v", wind)
	}
	mining := result.Projects[1]
	if mining.Province != "NL" || mining.Sector != domain.SectorMiningMetals || mining.CapexCAD != 3_864_000_000 {
		t.Fatalf("mining normalization = province %q sector %q capex %d", mining.Province, mining.Sector, mining.CapexCAD)
	}

	second, err := adapter.Parse([]byte(testSnapshot))
	if err != nil {
		t.Fatal(err)
	}
	if second.Projects[0].ID != wind.ID || second.Evidence[0].ID != result.Evidence[0].ID {
		t.Fatal("stable source input produced unstable identifiers")
	}
}

func TestLiveAdapterRejectsNonOfficialEndpoint(t *testing.T) {
	if _, err := NewLiveNRCanAdapter(nil, "https://example.com/query"); err == nil {
		t.Fatal("expected non-official endpoint to be rejected")
	}
	if _, err := NewLiveNRCanAdapter(nil, "http://maps-cartes.services.geo.ca/server_serveur/rest/services/NRCan/major_projects_inventory_en/MapServer/0/query"); err == nil {
		t.Fatal("expected non-HTTPS endpoint to be rejected")
	}
}

func TestFetchLiveUsesBoundedOfficialRequest(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() != "maps-cartes.services.geo.ca" {
			t.Fatalf("unexpected host %q", req.URL.Hostname())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"features":[]}`)),
			Request:    req,
		}, nil
	})}
	adapter, err := NewLiveNRCanAdapter(client, DefaultEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := adapter.Fetch(context.Background()); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if adapter.Health().Mode != "LIVE" || adapter.Health().LastSuccess.IsZero() {
		t.Fatalf("unexpected health: %#v", adapter.Health())
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }
