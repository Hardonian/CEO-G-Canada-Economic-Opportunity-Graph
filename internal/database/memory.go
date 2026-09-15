package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource conflict")
)

// MemoryStore provides a thread-safe in-memory implementation of Store.
type MemoryStore struct {
	mu                      sync.RWMutex
	projects                map[string]*domain.Project
	entities                map[string]*domain.Entity
	events                  map[string]*domain.Event
	relationships           map[string]*domain.Relationship
	scores                  map[string][]*domain.ProjectScore // projectID -> list of scores
	procurements            map[string]*domain.Procurement
	capitalItems            map[string]*domain.CapitalItem
	signals                 map[string]*domain.Signal
	opportunities           map[string]*domain.Opportunity
	claims                  map[string]*domain.Claim
	candidateProjects       map[string]*domain.CandidateProject
	projectPhases           map[string]*domain.ProjectPhase
	capitalRequirements     map[string]*domain.CapitalRequirement
	capitalNeeds            map[string]*domain.CapitalNeed
	milestones              map[string]*domain.Milestone
	readinessAssessments    map[string][]*domain.ReadinessAssessment
	auditEntries            map[string]*domain.AuditEntry
	evidence                map[string]*domain.Evidence
	tradeMetrics            map[string]*domain.TradeMetric
	slugIndex               map[string]string // slug -> project ID
	entityNameIndex         map[string]string // normalized name -> entity ID
	publishers              map[string]*domain.Publisher
	publisherPolicies       map[string]*domain.PublisherPolicy
	sources                 map[string]*domain.Source
	sourceURLIndex          map[string]string
	sourcePrivateConfigs    map[string]*domain.SourcePrivateConfig
	sourceCandidates        map[string]*domain.SourceCandidate
	sourceRelationships     map[string]*domain.SourceRelationship
	sourceRelationshipIndex map[string]string
	sourceVersions          map[string]*domain.SourceVersion
	sourceVersionHashIndex  map[string]string
	sourceChanges           map[string]*domain.SourceChange
	sourceChangeIndex       map[string]string
	sourceHealthChecks      map[string]*domain.SourceHealthCheck
	sourceCheckpoints       map[string]*domain.SourceCheckpoint
	mappingVersions         map[string]*domain.MappingVersion
	mappingVersionIndex     map[string]string
	ingestionJobs           map[string]*domain.IngestionJob
	ingestionJobDedupeIndex map[string]string
	outboxEvents            map[string]*domain.OutboxEvent
	outboxHashIndex         map[string]string

	// Secondary indexes for O(1) lookups.
	signalsByProject map[string][]string // projectID → signal IDs
	eventsByProject  map[string][]string // projectID → event IDs

	// Cached capex aggregate for O(1) radar stats.
	cachedTotalCapex int64
	capexDirty       bool
}

// NewMemoryStore initializes an empty in-memory repository.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		projects:                make(map[string]*domain.Project),
		entities:                make(map[string]*domain.Entity),
		events:                  make(map[string]*domain.Event),
		relationships:           make(map[string]*domain.Relationship),
		scores:                  make(map[string][]*domain.ProjectScore),
		procurements:            make(map[string]*domain.Procurement),
		capitalItems:            make(map[string]*domain.CapitalItem),
		signals:                 make(map[string]*domain.Signal),
		opportunities:           make(map[string]*domain.Opportunity),
		claims:                  make(map[string]*domain.Claim),
		candidateProjects:       make(map[string]*domain.CandidateProject),
		projectPhases:           make(map[string]*domain.ProjectPhase),
		capitalRequirements:     make(map[string]*domain.CapitalRequirement),
		capitalNeeds:            make(map[string]*domain.CapitalNeed),
		milestones:              make(map[string]*domain.Milestone),
		readinessAssessments:    make(map[string][]*domain.ReadinessAssessment),
		auditEntries:            make(map[string]*domain.AuditEntry),
		evidence:                make(map[string]*domain.Evidence),
		tradeMetrics:            make(map[string]*domain.TradeMetric),
		publishers:              make(map[string]*domain.Publisher),
		publisherPolicies:       make(map[string]*domain.PublisherPolicy),
		sources:                 make(map[string]*domain.Source),
		sourceURLIndex:          make(map[string]string),
		sourcePrivateConfigs:    make(map[string]*domain.SourcePrivateConfig),
		sourceCandidates:        make(map[string]*domain.SourceCandidate),
		sourceRelationships:     make(map[string]*domain.SourceRelationship),
		sourceRelationshipIndex: make(map[string]string),
		sourceVersions:          make(map[string]*domain.SourceVersion),
		sourceVersionHashIndex:  make(map[string]string),
		sourceChanges:           make(map[string]*domain.SourceChange),
		sourceChangeIndex:       make(map[string]string),
		sourceHealthChecks:      make(map[string]*domain.SourceHealthCheck),
		sourceCheckpoints:       make(map[string]*domain.SourceCheckpoint),
		mappingVersions:         make(map[string]*domain.MappingVersion),
		mappingVersionIndex:     make(map[string]string),
		ingestionJobs:           make(map[string]*domain.IngestionJob),
		ingestionJobDedupeIndex: make(map[string]string),
		outboxEvents:            make(map[string]*domain.OutboxEvent),
		outboxHashIndex:         make(map[string]string),
		signalsByProject:        make(map[string][]string),
		eventsByProject:         make(map[string][]string),
		capexDirty:              true,
	}
}

