package forecast

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestEvaluateIsDeterministicOrderIndependentAndDoesNotMutateInputs(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 15, 30, 0, 0, time.FixedZone("EST", -5*60*60))
	ctx := completeContext(asOf)
	request := Request{AsOf: asOf, HorizonsMonths: []int{60, 12, 36, 12}, Scenarios: []Scenario{StandardScenarios()[1], StandardScenarios()[0], StandardScenarios()[2]}}
	originalEventOrder := []string{ctx.Events[0].ID, ctx.Events[1].ID}
	originalHorizons := append([]int(nil), request.HorizonsMonths...)

	first, err := Evaluate(ctx, request)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	secondContext := ctx
	secondContext.Events = reverseCopy(ctx.Events)
	secondContext.CapitalItems = reverseCopy(ctx.CapitalItems)
	secondContext.Relationships = reverseCopy(ctx.Relationships)
	secondContext.Procurements = reverseCopy(ctx.Procurements)
	secondContext.Opportunities = reverseCopy(ctx.Opportunities)
	secondContext.Signals = reverseCopy(ctx.Signals)
	second, err := Evaluate(secondContext, request)
	if err != nil {
		t.Fatalf("Evaluate(reordered) error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("deterministic reports differ\nfirst: %#v\nsecond: %#v", first, second)
	}
	if first.CalculatedAt != asOf.UTC() || first.AsOf != asOf.UTC() {
		t.Fatalf("forecast used a wall clock instead of as-of: calculated=%v as_of=%v", first.CalculatedAt, first.AsOf)
	}
	if got := []string{ctx.Events[0].ID, ctx.Events[1].ID}; !slices.Equal(got, originalEventOrder) {
		t.Fatalf("Evaluate mutated event order: got %v want %v", got, originalEventOrder)
	}
	if !slices.Equal(request.HorizonsMonths, originalHorizons) {
		t.Fatalf("Evaluate mutated request horizons: got %v want %v", request.HorizonsMonths, originalHorizons)
	}
	if first.Assurance.ExternalNetworkRequired || first.Assurance.ExternalTelemetry || first.Assurance.RawEvidenceEmitted || !first.Assurance.Deterministic {
		t.Fatalf("unexpected assurance properties: %+v", first.Assurance)
	}
	if first.Confidence != ConfidenceHigh || first.DataQuality.Score != 100 {
		t.Fatalf("complete fixture should have high support, got confidence=%s quality=%v", first.Confidence, first.DataQuality.Score)
	}
	if len(first.Provenance) != 1 || first.Provenance[0].EvidenceID != "evidence-1" {
		t.Fatalf("unexpected provenance: %+v", first.Provenance)
	}
	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("json.Marshal(report) error = %v", err)
	}
	if strings.Contains(string(encoded), "SENSITIVE RAW SOURCE TEXT") || strings.Contains(string(encoded), "raw_snippet") {
		t.Fatalf("raw evidence escaped the minimized provenance boundary: %s", encoded)
	}
	for _, scenario := range first.Scenarios {
		for _, milestone := range scenario.Milestones {
			for _, horizon := range milestone.ByHorizon {
				r := horizon.CompletionLikelihoodIndexPct
				if r.Low > r.Base || r.Base > r.High || r.Low < 0 || r.High > 100 {
					t.Fatalf("invalid uncertainty range %+v", r)
				}
			}
		}
	}
}

