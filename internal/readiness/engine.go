package readiness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const MethodologyVersion = "investment-readiness-v1.0"

type Inputs struct {
	Project      *domain.Project
	Milestones   []*domain.Milestone
	CapitalNeeds []*domain.CapitalNeed
	CapitalItems []*domain.CapitalItem
}

type factor struct {
	value    float64
	known    bool
	evidence []string
	reason   string
}

// Calculate produces a deterministic, decomposed readiness assessment. The
// timestamp is supplied by the caller and excluded from the input hash.
func Calculate(inputs Inputs, calculatedAt time.Time) *domain.ReadinessAssessment {
	projectID := ""
	if inputs.Project != nil {
		projectID = inputs.Project.ID
	}
	factors := map[string]factor{
		"engineering":    engineering(inputs),
		"regulatory":     milestoneFactor(inputs.Milestones, []domain.MilestoneType{domain.MilestoneEnvironmentalApproval, domain.MilestonePermitReceived}),
		"financing":      financing(inputs),
		"commercial":     commercial(inputs),
		"site":           milestoneFactor(inputs.Milestones, []domain.MilestoneType{domain.MilestoneLandSecured, domain.MilestoneSiteSelected}),
		"infrastructure": milestoneFactor(inputs.Milestones, []domain.MilestoneType{domain.MilestoneGridConnection}),
		"offtake":        offtake(inputs),
		"execution":      milestoneFactor(inputs.Milestones, []domain.MilestoneType{domain.MilestoneEPCAwarded, domain.MilestoneConstructionStarted, domain.MilestoneCommercialOperation}),
	}

	values := make(map[string]float64, len(factors))
	evidence := make(map[string][]string, len(factors))
	unknown := make([]string, 0)
	explanation := make([]string, 0, len(factors))
	known := 0
	for name, value := range factors {
		values[name] = round(value.value)
		evidence[name] = sortedUnique(value.evidence)
		if value.known {
			known++
		} else {
			unknown = append(unknown, name)
		}
		if value.reason != "" {
			explanation = append(explanation, name+": "+value.reason)
		}
	}
	sort.Strings(unknown)
	sort.Strings(explanation)

	vector := domain.ReadinessVector{
		Engineering: values["engineering"], Regulatory: values["regulatory"],
		Financing: values["financing"], Commercial: values["commercial"],
		Site: values["site"], Infrastructure: values["infrastructure"],
		Offtake: values["offtake"], Execution: values["execution"],
	}
	overall := round((vector.Engineering + vector.Regulatory + vector.Financing + vector.Commercial +
		vector.Site + vector.Infrastructure + vector.Offtake + vector.Execution) / 8)
	coverage := round(float64(known) / 8 * 100)
	inputHash := hashInputs(projectID, values, evidence, unknown)

	return &domain.ReadinessAssessment{
		ID: "readiness-" + projectID + "-" + inputHash[:12], ProjectID: projectID,
		MethodologyVersion: MethodologyVersion, Vector: vector, InvestmentReadiness: overall,
		Coverage: coverage, Factors: values, FactorEvidence: evidence, UnknownFactors: unknown,
		Explanation: explanation, InputHash: inputHash, CalculatedAt: calculatedAt.UTC(),
	}
}

func engineering(inputs Inputs) factor {
	best := stageScore(inputs.Project)
	evidence := []string{}
	known := inputs.Project != nil && inputs.Project.CurrentStage != domain.StageUnknown
	for _, milestone := range inputs.Milestones {
		if milestone.Type == domain.MilestoneFEEDComplete || milestone.Type == domain.MilestoneFeasibilityComplete {
			candidate := milestoneScore(milestone.Status)
			if milestone.Type == domain.MilestoneFeasibilityComplete && candidate > 75 {
				candidate = 75
			}
			if candidate > best {
				best = candidate
			}
			known = true
			evidence = append(evidence, milestone.EvidenceIDs...)
		}
	}
	return factor{value: best, known: known, evidence: evidence, reason: "normalized stage and sourced engineering milestones"}
}

