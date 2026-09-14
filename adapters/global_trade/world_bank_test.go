package global_trade

import (
	"context"
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

func TestPinnedWorldBankSnapshotHasCompleteLineage(t *testing.T) {
	adapter := NewAdapter("../../data/fixtures/world_bank_trade_canada.json")
	raw, err := adapter.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.TradeMetrics) != len(indicatorCodes) || len(result.Evidence) != len(indicatorCodes) {
		t.Fatalf("metrics=%d evidence=%d, want %d each", len(result.TradeMetrics), len(result.Evidence), len(indicatorCodes))
	}
	if err := adapters.ValidateLineage(result); err != nil {
		t.Fatalf("lineage validation failed: %v", err)
	}
	evidenceByID := make(map[string]bool, len(result.Evidence))
	for _, evidence := range result.Evidence {
		evidenceByID[evidence.ID] = true
		if evidence.SourceID != "world-bank-indicators-api" || evidence.MappingVersion != mappingVersion {
			t.Fatalf("incomplete transformation lineage: %#v", evidence)
		}
	}
	for _, metric := range result.TradeMetrics {
		if !evidenceByID[metric.EvidenceID] || metric.Evidence == nil || metric.Evidence.ID != metric.EvidenceID {
			t.Fatalf("metric %s does not resolve to its evidence", metric.MetricCode)
		}
	}
}

func TestLineageRejectsDetachedTradeMetric(t *testing.T) {
	adapter := NewAdapter("../../data/fixtures/world_bank_trade_canada.json")
	raw, err := adapter.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	result.TradeMetrics[0].EvidenceID = "missing-evidence"
	if err := adapters.ValidateLineage(result); err == nil {
		t.Fatal("expected detached metric lineage to be rejected")
	}
}
