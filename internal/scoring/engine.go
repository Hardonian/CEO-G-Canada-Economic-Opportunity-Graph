package scoring

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
	"github.com/google/uuid"
)

const (
	VersionBuildability    = "buildability-v2.0"
	VersionInvestability   = "investability-v1.0"
	VersionSupplierability = "supplierability-v1.0"
	VersionStrategicity    = "strategicity-v1.0"
)

// ProjectContext aggregates relevant facts for deterministic scoring.
type ProjectContext struct {
	Project       *domain.Project
	CapitalItems  []*domain.CapitalItem
	Events        []*domain.Event
	Relationships []*domain.Relationship
	Procurements  []*domain.Procurement
	Opportunities []*domain.Opportunity
}

// CalculateBuildability computes a deterministic 0-100 score on project execution likelihood.
func CalculateBuildability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)
	weights := map[string]float64{
		"stage_progress": 0.15, "financing_readiness": 0.15,
		"indigenous_agreements": 0.15, "regulatory_environmental": 0.15,
		"site_control": 0.10, "offtake_commercial": 0.10,
		"infrastructure_readiness": 0.10, "execution_evidence": 0.10,
	}

	if p.CurrentStage != "" && p.CurrentStage != domain.StageUnknown {
		factors["stage_progress"] = stageProgressScore(p.CurrentStage)
	}

	if p.CapexCAD > 0 && isKnown(p.CapexStatus) && len(ctx.CapitalItems) > 0 {
		var committed int64
		for _, item := range ctx.CapitalItems {
			if item.AmountType == "exact" && (item.Status == domain.CapitalCommitted || item.Status == domain.CapitalClosed || item.Status == domain.CapitalDisbursed) {
				committed += item.AmountCAD
			}
		}
		factors["financing_readiness"] = round(math.Min(100, float64(committed)/float64(p.CapexCAD)*100))
	}

	for _, relationship := range ctx.Relationships {
		switch relationship.RelationType {
		case "indigenous_partner", "participates_in", "partners_with":
			factors["indigenous_agreements"] = 100
		case "offtakes_from", "offtaker":
			factors["offtake_commercial"] = 100
		case "depends_on", "enables":
			factors["infrastructure_readiness"] = 70
		}
	}

	for _, event := range ctx.Events {
		switch event.EventType {
		case "indigenous_agreement", "impact_benefit_agreement":
			factors["indigenous_agreements"] = 100
		case "offtake_agreement", "power_purchase_agreement":
			factors["offtake_commercial"] = 100
		case "site_control_secured", "land_secured":
			factors["site_control"] = 100
		case "regulatory.impact_assessment_filing", "regulatory_filing", "terms_of_reference":
			factors["regulatory_environmental"] = math.Max(factors["regulatory_environmental"], 50)
		case "regulatory.impact_assessment_decision_issued", "environmental_approval", "ministerial_decision":
			factors["regulatory_environmental"] = math.Max(factors["regulatory_environmental"], 90)
		case "regulatory.hold_point_removed":
			factors["regulatory_environmental"] = math.Max(factors["regulatory_environmental"], 95)
		case "project.commercial_operations_started", "construction_started":
			factors["execution_evidence"] = 100
		}
		if event.NewStage != nil && (*event.NewStage == domain.StageConstruction || *event.NewStage == domain.StageOperating) {
			factors["execution_evidence"] = 100
		}
	}

	var weighted, coverage float64
	var unknown []string
	for factor, weight := range weights {
		value, ok := factors[factor]
		if !ok {
			unknown = append(unknown, factor)
			continue
		}
		weighted += value * weight
		coverage += weight
	}
	sort.Strings(unknown)
	total := 0.0
	if coverage > 0 {
		total = clamp(weighted/coverage, 0, 100)
	}

	confidence := domain.ConfidenceUnknown
	if coverage >= 0.75 {
		confidence = domain.ConfidenceSupported
	} else if coverage >= 0.5 {
		confidence = domain.ConfidenceReported
	}
	inputHash := buildabilityInputHash(ctx)
	calculatedAt := latestInputTime(ctx)
	explanation := fmt.Sprintf("%s uses only evidenced factors; %.0f%% of factor weight is currently covered and %d factors remain unknown.", VersionBuildability, coverage*100, len(unknown))

	return &domain.ProjectScore{
		ID:             identity.StableID("score", VersionBuildability, p.ID+":"+inputHash),
		ProjectID:      p.ID,
		ScoreType:      "buildability",
		ScoreValue:     round(total),
		ScoreVersion:   VersionBuildability,
		Factors:        factors,
		UnknownFactors: unknown,
		Coverage:       round(coverage * 100),
		Confidence:     confidence,
		InputHash:      inputHash,
		Explanation:    explanation,
		CalculatedAt:   calculatedAt,
	}
}

