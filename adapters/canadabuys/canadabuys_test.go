package canadabuys

import "testing"

func TestParseUsesStableTenderAndEvidenceIDs(t *testing.T) {
	payload := []byte(`[{
  "tender_id":"PW-TEST-001",
  "title":"Grid infrastructure engineering services",
  "buyer":"Public Services and Procurement Canada",
  "buyer_type":"Federal",
  "status":"OPEN",
  "closing_date":"2026-10-01T16:00:00Z",
  "source_url":"https://canadabuys.canada.ca/example/PW-TEST-001",
  "categories":["engineering"]
}]`)
	adapter := NewCanadaBuysAdapter("")
	first, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if first.Procurements[0].ID != second.Procurements[0].ID {
		t.Fatalf("procurement ID changed across replay: %q != %q", first.Procurements[0].ID, second.Procurements[0].ID)
	}
	if first.Evidence[0].ID != second.Evidence[0].ID || first.Evidence[0].SourceRecordID != "PW-TEST-001" {
		t.Fatalf("evidence lineage is not stable: %#v", first.Evidence[0])
	}
}
