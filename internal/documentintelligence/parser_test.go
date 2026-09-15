package documentintelligence

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestParseCapexPreservesSemantics(t *testing.T) {
	tests := []struct {
		raw, currency string
		kind          domain.AmountType
		amount, min, max int64
	}{
		{"$1.9B", "UNSPECIFIED", domain.AmountExact, 1_900_000_000, 0, 0},
		{"US$1.4B", "USD", domain.AmountExact, 1_400_000_000, 0, 0},
		{"US$5B–US$6B", "USD", domain.AmountRange, 0, 5_000_000_000, 6_000_000_000},
		{"approximately US$400M–US$500M", "USD", domain.AmountRange, 0, 400_000_000, 500_000_000},
		{"$250M Phase 1", "UNSPECIFIED", domain.AmountExact, 250_000_000, 0, 0},
		{"$1B+ total", "UNSPECIFIED", domain.AmountMinimum, 1_000_000_000, 1_000_000_000, 0},
		{"Not available", "", domain.AmountNotAvailable, 0, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			got := ParseCapex(test.raw)
			if got.Currency != test.currency || got.AmountType != test.kind || deref(got.Amount) != test.amount || deref(got.Minimum) != test.min || deref(got.Maximum) != test.max {
				t.Fatalf("ParseCapex(%q) = %#v", test.raw, got)
			}
		})
	}
}

func TestMultiCardSegmentationDoesNotCrossAssociate(t *testing.T) {
	fixture := `--- PROJECT ---
Project: Aurora Port
Proponent: North Harbour
Stage: FEED + permitting
CapEx: C$1B Phase 1
Financing Objective: infrastructure equity + project finance
--- PROJECT ---
Project: Prairie Compute
Proponent: Wheatland Digital
Stage: pre-FID + financing
CapEx: Not available
Financing Objective: strategic partner + anchor tenant`
	cards, err := ExtractCards(fixture, "restricted-eval", domain.VisibilityLicensedPrivate, time.Unix(1, 0))
	if err != nil || len(cards) != 2 {
		t.Fatalf("ExtractCards = %#v, %v", cards, err)
	}
	if cards[0].Candidate.Proponent != "North Harbour" || cards[1].Candidate.Proponent != "Wheatland Digital" {
		t.Fatalf("proponents crossed cards: %#v", cards)
	}
	if cards[0].CapitalRequirement.Amount.Currency != "CAD" || cards[1].CapitalRequirement.Amount.AmountType != domain.AmountNotAvailable {
		t.Fatalf("capex crossed cards: %#v", cards)
	}
	if len(cards[0].CapitalNeed.Types) == 0 || len(cards[1].CapitalNeed.Types) == 0 || cards[0].CapitalNeed.Publishable || cards[1].CapitalNeed.Publishable {
		t.Fatalf("capital needs invalid: %#v", cards)
	}
}

func TestPrivateExtractorRejectsPublicVisibility(t *testing.T) {
	if _, err := ExtractCards("Project: X", "source", domain.VisibilityPublic, time.Now()); err == nil {
		t.Fatal("restricted extraction accepted public visibility")
	}
}

func deref(value *int64) int64 {
	if value == nil { return 0 }
	return *value
}
