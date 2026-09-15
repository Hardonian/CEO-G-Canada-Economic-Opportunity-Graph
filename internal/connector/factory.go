// Package connector provides a shared factory that wraps raw adapters with
// production middleware (retry, circuit breaker, cache, dedup, instrumentation)
// and exposes them through a Registry. The factory is intentionally small and
// deterministic: the same adapter list always produces the same middleware chain
// in the same order, so behaviour is reproducible across runs.
package connector

import (
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/middleware"
)

// DefaultMiddlewares returns the canonical production middleware chain. Each
// middleware is applied in order: retry first (outermost), then circuit
// breaker, cache, dedup, and instrumentation (innermost, closest to the raw
// adapter). The order matters: retries wrap circuit-breaking so a tripped
// breaker short-circuits without further attempts.
func DefaultMiddlewares() []middleware.AdapterMiddleware {
	return []middleware.AdapterMiddleware{
		middleware.Retry(middleware.DefaultRetryConfig()),
		middleware.CircuitBreaker(middleware.DefaultCircuitBreakerConfig()),
		middleware.Cache(middleware.DefaultCacheConfig()),
		middleware.Dedup(),
		middleware.Instrumented(),
	}
}

// BuildRegistry constructs a Registry from a list of adapters, wrapping each
// with the supplied middlewares. If middlewares is nil, the default chain is
// used. The returned registry preserves insertion order.
func BuildRegistry(adapterList []adapters.Adapter, middlewares ...middleware.AdapterMiddleware) *Registry {
	if middlewares == nil {
		middlewares = DefaultMiddlewares()
	}
	reg := NewRegistry()
	for _, adp := range adapterList {
		reg.Register(NewConnector(adp, middlewares...))
	}
	return reg
}

// BuildPipelineAdapters is a convenience wrapper around BuildRegistry that
// returns the middleware-wrapped adapters ready for the ingestion pipeline.
func BuildPipelineAdapters(adapterList []adapters.Adapter, middlewares ...middleware.AdapterMiddleware) []adapters.Adapter {
	return BuildRegistry(adapterList, middlewares...).BuildPipelineAdapters()
}

// AdapterNames returns the names of all adapters in insertion order, useful
// for logging and health reporting.
func AdapterNames(adapterList []adapters.Adapter) []string {
	names := make([]string, 0, len(adapterList))
	for _, adp := range adapterList {
		names = append(names, strings.TrimSpace(adp.Name()))
	}
	return names
}