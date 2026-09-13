package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// Cache is a bounded LRU cache keyed by adapter name + SHA-256(document).
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

func (c *cacheAdapter) Name() string { return c.next.Name() }
func (c *cacheAdapter) Tier() adapters.SourceTier { return c.next.Tier() }
func (c *cacheAdapter) Health() *adapters.SourceHealth { return c.next.Health() }

func (c *cacheAdapter) Fetch(ctx context.Context) ([]byte, error) {
	data, err := c.next.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	key := c.key(data)
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UTC()
	if ent, ok := c.entries[key]; ok {
		if now.Sub(ent.fetchedAt) < c.cfg.TTL {
			c.touch(key)
			return ent.data, nil
		}
		delete(c.entries, key)
	}
	c.trimLocked()
	c.entries[key] = cacheEntry{data: data, fetchedAt: now}
	c.order = append(c.order, key)
	return data, nil
}

func (c *cacheAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	return c.next.Parse(data)
}

func (c *cacheAdapter) key(data []byte) string {
	sum := sha256.Sum256(data)
	return c.next.Name() + ":" + hex.EncodeToString(sum[:])
}

func (c *cacheAdapter) touch(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, key)
}

func (c *cacheAdapter) trimLocked() {
	for len(c.order) > c.cfg.MaxEntries {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.entries, oldest)
	}
}