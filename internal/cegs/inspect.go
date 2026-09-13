package cegs

import (
	"encoding/json"
	"fmt"
)

// InspectionReport summarizes key profile metrics and data quality of a CEGS resource.
type InspectionReport struct {
	CEGSVersion      string   `json:"cegs_version"`
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	CanonicalName    string   `json:"canonical_name"`
	Stage            string   `json:"stage,omitempty"`
	EvidenceCount    int      `json:"evidence_count"`
	ConformanceLevel string   `json:"conformance_level"`
	Status           string   `json:"status,omitempty"`
	UnknownFields    int      `json:"unknown_fields"`
	Extensions       []string `json:"extensions"`
	SummaryText      string   `json:"summary_text"`
}

// Inspect parses a CEGS document and extracts an investor-grade trust profile and conformance summary.
func Inspect(data []byte) (*InspectionReport, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	ver, _ := raw["cegs"].(string)
	id, _ := raw["id"].(string)
	resType, _ := raw["type"].(string)
	name, _ := raw["canonical_name"].(string)
	stage, _ := raw["stage"].(string)
	status, _ := raw["source_status"].(string)

	evidenceCount := 0
	if prov, ok := raw["provenance"].([]interface{}); ok {
		evidenceCount = len(prov)
	}
	if ev, ok := raw["evidence"].([]interface{}); ok {
		evidenceCount = len(ev)
	}

	conformance := "CEGS Core"
	if evidenceCount > 0 {
		conformance = "CEGS Provenance"
	}
	if resType == "event" {
		conformance = "CEGS Historical"
	}

	var extensions []string
	if ext, ok := raw["extensions"].(map[string]interface{}); ok {
		for k := range ext {
			extensions = append(extensions, k)
		}
	}

	summary := fmt.Sprintf("CEGS %s | Type: %s | Name: %s", ver, resType, name)
	if stage != "" {
		summary += fmt.Sprintf(" | Stage: %s", stage)
	}
	summary += fmt.Sprintf(" | Evidence: %d | Conformance: %s", evidenceCount, conformance)

	return &InspectionReport{
		CEGSVersion:      ver,
		ID:               id,
		Type:             resType,
		CanonicalName:    name,
		Stage:            stage,
		EvidenceCount:    evidenceCount,
		ConformanceLevel: conformance,
		Status:           status,
		Extensions:       extensions,
		SummaryText:      summary,
	}, nil
}
