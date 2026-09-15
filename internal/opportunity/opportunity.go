// Package opportunity generates typed, evidence-linked investment opportunities
// from a project's capital needs, milestones, and requirements. Each opportunity
// is a first-class graph node with its own funnel state, not a passive projection
// of a project record.
package opportunity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const MethodologyVersion = "opportunity-generator-v1.0"

// Context is the input boundary for opportunity generation. It mirrors the
// scoring engine pattern: all inputs are supplied, nothing is fetched.
type Context struct {
	Project      *domain.Project
	CapitalNeeds []*domain.CapitalNeed
	Milestones   []*domain.Milestone
	Requirements []*domain.ProjectRequirement
	CapitalItems []*domain.CapitalItem
	FID          *domain.FIDIntelligence
}

// GenerateOpportunities produces typed opportunities from a project context.
// Each capital need produces a CAPITAL opportunity. Each unsatisfied requirement
// produces a PROCUREMENT or PARTNERSHIP opportunity. Opportunities inherit the
// visibility of their source records.
func GenerateOpportunities(ctx Context, generatedAt time.Time) []*domain.Opportunity {
	if ctx.Project == nil {
		return nil
	}
	var result []*domain.Opportunity
	result = append(result, capitalOpportunities(ctx, generatedAt)...)
	result = append(result, requirementOpportunities(ctx, generatedAt)...)
	result = append(result, offtakeOpportunities(ctx, generatedAt)...)

	// Assign funnel states based on evidence quality and milestone status.
	for _, opp := range result {
		opp.FunnelState = classifyFunnelState(opp, ctx)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func capitalOpportunities(ctx Context, at time.Time) []*domain.Opportunity {
	result := make([]*domain.Opportunity, 0, len(ctx.CapitalNeeds))
	for _, need := range ctx.CapitalNeeds {
		if need == nil {
			continue
		}
		opp := &domain.Opportunity{
			ID:               stableID(ctx.Project.ID + "|capital|" + need.ID),
			ProjectID:        ctx.Project.ID,
			ProjectName:      ctx.Project.Name,
			Title:            capitalTitle(ctx.Project.Name, need),
			Sector:           ctx.Project.Sector,
			RequirementClass: domain.RequirementConfirmed,
			Category:         "capital",
			OpportunityKind:  "CAPITAL",
			CapitalNeedID:    need.ID,
			Instruments:      need.Types,
			Counterparties:   need.Counterparties,
			EstimatedAmount:  need.Amount,
			EvidenceIDs:      need.EvidenceIDs,
			EvidenceQuality:  domain.ConfidenceReported,
			Visibility:       need.Visibility,
			Publishable:      need.Publishable,
			PublicationState: need.PublicationState,
			TriggerMilestone: triggerMilestoneForCapital(ctx),
			CreatedAt:        at,
			UpdatedAt:        at,
		}
		if need.Amount != nil && need.Amount.Amount != nil {
			opp.EstimatedCAD = *need.Amount.Amount
		}
		if len(need.EvidenceIDs) > 0 {
			opp.EvidenceQuality = domain.ConfidenceSupported
		}
		result = append(result, opp)
	}
	return result
}

func requirementOpportunities(ctx Context, at time.Time) []*domain.Opportunity {
	result := make([]*domain.Opportunity, 0, len(ctx.Requirements))
	for _, req := range ctx.Requirements {
		if req == nil || req.SatisfiedBy != "" {
			continue
		}
		kind := "PROCUREMENT"
		if req.Type == domain.RequirePower || req.Type == domain.RequireTransmission ||
			req.Type == domain.RequirePort || req.Type == domain.RequireRail {
			kind = "PARTNERSHIP"
		}

		reqClass := domain.RequirementDerived
		if req.Confidence == domain.RequirementStated {
			reqClass = domain.RequirementConfirmed
		}

		opp := &domain.Opportunity{
			ID:               stableID(ctx.Project.ID + "|requirement|" + req.ID),
			ProjectID:        ctx.Project.ID,
			ProjectName:      ctx.Project.Name,
			Title:            fmt.Sprintf("%s — %s requirement", ctx.Project.Name, strings.ReplaceAll(string(req.Type), "_", " ")),
			Sector:           ctx.Project.Sector,
			RequirementClass: reqClass,
			Category:         strings.ToLower(string(req.Type)),
			OpportunityKind:  kind,
			Description:      req.Description,
			EvidenceIDs:      req.EvidenceIDs,
			Visibility:       req.Visibility,
			Publishable:      req.Publishable,
			CreatedAt:        at,
			UpdatedAt:        at,
		}
		result = append(result, opp)
	}
	return result
}

func offtakeOpportunities(ctx Context, at time.Time) []*domain.Opportunity {
	var result []*domain.Opportunity
	for _, need := range ctx.CapitalNeeds {
		if need == nil {
			continue
		}
		for _, t := range need.Types {
			if t == domain.NeedOfftake || t == domain.NeedAnchorTenant {
				kind := "OFFTAKE"
				if t == domain.NeedAnchorTenant {
					kind = "TENANCY"
				}
				opp := &domain.Opportunity{
					ID:               stableID(ctx.Project.ID + "|" + string(t) + "|" + need.ID),
					ProjectID:        ctx.Project.ID,
					ProjectName:      ctx.Project.Name,
					Title:            fmt.Sprintf("%s — %s opportunity", ctx.Project.Name, strings.ToLower(string(t))),
					Sector:           ctx.Project.Sector,
					RequirementClass: domain.RequirementConfirmed,
					Category:         "commercial",
					OpportunityKind:  kind,
					CapitalNeedID:    need.ID,
					Counterparties:   need.Counterparties,
					EvidenceIDs:      need.EvidenceIDs,
					Visibility:       need.Visibility,
					Publishable:      need.Publishable,
					PublicationState: need.PublicationState,
					CreatedAt:        at,
					UpdatedAt:        at,
				}
				result = append(result, opp)
				break // one opportunity per need for this type
			}
		}
	}
	return result
}

func classifyFunnelState(opp *domain.Opportunity, ctx Context) domain.OpportunityFunnelState {
	switch {
	case opp.PublicationState == domain.PublicationPublicCanonical && len(opp.EvidenceIDs) >= 2:
		return domain.FunnelCorroborated
	case opp.PublicationState == domain.PublicationPublicCanonical:
		return domain.FunnelActiveOpportunity
	case len(opp.EvidenceIDs) > 0:
		return domain.FunnelQualifying
	default:
		return domain.FunnelDiscovered
	}
}

func capitalTitle(projectName string, need *domain.CapitalNeed) string {
	if len(need.Types) == 0 {
		return projectName + " — capital opportunity"
	}
	labels := make([]string, 0, len(need.Types))
	for _, t := range need.Types {
		labels = append(labels, strings.ToLower(strings.ReplaceAll(string(t), "_", " ")))
	}
	return fmt.Sprintf("%s — %s", projectName, strings.Join(labels, " / "))
}

func triggerMilestoneForCapital(ctx Context) string {
	if ctx.FID != nil && ctx.FID.Status != domain.FIDNotApplicable && ctx.FID.Status != domain.FIDAchieved {
		return "FID"
	}
	for _, m := range ctx.Milestones {
		if m.Type == domain.MilestoneFinancialClose && m.Status != domain.MilestoneComplete {
			return "FINANCIAL_CLOSE"
		}
	}
	return ""
}

func stableID(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return "opp-" + hex.EncodeToString(sum[:12])
}