func stageProgressScore(stage domain.LifecycleStage) float64 {
	switch stage {
	case domain.StageDiscovered, domain.StageAnnounced:
		return 20
	case domain.StageReferred, domain.StageEarlyDevelopment, domain.StageFeasibility:
		return 35
	case domain.StageEnvironmentalReview, domain.StagePermitting:
		return 55
	case domain.StageFinancing, domain.StageProcurement:
		return 70
	case domain.StageFIDLikely:
		return 85
	case domain.StageFID, domain.StageConstruction:
		return 95
	case domain.StageCommissioning, domain.StageOperating:
		return 100
	case domain.StageDelayed, domain.StagePaused:
		return 25
	case domain.StageCancelled:
		return 0
	default:
		return 0
	}
}

func isKnown(status domain.ConfidenceLevel) bool {
	return status == domain.ConfidenceVerified || status == domain.ConfidenceSupported || status == domain.ConfidenceReported
}

func buildabilityInputHash(ctx *ProjectContext) string {
	type input struct {
		Project       *domain.Project
		Events        []string
		Capital       []string
		Relationships []string
		Procurements  []string
	}
	value := input{Project: ctx.Project}
	for _, event := range ctx.Events {
		value.Events = append(value.Events, event.ID)
	}
	for _, item := range ctx.CapitalItems {
		value.Capital = append(value.Capital, item.ID)
	}
	for _, relationship := range ctx.Relationships {
		value.Relationships = append(value.Relationships, relationship.ID)
	}
	for _, procurement := range ctx.Procurements {
		value.Procurements = append(value.Procurements, procurement.ID)
	}
	sort.Strings(value.Events)
	sort.Strings(value.Capital)
	sort.Strings(value.Relationships)
	sort.Strings(value.Procurements)
	data, _ := json.Marshal(value)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func latestInputTime(ctx *ProjectContext) time.Time {
	latest := ctx.Project.UpdatedAt
	for _, event := range ctx.Events {
		if event.EventDate.After(latest) {
			latest = event.EventDate
		}
	}
	for _, item := range ctx.CapitalItems {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	for _, relationship := range ctx.Relationships {
		if relationship.CreatedAt.After(latest) {
			latest = relationship.CreatedAt
		}
	}
	for _, procurement := range ctx.Procurements {
		if procurement.CreatedAt.After(latest) {
			latest = procurement.CreatedAt
		}
	}
	if latest.IsZero() {
		return time.Unix(0, 0).UTC()
	}
	return latest.UTC()
}

// CalculateInvestability computes capital opportunity attractiveness (0-100).
func CalculateInvestability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)

	// 1. Capital Scale & Gap (25%)
	gapScore := 50.0
	if p.CapexCAD > 1_000_000_000 {
		gapScore = 90.0
	} else if p.CapexCAD > 250_000_000 {
		gapScore = 75.0
	} else if p.CapexCAD > 50_000_000 {
		gapScore = 60.0
	}
	factors["capital_scale"] = gapScore

	// 2. Financing Momentum (25%)
	momentumScore := 30.0
	for _, c := range ctx.CapitalItems {
		if c.Category == domain.CapitalCategoryCIB || c.Category == domain.CapitalCategoryCGF || c.Category == domain.CapitalCategoryEquity {
			momentumScore += 20.0
		}
	}
	factors["financing_momentum"] = math.Min(100.0, momentumScore)

	// 3. Strategic Importance (25%)
	stratScore := 50.0
	if p.Sector == domain.SectorCriticalMinerals || p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorAICompute {
		stratScore = 90.0
	} else if p.Sector == domain.SectorCleanEnergy || p.Sector == domain.SectorDefenceArctic {
		stratScore = 80.0
	}
	factors["strategic_importance"] = stratScore

	// 4. Execution Feasibility (25%)
	execScore := 50.0
	if p.CurrentStage == domain.StageFIDLikely || p.CurrentStage == domain.StageFID || p.CurrentStage == domain.StageConstruction {
		execScore = 85.0
	} else if p.CurrentStage == domain.StageEnvironmentalReview || p.CurrentStage == domain.StagePermitting {
		execScore = 65.0
	}
	factors["execution_feasibility"] = execScore

	total := (factors["capital_scale"] * 0.25) +
		(factors["financing_momentum"] * 0.25) +
		(factors["strategic_importance"] * 0.25) +
		(factors["execution_feasibility"] * 0.25)

	total = clamp(total, 0.0, 100.0)

	explanation := fmt.Sprintf("Investability evaluated under %s: strategic sector demand (%.0f), capital scale (%.0f), and financing velocity (%.0f).",
		VersionInvestability, factors["strategic_importance"], factors["capital_scale"], factors["financing_momentum"])

	return &domain.ProjectScore{
		ID:           uuid.New().String(),
		ProjectID:    p.ID,
		ScoreType:    "investability",
		ScoreValue:   round(total),
		ScoreVersion: VersionInvestability,
		Factors:      factors,
		Explanation:  explanation,
		CalculatedAt: time.Now(),
	}
}

