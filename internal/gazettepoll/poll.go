// Package gazettepoll provides a scheduled polling worker for provincial
// gazettes. It wraps the gazette adapters in a change-detection loop that
// re-fetches each gazette on a configurable interval, compares the document
// hash against the previous cycle, and only re-parses when content has
// actually changed. This keeps the ingestion pipeline cheap while still
// picking up new notices as they are published.
package gazettepoll

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/gazette"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// Config controls the polling worker behaviour.
type Config struct {
	Interval      time.Duration
	Provinces     []string
	FixtureBase   string
	MaxConcurrent int
}

// DefaultConfig returns a safe default configuration: poll every 5 minutes,
// watch all four jurisdictions, and process at most 2 adapters concurrently.
func DefaultConfig() Config {
	return Config{
		Interval:      5 * time.Minute,
		Provinces:     []string{"ON", "QC", "BC", "CA"},
		FixtureBase:   "data/fixtures/gazette_%s.json",
		MaxConcurrent: 2,
	}
}

// PollResult summarises a single polling cycle.
type PollResult struct {
	Province         string
	Changed          bool
	PreviousHash     string
	CurrentHash      string
	DocumentsSeen    int
	DocumentsChanged int
	Health           *adapters.SourceHealth
	Err              error
}

// Worker runs the gazette polling loop until ctx is cancelled.
type Worker struct {
	config    Config
	lastHashes map[string]string
	mu         sync.RWMutex
}

// NewWorker constructs a gazette polling worker.
func NewWorker(config Config) *Worker {
	return &Worker{
		config:     config,
		lastHashes: make(map[string]string),
	}
}

// PollOnce performs a single polling cycle across all configured provinces.
// It returns one PollResult per province. Adapters whose document hash is
// unchanged since the last cycle report Changed=false and skip re-parsing.
func (w *Worker) PollOnce(ctx context.Context) []PollResult {
	results := make([]PollResult, 0, len(w.config.Provinces))

	// Bound concurrency with a semaphore.
	sem := make(chan struct{}, w.config.MaxConcurrent)
	var wg sync.WaitGroup

	for _, province := range w.config.Provinces {
		wg.Add(1)
		go func(prov string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results = append(results, w.pollProvince(ctx, prov))
		}(province)
	}
	wg.Wait()
	return results
}

// Run starts the polling loop. It calls onCycle for every cycle, including
// the initial one, so callers can react to changes in real time.
func (w *Worker) Run(ctx context.Context, onCycle func([]PollResult)) {
	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	// Initial cycle — run immediately so the worker is useful on startup.
	onCycle(w.PollOnce(ctx))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			onCycle(w.PollOnce(ctx))
		}
	}
}

func (w *Worker) pollProvince(ctx context.Context, province string) PollResult {
	result := PollResult{Province: province}
	fixturePath := fmt.Sprintf(w.config.FixtureBase, strings.ToLower(province))
	adp := gazette.NewGazetteAdapter(province, fixturePath)

	raw, err := adp.Fetch(ctx)
	if err != nil {
		result.Err = fmt.Errorf("fetch %s gazette: %w", province, err)
		result.Health = adp.Health()
		return result
	}

	currentHash := adapters.HashDocument(raw)
	result.CurrentHash = currentHash
	result.Health = adp.Health()

	w.mu.RLock()
	prev := w.lastHashes[province]
	w.mu.RUnlock()
	result.PreviousHash = prev

	if prev != "" && prev == currentHash {
		result.Changed = false
		result.DocumentsSeen = adp.Health().DocumentsSeen
		return result
	}

	parsed, err := adp.Parse(raw)
	if err != nil {
		result.Err = fmt.Errorf("parse %s gazette: %w", province, err)
		return result
	}

	result.Changed = true
	result.DocumentsSeen = adp.Health().DocumentsSeen
	result.DocumentsChanged = len(parsed.Projects)

	w.mu.Lock()
	w.lastHashes[province] = currentHash
	w.mu.Unlock()

	log.Printf("[GAZETTE-POLL] %s changed: %d new/updated notices\n", province, result.DocumentsChanged)
	return result
}

// Health returns the aggregated health of all configured provinces.
func (w *Worker) Health() []*adapters.SourceHealth {
	w.mu.RLock()
	defer w.mu.RUnlock()
	healths := make([]*adapters.SourceHealth, 0, len(w.config.Provinces))
	for _, province := range w.config.Provinces {
		fixturePath := fmt.Sprintf(w.config.FixtureBase, strings.ToLower(province))
		adp := gazette.NewGazetteAdapter(province, fixturePath)
		healths = append(healths, adp.Health())
	}
	return healths
}

// LastHash returns the last seen document hash for a province.
func (w *Worker) LastHash(province string) string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.lastHashes[province]
}

// Ensure domain import is used (health status constants are referenced by callers).
var _ = domain.StatusHealthy