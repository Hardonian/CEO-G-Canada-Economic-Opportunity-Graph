package cegs

import (
	"fmt"
	"strings"
	"time"
)

const SpecVersion = "0.1"

// Envelope defines the universal CEGS resource wrapper.
type Envelope struct {
	CEGS          string                 `json:"cegs"`
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	CanonicalName string                 `json:"canonical_name"`
	Jurisdiction  string                 `json:"jurisdiction"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Provenance    []string               `json:"provenance,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}

// Project represents a canonical CEGS project resource.
type Project struct {
	Envelope
	Aliases      []string `json:"aliases,omitempty"`
	Description  string   `json:"description,omitempty"`
	Sector       string   `json:"sector"`
	Subsector    string   `json:"subsector,omitempty"`
	Stage        string   `json:"stage"`
	Capex        Monetary `json:"capex"`
	Proponents   []string `json:"proponents,omitempty"`
	Location     Location `json:"location"`
	SourceStatus string   `json:"source_status"`
}

// Monetary encapsulates standard CEGS currency values.
type Monetary struct {
	Amount     int64         `json:"amount,omitempty"`
	Currency   string        `json:"currency"`
	AmountType string        `json:"amount_type"` // reported, estimated, range, unknown
	Range      *CapitalRange `json:"range,omitempty"`
}

// CapitalRange represents min/max ranges.
type CapitalRange struct {
	Min int64 `json:"min"`
	Max int64 `json:"max"`
}

// Location represents standardized geographic placement.
type Location struct {
	Name      string   `json:"name"`
	Province  string   `json:"province"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

// Organization represents a corporate, public, or Indigenous entity.
type Organization struct {
	Envelope
	LegalName   string            `json:"legal_name,omitempty"`
	Aliases     []string          `json:"aliases,omitempty"`
	EntityType  string            `json:"entity_type"`
	Website     string            `json:"website,omitempty"`
	Identifiers map[string]string `json:"identifiers,omitempty"`
	Description string            `json:"description,omitempty"`
}

// Event records an immutable state transition.
type Event struct {
	CEGS        string                 `json:"cegs"`
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "event"
	EventType   string                 `json:"event_type"`
	Subject     string                 `json:"subject"`
	OccurredAt  time.Time              `json:"occurred_at"`
	ObservedAt  *time.Time             `json:"observed_at,omitempty"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Evidence    []string               `json:"evidence"`
	Attributes  map[string]interface{} `json:"attributes,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Relationship models an explicit typed graph edge.
type Relationship struct {
	CEGS             string     `json:"cegs"`
	ID               string     `json:"id"`
	Type             string     `json:"type"` // "relationship"
	RelationshipType string     `json:"relationship_type"`
	From             string     `json:"from"`
	To               string     `json:"to"`
	Status           string     `json:"status"`
	ValidFrom        time.Time  `json:"valid_from"`
	ValidTo          *time.Time `json:"valid_to,omitempty"`
	Evidence         []string   `json:"evidence,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// Evidence holds cryptographic provenance.
type Evidence struct {
	CEGS               string     `json:"cegs"`
	ID                 string     `json:"id"`
	Type               string     `json:"type"` // "evidence"
	SourceURL          string     `json:"source_url"`
	Publisher          string     `json:"publisher"`
	SourceTier         int        `json:"source_tier"`
	ContentHash        string     `json:"content_hash"`
	RetrievalTimestamp time.Time  `json:"retrieval_timestamp"`
	PublicationDate    *time.Time `json:"publication_date,omitempty"`
	EffectiveDate      *time.Time `json:"effective_date,omitempty"`
	Confidence         string     `json:"confidence"`
	ExtractionMethod   string     `json:"extraction_method"`
	RawSnippet         string     `json:"raw_snippet,omitempty"`
}

// FormatID generates standard CEGS URIs.
func FormatID(resourceType, jurisdiction, slug string) string {
	res := strings.ToLower(resourceType)
	jur := strings.ToLower(jurisdiction)
	sl := strings.ToLower(slug)
	sl = strings.ReplaceAll(sl, " ", "-")
	sl = strings.ReplaceAll(sl, "_", "-")
	if jur == "" {
		jur = "ca"
	}
	return fmt.Sprintf("cegs:%s:%s:%s", res, jur, sl)
}
