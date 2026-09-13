package cegs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	uriRegex  = regexp.MustCompile(`^cegs:[a-z0-9_-]+:[a-z0-9_-]+(:[a-z0-9_-]+)*$`)
	hashRegex = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// ValidationReport contains comprehensive structural and semantic audit results.
type ValidationReport struct {
	Valid            bool     `json:"valid"`
	ResourceType     string   `json:"resource_type"`
	ID               string   `json:"id"`
	ConformanceLevel string   `json:"conformance_level"` // Core, Provenance, Historical, Intelligence
	Errors           []string `json:"errors"`
	Warnings         []string `json:"warnings"`
	Summary          string   `json:"summary"`
}

// Validate analyzes a raw JSON byte slice against CEGS 0.1 normative invariants.
func Validate(data []byte) (*ValidationReport, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return &ValidationReport{
			Valid:   false,
			Errors:  []string{fmt.Sprintf("Invalid JSON: %v", err)},
			Summary: "FAILED: Payload is not valid JSON.",
		}, nil
	}

	report := &ValidationReport{
		Valid:            true,
		Errors:           make([]string, 0),
		Warnings:         make([]string, 0),
		ConformanceLevel: "CEGS Core",
	}

	// 1. Check Envelope Requirements
	cegsVer, _ := raw["cegs"].(string)
	if cegsVer != SpecVersion {
		report.Errors = append(report.Errors, fmt.Sprintf("Unsupported or missing CEGS version '%s', expected '%s'", cegsVer, SpecVersion))
	}

	id, _ := raw["id"].(string)
	if id == "" {
		report.Errors = append(report.Errors, "Missing required field 'id'")
	} else if !uriRegex.MatchString(id) {
		report.Errors = append(report.Errors, fmt.Sprintf("Invalid CEGS ID URI format: '%s'", id))
	}
	report.ID = id

	resType, _ := raw["type"].(string)
	if resType == "" {
		report.Errors = append(report.Errors, "Missing required field 'type'")
	}
	report.ResourceType = resType

	// Types other than event, relationship, evidence, score, signal, manifest require canonical_name
	if resType != "event" && resType != "relationship" && resType != "evidence" && resType != "score" && resType != "signal" && resType != "manifest" {
		name, _ := raw["canonical_name"].(string)
		if strings.TrimSpace(name) == "" {
			report.Errors = append(report.Errors, "Missing required field 'canonical_name'")
		}
	}

	// 2. Resource-Specific Checks
	switch resType {
	case "project":
		validateProject(raw, report)
	case "organization":
		validateOrganization(raw, report)
	case "event":
		validateEvent(raw, report)
	case "relationship":
		validateRelationship(raw, report)
	case "evidence":
		validateEvidence(raw, report)
	case "manifest":
		validateManifest(raw, report)
	case "source":
		validateSource(raw, report)
	}

	// 3. Determine Conformance Level
	if len(report.Errors) > 0 {
		report.Valid = false
		report.Summary = fmt.Sprintf("FAILED: %d validation error(s) detected.", len(report.Errors))
	} else {
		report.Summary = fmt.Sprintf("PASSED: Conforms to %s.", report.ConformanceLevel)
	}

	return report, nil
}

