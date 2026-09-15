package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// Resolver holds a reference to the persistent store and implements all query
// resolvers. Each method corresponds to a root Query field in the SDL.
type Resolver struct {
	store database.Store
}

// NewResolver constructs a Resolver backed by the given store.
func NewResolver(store database.Store) *Resolver {
	return &Resolver{store: store}
}

// ─── query: projects ─────────────────────────────────────────────────────────

func (r *Resolver) resolveProjects(ctx context.Context, args map[string]interface{}, subFields []string) (interface{}, error) {
	filter := database.ProjectFilter{
		Sector:   stringArg(args, "sector"),
		Province: stringArg(args, "province"),
		Stage:    stringArg(args, "stage"),
		Limit:    intArg(args, "limit", 50),
		Offset:   intArg(args, "offset", 0),
	}
	projects, _, err := r.store.ListProjects(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("projects: %w", err)
	}
	return marshalProjects(projects, subFields), nil
}

// ─── query: project ───────────────────────────────────────────────────────────

func (r *Resolver) resolveProject(ctx context.Context, args map[string]interface{}, subFields []string) (interface{}, error) {
	id := stringArg(args, "id")
	if id == "" {
		return nil, fmt.Errorf("project: argument 'id' is required")
	}
	proj, err := r.store.GetProject(ctx, id)
	if err != nil {
		// fall back to slug lookup
		proj, err = r.store.GetProjectBySlug(ctx, id)
		if err != nil {
			return nil, nil // not found → null in GraphQL
		}
	}
	return marshalProject(proj, subFields), nil
}

// ─── query: organizations ─────────────────────────────────────────────────────

func (r *Resolver) resolveOrganizations(ctx context.Context, args map[string]interface{}, subFields []string) (interface{}, error) {
	limit := intArg(args, "limit", 50)
	entities, err := r.store.ListEntities(ctx)
	if err != nil {
		return nil, fmt.Errorf("organizations: %w", err)
	}
	if limit > 0 && len(entities) > limit {
		entities = entities[:limit]
	}
	return marshalOrganizations(entities, subFields), nil
}

// ─── query: events ────────────────────────────────────────────────────────────

func (r *Resolver) resolveEvents(ctx context.Context, args map[string]interface{}, subFields []string) (interface{}, error) {
	limit := intArg(args, "limit", 20)
	projectID := stringArg(args, "projectId")
	var events []*domain.Event
	var err error
	if projectID != "" {
		events, err = r.store.ListEventsByProject(ctx, projectID)
	} else {
		events, err = r.store.ListRecentEvents(ctx, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("events: %w", err)
	}
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}
	return marshalEvents(events, subFields), nil
}

// ─── query: signals ───────────────────────────────────────────────────────────

func (r *Resolver) resolveSignals(ctx context.Context, args map[string]interface{}, subFields []string) (interface{}, error) {
	limit := intArg(args, "limit", 20)
	signals, err := r.store.ListSignals(ctx, 24*time.Hour, limit)
	if err != nil {
		return nil, fmt.Errorf("signals: %w", err)
	}
	return marshalSignals(signals, subFields), nil
}

// ─── query: reconciliation ────────────────────────────────────────────────────

func (r *Resolver) resolveReconciliation(ctx context.Context, _ map[string]interface{}, _ []string) (interface{}, error) {
	// Reconciliation stats are derived from radar stats as a lightweight proxy
	// until a dedicated store method is added.
	stats, err := r.store.GetRadarStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("reconciliation: %w", err)
	}
	return map[string]interface{}{
		"totalRecords": stats.TotalProjects,
		"merged":       0,
		"linked":       0,
		"conflicts":    0,
	}, nil
}

// ─── query: aiSovereignty ─────────────────────────────────────────────────────

func (r *Resolver) resolveAISovereignty(ctx context.Context, _ map[string]interface{}, _ []string) (interface{}, error) {
	entities, err := r.store.ListEntities(ctx)
	if err != nil {
		return nil, fmt.Errorf("aiSovereignty: %w", err)
	}
	out := make([]interface{}, 0, len(entities))
	for _, e := range entities {
		out = append(out, map[string]interface{}{
			"entityId":     e.ID,
			"entityName":   e.CommonName,
			"overallScore": 0.0,
			"tier":         "UNRATED",
		})
	}
	return out, nil
}

// ─── marshal helpers ──────────────────────────────────────────────────────────

// fullProject returns the complete project field map. marshalProject calls
// this and then optionally projects to a subset of fields.
func fullProject(p *domain.Project) map[string]interface{} {
	if p == nil {
		return nil
	}
	return map[string]interface{}{
		"id":            p.ID,
		"name":          p.Name,
		"slug":          p.Slug,
		"sector":        string(p.Sector),
		"province":      p.Province,
		"stage":         string(p.CurrentStage),
		"capexCAD":      float64(p.CapexCAD),
		"buildability":  nil,
		"investability": nil,
		"updatedAt":     p.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func marshalProject(p *domain.Project, fields []string) map[string]interface{} {
	m := fullProject(p)
	return projectFields(m, fields)
}

func marshalProjects(projects []*domain.Project, fields []string) []interface{} {
	out := make([]interface{}, 0, len(projects))
	for _, p := range projects {
		m := fullProject(p)
		out = append(out, projectFields(m, fields))
	}
	return out
}

// projectFields projects a full project map to only the requested fields.
// When fields is nil or empty, the full map is returned unchanged.
func projectFields(m map[string]interface{}, fields []string) map[string]interface{} {
	if len(fields) == 0 {
		return m
	}
	out := make(map[string]interface{}, len(fields))
	for _, f := range fields {
		if v, ok := m[f]; ok {
			out[f] = v
		}
	}
	return out
}

func marshalOrganizations(entities []*domain.Entity, fields []string) []interface{} {
	out := make([]interface{}, 0, len(entities))
	for _, e := range entities {
		m := map[string]interface{}{
			"id":         e.ID,
			"slug":       e.Slug,
			"commonName": e.CommonName,
			"legalName":  e.LegalName,
			"entityType": string(e.EntityType),
			"updatedAt":  e.UpdatedAt.UTC().Format(time.RFC3339),
		}
		out = append(out, projectFields(m, fields))
	}
	return out
}

func marshalEvents(events []*domain.Event, fields []string) []interface{} {
	out := make([]interface{}, 0, len(events))
	for _, ev := range events {
		m := map[string]interface{}{
			"id":          ev.ID,
			"projectId":   ev.ProjectID,
			"eventType":   string(ev.EventType),
			"eventDate":   ev.EventDate.UTC().Format(time.RFC3339),
			"title":       ev.Title,
			"description": ev.Description,
		}
		out = append(out, projectFields(m, fields))
	}
	return out
}

func marshalSignals(signals []*domain.Signal, fields []string) []interface{} {
	out := make([]interface{}, 0, len(signals))
	for _, s := range signals {
		m := map[string]interface{}{
			"id":          s.ID,
			"projectId":   s.ProjectID,
			"signalType":  string(s.Type),
			"strength":    s.Magnitude,
			"detectedAt":  s.Timestamp.UTC().Format(time.RFC3339),
			"description": s.Description,
		}
		out = append(out, projectFields(m, fields))
	}
	return out
}

// ─── argument helpers ─────────────────────────────────────────────────────────

func stringArg(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case float64:
			return int(n)
		case int64:
			return int(n)
		}
	}
	return defaultVal
}
