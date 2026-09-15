// Package velocity provides deterministic, time-decayed metrics that measure
// the rate of progress across capital deployment, physical execution, and
// opportunity emergence. Each velocity score is decomposed into factors and
// evidence-linked, following the same pattern as the scoring engine.
package velocity

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const MethodologyVersion = "velocity-v1.0"

// VelocityType identifies which velocity dimension is being measured.
type VelocityType string

const (
	VelocityCapital   VelocityType = "CAPITAL"
	VelocityExecution VelocityType = "EXECUTION"
)

// VelocityScore is a deterministic, decomposed velocity metric.
type VelocityScore struct {
	ProjectID          string              `json:"project_id"`
	Type               VelocityType        `json:"type"`
	Score              float64             `json:"score"`   // 0-100
	Trend              string              `json:"trend"`   // ACCELERATING, STEADY, DECELERATING, STALLED
	MethodologyVersion string              `json:"methodology_version"`
	Factors            map[string]float64  `json:"factors"`
	FactorEvidence     map[string][]string `json:"factor_evidence"`
	Explanation        []string            `json:"explanation"`
	InputHash          string              `json:"input_hash"`
	CalculatedAt       time.Time           `json:"calculated_at"`
}

// CapitalVelocityInputs contains the data needed for capital velocity.
type CapitalVelocityInputs struct {
	Project      *domain.Project
	CapitalItems []*domain.CapitalItem
	Events       []*domain.Event
	Milestones   []*domain.Milestone
	CapitalNeeds []*domain.CapitalNeed
}

// CalculateCapitalVelocity measures the rate of financing progress:
// new commitments, partner announcements, financing milestones, and
// capital gap closure — all time-decayed relative to asOf.
func CalculateCapitalVelocity(inputs CapitalVelocityInputs, asOf time.Time) *VelocityScore {
	if inputs.Project == nil {
		return nil
	}

	factors := map[string]float64{}
	evidence := map[string][]string{}
	explanation := []string{}

	// Factor 1: Recent capital commitments (last 12 months, time-decayed).
	recentCommitments, commitEvidence := recentCapitalFactor(inputs.CapitalItems, asOf, 365)
	factors["recent_commitments"] = recentCommitments
	evidence["recent_commitments"] = commitEvidence
	if recentCommitments > 0 {
		explanation = append(explanation, "recent_commitments: active capital deployment detected")
	}

	// Factor 2: Financing milestone progress.
	financingProgress, finEvidence := financingMilestoneFactor(inputs.Milestones)
	factors["financing_milestones"] = financingProgress
	evidence["financing_milestones"] = finEvidence

	// Factor 3: Capital gap closure rate.
	gapClosure := capitalGapClosureFactor(inputs.CapitalItems, inputs.CapitalNeeds, inputs.Project)
	factors["gap_closure"] = gapClosure

	// Factor 4: Event momentum (financing-related events in last 6 months).
	eventMomentum, eventEvidence := eventMomentumFactor(inputs.Events, asOf, 180,
		[]string{"financing_announced", "contract_awarded", "indigenous_agreement"})
	factors["event_momentum"] = eventMomentum
	evidence["event_momentum"] = eventEvidence

	// Weighted composite.
	score := clamp(
		recentCommitments*0.35+
			financingProgress*0.25+
			gapClosure*0.25+
			eventMomentum*0.15,
		0, 100)

	trend := classifyTrend(score, eventMomentum)
	sort.Strings(explanation)

	return &VelocityScore{
		ProjectID:          inputs.Project.ID,
		Type:               VelocityCapital,
		Score:              round(score),
		Trend:              trend,
		MethodologyVersion: MethodologyVersion,
		Factors:            factors,
		FactorEvidence:     evidence,
		Explanation:        explanation,
		InputHash:          hashInputs(inputs.Project.ID, factors),
		CalculatedAt:       asOf,
	}
}

// ExecutionVelocityInputs contains the data needed for execution velocity.
type ExecutionVelocityInputs struct {
	Project    *domain.Project
	Milestones []*domain.Milestone
	Events     []*domain.Event
}

