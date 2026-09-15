package gazettepoll

import (
	"context"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

func TestNewWorker_DefaultConfig(t *testing.T) {
	w := NewWorker(DefaultConfig())
	if w == nil {
		t.Fatal("expected non-nil worker")
	}
	if w.config.Interval != 5*time.Minute {
		t.Fatalf("expected 5m interval, got %v", w.config.Interval)
	}
	if len(w.config.Provinces) != 4 {
		t.Fatalf("expected 4 provinces, got %d", len(w.config.Provinces))
	}
}

func TestPollOnce_FirstRun(t *testing.T) {
	w := NewWorker(DefaultConfig())
	results := w.PollOnce(context.Background())
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error polling %s: %v", r.Province, r.Err)
		}
		if !r.Changed {
			t.Fatalf("expected first run to report changed for %s", r.Province)
		}
		if r.CurrentHash == "" {
			t.Fatalf("expected non-empty hash for %s", r.Province)
		}
		// Health should be non-nil.
		if r.Health == nil {
			t.Fatalf("expected non-nil health for %s", r.Province)
		}
	}
}

func TestPollOnce_IdempotentSecondRun(t *testing.T) {
	w := NewWorker(DefaultConfig())
	// First run populates lastHashes.
	w.PollOnce(context.Background())
	// Second run should report no changes.
	results := w.PollOnce(context.Background())
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	changedCount := 0
	for _, r := range results {
		if r.Changed {
			changedCount++
		}
		if r.Err != nil {
			t.Fatalf("unexpected error polling %s: %v", r.Province, r.Err)
		}
	}
	if changedCount != 0 {
		t.Fatalf("expected 0 changes on second run, got %d", changedCount)
	}
}

func TestPollOnce_ChangeDetection(t *testing.T) {
	w := NewWorker(DefaultConfig())
	// First run.
	first := w.PollOnce(context.Background())
	if len(first) == 0 {
		t.Fatal("expected results from first run")
	}
	// Second run — no changes expected.
	second := w.PollOnce(context.Background())
	for i := range second {
		if second[i].PreviousHash != first[i].CurrentHash {
			t.Fatalf("expected previous hash to match first run's current hash for %s", second[i].Province)
		}
	}
}

func TestWorker_Health(t *testing.T) {
	w := NewWorker(DefaultConfig())
	healths := w.Health()
	if len(healths) != 4 {
		t.Fatalf("expected 4 health entries, got %d", len(healths))
	}
	for _, h := range healths {
		if h == nil {
			t.Fatal("expected non-nil health entry")
		}
		if h.AdapterName == "" {
			t.Fatal("expected non-empty adapter name")
		}
	}
}

func TestWorker_LastHash(t *testing.T) {
	w := NewWorker(DefaultConfig())
	// Before any polling, LastHash should be empty.
	if h := w.LastHash("ON"); h != "" {
		t.Fatalf("expected empty hash before polling, got %s", h)
	}
	// After polling, LastHash should be populated.
	w.PollOnce(context.Background())
	if h := w.LastHash("ON"); h == "" {
		t.Fatal("expected non-empty hash after polling")
	}
}

func TestWorker_RunCancel(t *testing.T) {
	w := NewWorker(Config{
		Interval:      10 * time.Millisecond,
		Provinces:     []string{"ON"},
		FixtureBase:   "data/fixtures/gazette_%s.json",
		MaxConcurrent: 1,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cycles := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	w.Run(ctx, func(results []PollResult) {
		cycles++
	})
	// At least the initial cycle should have run.
	if cycles == 0 {
		t.Fatal("expected at least one cycle before cancellation")
	}
}

func TestPollResult_Fields(t *testing.T) {
	r := PollResult{
		Province:         "ON",
		Changed:          true,
		PreviousHash:     "abc",
		CurrentHash:      "def",
		DocumentsSeen:     100,
		DocumentsChanged: 5,
	}
	if r.Province != "ON" {
		t.Error("Province field not set")
	}
	if !r.Changed {
		t.Error("Changed field not set")
	}
	if r.CurrentHash != "def" {
		t.Error("CurrentHash field not set")
	}
}

// Ensure adapters import is used.
var _ = adapters.SourceHealth{}