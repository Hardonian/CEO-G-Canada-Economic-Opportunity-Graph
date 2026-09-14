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
)

const (
	VersionBuildability    = "buildability-v2.1"
	VersionInvestability   = "investability-v2.0"
	VersionSupplierability = "supplierability-v2.0"
	VersionStrategicity    = "strategicity-v2.0"
	VersionTradeResilience = "trade-resilience-v1.0"
)

// ProjectContext aggregates persisted, evidence-addressable scoring inputs.
type ProjectContext struct {
	Project       *domain.Project
	CapitalItems  []*domain.CapitalItem
	Events        []*domain.Event
	Relationships []*domain.Relationship
	Procurements  []*domain.Procurement
	Opportunities []*domain.Opportunity
	TradeMetrics  []*domain.TradeMetric
}

// CalculateBuildability computes project execution readiness from observed
// facts only. Missing factors are omitted and exposed through Coverage.
func CalculateBuildability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)
	factorEvidence := make(map[string][]string)
	weights := map[string]float64{
		"stage_progress": 0.15, "financing_readiness": 0.15,
		"indigenous_agreements": 0.15, "regulatory_environmental": 0.15,
		"site_control": 0.10, "offtake_commercial": 0.10,
		"infrastructure_readiness": 0.10, "execution_evidence": 0.10,
	}

	if p.CurrentStage != "" && p.CurrentStage != domain.StageUnknown {
		factors["stage_progress"] = stageProgressScore(p.CurrentStage)
		factorEvidence["stage_progress"] = append(factorEvidence["stage_progress"], p.EvidenceIDs...)
	}
	if p.CapexCAD > 0 && isKnown(p.CapexStatus) && len(ctx.CapitalItems) > 0 {
		var committed int64
		for _, item := range ctx.CapitalItems {
			if item.AmountType == "exact" && (item.Status == domain.CapitalCommitted || item.Status == domain.CapitalClosed || item.Status == domain.CapitalDisbursed) {
				committed += item.AmountCAD
				factorEvidence["financing_readiness"] = append(factorEvidence["financing_readiness"], item.EvidenceID)
			}
		}
		factors["financing_readiness"] = round(math.Min(100, float64(committed)/float64(p.CapexCAD)*100))
	}
	for _, relationship := range ctx.Relationships {
		switch relationship.RelationType {
		case "indigenous_partner", "participates_in", "partners_with":
			factors["indigenous_agreements"] = 100
			factorEvidence["indigenous_agreements"] = append(factorEvidence["indigenous_agreements"], relationship.EvidenceID)
		case "offtakes_from", "offtaker":
			factors["offtake_commercial"] = 100
			factorEvidence["offtake_commercial"] = append(factorEvidence["offtake_commercial"], relationship.EvidenceID)
		case "depends_on", "enables":
			factors["infrastructure_readiness"] = 70
			factorEvidence["infrastructure_readiness"] = append(factorEvidence["infrastructure_readiness"], relationship.EvidenceID)
		}
	}
	for _, event := range ctx.Events {
		switch event.EventType {
		case "indigenous_agreement", "impact_benefit_agreement":
			factors["indigenous_agreements"] = 100
			factorEvidence["indigenous_agreements"] = append(factorEvidence["indigenous_agreements"], event.EvidenceID)
		case "offtake_agreement", "power_purchase_agreement":
			factors["offtake_commercial"] = 100
			factorEvidence["offtake_commercial"] = append(factorEvidence["offtake_commercial"], event.EvidenceID)
		case "site_control_secured", "land_secured":
			factors["site_control"] = 100
			factorEvidence["site_control"] = append(factorEvidence["site_control"], event.EvidenceID)
		case "regulatory.impact_assessment_filing", "regulatory_filing", "terms_of_reference":
			factors["regulatory_environmental"] = math.Max(factors["regulatory_environmental"], 50)
			factorEvidence["regulatory_environmental"] = append(factorEvidence["regulatory_environmental"], event.EvidenceID)
		case "regulatory.impact_assessment_decision_issued", "environmental_approval", "ministerial_decision":
			factors["regulatory_environmental"] = math.Max(factors["regulatory_environmental"], 90)
			factorEvidence["regulatory_environmental"] = append(factorEvidence["regulatory_environmental"], event.EvidenceID)
		case "regulatory.hold_point_removed":
			factors["regulatory_environmental"] = math.Max(factors["regulatory_environmental"], 95)
			factorEvidence["regulatory_environmental"] = append(factorEvidence["regulatory_environmental"], event.EvidenceID)
		case "project.commercial_operations_started", "construction_started":
			factors["execution_evidence"] = 100
			factorEvidence["execution_evidence"] = append(factorEvidence["execution_evidence"], event.EvidenceID)
		}
		if event.NewStage != nil && (*event.NewStage == domain.StageConstruction || *event.NewStage == domain.StageOperating) {
			factors["execution_evidence"] = 100
			factorEvidence["execution_evidence"] = append(factorEvidence["execution_evidence"], event.EvidenceID)
		}
	}

	return finalize(ctx, "buildability", VersionBuildability, factors, weights, factorEvidence, false,
		"uses only evidenced execution factors; uncovered factor weight is reported rather than replaced with a default")
}

