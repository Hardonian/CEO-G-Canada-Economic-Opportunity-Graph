package connector

import (
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
)

// Registry holds a set of named connectors with lazy initialization
// and an aggregated health snapshot.
type Registry struct {
	mu         sync.RWMutex
	connectors map[string]*Connector
	order      []string // insertion order, stable iteration
}

// NewRegistry constructs an empty connector registry.
func NewRegistry() *Registry {
	return &Registry{connectors: make(map[string]*Connector)}
}

// Register inserts a connector. Registering a duplicate name replaces
// the previous connector but preserves insertion order.
func (r *Registry) Register(conn *Connector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.connectors[conn.Name()]; !exists {
		r.order = append(r.order, conn.Name())
	}
	r.connectors[conn.Name()] = conn
}

// Get returns the connector for a name, or nil if absent.
func (r *Registry) Get(name string) *Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connectors[name]
}

// All returns all connectors in insertion order.
func (r *Registry) All() []*Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*Connector, 0, len(r.order))
	for _, name := range r.order {
		list = append(list, r.connectors[name])
	}
	return list
}

// Names returns connector names in insertion order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, len(r.order))
	copy(names, r.order)
	return names
}

// HealthSnapshot aggregates health for every registered connector.
type HealthSnapshot struct {
	Connectors []*ConnectorStatus `json:"connectors"`
	Healthy    int                `json:"healthy"`
	Degraded   int                `json:"degraded"`
	Broken     int                `json:"broken"`
	Generated  time.Time          `json:"generated_at"`
}

// AggregateHealth polls each connector and builds a combined snapshot.
func (r *Registry) AggregateHealth() *HealthSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snap := &HealthSnapshot{Generated: time.Now().UTC()}
	for _, name := range r.order {
		conn := r.connectors[name]
		conn.PollHealth()
		status := conn.Status()
		snap.Connectors = append(snap.Connectors, &status)
		switch status.Status {
		case "HEALTHY":
			snap.Healthy++
		case "DEGRADED":
			snap.Degraded++
		default:
			snap.Broken++
		}
	}
	return snap
}

// BuildPipelineAdapters returns the middleware-wrapped adapters in
// insertion order for consumption by the ingestion pipeline.
func (r *Registry) BuildPipelineAdapters() []adapters.Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]adapters.Adapter, 0, len(r.order))
	for _, name := range r.order {
		list = append(list, r.connectors[name])
	}
	return list
}