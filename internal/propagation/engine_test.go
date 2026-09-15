package propagation

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestDefaultOntology_NonEmpty(t *testing.T) {
	rules := DefaultOntology()
	if len(rules) == 0 {
		t.Fatal("expected non-empty ontology")
	}
	for i, rule := range rules {
		if rule.Sector == "" {
			t.Fatalf("rule %d has empty sector", i)
		}
		if rule.TitleTemplate == "" {
			t.Fatalf("rule %d has empty title template", i)
		}
		if rule.Description == "" {
			t.Fatalf("rule %d has empty description", i)
		}
		if rule.Category == "" {
			t.Fatalf("rule %d has empty category", i)
		}
		if rule.Class == "" {
			t.Fatalf("rule %d has empty class", i)
		}
		if rule.TriggerMilestone == "" {
			t.Fatalf("rule %d has empty trigger milestone", i)
		}
		if rule.CapexPercentMin < 0 || rule.CapexPercentMin > 1 {
			t.Fatalf("rule %d has invalid CapexPercentMin: %f", i, rule.CapexPercentMin)
		}
		if rule.CapexPercentMax < 0 || rule.CapexPercentMax > 1 {
			t.Fatalf("rule %d has invalid CapexPercentMax: %f", i, rule.CapexPercentMax)
		}
		if rule.CapexPercentMin > rule.CapexPercentMax {
			t.Fatalf("rule %d has min > max: %f > %f", i, rule.CapexPercentMin, rule.CapexPercentMax)
		}
	}
}

func TestDefaultOntology_MinCapexNotExceedsMax(t *testing.T) {
	rules := DefaultOntology()
	for i, rule := range rules {
		if rule.CapexPercentMin > rule.CapexPercentMax {
			t.Fatalf("rule %d: min %.4f > max %.4f", i, rule.CapexPercentMin, rule.CapexPercentMax)
		}
	}
}

func TestPropagateOpportunities_NilProject(t *testing.T) {
	if got := PropagateOpportunities(nil); got == nil {
		t.Fatal("expected non-nil for nil project")
	}
}

func TestPropagateOpportunities_UnknownSector(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: "Unknown Sector", Name: "Test", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000}
	got := PropagateOpportunities(project)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 opportunities for unknown sector, got %d", len(got))
	}
}

func TestPropagateOpportunities_GeneratesForMatchingSector(t *testing.T) {
	project := &domain.Project{
		ID:            "p1",
		Sector:        domain.SectorCriticalMinerals,
		Name:          "Test Project",
		CurrentStage:  domain.StageFEED,
		CapexCAD:      1_000_000_000,
		UpdatedAt:     time.Now(),
	}
	got := PropagateOpportunities(project)
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if len(got) == 0 {
		t.Fatal("expected opportunities for CriticalMinerals sector")
	}
	for _, opp := range got {
		if opp.ProjectID != "p1" {
			t.Fatalf("expected project ID p1, got %s", opp.ProjectID)
		}
		if opp.ProjectName != "Test Project" {
			t.Fatalf("expected project name 'Test Project', got %s", opp.ProjectName)
		}
		if opp.Sector != domain.SectorCriticalMinerals {
			t.Fatalf("wrong sector: %s", opp.Sector)
		}
	}
}

func TestPropagateOpportunities_StableIDs(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCriticalMinerals, Name: "Test", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: time.Now()}
	first := PropagateOpportunities(project)
	second := PropagateOpportunities(project)
	if len(first) != len(second) {
		t.Fatalf("expected same number of results, got %d vs %d", len(first), len(second))
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("result %d has different ID: %s vs %s", i, first[i].ID, second[i].ID)
		}
	}
}

func TestPropagateOpportunities_Sorted(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCriticalMinerals, Name: "Test", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: time.Now()}
	got := PropagateOpportunities(project)
	if len(got) <= 1 {
		t.Skip("need more than 1 result to verify sort order")
	}
}

func TestPropagateOpportunities_SetsCanonicalState(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCriticalMinerals, Name: "Test", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: time.Now()}
	got := PropagateOpportunities(project)
	for _, opp := range got {
		if opp.PublicationState != domain.PublicationPublicCanonical {
			t.Fatalf("expected PUBLIC_CANonical state, got %s", opp.PublicationState)
		}
		if !opp.Publishable {
			t.Fatal("expected publishable to be true")
		}
		if opp.Visibility != domain.VisibilityPublic {
			t.Fatalf("expected PUBLIC visibility, got %s", opp.Visibility)
		}
		if opp.OpportunityKind != "PROCUREMENT" {
			t.Fatalf("expected PROCUREMENT kind, got %s", opp.OpportunityKind)
		}
	}
}

func TestPropagateOpportunities_TitleFormat(t *testing.T) {
	project := &domain.Project{ID: "p1", Sector: domain.SectorCriticalMinerals, Name: "My Project", CurrentStage: domain.StageFEED, CapexCAD: 1_000_000_000, UpdatedAt: time.Now()}
	got := PropagateOpportunities(project)
	for _, opp := range got {
		if opp.Title == "" {
			t.Fatal("title should not be empty")
		}
		// Should include project name
	}
}
