package middleware

import (
	"context"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

// CacheConfig holds LRU+TTL configuration for the adapter cache.
type CacheConfig struct {
	// MaxEntries is the maximum number of cached adapter results.
	// Zero defaults to 64.
	MaxEntries int
	// TTL is how long a cached result is considered fresh.
	// Zero defaults to 5 minutes.
	TTL time.Duration
}

func DefaultCacheConfig() CacheConfig {
	return CacheConfig{MaxEntries: 64, TTL: 5 * time.Minute}
}

type cacheAdapter struct {
	next    adapters.Adapter
	cfg     CacheConfig
	mu      sync.Mutex
	entries map[string]cacheEntry
	order   []string // LRU head (most recent) at end
}

type cacheEntry struct {
	data      []byte
	fetchedAt time.Time
}

// Cache is a bounded TTL cache keyed by stable adapter identity. A cache hit
// is decided before calling the upstream adapter, so it actually avoids a
// source request during the configured freshness window.
func Cache(config CacheConfig) AdapterMiddleware {
	if config.MaxEntries <= 0 {
		config.MaxEntries = 64
	}
	if config.TTL <= 0 {
		config.TTL = 5 * time.Minute
	}
	return func(next adapters.Adapter) adapters.Adapter {
		return &cacheAdapter{
			next:    next,
			cfg:     config,
			entries: make(map[string]cacheEntry),
		}
	}
}

func (c *cacheAdapter) Name() string                   { return c.next.Name() }
func (c *cacheAdapter) Tier() adapters.SourceTier      { return c.next.Tier() }
func (c *cacheAdapter) Health() *adapters.SourceHealth { return c.next.Health() }

func (c *cacheAdapter) Fetch(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := c.next.Name()
	now := time.Now().UTC()
	if ent, ok := c.entries[key]; ok {
		if now.Sub(ent.fetchedAt) < c.cfg.TTL {
			c.touch(key)
			return append([]byte(nil), ent.data...), nil
		}
		delete(c.entries, key)
		c.removeFromOrder(key)
	}

	data, err := c.next.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	c.entries[key] = cacheEntry{data: append([]byte(nil), data...), fetchedAt: now}
	c.order = append(c.order, key)
	c.trimLocked()
	return append([]byte(nil), data...), nil
}

func (c *cacheAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	return c.next.Parse(data)
}

func (c *cacheAdapter) touch(key string) {
	c.removeFromOrder(key)
	c.order = append(c.order, key)
}

func (c *cacheAdapter) removeFromOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}

func (c *cacheAdapter) trimLocked() {
	for len(c.order) > c.cfg.MaxEntries {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.entries, oldest)
	}
}
