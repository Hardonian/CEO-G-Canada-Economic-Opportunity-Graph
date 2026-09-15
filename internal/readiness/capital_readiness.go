package readiness

import (
	"sort"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const CapitalReadinessVersion = "capital-readiness-v1.0"

// CapitalReadinessInputs provides the project-local context for capital
// readiness assessment.
type CapitalReadinessInputs struct {
	Project      *domain.Project
	Milestones   []*domain.Milestone
	CapitalNeeds []*domain.CapitalNeed
	CapitalItems []*domain.CapitalItem
	FID          *domain.FIDIntelligence
}

// CapitalReadinessScore answers: "How prepared is this project to absorb
// the capital it says it needs?" This is distinct from InvestmentReadiness
// (which measures the project's overall completion) and from scoring
// (which measures relative attractiveness).
type CapitalReadinessScore struct {
	ProjectID          string              `json:"project_id"`
	Score              float64             `json:"score"` // 0-100
	MethodologyVersion string              `json:"methodology_version"`
	Factors            map[string]float64  `json:"factors"`
	FactorEvidence     map[string][]string `json:"factor_evidence"`
	UnknownFactors     []string            `json:"unknown_factors,omitempty"`
	Explanation        []string            `json:"explanation"`
	CalculatedAt       time.Time           `json:"calculated_at"`
}

// CalculateCapitalReadiness produces a decomposed score for a project's
// readiness to receive and deploy capital.
func CalculateCapitalReadiness(inputs CapitalReadinessInputs, calculatedAt time.Time) *CapitalReadinessScore {
	if inputs.Project == nil {
		return nil
	}

	factors := map[string]float64{}
	evidence := map[string][]string{}
	unknown := []string{}
	explanation := []string{}

	// Factor 1: Engineering maturity (has project completed FEED/detailed engineering?).
	eng := engineeringMaturity(inputs.Milestones)
	factors["engineering_maturity"] = eng.value
	evidence["engineering_maturity"] = eng.evidence
	if !eng.known {
		unknown = append(unknown, "engineering_maturity")
	}

	// Factor 2: Regulatory/permit readiness.
	reg := regulatoryReadiness(inputs.Milestones)
	factors["regulatory_readiness"] = reg.value
	evidence["regulatory_readiness"] = reg.evidence
	if !reg.known {
		unknown = append(unknown, "regulatory_readiness")
	}

	// Factor 3: Commercial foundation (offtake, anchor customer).
	com := commercialFoundation(inputs.Milestones, inputs.CapitalNeeds)
	factors["commercial_foundation"] = com.value
	evidence["commercial_foundation"] = com.evidence
	if !com.known {
		unknown = append(unknown, "commercial_foundation")
	}

	// Factor 4: Site control.
	site := siteControl(inputs.Milestones)
	factors["site_control"] = site.value
	evidence["site_control"] = site.evidence
	if !site.known {
		unknown = append(unknown, "site_control")
	}

	// Factor 5: Capital already committed (gap closure).
	gap := gapClosure(inputs.CapitalItems, inputs.Project)
	factors["committed_capital"] = gap.value
	evidence["committed_capital"] = gap.evidence
	if !gap.known {
		unknown = append(unknown, "committed_capital")
	}

	// Factor 6: FID proximity.
	fid := fidProximity(inputs.FID)
	factors["fid_proximity"] = fid.value
	evidence["fid_proximity"] = fid.evidence
	if !fid.known {
		unknown = append(unknown, "fid_proximity")
	}

	// Factor 7: Financing specificity (does the project know what it needs?).
	spec := financingSpecificity(inputs.CapitalNeeds)
	factors["financing_specificity"] = spec.value
	evidence["financing_specificity"] = spec.evidence
	if !spec.known {
		unknown = append(unknown, "financing_specificity")
	}

	// Weighted composite.
	score := roundScore(
		eng.value*0.18 +
			reg.value*0.15 +
			com.value*0.15 +
			site.value*0.10 +
			gap.value*0.17 +
			fid.value*0.10 +
			spec.value*0.15)

	sort.Strings(unknown)
	sort.Strings(explanation)

	return &CapitalReadinessScore{
		ProjectID:          inputs.Project.ID,
		Score:              score,
		MethodologyVersion: CapitalReadinessVersion,
		Factors:            factors,
		FactorEvidence:     evidence,
		UnknownFactors:     unknown,
		Explanation:        explanation,
		CalculatedAt:       calculatedAt,
	}
}

// --- Internal factor assessments ---

type factorResult struct {
	value    float64
	known    bool
	evidence []string
}

func engineeringMaturity(milestones []*domain.Milestone) factorResult {
	r := factorResult{}
	for _, m := range milestones {
		switch m.Type {
		case domain.MilestoneFEEDComplete:
			if m.Status == domain.MilestoneComplete {
				r.value = 80
				r.known = true
			} else if m.Status == domain.MilestoneInProgress {
				r.value = 50
				r.known = true
			}
		case domain.MilestoneFeasibilityComplete:
			if m.Status == domain.MilestoneComplete && r.value < 40 {
				r.value = 40
				r.known = true
			}
		}
		r.evidence = append(r.evidence, m.EvidenceIDs...)
	}
	return r
}

func regulatoryReadiness(milestones []*domain.Milestone) factorResult {
	r := factorResult{}
	envDone, permitDone := false, false
	for _, m := range milestones {
		if m.Type == domain.MilestoneEnvironmentalApproval && m.Status == domain.MilestoneComplete {
			envDone = true
			r.evidence = append(r.evidence, m.EvidenceIDs...)
		}
		if m.Type == domain.MilestonePermitReceived && m.Status == domain.MilestoneComplete {
			permitDone = true
			r.evidence = append(r.evidence, m.EvidenceIDs...)
		}
	}
	if envDone && permitDone {
		r.value, r.known = 100, true
	} else if envDone || permitDone {
		r.value, r.known = 60, true
	}
	return r
}

func commercialFoundation(milestones []*domain.Milestone, needs []*domain.CapitalNeed) factorResult {
	r := factorResult{}
	for _, m := range milestones {
		if m.Type == domain.MilestoneOfftakeSigned && m.Status == domain.MilestoneComplete {
			r.value = 90
			r.known = true
			r.evidence = append(r.evidence, m.EvidenceIDs...)
			return r
		}
	}
	// Check if capital needs mention commercial partnerships.
	for _, n := range needs {
		for _, t := range n.Types {
			if t == domain.NeedOfftake || t == domain.NeedAnchorTenant {
				r.value = 30
				r.known = true
				return r
			}
		}
	}
	return r
}

func siteControl(milestones []*domain.Milestone) factorResult {
	r := factorResult{}
	for _, m := range milestones {
		if (m.Type == domain.MilestoneLandSecured || m.Type == domain.MilestoneSiteSelected) &&
			m.Status == domain.MilestoneComplete {
			r.value = 100
			r.known = true
			r.evidence = append(r.evidence, m.EvidenceIDs...)
			return r
		}
	}
	return r
}

func gapClosure(items []*domain.CapitalItem, project *domain.Project) factorResult {
	r := factorResult{}
	if project.CapexCAD == 0 {
		return r
	}
	var committed int64
	for _, item := range items {
		if item.Status == domain.CapitalCommitted || item.Status == domain.CapitalSigned ||
			item.Status == domain.CapitalDisbursed || item.Status == domain.CapitalClosed {
			committed += item.AmountCAD
			if item.EvidenceID != "" {
				r.evidence = append(r.evidence, item.EvidenceID)
			}
		}
	}
	ratio := float64(committed) / float64(project.CapexCAD)
	if ratio > 1 {
		ratio = 1
	}
	r.value = ratio * 100
	r.known = true
	return r
}

func fidProximity(fid *domain.FIDIntelligence) factorResult {
	r := factorResult{}
	if fid == nil {
		return r
	}
	r.evidence = fid.EvidenceIDs
	switch fid.Status {
	case domain.FIDAchieved:
		r.value, r.known = 100, true
	case domain.FIDExpected:
		r.value, r.known = 80, true
	case domain.FIDTarget:
		r.value, r.known = 60, true
	case domain.FIDDelayed, domain.FIDDeferred:
		r.value, r.known = 20, true
	case domain.FIDNegative:
		r.value, r.known = 0, true
	default:
		r.known = false
	}
	return r
}

func financingSpecificity(needs []*domain.CapitalNeed) factorResult {
	r := factorResult{}
	if len(needs) == 0 {
		return r
	}
	score := 0.0
	for _, n := range needs {
		if len(n.Types) > 0 {
			score += 30
		}
		if n.Amount != nil && n.Amount.Amount != nil {
			score += 40
		}
		if len(n.Counterparties) > 0 {
			score += 30
		}
	}
	r.value = score / float64(len(needs))
	if r.value > 100 {
		r.value = 100
	}
	r.known = true
	return r
}

func roundScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return float64(int(v*100+0.5)) / 100
}