func stageScore(project *domain.Project) float64 {
	if project == nil {
		return 0
	}
	switch project.CurrentStage {
	case domain.StageConcept, domain.StageDiscovered, domain.StageAnnounced:
		return 15
	case domain.StagePreDevelopment, domain.StageEarlyDevelopment:
		return 25
	case domain.StageFeasibility, domain.StagePreFEED:
		return 45
	case domain.StageFEED, domain.StageEnvironmentalReview, domain.StagePermitting, domain.StageFinancing:
		return 60
	case domain.StageDetailedEngineering, domain.StageFIDLikely, domain.StageProcurement:
		return 75
	case domain.StageFID, domain.StageConstructionReady:
		return 90
	case domain.StageConstruction, domain.StageCommissioning, domain.StageOperating, domain.StageExpansion:
		return 100
	default:
		return 0
	}
}

func milestoneFactor(milestones []*domain.Milestone, types []domain.MilestoneType) factor {
	allowed := make(map[domain.MilestoneType]struct{}, len(types))
	for _, kind := range types {
		allowed[kind] = struct{}{}
	}
	best := 0.0
	known := false
	evidence := []string{}
	for _, milestone := range milestones {
		if milestone == nil {
			continue
		}
		if _, ok := allowed[milestone.Type]; !ok {
			continue
		}
		known = true
		if score := milestoneScore(milestone.Status); score > best {
			best = score
		}
		evidence = append(evidence, milestone.EvidenceIDs...)
	}
	return factor{value: best, known: known, evidence: evidence, reason: "sourced milestone status; unknown is not treated as complete"}
}

func milestoneScore(status domain.MilestoneStatus) float64 {
	switch status {
	case domain.MilestoneComplete:
		return 100
	case domain.MilestoneInProgress:
		return 60
	case domain.MilestonePlanned:
		return 30
	case domain.MilestoneDelayed:
		return 20
	case domain.MilestoneCancelled:
		return 0
	default:
		return 0
	}
}

func financing(inputs Inputs) factor {
	if len(inputs.CapitalNeeds) == 0 && len(inputs.CapitalItems) == 0 {
		return factor{reason: "no public financing need or commitment evidence"}
	}
	score := 20.0
	evidence := []string{}
	for _, need := range inputs.CapitalNeeds {
		if need == nil {
			continue
		}
		evidence = append(evidence, need.EvidenceIDs...)
		if len(need.Types) > 0 {
			score = math.Max(score, 40)
		}
		if need.Amount != nil && need.Amount.AmountType != domain.AmountNotAvailable {
			score = math.Max(score, 55)
		}
	}
	for _, item := range inputs.CapitalItems {
		if item == nil {
			continue
		}
		evidence = append(evidence, item.EvidenceID)
		switch item.Status {
		case domain.CapitalCommitted, domain.CapitalSigned:
			score = math.Max(score, 75)
		case domain.CapitalClosed, domain.CapitalDisbursed:
			score = math.Max(score, 100)
		}
	}
	return factor{value: score, known: true, evidence: evidence, reason: "financing specificity and observed commitment status"}
}

func commercial(inputs Inputs) factor {
	base := milestoneFactor(inputs.Milestones, []domain.MilestoneType{domain.MilestoneOfftakeSigned})
	for _, need := range inputs.CapitalNeeds {
		for _, kind := range need.Types {
			if kind == domain.NeedAnchorTenant || kind == domain.NeedCommercialPartnership || kind == domain.NeedOfftake {
				base.known = true
				base.value = math.Max(base.value, 35)
				base.evidence = append(base.evidence, need.EvidenceIDs...)
			}
		}
	}
	base.reason = "public commercial agreements and explicitly stated counterparty needs"
	return base
}

