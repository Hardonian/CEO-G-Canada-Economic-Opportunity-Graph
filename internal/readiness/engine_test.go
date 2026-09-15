package readiness

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestReadinessIsDeterministicAndDecomposed(t *testing.T) {
	project := &domain.Project{ID: "project-1", CurrentStage: domain.StageFEED}
	milestones := []*domain.Milestone{
		{Type: domain.MilestoneFEEDComplete, Status: domain.MilestoneComplete, EvidenceIDs: []string{"e-feed"}},
		{Type: domain.MilestonePermitReceived, Status: domain.MilestoneInProgress, EvidenceIDs: []string{"e-permit"}},
	}
	need := &domain.CapitalNeed{Types: []domain.CapitalNeedType{domain.NeedProjectFinance}, EvidenceIDs: []string{"e-need"}}
	first := Calculate(Inputs{Project: project, Milestones: milestones, CapitalNeeds: []*domain.CapitalNeed{need}}, time.Unix(10, 0))
	second := Calculate(Inputs{Project: project, Milestones: milestones, CapitalNeeds: []*domain.CapitalNeed{need}}, time.Unix(20, 0))
	if first.InputHash != second.InputHash || first.InvestmentReadiness != second.InvestmentReadiness || first.Vector != second.Vector {
		t.Fatalf("same canonical state changed score: %#v %#v", first, second)
	}
	if first.Vector.Engineering != 100 || first.Vector.Regulatory != 60 || first.Coverage <= 0 || len(first.UnknownFactors) == 0 {
		t.Fatalf("unexpected decomposition: %#v", first)
	}
}

func TestCapitalGapDoesNotCountSeekingOrContingentCapital(t *testing.T) {
	total := int64(1_000)
	committed := int64(300)
	seeking := int64(500)
	requirement := &domain.CapitalRequirement{Amount: domain.MonetaryAmount{Amount: &total, Currency: "USD", AmountType: domain.AmountExact}}
	items := []*domain.CapitalItem{
		{Status: domain.CapitalCommitted, StackTreatment: "ADDITIVE", OriginalAmount: &domain.MonetaryAmount{Amount: &committed, Currency: "USD", AmountType: domain.AmountExact}},
		{Status: domain.CapitalSeeking, StackTreatment: "ADDITIVE", OriginalAmount: &domain.MonetaryAmount{Amount: &seeking, Currency: "USD", AmountType: domain.AmountExact}},
		{Status: domain.CapitalCommitted, StackTreatment: "CONTINGENT", OriginalAmount: &domain.MonetaryAmount{Amount: &seeking, Currency: "USD", AmountType: domain.AmountExact}},
	}
	gap := CalculateCapitalGap("project-1", requirement, items, false)
	if gap.State != domain.CapitalGapPartial || gap.PotentialGap == nil || *gap.PotentialGap.Amount != 700 {
		t.Fatalf("unsafe gap calculation: %#v", gap)
	}
}
