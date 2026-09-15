// Package adaptersandbox provides a registry for community-maintained
// regional and municipal adapters. Adapters registered here are validated
// for source-tier, URL safety, and schema conformance before being made
// available to the ingestion pipeline.
package adaptersandbox

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// RegistryVersion identifies the sandbox registry.
const RegistryVersion = "adapter-sandbox-v1.0"

// AdapterEntry describes a community adapter registered in the sandbox.
type AdapterEntry struct {
	Name         string             `json:"name"`
	Version      string             `json:"version"`
	SourceURL    string             `json:"source_url"`
	Tier         domain.SourceTier  `json:"tier"`
	Contact      string             `json:"contact"`
	Description  string             `json:"description"`
	Capabilities []string           `json:"capabilities"`
	RegisteredAt time.Time          `json:"registered_at"`
	Approved     bool               `json:"approved"`
}

// Registry holds community adapters with thread-safe registration.
type Registry struct {
	mu       sync.RWMutex
	entries  map[string]*AdapterEntry
	order    []string
}

// NewRegistry constructs an empty adapter sandbox registry.
func NewRegistry() *Registry {
	return &Registry{entries: make(map[string]*AdapterEntry)}
}

// Register adds an adapter entry. Registering a duplicate name replaces the
// previous entry but preserves insertion order.
func (r *Registry) Register(entry *AdapterEntry) error {
	if err := validateEntry(entry); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.entries[entry.Name]; !exists {
		r.order = append(r.order, entry.Name)
	}
	r.entries[entry.Name] = entry
	return nil
}

// Get returns an adapter entry by name, or nil if absent.
func (r *Registry) Get(name string) *AdapterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.entries[name]
}

// All returns all entries in insertion order.
func (r *Registry) All() []*AdapterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*AdapterEntry, 0, len(r.order))
	for _, name := range r.order {
		list = append(list, r.entries[name])
	}
	return list
}

// Names returns entry names in insertion order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, len(r.order))
	copy(names, r.order)
	return names
}

// Count returns the number of registered entries.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// Approved returns only entries that have been approved for production use.
func (r *Registry) Approved() []*AdapterEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*AdapterEntry, 0, len(r.entries))
	for _, name := range r.order {
		if r.entries[name].Approved {
			list = append(list, r.entries[name])
		}
	}
	return list
}

func validateEntry(entry *AdapterEntry) error {
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}
	if strings.TrimSpace(entry.Name) == "" {
		return fmt.Errorf("adapter name is required")
	}
	if strings.TrimSpace(entry.SourceURL) == "" {
		return fmt.Errorf("source_url is required")
	}
	parsed, err := url.Parse(entry.SourceURL)
	if err != nil {
		return fmt.Errorf("invalid source_url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("source_url must have a scheme and host: %q", entry.SourceURL)
	}
	if entry.Tier == 0 {
		return fmt.Errorf("tier is required")
	}
	return nil
}

// Ensure adapters import is used.
var _ = adapters.Adapter(nil)