func offtake(inputs Inputs) factor {
	base := milestoneFactor(inputs.Milestones, []domain.MilestoneType{domain.MilestoneOfftakeSigned})
	for _, need := range inputs.CapitalNeeds {
		for _, kind := range need.Types {
			if kind == domain.NeedOfftake {
				base.known = true
				base.value = math.Max(base.value, 30)
				base.evidence = append(base.evidence, need.EvidenceIDs...)
			}
		}
	}
	base.reason = "signed offtake milestones outrank an observed open offtake need"
	return base
}

func hashInputs(projectID string, values map[string]float64, evidence map[string][]string, unknown []string) string {
	payload := struct {
		ProjectID string              `json:"project_id"`
		Version   string              `json:"version"`
		Values    map[string]float64  `json:"values"`
		Evidence  map[string][]string `json:"evidence"`
		Unknown   []string            `json:"unknown"`
	}{projectID, MethodologyVersion, values, evidence, unknown}
	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func sortedUnique(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			if _, ok := seen[value]; !ok {
				seen[value] = struct{}{}
				result = append(result, value)
			}
		}
	}
	sort.Strings(result)
	return result
}

func round(value float64) float64 { return math.Round(value*10) / 10 }

// CalculateCapitalGap refuses to blend currencies or count contingent and
// overlapping support. coverageComplete must be established by a source audit;
// otherwise the result is explicitly PARTIAL.
func CalculateCapitalGap(projectID string, requirement *domain.CapitalRequirement, items []*domain.CapitalItem, coverageComplete bool) domain.CapitalGap {
	result := domain.CapitalGap{ProjectID: projectID, State: domain.CapitalGapUnknown, Confidence: domain.ConfidenceUnknown}
	if requirement == nil || requirement.Amount.AmountType == domain.AmountNotAvailable || requirement.Amount.Currency == "" {
		result.Explanation = []string{"Total project requirement is not known."}
		return result
	}
	total, ok := singleAmount(requirement.Amount)
	if !ok {
		result.State = domain.CapitalGapPartial
		result.TotalRequirement = &requirement.Amount
		result.Explanation = []string{"A range or one-sided estimate is preserved and is not collapsed into a definitive gap."}
		return result
	}
	committed := int64(0)
	for _, item := range items {
		if item == nil || item.OriginalAmount == nil || item.OriginalAmount.Currency != requirement.Amount.Currency || item.StackTreatment != "ADDITIVE" {
			continue
		}
		switch item.Status {
		case domain.CapitalCommitted, domain.CapitalSigned, domain.CapitalClosed, domain.CapitalDisbursed:
		default:
			continue
		}
		amount, exact := singleAmount(*item.OriginalAmount)
		if exact {
			committed += amount
		}
	}
	if committed > total {
		result.State = domain.CapitalGapConflicted
		result.Confidence = domain.ConfidenceConflict
		result.TotalRequirement = &requirement.Amount
		result.Explanation = []string{"Compatible committed capital exceeds the stated requirement; sources require reconciliation."}
		return result
	}
	committedAmount := domain.MonetaryAmount{Amount: &committed, Currency: requirement.Amount.Currency, AmountType: domain.AmountExact}
	gap := total - committed
	gapAmount := domain.MonetaryAmount{Amount: &gap, Currency: requirement.Amount.Currency, AmountType: domain.AmountExact}
	result.TotalRequirement = &requirement.Amount
	result.CompatibleCommitted = &committedAmount
	result.PotentialGap = &gapAmount
	result.State = domain.CapitalGapPartial
	result.Confidence = domain.ConfidenceSupported
	result.Explanation = []string{"Only additive, same-currency, committed/signed/closed/disbursed amounts were counted."}
	if coverageComplete {
		result.State = domain.CapitalGapKnown
		result.Confidence = domain.ConfidenceVerified
	} else {
		result.Explanation = append(result.Explanation, "Capital-stack coverage is incomplete; the value is a potential documented gap.")
	}
	return result
}

func singleAmount(amount domain.MonetaryAmount) (int64, bool) {
	if amount.Amount != nil && (amount.AmountType == domain.AmountExact || amount.AmountType == domain.AmountApproximate) {
		return *amount.Amount, true
	}
	return 0, false
}

var _ = fmt.Sprintf