func TestEvaluateScenariosMoveScheduleLikelihoodAndCapitalInExpectedDirection(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	report, err := Evaluate(completeContext(asOf), Request{AsOf: asOf, HorizonsMonths: []int{36}, Scenarios: StandardScenarios()})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	baseline := scenarioByID(t, report, "baseline")
	stress := scenarioByID(t, report, "delivery_stress")
	accelerated := scenarioByID(t, report, "accelerated_delivery")
	baseConstruction := milestoneByStage(t, baseline, domain.StageConstruction)
	stressConstruction := milestoneByStage(t, stress, domain.StageConstruction)
	acceleratedConstruction := milestoneByStage(t, accelerated, domain.StageConstruction)

	if !stressConstruction.ConditionalSchedule.Base.After(*baseConstruction.ConditionalSchedule.Base) {
		t.Fatalf("delivery stress should delay construction: stress=%v baseline=%v", stressConstruction.ConditionalSchedule.Base, baseConstruction.ConditionalSchedule.Base)
	}
	if !acceleratedConstruction.ConditionalSchedule.Base.Before(*baseConstruction.ConditionalSchedule.Base) {
		t.Fatalf("accelerated scenario should advance construction: accelerated=%v baseline=%v", acceleratedConstruction.ConditionalSchedule.Base, baseConstruction.ConditionalSchedule.Base)
	}
	if stressConstruction.ByHorizon[0].CompletionLikelihoodIndexPct.Base >= baseConstruction.ByHorizon[0].CompletionLikelihoodIndexPct.Base {
		t.Fatalf("delivery stress should lower 36-month construction index: stress=%+v baseline=%+v", stressConstruction.ByHorizon[0], baseConstruction.ByHorizon[0])
	}
	if stress.Capital.ScenarioAdjustedCapexCAD.Base != 1_200_000_000 || baseline.Capital.ScenarioAdjustedCapexCAD.Base != 1_000_000_000 {
		t.Fatalf("capex stress not applied transparently: stress=%+v baseline=%+v", stress.Capital, baseline.Capital)
	}
	if baseline.Capital.EvidencedCommittedCAD != 300_000_000 || baseline.Capital.IndicativeFundingGapCAD.Base != 700_000_000 {
		t.Fatalf("unexpected evidenced capital outlook: %+v", baseline.Capital)
	}
	if baseline.Capital.ConfirmedProcurementCAD != 50_000_000 || baseline.Capital.ConfirmedOpportunityCAD != 20_000_000 || baseline.Capital.DerivedOpportunityCAD != 100_000_000 {
		t.Fatalf("planning pipeline totals are wrong: %+v", baseline.Capital)
	}
	if assumptionByName(t, stress, "capex_escalation").Source != AssumptionScenario {
		t.Fatal("scenario capex stress was not labelled as a caller assumption")
	}
}

func TestEvaluateIgnoresInputsNotKnowableAtAsOf(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	ctx := completeContext(asOf)
	futureEvidence := &domain.Evidence{ID: "future-evidence", Publisher: "Future publisher", SourceTier: domain.SourceTier1,
		RetrievalTimestamp: asOf.Add(24 * time.Hour), Confidence: domain.ConfidenceVerified, ContentHash: "future"}
	ctx.Evidence = append(ctx.Evidence, futureEvidence)
	ctx.CapitalItems = append(ctx.CapitalItems, &domain.CapitalItem{ID: "future-capital", ProjectID: ctx.Project.ID, Status: domain.CapitalCommitted,
		Category: domain.CapitalCategoryDebt, AmountCAD: 600_000_000, AmountType: "exact", EvidenceID: futureEvidence.ID, CreatedAt: asOf.Add(time.Hour)})
	ctx.Signals = append(ctx.Signals, &domain.Signal{ID: "future-signal", ProjectID: ctx.Project.ID, Type: domain.SignalFinancingAcceleration,
		Timestamp: asOf.Add(time.Hour), Magnitude: 1, Confidence: 1, EvidenceID: futureEvidence.ID})

	report, err := Evaluate(ctx, Request{AsOf: asOf})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	baseline := scenarioByID(t, report, "baseline")
	if baseline.Capital.EvidencedCommittedCAD != 300_000_000 {
		t.Fatalf("future capital leaked into forecast: %+v", baseline.Capital)
	}
	for _, warning := range []string{"future_capital_item_ignored", "future_signal_ignored", "future_retrieved_evidence_ignored"} {
		if !slices.Contains(report.Warnings, warning) {
			t.Fatalf("missing temporal warning %q in %v", warning, report.Warnings)
		}
	}
	if len(report.Provenance) != 1 || report.Provenance[0].EvidenceID == futureEvidence.ID {
		t.Fatalf("future evidence leaked into provenance: %+v", report.Provenance)
	}
}

func TestEvaluateRejectsTemporalLeakageCrossProjectAndDuplicateRecords(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	t.Run("newer project snapshot", func(t *testing.T) {
		ctx := completeContext(asOf)
		ctx.Project = cloneProject(ctx.Project)
		ctx.Project.UpdatedAt = asOf.Add(time.Second)
		_, err := Evaluate(ctx, Request{AsOf: asOf})
		if !errors.Is(err, ErrTemporalLeakage) {
			t.Fatalf("error = %v, want ErrTemporalLeakage", err)
		}
	})
	t.Run("cross-project record", func(t *testing.T) {
		ctx := completeContext(asOf)
		ctx.Events = append([]*domain.Event(nil), ctx.Events...)
		copyEvent := *ctx.Events[0]
		copyEvent.ProjectID = "different-project"
		ctx.Events[0] = &copyEvent
		_, err := Evaluate(ctx, Request{AsOf: asOf})
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "different-project") {
			t.Fatalf("error = %v, want cross-project ErrInvalidInput", err)
		}
	})
	t.Run("duplicate record", func(t *testing.T) {
		ctx := completeContext(asOf)
		ctx.CapitalItems = append(ctx.CapitalItems, ctx.CapitalItems[0])
		_, err := Evaluate(ctx, Request{AsOf: asOf})
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "duplicate capital item") {
			t.Fatalf("error = %v, want duplicate ErrInvalidInput", err)
		}
	})
}