// CalculateExecutionVelocity measures the rate of physical project execution:
// engineering progress, permits, procurement awards, construction milestones.
func CalculateExecutionVelocity(inputs ExecutionVelocityInputs, asOf time.Time) *VelocityScore {
	if inputs.Project == nil {
		return nil
	}

	factors := map[string]float64{}
	evidence := map[string][]string{}
	explanation := []string{}

	// Factor 1: Milestone completion rate.
	completionRate, msEvidence := milestoneCompletionFactor(inputs.Milestones)
	factors["milestone_completion"] = completionRate
	evidence["milestone_completion"] = msEvidence

	// Factor 2: Engineering/construction milestone progress.
	execProgress, execEvidence := executionMilestoneFactor(inputs.Milestones)
	factors["execution_progress"] = execProgress
	evidence["execution_progress"] = execEvidence

	// Factor 3: Recent execution events (regulatory, construction, contract).
	eventMomentum, eventEvidence := eventMomentumFactor(inputs.Events, asOf, 180,
		[]string{"stage_change", "regulatory_filing", "contract_awarded"})
	factors["event_momentum"] = eventMomentum
	evidence["event_momentum"] = eventEvidence

	// Factor 4: Stage advancement rate.
	stageAdvancement := stageAdvancementFactor(inputs.Project.CurrentStage)
	factors["stage_advancement"] = stageAdvancement

	score := clamp(
		completionRate*0.30+
			execProgress*0.30+
			eventMomentum*0.20+
			stageAdvancement*0.20,
		0, 100)

	trend := classifyTrend(score, eventMomentum)
	sort.Strings(explanation)

	return &VelocityScore{
		ProjectID:          inputs.Project.ID,
		Type:               VelocityExecution,
		Score:              round(score),
		Trend:              trend,
		MethodologyVersion: MethodologyVersion,
		Factors:            factors,
		FactorEvidence:     evidence,
		Explanation:        explanation,
		InputHash:          hashInputs(inputs.Project.ID, factors),
		CalculatedAt:       asOf,
	}
}

// --- Internal factor calculations ---

func recentCapitalFactor(items []*domain.CapitalItem, asOf time.Time, windowDays int) (float64, []string) {
	if len(items) == 0 {
		return 0, nil
	}
	var totalDecayed float64
	var totalCapex int64
	evidence := []string{}
	cutoff := asOf.AddDate(0, 0, -windowDays)
	for _, item := range items {
		if item.CreatedAt.Before(cutoff) {
			continue
		}
		if item.Status != domain.CapitalCommitted && item.Status != domain.CapitalSigned &&
			item.Status != domain.CapitalDisbursed && item.Status != domain.CapitalClosed {
			continue
		}
		age := asOf.Sub(item.CreatedAt).Hours() / 24
		decay := math.Exp(-age / float64(windowDays) * 2)
		totalDecayed += float64(item.AmountCAD) * decay
		totalCapex += item.AmountCAD
		if item.EvidenceID != "" {
			evidence = append(evidence, item.EvidenceID)
		}
	}
	if totalCapex == 0 {
		return 0, evidence
	}
	// Normalize: $1B+ committed recently → 100.
	score := math.Min(totalDecayed/1e9*100, 100)
	return score, evidence
}

func financingMilestoneFactor(milestones []*domain.Milestone) (float64, []string) {
	financingTypes := map[domain.MilestoneType]float64{
		domain.MilestoneFinancingCommitted: 40,
		domain.MilestoneFinancialClose:     60,
		domain.MilestoneFID:               80,
		domain.MilestoneOfftakeSigned:      30,
	}
	var score float64
	evidence := []string{}
	for _, m := range milestones {
		weight, ok := financingTypes[m.Type]
		if !ok {
			continue
		}
		switch m.Status {
		case domain.MilestoneComplete:
			score += weight
		case domain.MilestoneInProgress:
			score += weight * 0.5
		case domain.MilestonePlanned:
			score += weight * 0.1
		}
		evidence = append(evidence, m.EvidenceIDs...)
	}
	return clamp(score, 0, 100), evidence
}

