package ingestion

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
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
	Status                domain.IntelligenceStatus `json:"status"`
	Errors                []string        `json:"errors,omitempty"`
	Duration              time.Duration `json:"duration"`
}

// Run executes all registered adapters and performs graph enrichment.
func (p *Pipeline) Run(ctx context.Context) (*IngestionReport, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	start := time.Now()
	report := &IngestionReport{Status: domain.StatusHealthy}
	var runErrors []error
	var ingestedProjects []*domain.Project
	sourcesSucceeded := 0

	for _, adp := range p.adapters {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		raw, err := adp.Fetch(ctx)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s fetch: %v", adp.Name(), err))
			runErrors = append(runErrors, err)
			report.SourceHealths = append(report.SourceHealths, adp.Health())
			continue
		}

		hash := adapters.HashDocument(raw)
		if last, ok := p.lastHashes[adp.Name()]; ok && last == hash {
			// Document unchanged - skip reprocessing identical payload
			sourcesSucceeded++
			report.SourceHealths = append(report.SourceHealths, adp.Health())
			continue
		}
		p.lastHashes[adp.Name()] = hash

		parsed, err := adp.Parse(raw)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s parse: %v", adp.Name(), err))
			runErrors = append(runErrors, err)
			report.SourceHealths = append(report.SourceHealths, adp.Health())
			continue
		}
		sourcesSucceeded++

		// 1. Ingest Evidence
		for _, ev := range parsed.Evidence {
			if err := p.store.SaveEvidence(ctx, ev); err != nil { return report, fmt.Errorf("save evidence %s: %w", ev.ID, err) }
		}

		// 2. Resolve & Ingest Entities
		entityMap := make(map[string]string) // raw entity ID -> resolved entity ID
		for _, ent := range parsed.Entities {
			resolved, err := p.store.FindEntityByLegalOrAlias(ctx, ent.LegalName)
			if err == nil && resolved != nil {
				entityMap[ent.ID] = resolved.ID
			} else {
				if err := p.store.SaveEntity(ctx, ent); err != nil { return report, fmt.Errorf("save entity %s: %w", ent.ID, err) }
				entityMap[ent.ID] = ent.ID
				report.EntitiesResolved++
			}
		}

		// 3. Ingest Projects & Link Proponents. Canonical project IDs are
		// stable; slug matching exists only as a compatibility fallback.
		projectMap := make(map[string]string)
		for _, proj := range parsed.Projects {
			rawProjectID := proj.ID
			if resolvedID, ok := entityMap[proj.ProponentID]; ok {
				proj.ProponentID = resolvedID
			}

			// Check if project already exists
			existing, err := p.store.GetProjectBySlug(ctx, proj.Slug)
			if err == nil && existing != nil {
				// Record stage change event if progressed
				if existing.CurrentStage != proj.CurrentStage {
					eventTime := proj.LastMeaningfulUpdate
					if eventTime.IsZero() { eventTime = time.Now().UTC() }
					ev := &domain.Event{
						ID:            identity.StableID("event", "stage-change", existing.ID+":"+string(existing.CurrentStage)+":"+string(proj.CurrentStage)+":"+eventTime.Format(time.RFC3339Nano)),
						ProjectID:     existing.ID,
						EventType:     "project.stage_changed",
						EventDate:     eventTime,
						PreviousStage: &existing.CurrentStage,
						NewStage:      &proj.CurrentStage,
						Title:         fmt.Sprintf("Stage updated to %s", proj.CurrentStage),
						Description:   fmt.Sprintf("Project progressed from %s to %s.", existing.CurrentStage, proj.CurrentStage),
						CreatedAt:     eventTime,
					}
					if err := p.store.SaveEvent(ctx, ev); err != nil { return report, fmt.Errorf("save stage event %s: %w", ev.ID, err) }
					report.EventsRecorded++
				}
				proj.ID = existing.ID
			}
			projectMap[rawProjectID] = proj.ID
			if err := p.store.SaveProject(ctx, proj); err != nil { return report, fmt.Errorf("save project %s: %w", proj.ID, err) }
			report.ProjectsIngested++
			ingestedProjects = append(ingestedProjects, proj)
		}

		// 4. Ingest graph facts before computing any derived intelligence.
		for _, ev := range parsed.Events {
			if mapped, ok := projectMap[ev.ProjectID]; ok { ev.ProjectID = mapped }
			if err := p.store.SaveEvent(ctx, ev); err != nil { return report, fmt.Errorf("save event %s: %w", ev.ID, err) }
			report.EventsRecorded++
		}

		for _, rel := range parsed.Relationships {
			if mapped, ok := projectMap[rel.ProjectID]; ok { rel.ProjectID = mapped }
			if mapped, ok := entityMap[rel.SourceEntityID]; ok { rel.SourceEntityID = mapped }
			if mapped, ok := entityMap[rel.TargetEntityID]; ok { rel.TargetEntityID = mapped }
			if err := p.store.SaveRelationship(ctx, rel); err != nil { return report, fmt.Errorf("save relationship %s: %w", rel.ID, err) }
		}

		for _, pr := range parsed.Procurements {
			if mapped, ok := projectMap[pr.ProjectID]; ok { pr.ProjectID = mapped }
			if err := p.store.SaveProcurement(ctx, pr); err != nil { return report, fmt.Errorf("save procurement %s: %w", pr.ID, err) }
			report.ProcurementsIngested++
		}

		for _, capItem := range parsed.CapitalItems {
			if mapped, ok := projectMap[capItem.ProjectID]; ok { capItem.ProjectID = mapped }
			if err := p.store.SaveCapitalItem(ctx, capItem); err != nil { return report, fmt.Errorf("save capital item %s: %w", capItem.ID, err) }
			report.CapitalItemsIngested++
		}

		report.SourceHealths = append(report.SourceHealths, adp.Health())
	}

	// 5. Generate template-derived needs, evidence-backed signals, and
	// versioned scores only after their persisted inputs are queryable.
	for _, project := range ingestedProjects {
		opportunities := propagation.PropagateOpportunities(project)
		for _, opportunity := range opportunities {
			if err := p.store.SaveOpportunity(ctx, opportunity); err != nil { return report, fmt.Errorf("save opportunity %s: %w", opportunity.ID, err) }
			report.OpportunitiesDerived++
		}

		events, err := p.store.ListEventsByProject(ctx, project.ID)
		if err != nil { return report, err }
		capitalItems, err := p.store.ListCapitalItemsByProject(ctx, project.ID)
		if err != nil { return report, err }
		relationships, err := p.store.ListRelationshipsByProject(ctx, project.ID)
		if err != nil { return report, err }
		procurements, err := p.store.ListProcurementsByProject(ctx, project.ID)
		if err != nil { return report, err }
		opportunities, err = p.store.ListOpportunitiesByProject(ctx, project.ID)
		if err != nil { return report, err }

		scoreContext := &scoring.ProjectContext{Project: project, CapitalItems: capitalItems, Events: events, Relationships: relationships, Procurements: procurements, Opportunities: opportunities}
		if err := p.store.SaveScore(ctx, scoring.CalculateBuildability(scoreContext)); err != nil { return report, err }

		for _, signal := range signals.DetectSignals(project, events, capitalItems, procurements) {
			if err := p.store.SaveSignal(ctx, signal); err != nil { return report, err }
			report.SignalsGenerated++
		}
	}

	report.Duration = time.Since(start)
	if len(runErrors) > 0 {
		report.Status = domain.StatusDegraded
		if sourcesSucceeded > 0 { report.Status = domain.StatusPartial }
	}
	if sourcesSucceeded == 0 && len(runErrors) > 0 { return report, errors.Join(runErrors...) }
	return report, nil
}
