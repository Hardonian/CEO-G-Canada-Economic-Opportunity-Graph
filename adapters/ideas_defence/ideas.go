package ideas_defence

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

type rawIDEaSRecord struct {
	ChallengeID string  `json:"challenge_id"`
	Title       string  `json:"title"`
	Publisher   string  `json:"publisher"`
	Province    string  `json:"province"`
	Location    string  `json:"location"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Sector      string  `json:"sector"`
	Subsector   string  `json:"subsector"`
	Stage       string  `json:"stage"`
	CapexCAD    int64   `json:"capex_cad"`
	SourceURL   string  `json:"source_url"`
	Summary     string  `json:"summary"`
}

type IDEaSAdapter struct {
	fixturePath string
	health      adapters.SourceHealth
}

func NewIDEaSAdapter(fixturePath string) *IDEaSAdapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/ideas_defence.json"
	}
	return &IDEaSAdapter{
		fixturePath: fixturePath,
		health: adapters.SourceHealth{
			AdapterName: "ideas_defence_arctic",
			Tier:        domain.SourceTier1,
			Status:      "UNKNOWN",
			Mode:        "CURATED_SNAPSHOT",
		},
	}
}

func (a *IDEaSAdapter) Name() string {
	return "ideas_defence_arctic"
}

func (a *IDEaSAdapter) Tier() domain.SourceTier {
	return domain.SourceTier1
}

func (a *IDEaSAdapter) Health() *adapters.SourceHealth {
	return &a.health
}

func (a *IDEaSAdapter) Fetch(ctx context.Context) ([]byte, error) {
	a.health.LastAttempt = time.Now()
	data, err := adapters.ReadBoundedFile(a.fixturePath, adapters.MaxFixtureBytes)
	if err != nil {
		a.health.Status = "DEGRADED"
		a.health.LastError = err.Error()
		return nil, fmt.Errorf("failed to read IDEaS fixture: %w", err)
	}
	a.health.LastSuccess = time.Now()
	a.health.Status = "HEALTHY"
	return data, nil
}

func (a *IDEaSAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	var records []rawIDEaSRecord
	if err := json.Unmarshal(data, &records); err != nil {
		a.health.ParseFailures++
		return nil, fmt.Errorf("failed to parse IDEaS JSON: %w", err)
	}

	res := &adapters.IngestionResult{}

	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)

	for _, rec := range records {
		slug := strings.ToLower(strings.ReplaceAll(rec.Title, " ", "-"))
		if len(slug) > 60 {
			slug = slug[:60]
		}
		projID := identity.StableID("project", "ideas", rec.ChallengeID)
		recordHash, err := adapters.HashRecord(rec)
		if err != nil {
			return nil, fmt.Errorf("hash IDEaS record %q: %w", rec.ChallengeID, err)
		}
		evID := identity.StableID("evidence", "ideas", rec.ChallengeID+":"+recordHash)
		now := time.Now().UTC()

		evidence := &domain.Evidence{
			ID:                 evID,
			SourceURL:          rec.SourceURL,
			Publisher:          rec.Publisher,
			SourceTier:         domain.SourceTier1,
			Visibility:         domain.VisibilityPublicAttribution,
			Publishable:        true,
			RetrievalTimestamp: now,
			Confidence:         domain.ConfidenceVerified,
			ExtractionMethod:   "dnd_ideas_challenge_adapter",
			ContentHash:        recordHash,
			HashScope:          "normalized_source_record",
			SourceClass:        "GOVERNMENT_PROGRAM",
			SourceRecordID:     rec.ChallengeID,
			ParserVersion:      "ideas-v1",
			RawSnippet:         rec.Summary,
		}

		project := &domain.Project{
			ID:                   projID,
			Slug:                 slug,
			Name:                 rec.Title,
			Summary:              rec.Summary,
			Sector:               domain.Sector(rec.Sector),
			Subsector:            rec.Subsector,
			Province:             rec.Province,
			LocationName:         rec.Location,
			Latitude:             rec.Latitude,
			Longitude:            rec.Longitude,
			CurrentStage:         domain.LifecycleStage(rec.Stage),
			CapexCAD:             rec.CapexCAD,
			Confidence:           domain.ConfidenceVerified,
			IsSynthetic:          false,
			LastMeaningfulUpdate: now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		res.Evidence = append(res.Evidence, evidence)
		res.Projects = append(res.Projects, project)
	}

	return res, nil
}