func capitalGapClosureFactor(items []*domain.CapitalItem, needs []*domain.CapitalNeed, project *domain.Project) float64 {
	if project.CapexCAD == 0 {
		return 0
	}
	var committed int64
	for _, item := range items {
		if item.Status == domain.CapitalCommitted || item.Status == domain.CapitalSigned ||
			item.Status == domain.CapitalDisbursed || item.Status == domain.CapitalClosed {
			committed += item.AmountCAD
		}
	}
	ratio := float64(committed) / float64(project.CapexCAD)
	return clamp(ratio*100, 0, 100)
}

func eventMomentumFactor(events []*domain.Event, asOf time.Time, windowDays int, types []string) (float64, []string) {
	typeSet := map[string]bool{}
	for _, t := range types {
		typeSet[t] = true
	}
	var decayedCount float64
	evidence := []string{}
	cutoff := asOf.AddDate(0, 0, -windowDays)
	for _, ev := range events {
		if ev.EventDate.Before(cutoff) || !typeSet[ev.EventType] {
			continue
		}
		age := asOf.Sub(ev.EventDate).Hours() / 24
		decay := math.Exp(-age / float64(windowDays) * 2)
		decayedCount += decay
		if ev.EvidenceID != "" {
			evidence = append(evidence, ev.EvidenceID)
		}
	}
	// 5+ weighted events → 100.
	return clamp(decayedCount/5*100, 0, 100), evidence
}

func milestoneCompletionFactor(milestones []*domain.Milestone) (float64, []string) {
	if len(milestones) == 0 {
		return 0, nil
	}
	completed := 0
	evidence := []string{}
	for _, m := range milestones {
		if m.Status == domain.MilestoneComplete {
			completed++
			evidence = append(evidence, m.EvidenceIDs...)
		}
	}
	return clamp(float64(completed)/float64(len(milestones))*100, 0, 100), evidence
}

func executionMilestoneFactor(milestones []*domain.Milestone) (float64, []string) {
	executionTypes := map[domain.MilestoneType]float64{
		domain.MilestoneEnvironmentalApproval: 15,
		domain.MilestonePermitReceived:        15,
		domain.MilestoneGridConnection:        10,
		domain.MilestoneEPCAwarded:            20,
		domain.MilestoneConstructionStarted:   25,
		domain.MilestoneCommercialOperation:   15,
	}
	var score float64
	evidence := []string{}
	for _, m := range milestones {
		weight, ok := executionTypes[m.Type]
		if !ok {
			continue
		}
		if m.Status == domain.MilestoneComplete {
			score += weight
			evidence = append(evidence, m.EvidenceIDs...)
		} else if m.Status == domain.MilestoneInProgress {
			score += weight * 0.4
		}
	}
	return clamp(score, 0, 100), evidence
}

func stageAdvancementFactor(stage domain.LifecycleStage) float64 {
	stageScore := map[domain.LifecycleStage]float64{
		domain.StageConcept:            5,
		domain.StagePreDevelopment:     10,
		domain.StageAnnounced:          15,
		domain.StageFeasibility:        25,
		domain.StagePreFEED:            30,
		domain.StageFEED:               40,
		domain.StageDetailedEngineering: 50,
		domain.StagePermitting:         55,
		domain.StageFinancing:          60,
		domain.StageFIDLikely:          70,
		domain.StageFID:                75,
		domain.StageConstructionReady:  80,
		domain.StageConstruction:       85,
		domain.StageCommissioning:      90,
		domain.StageOperating:          100,
	}
	if s, ok := stageScore[stage]; ok {
		return s
	}
	return 0
}

func classifyTrend(score, momentum float64) string {
	switch {
	case momentum >= 60 && score >= 40:
		return "ACCELERATING"
	case momentum >= 20:
		return "STEADY"
	case score > 0:
		return "DECELERATING"
	default:
		return "STALLED"
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round(v float64) float64 {
	return math.Round(v*100) / 100
}

func hashInputs(projectID string, factors map[string]float64) string {
	b, _ := json.Marshal(map[string]interface{}{
		"project_id": projectID,
		"factors":    factors,
		"version":    MethodologyVersion,
	})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:16])
}
