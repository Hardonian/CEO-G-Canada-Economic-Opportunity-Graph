package connector

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/middleware"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// Connector wraps a single Adapter with its middleware chain, health
// polling cadence, and derived snapshot state.
type Connector struct {
	adapter      adapters.Adapter
	chain        adapters.Adapter // adapter wrapped by middleware
	health       *adapters.SourceHealth
	mu           sync.RWMutex
	previousHash string
	lastHash     string
}

// NewConnector builds a Connector with the standard production middlewares:
// retry, circuit breaker, cache, dedup, instrumentation. Supply nil middlewares
// to use the bare adapter.
func NewConnector(adapter adapters.Adapter, middlewares ...middleware.AdapterMiddleware) *Connector {
	chain := adapter
	for _, m := range middlewares {
		if m != nil {
			chain = m(chain)
		}
	}
	return &Connector{
		adapter: adapter,
		chain:   chain,
		health:  adapter.Health(),
	}
}

// Name returns the underlying adapter's name.
func (c *Connector) Name() string { return c.adapter.Name() }

// Tier returns the underlying adapter's tier.
func (c *Connector) Tier() domain.SourceTier { return c.adapter.Tier() }

// Fetch runs the middleware-wrapped adapter and tracks the last hash.
func (c *Connector) Fetch(ctx context.Context) ([]byte, error) {
	data, err := c.chain.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	hash := adapters.HashDocument(data)
	c.mu.Lock()
	c.previousHash = c.lastHash
	c.lastHash = hash
	c.mu.Unlock()
	return data, err
}

// Parse delegates to the middleware-wrapped adapter.
func (c *Connector) Parse(data []byte) (*adapters.IngestionResult, error) {
	return c.chain.Parse(data)
}

// Health returns the latest health snapshot.
func (c *Connector) Health() *adapters.SourceHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snap := *c.health
	return &snap
}

// LastHash returns the SHA-256 of the last successfully fetched payload.
func (c *Connector) LastHash() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastHash
}

// HasDuplicatePayload reports whether the last fetch produced the same
// content hash as the previous fetch.
func (c *Connector) HasDuplicatePayload() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.previousHash != "" && c.previousHash == c.lastHash
}

// PollHealth refreshes the connector health snapshot.
func (c *Connector) PollHealth() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if h := c.adapter.Health(); h != nil {
		c.health = h
	}
}

// ConnectorStatus aggregates health and freshness information.
type ConnectorStatus struct {
	Name        string    `json:"name"`
	Tier        string    `json:"tier"`
	Status      string    `json:"status"`
	LastAttempt time.Time `json:"last_attempt"`
	LastSuccess time.Time `json:"last_success"`
	LastHash    string    `json:"last_hash"`
	Duplicate   bool      `json:"duplicate"`
}

// Status returns a snapshot suitable for monitoring/export.
func (c *Connector) Status() ConnectorStatus {
	h := c.Health()
	c.mu.RLock()
	lastHash := c.lastHash
	duplicate := c.previousHash != "" && c.previousHash == c.lastHash
	c.mu.RUnlock()
	return ConnectorStatus{
		Name:        h.AdapterName,
		Tier:        strconv.Itoa(int(h.Tier)),
		Status:      h.Status,
		LastAttempt: h.LastAttempt,
		LastSuccess: h.LastSuccess,
		LastHash:    lastHash,
		Duplicate:   duplicate,
	}
}