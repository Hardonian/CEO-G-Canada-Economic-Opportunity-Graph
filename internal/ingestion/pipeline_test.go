package ingestion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

type checkpointTestAdapter struct {
	payload    []byte
	parseErr   error
	parseCalls int
	health     adapters.SourceHealth
}

type projectTestAdapter struct {
	name    string
	project *domain.Project
	health  adapters.SourceHealth
}

func (a *projectTestAdapter) Name() string            { return a.name }
func (a *projectTestAdapter) Tier() domain.SourceTier { return domain.SourceTier1 }
func (a *projectTestAdapter) Health() *adapters.SourceHealth {
	copy := a.health
	return &copy
}
func (a *projectTestAdapter) Fetch(context.Context) ([]byte, error) { return []byte(a.name), nil }
func (a *projectTestAdapter) Parse([]byte) (*adapters.IngestionResult, error) {
	copy := *a.project
	return &adapters.IngestionResult{Projects: []*domain.Project{&copy}}, nil
}

func newCheckpointTestAdapter(parseErr error) *checkpointTestAdapter {
	return &checkpointTestAdapter{
		payload:  []byte(`{"records":[{"id":"one"}]}`),
		parseErr: parseErr,
		health: adapters.SourceHealth{
			AdapterName: "checkpoint-test",
			Tier:        domain.SourceTier1,
			Status:      "UNKNOWN",
			Mode:        "CURATED_SNAPSHOT",
		},
	}
}

func (a *checkpointTestAdapter) Name() string            { return a.health.AdapterName }
func (a *checkpointTestAdapter) Tier() domain.SourceTier { return domain.SourceTier1 }
func (a *checkpointTestAdapter) Health() *adapters.SourceHealth {
	snapshot := a.health
	return &snapshot
}
func (a *checkpointTestAdapter) Fetch(context.Context) ([]byte, error) {
	a.health.LastAttempt = time.Now().UTC()
	return append([]byte(nil), a.payload...), nil
}
func (a *checkpointTestAdapter) Parse([]byte) (*adapters.IngestionResult, error) {
	a.parseCalls++
	if a.parseErr != nil {
		return nil, a.parseErr
	}
	return &adapters.IngestionResult{}, nil
}

func TestParseFailureDoesNotAdvanceLastKnownGoodHash(t *testing.T) {
	adapter := newCheckpointTestAdapter(errors.New("schema drift"))
	pipeline := NewPipeline(database.NewMemoryStore(), []adapters.Adapter{adapter})

	for run := 1; run <= 2; run++ {
		if _, err := pipeline.Run(context.Background()); err == nil {
			t.Fatalf("run %d: expected parse failure", run)
		}
	}

	if adapter.parseCalls != 2 {
		t.Fatalf("parse calls = %d, want 2; failed payload must remain retryable", adapter.parseCalls)
	}
}

func TestSuccessfulPayloadAdvancesLastKnownGoodHash(t *testing.T) {
	adapter := newCheckpointTestAdapter(nil)
	pipeline := NewPipeline(database.NewMemoryStore(), []adapters.Adapter{adapter})

	for run := 1; run <= 2; run++ {
		if _, err := pipeline.Run(context.Background()); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}

	if adapter.parseCalls != 1 {
		t.Fatalf("parse calls = %d, want 1 for an unchanged last-known-good payload", adapter.parseCalls)
	}
}

func TestCanonicalProjectIsScoredOnceAfterMultiAdapterMerge(t *testing.T) {
	store := database.NewMemoryStore()
	first := &projectTestAdapter{name: "first", project: &domain.Project{
		ID: "project-a", Slug: "shared-project", Name: "Shared Project", Sector: domain.SectorTransportation,
		Province: "BC", CurrentStage: domain.StageProcurement, CapexCAD: 1_000_000_000,
		CapexStatus: domain.ConfidenceReported, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	second := &projectTestAdapter{name: "second", project: &domain.Project{
		ID: "project-b", Slug: "shared-project", Name: "Shared Project", Sector: domain.SectorTransportation,
		Province: "BC", CurrentStage: domain.StageConstruction, CapexCAD: 1_000_000_000,
		CapexStatus: domain.ConfidenceReported, UpdatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}}
	pipeline := NewPipeline(store, []adapters.Adapter{first, second})
	if _, err := pipeline.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	history, err := store.ListScoreHistory(context.Background(), "project-a", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 5 {
		t.Fatalf("score history has %d records, want exactly one record for each of five dimensions", len(history))
	}
}