func validateSource(raw map[string]interface{}, rep *ValidationReport) {
	kind, _ := raw["source_kind"].(string)
	if !oneOf(kind, "CATALOG", "DATASET", "RESOURCE", "API", "FEED", "DOCUMENT_REPOSITORY", "DOCUMENT", "WEB_PAGE") {
		rep.Errors = append(rep.Errors, "Source has an invalid or missing 'source_kind'")
	}
	canonicalURL, _ := raw["canonical_url"].(string)
	parsed, err := url.Parse(canonicalURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		rep.Errors = append(rep.Errors, "Source 'canonical_url' must be an absolute HTTP(S) URL")
	}
	for _, field := range []string{"source_family", "access_method"} {
		if value, _ := raw[field].(string); strings.TrimSpace(value) == "" {
			rep.Errors = append(rep.Errors, fmt.Sprintf("Source missing '%s'", field))
		}
	}
	authority, ok := raw["authority_tier"].(float64)
	if !ok || authority < 1 || authority > 5 || authority != float64(int(authority)) {
		rep.Errors = append(rep.Errors, "Source 'authority_tier' must be an integer between 1 and 5")
	}
	lifecycle, _ := raw["lifecycle"].(string)
	if !oneOf(lifecycle, "DISCOVERED", "CLASSIFIED", "TESTED", "APPROVED", "ACTIVE", "REJECTED", "BLOCKED", "RETIRED") {
		rep.Errors = append(rep.Errors, "Source has an invalid or missing 'lifecycle'")
	}
	health, _ := raw["health"].(string)
	if !oneOf(health, "UNKNOWN", "HEALTHY", "STALE", "DEGRADED", "BROKEN", "DISABLED") {
		rep.Errors = append(rep.Errors, "Source has an invalid or missing 'health'")
	}
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func validateProject(raw map[string]interface{}, rep *ValidationReport) {
	sector, _ := raw["sector"].(string)
	if sector == "" {
		rep.Errors = append(rep.Errors, "Project missing 'sector'")
	}

	stage, _ := raw["stage"].(string)
	if stage == "" {
		rep.Errors = append(rep.Errors, "Project missing 'stage'")
	}

	capex, ok := raw["capex"].(map[string]interface{})
	if !ok {
		rep.Errors = append(rep.Errors, "Project missing structured 'capex' object")
	} else {
		curr, _ := capex["currency"].(string)
		if curr != "CAD" {
			rep.Errors = append(rep.Errors, fmt.Sprintf("Expected capex.currency 'CAD', got '%s'", curr))
		}
	}

	prov, ok := raw["provenance"].([]interface{})
	if ok && len(prov) > 0 {
		rep.ConformanceLevel = "CEGS Provenance"
	} else {
		rep.Warnings = append(rep.Warnings, "Project has no provenance evidence references")
	}
}

func validateOrganization(raw map[string]interface{}, rep *ValidationReport) {
	eType, _ := raw["entity_type"].(string)
	if eType == "" {
		rep.Errors = append(rep.Errors, "Organization missing 'entity_type'")
	}
}

func validateEvent(raw map[string]interface{}, rep *ValidationReport) {
	eType, _ := raw["event_type"].(string)
	if eType == "" {
		rep.Errors = append(rep.Errors, "Event missing 'event_type'")
	}
	subj, _ := raw["subject"].(string)
	if subj == "" {
		rep.Errors = append(rep.Errors, "Event missing 'subject'")
	}
	occ, _ := raw["occurred_at"].(string)
	if occ == "" {
		rep.Errors = append(rep.Errors, "Event missing 'occurred_at'")
	} else if _, err := time.Parse(time.RFC3339, occ); err != nil {
		rep.Errors = append(rep.Errors, fmt.Sprintf("Event occurred_at must be RFC3339 timestamp: %v", err))
	}

	evidence, ok := raw["evidence"].([]interface{})
	if !ok || len(evidence) == 0 {
		rep.Errors = append(rep.Errors, "Event requires at least 1 evidence reference")
	} else {
		rep.ConformanceLevel = "CEGS Historical"
	}
}

func validateRelationship(raw map[string]interface{}, rep *ValidationReport) {
	relType, _ := raw["relationship_type"].(string)
	if relType == "" {
		rep.Errors = append(rep.Errors, "Relationship missing 'relationship_type'")
	}
	from, _ := raw["from"].(string)
	if from == "" {
		rep.Errors = append(rep.Errors, "Relationship missing 'from'")
	}
	to, _ := raw["to"].(string)
	if to == "" {
		rep.Errors = append(rep.Errors, "Relationship missing 'to'")
	}
}

func validateEvidence(raw map[string]interface{}, rep *ValidationReport) {
	sourceURL, _ := raw["source_url"].(string)
	if sourceURL == "" {
		rep.Errors = append(rep.Errors, "Evidence missing 'source_url'")
	}
	publisher, _ := raw["publisher"].(string)
	if publisher == "" {
		rep.Errors = append(rep.Errors, "Evidence missing 'publisher'")
	}
	tier, ok := raw["source_tier"].(float64)
	if !ok || tier < 1 || tier > 4 {
		rep.Errors = append(rep.Errors, "Evidence 'source_tier' must be an integer between 1 and 4")
	}
	hash, _ := raw["content_hash"].(string)
	if !hashRegex.MatchString(hash) {
		rep.Errors = append(rep.Errors, fmt.Sprintf("Evidence 'content_hash' must be 64-character hex SHA-256 string, got '%s'", hash))
	}
	rep.ConformanceLevel = "CEGS Provenance"
}

func validateManifest(raw map[string]interface{}, rep *ValidationReport) {
	datasetID, _ := raw["dataset_id"].(string)
	if datasetID == "" {
		rep.Errors = append(rep.Errors, "Manifest missing 'dataset_id'")
	}
	publisher, _ := raw["publisher"].(string)
	if publisher == "" {
		rep.Errors = append(rep.Errors, "Manifest missing 'publisher'")
	}
	license, _ := raw["license"].(string)
	if license == "" {
		rep.Errors = append(rep.Errors, "Manifest missing 'license'")
	}
}