// CalculateInvestability computes capital opportunity attractiveness.
func CalculateInvestability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)
	evidence := make(map[string][]string)
	weights := map[string]float64{"capital_scale": .25, "financing_momentum": .25, "strategic_importance": .25, "execution_feasibility": .25}

	if p.CapexCAD > 0 && isKnown(p.CapexStatus) {
		score := 50.0
		if p.CapexCAD > 1_000_000_000 {
			score = 90
		} else if p.CapexCAD > 250_000_000 {
			score = 75
		} else if p.CapexCAD > 50_000_000 {
			score = 60
		}
		factors["capital_scale"] = score
		evidence["capital_scale"] = append(evidence["capital_scale"], p.EvidenceIDs...)
	}
	if len(ctx.CapitalItems) > 0 {
		score := 30.0
		for _, item := range ctx.CapitalItems {
			if item.Category == domain.CapitalCategoryCIB || item.Category == domain.CapitalCategoryCGF || item.Category == domain.CapitalCategoryEquity {
				score += 20
			}
			evidence["financing_momentum"] = append(evidence["financing_momentum"], item.EvidenceID)
		}
		factors["financing_momentum"] = math.Min(100, score)
	}
	if p.Sector != "" {
		score := 50.0
		if p.Sector == domain.SectorCriticalMinerals || p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorAICompute {
			score = 90
		} else if p.Sector == domain.SectorCleanEnergy || p.Sector == domain.SectorDefenceArctic {
			score = 80
		}
		factors["strategic_importance"] = score
		evidence["strategic_importance"] = append(evidence["strategic_importance"], p.EvidenceIDs...)
	}
	if p.CurrentStage != "" && p.CurrentStage != domain.StageUnknown {
		score := 50.0
		if p.CurrentStage == domain.StageFIDLikely || p.CurrentStage == domain.StageFID || p.CurrentStage == domain.StageConstruction {
			score = 85
		} else if p.CurrentStage == domain.StageEnvironmentalReview || p.CurrentStage == domain.StagePermitting {
			score = 65
		}
		factors["execution_feasibility"] = score
		evidence["execution_feasibility"] = append(evidence["execution_feasibility"], p.EvidenceIDs...)
	}
	return finalize(ctx, "investability", VersionInvestability, factors, weights, evidence, false,
		"combines disclosed capital scale, observed financing events, classified strategic demand, and evidenced lifecycle maturity")
}

