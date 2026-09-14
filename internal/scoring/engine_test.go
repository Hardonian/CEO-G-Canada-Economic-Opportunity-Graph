package scoring

import (
	"context"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/global_trade"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestTradeResilienceScoreIsDeterministicAndEvidenceComplete(t *testing.T) {
	adapter := global_trade.NewAdapter("../../data/fixtures/world_bank_trade_canada.json")
	raw, err := adapter.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := adapter.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	project := &domain.Project{ID: "project-1", Sector: domain.SectorTransportation, CurrentStage: domain.StageProcurement, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	ctx := &ProjectContext{Project: project, TradeMetrics: parsed.TradeMetrics}
	first := CalculateTradeResilience(ctx)
	second := CalculateTradeResilience(ctx)
	if first.ID != second.ID || first.InputHash != second.InputHash || first.ScoreValue != second.ScoreValue || !first.CalculatedAt.Equal(second.CalculatedAt) {
		t.Fatalf("score is not deterministic: %#v %#v", first, second)
	}
	if first.Coverage != 100 || len(first.UnknownFactors) != 0 || len(first.EvidenceIDs) != 5 {
		t.Fatalf("unexpected coverage or evidence: %#v", first)
	}
	available := make(map[string]bool, len(parsed.Evidence))
	for _, evidence := range parsed.Evidence {
		available[evidence.ID] = true
	}
	for _, evidenceID := range first.EvidenceIDs {
		if !available[evidenceID] {
			t.Fatalf("score evidence %q is not resolvable", evidenceID)
		}
	}
	copyMetric := *parsed.TradeMetrics[7]
	copyMetric.Value++
	mutated := append([]*domain.TradeMetric(nil), parsed.TradeMetrics...)
	mutated[7] = &copyMetric
	changed := CalculateTradeResilience(&ProjectContext{Project: project, TradeMetrics: mutated})
	if changed.InputHash == first.InputHash || changed.ScoreValue == first.ScoreValue {
		t.Fatal("changing a scored observation must change both the input hash and result")
	}
}

func TestSupplierabilityCarriesWorldBankFactorLineage(t *testing.T) {
	adapter := global_trade.NewAdapter("../../data/fixtures/world_bank_trade_canada.json")
	raw, _ := adapter.Fetch(context.Background())
	parsed, err := adapter.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	project := &domain.Project{ID: "project-1", Sector: domain.SectorTransportation, CurrentStage: domain.StageProcurement, EvidenceIDs: []string{"project-evidence"}, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	score := CalculateSupplierability(&ProjectContext{Project: project, TradeMetrics: parsed.TradeMetrics})
	ids := score.FactorEvidence["trade_logistics"]
	if len(ids) != 1 || ids[0] == "" || score.Factors["trade_logistics"] != 75 {
		t.Fatalf("trade logistics factor has incomplete lineage: %#v", score)
	}
}
