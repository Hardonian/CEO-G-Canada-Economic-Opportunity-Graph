// Package trust computes evidence quality independently from project quality.
package trust

import (
	"math"
	"sort"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const MethodologyVersion = "evidence-quality-v1.0"

type Report struct {
	ProjectID             string                    `json:"project_id"`
	MethodologyVersion    string                    `json:"methodology_version"`
	Status                domain.IntelligenceStatus `json:"status"`
	OverallScore          *float64                  `json:"overall_score,omitempty"`
	PrimarySourceCoverage float64                   `json:"primary_source_coverage"`
	FieldCoverage         float64                   `json:"field_coverage"`
	RelationshipEvidence  float64                   `json:"relationship_evidence_coverage"`
	EvidenceCount         int                       `json:"evidence_count"`
	ConflictingClaims     int                       `json:"conflicting_claims"`
	CriticalUnknowns      []string                  `json:"critical_unknowns"`
	Staleness             string                    `json:"staleness"`
	LatestEffectiveDate   *time.Time                `json:"latest_effective_date,omitempty"`
}

// Evaluate is deterministic for a supplied asOf timestamp. Evidence quality
// never feeds Buildability; consumers can compare the two explicitly.
func Evaluate(project *domain.Project, evidence []*domain.Evidence, relationships []*domain.Relationship, asOf time.Time) Report {
	report := Report{ProjectID: project.ID, MethodologyVersion: MethodologyVersion, Status: domain.StatusUnavailable, Staleness: "UNKNOWN"}
	if len(evidence) == 0 {
		return report
	}

	report.EvidenceCount = len(evidence)
	primary := 0
	var latest time.Time
	for _, item := range evidence {
		if item.SourceTier == domain.SourceTier1 {
			primary++
		}
		if item.Confidence == domain.ConfidenceConflicted {
			report.ConflictingClaims++
		}
		candidate := item.RetrievalTimestamp
		if item.EffectiveDate != nil {
			candidate = *item.EffectiveDate
		}
		if candidate.After(latest) {
			latest = candidate
		}
	}
	report.PrimarySourceCoverage = percent(primary, len(evidence))
	if !latest.IsZero() {
		latest = latest.UTC()
		report.LatestEffectiveDate = &latest
		age := asOf.Sub(latest)
		switch {
		case age < 0:
			report.Staleness = "FUTURE_DATED"
		case age <= 90*24*time.Hour:
			report.Staleness = "LOW"
		case age <= 365*24*time.Hour:
			report.Staleness = "MEDIUM"
		default:
			report.Staleness = "HIGH"
		}
	}

	knownFields := 0
	if project.CurrentStage != "" && project.CurrentStage != domain.StageUnknown {
		knownFields++
	} else {
		report.CriticalUnknowns = append(report.CriticalUnknowns, "current_stage")
	}
	if project.CapexCAD > 0 && isKnown(project.CapexStatus) {
		knownFields++
	} else {
		report.CriticalUnknowns = append(report.CriticalUnknowns, "capex")
	}
	if project.ProponentID != "" {
		knownFields++
	} else {
		report.CriticalUnknowns = append(report.CriticalUnknowns, "proponent")
	}
	if project.LocationName != "" {
		knownFields++
	} else {
		report.CriticalUnknowns = append(report.CriticalUnknowns, "location")
	}
	if project.Sector != "" {
		knownFields++
	} else {
		report.CriticalUnknowns = append(report.CriticalUnknowns, "sector")
	}
	if project.Summary != "" {
		knownFields++
	} else {
		report.CriticalUnknowns = append(report.CriticalUnknowns, "summary")
	}
	report.FieldCoverage = percent(knownFields, 6)
	sort.Strings(report.CriticalUnknowns)

	if len(relationships) == 0 {
		report.RelationshipEvidence = 0
	} else {
		withEvidence := 0
		for _, relationship := range relationships {
			if relationship.EvidenceID != "" {
				withEvidence++
			}
		}
		report.RelationshipEvidence = percent(withEvidence, len(relationships))
	}

	stalenessScore := 0.0
	switch report.Staleness {
	case "LOW":
		stalenessScore = 100
	case "MEDIUM":
		stalenessScore = 65
	case "HIGH":
		stalenessScore = 20
	}
	overall := round(report.PrimarySourceCoverage*0.35 + report.FieldCoverage*0.35 + report.RelationshipEvidence*0.15 + stalenessScore*0.15)
	report.OverallScore = &overall
	report.Status = domain.StatusHealthy
	if len(report.CriticalUnknowns) > 0 || report.ConflictingClaims > 0 {
		report.Status = domain.StatusPartial
	}
	if report.Staleness == "HIGH" || report.Staleness == "FUTURE_DATED" {
		report.Status = domain.StatusStale
	}
	return report
}

func isKnown(value domain.ConfidenceLevel) bool {
	return value == domain.ConfidenceVerified || value == domain.ConfidenceSupported || value == domain.ConfidenceReported
}

func percent(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return round(float64(numerator) / float64(denominator) * 100)
}

func round(value float64) float64 { return math.Round(value*10) / 10 }
