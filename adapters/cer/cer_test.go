package cer

import "testing"

func TestParseUsesStableFilingIDs(t *testing.T) {
	payload := []byte(`[{
  "filing_id":"CER-42",
  "facility_name":"Northern Grid Link",
  "proponent_name":"Example Utility",
  "province":"ON",
  "sector":"Clean Energy & Grid",
  "subsector":"Transmission",
  "stage":"PERMITTING",
  "filing_url":"https://www.cer-rec.gc.ca/example/CER-42",
  "summary":"A public test filing."
}]`)
	adapter := NewCERAdapter("")
	first, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if first.Projects[0].ID != second.Projects[0].ID || first.Entities[0].ID != second.Entities[0].ID || first.Evidence[0].ID != second.Evidence[0].ID {
		t.Fatal("CER identities changed across identical replay")
	}
}
