package cegs

import (
	"encoding/json"
	"testing"
)

func TestMigrateSchema_NormalisesLegacyFields(t *testing.T) {
	legacy := `{
		"$id": "cegs:legacy:project",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {"name": {"type": "string"}},
		"required": ["name"]
	}`

	migrated, err := MigrateSchema([]byte(legacy))
	if err != nil {
		t.Fatalf("MigrateSchema() error = %v", err)
	}
	if migrated.ID != "legacy:project" {
		t.Errorf("id = %q, want %q", migrated.ID, "legacy:project")
	}
	if migrated.Version != SpecVersion {
		t.Errorf("version = %q, want %q", migrated.Version, SpecVersion)
	}
	if migrated.Type != "object" {
		t.Errorf("type = %q, want object", migrated.Type)
	}
	if len(migrated.Vocabulary) == 0 || migrated.Vocabulary[0] != "cegs-1.0" {
		t.Errorf("vocabulary = %v, want [cegs-1.0]", migrated.Vocabulary)
	}
	if migrated.MigratedFrom != "cegs:legacy:project" {
		t.Errorf("migrated_from = %q, want cegs:legacy:project", migrated.MigratedFrom)
	}
	if migrated.MigrationNote == "" {
		t.Error("migration_note should not be empty")
	}
}

func TestMigrateSchema_InvalidJSON(t *testing.T) {
	_, err := MigrateSchema([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestMigrateSchema_EmptyID(t *testing.T) {
	legacy := `{"type":"object"}`
	migrated, err := MigrateSchema([]byte(legacy))
	if err != nil {
		t.Fatalf("MigrateSchema() error = %v", err)
	}
	if migrated.ID != "cegs:unknown" {
		t.Errorf("id = %q, want cegs:unknown", migrated.ID)
	}
}

func TestMigrateSchema_EmptyType(t *testing.T) {
	legacy := `{"$id":"cegs:test"}`
	migrated, err := MigrateSchema([]byte(legacy))
	if err != nil {
		t.Fatalf("MigrateSchema() error = %v", err)
	}
	if migrated.Type != "object" {
		t.Errorf("type = %q, want object", migrated.Type)
	}
}

func TestValidateMigrated_Valid(t *testing.T) {
	m := &MigratedSchema{
		ID:      "cegs:test",
		Version: SpecVersion,
		Type:    "object",
	}
	if err := ValidateMigrated(m); err != nil {
		t.Errorf("ValidateMigrated() error = %v", err)
	}
}

func TestValidateMigrated_Nil(t *testing.T) {
	if err := ValidateMigrated(nil); err == nil {
		t.Fatal("expected error for nil schema")
	}
}

func TestValidateMigrated_WrongVersion(t *testing.T) {
	m := &MigratedSchema{
		ID:      "cegs:test",
		Version: "0.0",
		Type:    "object",
	}
	if err := ValidateMigrated(m); err == nil {
		t.Fatal("expected error for wrong version")
	}
}

func TestValidateMigrated_EmptyID(t *testing.T) {
	m := &MigratedSchema{
		Version: SpecVersion,
		Type:    "object",
	}
	if err := ValidateMigrated(m); err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestMigrateSchema_PreservesProperties(t *testing.T) {
	legacy := `{
		"$id": "cegs:test:project",
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"capex": {"type": "number"}
		},
		"required": ["name"]
	}`
	migrated, err := MigrateSchema([]byte(legacy))
	if err != nil {
		t.Fatalf("MigrateSchema() error = %v", err)
	}
	if len(migrated.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(migrated.Properties))
	}
	if len(migrated.Required) != 1 || migrated.Required[0] != "name" {
		t.Errorf("required = %v, want [name]", migrated.Required)
	}
}

// Ensure json import is used.
var _ = json.Marshal