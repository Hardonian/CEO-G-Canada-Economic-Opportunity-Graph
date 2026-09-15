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

// DefaultConfig returns a safe default configuration.
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

// FetchFunc fetches raw bytes for a province.
type FetchFunc func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error)

// Worker runs the gazette polling loop until ctx is cancelled.
type Worker struct {
	config     Config
	lastHashes map[string]string
	mu         sync.RWMutex
	fetcher    FetchFunc
}

// NewWorker constructs a gazette polling worker.
func NewWorker(config Config) *Worker {
	return &Worker{
		config:     config,
		lastHashes: make(map[string]string),
		fetcher:    defaultFetcher(config),
	}
}

// NewTestWorker constructs a polling worker that uses a custom fetch function.
func NewTestWorker(config Config, fetcher FetchFunc) *Worker {
	return &Worker{
		config:     config,
		lastHashes: make(map[string]string),
		fetcher:    fetcher,
	}
}

func defaultFetcher(config Config) FetchFunc {
	return func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		fixturePath := fmt.Sprintf(config.FixtureBase, strings.ToLower(province))
		adp := gazette.NewGazetteAdapter(province, fixturePath)
		raw, err := adp.Fetch(ctx)
		return raw, adp.Health(), err
	}
}

// PollOnce performs a single polling cycle across all configured provinces.
func (w *Worker) PollOnce(ctx context.Context) []PollResult {
	results := make([]PollResult, 0, len(w.config.Provinces))
	ch := make(chan PollResult, len(w.config.Provinces))
	sem := make(chan struct{}, w.config.MaxConcurrent)
	var wg sync.WaitGroup

	for _, province := range w.config.Provinces {
		wg.Add(1)
		go func(prov string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			ch <- w.pollProvince(ctx, prov)
		}(province)
	}
	wg.Wait()
	close(ch)
	for r := range ch {
		results = append(results, r)
	}
	return results
}

// Run starts the polling loop.
func (w *Worker) Run(ctx context.Context, onCycle func([]PollResult)) {
	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

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

	raw, health, err := w.fetcher(ctx, province)
	result.Health = health
	if err != nil {
		result.Err = fmt.Errorf("fetch %s gazette: %w", province, err)
		return result
	}

	currentHash := adapters.HashDocument(raw)
	result.CurrentHash = currentHash

	w.mu.RLock()
	prev := w.lastHashes[province]
	w.mu.RUnlock()
	result.PreviousHash = prev

	if prev != "" && prev == currentHash {
		result.Changed = false
		if health != nil {
			result.DocumentsSeen = health.DocumentsSeen
		}
		return result
	}

	fixturePath := fmt.Sprintf(w.config.FixtureBase, strings.ToLower(province))
	adp := gazette.NewGazetteAdapter(province, fixturePath)
	parsed, err := adp.Parse(raw)
	if err != nil {
		result.Err = fmt.Errorf("parse %s gazette: %w", province, err)
		return result
	}

	result.Changed = true
	if health != nil {
		result.DocumentsSeen = health.DocumentsSeen
	}
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

// Ensure domain import is used.
var _ = domain.StatusHealthy