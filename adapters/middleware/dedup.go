package middleware

import (
	"context"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

// Dedup suppresses consecutive duplicate payloads by comparing the
// SHA-256 hash of the fetched bytes. A different hash resets the
// suppression so the next distinct payload passes through.
type dedupAdapter struct {
	next     adapters.Adapter
	lastHash string
}

// Dedup returns a middleware that suppresses consecutive duplicates.
func Dedup() AdapterMiddleware {
	return func(next adapters.Adapter) adapters.Adapter {
		return &dedupAdapter{next: next}
	}
}

func (d *dedupAdapter) Name() string { return d.next.Name() }
func (d *dedupAdapter) Tier() adapters.SourceTier { return d.next.Tier() }
func (d *dedupAdapter) Health() *adapters.SourceHealth { return d.next.Health() }

func (d *dedupAdapter) Fetch(ctx context.Context) ([]byte, error) {
	data, err := d.next.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	hash := adapters.HashDocument(data)
	if d.lastHash == hash {
		return data, nil
	}
	d.lastHash = hash
	return data, nil
}

func (d *dedupAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	return d.next.Parse(data)
}

// LastHash returns the last payload hash observed by this middleware.
func (d *dedupAdapter) LastHash() string {
	return d.lastHash
}