package iaac

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

type rawIAACProject struct {
	RegistryID         string  `json:"registry_id"`
	ProjectName        string  `json:"project_name"`
	ProponentName      string  `json:"proponent_name"`
	Province           string  `json:"province"`
	Region             string  `json:"region"`
	Latitude           float64 `json:"latitude"`
	Longitude          float64 `json:"longitude"`
	Sector             string  `json:"sector"`
	Subsector          string  `json:"subsector"`
	CurrentStatus      string  `json:"current_status"`
	EstimatedCapexCAD  int64   `json:"estimated_capex_cad"`
	AnnouncementDate   string  `json:"announcement_date"`
	LastRegistryUpdate string  `json:"last_registry_update"`
	RegistryURL        string  `json:"registry_url"`
	Summary            string  `json:"summary"`
}

type IAACAdapter struct {
	fixturePath string
	health      adapters.SourceHealth
}

func NewIAACAdapter(fixturePath string) *IAACAdapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/iaac_projects.json"
	}
	return &IAACAdapter{
		fixturePath: fixturePath,
		health: adapters.SourceHealth{
			AdapterName: "iaac_registry",
			Tier:        domain.SourceTier1,
			Status:      "UNKNOWN",
			Mode:        "CURATED_SNAPSHOT",
		},
	}
}

func (a *IAACAdapter) Name() string {
	return "iaac_registry"
}

func (a *IAACAdapter) Tier() domain.SourceTier {
	return domain.SourceTier1
}

func (a *IAACAdapter) Health() *adapters.SourceHealth {
	return &a.health
}

func (a *IAACAdapter) Fetch(ctx context.Context) ([]byte, error) {
	a.health.LastAttempt = time.Now()
	data, err := adapters.ReadBoundedFile(a.fixturePath, adapters.MaxFixtureBytes)
	if err != nil {
		a.health.Status = "DEGRADED"
		a.health.LastError = err.Error()
		return nil, fmt.Errorf("failed to read IAAC source data: %w", err)
	}
	a.health.LastSuccess = time.Now()
	a.health.Status = "HEALTHY"
	return data, nil
}

func (a *IAACAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	var records []rawIAACProject
	if err := json.Unmarshal(data, &records); err != nil {
		a.health.ParseFailures++
		return nil, fmt.Errorf("failed to parse IAAC JSON: %w", err)
	}

	res := &adapters.IngestionResult{}

	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)

	for _, rec := range records {
		slug := strings.ToLower(strings.ReplaceAll(rec.ProjectName, " ", "-"))
		propSlug := strings.ToLower(strings.ReplaceAll(rec.ProponentName, " ", "-"))

		propID := identity.StableID("entity", "iaac", propSlug)
		projID := identity.StableID("project", "iaac", rec.RegistryID)
		recordHash, err := adapters.HashRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("hash IAAC registry record %q: %w", rec.RegistryID, err)
		}
		evID := identity.StableID("evidence", "iaac", rec.RegistryID+":"+recordHash)

		now := time.Now().UTC()

		evidence := &domain.Evidence{
			ID:                 evID,
			SourceURL:          rec.RegistryURL,
			Publisher:          "Impact Assessment Agency of Canada",
			SourceTier:         domain.SourceTier1,
			RetrievalTimestamp: now,
			Confidence:         domain.ConfidenceVerified,
			ExtractionMethod:   "official_iaac_json_adapter",
			ContentHash:        recordHash,
			HashScope:          "normalized_source_record",
			SourceClass:        "REGULATOR_REGISTRY",
			SourceRecordID:     rec.RegistryID,
			ParserVersion:      "iaac-v1",
			RawSnippet:         rec.Summary,
		}

		proponent := &domain.Entity{
			ID:           propID,
			Slug:         propSlug,
			LegalName:    rec.ProponentName,
			CommonName:   rec.ProponentName,
			Aliases:      []string{rec.ProponentName},
			EntityType:   "Corporation",
			Jurisdiction: "CA:" + rec.Province,
			EvidenceID:   evID,
			Evidence:     evidence,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if strings.Contains(rec.ProponentName, "Ontario Power Generation") {
			proponent.EntityType = "CrownCorp"
			proponent.Aliases = append(proponent.Aliases, "OPG")
		} else if strings.Contains(rec.ProponentName, "Port Authority") {
			proponent.EntityType = "CrownCorp"
		}

		stage := domain.LifecycleStage(rec.CurrentStatus)
		if stage == "" {
			stage = domain.StageEnvironmentalReview
		}

		project := &domain.Project{
			ID:                   projID,
			Slug:                 slug,
			Name:                 rec.ProjectName,
			Summary:              rec.Summary,
			Sector:               domain.Sector(rec.Sector),
			Subsector:            rec.Subsector,
			Province:             rec.Province,
			LocationName:         rec.Region,
			Latitude:             rec.Latitude,
			Longitude:            rec.Longitude,
			CurrentStage:         stage,
			CapexCAD:             rec.EstimatedCapexCAD,
			ProponentID:          propID,
			Proponent:            proponent,
			Confidence:           domain.ConfidenceVerified,
			IsSynthetic:          false,
			LastMeaningfulUpdate: now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		event := &domain.Event{
			ID:          identity.StableID("event", "iaac", rec.RegistryID+":"+rec.LastRegistryUpdate+":"+rec.CurrentStatus),
			ProjectID:   projID,
			EventType:   "regulatory.impact_assessment_filing",
			EventDate:   now,
			NewStage:    &stage,
			Title:       fmt.Sprintf("Impact Assessment filing recorded for %s", rec.ProjectName),
			Description: rec.Summary,
			EvidenceID:  evID,
			Evidence:    evidence,
			CreatedAt:   now,
		}

		rel := &domain.Relationship{
			ID:             identity.StableID("relationship", "iaac", rec.RegistryID+":proponent:"+propID),
			ProjectID:      projID,
			SourceEntityID: propID,
			TargetEntityID: propID,
			RelationType:   "proponent",
			Confidence:     domain.ConfidenceVerified,
			EvidenceID:     evID,
			CreatedAt:      now,
		}

		res.Evidence = append(res.Evidence, evidence)
		res.Entities = append(res.Entities, proponent)
		res.Projects = append(res.Projects, project)
		res.Events = append(res.Events, event)
		res.Relationships = append(res.Relationships, rel)
	}

	return res, nil
}