func (m *MemoryStore) SaveProject(ctx context.Context, p *domain.Project) error {
	if p == nil || p.ID == "" {
		return errors.New("project id is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.Slug != "" {
		if owner, ok := m.slugIndex[p.Slug]; ok && owner != p.ID {
			return fmt.Errorf("project slug %q belongs to %s: %w", p.Slug, owner, ErrConflict)
		}
	}
	if existing, ok := m.projects[p.ID]; ok {
		if !existing.CreatedAt.IsZero() && (p.CreatedAt.IsZero() || existing.CreatedAt.Before(p.CreatedAt)) {
			p.CreatedAt = existing.CreatedAt
		}
		p.EvidenceIDs = mergeStrings(existing.EvidenceIDs, p.EvidenceIDs)
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
		if !existing.LastMeaningfulUpdate.IsZero() && (p.LastMeaningfulUpdate.IsZero() || p.LastMeaningfulUpdate.Before(existing.LastMeaningfulUpdate)) {
			p.CurrentStage = existing.CurrentStage
			p.LastMeaningfulUpdate = existing.LastMeaningfulUpdate
		}
		if existing.Slug != "" && existing.Slug != p.Slug && m.slugIndex[existing.Slug] == p.ID {
			delete(m.slugIndex, existing.Slug)
		}
	}
	m.projects[p.ID] = p
	// Keep slug index consistent.
	if m.slugIndex == nil {
		m.slugIndex = make(map[string]string, len(m.projects))
	}
	if p.Slug != "" {
		m.slugIndex[p.Slug] = p.ID
	}
	// Mark capex aggregate dirty for lazy recompute.
	m.capexDirty = true
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
	if m.slugIndex == nil {
		m.ensureSlugIndexLocked()
	}
	if id, ok := m.slugIndex[slug]; ok {
		if p, ok := m.projects[id]; ok {
			return p, nil
		}
	}
	// Fallback for any projects not captured by the index yet.
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
		case "buildability", "investability", "supplierability", "strategicity", "trade_resilience":
			scoreI, scoreJ := 0.0, 0.0
			if result[i].Scores != nil {
				scoreI = result[i].Scores[filter.SortBy]
			}
			if result[j].Scores != nil {
				scoreJ = result[j].Scores[filter.SortBy]
			}
			if filter.SortDir == "asc" {
				return scoreI < scoreJ
			}
			return scoreI > scoreJ
		default:
			return result[i].LastMeaningfulUpdate.After(result[j].LastMeaningfulUpdate)
		}
	})

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
	// Keep entity name index consistent.
	if m.entityNameIndex == nil {
		m.entityNameIndex = make(map[string]string, len(m.entities))
	}
	for _, name := range entityNames(e) {
		key := domain.NormalizeLookupName(name)
		if key != "" {
			m.entityNameIndex[key] = e.ID
		}
	}
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
	if m.entityNameIndex == nil {
		m.ensureEntityNameIndexLocked()
	}
	key := domain.NormalizeLookupName(name)
	if id, ok := m.entityNameIndex[key]; ok {
		if e, ok := m.entities[id]; ok {
			return e, nil
		}
	}
	// Fallback linear scan for entities whose names were added before
	// the index was built.
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
	if ev == nil || ev.ID == "" || ev.ProjectID == "" || ev.EvidenceID == "" {
		return errors.New("event id, project id, and evidence id are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[ev.ProjectID]; !ok {
		return fmt.Errorf("event project %q: %w", ev.ProjectID, ErrNotFound)
	}
	if _, ok := m.evidence[ev.EvidenceID]; !ok {
		return fmt.Errorf("event evidence %q: %w", ev.EvidenceID, ErrNotFound)
	}
	m.events[ev.ID] = ev
	// Maintain events-by-project secondary index.
	m.eventsByProject[ev.ProjectID] = append(m.eventsByProject[ev.ProjectID], ev.ID)
	return nil
}

