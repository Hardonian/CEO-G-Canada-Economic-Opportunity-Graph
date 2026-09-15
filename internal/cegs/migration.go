// Package cegs provides the CEGS 1.0 migration toolkit. It converts legacy
// JSON Schemas (Draft 2020-12) into the canonical CEGS 1.0 vocabulary and
// validates that migrated records conform to the locked specification.
package cegs

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MigrationVersion identifies the migration toolkit.
const MigrationVersion = "cegs-migration-v1.0"

// LegacySchema represents a Draft 2020-12 JSON Schema that may contain
// deprecated vocabulary or non-canonical identifiers.
type LegacySchema struct {
	ID         string                 `json:"$id"`
	Schema     string                 `json:"$schema"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

// MigratedSchema is the canonical CEGS 1.0 representation.
type MigratedSchema struct {
	ID            string                 `json:"id"`
	Version       string                 `json:"version"`
	Type          string                 `json:"type"`
	Vocabulary    []string               `json:"vocabulary"`
	Properties    map[string]interface{} `json:"properties"`
	Required      []string               `json:"required"`
	MigratedFrom  string                 `json:"migrated_from"`
	MigrationNote string                 `json:"migration_note"`
}

// MigrateSchema converts a legacy JSON Schema into a canonical CEGS 1.0 schema.
func MigrateSchema(raw []byte) (*MigratedSchema, error) {
	var legacy LegacySchema
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return nil, fmt.Errorf("parse legacy schema: %w", err)
	}

	migrated := &MigratedSchema{
		ID:           strings.TrimPrefix(legacy.ID, "cegs:"),
		Version:      SpecVersion,
		Type:         legacy.Type,
		Vocabulary:   []string{"cegs-1.0"},
		Properties:   legacy.Properties,
		Required:     legacy.Required,
		MigratedFrom: legacy.ID,
		MigrationNote: fmt.Sprintf("Migrated from %s using %s. Canonical vocabulary is cegs-1.0.", legacy.ID, MigrationVersion),
	}

	if migrated.ID == "" {
		migrated.ID = "cegs:unknown"
	}
	if migrated.Type == "" {
		migrated.Type = "object"
	}

	return migrated, nil
}

// ValidateMigrated checks that a migrated schema conforms to the CEGS 1.0 spec.
func ValidateMigrated(m *MigratedSchema) error {
	if m == nil {
		return fmt.Errorf("migrated schema is nil")
	}
	if m.Version != SpecVersion {
		return fmt.Errorf("version %q does not match CEGS 1.0 (%q)", m.Version, SpecVersion)
	}
	if m.ID == "" || m.ID == "cegs:unknown" {
		return fmt.Errorf("migrated schema has no valid id")
	}
	return nil
}