package cer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/google/uuid"
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
			Status:      "HEALTHY",
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
	data, err := os.ReadFile(a.fixturePath)
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

	hash := adapters.HashDocument(data)
	res := &adapters.IngestionResult{}

	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)

	for _, rec := range records {
		slug := strings.ToLower(strings.ReplaceAll(rec.FacilityName, " ", "-"))
		propSlug := strings.ToLower(strings.ReplaceAll(rec.ProponentName, " ", "-"))

		propID := uuid.New().String()
		projID := uuid.New().String()
		evID := uuid.New().String()
		now := time.Now()

		evidence := &domain.Evidence{
			ID:                 evID,
			SourceURL:          rec.FilingURL,
			Publisher:          "Canada Energy Regulator (CER) / Régie de l'énergie du Canada",
			SourceTier:         domain.SourceTier1,
			RetrievalTimestamp: now,
			Confidence:         domain.ConfidenceVerified,
			ExtractionMethod:   "cer_regulatory_filing_adapter",
			ContentHash:        hash,
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
