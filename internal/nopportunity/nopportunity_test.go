package opportunity

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestGenerateOpportunities_NilProject(t *testing.T) {
	now := time.Now()
	if got := GenerateOpportunities(Context{Project: nil}, now); got != nil {
		t.Fatalf("expected nil for nil project, got %v", got)
	}
}

func TestGenerateOpportunities_EmptyContext(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	got := GenerateOpportunities(Context{Project: project}, now)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 opportunities for empty context, got %d", len(got))
	}
}

func TestGenerateOpportunities_CapitalNeeds(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	needs := []*domain.CapitalNeed{
		{ID: "n1", Types: []domain.CapitalNeedType{domain.NeedEquity}, EvidenceIDs: []string{"e1"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
		{ID: "n2", Types: []domain.CapitalNeedType{domain.NeedDebt}, EvidenceIDs: []string{"e2"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
	}
	got := GenerateOpportunities(Context{
		Project:      project,
		CapitalNeeds: needs,
	}, now)
	capitalCount := 0
	for _, opp := range got {
		if opp.OpportunityKind == "CAPITAL" {
			capitalCount++
			if opp.CapitalNeedID != "n1" && opp.CapitalNeedID != "n2" {
				t.Fatalf("unexpected capital need ID: %s", opp.CapitalNeedID)
			}
		}
	}
	if capitalCount != 2 {
		t.Fatalf("expected 2 capital opportunities, got %d", capitalCount)
	}
}

func TestGenerateOpportunities_Requirements(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	reqs := []*domain.ProjectRequirement{
		{ID: "r1", Type: domain.RequirePower, Description: "Power needed", Confidence: domain.RequirementDerived, Visibility: domain.VisibilityPublic, Publishable: true, EvidenceIDs: []string{"e1"}},
		{ID: "r2", Type: domain.RequireRoad, Description: "Road needed", Confidence: domain.RequirementStated, Visibility: domain.VisibilityPublic, Publishable: true, EvidenceIDs: []string{"e2"}},
	}
	got := GenerateOpportunities(Context{
		Project:       project,
		Requirements:  reqs,
	}, now)
	procurementCount := 0
	for _, opp := range got {
		if opp.OpportunityKind == "PROCUREMENT" || opp.OpportunityKind == "PARTNERSHIP" {
			procurementCount++
		}
	}
	if procurementCount != 2 {
		t.Fatalf("expected 2 requirement opportunities, got %d", procurementCount)
	}
}

func TestGenerateOpportunities_FilteredBySatisfied(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	reqs := []*domain.ProjectRequirement{
		{ID: "r1", Type: domain.RequirePower, SatisfiedBy: "other-project", Visibility: domain.VisibilityPublic, Publishable: true},
	}
	got := GenerateOpportunities(Context{Project: project, Requirements: reqs}, now)
	for _, opp := range got {
		if opp.OpportunityKind == "PROCUREMENT" || opp.OpportunityKind == "PARTNERSHIP" {
			t.Fatalf("satisfied requirement should not generate opportunity, got %v", opp)
		}
	}
}

func TestGenerateOpportunities_Offtake(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	needs := []*domain.CapitalNeed{
		{ID: "n1", Types: []domain.CapitalNeedType{domain.NeedOfftake}, Counterparties: []domain.CounterpartyType{domain.CounterpartyOfftaker}},
		{ID: "n2", Types: []domain.CapitalNeedType{domain.NeedAnchorTenant}, Counterparties: []domain.CounterpartyType{domain.CounterpartyAnchorTenant}},
	}
	got := GenerateOpportunities(Context{Project: project, CapitalNeeds: needs}, now)
	offtakeCount := 0
	tenancyCount := 0
	for _, opp := range got {
		switch opp.OpportunityKind {
		case "OFFTAKE":
			offtakeCount++
		case "TENANCY":
			tenancyCount++
		}
	}
	if offtakeCount != 1 {
		t.Fatalf("expected 1 offtake opportunity, got %d", offtakeCount)
	}
	if tenancyCount != 1 {
		t.Fatalf("expected 1 tenancy opportunity, got %d", tenancyCount)
	}
}

func TestGenerateOpportunities_FunnelStates(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	needs := []*domain.CapitalNeed{
		{ID: "n1", Types: []domain.CapitalNeedType{domain.NeedEquity}, EvidenceIDs: []string{"e1", "e2"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
		{ID: "n2", Types: []domain.CapitalNeedType{domain.NeedDebt}, EvidenceIDs: []string{"e3"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
		{ID: "n3", Types: []domain.CapitalNeedType{domain.NeedEquity}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
	}
	got := GenerateOpportunities(Context{Project: project, CapitalNeeds: needs}, now)
	for _, opp := range got {
		switch {
		case opp.PublicationState == domain.PublicationPublicCanonical && len(opp.EvidenceIDs) >= 2:
			if opp.FunnelState != domain.FunnelCorroborated {
				t.Fatalf("expected CORROBORATED for %s, got %s", opp.ID, opp.FunnelState)
			}
		case opp.PublicationState == domain.PublicationPublicCanonical:
			if opp.FunnelState != domain.FunnelActiveOpportunity {
				t.Fatalf("expected ACTIVE_OPPORTUNITY for %s, got %s", opp.ID, opp.FunnelState)
			}
		case len(opp.EvidenceIDs) > 0:
			if opp.FunnelState != domain.FunnelQualifying {
				t.Fatalf("expected QUALIFYING for %s, got %s", opp.ID, opp.FunnelState)
			}
		default:
			if opp.FunnelState != domain.FunnelDiscovered {
				t.Fatalf("expected DISCOVERED for %s, got %s", opp.ID, opp.FunnelState)
			}
		}
	}
}

func TestGenerateOpportunities_Sorted(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	needs := []*domain.CapitalNeed{
		{ID: "n2", Types: []domain.CapitalNeedType{domain.NeedDebt}, EvidenceIDs: []string{"e1"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
		{ID: "n1", Types: []domain.CapitalNeedType{domain.NeedEquity}, EvidenceIDs: []string{"e2"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
	}
	got := GenerateOpportunities(Context{Project: project, CapitalNeeds: needs}, now)
	for i := 1; i < len(got); i++ {
		if got[i-1].ID > got[i].ID {
			t.Fatalf("results not sorted by ID: %s > %s", got[i-1].ID, got[i].ID)
		}
	}
}

func TestGenerateOpportunities_TitleFormat(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "My Project", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	needs := []*domain.CapitalNeed{
		{ID: "n1", Types: []domain.CapitalNeedType{domain.NeedEquity, domain.NeedDebt}, EvidenceIDs: []string{"e1"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
	}
	got := GenerateOpportunities(Context{Project: project, CapitalNeeds: needs}, now)
	for _, opp := range got {
		if opp.Title == "" {
			t.Fatalf("title should not be empty for %s", opp.ID)
		}
		if opp.ProjectName != "My Project" {
			t.Fatalf("wrong project name in opportunity: %s", opp.ProjectName)
		}
	}
}

func TestGenerateOpportunities_InheritsNeedProperties(t *testing.T) {
	now := time.Now()
	project := &domain.Project{ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: now}
	needs := []*domain.CapitalNeed{
		{ID: "n1", Types: []domain.CapitalNeedType{domain.NeedEquity}, Counterparties: []domain.CounterpartyType{domain.CounterpartyInfrastructureFund}, EvidenceIDs: []string{"e1"}, Visibility: domain.VisibilityPublic, Publishable: true, PublicationState: domain.PublicationPublicCanonical},
	}
	got := GenerateOpportunities(Context{Project: project, CapitalNeeds: needs}, now)
	if len(got) != 1 {
		t.Fatalf("expected 1 opportunity, got %d", len(got))
	}
	opp := got[0]
	if opp.CapitalNeedID != "n1" {
		t.Fatalf("wrong capital_need_id: %s", opp.CapitalNeedID)
	}
	if len(opp.Counterparties) != 1 || opp.Counterparties[0] != domain.CounterpartyInfrastructureFund {
		t.Fatalf("wrong counterparties: %v", opp.Counterparties)
	}
	if len(opp.Instruments) != 1 || opp.Instruments[0] != domain.NeedEquity {
		t.Fatalf("wrong instruments: %v", opp.Instruments)
	}
}
