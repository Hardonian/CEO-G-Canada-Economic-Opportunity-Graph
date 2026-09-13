package database

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

var (
	ErrNotFound = errors.New("resource not found")
)

// MemoryStore provides a thread-safe in-memory implementation of Store.
type MemoryStore struct {
	mu            sync.RWMutex
	projects      map[string]*domain.Project
	entities      map[string]*domain.Entity
	events        map[string]*domain.Event
	relationships map[string]*domain.Relationship
	scores        map[string][]*domain.ProjectScore // projectID -> list of scores
	procurements  map[string]*domain.Procurement
	capitalItems  map[string]*domain.CapitalItem
	signals       []*domain.Signal
	opportunities map[string]*domain.Opportunity
	evidence      map[string]*domain.Evidence
}

// NewMemoryStore initializes an empty in-memory repository.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects:      make(map[string]*domain.Project),
		entities:      make(map[string]*domain.Entity),
		events:        make(map[string]*domain.Event),
		relationships: make(map[string]*domain.Relationship),
		scores:        make(map[string][]*domain.ProjectScore),
		procurements:  make(map[string]*domain.Procurement),
		capitalItems:  make(map[string]*domain.CapitalItem),
		signals:       make([]*domain.Signal, 0),
		opportunities: make(map[string]*domain.Opportunity),
		evidence:      make(map[string]*domain.Evidence),
	}
}

func (m *MemoryStore) SaveProject(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[p.ID] = p
	return nil
}

func (m *MemoryStore) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.projects[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (m *MemoryStore) GetProjectBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MemoryStore) ListProjects(ctx context.Context, filter ProjectFilter) ([]*domain.Project, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.Project
	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, p := range m.projects {
		if filter.Sector != "" && string(p.Sector) != filter.Sector {
			continue
		}
		if filter.Province != "" && p.Province != filter.Province {
			continue
		}
		if filter.Stage != "" && string(p.CurrentStage) != filter.Stage {
			continue
		}
		if filter.MinCapexCAD > 0 && p.CapexCAD < filter.MinCapexCAD {
			continue
		}
		if filter.IsSynthetic != nil && p.IsSynthetic != *filter.IsSynthetic {
			continue
		}
		if searchLower != "" {
			combined := strings.ToLower(p.Name + " " + p.Summary + " " + p.LocationName + " " + p.Subsector)
			if !strings.Contains(combined, searchLower) {
				continue
			}
		}
		result = append(result, p)
	}

	total := len(result)

	// Sort results
	sort.Slice(result, func(i, j int) bool {
		switch filter.SortBy {
		case "capex":
			if filter.SortDir == "asc" {
				return result[i].CapexCAD < result[j].CapexCAD
			}
			return result[i].CapexCAD > result[j].CapexCAD
		case "name":
			if filter.SortDir == "desc" {
				return result[i].Name > result[j].Name
			}
			return result[i].Name < result[j].Name
		case "buildability":
			bI := 0.0
			bJ := 0.0
			if result[i].Scores != nil {
				bI = result[i].Scores["buildability"]
			}
			if result[j].Scores != nil {
				bJ = result[j].Scores["buildability"]
			}
			if filter.SortDir == "asc" {
				return bI < bJ
			}
			return bI > bJ
		case "investability":
			invI := 0.0
			invJ := 0.0
			if result[i].Scores != nil {
				invI = result[i].Scores["investability"]
			}
			if result[j].Scores != nil {
				invJ = result[j].Scores["investability"]
			}
			if filter.SortDir == "asc" {
				return invI < invJ
			}
			return invI > invJ
		default: // default: updated / last meaningful update desc
			return result[i].LastMeaningfulUpdate.After(result[j].LastMeaningfulUpdate)
		}
	})

	// Pagination
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []*domain.Project{}, total, nil
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	end := offset + limit
	if end > total {
		end = total
	}

	return result[offset:end], total, nil
}

func (m *MemoryStore) SaveEntity(ctx context.Context, e *domain.Entity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entities[e.ID] = e
	return nil
}

func (m *MemoryStore) GetEntity(ctx context.Context, id string) (*domain.Entity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entities[id]
	if !ok {
		return nil, ErrNotFound
	}
	return e, nil
}

func (m *MemoryStore) FindEntityByLegalOrAlias(ctx context.Context, name string) (*domain.Entity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	norm := strings.ToLower(strings.TrimSpace(name))
	for _, e := range m.entities {
		if strings.ToLower(e.LegalName) == norm || strings.ToLower(e.CommonName) == norm {
			return e, nil
		}
		for _, alias := range e.Aliases {
			if strings.ToLower(alias) == norm {
				return e, nil
			}
		}
	}
	return nil, ErrNotFound
}

func (m *MemoryStore) ListEntities(ctx context.Context) ([]*domain.Entity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Entity
	for _, e := range m.entities {
		list = append(list, e)
	}
	return list, nil
}

func (m *MemoryStore) SaveEvent(ctx context.Context, ev *domain.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[ev.ID] = ev
	return nil
}

func (m *MemoryStore) ListEventsByProject(ctx context.Context, projectID string) ([]*domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Event
	for _, ev := range m.events {
		if ev.ProjectID == projectID {
			list = append(list, ev)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].EventDate.After(list[j].EventDate)
	})
	return list, nil
}

func (m *MemoryStore) ListRecentEvents(ctx context.Context, limit int) ([]*domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Event
	for _, ev := range m.events {
		list = append(list, ev)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].EventDate.After(list[j].EventDate)
	})
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (m *MemoryStore) SaveRelationship(ctx context.Context, r *domain.Relationship) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.relationships[r.ID] = r
	return nil
}

