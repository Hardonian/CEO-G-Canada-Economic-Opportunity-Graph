package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// SourceHealth tracks real-time telemetry and operational reliability for an adapter.
type SourceHealth struct {
	AdapterName      string            `json:"adapter_name"`
	Tier             domain.SourceTier `json:"tier"`
	Status           string            `json:"status"` // HEALTHY, STALE, DEGRADED, BROKEN, DISABLED
	LastAttempt      time.Time         `json:"last_attempt"`
	LastSuccess      time.Time         `json:"last_success"`
	DocumentsSeen    int               `json:"documents_seen"`
	DocumentsChanged int               `json:"documents_changed"`
	ParseFailures    int               `json:"parse_failures"`
	LastError        string            `json:"last_error,omitempty"`
}

// IngestionResult aggregates normalized domain records extracted by an adapter.
type IngestionResult struct {
	Projects      []*domain.Project
	Entities      []*domain.Entity
	Events        []*domain.Event
	Relationships []*domain.Relationship
	Procurements  []*domain.Procurement
	CapitalItems  []*domain.CapitalItem
	Evidence      []*domain.Evidence
}

// Adapter defines the contract for all modular data source adapters.
type Adapter interface {
	Name() string
	Tier() domain.SourceTier
	Fetch(ctx context.Context) ([]byte, error)
	Parse(data []byte) (*IngestionResult, error)
	Health() *SourceHealth
}

// HashDocument produces a deterministic SHA-256 hex string for change detection.
func HashDocument(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