func (m *MemoryStore) ListEventsByProject(ctx context.Context, projectID string) ([]*domain.Event, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := m.eventsByProject[projectID]
	list := make([]*domain.Event, 0, len(ids))
	for _, id := range ids {
		if ev, ok := m.events[id]; ok {
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
		if !updated {
			p.ScoreDetails = append(p.ScoreDetails, s)
		}
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
	// Maintain signals-by-project secondary index.
	if s.ProjectID != "" {
		m.signalsByProject[s.ProjectID] = append(m.signalsByProject[s.ProjectID], s.ID)
	}
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
	if o == nil || o.ID == "" || o.ProjectID == "" {
		return errors.New("opportunity id and project id are required")
	}
	if o.Publishable && !o.Visibility.Public() {
		return errors.New("restricted opportunity cannot be publishable")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.opportunities[o.ID] = o
	return nil
}

func (m *MemoryStore) GetOpportunity(ctx context.Context, id string) (*domain.Opportunity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.opportunities[id]
	if !ok { return nil, ErrNotFound }
	return o, nil
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

func (m *MemoryStore) SaveClaim(ctx context.Context, claim *domain.Claim) error {
	if claim == nil || claim.ID == "" || claim.SubjectID == "" || claim.Predicate == "" || !claim.SourceVisibility.Valid() {
		return errors.New("claim id, subject, predicate, and explicit visibility are required")
	}
	if claim.Publishable && !claim.SourceVisibility.Public() {
		return errors.New("restricted claim cannot be publishable")
	}
	m.mu.Lock(); defer m.mu.Unlock()
	m.claims[claim.ID] = claim
	return nil
}

func (m *MemoryStore) GetClaim(ctx context.Context, id string) (*domain.Claim, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	claim, ok := m.claims[id]
	if !ok { return nil, ErrNotFound }
	return claim, nil
}

func (m *MemoryStore) ListClaimsBySubject(ctx context.Context, subjectID string) ([]*domain.Claim, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	list := []*domain.Claim{}
	for _, claim := range m.claims { if claim.SubjectID == subjectID { list = append(list, claim) } }
	sort.Slice(list, func(i, j int) bool { return list[i].ObservedAt.Before(list[j].ObservedAt) })
	return list, nil
}

func (m *MemoryStore) SaveCandidateProject(ctx context.Context, candidate *domain.CandidateProject) error {
	if candidate == nil || candidate.ID == "" || candidate.SourceID == "" || !candidate.Visibility.Valid() || candidate.Visibility.Public() {
		return errors.New("candidate project requires a private or restricted source visibility")
	}
	m.mu.Lock(); defer m.mu.Unlock()
	m.candidateProjects[candidate.ID] = candidate
	return nil
}

func (m *MemoryStore) GetCandidateProject(ctx context.Context, id string) (*domain.CandidateProject, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	candidate, ok := m.candidateProjects[id]
	if !ok { return nil, ErrNotFound }
	return candidate, nil
}

func (m *MemoryStore) SaveProjectPhase(ctx context.Context, phase *domain.ProjectPhase) error {
	if phase == nil || phase.ID == "" || phase.ProjectID == "" { return errors.New("phase id and project id are required") }
	if phase.Publishable && !phase.Visibility.Public() { return errors.New("restricted phase cannot be publishable") }
	m.mu.Lock(); defer m.mu.Unlock(); m.projectPhases[phase.ID] = phase; return nil
}

func (m *MemoryStore) ListProjectPhases(ctx context.Context, projectID string) ([]*domain.ProjectPhase, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); list := []*domain.ProjectPhase{}
	for _, value := range m.projectPhases { if value.ProjectID == projectID { list = append(list, value) } }
	sort.Slice(list, func(i, j int) bool { return list[i].Sequence < list[j].Sequence }); return list, nil
}

func (m *MemoryStore) SaveCapitalRequirement(ctx context.Context, requirement *domain.CapitalRequirement) error {
	if requirement == nil || requirement.ID == "" || requirement.ProjectID == "" { return errors.New("capital requirement id and project id are required") }
	if requirement.Publishable && !requirement.Visibility.Public() { return errors.New("restricted capital requirement cannot be publishable") }
	m.mu.Lock(); defer m.mu.Unlock(); m.capitalRequirements[requirement.ID] = requirement; return nil
}

func (m *MemoryStore) ListCapitalRequirements(ctx context.Context, projectID string) ([]*domain.CapitalRequirement, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); list := []*domain.CapitalRequirement{}
	for _, value := range m.capitalRequirements { if value.ProjectID == projectID { list = append(list, value) } }
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.Before(list[j].CreatedAt) }); return list, nil
}

func (m *MemoryStore) SaveCapitalNeed(ctx context.Context, need *domain.CapitalNeed) error {
	if need == nil || need.ID == "" || need.ProjectID == "" { return errors.New("capital need id and project id are required") }
	if need.Publishable && !need.Visibility.Public() { return errors.New("restricted capital need cannot be publishable") }
	m.mu.Lock(); defer m.mu.Unlock(); m.capitalNeeds[need.ID] = need; return nil
}

func (m *MemoryStore) GetCapitalNeed(ctx context.Context, id string) (*domain.CapitalNeed, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); need, ok := m.capitalNeeds[id]; if !ok { return nil, ErrNotFound }; return need, nil
}

