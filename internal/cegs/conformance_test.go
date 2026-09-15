package cegs_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
)

// conformanceCase describes a single conformance test vector.
type conformanceCase struct {
	name             string
	doc              map[string]interface{}
	expectValid      bool
	expectLevel      string // empty means "don't check"
	expectErrSubstr  string // non-empty means at least one error must contain this substring
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func envelope(resType string, extras map[string]interface{}) map[string]interface{} {
	base := map[string]interface{}{
		"cegs":           cegs.SpecVersion,
		"id":             "cegs:" + resType + ":ca:on:test-" + resType,
		"type":           resType,
		"canonical_name": "Test " + resType,
		"jurisdiction":   "CA:ON",
		"created_at":     time.Now().UTC().Format(time.RFC3339),
		"updated_at":     time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range extras {
		base[k] = v
	}
	return base
}

// ─── test tables ──────────────────────────────────────────────────────────────

func conformanceCases() []conformanceCase {
	now := time.Now().UTC().Format(time.RFC3339)
	provenance := []string{"cegs:evidence:ca:test-evidence-001"}

	return []conformanceCase{
		// ── PROJECT ─────────────────────────────────────────────────────────
		{
			name: "project/core-valid",
			doc: envelope("project", map[string]interface{}{
				"sector": "Clean Energy",
				"stage":  "CONSTRUCTION",
				"capex":  map[string]interface{}{"amount": 1000000000, "currency": "CAD", "amount_type": "reported"},
				"location": map[string]interface{}{
					"name":     "Windsor",
					"province": "ON",
				},
			}),
			expectValid: true,
			expectLevel: "CEGS Core",
		},
		{
			name: "project/provenance-valid",
			doc: envelope("project", map[string]interface{}{
				"sector":     "Infrastructure",
				"stage":      "PERMITTING",
				"capex":      map[string]interface{}{"amount": 500000000, "currency": "CAD", "amount_type": "estimated"},
				"provenance": provenance,
			}),
			expectValid: true,
			expectLevel: "CEGS Provenance",
		},
		{
			name: "project/missing-sector",
			doc: envelope("project", map[string]interface{}{
				"stage": "PLANNING",
				"capex": map[string]interface{}{"currency": "CAD", "amount_type": "unknown"},
			}),
			expectValid:     false,
			expectErrSubstr: "sector",
		},
		{
			name: "project/missing-capex",
			doc: envelope("project", map[string]interface{}{
				"sector": "Mining",
				"stage":  "PLANNING",
			}),
			expectValid:     false,
			expectErrSubstr: "capex",
		},
		{
			name: "project/wrong-currency",
			doc: envelope("project", map[string]interface{}{
				"sector": "LNG",
				"stage":  "CONSTRUCTION",
				"capex":  map[string]interface{}{"amount": 1000, "currency": "USD", "amount_type": "reported"},
			}),
			expectValid:     false,
			expectErrSubstr: "CAD",
		},
		{
			name: "project/bad-id-format",
			doc: map[string]interface{}{
				"cegs":           cegs.SpecVersion,
				"id":             "not-a-cegs-uri",
				"type":           "project",
				"canonical_name": "Bad ID Project",
				"sector":         "Clean Energy",
				"stage":          "PLANNING",
				"capex":          map[string]interface{}{"currency": "CAD", "amount_type": "unknown"},
				"created_at":     now,
				"updated_at":     now,
			},
			expectValid:     false,
			expectErrSubstr: "CEGS ID URI",
		},

		// ── ORGANIZATION ────────────────────────────────────────────────────
		{
			name: "organization/core-valid",
			doc: envelope("organization", map[string]interface{}{
				"entity_type": "CORPORATE",
				"legal_name":  "Acme Energy Corp.",
			}),
			expectValid: true,
			expectLevel: "CEGS Core",
		},
		{
			name: "organization/missing-entity-type",
			doc: envelope("organization", map[string]interface{}{
				"legal_name": "No Type Corp.",
			}),
			expectValid:     false,
			expectErrSubstr: "entity_type",
		},

		// ── EVENT ───────────────────────────────────────────────────────────
		{
			name: "event/historical-valid",
			doc: map[string]interface{}{
				"cegs":        cegs.SpecVersion,
				"id":          "cegs:event:ca:on:test-event-001",
				"type":        "event",
				"event_type":  "stage_change",
				"subject":     "cegs:project:ca:on:test-project",
				"occurred_at": now,
				"title":       "Project moved to CONSTRUCTION",
				"description": "FID achieved, shovels in the ground.",
				"evidence":    provenance,
				"created_at":  now,
			},
			expectValid: true,
			expectLevel: "CEGS Historical",
		},
		{
			name: "event/missing-event-type",
			doc: map[string]interface{}{
				"cegs":        cegs.SpecVersion,
				"id":          "cegs:event:ca:on:bad-event",
				"type":        "event",
				"subject":     "cegs:project:ca:on:test",
				"occurred_at": now,
				"title":       "Missing event_type",
				"description": "Should fail",
				"evidence":    provenance,
				"created_at":  now,
			},
			expectValid:     false,
			expectErrSubstr: "event_type",
		},
		{
			name: "event/missing-evidence",
			doc: map[string]interface{}{
				"cegs":        cegs.SpecVersion,
				"id":          "cegs:event:ca:on:no-evidence",
				"type":        "event",
				"event_type":  "financial_close",
				"subject":     "cegs:project:ca:on:test",
				"occurred_at": now,
				"title":       "No evidence",
				"description": "Should fail",
				"created_at":  now,
			},
			expectValid:     false,
			expectErrSubstr: "evidence",
		},
		{
			name: "event/bad-occurred-at",
			doc: map[string]interface{}{
				"cegs":        cegs.SpecVersion,
				"id":          "cegs:event:ca:on:bad-time",
				"type":        "event",
				"event_type":  "stage_change",
				"subject":     "cegs:project:ca:on:test",
				"occurred_at": "not-a-timestamp",
				"title":       "Bad timestamp",
				"description": "Should fail",
				"evidence":    provenance,
				"created_at":  now,
			},
			expectValid:     false,
			expectErrSubstr: "occurred_at",
		},

		// ── RELATIONSHIP ─────────────────────────────────────────────────────
		{
			name: "relationship/core-valid",
			doc: map[string]interface{}{
				"cegs":              cegs.SpecVersion,
				"id":                "cegs:relationship:ca:on:test-rel-001",
				"type":              "relationship",
				"relationship_type": "PROPONENT",
				"from":              "cegs:org:ca:acme-energy",
				"to":                "cegs:project:ca:on:test-project",
				"status":            "ACTIVE",
				"valid_from":        now,
				"created_at":        now,
			},
			expectValid: true,
			expectLevel: "CEGS Core",
		},
		{
			name: "relationship/missing-relationship-type",
			doc: map[string]interface{}{
				"cegs":       cegs.SpecVersion,
				"id":         "cegs:relationship:ca:on:bad-rel",
				"type":       "relationship",
				"from":       "cegs:org:ca:acme",
				"to":         "cegs:project:ca:on:test",
				"created_at": now,
			},
			expectValid:     false,
			expectErrSubstr: "relationship_type",
		},
		{
			name: "relationship/missing-from",
			doc: map[string]interface{}{
				"cegs":              cegs.SpecVersion,
				"id":                "cegs:relationship:ca:on:no-from",
				"type":              "relationship",
				"relationship_type": "LENDER",
				"to":                "cegs:project:ca:on:test",
				"created_at":        now,
			},
			expectValid:     false,
			expectErrSubstr: "from",
		},

		// ── EVIDENCE ─────────────────────────────────────────────────────────
		{
			name: "evidence/provenance-valid",
			doc: map[string]interface{}{
				"cegs":                 cegs.SpecVersion,
				"id":                   "cegs:evidence:ca:iaac-test-001",
				"type":                 "evidence",
				"source_url":           "https://example.gc.ca/dataset/001",
				"publisher":            "IAAC",
				"source_tier":          1,
				"content_hash":         "aabbccdd" + "aabbccdd" + "aabbccdd" + "aabbccdd" + "aabbccdd" + "aabbccdd" + "aabbccdd" + "aabbccd0",
				"retrieval_timestamp":  now,
				"confidence":           "verified",
				"extraction_method":    "deterministic_adapter",
			},
			expectValid: true,
			expectLevel: "CEGS Provenance",
		},
		{
			name: "evidence/bad-hash",
			doc: map[string]interface{}{
				"cegs":                "cegs:evidence:ca:bad-hash",
				"id":                  "cegs:evidence:ca:bad-hash",
				"type":                "evidence",
				"source_url":          "https://example.gc.ca/dataset/002",
				"publisher":           "IAAC",
				"source_tier":         1,
				"content_hash":        "not-a-valid-sha256",
				"retrieval_timestamp": now,
				"confidence":          "high",
				"extraction_method":   "manual",
			},
			expectValid:     false,
			expectErrSubstr: "content_hash",
		},
		{
			name: "evidence/tier-out-of-range",
			doc: map[string]interface{}{
				"cegs":                cegs.SpecVersion,
				"id":                  "cegs:evidence:ca:bad-tier",
				"type":                "evidence",
				"source_url":          "https://example.gc.ca/dataset/003",
				"publisher":           "IAAC",
				"source_tier":         5,
				"content_hash":        "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd",
				"retrieval_timestamp": now,
				"confidence":          "high",
				"extraction_method":   "deterministic_adapter",
			},
			expectValid:     false,
			expectErrSubstr: "source_tier",
		},

		// ── SOURCE ────────────────────────────────────────────────────────────
		{
			name: "source/core-valid",
			doc: envelope("source", map[string]interface{}{
				"source_kind":   "DATASET",
				"canonical_url": "https://open.canada.ca/data/en/dataset/test",
				"source_family": "federal-opendata",
				"access_method": "REST_API",
				"authority_tier": 1,
				"lifecycle":     "ACTIVE",
				"health":        "HEALTHY",
			}),
			expectValid: true,
			expectLevel: "CEGS Core",
		},
		{
			name: "source/invalid-lifecycle",
			doc: envelope("source", map[string]interface{}{
				"source_kind":   "FEED",
				"canonical_url": "https://example.gc.ca/feed",
				"source_family": "gazette",
				"access_method": "SCRAPE",
				"authority_tier": 2,
				"lifecycle":     "UNKNOWN_INVALID",
				"health":        "HEALTHY",
			}),
			expectValid:     false,
			expectErrSubstr: "lifecycle",
		},

		// ── VERSION CHECK ─────────────────────────────────────────────────────
		{
			name: "envelope/wrong-cegs-version",
			doc: map[string]interface{}{
				"cegs":           "0.0.1",
				"id":             "cegs:project:ca:on:test",
				"type":           "project",
				"canonical_name": "Old version",
				"sector":         "Infrastructure",
				"stage":          "PLANNING",
				"capex":          map[string]interface{}{"currency": "CAD", "amount_type": "unknown"},
				"created_at":     now,
				"updated_at":     now,
			},
			expectValid:     false,
			expectErrSubstr: "version",
		},
		{
			name: "envelope/missing-id",
			doc: map[string]interface{}{
				"cegs":           cegs.SpecVersion,
				"type":           "project",
				"canonical_name": "No ID",
				"sector":         "Infrastructure",
				"stage":          "PLANNING",
				"capex":          map[string]interface{}{"currency": "CAD", "amount_type": "unknown"},
				"created_at":     now,
				"updated_at":     now,
			},
			expectValid:     false,
			expectErrSubstr: "'id'",
		},
	}
}

// ─── conformance runner ───────────────────────────────────────────────────────

func TestCEGS10ConformanceSuite(t *testing.T) {
	for _, tc := range conformanceCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data := mustMarshal(t, tc.doc)
			report, err := cegs.Validate(data)
			if err != nil {
				t.Fatalf("unexpected Validate error: %v", err)
			}

			if report.Valid != tc.expectValid {
				t.Errorf("Valid=%v, want %v; errors=%v", report.Valid, tc.expectValid, report.Errors)
			}
			if tc.expectLevel != "" && report.ConformanceLevel != tc.expectLevel {
				t.Errorf("ConformanceLevel=%q, want %q", report.ConformanceLevel, tc.expectLevel)
			}
			if tc.expectErrSubstr != "" {
				found := false
				for _, e := range report.Errors {
					if containsCI(e, tc.expectErrSubstr) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got: %v", tc.expectErrSubstr, report.Errors)
				}
			}
		})
	}
}

