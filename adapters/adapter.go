package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
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
	LastChange       time.Time         `json:"last_change"`
	DocumentsSeen    int               `json:"documents_seen"`
	DocumentsChanged int               `json:"documents_changed"`
	ParseFailures    int               `json:"parse_failures"`
	LastError        string            `json:"last_error,omitempty"`
	RateLimitState   string            `json:"rate_limit_state"`
	Mode             string            `json:"mode"` // CURATED_SNAPSHOT or LIVE
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

const MaxFixtureBytes int64 = 5 << 20

// ReadBoundedFile protects fixture ingestion from unexpectedly large or
// replaced files. Live adapters must apply equivalent response body limits.
func ReadBoundedFile(path string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = MaxFixtureBytes
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := io.LimitReader(f, maxBytes+1)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("source document exceeds %d byte limit", maxBytes)
	}
	return data, nil
}

// HashRecord hashes the normalized record rather than a whole multi-record
// fixture, keeping evidence revisions scoped to the assertion they support.
func HashRecord(record any) (string, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	return HashDocument(data), nil
}
