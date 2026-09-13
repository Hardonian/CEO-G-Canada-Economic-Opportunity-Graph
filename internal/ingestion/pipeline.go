package ingestion

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/propagation"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/scoring"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/signals"
)

// Pipeline orchestrates fetching, change-detection, entity resolution, and scoring.
type Pipeline struct {
	store      database.Store
	adapters   []adapters.Adapter
	mu         sync.Mutex
	lastHashes map[string]string // adapterName -> SHA-256 hash
}

// NewPipeline constructs an ingestion pipeline over a Store.
func NewPipeline(store database.Store, adapterList []adapters.Adapter) *Pipeline {
	return &Pipeline{
		store:      store,
		adapters:   adapterList,
		lastHashes: make(map[string]string),
	}
}

// IngestionReport summarizes the outcome of a pipeline run.
type IngestionReport struct {
	ProjectsIngested      int           `json:"projects_ingested"`
	EntitiesResolved      int           `json:"entities_resolved"`
	EventsRecorded        int           `json:"events_recorded"`
	ProcurementsIngested  int           `json:"procurements_ingested"`
	CapitalItemsIngested  int           `json:"capital_items_ingested"`
	OpportunitiesDerived  int           `json:"opportunities_derived"`
	SignalsGenerated      int           `json:"signals_generated"`
	SourceHealths         []*adapters.SourceHealth `json:"source_healths"`
	Duration              time.Duration `json:"duration"`
}

// Run executes all registered adapters and performs graph enrichment.
func (p *Pipeline) Run(ctx context.Context) (*IngestionReport, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	start := time.Now()
	report := &IngestionReport{}

	for _, adp := range p.adapters {
		raw, err := adp.Fetch(ctx)
		if err != nil {
			continue
		}

		hash := adapters.HashDocument(raw)
		if last, ok := p.lastHashes[adp.Name()]; ok && last == hash {
			// Document unchanged - skip reprocessing identical payload
			continue
		}
		p.lastHashes[adp.Name()] = hash

		parsed, err := adp.Parse(raw)
		if err != nil {
			continue
		}

		// 1. Ingest Evidence
		for _, ev := range parsed.Evidence {
			_ = p.store.SaveEvidence(ctx, ev)
		}

		// 2. Resolve & Ingest Entities
		entityMap := make(map[string]string) // raw entity ID -> resolved entity ID
		for _, ent := range parsed.Entities {
			resolved, err := p.store.FindEntityByLegalOrAlias(ctx, ent.LegalName)
			if err == nil && resolved != nil {
				entityMap[ent.ID] = resolved.ID
			} else {
				_ = p.store.SaveEntity(ctx, ent)
				entityMap[ent.ID] = ent.ID
				report.EntitiesResolved++
			}
		}

		// 3. Ingest Projects & Link Proponents
		for _, proj := range parsed.Projects {
			if resolvedID, ok := entityMap[proj.ProponentID]; ok {
				proj.ProponentID = resolvedID
			}

			// Check if project already exists
			existing, err := p.store.GetProjectBySlug(ctx, proj.Slug)
			if err == nil && existing != nil {
				// Record stage change event if progressed
				if existing.CurrentStage != proj.CurrentStage {
					ev := &domain.Event{
						ID:            fmt.Sprintf("ev-stage-%d", time.Now().UnixNano()),
						ProjectID:     existing.ID,
						EventType:     "project.stage_changed",
						EventDate:     time.Now(),
						PreviousStage: &existing.CurrentStage,
						NewStage:      &proj.CurrentStage,
						Title:         fmt.Sprintf("Stage updated to %s", proj.CurrentStage),
						Description:   fmt.Sprintf("Project progressed from %s to %s.", existing.CurrentStage, proj.CurrentStage),
						CreatedAt:     time.Now(),
					}
					_ = p.store.SaveEvent(ctx, ev)
					report.EventsRecorded++
				}
				proj.ID = existing.ID
			}

			_ = p.store.SaveProject(ctx, proj)
			report.ProjectsIngested++

			// 4. Propagate Opportunities
			opps := propagation.PropagateOpportunities(proj)
			for _, opp := range opps {
				_ = p.store.SaveOpportunity(ctx, opp)
				report.OpportunitiesDerived++
			}

			// 5. Deterministic Scoring
			sCtx := &scoring.ProjectContext{
				Project: proj,
			}
			bScore := scoring.CalculateBuildability(sCtx)
			iScore := scoring.CalculateInvestability(sCtx)
			supScore := scoring.CalculateSupplierability(sCtx)
			stratScore := scoring.CalculateStrategicity(sCtx)

			_ = p.store.SaveScore(ctx, bScore)
			_ = p.store.SaveScore(ctx, iScore)
			_ = p.store.SaveScore(ctx, supScore)
			_ = p.store.SaveScore(ctx, stratScore)

			// 6. Signal & Momentum Detection
			sigs := signals.DetectSignals(proj, nil, nil, nil)
			for _, s := range sigs {
				_ = p.store.SaveSignal(ctx, s)
				report.SignalsGenerated++
			}
		}

		// 7. Ingest Events
		for _, ev := range parsed.Events {
			_ = p.store.SaveEvent(ctx, ev)
			report.EventsRecorded++
		}

		// 8. Ingest Procurements
		for _, pr := range parsed.Procurements {
			_ = p.store.SaveProcurement(ctx, pr)
			report.ProcurementsIngested++
		}

		// 9. Ingest Capital Items
		for _, capItem := range parsed.CapitalItems {
			_ = p.store.SaveCapitalItem(ctx, capItem)
			report.CapitalItemsIngested++
		}

		report.SourceHealths = append(report.SourceHealths, adp.Health())
	}

	report.Duration = time.Since(start)
	return report, nil
}