// CalculateSupplierability computes contracting and supply-chain opportunity
// density. The official World Bank logistics observation contributes 15%.
func CalculateSupplierability(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := make(map[string]float64)
	evidence := make(map[string][]string)
	weights := map[string]float64{"procurement_proximity": .25, "tender_volume": .20, "technical_intensity": .20, "downstream_opportunities": .20, "trade_logistics": .15}

	if p.CurrentStage != "" && p.CurrentStage != domain.StageUnknown {
		score := 30.0
		switch p.CurrentStage {
		case domain.StageProcurement:
			score = 95
		case domain.StageFIDLikely, domain.StageFID:
			score = 85
		case domain.StagePermitting:
			score = 65
		case domain.StageConstruction:
			score = 75
		}
		factors["procurement_proximity"] = score
		evidence["procurement_proximity"] = append(evidence["procurement_proximity"], p.EvidenceIDs...)
	}
	if len(ctx.Procurements) > 0 {
		factors["tender_volume"] = math.Min(100, 20+float64(len(ctx.Procurements))*25)
		for _, item := range ctx.Procurements {
			evidence["tender_volume"] = append(evidence["tender_volume"], item.EvidenceID)
		}
	}
	if p.Sector != "" {
		score := 50.0
		if p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorAICompute {
			score = 95
		} else if p.Sector == domain.SectorCriticalMinerals || p.Sector == domain.SectorTransportation {
			score = 80
		}
		factors["technical_intensity"] = score
		evidence["technical_intensity"] = append(evidence["technical_intensity"], p.EvidenceIDs...)
	}
	if len(ctx.Opportunities) > 0 {
		factors["downstream_opportunities"] = math.Min(100, 20+float64(len(ctx.Opportunities))*15)
		evidence["downstream_opportunities"] = append(evidence["downstream_opportunities"], p.EvidenceIDs...)
	}
	if metric := latestMetric(ctx.TradeMetrics, "LP.LPI.OVRL.XQ"); metric != nil {
		factors["trade_logistics"] = scaledMetric(metric)
		evidence["trade_logistics"] = []string{metric.EvidenceID}
	}
	return finalize(ctx, "supplierability", VersionSupplierability, factors, weights, evidence, true,
		"combines project procurement signals with the official Canada logistics-performance context; missing tender or trade inputs remain explicit")
}

// CalculateStrategicity computes national economic and sovereignty significance.
func CalculateStrategicity(ctx *ProjectContext) *domain.ProjectScore {
	p := ctx.Project
	factors := map[string]float64{"critical_minerals": 20, "energy_security": 20, "sovereign_infrastructure": 20, "ai_sovereignty": 20}
	evidence := make(map[string][]string)
	for factor := range factors {
		evidence[factor] = append(evidence[factor], p.EvidenceIDs...)
	}
	if p.Sector == domain.SectorCriticalMinerals {
		factors["critical_minerals"] = 95
	}
	if p.Sector == domain.SectorNuclearEnergy || p.Sector == domain.SectorCleanEnergy {
		factors["energy_security"] = 90
	}
	if p.Sector == domain.SectorDefenceArctic || p.Province == "YT" || p.Province == "NT" || p.Province == "NU" {
		factors["sovereign_infrastructure"] = 95
	} else if p.Sector == domain.SectorTransportation {
		factors["sovereign_infrastructure"] = 80
	}
	if p.Sector == domain.SectorAICompute {
		factors["ai_sovereignty"] = 95
	}
	weights := map[string]float64{"critical_minerals": .25, "energy_security": .25, "sovereign_infrastructure": .25, "ai_sovereignty": .25}
	return finalize(ctx, "strategicity", VersionStrategicity, factors, weights, evidence, false,
		"classifies evidenced project sector and geography against published Canadian industrial-sovereignty dimensions")
}

// CalculateTradeResilience produces a transparent national-context score for
// every Canadian project. It does not claim project-specific import exposure.
func CalculateTradeResilience(ctx *ProjectContext) *domain.ProjectScore {
	factors := make(map[string]float64)
	evidence := make(map[string][]string)
	weights := map[string]float64{"logistics_performance": .45, "trade_participation": .20, "two_way_merchandise_balance": .20, "high_technology_export_intensity": .15}

	if metric := latestMetric(ctx.TradeMetrics, "LP.LPI.OVRL.XQ"); metric != nil {
		factors["logistics_performance"] = scaledMetric(metric)
		evidence["logistics_performance"] = []string{metric.EvidenceID}
	}
	if metric := latestMetric(ctx.TradeMetrics, "NE.TRD.GNFS.ZS"); metric != nil {
		factors["trade_participation"] = clamp(metric.Value, 0, 100)
		evidence["trade_participation"] = []string{metric.EvidenceID}
	}
	exports := latestMetric(ctx.TradeMetrics, "TX.VAL.MRCH.CD.WT")
	imports := latestMetric(ctx.TradeMetrics, "TM.VAL.MRCH.CD.WT")
	if exports != nil && imports != nil && exports.Value+imports.Value > 0 {
		factors["two_way_merchandise_balance"] = clamp(100*(1-math.Abs(exports.Value-imports.Value)/(exports.Value+imports.Value)), 0, 100)
		evidence["two_way_merchandise_balance"] = []string{exports.EvidenceID, imports.EvidenceID}
	}
	if metric := latestMetric(ctx.TradeMetrics, "TX.VAL.TECH.MF.ZS"); metric != nil {
		factors["high_technology_export_intensity"] = clamp(metric.Value, 0, 100)
		evidence["high_technology_export_intensity"] = []string{metric.EvidenceID}
	}
	return finalize(ctx, "trade_resilience", VersionTradeResilience, factors, weights, evidence, true,
		"is a Canada-level context index: 45% normalized LPI, 20% trade-to-GDP, 20% two-way merchandise-flow balance, and 15% high-technology export share")
}

