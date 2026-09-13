package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

// CircuitBreakerConfig controls failure thresholds and cooldown.
type CircuitBreakerConfig struct {
	// MaxFailures is the consecutive Fetch failures before opening.
	// Zero defaults to 5.
	MaxFailures int
	// Cooldown is how long the breaker stays open before probing.
	// Zero defaults to 30 seconds.
	Cooldown time.Duration
}

func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{MaxFailures: 5, Cooldown: 30 * time.Second}
}

// CircuitBreaker wraps an adapter with half-open/circuit-open logic.
type circuitBreakerAdapter struct {
	next     adapters.Adapter
	cfg      CircuitBreakerConfig
	mu       sync.Mutex
	failures int
	lastFail time.Time
	state    breakerState // closed | open | halfOpen
}

type breakerState int

const (
	stateClosed breakerState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker returns a middleware that opens after consecutive failures.
func CircuitBreaker(config CircuitBreakerConfig) AdapterMiddleware {
	if config.MaxFailures <= 0 {
		config.MaxFailures = 5
	}
	if config.Cooldown <= 0 {
		config.Cooldown = 30 * time.Second
	}
	return func(next adapters.Adapter) adapters.Adapter {
		return &circuitBreakerAdapter{next: next, cfg: config}
	}
}

func (c *circuitBreakerAdapter) Name() string                   { return c.next.Name() }
func (c *circuitBreakerAdapter) Tier() adapters.SourceTier      { return c.next.Tier() }
func (c *circuitBreakerAdapter) Health() *adapters.SourceHealth { return c.next.Health() }

func (c *circuitBreakerAdapter) Fetch(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	switch c.state {
	case stateOpen:
		if time.Since(c.lastFail) < c.cfg.Cooldown {
			c.mu.Unlock()
			return nil, &CircuitOpenError{adapter: c.next.Name()}
		}
		c.state = stateHalfOpen
	case stateHalfOpen:
		// Exactly one recovery probe is allowed. Other callers fail fast
		// instead of recursively calling Fetch and risking an unbounded spin.
		c.mu.Unlock()
		return nil, &CircuitOpenError{adapter: c.next.Name()}
	default:
	}
	c.mu.Unlock()

	data, err := c.next.Fetch(ctx)
	if err != nil {
		c.mu.Lock()
		c.failures++
		c.lastFail = time.Now().UTC()
		if c.failures >= c.cfg.MaxFailures {
			c.state = stateOpen
		}
		c.mu.Unlock()
		return nil, err
	}

	c.mu.Lock()
	c.failures = 0
	c.state = stateClosed
	c.mu.Unlock()
	return data, nil
}

func (c *circuitBreakerAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	return c.next.Parse(data)
}

// CircuitOpenError is returned when the breaker is open.
type CircuitOpenError struct {
	adapter string
}

func (e *CircuitOpenError) Error() string { return "circuit open for adapter " + e.adapter }
func (e *CircuitOpenError) Unwrap() error { return context.Canceled }