func (m *MemoryStore) ListRelationshipsByProject(ctx context.Context, projectID string) ([]*domain.Relationship, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Relationship
	for _, r := range m.relationships {
		if r.ProjectID == projectID {
			list = append(list, r)
		}
	}
	return list, nil
}

func (m *MemoryStore) SaveScore(ctx context.Context, s *domain.ProjectScore) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scores[s.ProjectID] = append(m.scores[s.ProjectID], s)

	// Also update the project's fast-lookup map
	if p, ok := m.projects[s.ProjectID]; ok {
		if p.Scores == nil {
			p.Scores = make(map[string]float64)
		}
		p.Scores[s.ScoreType] = s.ScoreValue
	}
	return nil
}

func (m *MemoryStore) GetLatestScores(ctx context.Context, projectID string) (map[string]*domain.ProjectScore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	scores, ok := m.scores[projectID]
	if !ok {
		return make(map[string]*domain.ProjectScore), nil
	}
	latest := make(map[string]*domain.ProjectScore)
	for _, s := range scores {
		existing, has := latest[s.ScoreType]
		if !has || s.CalculatedAt.After(existing.CalculatedAt) {
			latest[s.ScoreType] = s
		}
	}
	return latest, nil
}

func (m *MemoryStore) ListRankings(ctx context.Context, scoreType string, limit int) ([]*domain.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Project
	for _, p := range m.projects {
		if p.Scores != nil {
			if _, ok := p.Scores[scoreType]; ok {
				list = append(list, p)
			}
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Scores[scoreType] > list[j].Scores[scoreType]
	})
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (m *MemoryStore) SaveProcurement(ctx context.Context, p *domain.Procurement) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.procurements[p.ID] = p
	return nil
}

func (m *MemoryStore) ListProcurements(ctx context.Context, limit, offset int) ([]*domain.Procurement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Procurement
	for _, p := range m.procurements {
		list = append(list, p)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})
	total := len(list)
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return []*domain.Procurement{}, nil
	}
	if limit <= 0 {
		limit = 50
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return list[offset:end], nil
}

func (m *MemoryStore) SaveCapitalItem(ctx context.Context, c *domain.CapitalItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capitalItems[c.ID] = c
	return nil
}

func (m *MemoryStore) ListCapitalItemsByProject(ctx context.Context, projectID string) ([]*domain.CapitalItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.CapitalItem
	for _, c := range m.capitalItems {
		if c.ProjectID == projectID {
			list = append(list, c)
		}
	}
	return list, nil
}

func (m *MemoryStore) ListAllCapitalItems(ctx context.Context) ([]*domain.CapitalItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.CapitalItem
	for _, c := range m.capitalItems {
		list = append(list, c)
	}
	return list, nil
}

func (m *MemoryStore) SaveSignal(ctx context.Context, s *domain.Signal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, s)
	return nil
}

func (m *MemoryStore) ListSignals(ctx context.Context, since time.Duration, limit int) ([]*domain.Signal, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cutoff := time.Now().Add(-since)
	var list []*domain.Signal
	for _, s := range m.signals {
		if since == 0 || s.Timestamp.After(cutoff) {
			list = append(list, s)
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Timestamp.After(list[j].Timestamp)
	})
	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (m *MemoryStore) SaveOpportunity(ctx context.Context, o *domain.Opportunity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.opportunities[o.ID] = o
	return nil
}

func (m *MemoryStore) ListOpportunitiesByProject(ctx context.Context, projectID string) ([]*domain.Opportunity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Opportunity
	for _, o := range m.opportunities {
		if o.ProjectID == projectID {
			list = append(list, o)
		}
	}
	return list, nil
}

func (m *MemoryStore) ListAllOpportunities(ctx context.Context) ([]*domain.Opportunity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Opportunity
	for _, o := range m.opportunities {
		list = append(list, o)
	}
	return list, nil
}

func (m *MemoryStore) SaveEvidence(ctx context.Context, e *domain.Evidence) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.evidence[e.ID] = e
	return nil
}

func (m *MemoryStore) GetEvidence(ctx context.Context, id string) (*domain.Evidence, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.evidence[id]
	if !ok {
		return nil, ErrNotFound
	}
	return e, nil
}

func (m *MemoryStore) GetRadarStats(ctx context.Context) (*RadarStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &RadarStats{
		TotalProjects:     len(m.projects),
		SectorBreakdown:   make(map[string]int64),
		ProvinceBreakdown: make(map[string]int64),
	}

	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	for _, p := range m.projects {
		stats.TotalCapexCAD += p.CapexCAD
		stats.SectorBreakdown[string(p.Sector)] += p.CapexCAD
		stats.ProvinceBreakdown[p.Province] += p.CapexCAD

		if p.Scores != nil {
			if b, ok := p.Scores["buildability"]; ok && b >= 75.0 {
				stats.AcceleratingProjectsCount++
			} else if b, ok := p.Scores["buildability"]; ok && b < 40.0 {
				stats.StalledProjectsCount++
			}
		}
	}

	for _, ev := range m.events {
		if ev.EventDate.After(weekAgo) {
			if p, ok := m.projects[ev.ProjectID]; ok {
				// Estimate active capital motion proportionally
				stats.CapitalMovingWeekCAD += p.CapexCAD / 10
			}
		}
	}

	stats.ActiveProcurementsCount = len(m.procurements)
	return stats, nil
}
