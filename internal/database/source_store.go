package database

import (
	"context"
	"errors"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

var (
	ErrSourceConflict    = errors.New("public data mesh uniqueness conflict")
	ErrInvalidTransition = errors.New("invalid public data mesh state transition")
	ErrLeaseLost         = errors.New("public data mesh lease is not held by this worker")
)

type SourceRelationshipFilter struct {
	SourceID        string
	FromSourceID    string
	ToSourceID      string
	RelationshipType domain.SourceRelationshipType
	Limit           int
	Offset          int
}

type SourceVersionFilter struct {
	SourceID     string
	LastKnownGood *bool
	Limit        int
	Offset       int
}

type SourceChangeFilter struct {
	SourceID   string
	ChangeType domain.SourceChangeType
	Materiality domain.Materiality
	Since      *time.Time
	Limit      int
	Offset     int
}

type SourceHealthFilter struct {
	SourceID string
	Status   domain.SourceHealthStatus
	Since    *time.Time
	Limit    int
	Offset   int
}

type MappingVersionFilter struct {
	SourceID string
	Name     string
	Status   domain.MappingStatus
	Limit    int
	Offset   int
}

type CandidateTransitionRequest struct {
	CandidateID string
	To          domain.SourceLifecycleStatus
	ReviewedBy  string
	Reason      string
	At          time.Time
}

type ClaimJobsRequest struct {
	WorkerID     string
	Queues       []domain.IngestionQueue
	Limit        int
	Now          time.Time
	LeaseDuration time.Duration
}

type CompleteJobRequest struct {
	JobID    string
	WorkerID string
	At       time.Time
}

type FailJobRequest struct {
	JobID       string
	WorkerID    string
	FailureStage string
	Error       string
	RetryAt     *time.Time
	At          time.Time
}

type ReplayJobRequest struct {
	DeadLetterJobID string
	NewJobID        string
	DedupeKey       string
	AvailableAt     time.Time
	RequestedAt     time.Time
}

type ClaimOutboxRequest struct {
	WorkerID     string
	Limit        int
	Now          time.Time
	LeaseDuration time.Duration
}

type CompleteOutboxRequest struct {
	EventID  string
	WorkerID string
	At       time.Time
}

type FailOutboxRequest struct {
	EventID   string
	WorkerID  string
	Error     string
	RetryAt   time.Time
	At        time.Time
}

// SourceObservationCommit is an atomic control-plane unit. Database-backed
// implementations must persist every supplied record and complete JobID in one
// transaction. The graph projection is deliberately outside SourceStore.
type SourceObservationCommit struct {
	JobID       string
	WorkerID    string
	Version     *domain.SourceVersion
	Changes     []*domain.SourceChange
	Health      *domain.SourceHealthCheck
	Checkpoint  *domain.SourceCheckpoint
	Outbox      []*domain.OutboxEvent
	CompletedAt time.Time
}

// SourceStore is the durable Public Data Mesh control-plane contract. It is
// separate from Store so existing graph consumers and test doubles do not gain
// unrelated source-management methods.
type SourceStore interface {
	UpsertPublisher(ctx context.Context, publisher *domain.Publisher) error
	GetPublisher(ctx context.Context, id string) (*domain.Publisher, error)
	ListPublishers(ctx context.Context) ([]*domain.Publisher, error)
	SavePublisherPolicy(ctx context.Context, policy *domain.PublisherPolicy) error
	GetPublisherPolicy(ctx context.Context, publisherID string) (*domain.PublisherPolicy, error)

	UpsertSource(ctx context.Context, source *domain.Source) error
	GetSource(ctx context.Context, id string) (*domain.Source, error)
	ListSources(ctx context.Context, filter domain.SourceFilter) ([]*domain.Source, int, error)
	SaveSourcePrivateConfig(ctx context.Context, config *domain.SourcePrivateConfig) error
	GetSourcePrivateConfig(ctx context.Context, sourceID string) (*domain.SourcePrivateConfig, error)

	UpsertSourceCandidate(ctx context.Context, candidate *domain.SourceCandidate) error
	GetSourceCandidate(ctx context.Context, id string) (*domain.SourceCandidate, error)
	ListSourceCandidates(ctx context.Context, filter domain.SourceCandidateFilter) ([]*domain.SourceCandidate, int, error)
	TransitionSourceCandidate(ctx context.Context, request CandidateTransitionRequest) (*domain.SourceCandidate, error)

	SaveSourceRelationship(ctx context.Context, relationship *domain.SourceRelationship) error
	ListSourceRelationships(ctx context.Context, filter SourceRelationshipFilter) ([]*domain.SourceRelationship, int, error)

	SaveSourceVersion(ctx context.Context, version *domain.SourceVersion) error
	GetSourceVersion(ctx context.Context, id string) (*domain.SourceVersion, error)
	ListSourceVersions(ctx context.Context, filter SourceVersionFilter) ([]*domain.SourceVersion, int, error)
	GetLastKnownGoodSourceVersion(ctx context.Context, sourceID string) (*domain.SourceVersion, error)

	SaveSourceChange(ctx context.Context, change *domain.SourceChange) error
	ListSourceChanges(ctx context.Context, filter SourceChangeFilter) ([]*domain.SourceChange, int, error)

	SaveSourceHealthCheck(ctx context.Context, check *domain.SourceHealthCheck) error
	ListSourceHealthChecks(ctx context.Context, filter SourceHealthFilter) ([]*domain.SourceHealthCheck, int, error)
	GetLatestSourceHealthCheck(ctx context.Context, sourceID string) (*domain.SourceHealthCheck, error)

	SaveSourceCheckpoint(ctx context.Context, checkpoint *domain.SourceCheckpoint) error
	GetSourceCheckpoint(ctx context.Context, sourceID, streamKey string) (*domain.SourceCheckpoint, error)

	SaveMappingVersion(ctx context.Context, mapping *domain.MappingVersion) error
	GetMappingVersion(ctx context.Context, id string) (*domain.MappingVersion, error)
	ListMappingVersions(ctx context.Context, filter MappingVersionFilter) ([]*domain.MappingVersion, int, error)
	GetActiveMappingVersion(ctx context.Context, sourceID, name string) (*domain.MappingVersion, error)

	EnqueueIngestionJob(ctx context.Context, job *domain.IngestionJob) (*domain.IngestionJob, error)
	GetIngestionJob(ctx context.Context, id string) (*domain.IngestionJob, error)
	ListIngestionJobs(ctx context.Context, filter domain.IngestionJobFilter) ([]*domain.IngestionJob, int, error)
	ClaimIngestionJobs(ctx context.Context, request ClaimJobsRequest) ([]*domain.IngestionJob, error)
	CompleteIngestionJob(ctx context.Context, request CompleteJobRequest) (*domain.IngestionJob, error)
	FailIngestionJob(ctx context.Context, request FailJobRequest) (*domain.IngestionJob, error)
	ReplayDeadLetterJob(ctx context.Context, request ReplayJobRequest) (*domain.IngestionJob, error)

	SaveOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error
	ListOutboxEvents(ctx context.Context, status domain.OutboxStatus, limit, offset int) ([]*domain.OutboxEvent, int, error)
	ClaimOutboxEvents(ctx context.Context, request ClaimOutboxRequest) ([]*domain.OutboxEvent, error)
	CompleteOutboxEvent(ctx context.Context, request CompleteOutboxRequest) (*domain.OutboxEvent, error)
	FailOutboxEvent(ctx context.Context, request FailOutboxRequest) (*domain.OutboxEvent, error)

	CommitSourceObservation(ctx context.Context, commit SourceObservationCommit) error
	GetSourceCoverage(ctx context.Context, asOf time.Time) (*domain.SourceCoverage, error)
}
