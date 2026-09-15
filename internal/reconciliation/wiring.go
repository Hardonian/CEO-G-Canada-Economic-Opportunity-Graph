// Package reconciliation provides multi-jurisdiction reconciliation of project
// records across Federal, Provincial, Municipal, and Indigenous sources. It is
// wired into the ingestion pipeline so that cross-source disagreements are
// surfaced as CONFLICT actions rather than silently overwritten.
package reconciliation

import (
	"context"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// Reconcile runs multi-jurisdiction reconciliation over the projects currently
// held in the store. It builds JurisdictionRecord views from each project's
// metadata and external identifiers, then delegates to the deterministic
// ReconcileJurisdictions engine. The report is returned; no store writes are
// performed by this function so callers can decide how to act on conflicts.
func Reconcile(ctx context.Context, store database.Store) *ReconciliationReport {
	projects, _, err := store.ListProjects(ctx, database.ProjectFilter{Limit: 10_000})
	if err != nil {
		return &ReconciliationReport{MethodologyVersion: MethodologyVersion}
	}

	records := make([]*JurisdictionRecord, 0, len(projects))
	for _, p := range projects {
		if p == nil {
			continue
		}
		level := jurisdictionLevel(p)
		records = append(records, &JurisdictionRecord{
			ProjectID:     p.ID,
			Jurisdiction:  string(level),
			Level:         level,
			ExternalID:    firstExternalID(p.ExternalIDs),
			SourceURL:     firstSourceURL(p.ExternalIDs),
			Stage:         p.CurrentStage,
			Title:         p.Name,
			Summary:       p.Summary,
			Location:      p.LocationName,
			CapexCAD:      p.CapexCAD,
			EffectiveDate: formatEffectiveDate(p.LastMeaningfulUpdate),
			Confidence:    p.Confidence,
			Metadata:      copyMetadata(p.Metadata),
		})
	}

	return ReconcileJurisdictions(records)
}

// ReconcileRecords is a thin wrapper around ReconcileJurisdictions for callers
// who already have a slice of jurisdiction records.
func ReconcileRecords(records []*JurisdictionRecord) *ReconciliationReport {
	return ReconcileJurisdictions(records)
}

func jurisdictionLevel(project *domain.Project) JurisdictionLevel {
	if project == nil {
		return JurisdictionFederal
	}
	switch strings.ToUpper(strings.TrimSpace(project.Province)) {
	case "Federal", "FEDERAL":
		return JurisdictionFederal
	default:
		if strings.Contains(strings.ToLower(project.Province), "municipal") ||
			strings.Contains(strings.ToLower(project.LocationName), "municipal") {
			return JurisdictionMunicipal
		}
		if hasIndigenousTheme(project) {
			return JurisdictionIndigenous
		}
		// Provincial codes are the common case; everything else is treated as
		// provincial so the reconciliation engine still gets a chance to match.
		return JurisdictionProvincial
	}
}

func hasIndigenousTheme(project *domain.Project) bool {
	for _, theme := range project.StrategicThemes {
		if theme == domain.ThemeIndigenousOwnership {
			return true
		}
	}
	return false
}

func firstExternalID(externalIDs map[string]string) string {
	if len(externalIDs) == 0 {
		return ""
	}
	for _, value := range externalIDs {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstSourceURL(externalIDs map[string]string) string {
	if len(externalIDs) == 0 {
		return ""
	}
	for key, value := range externalIDs {
		if strings.HasSuffix(strings.ToLower(key), "url") && value != "" {
			return value
		}
	}
	return ""
}

func formatEffectiveDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	if len(metadata) == 0 {
		return nil
	}
	copy := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		copy[key] = value
	}
	return copy
}