package identity

import "testing"

func TestSlugPreservesReadableFrenchNames(t *testing.T) {
	tests := map[string]string{
		"Contrecœur Port — Québec": "contrecoeur-port-quebec",
		"Énergie Côte-Nord":        "energie-cote-nord",
		"Îles-de-la-Madeleine":     "iles-de-la-madeleine",
	}
	for input, want := range tests {
		if got := Slug(input); got != want {
			t.Errorf("Slug(%q) = %q, want %q", input, got, want)
		}
	}
}