func (m *MemoryStore) ListCapitalNeeds(ctx context.Context, projectID string) ([]*domain.CapitalNeed, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); list := []*domain.CapitalNeed{}
	for _, value := range m.capitalNeeds { if projectID == "" || value.ProjectID == projectID { list = append(list, value) } }
	sort.Slice(list, func(i, j int) bool { return list[i].UpdatedAt.After(list[j].UpdatedAt) }); return list, nil
}

func (m *MemoryStore) SaveMilestone(ctx context.Context, milestone *domain.Milestone) error {
	if milestone == nil || milestone.ID == "" || milestone.ProjectID == "" { return errors.New("milestone id and project id are required") }
	if milestone.Publishable && !milestone.Visibility.Public() { return errors.New("restricted milestone cannot be publishable") }
	m.mu.Lock(); defer m.mu.Unlock(); m.milestones[milestone.ID] = milestone; return nil
}

func (m *MemoryStore) ListMilestones(ctx context.Context, projectID string) ([]*domain.Milestone, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); list := []*domain.Milestone{}
	for _, value := range m.milestones { if projectID == "" || value.ProjectID == projectID { list = append(list, value) } }
	sort.Slice(list, func(i, j int) bool {
		if list[i].TargetDate == nil { return false }; if list[j].TargetDate == nil { return true }; return list[i].TargetDate.Before(*list[j].TargetDate)
	}); return list, nil
}

func (m *MemoryStore) SaveReadinessAssessment(ctx context.Context, assessment *domain.ReadinessAssessment) error {
	if assessment == nil || assessment.ID == "" || assessment.ProjectID == "" || assessment.MethodologyVersion == "" { return errors.New("readiness id, project id, and methodology are required") }
	m.mu.Lock(); defer m.mu.Unlock()
	for _, value := range m.readinessAssessments[assessment.ProjectID] { if value.InputHash == assessment.InputHash && value.MethodologyVersion == assessment.MethodologyVersion { return nil } }
	m.readinessAssessments[assessment.ProjectID] = append(m.readinessAssessments[assessment.ProjectID], assessment); return nil
}

func (m *MemoryStore) GetLatestReadinessAssessment(ctx context.Context, projectID string) (*domain.ReadinessAssessment, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); values := m.readinessAssessments[projectID]; if len(values) == 0 { return nil, ErrNotFound }
	latest := values[0]; for _, value := range values[1:] { if value.CalculatedAt.After(latest.CalculatedAt) { latest = value } }; return latest, nil
}

func (m *MemoryStore) SaveAuditEntry(ctx context.Context, entry *domain.AuditEntry) error {
	if entry == nil || entry.ID == "" || entry.SubjectID == "" { return errors.New("audit id and subject id are required") }
	m.mu.Lock(); defer m.mu.Unlock(); m.auditEntries[entry.ID] = entry; return nil
}

func (m *MemoryStore) ListAuditEntries(ctx context.Context, subjectID string) ([]*domain.AuditEntry, error) {
	m.mu.RLock(); defer m.mu.RUnlock(); list := []*domain.AuditEntry{}
	for _, value := range m.auditEntries { if subjectID == "" || value.SubjectID == subjectID { list = append(list, value) } }
	sort.Slice(list, func(i, j int) bool { return list[i].OccurredAt.Before(list[j].OccurredAt) }); return list, nil
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

func (m *MemoryStore) SaveTradeMetric(ctx context.Context, metric *domain.TradeMetric) error {
	if metric == nil || metric.ID == "" || metric.Geography == "" || metric.MetricCode == "" || metric.EvidenceID == "" {
		return fmt.Errorf("invalid trade metric")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.evidence[metric.EvidenceID]; !ok {
		return fmt.Errorf("trade metric evidence %q: %w", metric.EvidenceID, ErrNotFound)
	}
	m.tradeMetrics[metric.ID] = metric
	return nil
}

func (m *MemoryStore) ListTradeMetrics(ctx context.Context, geography string) ([]*domain.TradeMetric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics := make([]*domain.TradeMetric, 0, len(m.tradeMetrics))
	for _, metric := range m.tradeMetrics {
		if geography == "" || metric.Geography == geography {
			metrics = append(metrics, metric)
		}
	}
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].MetricCode == metrics[j].MetricCode {
			return metrics[i].ReferencePeriod > metrics[j].ReferencePeriod
		}
		return metrics[i].MetricCode < metrics[j].MetricCode
	})
	return metrics, nil
}

