package iaac

import "testing"

func TestParseUsesStableRegistryEventIDs(t *testing.T) {
	payload := []byte(`[{
  "registry_id":"IAAC-83857",
  "project_name":"Example Critical Minerals Project",
  "proponent_name":"Example Mining Inc.",
  "province":"ON",
  "region":"Northern Ontario",
  "sector":"Critical Minerals",
  "subsector":"Nickel",
  "current_status":"PERMITTING",
  "last_registry_update":"2026-09-01T12:00:00Z",
  "registry_url":"https://iaac-aeic.gc.ca/example/83857",
  "summary":"A public registry test record."
}]`)
	adapter := NewIAACAdapter("")
	first, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	second, err := adapter.Parse(payload)
	if err != nil {
		t.Fatal(err)
	}
	if first.Projects[0].ID != second.Projects[0].ID || first.Events[0].ID != second.Events[0].ID || first.Relationships[0].ID != second.Relationships[0].ID {
		t.Fatal("IAAC identities changed across identical replay")
	}
}