// ─── migration round-trip tests ───────────────────────────────────────────────

func TestCEGS10MigrationRoundTrip(t *testing.T) {
	resourceTypes := []string{
		"project", "organization", "event", "relationship",
		"evidence", "source", "signal", "opportunity", "procurement",
	}

	for _, resType := range resourceTypes {
		resType := resType
		t.Run("migration/"+resType, func(t *testing.T) {
			// Build a minimal legacy JSON Schema for the resource type.
			legacyID := "cegs:" + resType + ":ca:legacy-schema-" + resType
			legacy := map[string]interface{}{
				"$id":     legacyID,
				"$schema": "https://json-schema.org/draft/2020-12/schema",
				"type":    "object",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string"},
					"type": map[string]interface{}{"type": "string", "const": resType},
				},
				"required": []string{"id", "type"},
			}
			raw, err := json.Marshal(legacy)
			if err != nil {
				t.Fatalf("marshal legacy schema: %v", err)
			}

			migrated, err := cegs.MigrateSchema(raw)
			if err != nil {
				t.Fatalf("MigrateSchema failed: %v", err)
			}

			// Validate the migrated schema struct.
			if err := cegs.ValidateMigrated(migrated); err != nil {
				t.Errorf("ValidateMigrated failed: %v", err)
			}

			// Check canonical vocabulary.
			if len(migrated.Vocabulary) == 0 || migrated.Vocabulary[0] != "cegs-1.0" {
				t.Errorf("expected vocabulary [cegs-1.0], got %v", migrated.Vocabulary)
			}

			// Check migration provenance is preserved.
			if migrated.MigratedFrom != legacyID {
				t.Errorf("expected MigratedFrom=%q, got %q", legacyID, migrated.MigratedFrom)
			}

			// MigrationNote must mention the toolkit version.
			if !containsCI(migrated.MigrationNote, cegs.MigrationVersion) {
				t.Errorf("MigrationNote %q does not reference migration version %q", migrated.MigrationNote, cegs.MigrationVersion)
			}

			// Properties from the legacy schema must survive.
			if _, ok := migrated.Properties["id"]; !ok {
				t.Errorf("migrated schema lost 'id' property")
			}
		})
	}
}

// ─── nil / corrupt input tests ────────────────────────────────────────────────

func TestCEGS10CorruptInputs(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"not-json", []byte("not json at all")},
		{"empty-object", []byte("{}")},
		{"null", []byte("null")},
		{"array", []byte("[1, 2, 3]")},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			report, err := cegs.Validate(tc.input)
			// Validate must never return a Go error — it returns a ValidationReport.
			if err != nil {
				t.Fatalf("Validate returned unexpected Go error: %v", err)
			}
			if report == nil {
				t.Fatal("Validate returned nil report")
			}
			// Corrupt / empty inputs should always be invalid.
			if report.Valid {
				t.Errorf("expected invalid for input %q, got valid", tc.input)
			}
		})
	}
}

// ─── utils ────────────────────────────────────────────────────────────────────

// containsCI is a case-insensitive contains check.
func containsCI(s, substr string) bool {
	// Use a simple fold: convert both to lowercase.
	lower := func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}
	sl := applyRune(s, lower)
	subl := applyRune(substr, lower)
	return len(sl) >= len(subl) && contains(sl, subl)
}

func applyRune(s string, fn func(rune) rune) string {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = fn(r)
	}
	return string(runes)
}

func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