func (m *MemoryStore) GetRadarStats(ctx context.Context) (*RadarStats, error) {
	m.mu.RLock()
	// Use cached capex when clean; recompute lazily when dirty.
	totalCapex := m.cachedTotalCapex
	if m.capexDirty {
		m.mu.RUnlock()
		m.mu.Lock()
		totalCapex = m.recomputeCapexLocked()
		m.cachedTotalCapex = totalCapex
		m.capexDirty = false
		m.mu.Unlock()
		m.mu.RLock()
	}

	stats := &RadarStats{
		TotalProjects:     len(m.projects),
		TotalCapexCAD:     totalCapex,
		SectorBreakdown:   make(map[string]int64),
		ProvinceBreakdown: make(map[string]int64),
		DataStatus:        domain.StatusHealthy,
		GeneratedAt:       time.Now().UTC(),
	}
	if len(m.projects) == 0 {
		stats.DataStatus = domain.StatusUnavailable
		m.mu.RUnlock()
		return stats, nil
	}

	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	for _, p := range m.projects {
		if p.CapexStatus == domain.ConfidenceVerified || p.CapexStatus == domain.ConfidenceSupported || p.CapexStatus == domain.ConfidenceReported {
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
	m.mu.RUnlock()

	stats.AcceleratingProjectsCount = len(accelerating)
	stats.StalledProjectsCount = len(stalled)

	// Capital items and procurements — iterate without holding the lock.
	m.mu.RLock()
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
	m.mu.RUnlock()
	return stats, nil
}

// recomputeCapexLocked recalculates the total CAPEX from all projects. Must be
// called with the write lock held.
func (m *MemoryStore) recomputeCapexLocked() int64 {
	var total int64
	for _, p := range m.projects {
		if p.CapexStatus == domain.ConfidenceVerified ||
			p.CapexStatus == domain.ConfidenceSupported ||
			p.CapexStatus == domain.ConfidenceReported {
			total += p.CapexCAD
		}
	}
	return total
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

func entityNames(entity *domain.Entity) []string {
	if entity == nil {
		return nil
	}
	names := make([]string, 0, 2+len(entity.Aliases))
	names = append(names, entity.LegalName, entity.CommonName)
	names = append(names, entity.Aliases...)
	return names
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return ctx.Err()
}

func clonePublisher(p *domain.Publisher) *domain.Publisher {
	if p == nil {
		return nil
	}
	copy := *p
	return &copy
}

func clonePublisherPolicy(p *domain.PublisherPolicy) *domain.PublisherPolicy {
	if p == nil {
		return nil
	}
	copy := *p
	if p.AllowedSchemes != nil {
		copy.AllowedSchemes = append([]string(nil), p.AllowedSchemes...)
	}
	return &copy
}

func cloneSource(s *domain.Source) *domain.Source {
	if s == nil {
		return nil
	}
	copy := *s
	if s.Geographies != nil {
		copy.Geographies = append([]string(nil), s.Geographies...)
	}
	if s.SubjectTags != nil {
		copy.SubjectTags = append([]string(nil), s.SubjectTags...)
	}
	if s.SectorTags != nil {
		copy.SectorTags = append([]string(nil), s.SectorTags...)
	}
	if s.Languages != nil {
		copy.Languages = append([]string(nil), s.Languages...)
	}
	if s.Capabilities != nil {
		copy.Capabilities = make(map[string]interface{}, len(s.Capabilities))
		for k, v := range s.Capabilities {
			copy.Capabilities[k] = v
		}
	}
	if s.Metadata != nil {
		copy.Metadata = make(map[string]interface{}, len(s.Metadata))
		for k, v := range s.Metadata {
			copy.Metadata[k] = v
		}
	}
	if s.LastCheckedAt != nil {
		t := *s.LastCheckedAt
		copy.LastCheckedAt = &t
	}
	if s.LastSuccessAt != nil {
		t := *s.LastSuccessAt
		copy.LastSuccessAt = &t
	}
	if s.LastChangeAt != nil {
		t := *s.LastChangeAt
		copy.LastChangeAt = &t
	}
	if s.RegisteredAt != nil {
		t := *s.RegisteredAt
		copy.RegisteredAt = &t
	}
	return &copy
}

func cloneSourcePrivateConfig(c *domain.SourcePrivateConfig) *domain.SourcePrivateConfig {
	if c == nil {
		return nil
	}
	copy := *c
	if c.AdapterConfig != nil {
		copy.AdapterConfig = make(map[string]interface{}, len(c.AdapterConfig))
		for k, v := range c.AdapterConfig {
			copy.AdapterConfig[k] = v
		}
	}
	if c.CredentialRefs != nil {
		copy.CredentialRefs = make(map[string]string, len(c.CredentialRefs))
		for k, v := range c.CredentialRefs {
			copy.CredentialRefs[k] = v
		}
	}
	return &copy
}

func cloneSourceCandidate(c *domain.SourceCandidate) *domain.SourceCandidate {
	if c == nil {
		return nil
	}
	copy := *c
	if c.Metadata != nil {
		copy.Metadata = make(map[string]interface{}, len(c.Metadata))
		for k, v := range c.Metadata {
			copy.Metadata[k] = v
		}
	}
	if c.ReviewedAt != nil {
		t := *c.ReviewedAt
		copy.ReviewedAt = &t
	}
	return &copy
}

func cloneSourceRelationship(r *domain.SourceRelationship) *domain.SourceRelationship {
	if r == nil {
		return nil
	}
	copy := *r
	if r.Metadata != nil {
		copy.Metadata = make(map[string]interface{}, len(r.Metadata))
		for k, v := range r.Metadata {
			copy.Metadata[k] = v
		}
	}
	if r.ValidFrom != nil {
		t := *r.ValidFrom
		copy.ValidFrom = &t
	}
	if r.ValidTo != nil {
		t := *r.ValidTo
		copy.ValidTo = &t
	}
	return &copy
}

func cloneSourceVersion(v *domain.SourceVersion) *domain.SourceVersion {
	if v == nil {
		return nil
	}
	copy := *v
	if v.Snapshot.ObjectKey != "" || v.Snapshot.Locator != "" || v.Snapshot.SizeBytes != 0 || v.Snapshot.MediaType != "" || v.Snapshot.Charset != "" || v.Snapshot.ContentEncoding != "" || v.Snapshot.SourceProjection != "" || v.Snapshot.License != "" || v.Snapshot.RetainUntil != nil {
		snap := v.Snapshot
		if v.Snapshot.RetainUntil != nil {
			t := *v.Snapshot.RetainUntil
			snap.RetainUntil = &t
		}
		copy.Snapshot = snap
	}
	if v.Metadata != nil {
		copy.Metadata = make(map[string]interface{}, len(v.Metadata))
		for k, val := range v.Metadata {
			copy.Metadata[k] = val
		}
	}
	if v.PublishedAt != nil {
		t := *v.PublishedAt
		copy.PublishedAt = &t
	}
	if v.EffectiveAt != nil {
		t := *v.EffectiveAt
		copy.EffectiveAt = &t
	}
	if v.RemoteModifiedAt != nil {
		t := *v.RemoteModifiedAt
		copy.RemoteModifiedAt = &t
	}
	return &copy
}

func cloneSourceChange(c *domain.SourceChange) *domain.SourceChange {
	if c == nil {
		return nil
	}
	dup := *c
	if c.ChangedFields != nil {
		dup.ChangedFields = make([]domain.FieldChange, len(c.ChangedFields))
		for i, fc := range c.ChangedFields {
			dup.ChangedFields[i] = fc
		}
	}
	if c.RawDiff != nil {
		dup.RawDiff = make(json.RawMessage, len(c.RawDiff))
		copy(dup.RawDiff, c.RawDiff)
	}
	if c.Metadata != nil {
		dup.Metadata = make(map[string]interface{}, len(c.Metadata))
		for k, v := range c.Metadata {
			dup.Metadata[k] = v
		}
	}
	if c.EffectiveAt != nil {
		t := *c.EffectiveAt
		dup.EffectiveAt = &t
	}
	return &dup
}

func cloneSourceHealthCheck(h *domain.SourceHealthCheck) *domain.SourceHealthCheck {
	if h == nil {
		return nil
	}
	copy := *h
	return &copy
}

func cloneSourceCheckpoint(c *domain.SourceCheckpoint) *domain.SourceCheckpoint {
	if c == nil {
		return nil
	}
	copy := *c
	if c.OpaqueState != nil {
		copy.OpaqueState = make(map[string]interface{}, len(c.OpaqueState))
		for k, v := range c.OpaqueState {
			copy.OpaqueState[k] = v
		}
	}
	if c.LastSeenRemoteAt != nil {
		t := *c.LastSeenRemoteAt
		copy.LastSeenRemoteAt = &t
	}
	if c.LastProcessedAt != nil {
		t := *c.LastProcessedAt
		copy.LastProcessedAt = &t
	}
	if c.LastModified != "" {
		copy.LastModified = c.LastModified
	}
	if c.ETag != "" {
		copy.ETag = c.ETag
	}
	return &copy
}

func cloneMappingVersion(m *domain.MappingVersion) *domain.MappingVersion {
	if m == nil {
		return nil
	}
	dup := *m
	if m.Definition != nil {
		dup.Definition = make(json.RawMessage, len(m.Definition))
		copy(dup.Definition, m.Definition)
	}
	if m.ActivatedAt != nil {
		t := *m.ActivatedAt
		dup.ActivatedAt = &t
	}
	if m.RetiredAt != nil {
		t := *m.RetiredAt
		dup.RetiredAt = &t
	}
	return &dup
}

func cloneIngestionJob(j *domain.IngestionJob) *domain.IngestionJob {
	if j == nil {
		return nil
	}
	copy := *j
	if j.Payload != nil {
		copy.Payload = make(map[string]interface{}, len(j.Payload))
		for k, v := range j.Payload {
			copy.Payload[k] = v
		}
	}
	if j.LeaseExpiresAt != nil {
		t := *j.LeaseExpiresAt
		copy.LeaseExpiresAt = &t
	}
	if j.StartedAt != nil {
		t := *j.StartedAt
		copy.StartedAt = &t
	}
	if j.CompletedAt != nil {
		t := *j.CompletedAt
		copy.CompletedAt = &t
	}
	if j.DeadLetteredAt != nil {
		t := *j.DeadLetteredAt
		copy.DeadLetteredAt = &t
	}
	return &copy
}

func cloneOutboxEvent(e *domain.OutboxEvent) *domain.OutboxEvent {
	if e == nil {
		return nil
	}
	copy := *e
	if e.Payload != nil {
		copy.Payload = make(map[string]interface{}, len(e.Payload))
		for k, v := range e.Payload {
			copy.Payload[k] = v
		}
	}
	if e.LeaseExpiresAt != nil {
		t := *e.LeaseExpiresAt
		copy.LeaseExpiresAt = &t
	}
	if e.EffectiveAt != nil {
		t := *e.EffectiveAt
		copy.EffectiveAt = &t
	}
	if e.PublishedAt != nil {
		t := *e.PublishedAt
		copy.PublishedAt = &t
	}
	return &copy
}

func canonicalSourceURLKey(rawURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func sourceRelationshipKey(r *domain.SourceRelationship) string {
	return fmt.Sprintf("%s|%s|%s", r.FromSourceID, r.RelationshipType, r.ToSourceID)
}

func sourceVersionHashKey(sourceID, contentHash string) string {
	h := sha256.Sum256([]byte(sourceID + "|" + contentHash))
	return hex.EncodeToString(h[:])
}

func sourceCheckpointKey(sourceID, streamKey string) string {
	return sourceID + "|" + streamKey
}

func mappingVersionKey(sourceID, name, version string) string {
	return sourceID + "|" + name + "|" + version
}

func validateSource(s *domain.Source) error {
	if s == nil {
		return fmt.Errorf("source is nil")
	}
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("source id is required")
	}
	if strings.TrimSpace(s.PublisherID) == "" {
		return fmt.Errorf("publisher id is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("source name is required")
	}
	if strings.TrimSpace(s.CanonicalURL) == "" {
		return fmt.Errorf("canonical url is required")
	}
	if !s.Kind.Valid() {
		return fmt.Errorf("invalid source kind")
	}
	if !s.Family.Valid() {
		return fmt.Errorf("invalid source family")
	}
	if !s.AuthorityTier.Valid() {
		return fmt.Errorf("invalid authority tier")
	}
	if !s.LifecycleStatus.Valid() {
		return fmt.Errorf("invalid lifecycle status")
	}
	if !s.HealthStatus.Valid() {
		return fmt.Errorf("invalid health status")
	}
	if !s.TermsStatus.Valid() {
		return fmt.Errorf("invalid terms status")
	}
	if !s.RobotsStatus.Valid() {
		return fmt.Errorf("invalid robots status")
	}
	return nil
}

func validateSourceVersion(v *domain.SourceVersion) error {
	if v == nil {
		return fmt.Errorf("source version is nil")
	}
	if strings.TrimSpace(v.ID) == "" {
		return fmt.Errorf("version id is required")
	}
	if strings.TrimSpace(v.SourceID) == "" {
		return fmt.Errorf("source id is required")
	}
	if v.Sequence < 0 {
		return fmt.Errorf("sequence must be non-negative")
	}
	if strings.TrimSpace(v.ContentHash) == "" {
		return fmt.Errorf("content hash is required")
	}
	if v.ObservedAt.IsZero() {
		return fmt.Errorf("observed at is required")
	}
	if v.RetrievedAt.IsZero() {
		return fmt.Errorf("retrieved at is required")
	}
	if v.ParserVersion == "" {
		return fmt.Errorf("parser version is required")
	}
	if v.MappingVersion == "" {
		return fmt.Errorf("mapping version is required")
	}
	if v.PipelineVersion == "" {
		return fmt.Errorf("pipeline version is required")
	}
	return nil
}

func validateSourceChange(c *domain.SourceChange) error {
	if c == nil {
		return fmt.Errorf("source change is nil")
	}
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("change id is required")
	}
	if strings.TrimSpace(c.SourceID) == "" {
		return fmt.Errorf("source id is required")
	}
	if strings.TrimSpace(c.ToVersionID) == "" {
		return fmt.Errorf("to version id is required")
	}
	if !c.ChangeType.Valid() {
		return fmt.Errorf("invalid change type")
	}
	if !c.Materiality.Valid() {
		return fmt.Errorf("invalid materiality")
	}
	if c.ObservedAt.IsZero() {
		return fmt.Errorf("observed at is required")
	}
	if strings.TrimSpace(c.Fingerprint) == "" {
		return fmt.Errorf("fingerprint is required")
	}
	if c.PipelineVersion == "" {
		return fmt.Errorf("pipeline version is required")
	}
	return nil
}

func validateSourceHealthCheck(h *domain.SourceHealthCheck) error {
	if h == nil {
		return fmt.Errorf("source health check is nil")
	}
	if strings.TrimSpace(h.ID) == "" {
		return fmt.Errorf("health check id is required")
	}
	if strings.TrimSpace(h.SourceID) == "" {
		return fmt.Errorf("source id is required")
	}
	if h.CheckedAt.IsZero() {
		return fmt.Errorf("checked at is required")
	}
	if !h.Status.Valid() {
		return fmt.Errorf("invalid health status")
	}
	return nil
}

func validateSourceCheckpoint(c *domain.SourceCheckpoint) error {
	if c == nil {
		return fmt.Errorf("source checkpoint is nil")
	}
	if strings.TrimSpace(c.SourceID) == "" {
		return fmt.Errorf("source id is required")
	}
	if strings.TrimSpace(c.StreamKey) == "" {
		return fmt.Errorf("stream key is required")
	}
	if c.Revision < 0 {
		return fmt.Errorf("revision must be non-negative")
	}
	return nil
}

func validateOutboxEvent(e *domain.OutboxEvent) error {
	if e == nil {
		return fmt.Errorf("outbox event is nil")
	}
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("event id is required")
	}
	if strings.TrimSpace(e.SourceID) == "" {
		return fmt.Errorf("source id is required")
	}
	if strings.TrimSpace(e.EventType) == "" {
		return fmt.Errorf("event type is required")
	}
	if e.ObservedAt.IsZero() {
		return fmt.Errorf("observed at is required")
	}
	if strings.TrimSpace(e.Hash) == "" {
		return fmt.Errorf("hash is required")
	}
	if e.PipelineVersion == "" {
		return fmt.Errorf("pipeline version is required")
	}
	if !e.Status.Valid() {
		return fmt.Errorf("invalid outbox status")
	}
	return nil
}

func validateIngestionJob(j *domain.IngestionJob) error {
	if j == nil {
		return fmt.Errorf("ingestion job is nil")
	}
	if strings.TrimSpace(j.ID) == "" {
		return fmt.Errorf("job id is required")
	}
	if strings.TrimSpace(j.SourceID) == "" {
		return fmt.Errorf("source id is required")
	}
	if strings.TrimSpace(j.DedupeKey) == "" {
		return fmt.Errorf("dedupe key is required")
	}
	if !j.Queue.Valid() {
		return fmt.Errorf("invalid queue")
	}
	if !j.Mode.Valid() {
		return fmt.Errorf("invalid mode")
	}
	if j.Priority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}
	if j.MaxAttempts <= 0 {
		return fmt.Errorf("max attempts must be positive")
	}
	if j.AvailableAt.IsZero() {
		return fmt.Errorf("available at is required")
	}
	if j.PipelineVersion == "" {
		return fmt.Errorf("pipeline version is required")
	}
	return nil
}

func validateJobLease(job *domain.IngestionJob, workerID string, at time.Time) error {
	if job.LeaseOwner != workerID {
		return ErrLeaseLost
	}
	if job.LeaseExpiresAt == nil || !job.LeaseExpiresAt.After(at) {
		return ErrLeaseLost
	}
	return nil
}

func validateOutboxLease(event *domain.OutboxEvent, workerID string, at time.Time) error {
	if event.LeaseOwner != workerID {
		return ErrLeaseLost
	}
	if event.LeaseExpiresAt == nil || !event.LeaseExpiresAt.After(at) {
		return ErrLeaseLost
	}
	return nil
}

func pageBounds(total, offset, limit int) (int, int) {
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return total, total
	}
	if limit <= 0 {
		limit = 50
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return offset, end
}

func sourceMatchesFilter(source *domain.Source, filter domain.SourceFilter) bool {
	if filter.PublisherID != "" && source.PublisherID != filter.PublisherID {
		return false
	}
	if filter.ParentSourceID != "" && source.ParentSourceID != filter.ParentSourceID {
		return false
	}
	if filter.Jurisdiction != "" && source.Jurisdiction != filter.Jurisdiction {
		return false
	}
	if filter.Sector != "" {
		found := false
		for _, tag := range source.SectorTags {
			if tag == filter.Sector {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Language != "" {
		found := false
		for _, lang := range source.Languages {
			if lang == filter.Language {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Family != "" && source.Family != filter.Family {
		return false
	}
	if filter.Kind != "" && source.Kind != filter.Kind {
		return false
	}
	if filter.LifecycleStatus != "" && source.LifecycleStatus != filter.LifecycleStatus {
		return false
	}
	if filter.HealthStatus != "" && source.HealthStatus != filter.HealthStatus {
		return false
	}
	if filter.AuthorityTier != 0 && source.AuthorityTier != filter.AuthorityTier {
		return false
	}
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		combined := strings.ToLower(source.Name + " " + source.CanonicalURL + " " + strings.Join(source.SubjectTags, " "))
		if !strings.Contains(combined, searchLower) {
			return false
		}
	}
	return true
}
