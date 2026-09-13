package middleware

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

type middlewareTestAdapter struct {
	mu       sync.Mutex
	calls    int
	payload  []byte
	fetchErr error
	health   adapters.SourceHealth
	started  chan struct{}
	release  chan struct{}
}

func newMiddlewareTestAdapter() *middlewareTestAdapter {
	return &middlewareTestAdapter{
		payload: []byte(`{"ok":true}`),
		health:  adapters.SourceHealth{AdapterName: "middleware-test", Tier: domain.SourceTier1, Status: "UNKNOWN"},
	}
}

func (a *middlewareTestAdapter) Name() string            { return a.health.AdapterName }
func (a *middlewareTestAdapter) Tier() domain.SourceTier { return a.health.Tier }
func (a *middlewareTestAdapter) Health() *adapters.SourceHealth {
	snapshot := a.health
	return &snapshot
}
func (a *middlewareTestAdapter) Fetch(ctx context.Context) ([]byte, error) {
	a.mu.Lock()
	a.calls++
	err := a.fetchErr
	started, release := a.started, a.release
	a.mu.Unlock()
	if started != nil {
		select {
		case started <- struct{}{}:
		default:
		}
	}
	if release != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-release:
		}
	}
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), a.payload...), nil
}
func (a *middlewareTestAdapter) Parse([]byte) (*adapters.IngestionResult, error) {
	return &adapters.IngestionResult{}, nil
}
func (a *middlewareTestAdapter) callCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.calls
}

func TestCacheAvoidsUpstreamFetchWithinTTL(t *testing.T) {
	upstream := newMiddlewareTestAdapter()
	cached := Cache(CacheConfig{MaxEntries: 1, TTL: time.Minute})(upstream)
	first, err := cached.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	first[0] = 'x'
	second, err := cached.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if upstream.callCount() != 1 {
		t.Fatalf("upstream calls = %d, want 1", upstream.callCount())
	}
	if string(second) != `{"ok":true}` {
		t.Fatalf("cached bytes were mutated by caller: %s", second)
	}
}

func TestCircuitBreakerAllowsOneHalfOpenProbe(t *testing.T) {
	upstream := newMiddlewareTestAdapter()
	upstream.fetchErr = errors.New("offline")
	breaker := CircuitBreaker(CircuitBreakerConfig{MaxFailures: 1, Cooldown: time.Millisecond})(upstream)
	if _, err := breaker.Fetch(context.Background()); err == nil {
		t.Fatal("first failure unexpectedly succeeded")
	}
	time.Sleep(2 * time.Millisecond)

	upstream.mu.Lock()
	upstream.fetchErr = nil
	upstream.started = make(chan struct{}, 1)
	upstream.release = make(chan struct{})
	upstream.mu.Unlock()

	probeDone := make(chan error, 1)
	go func() {
		_, err := breaker.Fetch(context.Background())
		probeDone <- err
	}()
	select {
	case <-upstream.started:
	case <-time.After(time.Second):
		t.Fatal("half-open probe did not start")
	}
	if _, err := breaker.Fetch(context.Background()); err == nil {
		t.Fatal("second half-open caller bypassed the recovery probe")
	}
	close(upstream.release)
	if err := <-probeDone; err != nil {
		t.Fatalf("recovery probe failed: %v", err)
	}
	if upstream.callCount() != 2 {
		t.Fatalf("upstream calls = %d, want initial failure plus one probe", upstream.callCount())
	}
}
