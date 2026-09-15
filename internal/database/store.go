package database

import (
	"context"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// ProjectFilter holds query criteria for project searches.
type ProjectFilter struct {
	Sector      string
	Province    string
	Stage       string
	MinCapexCAD int64
	Search      string
	SortBy      string // capex, name, buildability, investability, updated
	SortDir     string // asc, desc
	Limit       int
	Offset      int
	IsSynthetic *bool
}

// RadarStats provides aggregate metrics for the Canada Capital Radar.
type RadarStats struct {
	TotalProjects             int                       `json:"total_projects"`
	TotalCapexCAD             int64                     `json:"total_capex_cad"`
	CapitalMovingWeekCAD      int64                     `json:"capital_moving_week_cad"`
	AcceleratingProjectsCount int                       `json:"accelerating_projects_count"`
	StalledProjectsCount      int                       `json:"stalled_projects_count"`
	ActiveProcurementsCount   int                       `json:"active_procurements_count"`
	SectorBreakdown           map[string]int64          `json:"sector_breakdown"`
	ProvinceBreakdown         map[string]int64          `json:"province_breakdown"`
	UnknownCapexProjects      int                       `json:"unknown_capex_projects"`
	DataStatus                domain.IntelligenceStatus `json:"data_status"`
	GeneratedAt               time.Time                 `json:"generated_at"`
}

// Store defines persistence operations for CanadaOpportunityGraph.
type Store interface {
	SaveProject(ctx context.Context, p *domain.Project) error
	GetProject(ctx context.Context, id string) (*domain.Project, error)
	GetProjectBySlug(ctx context.Context, slug string) (*domain.Project, error)
	ListProjects(ctx context.Context, filter ProjectFilter) ([]*domain.Project, int, error)

	SaveEntity(ctx context.Context, e *domain.Entity) error
	GetEntity(ctx context.Context, id string) (*domain.Entity, error)
	FindEntityByLegalOrAlias(ctx context.Context, name string) (*domain.Entity, error)
	ListEntities(ctx context.Context) ([]*domain.Entity, error)

	SaveEvent(ctx context.Context, ev *domain.Event) error
	ListEventsByProject(ctx context.Context, projectID string) ([]*domain.Event, error)
	ListRecentEvents(ctx context.Context, limit int) ([]*domain.Event, error)

	SaveRelationship(ctx context.Context, r *domain.Relationship) error
	ListRelationshipsByProject(ctx context.Context, projectID string) ([]*domain.Relationship, error)

	SaveScore(ctx context.Context, s *domain.ProjectScore) error
	GetLatestScores(ctx context.Context, projectID string) (map[string]*domain.ProjectScore, error)
	ListScoreHistory(ctx context.Context, projectID, scoreType string) ([]*domain.ProjectScore, error)
	ListRankings(ctx context.Context, scoreType string, limit int) ([]*domain.Project, error)

	SaveProcurement(ctx context.Context, p *domain.Procurement) error
	ListProcurements(ctx context.Context, limit, offset int) ([]*domain.Procurement, error)
	ListProcurementsByProject(ctx context.Context, projectID string) ([]*domain.Procurement, error)

	SaveCapitalItem(ctx context.Context, c *domain.CapitalItem) error
	ListCapitalItemsByProject(ctx context.Context, projectID string) ([]*domain.CapitalItem, error)
	ListAllCapitalItems(ctx context.Context) ([]*domain.CapitalItem, error)

	SaveSignal(ctx context.Context, s *domain.Signal) error
	ListSignals(ctx context.Context, since time.Duration, limit int) ([]*domain.Signal, error)

	SaveOpportunity(ctx context.Context, o *domain.Opportunity) error
	GetOpportunity(ctx context.Context, id string) (*domain.Opportunity, error)
	ListOpportunitiesByProject(ctx context.Context, projectID string) ([]*domain.Opportunity, error)
	ListAllOpportunities(ctx context.Context) ([]*domain.Opportunity, error)

	SaveClaim(ctx context.Context, claim *domain.Claim) error
	GetClaim(ctx context.Context, id string) (*domain.Claim, error)
	ListClaimsBySubject(ctx context.Context, subjectID string) ([]*domain.Claim, error)
	SaveCandidateProject(ctx context.Context, candidate *domain.CandidateProject) error
	GetCandidateProject(ctx context.Context, id string) (*domain.CandidateProject, error)
	SaveProjectPhase(ctx context.Context, phase *domain.ProjectPhase) error
	ListProjectPhases(ctx context.Context, projectID string) ([]*domain.ProjectPhase, error)
	SaveCapitalRequirement(ctx context.Context, requirement *domain.CapitalRequirement) error
	ListCapitalRequirements(ctx context.Context, projectID string) ([]*domain.CapitalRequirement, error)
	SaveCapitalNeed(ctx context.Context, need *domain.CapitalNeed) error
	GetCapitalNeed(ctx context.Context, id string) (*domain.CapitalNeed, error)
	ListCapitalNeeds(ctx context.Context, projectID string) ([]*domain.CapitalNeed, error)
	SaveMilestone(ctx context.Context, milestone *domain.Milestone) error
	ListMilestones(ctx context.Context, projectID string) ([]*domain.Milestone, error)
	SaveReadinessAssessment(ctx context.Context, assessment *domain.ReadinessAssessment) error
	GetLatestReadinessAssessment(ctx context.Context, projectID string) (*domain.ReadinessAssessment, error)
	SaveAuditEntry(ctx context.Context, entry *domain.AuditEntry) error
	ListAuditEntries(ctx context.Context, subjectID string) ([]*domain.AuditEntry, error)

	SaveEvidence(ctx context.Context, e *domain.Evidence) error
	GetEvidence(ctx context.Context, id string) (*domain.Evidence, error)

	SaveTradeMetric(ctx context.Context, metric *domain.TradeMetric) error
	ListTradeMetrics(ctx context.Context, geography string) ([]*domain.TradeMetric, error)

	GetRadarStats(ctx context.Context) (*RadarStats, error)
}
