package scoring

import (
	"fmt"
	"math"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/google/uuid"
)

const (
	VersionBuildability   = "buildability-v1.0"
	VersionInvestability  = "investability-v1.0"
	VersionSupplierability = "supplierability-v1.0"
	VersionStrategicity   = "strategicity-v1.0"
)

// ProjectContext aggregates relevant facts for deterministic scoring.
type ProjectContext struct {
	Project        *domain.Project
	CapitalItems   []*domain.CapitalItem
	Events         []*domain.Event
	Relationships  []*domain.Relationship
	Procurements   []*domain.Procurement
	Opportunities  []*domain.Opportunity
}

// CalculateBuildability computes a deterministic 0-100 score on project execution likelihood.
func CalculateBuildability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)

	// 1. Stage Progress (15%)
	stageScore := 10.0
	switch p.CurrentStage {
	case domain.StageDiscovered, domain.StageAnnounced:
		stageScore = 20.0
	case domain.StageEarlyDevelopment, domain.StageFeasibility:
		stageScore = 35.0
	case domain.StageEnvironmentalReview, domain.StagePermitting:
		stageScore = 55.0
	case domain.StageFinancing, domain.StageProcurement:
		stageScore = 70.0
	case domain.StageFIDLikely:
		stageScore = 85.0
	case domain.StageFID, domain.StageConstruction:
		stageScore = 95.0
	case domain.StageCommissioning, domain.StageOperating:
		stageScore = 100.0
	case domain.StageDelayed, domain.StagePaused:
		stageScore = 25.0
	case domain.StageCancelled:
		stageScore = 0.0
	}
	factors["stage_progress"] = stageScore

	// 2. Financing Readiness (15%)
	var totalCommitted int64
	for _, c := range ctx.CapitalItems {
		if c.Status == domain.CapitalCommitted || c.Status == domain.CapitalClosed || c.Status == domain.CapitalDisbursed {
			totalCommitted += c.AmountCAD
		}
	}
	finScore := 20.0
	if p.CapexCAD > 0 {
		ratio := float64(totalCommitted) / float64(p.CapexCAD)
		finScore = math.Min(100.0, 20.0+(ratio*80.0))
	}
	factors["financing_readiness"] = round(finScore)

	// 3. Indigenous Agreements (15%)
	indigScore := 20.0
	for _, r := range ctx.Relationships {
		if r.RelationType == "indigenous_partner" {
			indigScore += 35.0
		}
	}
	for _, ev := range ctx.Events {
		if ev.EventType == "indigenous_agreement" || ev.EventType == "impact_benefit_agreement" {
			indigScore += 25.0
		}
	}
	factors["indigenous_agreements"] = math.Min(100.0, indigScore)

	// 4. Regulatory & Environmental Progress (15%)
	regScore := 30.0
	for _, ev := range ctx.Events {
		if ev.EventType == "regulatory_filing" || ev.EventType == "terms_of_reference" {
			regScore += 15.0
		}
		if ev.EventType == "environmental_approval" || ev.EventType == "ministerial_decision" {
			regScore += 40.0
		}
	}
	factors["regulatory_environmental"] = math.Min(100.0, regScore)

	// 5. Land & Site Control (10%)
	siteScore := 40.0
	if p.LocationName != "" && p.Latitude != 0 && p.Longitude != 0 {
		siteScore += 30.0
	}
	if p.CurrentStage == domain.StageConstruction || p.CurrentStage == domain.StageFID {
		siteScore = 100.0
	}
	factors["site_control"] = math.Min(100.0, siteScore)

	// 6. Commercial & Offtake Agreements (10%)
	offtakeScore := 20.0
	for _, r := range ctx.Relationships {
		if r.RelationType == "offtaker" {
			offtakeScore += 40.0
		}
	}
	for _, ev := range ctx.Events {
		if ev.EventType == "offtake_agreement" || ev.EventType == "power_purchase_agreement" {
			offtakeScore += 30.0
		}
	}
	factors["offtake_commercial"] = math.Min(100.0, offtakeScore)

	// 7. Energy & Infrastructure Readiness (10%)
	infraScore := 45.0
	if p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorCleanEnergy {
		infraScore += 25.0
	}
	factors["infra_readiness"] = math.Min(100.0, infraScore)

	// 8. Proponent Credibility (5%)
	propScore := 50.0
	if p.Proponent != nil && (p.Proponent.EntityType == "CrownCorp" || p.Proponent.EntityType == "Utility") {
		propScore = 95.0
	} else if p.Proponent != nil && p.Proponent.EntityType == "Corporation" {
		propScore = 75.0
	}
	factors["proponent_credibility"] = propScore

	// Weighted sum
	total := (factors["stage_progress"] * 0.15) +
		(factors["financing_readiness"] * 0.15) +
		(factors["indigenous_agreements"] * 0.15) +
		(factors["regulatory_environmental"] * 0.15) +
		(factors["site_control"] * 0.10) +
		(factors["offtake_commercial"] * 0.10) +
		(factors["infra_readiness"] * 0.10) +
		(factors["proponent_credibility"] * 0.10)

	total = clamp(total, 0.0, 100.0)

	explanation := fmt.Sprintf("Buildability calculated under %s: stage maturity (%.0f), financing (%.0f), Indigenous consensus (%.0f), and regulatory clearance (%.0f).",
		VersionBuildability, factors["stage_progress"], factors["financing_readiness"], factors["indigenous_agreements"], factors["regulatory_environmental"])

	return &domain.ProjectScore{
		ID:           uuid.New().String(),
		ProjectID:    p.ID,
		ScoreType:    "buildability",
		ScoreValue:   round(total),
		ScoreVersion: VersionBuildability,
		Factors:      factors,
		Explanation:  explanation,
		CalculatedAt: time.Now(),
	}
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
