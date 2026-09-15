package publication

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestPublicEvidence_RestrictedNeverPasses(t *testing.T) {
	cases := []struct {
		name       string
		visibility domain.VisibilityClass
		pub        bool
		want       bool
	}{
		{"public+publishable", domain.VisibilityPublic, true, true},
		{"public+not_publishable", domain.VisibilityPublic, false, false},
		{"attribution+publishable", domain.VisibilityPublicAttribution, true, true},
		{"licensed_private", domain.VisibilityLicensedPrivate, true, false},
		{"user_private", domain.VisibilityUserPrivate, true, false},
		{"internal_restricted", domain.VisibilityInternalRestricted, true, false},
		{"nil", domain.VisibilityPublic, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := &domain.Evidence{Visibility: tc.visibility, Publishable: tc.pub}
			if got := PublicEvidence(e); got != tc.want {
				t.Errorf("PublicEvidence(%s, publishable=%v) = %v, want %v",
					tc.visibility, tc.pub, got, tc.want)
			}
		})
	}
	t.Run("nil_evidence", func(t *testing.T) {
		if PublicEvidence(nil) {
			t.Error("PublicEvidence(nil) should be false")
		}
	})
}

func TestPublicClaim_PrivateNeverPromoted(t *testing.T) {
	now := time.Now()
	baseClaim := func(vis domain.VisibilityClass, pub bool, state domain.PublicationState) *domain.Claim {
		return &domain.Claim{
			ID:               "test-claim",
			SourceVisibility: vis,
			Publishable:      pub,
			PublicationState: state,
			Status:           domain.ConfidenceReported,
			CreatedAt:        now,
		}
	}

	cases := []struct {
		name string
		c    *domain.Claim
		want bool
	}{
		{"public_canonical", baseClaim(domain.VisibilityPublic, true, domain.PublicationPublicCanonical), true},
		{"private_canonical_blocked", baseClaim(domain.VisibilityLicensedPrivate, true, domain.PublicationPublicCanonical), false},
		{"public_private_only_blocked", baseClaim(domain.VisibilityPublic, true, domain.PublicationPrivateOnly), false},
		{"restricted_blocked", baseClaim(domain.VisibilityInternalRestricted, true, domain.PublicationPublicCanonical), false},
		{"not_publishable_blocked", baseClaim(domain.VisibilityPublic, false, domain.PublicationPublicCanonical), false},
		{"unverified_blocked", baseClaim(domain.VisibilityPublic, true, domain.PublicationUnverified), false},
		{"conflicted_blocked", baseClaim(domain.VisibilityPublic, true, domain.PublicationConflicted), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PublicClaim(tc.c); got != tc.want {
				t.Errorf("PublicClaim(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
	t.Run("nil_claim", func(t *testing.T) {
		if PublicClaim(nil) {
			t.Error("PublicClaim(nil) should be false")
		}
	})
}

func TestPublicOpportunity_PrivateNeverLeaks(t *testing.T) {
	opp := func(vis domain.VisibilityClass, pub bool, state domain.PublicationState) *domain.Opportunity {
		return &domain.Opportunity{
			Visibility:       vis,
			Publishable:      pub,
			PublicationState: state,
		}
	}
	cases := []struct {
		name string
		o    *domain.Opportunity
		want bool
	}{
		{"public_canonical", opp(domain.VisibilityPublic, true, domain.PublicationPublicCanonical), true},
		{"private_blocked", opp(domain.VisibilityLicensedPrivate, true, domain.PublicationPublicCanonical), false},
		{"not_publishable", opp(domain.VisibilityPublic, false, domain.PublicationPublicCanonical), false},
		{"not_canonical", opp(domain.VisibilityPublic, true, domain.PublicationPrivateOnly), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := PublicOpportunity(tc.o); got != tc.want {
				t.Errorf("PublicOpportunity(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestPublicCapitalNeed_PrivateNeverLeaks(t *testing.T) {
	need := func(vis domain.VisibilityClass, pub bool, state domain.PublicationState) *domain.CapitalNeed {
		return &domain.CapitalNeed{
			Visibility:       vis,
			Publishable:      pub,
			PublicationState: state,
		}
	}
	if PublicCapitalNeed(need(domain.VisibilityLicensedPrivate, true, domain.PublicationPublicCanonical)) {
		t.Error("licensed private capital need should not pass")
	}
	if PublicCapitalNeed(need(domain.VisibilityUserPrivate, true, domain.PublicationPublicCanonical)) {
		t.Error("user private capital need should not pass")
	}
	if !PublicCapitalNeed(need(domain.VisibilityPublic, true, domain.PublicationPublicCanonical)) {
		t.Error("public canonical capital need should pass")
	}
}

func TestReconcileClaims_RestrictedStaysPrivate(t *testing.T) {
	now := time.Now()
	restricted := &domain.Claim{
		ID:               "restricted-1",
		SubjectID:        "proj-1",
		Predicate:        "capex_cad",
		Value:            []byte(`1000000000`),
		SourceID:         "src-private",
		SourceVisibility: domain.VisibilityInternalRestricted,
		Status:           domain.ConfidenceReported,
		Publishable:      false,
		PublicationState: domain.PublicationPrivateOnly,
		CreatedAt:        now,
	}

	t.Run("restricted_only_no_corroboration", func(t *testing.T) {
		results, _ := ReconcileClaims(restricted, nil, "system", now)
		for _, r := range results {
			if r.Publishable {
				t.Error("restricted claim without corroboration must not become publishable")
			}
			if r.PublicationState == domain.PublicationPublicCanonical {
				t.Error("restricted claim must not reach PUBLIC_CANONICAL without public corroboration")
			}
		}
	})

	t.Run("restricted_with_public_corroboration", func(t *testing.T) {
		publicObs := &domain.Claim{
			ID:               "public-1",
			SubjectID:        "proj-1",
			Predicate:        "capex_cad",
			Value:            []byte(`1000000000`),
			SourceID:         "src-public",
			SourceVisibility: domain.VisibilityPublic,
			Status:           domain.ConfidenceReported,
			Publishable:      true,
			PublicationState: domain.PublicationPublicCanonical,
			CreatedAt:        now,
		}
		results, audits := ReconcileClaims(restricted, []*domain.Claim{publicObs}, "system", now)

		// The restricted source claim should remain private.
		restrictedResult := results[0]
		if restrictedResult.Publishable {
			t.Error("original restricted claim should remain not-publishable")
		}
		if restrictedResult.PublicationState != domain.PublicationPrivateOnly {
			t.Errorf("restricted claim should be PRIVATE_ONLY, got %s", restrictedResult.PublicationState)
		}

		// The public corroborating claim should be promoted.
		foundPublic := false
		for _, r := range results {
			if r.ID == "public-1" && r.PublicationState == domain.PublicationPublicCanonical {
				foundPublic = true
			}
		}
		if !foundPublic {
			t.Error("public corroborating claim should be promoted to PUBLIC_CANONICAL")
		}

		if len(audits) == 0 {
			t.Error("promotion should generate an audit entry")
		}
	})

	t.Run("public_conflict", func(t *testing.T) {
		conflicting := &domain.Claim{
			ID:               "public-conflict",
			SubjectID:        "proj-1",
			Predicate:        "capex_cad",
			Value:            []byte(`2000000000`), // Different value
			SourceID:         "src-public-2",
			SourceVisibility: domain.VisibilityPublic,
			Status:           domain.ConfidenceReported,
			Publishable:      true,
			PublicationState: domain.PublicationPublicCanonical,
			CreatedAt:        now,
		}
		results, _ := ReconcileClaims(restricted, []*domain.Claim{conflicting}, "system", now)
		for _, r := range results {
			if r.ID == restricted.ID && r.Status != domain.ConfidenceConflict {
				t.Errorf("restricted claim with conflicting public data should be CONFLICTED, got %s", r.Status)
			}
		}
	})
}
