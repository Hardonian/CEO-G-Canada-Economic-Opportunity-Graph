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
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(DefaultConfig(), fetcher)
	results := w.PollOnce(context.Background())
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	successCount := 0
	for _, r := range results {
		if r.Err == nil && r.CurrentHash != "" {
			successCount++
		}
	}
	if successCount < 4 {
		t.Fatalf("expected 4 successful polls, got %d", successCount)
	}
}

func TestPollOnce_IdempotentSecondRun(t *testing.T) {
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(DefaultConfig(), fetcher)
	w.PollOnce(context.Background())
	results := w.PollOnce(context.Background())
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	changedCount := 0
	for _, r := range results {
		if r.Changed {
			changedCount++
		}
	}
	// Only provinces that succeeded on the first run should be unchanged on the second.
	if changedCount > 0 {
		t.Fatalf("expected 0 changes on second run, got %d", changedCount)
	}
}

func TestPollOnce_ChangeDetection(t *testing.T) {
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(Config{
		Interval:      1 * time.Minute,
		Provinces:     []string{"ON"},
		FixtureBase:   "data/fixtures/gazette_%s.json",
		MaxConcurrent: 1,
	}, fetcher)
	first := w.PollOnce(context.Background())
	if len(first) == 0 {
		t.Fatal("expected results from first run")
	}
	second := w.PollOnce(context.Background())
	for i := range second {
		if second[i].PreviousHash != first[i].CurrentHash {
			t.Fatalf("expected previous hash to match first run's current hash for %s", second[i].Province)
		}
		if second[i].Changed {
			t.Fatalf("expected no change on second run for %s", second[i].Province)
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

func TestWorker_HealthWithInlineFixtures(t *testing.T) {
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(Config{
		Interval:      1 * time.Minute,
		Provinces:     []string{"ON"},
		FixtureBase:   "data/fixtures/gazette_%s.json",
		MaxConcurrent: 1,
	}, fetcher)
	// Health() still uses the default file-based adapter, so just verify it
	// returns one entry with a non-empty adapter name.
	healths := w.Health()
	if len(healths) != 1 {
		t.Fatalf("expected 1 health entry, got %d", len(healths))
	}
	if healths[0] == nil || healths[0].AdapterName == "" {
		t.Fatal("expected non-nil health with non-empty adapter name")
	}

	// PollOnce should succeed with the inline fixture.
	results := w.PollOnce(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("unexpected error: %v", results[0].Err)
	}
	if !results[0].Changed {
		t.Fatal("expected first run to report changed")
	}
	if results[0].DocumentsChanged != 1 {
		t.Fatalf("expected 1 changed document, got %d", results[0].DocumentsChanged)
	}
}

func TestWorker_LastHash(t *testing.T) {
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(Config{
		Interval:      1 * time.Minute,
		Provinces:     []string{"ON"},
		FixtureBase:   "data/fixtures/gazette_%s.json",
		MaxConcurrent: 1,
	}, fetcher)
	if h := w.LastHash("ON"); h != "" {
		t.Fatalf("expected empty hash before polling, got %s", h)
	}
	w.PollOnce(context.Background())
	if h := w.LastHash("ON"); h == "" {
		t.Fatal("expected non-empty hash after polling")
	}
}

func TestWorker_RunCancel(t *testing.T) {
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(Config{
		Interval:      10 * time.Millisecond,
		Provinces:     []string{"ON"},
		FixtureBase:   "data/fixtures/gazette_%s.json",
		MaxConcurrent: 1,
	}, fetcher)
	ctx, cancel := context.WithCancel(context.Background())
	cycles := 0
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	w.Run(ctx, func(results []PollResult) {
		cycles++
	})
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

func TestNewTestWorker_CustomFetcher(t *testing.T) {
	const fixture = `{"source":"test","source_url":"https://example.com","retrieved_at":"2026-09-13T00:00:00Z","effective_at":"2026-09-13T00:00:00Z","dataset_vintage":"2026-Q3","notices":[{"notice_id":"TEST-001","title":"Test Project","content":"Test content","publish_date":"2026-09-10","source_url":"https://example.com/1","category":"Environmental Assessment","proponent_name":"Test Proponent","project_name":"Test Project","location":"Test Location","amount_text":"C$100 million","metadata":{"naics_code":"221114"}}]}`
	fetcher := FetchFunc(func(ctx context.Context, province string) ([]byte, *adapters.SourceHealth, error) {
		return []byte(fixture), nil, nil
	})
	w := NewTestWorker(DefaultConfig(), fetcher)
	results := w.PollOnce(context.Background())
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	successCount := 0
	for _, r := range results {
		if r.Err == nil && r.Changed {
			successCount++
		}
	}
	if successCount < 1 {
		t.Fatalf("expected at least 1 successful parse, got %d (results=%+v)", successCount, results)
	}
}