// Package indigenouslinker wires the Indigenous Services Canada Business
// Directory cross-referencing into the ingestion pipeline. After each cycle
// it scans every project and procurement for eligible Indigenous businesses
// and records the resulting partnerships as graph relationships, so the
// downstream graph can answer "which Indigenous businesses could partner on
// this project?" without a separate lookup.
package indigenouslinker

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/indigenous"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

// LinkResult summarises a cross-referencing cycle.
type LinkResult struct {
	ProjectsScanned     int
	ProcurementsScanned int
	RelationshipsCreated int
	Errors              []string
}

// Link runs cross-referencing over the projects and procurements currently in
// the store. It is safe to call repeatedly: duplicate relationships are
// idempotent because the relationship ID is derived from the project and
// entity identifiers.
func Link(ctx context.Context, store database.Store) *LinkResult {
	result := &LinkResult{}

	// Collect all Indigenous businesses from the store.
	entities, err := store.ListEntities(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("list entities: %v", err))
		return result
	}

	indigenousBusinesses := make([]*domain.Entity, 0)
	for _, e := range entities {
		if e != nil && e.EntityType == "IndigenousBusiness" {
			indigenousBusinesses = append(indigenousBusinesses, e)
		}
	}
	if len(indigenousBusinesses) == 0 {
		log.Printf("[INDIGENOUS-LINKER] No Indigenous businesses found in store; skipping cross-reference.")
		return result
	}

	// Scan projects.
	projects, _, err := store.ListProjects(ctx, database.ProjectFilter{Limit: 10_000})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("list projects: %v", err))
		return result
	}
	result.ProjectsScanned = len(projects)
	for _, p := range projects {
		if p == nil {
			continue
		}
		created := linkProjectPartners(ctx, store, p, indigenousBusinesses)
		result.RelationshipsCreated += created
	}

	// Scan procurements.
	procurements, err := store.ListProcurements(ctx, 10_000, 0)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("list procurements: %v", err))
		return result
	}
	result.ProcurementsScanned = len(procurements)
	for _, pr := range procurements {
		if pr == nil {
			continue
		}
		created := linkProcurementSetAsides(ctx, store, pr, indigenousBusinesses)
		result.RelationshipsCreated += created
	}

	log.Printf("[INDIGENOUS-LINKER] Scanned %d projects, %d procurements; created %d relationships\n",
		result.ProjectsScanned, result.ProcurementsScanned, result.RelationshipsCreated)
	return result
}

func linkProjectPartners(ctx context.Context, store database.Store, project *domain.Project, businesses []*domain.Entity) int {
	created := 0
	partners := indigenous.FindIndigenousPartnersForProject(businesses, project)
	for _, partner := range partners {
		rel := &domain.Relationship{
			ID:             identity.StableID("relationship", "indigenous-linker", project.ID+":"+partner.ID+":indigenous_partner"),
			ProjectID:      project.ID,
			SourceEntityID: partner.ID,
			RelationType:   "indigenous_partner",
			Confidence:     domain.ConfidenceSupported,
			EvidenceID:     firstEvidenceID(project),
			CreatedAt:      time.Now().UTC(),
		}
		if err := store.SaveRelationship(ctx, rel); err != nil {
			log.Printf("[INDIGENOUS-LINKER] failed to save relationship %s: %v\n", rel.ID, err)
			continue
		}
		created++
	}
	return created
}

func linkProcurementSetAsides(ctx context.Context, store database.Store, procurement *domain.Procurement, businesses []*domain.Entity) int {
	created := 0
	matches := indigenous.CrossReferenceProcurement(businesses, procurement)
	for _, biz := range matches {
		rel := &domain.Relationship{
			ID:             identity.StableID("relationship", "indigenous-linker", procurement.ID+":"+biz.ID+":set_aside_eligible"),
			ProjectID:      procurement.ProjectID,
			SourceEntityID: biz.ID,
			RelationType:   "set_aside_eligible",
			Confidence:     domain.ConfidenceSupported,
			EvidenceID:     firstProcurementEvidence(procurement),
			CreatedAt:      time.Now().UTC(),
		}
		if err := store.SaveRelationship(ctx, rel); err != nil {
			log.Printf("[INDIGENOUS-LINKER] failed to save relationship %s: %v\n", rel.ID, err)
			continue
		}
		created++
	}
	return created
}

func firstEvidenceID(p *domain.Project) string {
	if len(p.EvidenceIDs) > 0 {
		return p.EvidenceIDs[0]
	}
	return ""
}

func firstProcurementEvidence(p *domain.Procurement) string {
	if p.EvidenceID != "" {
		return p.EvidenceID
	}
	return ""
}

// Ensure strings import is used.
var _ = strings.TrimSpace