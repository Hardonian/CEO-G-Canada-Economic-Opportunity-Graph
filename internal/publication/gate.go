// Package publication owns the one-way boundary between restricted workspace
// intelligence and public products. Public callers should use these predicates
// instead of duplicating visibility checks.
package publication

import (
	"bytes"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func PublicEvidence(e *domain.Evidence) bool {
	return e != nil && e.Visibility.Public() && e.Publishable
}

func PublicClaim(c *domain.Claim) bool {
	if c == nil || !c.SourceVisibility.Public() || !c.Publishable || c.PublicationState != domain.PublicationPublicCanonical {
		return false
	}
	switch c.Status {
	case domain.ConfidenceReported, domain.ConfidenceSupported, domain.ConfidenceVerified,
		domain.ConfidenceInferred, domain.ConfidenceStale:
		return true
	default:
		return false
	}
}

func PublicOpportunity(o *domain.Opportunity) bool {
	return o != nil && o.Visibility.Public() && o.Publishable && o.PublicationState == domain.PublicationPublicCanonical
}

func PublicCapitalNeed(n *domain.CapitalNeed) bool {
	return n != nil && n.Visibility.Public() && n.Publishable && n.PublicationState == domain.PublicationPublicCanonical
}

func PublicMilestone(m *domain.Milestone) bool {
	return m != nil && m.Visibility.Public() && m.Publishable
}

func PublicCapitalItem(c *domain.CapitalItem) bool {
	return c != nil && c.Visibility.Public() && c.Publishable
}

func PublicCapitalRequirement(c *domain.CapitalRequirement) bool {
	return c != nil && c.Visibility.Public() && c.Publishable
}

// ReconcileClaims promotes only the publishable public claims. Restricted
// claims participate as candidate context, but can never be returned as public
// canonical records or be used as the public evidence carrier.
func ReconcileClaims(candidate *domain.Claim, observations []*domain.Claim, actor string, now time.Time) ([]*domain.Claim, []*domain.AuditEntry) {
	if candidate == nil {
		return nil, nil
	}
	matching := make([]*domain.Claim, 0)
	conflicting := make([]*domain.Claim, 0)
	for _, observation := range observations {
		if observation == nil || !observation.SourceVisibility.Public() || !observation.Publishable {
			continue
		}
		if observation.SubjectID != candidate.SubjectID || observation.Predicate != candidate.Predicate {
			continue
		}
		if equivalentValue(candidate.Value, observation.Value) {
			matching = append(matching, observation)
		} else {
			conflicting = append(conflicting, observation)
		}
	}

	audits := make([]*domain.AuditEntry, 0)
	if len(matching) == 0 {
		copyCandidate := *candidate
		copyCandidate.Publishable = false
		copyCandidate.PublicationState = domain.PublicationPrivateOnly
		if len(conflicting) > 0 {
			copyCandidate.Status = domain.ConfidenceConflict
			copyCandidate.PublicationState = domain.PublicationConflicted
			copyCandidate.ConflictsWith = claimIDs(conflicting)
		}
		return []*domain.Claim{&copyCandidate}, audits
	}

	independent := independentSources(matching)
	result := make([]*domain.Claim, 0, len(matching)+1)
	privateCopy := *candidate
	privateCopy.Publishable = false
	privateCopy.PublicationState = domain.PublicationPrivateOnly
	result = append(result, &privateCopy)
	for _, match := range matching {
		promoted := *match
		promoted.PublicationState = domain.PublicationPublicCanonical
		promoted.CorroborationCount = len(matching)
		promoted.IndependentSourceCount = independent
		if independent > 1 && promoted.Status == domain.ConfidenceReported {
			promoted.Status = domain.ConfidenceSupported
		}
		result = append(result, &promoted)
	}
	audits = append(audits, &domain.AuditEntry{
		ID: "audit-promote-" + candidate.ID,
		Action: domain.AuditClaimPromoted,
		Actor: actor,
		SubjectID: candidate.ID,
		Reason: "Matching publishable public evidence independently corroborated the restricted candidate claim.",
		FromVisibility: candidate.SourceVisibility,
		ToVisibility: matching[0].SourceVisibility,
		OccurredAt: now.UTC(),
	})
	return result, audits
}

func equivalentValue(left, right []byte) bool {
	return bytes.Equal(bytes.TrimSpace(left), bytes.TrimSpace(right)) ||
		strings.EqualFold(strings.TrimSpace(string(left)), strings.TrimSpace(string(right)))
}

func independentSources(claims []*domain.Claim) int {
	seen := map[string]struct{}{}
	for _, claim := range claims {
		key := strings.TrimSpace(claim.SourceID)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	return len(seen)
}

func claimIDs(claims []*domain.Claim) []string {
	ids := make([]string, 0, len(claims))
	for _, claim := range claims {
		ids = append(ids, claim.ID)
	}
	sort.Strings(ids)
	return ids
}