// CalculateSupplierability computes contracting and supply-chain opportunity density (0-100).
func CalculateSupplierability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)

	// 1. Procurement Proximity (30%)
	procScore := 20.0
	switch p.CurrentStage {
	case domain.StageProcurement:
		procScore = 95.0
	case domain.StageFIDLikely, domain.StageFID:
		procScore = 85.0
	case domain.StagePermitting:
		procScore = 65.0
	case domain.StageConstruction:
		procScore = 75.0
	default:
		procScore = 30.0
	}
	factors["procurement_proximity"] = procScore

	// 2. Active Procurements Volume (25%)
	activeCount := len(ctx.Procurements)
	activeScore := math.Min(100.0, 20.0+float64(activeCount)*25.0)
	factors["tender_volume"] = activeScore

	// 3. Technical Complexity & Engineering Intensity (25%)
	techScore := 50.0
	if p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorAICompute {
		techScore = 95.0
	} else if p.Sector == domain.SectorCriticalMinerals || p.Sector == domain.SectorTransportation {
		techScore = 80.0
	}
	factors["technical_intensity"] = techScore

	// 4. Downstream Opportunities (20%)
	oppScore := math.Min(100.0, 20.0+float64(len(ctx.Opportunities))*15.0)
	factors["downstream_opportunities"] = oppScore

	total := (factors["procurement_proximity"] * 0.30) +
		(factors["tender_volume"] * 0.25) +
		(factors["technical_intensity"] * 0.25) +
		(factors["downstream_opportunities"] * 0.20)

	total = clamp(total, 0.0, 100.0)

	explanation := fmt.Sprintf("Supplierability determined under %s: procurement lifecycle proximity (%.0f), tender pipeline (%.0f), and technical complexity (%.0f).",
		VersionSupplierability, factors["procurement_proximity"], factors["tender_volume"], factors["technical_intensity"])

	return &domain.ProjectScore{
		ID:           uuid.New().String(),
		ProjectID:    p.ID,
		ScoreType:    "supplierability",
		ScoreValue:   round(total),
		ScoreVersion: VersionSupplierability,
		Factors:      factors,
		Explanation:  explanation,
		CalculatedAt: time.Now(),
	}
}

// CalculateStrategicity computes national economic and sovereignty significance (0-100).
func CalculateStrategicity(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)

	// Critical minerals (25%)
	critScore := 20.0
	if p.Sector == domain.SectorCriticalMinerals {
		critScore = 95.0
	}
	factors["critical_minerals"] = critScore

	// Energy Security & Clean Transition (25%)
	energyScore := 20.0
	if p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorCleanEnergy {
		energyScore = 90.0
	}
	factors["energy_security"] = energyScore

	// Sovereign Infrastructure & Defence / Arctic (25%)
	sovScore := 20.0
	if p.Sector == domain.SectorDefenceArctic || p.Province == "YT" || p.Province == "NT" || p.Province == "NU" {
		sovScore = 95.0
	} else if p.Sector == domain.SectorTransportation {
		sovScore = 80.0
	}
	factors["sovereign_infrastructure"] = sovScore

	// AI Compute & Digital Sovereignty (25%)
	aiScore := 20.0
	if p.Sector == domain.SectorAICompute {
		aiScore = 95.0
	}
	factors["ai_sovereignty"] = aiScore

	total := (factors["critical_minerals"] * 0.25) +
		(factors["energy_security"] * 0.25) +
		(factors["sovereign_infrastructure"] * 0.25) +
		(factors["ai_sovereignty"] * 0.25)

	total = clamp(total, 0.0, 100.0)

	explanation := fmt.Sprintf("Strategicity calculated under %s based on federal industrial mandates, critical mineral priorities, and sovereign infrastructure resilience.",
		VersionStrategicity)

	return &domain.ProjectScore{
		ID:           uuid.New().String(),
		ProjectID:    p.ID,
		ScoreType:    "strategicity",
		ScoreValue:   round(total),
		ScoreVersion: VersionStrategicity,
		Factors:      factors,
		Explanation:  explanation,
		CalculatedAt: time.Now(),
	}
}

func round(val float64) float64 {
	return math.Round(val*10) / 10
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
