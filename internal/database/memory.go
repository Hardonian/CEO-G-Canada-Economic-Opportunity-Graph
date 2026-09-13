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
	signals       map[string]*domain.Signal
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
		signals:       make(map[string]*domain.Signal),
		opportunities: make(map[string]*domain.Opportunity),
		evidence:      make(map[string]*domain.Evidence),
	}
}

func (m *MemoryStore) SaveProject(ctx context.Context, p *domain.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.projects[p.ID]; ok {
		if !existing.CreatedAt.IsZero() && (p.CreatedAt.IsZero() || existing.CreatedAt.Before(p.CreatedAt)) {
			p.CreatedAt = existing.CreatedAt
		}
		p.EvidenceIDs = mergeStrings(existing.EvidenceIDs, p.EvidenceIDs)
		// Preserve useful source facts when a newer source is deliberately
		// silent. Zero coordinates and UNKNOWN capital are absence, not an
		// instruction to erase a value reported by another source.
		if p.Latitude == 0 && p.Longitude == 0 && (existing.Latitude != 0 || existing.Longitude != 0) {
			p.Latitude = existing.Latitude
			p.Longitude = existing.Longitude
		}
		if p.CapexCAD == 0 && p.CapexStatus == domain.ConfidenceUnknown && existing.CapexCAD > 0 {
			p.CapexCAD = existing.CapexCAD
			p.CapexStatus = existing.CapexStatus
		}
		p.ExternalIDs = mergeStringMap(existing.ExternalIDs, p.ExternalIDs)
		p.Metadata = mergeMetadata(existing.Metadata, p.Metadata)
	}
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
	for _, existing := range m.scores[s.ProjectID] {
		if existing.ScoreType == s.ScoreType && existing.ScoreVersion == s.ScoreVersion && existing.InputHash == s.InputHash {
			return nil
		}
	}
	var previous *domain.ProjectScore
	for _, candidate := range m.scores[s.ProjectID] {
		if candidate.ScoreType == s.ScoreType && (previous == nil || candidate.CalculatedAt.After(previous.CalculatedAt)) {
			previous = candidate
		}
	}
	if previous != nil {
		previousValue := previous.ScoreValue
		movement := s.ScoreValue - previous.ScoreValue
		s.PreviousValue = &previousValue
		s.Movement = &movement
		if movement != 0 {
			s.MovementReasons = append(s.MovementReasons, "Persisted scoring inputs changed; compare input hashes and factor decomposition.")
		}
	}
	m.scores[s.ProjectID] = append(m.scores[s.ProjectID], s)

	// Also update the project's fast-lookup map
	if p, ok := m.projects[s.ProjectID]; ok {
		if p.Scores == nil {
			p.Scores = make(map[string]float64)
		}
		p.Scores[s.ScoreType] = s.ScoreValue
		updated := false
		for i, detail := range p.ScoreDetails {
			if detail.ScoreType == s.ScoreType {
				p.ScoreDetails[i] = s
				updated = true
				break
			}
		}
		if !updated { p.ScoreDetails = append(p.ScoreDetails, s) }
	}
	return nil
}

func (m *MemoryStore) ListScoreHistory(ctx context.Context, projectID, scoreType string) ([]*domain.ProjectScore, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*domain.ProjectScore
	for _, score := range m.scores[projectID] {
		if scoreType == "" || score.ScoreType == scoreType {
			result = append(result, score)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CalculatedAt.Before(result[j].CalculatedAt) })
	return result, nil
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

func (m *MemoryStore) ListProcurementsByProject(ctx context.Context, projectID string) ([]*domain.Procurement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*domain.Procurement
	for _, procurement := range m.procurements {
		if procurement.ProjectID == projectID {
			list = append(list, procurement)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
	return list, nil
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
	m.signals[s.ID] = s
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
		DataStatus:        domain.StatusHealthy,
		GeneratedAt:       time.Now().UTC(),
	}

	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	for _, p := range m.projects {
		if p.CapexStatus == domain.ConfidenceVerified || p.CapexStatus == domain.ConfidenceSupported || p.CapexStatus == domain.ConfidenceReported {
			stats.TotalCapexCAD += p.CapexCAD
			stats.SectorBreakdown[string(p.Sector)] += p.CapexCAD
			stats.ProvinceBreakdown[p.Province] += p.CapexCAD
		} else {
			stats.UnknownCapexProjects++
		}
	}

	accelerating := make(map[string]bool)
	stalled := make(map[string]bool)
	for _, signal := range m.signals {
		if !signal.Timestamp.After(weekAgo) {
			continue
		}
		switch signal.Type {
		case domain.SignalTimelineSlip, domain.SignalProjectDelay, domain.SignalPoliticalSupportLoss:
			stalled[signal.ProjectID] = true
		default:
			accelerating[signal.ProjectID] = true
		}
	}
	stats.AcceleratingProjectsCount = len(accelerating)
	stats.StalledProjectsCount = len(stalled)

	for _, item := range m.capitalItems {
		if item.CreatedAt.After(weekAgo) && item.AmountType == "exact" && (item.Status == domain.CapitalCommitted || item.Status == domain.CapitalClosed || item.Status == domain.CapitalDisbursed) {
			stats.CapitalMovingWeekCAD += item.AmountCAD
		}
	}
	for _, procurement := range m.procurements {
		stage := strings.ToUpper(procurement.Stage)
		if stage != "AWARD" && stage != "CANCELLATION" && stage != "COMPLETE" {
			stats.ActiveProcurementsCount++
		}
	}
	if stats.UnknownCapexProjects > 0 {
		stats.DataStatus = domain.StatusPartial
	}
	return stats, nil
}

func mergeStrings(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	result := make([]string, 0, len(left)+len(right))
	for _, values := range [][]string{left, right} {
		for _, value := range values {
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

func mergeStringMap(left, right map[string]string) map[string]string {
	if len(left) == 0 && len(right) == 0 {
		return nil
	}
	merged := make(map[string]string, len(left)+len(right))
	for key, value := range left {
		merged[key] = value
	}
	for key, value := range right {
		merged[key] = value
	}
	return merged
}

func mergeMetadata(left, right map[string]interface{}) map[string]interface{} {
	if len(left) == 0 && len(right) == 0 {
		return nil
	}
	merged := make(map[string]interface{}, len(left)+len(right))
	for key, value := range left {
		merged[key] = value
	}
	for key, value := range right {
		merged[key] = value
	}
	return merged
}
