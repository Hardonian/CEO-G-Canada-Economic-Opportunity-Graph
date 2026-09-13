package cegs

import (
	"encoding/json"
	"fmt"
)

// SemanticChange represents a typed economic or factual difference between two CEGS documents.
type SemanticChange struct {
	Code        string      `json:"code"` // e.g. PROJECT_STAGE_CHANGED, CAPEX_CHANGED
	Field       string      `json:"field"`
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Description string      `json:"description"`
}

// DiffReport holds all semantic changes identified between two CEGS states.
type DiffReport struct {
	ResourceID   string            `json:"resource_id"`
	ResourceType string            `json:"resource_type"`
	HasChanges   bool              `json:"has_changes"`
	Changes      []*SemanticChange `json:"changes"`
}

// Diff compares two CEGS JSON states and detects meaningful economic transitions.
func Diff(oldData, newData []byte) (*DiffReport, error) {
	var oldDoc, newDoc map[string]interface{}
	if err := json.Unmarshal(oldData, &oldDoc); err != nil {
		return nil, fmt.Errorf("failed to parse old JSON: %w", err)
	}
	if err := json.Unmarshal(newData, &newDoc); err != nil {
		return nil, fmt.Errorf("failed to parse new JSON: %w", err)
	}

	report := &DiffReport{
		ResourceID:   fmt.Sprintf("%v", newDoc["id"]),
		ResourceType: fmt.Sprintf("%v", newDoc["type"]),
		Changes:      make([]*SemanticChange, 0),
	}

	// 1. Check Canonical Name Change
	oldName, _ := oldDoc["canonical_name"].(string)
	newName, _ := newDoc["canonical_name"].(string)
	if oldName != newName && newName != "" {
		report.Changes = append(report.Changes, &SemanticChange{
			Code:        "NAME_CHANGED",
			Field:       "canonical_name",
			OldValue:    oldName,
			NewValue:    newName,
			Description: fmt.Sprintf("Canonical name updated from '%s' to '%s'", oldName, newName),
		})
	}

	// 2. Check Stage Transitions
	oldStage, _ := oldDoc["stage"].(string)
	newStage, _ := newDoc["stage"].(string)
	if oldStage != newStage && newStage != "" {
		report.Changes = append(report.Changes, &SemanticChange{
			Code:        "PROJECT_STAGE_CHANGED",
			Field:       "stage",
			OldValue:    oldStage,
			NewValue:    newStage,
			Description: fmt.Sprintf("Lifecycle stage progressed from %s to %s", oldStage, newStage),
		})
	}

	// 3. Check CAPEX Revisions
	oldCapex, _ := extractCapex(oldDoc)
	newCapex, _ := extractCapex(newDoc)
	if oldCapex != newCapex {
		report.Changes = append(report.Changes, &SemanticChange{
			Code:        "CAPEX_CHANGED",
			Field:       "capex.amount",
			OldValue:    oldCapex,
			NewValue:    newCapex,
			Description: fmt.Sprintf("Reported CAPEX changed from $%d CAD to $%d CAD", oldCapex, newCapex),
		})
	}

	// 4. Check Status Changes
	oldStatus, _ := oldDoc["source_status"].(string)
	newStatus, _ := newDoc["source_status"].(string)
	if oldStatus != newStatus && newStatus != "" {
		report.Changes = append(report.Changes, &SemanticChange{
			Code:        "STATUS_CHANGED",
			Field:       "source_status",
			OldValue:    oldStatus,
			NewValue:    newStatus,
			Description: fmt.Sprintf("Verification confidence updated from %s to %s", oldStatus, newStatus),
		})
	}

	// 5. Check Evidence Expansion
	oldProv, _ := oldDoc["provenance"].([]interface{})
	newProv, _ := newDoc["provenance"].([]interface{})
	if len(newProv) > len(oldProv) {
		report.Changes = append(report.Changes, &SemanticChange{
			Code:        "EVIDENCE_ADDED",
			Field:       "provenance",
			OldValue:    len(oldProv),
			NewValue:    len(newProv),
			Description: fmt.Sprintf("Added %d new corroborating evidence reference(s)", len(newProv)-len(oldProv)),
		})
	}

	report.HasChanges = len(report.Changes) > 0
	return report, nil
}

func extractCapex(doc map[string]interface{}) (int64, bool) {
	capex, ok := doc["capex"].(map[string]interface{})
	if !ok {
		return 0, false
	}
	amt, ok := capex["amount"].(float64)
	if !ok {
		return 0, false
	}
	return int64(amt), true
}
