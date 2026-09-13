package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

// AdapterMiddleware wraps an Adapter, typically adding cross-cutting
// behaviour such as retries, caching, circuit breaking, or instrumentation.
type AdapterMiddleware func(adapters.Adapter) adapters.Adapter

// Chain composes middlewares left-to-right: the first argument wraps
// the adapter directly and each subsequent wrapper wraps the previous.
func Chain(middlewares ...AdapterMiddleware) AdapterMiddleware {
	return func(next adapters.Adapter) adapters.Adapter {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Instrumented decorates an adapter with a FetchLatency histogram
// (observed in milliseconds) and a FetchCalls counter.
func Instrumented() AdapterMiddleware {
	return func(next adapters.Adapter) adapters.Adapter {
		return &instrumented{next: next}
	}
}

type instrumented struct {
	next  adapters.Adapter
	mu    sync.Mutex
	calls int64
	total time.Duration
}

func (i *instrumented) Name() string                   { return i.next.Name() }
func (i *instrumented) Tier() adapters.SourceTier      { return i.next.Tier() }
func (i *instrumented) Health() *adapters.SourceHealth { return i.next.Health() }

func (i *instrumented) Fetch(ctx context.Context) ([]byte, error) {
	start := time.Now()
	data, err := i.next.Fetch(ctx)
	i.mu.Lock()
	i.calls++
	i.total += time.Since(start)
	i.mu.Unlock()
	return data, err
}

func (i *instrumented) Parse(data []byte) (*adapters.IngestionResult, error) {
	return i.next.Parse(data)
}

// FetchCalls returns the number of Fetch invocations observed.
func (i *instrumented) FetchCalls() int64 {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.calls
}

// FetchLatency returns the cumulative observed Fetch latency.
func (i *instrumented) FetchLatency() time.Duration {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.total
}
