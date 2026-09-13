// Package official ingests a reviewed, source-linked snapshot of primary
// public records. It is intentionally explicit about operating in curated
// snapshot mode; it does not pretend a checked-in fixture is a live API.
package official

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

const (
	adapterName     = "official_primary_sources"
	pipelineVersion = "official-snapshot-v1"
	parserVersion   = "official-json-v1"
)

type sourceRecord struct {
	SourceID        string                 `json:"source_id"`
	Publisher       string                 `json:"publisher"`
	SourceURL       string                 `json:"source_url"`
	PublicationDate string                 `json:"publication_date"`
	EffectiveDate   string                 `json:"effective_date"`
	RetrievedAt     string                 `json:"retrieved_at"`
	SourceClass     string                 `json:"source_class"`
	SourceTier      domain.SourceTier      `json:"source_tier"`
	Locator         string                 `json:"locator"`
	Excerpt         string                 `json:"excerpt"`
	Confidence      domain.ConfidenceLevel `json:"confidence"`
}

type proponentRecord struct {
	LegalName    string            `json:"legal_name"`
	CommonName   string            `json:"common_name"`
	Aliases      []string          `json:"aliases"`
	EntityType   string            `json:"entity_type"`
	Jurisdiction string            `json:"jurisdiction"`
	Identifiers  map[string]string `json:"identifiers"`
}

type projectRecord struct {
	ExternalID   string                 `json:"external_id"`
	ExternalIDs  map[string]string      `json:"external_ids"`
	Name         string                 `json:"name"`
	Summary      string                 `json:"summary"`
	Sector       domain.Sector          `json:"sector"`
	Subsector    string                 `json:"subsector"`
	Province     string                 `json:"province"`
	LocationName string                 `json:"location_name"`
	Latitude     *float64               `json:"latitude"`
	Longitude    *float64               `json:"longitude"`
	Stage        domain.LifecycleStage  `json:"stage"`
	CapexCAD     *int64                 `json:"capex_cad"`
	CapexStatus  domain.ConfidenceLevel `json:"capex_status"`
	Confidence   domain.ConfidenceLevel `json:"confidence"`
	Proponent    *proponentRecord       `json:"proponent"`
	Sources      []sourceRecord         `json:"sources"`
	Events       []eventRecord          `json:"events"`
	CapitalItems []capitalRecord        `json:"capital_items"`
}

