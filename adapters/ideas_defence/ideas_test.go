package ideas_defence

import "testing"

func TestParseUsesStableChallengeIDs(t *testing.T) {
	payload := []byte(`[{
  "challenge_id":"IDEAS-2026-01",
  "title":"Arctic logistics challenge",
  "publisher":"Department of National Defence",
  "province":"NU",
  "location":"Nunavut",
  "sector":"Defence & Arctic",
  "subsector":"Logistics",
  "stage":"DISCOVERED",
  "source_url":"https://canada.ca/example/IDEAS-2026-01",
  "summary":"A public unclassified challenge."
}]`)
	adapter := NewIDEaSAdapter("")
	first, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if first.Projects[0].ID != second.Projects[0].ID || first.Evidence[0].ID != second.Evidence[0].ID {
		t.Fatal("IDEaS identities changed across identical replay")
	}
}
