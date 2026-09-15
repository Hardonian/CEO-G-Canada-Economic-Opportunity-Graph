package cer

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

type rawCERFacility struct {
	FilingID      string  `json:"filing_id"`
	FacilityName  string  `json:"facility_name"`
	ProponentName string  `json:"proponent_name"`
	Province      string  `json:"province"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Sector        string  `json:"sector"`
	Subsector     string  `json:"subsector"`
	Stage         string  `json:"stage"`
	CapexCAD      int64   `json:"capex_cad"`
	FilingURL     string  `json:"filing_url"`
	Summary       string  `json:"summary"`
}

type CERAdapter struct {
	fixturePath string
	health      adapters.SourceHealth
}

func NewCERAdapter(fixturePath string) *CERAdapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/cer_facilities.json"
	}
	return &CERAdapter{
		fixturePath: fixturePath,
		health: adapters.SourceHealth{
			AdapterName: "cer_facilities",
			Tier:        domain.SourceTier1,
			Status:      "UNKNOWN",
			Mode:        "CURATED_SNAPSHOT",
		},
	}
}

func (a *CERAdapter) Name() string {
	return "cer_facilities"
}

func (a *CERAdapter) Tier() domain.SourceTier {
	return domain.SourceTier1
}

func (a *CERAdapter) Health() *adapters.SourceHealth {
	return &a.health
}

func (a *CERAdapter) Fetch(ctx context.Context) ([]byte, error) {
	a.health.LastAttempt = time.Now()
	data, err := adapters.ReadBoundedFile(a.fixturePath, adapters.MaxFixtureBytes)
	if err != nil {
		a.health.Status = "DEGRADED"
		a.health.LastError = err.Error()
		return nil, fmt.Errorf("failed to read CER fixture data: %w", err)
	}
	a.health.LastSuccess = time.Now()
	a.health.Status = "HEALTHY"
	return data, nil
}

func (a *CERAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	var records []rawCERFacility
	if err := json.Unmarshal(data, &records); err != nil {
		a.health.ParseFailures++
		return nil, fmt.Errorf("failed to parse CER JSON: %w", err)
	}

	res := &adapters.IngestionResult{}

	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)

	for _, rec := range records {
		slug := strings.ToLower(strings.ReplaceAll(rec.FacilityName, " ", "-"))
		propSlug := strings.ToLower(strings.ReplaceAll(rec.ProponentName, " ", "-"))

		propID := identity.StableID("entity", "cer", propSlug)
		projID := identity.StableID("project", "cer", rec.FilingID+":"+slug)
		recordHash, err := adapters.HashRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("hash CER filing %q: %w", rec.FilingID, err)
		}
		evID := identity.StableID("evidence", "cer", rec.FilingID+":"+recordHash)
		now := time.Now().UTC()

		evidence := &domain.Evidence{
			ID:                 evID,
			SourceURL:          rec.FilingURL,
			Publisher:          "Canada Energy Regulator (CER) / Régie de l'énergie du Canada",
			SourceTier:         domain.SourceTier1,
			Visibility:         domain.VisibilityPublicAttribution,
			Publishable:        true,
			RetrievalTimestamp: now,
			Confidence:         domain.ConfidenceVerified,
			ExtractionMethod:   "cer_regulatory_filing_adapter",
			ContentHash:        recordHash,
			HashScope:          "normalized_source_record",
			SourceClass:        "REGULATOR_FILING",
			SourceRecordID:     rec.FilingID,
			ParserVersion:      "cer-v1",
			RawSnippet:         rec.Summary,
		}

		proponent := &domain.Entity{
			ID:           propID,
			Slug:         propSlug,
			LegalName:    rec.ProponentName,
			CommonName:   rec.ProponentName,
			Aliases:      []string{rec.ProponentName},
			EntityType:   "Utility",
			Jurisdiction: "CA:" + rec.Province,
			EvidenceID:   evID,
			Evidence:     evidence,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		project := &domain.Project{
			ID:                   projID,
			Slug:                 slug,
			Name:                 rec.FacilityName,
			Summary:              rec.Summary,
			Sector:               domain.Sector(rec.Sector),
			Subsector:            rec.Subsector,
			Province:             rec.Province,
			LocationName:         rec.Province,
			Latitude:             rec.Latitude,
			Longitude:            rec.Longitude,
			CurrentStage:         domain.LifecycleStage(rec.Stage),
			CapexCAD:             rec.CapexCAD,
			ProponentID:          propID,
			Proponent:            proponent,
			Confidence:           domain.ConfidenceVerified,
			IsSynthetic:          false,
			LastMeaningfulUpdate: now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		res.Evidence = append(res.Evidence, evidence)
		res.Entities = append(res.Entities, proponent)
		res.Projects = append(res.Projects, project)
	}

	return res, nil
}
