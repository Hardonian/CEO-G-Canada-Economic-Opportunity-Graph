package publication

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func claim(id, source string, visibility domain.VisibilityClass, value string) *domain.Claim {
	raw, _ := json.Marshal(value)
	return &domain.Claim{ID: id, SubjectID: "candidate-1", Predicate: "capex", Value: raw,
		SourceID: source, SourceVisibility: visibility, Publishable: visibility.Public(),
		Status: domain.ConfidenceReported, ObservedAt: time.Unix(1, 0)}
}

func TestRestrictedClaimRequiresIndependentPublishableObservation(t *testing.T) {
	private := claim("private", "prospectus", domain.VisibilityLicensedPrivate, "USD 1.4B")
	withoutPublic, _ := ReconcileClaims(private, nil, "system", time.Unix(2, 0))
	if len(withoutPublic) != 1 || PublicClaim(withoutPublic[0]) {
		t.Fatalf("restricted-only claim crossed publication gate: %#v", withoutPublic)
	}

	public := claim("public", "issuer-release", domain.VisibilityPublicAttribution, "USD 1.4B")
	reconciled, audits := ReconcileClaims(private, []*domain.Claim{public}, "system", time.Unix(2, 0))
	if len(reconciled) != 2 || PublicClaim(reconciled[0]) || !PublicClaim(reconciled[1]) {
		t.Fatalf("unexpected reconciliation result: %#v", reconciled)
	}
	if reconciled[1].EvidenceID == private.EvidenceID || len(audits) != 1 {
		t.Fatal("promotion did not preserve the public evidence carrier and audit event")
	}
}

func TestPublicConflictDoesNotPromotePrivateValue(t *testing.T) {
	private := claim("private", "prospectus", domain.VisibilityInternalRestricted, "CAD 2B")
	public := claim("public", "regulator", domain.VisibilityPublicAttribution, "CAD 2.6B")
	result, _ := ReconcileClaims(private, []*domain.Claim{public}, "system", time.Now())
	if len(result) != 1 || result[0].PublicationState != domain.PublicationConflicted || PublicClaim(result[0]) {
		t.Fatalf("conflict was not held behind gate: %#v", result)
	}
}
