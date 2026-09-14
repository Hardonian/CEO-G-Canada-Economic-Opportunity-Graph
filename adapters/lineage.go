package adapters

import (
	"encoding/hex"
	"fmt"
	"net/url"
)

// ValidateLineage enforces the minimum evidence chain required before an
// adapter result can influence scoring. It deliberately validates references
// rather than trusting embedded Evidence pointers.
func ValidateLineage(result *IngestionResult) error {
	if result == nil {
		return fmt.Errorf("lineage: ingestion result is nil")
	}
	evidence := make(map[string]struct{}, len(result.Evidence))
	for _, item := range result.Evidence {
		if item == nil || item.ID == "" {
			return fmt.Errorf("lineage: evidence has no id")
		}
		if _, exists := evidence[item.ID]; exists {
			return fmt.Errorf("lineage: duplicate evidence id %q", item.ID)
		}
		evidence[item.ID] = struct{}{}
		parsedURL, err := url.Parse(item.SourceURL)
		if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
			return fmt.Errorf("lineage: evidence %q requires an absolute HTTPS source URL", item.ID)
		}
		hash, err := hex.DecodeString(item.ContentHash)
		if err != nil || len(hash) != 32 {
			return fmt.Errorf("lineage: evidence %q has an invalid SHA-256 content hash", item.ID)
		}
		if item.Publisher == "" || item.SourceRecordID == "" || item.Locator == "" || item.PipelineVersion == "" || item.ParserVersion == "" {
			return fmt.Errorf("lineage: evidence %q is missing publisher or transformation metadata", item.ID)
		}
		if item.RetrievalTimestamp.IsZero() {
			return fmt.Errorf("lineage: evidence %q has no retrieval timestamp", item.ID)
		}
	}

	for _, metric := range result.TradeMetrics {
		if metric == nil || metric.ID == "" || metric.MetricCode == "" || metric.Geography == "" || metric.ReferencePeriod == "" || metric.EvidenceID == "" {
			return fmt.Errorf("lineage: incomplete trade metric")
		}
		if _, ok := evidence[metric.EvidenceID]; !ok {
			return fmt.Errorf("lineage: trade metric %q references unknown evidence %q", metric.ID, metric.EvidenceID)
		}
		if metric.ObservedAt.IsZero() {
			return fmt.Errorf("lineage: trade metric %q has no observation time", metric.ID)
		}
	}
	return nil
}