type eventRecord struct {
	ExternalID    string                 `json:"external_id"`
	EventType     string                 `json:"event_type"`
	EventDate     string                 `json:"event_date"`
	PreviousStage *domain.LifecycleStage `json:"previous_stage"`
	NewStage      *domain.LifecycleStage `json:"new_stage"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	SourceID      string                 `json:"source_id"`
}

type capitalRecord struct {
	ExternalID   string                 `json:"external_id"`
	Category     domain.CapitalCategory `json:"category"`
	Status       domain.CapitalStatus   `json:"status"`
	AmountCAD    int64                  `json:"amount_cad"`
	AmountType   string                 `json:"amount_type"`
	ProviderName string                 `json:"provider_name"`
	Notes        string                 `json:"notes"`
	EventDate    string                 `json:"event_date"`
	SourceID     string                 `json:"source_id"`
}

type Adapter struct {
	fixturePath string
	health      adapters.SourceHealth
}

func NewAdapter(fixturePath string) *Adapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/official_records.json"
	}
	return &Adapter{fixturePath: fixturePath, health: adapters.SourceHealth{
		AdapterName:    adapterName,
		Tier:           domain.SourceTier1,
		Status:         string(domain.StatusHealthy),
		RateLimitState: "NOT_APPLICABLE",
		Mode:           "CURATED_SNAPSHOT",
	}}
}

func (a *Adapter) Name() string                   { return adapterName }
func (a *Adapter) Tier() domain.SourceTier        { return domain.SourceTier1 }
func (a *Adapter) Health() *adapters.SourceHealth { return &a.health }

func (a *Adapter) Fetch(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	a.health.LastAttempt = time.Now().UTC()
	data, err := adapters.ReadBoundedFile(a.fixturePath, adapters.MaxFixtureBytes)
	if err != nil {
		a.health.Status = string(domain.StatusDegraded)
		a.health.LastError = err.Error()
		return nil, fmt.Errorf("read official source snapshot: %w", err)
	}
	a.health.LastSuccess = time.Now().UTC()
	a.health.Status = string(domain.StatusHealthy)
	a.health.LastError = ""
	return data, nil
}

func (a *Adapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	var records []projectRecord
	if err := json.Unmarshal(data, &records); err != nil {
		a.health.ParseFailures++
		a.health.Status = string(domain.StatusDegraded)
		return nil, fmt.Errorf("parse official source snapshot: %w", err)
	}

	result := &adapters.IngestionResult{}
	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)
	var lastChange time.Time

	for _, rec := range records {
		if err := validateRecord(rec); err != nil {
			a.health.ParseFailures++
			return nil, err
		}
		projectID := identity.StableID("project", adapterName, rec.ExternalID)
		sourceEvidence := make(map[string]*domain.Evidence, len(rec.Sources))
		var evidenceIDs []string
		var createdAt, updatedAt time.Time

		for _, src := range rec.Sources {
			publication, err := parseTime(src.PublicationDate, "publication_date", src.SourceID)
			if err != nil {
				return nil, err
			}
			effective, err := parseTime(src.EffectiveDate, "effective_date", src.SourceID)
			if err != nil {
				return nil, err
			}
			retrieved, err := parseTime(src.RetrievedAt, "retrieved_at", src.SourceID)
			if err != nil {
				return nil, err
			}
			hash, err := adapters.HashRecord(src)
			if err != nil {
				return nil, err
			}
			evidenceID := identity.StableID("evidence", adapterName, src.SourceID+":"+hash)
			evidence := &domain.Evidence{
				ID: evidenceID, SourceURL: src.SourceURL, Publisher: src.Publisher,
				SourceTier: src.SourceTier, RetrievalTimestamp: retrieved,
				PublicationDate: &publication, EffectiveDate: &effective,
				Confidence: src.Confidence, ExtractionMethod: "human_reviewed_primary_source_snapshot",
				ContentHash: hash, HashScope: "normalized_source_record", SourceClass: src.SourceClass, SourceRecordID: src.SourceID,
				Locator: src.Locator, PipelineVersion: pipelineVersion, ParserVersion: parserVersion,
				RawSnippet: src.Excerpt,
			}
			sourceEvidence[src.SourceID] = evidence
			evidenceIDs = append(evidenceIDs, evidenceID)
			result.Evidence = append(result.Evidence, evidence)
			if createdAt.IsZero() || publication.Before(createdAt) {
				createdAt = publication
			}
			if effective.After(updatedAt) {
				updatedAt = effective
			}
			if effective.After(lastChange) {
				lastChange = effective
			}
		}

		var proponent *domain.Entity
		var proponentID string
		if rec.Proponent != nil {
			proponentID = identity.StableID("entity", "ca", rec.Proponent.LegalName)
			proponent = &domain.Entity{
				ID: proponentID, Slug: identity.Slug(rec.Proponent.CommonName),
				LegalName: rec.Proponent.LegalName, CommonName: rec.Proponent.CommonName,
				Aliases: rec.Proponent.Aliases, EntityType: rec.Proponent.EntityType,
				Jurisdiction: rec.Proponent.Jurisdiction, Identifiers: rec.Proponent.Identifiers,
				EvidenceID: evidenceIDs[0], Evidence: result.Evidence[len(result.Evidence)-len(rec.Sources)],
				CreatedAt: createdAt, UpdatedAt: updatedAt,
			}
			result.Entities = append(result.Entities, proponent)
		}

		project := &domain.Project{
			ID: projectID, Slug: identity.Slug(rec.Name), Name: rec.Name, Summary: rec.Summary,
			Sector: rec.Sector, Subsector: rec.Subsector, Province: rec.Province,
			LocationName: rec.LocationName, CurrentStage: rec.Stage, CapexStatus: rec.CapexStatus,
			ProponentID: proponentID, Proponent: proponent, Confidence: rec.Confidence,
			EvidenceIDs: evidenceIDs, ExternalIDs: rec.ExternalIDs,
			LastMeaningfulUpdate: updatedAt, CreatedAt: createdAt, UpdatedAt: updatedAt,
		}
		if rec.Latitude != nil {
			project.Latitude = *rec.Latitude
		}
		if rec.Longitude != nil {
			project.Longitude = *rec.Longitude
		}
		if rec.CapexCAD != nil {
			project.CapexCAD = *rec.CapexCAD
		}
		result.Projects = append(result.Projects, project)

		if proponent != nil {
			result.Relationships = append(result.Relationships, &domain.Relationship{
				ID:        identity.StableID("relationship", adapterName, rec.ExternalID+":proponent:"+proponentID),
				ProjectID: projectID, SourceEntityID: proponentID, RelationType: "develops",
				Confidence: rec.Confidence, EvidenceID: evidenceIDs[0], CreatedAt: createdAt,
			})
		}

		for _, event := range rec.Events {
			eventDate, err := parseTime(event.EventDate, "event_date", event.ExternalID)
			if err != nil {
				return nil, err
			}
			evidence, ok := sourceEvidence[event.SourceID]
			if !ok {
				return nil, fmt.Errorf("event %q references unknown source %q", event.ExternalID, event.SourceID)
			}
			result.Events = append(result.Events, &domain.Event{
				ID: identity.StableID("event", adapterName, event.ExternalID), ProjectID: projectID,
				EventType: event.EventType, EventDate: eventDate, PreviousStage: event.PreviousStage,
				NewStage: event.NewStage, Title: event.Title, Description: event.Description,
				EvidenceID: evidence.ID, Evidence: evidence, CreatedAt: eventDate,
			})
		}

		for _, item := range rec.CapitalItems {
			eventDate, err := parseTime(item.EventDate, "event_date", item.ExternalID)
			if err != nil {
				return nil, err
			}
			evidence, ok := sourceEvidence[item.SourceID]
			if !ok {
				return nil, fmt.Errorf("capital item %q references unknown source %q", item.ExternalID, item.SourceID)
			}
			result.CapitalItems = append(result.CapitalItems, &domain.CapitalItem{
				ID: identity.StableID("capital", adapterName, item.ExternalID), ProjectID: projectID,
				Category: item.Category, Status: item.Status, AmountCAD: item.AmountCAD,
				AmountType: item.AmountType, ProviderName: item.ProviderName, Notes: item.Notes,
				EvidenceID: evidence.ID, Evidence: evidence, CreatedAt: eventDate,
			})
		}
	}
	a.health.LastChange = lastChange
	return result, nil
}

func validateRecord(rec projectRecord) error {
	if rec.ExternalID == "" || rec.Name == "" || rec.Province == "" || len(rec.Sources) == 0 {
		return fmt.Errorf("official record must include external_id, name, province, and at least one source")
	}
	if rec.Stage == "" {
		return fmt.Errorf("official record %q has no explicit stage state", rec.ExternalID)
	}
	if rec.CapexCAD == nil && rec.CapexStatus != domain.ConfidenceUnknown {
		return fmt.Errorf("official record %q omits capex but status is %s", rec.ExternalID, rec.CapexStatus)
	}
	if rec.CapexCAD != nil && *rec.CapexCAD < 0 {
		return fmt.Errorf("official record %q has negative capex", rec.ExternalID)
	}
	for _, src := range rec.Sources {
		if src.SourceID == "" || src.SourceURL == "" || src.Publisher == "" || src.Excerpt == "" {
			return fmt.Errorf("official record %q contains incomplete source metadata", rec.ExternalID)
		}
		if src.SourceTier < domain.SourceTier1 || src.SourceTier > domain.SourceTier4 {
			return fmt.Errorf("official source %q has invalid source tier %d", src.SourceID, src.SourceTier)
		}
		if src.Confidence != domain.ConfidenceVerified && src.Confidence != domain.ConfidenceSupported && src.Confidence != domain.ConfidenceReported {
			return fmt.Errorf("official source %q uses unsupported confidence %q", src.SourceID, src.Confidence)
		}
	}
	return nil
}

func parseTime(value, field, id string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s for %q: %w", field, id, err)
	}
	return t, nil
}
