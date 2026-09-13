package nrcan_major_projects

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

type rawNRCanRecord struct {
	NRCanID          string  `json:"nrcan_id"`
	ProjectName      string  `json:"project_name"`
	ProponentName    string  `json:"proponent_name"`
	Province         string  `json:"province"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Sector           string  `json:"sector"`
	Subsector        string  `json:"subsector"`
	Stage            string  `json:"stage"`
	CapexCAD         int64   `json:"capex_cad"`
	CIBFinancingCAD  int64   `json:"cib_financing_cad"`
	NRCanGrantCAD    int64   `json:"nrcan_grant_cad"`
	AnnouncementDate string  `json:"announcement_date"`
	Summary          string  `json:"summary"`
}

type NRCanAdapter struct {
	fixturePath string
	health      adapters.SourceHealth
}

func NewNRCanAdapter(fixturePath string) *NRCanAdapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/nrcan_major_projects.json"
	}
	return &NRCanAdapter{
		fixturePath: fixturePath,
		health: adapters.SourceHealth{
			AdapterName: "nrcan_major_projects",
			Tier:        domain.SourceTier1,
			Status:      "HEALTHY",
		},
	}
}

func (a *NRCanAdapter) Name() string {
	return "nrcan_major_projects"
}

func (a *NRCanAdapter) Tier() domain.SourceTier {
	return domain.SourceTier1
}

func (a *NRCanAdapter) Health() *adapters.SourceHealth {
	return &a.health
}

func (a *NRCanAdapter) Fetch(ctx context.Context) ([]byte, error) {
	a.health.LastAttempt = time.Now()
	data, err := os.ReadFile(a.fixturePath)
	if err != nil {
		a.health.Status = "DEGRADED"
		a.health.LastError = err.Error()
		return nil, fmt.Errorf("failed to read NRCan fixture data: %w", err)
	}
	a.health.LastSuccess = time.Now()
	a.health.Status = "HEALTHY"
	return data, nil
}

func (a *NRCanAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	var records []rawNRCanRecord
	if err := json.Unmarshal(data, &records); err != nil {
		a.health.ParseFailures++
		return nil, fmt.Errorf("failed to parse NRCan JSON: %w", err)
	}

	hash := adapters.HashDocument(data)
	res := &adapters.IngestionResult{}

	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)

	for _, rec := range records {
		slug := strings.ToLower(strings.ReplaceAll(rec.ProjectName, " ", "-"))
		propSlug := strings.ToLower(strings.ReplaceAll(rec.ProponentName, " ", "-"))

		propID := uuid.New().String()
		projID := uuid.New().String()
		evID := uuid.New().String()
		now := time.Now()

		evidence := &domain.Evidence{
			ID:                 evID,
			SourceURL:          "https://natural-resources.canada.ca/transparency/reporting-and-accountability/plans-and-performance-reports/major-projects-inventory",
			Publisher:          "Natural Resources Canada (NRCan)",
			SourceTier:         domain.SourceTier1,
			RetrievalTimestamp: now,
			Confidence:         domain.ConfidenceVerified,
			ExtractionMethod:   "nrcan_mpi_tabular_adapter",
			ContentHash:        hash,
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

		stage := domain.LifecycleStage(rec.Stage)
		if stage == "" {
			stage = domain.StageFeasibility
		}

		project := &domain.Project{
			ID:                   projID,
			Slug:                 slug,
			Name:                 rec.ProjectName,
			Summary:              rec.Summary,
			Sector:               domain.Sector(rec.Sector),
			Subsector:            rec.Subsector,
			Province:             rec.Province,
			LocationName:         rec.Province,
			Latitude:             rec.Latitude,
			Longitude:            rec.Longitude,
			CurrentStage:         stage,
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

		// Capital events
		if rec.CIBFinancingCAD > 0 {
			res.CapitalItems = append(res.CapitalItems, &domain.CapitalItem{
				ID:           uuid.New().String(),
				ProjectID:    projID,
				Category:     domain.CapitalCategoryCIB,
				Status:       domain.CapitalCommitted,
				AmountCAD:    rec.CIBFinancingCAD,
				ProviderName: "Canada Infrastructure Bank (CIB)",
				EvidenceID:   evID,
				Evidence:     evidence,
				CreatedAt:    now,
			})
		}
		if rec.NRCanGrantCAD > 0 {
			res.CapitalItems = append(res.CapitalItems, &domain.CapitalItem{
				ID:           uuid.New().String(),
				ProjectID:    projID,
				Category:     domain.CapitalCategoryGrant,
				Status:       domain.CapitalDisbursed,
				AmountCAD:    rec.NRCanGrantCAD,
				ProviderName: "Natural Resources Canada - Clean Energy Fund",
				EvidenceID:   evID,
				Evidence:     evidence,
				CreatedAt:    now,
			})
		}
	}

	return res, nil
}