func TestEvaluateRejectsUnboundedOrNonFiniteScenarioInputs(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	tests := []Scenario{
		{ID: "bad space", Name: "Bad ID"},
		{ID: "nan", Name: "NaN", DemandDelta: math.NaN()},
		{ID: "delay", Name: "Delay", RegulatoryDelayMonths: 121},
		{ID: "stress", Name: "Stress", SupplyChainStress: 1.01},
	}
	for _, scenario := range tests {
		t.Run(scenario.ID, func(t *testing.T) {
			_, err := Evaluate(completeContext(asOf), Request{AsOf: asOf, Scenarios: []Scenario{scenario}})
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestEvaluateRejectsInvalidStageAndUnsafeProvenanceURL(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	t.Run("invalid lifecycle stage", func(t *testing.T) {
		ctx := completeContext(asOf)
		ctx.Project = cloneProject(ctx.Project)
		ctx.Project.CurrentStage = domain.LifecycleStage("MADE_UP_STAGE")
		_, err := Evaluate(ctx, Request{AsOf: asOf})
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "unsupported lifecycle stage") {
			t.Fatalf("error = %v, want invalid lifecycle error", err)
		}
	})
	t.Run("unsafe provenance URL", func(t *testing.T) {
		ctx := completeContext(asOf)
		copyEvidence := *ctx.Evidence[0]
		copyEvidence.SourceURL = "javascript:alert(1)"
		ctx.Evidence = []*domain.Evidence{&copyEvidence}
		_, err := Evaluate(ctx, Request{AsOf: asOf})
		if !errors.Is(err, ErrInvalidInput) || !strings.Contains(err.Error(), "http(s)") {
			t.Fatalf("error = %v, want safe URL validation error", err)
		}
	})
}

func TestEvaluateMakesInsufficientDataAndCancelledStateExplicit(t *testing.T) {
	asOf := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	minimal := Context{Project: &domain.Project{ID: "minimal", Name: "Minimal", CurrentStage: domain.StageUnknown, CapexStatus: domain.ConfidenceUnknown}}
	report, err := Evaluate(minimal, Request{AsOf: asOf})
	if err != nil {
		t.Fatalf("Evaluate(minimal) error = %v", err)
	}
	if report.Confidence != ConfidenceInsufficient || report.Status != domain.StatusDegraded {
		t.Fatalf("minimal data was overstated: confidence=%s status=%s quality=%+v", report.Confidence, report.Status, report.DataQuality)
	}
	for _, gap := range []string{"capex", "current_stage", "referenced_evidence"} {
		if !slices.Contains(report.DataQuality.CriticalGaps, gap) {
			t.Fatalf("missing explicit gap %q in %v", gap, report.DataQuality.CriticalGaps)
		}
	}
	if len(report.Limitations) < 4 {
		t.Fatalf("material limitations are missing: %v", report.Limitations)
	}

	cancelledContext := completeContext(asOf)
	cancelledContext.Project = cloneProject(cancelledContext.Project)
	cancelledContext.Project.CurrentStage = domain.StageCancelled
	cancelled, err := Evaluate(cancelledContext, Request{AsOf: asOf})
	if err != nil {
		t.Fatalf("Evaluate(cancelled) error = %v", err)
	}
	for _, milestone := range cancelled.Scenarios[0].Milestones {
		if milestone.ConditionalSchedule.Base != nil || milestone.ByHorizon[0].CompletionLikelihoodIndexPct != (LikelihoodRange{}) {
			t.Fatalf("cancelled project asserted a future milestone: %+v", milestone)
		}
	}
}

func TestAddCalendarMonthsClampsMonthEnd(t *testing.T) {
	start := time.Date(2025, time.January, 31, 12, 0, 0, 0, time.UTC)
	got := addCalendarMonths(start, 1)
	want := time.Date(2025, time.February, 28, 12, 0, 0, 0, time.UTC)
	if got != want {
		t.Fatalf("addCalendarMonths() = %v, want %v", got, want)
	}
}

func completeContext(asOf time.Time) Context {
	asOf = asOf.UTC()
	previous := domain.StageFeasibility
	permitting := domain.StagePermitting
	evidence := &domain.Evidence{
		ID: "evidence-1", SourceURL: "https://example.gc.ca/project/1", Publisher: "Government of Canada", SourceTier: domain.SourceTier1,
		RetrievalTimestamp: asOf.Add(-5 * 24 * time.Hour), Confidence: domain.ConfidenceVerified, ContentHash: "sha256:abc",
		RawSnippet: "SENSITIVE RAW SOURCE TEXT",
	}
	return Context{
		Project: &domain.Project{
			ID: "project-1", Name: "National Project", Sector: domain.SectorCriticalMinerals, Province: "ON", LocationName: "Northern Ontario",
			CurrentStage: domain.StagePermitting, CapexCAD: 1_000_000_000, CapexStatus: domain.ConfidenceVerified, ProponentID: "entity-1",
			Confidence: domain.ConfidenceVerified, EvidenceIDs: []string{evidence.ID}, LastMeaningfulUpdate: asOf.Add(-10 * 24 * time.Hour), UpdatedAt: asOf.Add(-10 * 24 * time.Hour),
		},
		Events: []*domain.Event{
			{ID: "event-regulatory", ProjectID: "project-1", EventType: "regulatory_filing", EventDate: asOf.Add(-30 * 24 * time.Hour), PreviousStage: &previous, NewStage: &permitting, EvidenceID: evidence.ID, CreatedAt: asOf.Add(-29 * 24 * time.Hour)},
			{ID: "event-indigenous", ProjectID: "project-1", EventType: "indigenous_agreement", EventDate: asOf.Add(-60 * 24 * time.Hour), EvidenceID: evidence.ID, CreatedAt: asOf.Add(-59 * 24 * time.Hour)},
		},
		CapitalItems: []*domain.CapitalItem{
			{ID: "capital-committed", ProjectID: "project-1", Category: domain.CapitalCategoryDebt, Status: domain.CapitalCommitted, AmountCAD: 300_000_000, AmountType: "exact", EvidenceID: evidence.ID, CreatedAt: asOf.Add(-20 * 24 * time.Hour)},
			{ID: "capital-conditional", ProjectID: "project-1", Category: domain.CapitalCategoryLoanGuarantee, Status: domain.CapitalConditionallyCommitted, AmountCAD: 100_000_000, AmountType: "exact", EvidenceID: evidence.ID, CreatedAt: asOf.Add(-15 * 24 * time.Hour)},
		},
		Relationships: []*domain.Relationship{
			{ID: "relationship-offtake", ProjectID: "project-1", RelationType: "offtaker", Confidence: domain.ConfidenceVerified, EvidenceID: evidence.ID, CreatedAt: asOf.Add(-40 * 24 * time.Hour)},
		},
		Procurements: []*domain.Procurement{
			{ID: "procurement-1", ProjectID: "project-1", Stage: "RFP", EstimatedCAD: 50_000_000, RequirementClass: domain.RequirementConfirmed, EvidenceID: evidence.ID, CreatedAt: asOf.Add(-7 * 24 * time.Hour), ClosingDate: timePointer(asOf.Add(30 * 24 * time.Hour))},
		},
		Opportunities: []*domain.Opportunity{
			{ID: "opportunity-confirmed", ProjectID: "project-1", RequirementClass: domain.RequirementConfirmed, EstimatedCAD: 20_000_000, EstimateStatus: domain.ConfidenceReported, CreatedAt: asOf.Add(-7 * 24 * time.Hour)},
			{ID: "opportunity-derived", ProjectID: "project-1", RequirementClass: domain.RequirementDerived, EstimatedCAD: 100_000_000, EstimateStatus: domain.ConfidenceInferred, CreatedAt: asOf.Add(-7 * 24 * time.Hour)},
		},
		Signals: []*domain.Signal{
			{ID: "signal-1", ProjectID: "project-1", Type: domain.SignalFinancingAcceleration, Timestamp: asOf.Add(-20 * 24 * time.Hour), Magnitude: 0.8, Confidence: 0.9, EvidenceID: evidence.ID},
		},
		Evidence: []*domain.Evidence{evidence},
	}
}

func scenarioByID(t *testing.T, report *Report, id string) ScenarioForecast {
	t.Helper()
	for _, scenario := range report.Scenarios {
		if scenario.Scenario.ID == id {
			return scenario
		}
	}
	t.Fatalf("scenario %q not found", id)
	return ScenarioForecast{}
}

func milestoneByStage(t *testing.T, scenario ScenarioForecast, stage domain.LifecycleStage) MilestoneForecast {
	t.Helper()
	for _, milestone := range scenario.Milestones {
		if milestone.Stage == stage {
			return milestone
		}
	}
	t.Fatalf("milestone %q not found", stage)
	return MilestoneForecast{}
}

func assumptionByName(t *testing.T, scenario ScenarioForecast, name string) Assumption {
	t.Helper()
	for _, assumption := range scenario.Assumptions {
		if assumption.Parameter == name {
			return assumption
		}
	}
	t.Fatalf("assumption %q not found", name)
	return Assumption{}
}

func reverseCopy[T any](values []T) []T {
	result := append([]T(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func cloneProject(value *domain.Project) *domain.Project {
	result := *value
	result.EvidenceIDs = append([]string(nil), value.EvidenceIDs...)
	return &result
}

func timePointer(value time.Time) *time.Time { return &value }