func finalize(ctx *ProjectContext, scoreType, version string, factors, weights map[string]float64, factorEvidence map[string][]string, includeTrade bool, methodology string) *domain.ProjectScore {
	var weighted, coverage float64
	unknown := make([]string, 0)
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
	evidenceIDs := flattenEvidence(factorEvidence)
	inputHash := scoringInputHash(ctx.Project.ID, scoreType, version, factors, evidenceIDs)
	confidence := domain.ConfidenceUnknown
	if coverage >= .75 {
		confidence = domain.ConfidenceSupported
	} else if coverage >= .5 {
		confidence = domain.ConfidenceReported
	}
	unknownSummary := fmt.Sprintf("%d factors are unknown", len(unknown))
	if len(unknown) == 1 {
		unknownSummary = "1 factor is unknown"
	}
	return &domain.ProjectScore{
		ID: identity.StableID("score", version, ctx.Project.ID+":"+inputHash), ProjectID: ctx.Project.ID,
		ScoreType: scoreType, ScoreValue: round(total), ScoreVersion: version, Factors: factors,
		FactorEvidence: normalizeFactorEvidence(factorEvidence), EvidenceIDs: evidenceIDs,
		UnknownFactors: unknown, Coverage: round(coverage * 100), Confidence: confidence, InputHash: inputHash,
		Explanation:  fmt.Sprintf("%s %s; %.0f%% of configured factor weight is covered and %s.", version, methodology, coverage*100, unknownSummary),
		CalculatedAt: latestInputTime(ctx, includeTrade),
	}
}

func latestMetric(metrics []*domain.TradeMetric, code string) *domain.TradeMetric {
	var latest *domain.TradeMetric
	for _, metric := range metrics {
		if metric != nil && metric.MetricCode == code && (latest == nil || metric.ReferencePeriod > latest.ReferencePeriod) {
			latest = metric
		}
	}
	return latest
}

func scaledMetric(metric *domain.TradeMetric) float64 {
	if metric.ScaleMin != nil && metric.ScaleMax != nil && *metric.ScaleMax > *metric.ScaleMin {
		return clamp((metric.Value-*metric.ScaleMin)/(*metric.ScaleMax-*metric.ScaleMin)*100, 0, 100)
	}
	return clamp(metric.Value, 0, 100)
}

func flattenEvidence(factors map[string][]string) []string {
	var ids []string
	for _, values := range factors {
		ids = append(ids, values...)
	}
	return sortedUnique(ids)
}

func normalizeFactorEvidence(value map[string][]string) map[string][]string {
	result := make(map[string][]string, len(value))
	for factor, ids := range value {
		if normalized := sortedUnique(ids); len(normalized) > 0 {
			result[factor] = normalized
		}
	}
	return result
}

func sortedUnique(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func scoringInputHash(projectID, scoreType, version string, factors map[string]float64, evidenceIDs []string) string {
	payload := struct {
		ProjectID   string             `json:"project_id"`
		ScoreType   string             `json:"score_type"`
		Version     string             `json:"version"`
		Factors     map[string]float64 `json:"factors"`
		EvidenceIDs []string           `json:"evidence_ids"`
	}{projectID, scoreType, version, factors, evidenceIDs}
	data, _ := json.Marshal(payload)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func latestInputTime(ctx *ProjectContext, includeTrade bool) time.Time {
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
	for _, item := range ctx.Relationships {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	for _, item := range ctx.Procurements {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	for _, item := range ctx.Opportunities {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	if includeTrade {
		for _, item := range ctx.TradeMetrics {
			if item.ObservedAt.After(latest) {
				latest = item.ObservedAt
			}
		}
	}
	if latest.IsZero() {
		return time.Unix(0, 0).UTC()
	}
	return latest.UTC()
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

func round(value float64) float64 { return math.Round(value*10) / 10 }

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
