package domain

import (
	"encoding/json"
	"time"
)

// SourceAuthorityTier classifies a source itself. It is deliberately separate
// from SourceTier: Tier 5 sources may be useful discovery leads, but cannot be
// attached to canonical factual claims as CEGS evidence without corroboration.
type SourceAuthorityTier int

const (
	SourceAuthorityTier1 SourceAuthorityTier = 1
	SourceAuthorityTier2 SourceAuthorityTier = 2
	SourceAuthorityTier3 SourceAuthorityTier = 3
	SourceAuthorityTier4 SourceAuthorityTier = 4
	SourceAuthorityTier5 SourceAuthorityTier = 5
)

func (tier SourceAuthorityTier) Valid() bool { return tier >= SourceAuthorityTier1 && tier <= SourceAuthorityTier5 }

// SourceLifecycleStatus is the review and activation lifecycle. It must never
// be conflated with SourceHealthStatus, which describes operational condition.
type SourceLifecycleStatus string

const (
	SourceLifecycleDiscovered SourceLifecycleStatus = "DISCOVERED"
	SourceLifecycleClassified SourceLifecycleStatus = "CLASSIFIED"
	SourceLifecycleTested     SourceLifecycleStatus = "TESTED"
	SourceLifecycleApproved   SourceLifecycleStatus = "APPROVED"
	SourceLifecycleActive     SourceLifecycleStatus = "ACTIVE"
	SourceLifecycleRejected   SourceLifecycleStatus = "REJECTED"
	SourceLifecycleBlocked    SourceLifecycleStatus = "BLOCKED"
	SourceLifecycleRetired    SourceLifecycleStatus = "RETIRED"
)

func (status SourceLifecycleStatus) Valid() bool {
	switch status {
	case SourceLifecycleDiscovered, SourceLifecycleClassified, SourceLifecycleTested,
		SourceLifecycleApproved, SourceLifecycleActive, SourceLifecycleRejected,
		SourceLifecycleBlocked, SourceLifecycleRetired:
		return true
	default:
		return false
	}
}

// CanTransitionSourceLifecycle encodes the candidate review gate. Recovery
// from BLOCKED/RETIRED is explicit; a source cannot silently jump from a newly
// discovered candidate to production activation.
func CanTransitionSourceLifecycle(from, to SourceLifecycleStatus) bool {
	if from == to {
		return from.Valid()
	}
	switch from {
	case SourceLifecycleDiscovered:
		return to == SourceLifecycleClassified || to == SourceLifecycleRejected || to == SourceLifecycleBlocked
	case SourceLifecycleClassified:
		return to == SourceLifecycleTested || to == SourceLifecycleRejected || to == SourceLifecycleBlocked
	case SourceLifecycleTested:
		return to == SourceLifecycleApproved || to == SourceLifecycleRejected || to == SourceLifecycleBlocked
	case SourceLifecycleApproved:
		return to == SourceLifecycleActive || to == SourceLifecycleRejected || to == SourceLifecycleBlocked
	case SourceLifecycleActive:
		return to == SourceLifecycleBlocked || to == SourceLifecycleRetired
	case SourceLifecycleBlocked:
		return to == SourceLifecycleClassified || to == SourceLifecycleTested || to == SourceLifecycleApproved ||
			to == SourceLifecycleActive || to == SourceLifecycleRejected || to == SourceLifecycleRetired
	case SourceLifecycleRejected:
		return to == SourceLifecycleDiscovered || to == SourceLifecycleClassified
	case SourceLifecycleRetired:
		return to == SourceLifecycleApproved || to == SourceLifecycleActive
	default:
		return false
	}
}

type SourceHealthStatus string

const (
	SourceHealthUnknown  SourceHealthStatus = "UNKNOWN"
	SourceHealthHealthy  SourceHealthStatus = "HEALTHY"
	SourceHealthStale    SourceHealthStatus = "STALE"
	SourceHealthDegraded SourceHealthStatus = "DEGRADED"
	SourceHealthBroken   SourceHealthStatus = "BROKEN"
	SourceHealthDisabled SourceHealthStatus = "DISABLED"
)

func (status SourceHealthStatus) Valid() bool {
	switch status {
	case SourceHealthUnknown, SourceHealthHealthy, SourceHealthStale,
		SourceHealthDegraded, SourceHealthBroken, SourceHealthDisabled:
		return true
	default:
		return false
	}
}

type SourceKind string

const (
	SourceKindCatalog            SourceKind = "CATALOG"
	SourceKindDataset            SourceKind = "DATASET"
	SourceKindResource           SourceKind = "RESOURCE"
	SourceKindAPI                SourceKind = "API"
	SourceKindFeed               SourceKind = "FEED"
	SourceKindDocumentRepository SourceKind = "DOCUMENT_REPOSITORY"
	SourceKindDocument           SourceKind = "DOCUMENT"
	SourceKindWebPage            SourceKind = "WEB_PAGE"
)

type SourceFamily string

const (
	SourceFamilyREST               SourceFamily = "REST_API"
	SourceFamilyGraphQL            SourceFamily = "GRAPHQL_API"
	SourceFamilyOpenAPI            SourceFamily = "OPENAPI"
	SourceFamilyCKAN               SourceFamily = "CKAN"
	SourceFamilySocrata            SourceFamily = "SOCRATA"
	SourceFamilyArcGIS             SourceFamily = "ARCGIS_REST"
	SourceFamilyArcGISFeature      SourceFamily = "ARCGIS_FEATURE_SERVER"
	SourceFamilyArcGISMap          SourceFamily = "ARCGIS_MAP_SERVER"
	SourceFamilyGeoJSON            SourceFamily = "GEOJSON"
	SourceFamilyWFS                SourceFamily = "WFS"
	SourceFamilyWMS                SourceFamily = "WMS"
	SourceFamilyWMTS               SourceFamily = "WMTS"
	SourceFamilySDMX               SourceFamily = "SDMX"
	SourceFamilyCSV                SourceFamily = "CSV"
	SourceFamilyTSV                SourceFamily = "TSV"
	SourceFamilySpreadsheet        SourceFamily = "XLS_XLSX"
	SourceFamilyJSON               SourceFamily = "JSON"
	SourceFamilyJSONL              SourceFamily = "JSONL"
	SourceFamilyXML                SourceFamily = "XML"
	SourceFamilyRSS                SourceFamily = "RSS"
	SourceFamilyAtom               SourceFamily = "ATOM"
	SourceFamilyDCAT               SourceFamily = "DCAT"
	SourceFamilyJSONLD             SourceFamily = "JSON_LD"
	SourceFamilyRDF                SourceFamily = "RDF"
	SourceFamilySPARQL             SourceFamily = "SPARQL"
	SourceFamilySitemap            SourceFamily = "SITEMAP"
	SourceFamilyHTML               SourceFamily = "HTML"
	SourceFamilyPDF                SourceFamily = "PDF"
	SourceFamilyDocumentRepository SourceFamily = "DOCUMENT_REPOSITORY"
	SourceFamilySearchPortal       SourceFamily = "SEARCH_PORTAL"
	SourceFamilyBulkZIP            SourceFamily = "BULK_ZIP"
	SourceFamilyGit                SourceFamily = "PUBLIC_GIT"
	SourceFamilyCloudObject        SourceFamily = "PUBLIC_CLOUD_OBJECT"
	SourceFamilyWebhook            SourceFamily = "WEBHOOK"
	SourceFamilyEmailFeed          SourceFamily = "EMAIL_FEED"
	SourceFamilyOther              SourceFamily = "OTHER"
)

type AccessPolicyStatus string

const (
	AccessPolicyUnknown       AccessPolicyStatus = "UNKNOWN"
	AccessPolicyAllowed       AccessPolicyStatus = "ALLOWED"
	AccessPolicyRestricted    AccessPolicyStatus = "RESTRICTED"
	AccessPolicyBlocked       AccessPolicyStatus = "BLOCKED"
	AccessPolicyNotApplicable AccessPolicyStatus = "NOT_APPLICABLE"
)

// Publisher is the public-source publisher identity. EntityID optionally links
// the publisher to the canonical economic graph without duplicating that graph.
type Publisher struct {
	ID              string              `json:"id"`
	EntityID        string              `json:"entity_id,omitempty"`
	Name            string              `json:"name"`
	CanonicalDomain string              `json:"canonical_domain,omitempty"`
	Jurisdiction    string              `json:"jurisdiction,omitempty"`
	AuthorityTier   SourceAuthorityTier `json:"authority_tier"`
	VerifiedAt      *time.Time           `json:"verified_at,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// PublisherPolicy centralizes polite-access controls shared by a publisher's
// sources. It contains policy, not credentials.
type PublisherPolicy struct {
	PublisherID               string    `json:"publisher_id"`
	MaxConcurrency            int       `json:"max_concurrency"`
	RequestsPerMinute         int       `json:"requests_per_minute"`
	Burst                     int       `json:"burst"`
	MinimumPollIntervalSeconds int64     `json:"minimum_poll_interval_seconds"`
	DefaultAttribution        string    `json:"default_attribution,omitempty"`
	DefaultLicense            string    `json:"default_license,omitempty"`
	AllowedSchemes            []string  `json:"allowed_schemes,omitempty"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// Source is the durable public-data census record. ParentSourceID models the
// efficient Publisher -> Catalog -> Dataset -> Resource hierarchy.
type Source struct {
	ID                    string                `json:"id"`
	PublisherID           string                `json:"publisher_id"`
	ParentSourceID        string                `json:"parent_source_id,omitempty"`
	Name                  string                `json:"name"`
	CanonicalURL          string                `json:"canonical_url"`
	Kind                  SourceKind            `json:"source_kind"`
	Jurisdiction          string                `json:"jurisdiction,omitempty"`
	Geographies           []string              `json:"geographies,omitempty"`
	Family                SourceFamily          `json:"source_family"`
	AccessMethod          string                `json:"access_method,omitempty"`
	ContentType           string                `json:"content_type,omitempty"`
	AuthorityTier         SourceAuthorityTier   `json:"authority_tier"`
	SubjectTags           []string              `json:"subject_tags,omitempty"`
	SectorTags            []string              `json:"sector_tags,omitempty"`
	Languages             []string              `json:"languages,omitempty"`
	UpdateFrequency       string                `json:"update_frequency,omitempty"`
	LifecycleStatus       SourceLifecycleStatus `json:"lifecycle_status"`
	HealthStatus          SourceHealthStatus    `json:"health_status"`
	TermsStatus           AccessPolicyStatus    `json:"terms_status"`
	RobotsStatus          AccessPolicyStatus    `json:"robots_status"`
	License               string                `json:"license,omitempty"`
	AuthenticationRequired bool                  `json:"authentication_required"`
	Parser                string                `json:"parser,omitempty"`
	Adapter               string                `json:"adapter,omitempty"`
	PollingPolicy         string                `json:"polling_policy,omitempty"`
	IncrementalCapability string                `json:"incremental_capability,omitempty"`
	HistoricalDepth       string                `json:"historical_depth,omitempty"`
	EstimatedValue        float64               `json:"estimated_value,omitempty"`
	EstimatedCost         float64               `json:"estimated_cost,omitempty"`
	CoverageScore         float64               `json:"coverage_score,omitempty"`
	Capabilities          map[string]interface{} `json:"capabilities,omitempty"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
	DiscoveredAt          time.Time              `json:"discovered_at"`
	RegisteredAt          *time.Time             `json:"registered_at,omitempty"`
	LastCheckedAt         *time.Time             `json:"last_checked_at,omitempty"`
	LastSuccessAt         *time.Time             `json:"last_success_at,omitempty"`
	LastChangeAt          *time.Time             `json:"last_change_at,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// SourcePrivateConfig is persisted separately and must not be serialized into
// public source APIs or exports. CredentialRefs are names/paths, never secrets.
type SourcePrivateConfig struct {
	SourceID       string                 `json:"source_id"`
	AdapterConfig  map[string]interface{} `json:"-"`
	CredentialRefs map[string]string      `json:"-"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// SourceCandidate retains discovery provenance and human-review state for a
// Source. LifecycleStatus is mirrored to Source by the store on transition.
type SourceCandidate struct {
	ID                   string                `json:"id"`
	SourceID             string                `json:"source_id"`
	DiscoveredBySourceID string                `json:"discovered_by_source_id,omitempty"`
	DiscoveryJobID       string                `json:"discovery_job_id,omitempty"`
	DiscoveryURL         string                `json:"discovery_url"`
	LifecycleStatus      SourceLifecycleStatus `json:"lifecycle_status"`
	AuthorityScore       float64               `json:"authority_score,omitempty"`
	EconomicRelevance    float64               `json:"economic_relevance,omitempty"`
	UniquenessScore      float64               `json:"uniqueness_score,omitempty"`
	StructurednessScore  float64               `json:"structuredness_score,omitempty"`
	HistoricalDepthScore float64               `json:"historical_depth_score,omitempty"`
	EstimatedValue       float64               `json:"estimated_value,omitempty"`
	EstimatedCost        float64               `json:"estimated_cost,omitempty"`
	CoverageGain         float64               `json:"coverage_gain,omitempty"`
	ReviewRequired       bool                  `json:"review_required"`
	ReviewedBy           string                `json:"reviewed_by,omitempty"`
	ReviewedAt           *time.Time             `json:"reviewed_at,omitempty"`
	DecisionReason       string                `json:"decision_reason,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

type SourceRelationshipType string

const (
	SourceRelationshipContains          SourceRelationshipType = "CONTAINS"
	SourceRelationshipReferences        SourceRelationshipType = "REFERENCES"
	SourceRelationshipMirrors           SourceRelationshipType = "MIRRORS"
	SourceRelationshipSupersedes        SourceRelationshipType = "SUPERSEDES"
	SourceRelationshipSiblingLanguage   SourceRelationshipType = "SIBLING_LANGUAGE"
	SourceRelationshipDerivedFrom       SourceRelationshipType = "DERIVED_FROM"
)

type SourceRelationship struct {
	ID               string                 `json:"id"`
	FromSourceID     string                 `json:"from_source_id"`
	ToSourceID       string                 `json:"to_source_id"`
	RelationshipType SourceRelationshipType `json:"relationship_type"`
	EvidenceID       string                 `json:"evidence_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ValidFrom        *time.Time             `json:"valid_from,omitempty"`
	ValidTo          *time.Time             `json:"valid_to,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

type SnapshotStorageMode string

const (
	SnapshotStorageFull     SnapshotStorageMode = "FULL"
	SnapshotStorageMetadata SnapshotStorageMode = "METADATA_ONLY"
	SnapshotStorageHash     SnapshotStorageMode = "HASH_ONLY"
	SnapshotStorageLocator  SnapshotStorageMode = "URL_LOCATOR_ONLY"
)

type SourceSnapshotMetadata struct {
	StorageMode      SnapshotStorageMode `json:"storage_mode"`
	ObjectKey        string              `json:"object_key,omitempty"`
	Locator          string              `json:"locator,omitempty"`
	SizeBytes        int64               `json:"size_bytes,omitempty"`
	MediaType        string              `json:"media_type,omitempty"`
	Charset          string              `json:"charset,omitempty"`
	ContentEncoding  string              `json:"content_encoding,omitempty"`
	SourceProjection string              `json:"source_projection,omitempty"`
	License          string              `json:"license,omitempty"`
	RetainUntil      *time.Time           `json:"retain_until,omitempty"`
}

// SourceVersion is an immutable, changed representation of a source. Routine
// unchanged checks are health records and do not manufacture new versions.
type SourceVersion struct {
	ID                 string                 `json:"id"`
	SourceID           string                 `json:"source_id"`
	Sequence           int64                  `json:"sequence"`
	ObservedAt         time.Time              `json:"observed_at"`
	RetrievedAt        time.Time              `json:"retrieved_at"`
	PublishedAt        *time.Time             `json:"published_at,omitempty"`
	EffectiveAt        *time.Time             `json:"effective_at,omitempty"`
	RemoteModifiedAt   *time.Time             `json:"remote_modified_at,omitempty"`
	ContentHash        string                 `json:"content_hash"`
	SemanticHash       string                 `json:"semantic_hash,omitempty"`
	ETag               string                 `json:"etag,omitempty"`
	LastModified       string                 `json:"last_modified,omitempty"`
	SchemaFingerprint  string                 `json:"schema_fingerprint,omitempty"`
	ParserVersion      string                 `json:"parser_version"`
	MappingVersion     string                 `json:"mapping_version"`
	PipelineVersion    string                 `json:"pipeline_version"`
	LastKnownGood      bool                   `json:"last_known_good"`
	Snapshot           SourceSnapshotMetadata `json:"snapshot"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
}

type SourceChangeType string

const (
	SourceChangeDiscovered       SourceChangeType = "SOURCE_DISCOVERED"
	SourceChangeChanged          SourceChangeType = "SOURCE_CHANGED"
	SourceChangeMoved            SourceChangeType = "SOURCE_MOVED"
	SourceChangeBroken           SourceChangeType = "SOURCE_BROKEN"
	SourceChangeRestored         SourceChangeType = "SOURCE_RESTORED"
	SourceChangeRetired          SourceChangeType = "SOURCE_RETIRED"
	SourceChangeFormatChanged    SourceChangeType = "SOURCE_FORMAT_CHANGED"
	SourceChangeLicenseChanged   SourceChangeType = "SOURCE_LICENSE_CHANGED"
	SourceChangeSchemaDrift      SourceChangeType = "SOURCE_SCHEMA_DRIFT"
	SourceChangeProjectStage     SourceChangeType = "PROJECT_STAGE_CHANGED"
	SourceChangeCAPEX            SourceChangeType = "CAPEX_CHANGED"
	SourceChangeDate             SourceChangeType = "DATE_CHANGED"
	SourceChangeFinancing        SourceChangeType = "FINANCING_CHANGED"
	SourceChangePermitStatus     SourceChangeType = "PERMIT_STATUS_CHANGED"
	SourceChangeProcurement      SourceChangeType = "PROCUREMENT_CHANGED"
	SourceChangeOrganization     SourceChangeType = "ORGANIZATION_CHANGED"
)

type Materiality string

const (
	MaterialityTrivial  Materiality = "TRIVIAL"
	MaterialityMinor    Materiality = "MINOR"
	MaterialityMaterial Materiality = "MATERIAL"
	MaterialityMajor    Materiality = "MAJOR"
)

type FieldChange struct {
	Code     string      `json:"code"`
	Path     string      `json:"path"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

type SourceChange struct {
	ID               string                 `json:"id"`
	SourceID         string                 `json:"source_id"`
	ResourceID       string                 `json:"resource_id,omitempty"`
	FromVersionID    string                 `json:"from_version_id,omitempty"`
	ToVersionID      string                 `json:"to_version_id"`
	ChangeType       SourceChangeType       `json:"change_type"`
	Materiality      Materiality            `json:"materiality"`
	ObservedAt       time.Time              `json:"observed_at"`
	EffectiveAt      *time.Time             `json:"effective_at,omitempty"`
	ChangedFields    []FieldChange          `json:"changed_fields,omitempty"`
	RawDiff          json.RawMessage        `json:"raw_diff,omitempty"`
	SemanticSummary  string                 `json:"semantic_summary,omitempty"`
	Fingerprint      string                 `json:"fingerprint"`
	PipelineVersion  string                 `json:"pipeline_version"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

type SourceHealthCheck struct {
	ID                  string             `json:"id"`
	SourceID            string             `json:"source_id"`
	JobID               string             `json:"job_id,omitempty"`
	CheckedAt           time.Time          `json:"checked_at"`
	Status              SourceHealthStatus `json:"status"`
	HTTPStatus          int                `json:"http_status,omitempty"`
	LatencyMilliseconds int64              `json:"latency_ms,omitempty"`
	DocumentsSeen       int64              `json:"documents_seen,omitempty"`
	DocumentsChanged    int64              `json:"documents_changed,omitempty"`
	BytesReceived       int64              `json:"bytes_received,omitempty"`
	ParseFailures       int64              `json:"parse_failures,omitempty"`
	ConsecutiveFailures int                `json:"consecutive_failures,omitempty"`
	RateLimitState      string             `json:"rate_limit_state,omitempty"`
	ErrorCode           string             `json:"error_code,omitempty"`
	ErrorMessage        string             `json:"error_message,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
}

type SourceCheckpoint struct {
	SourceID          string                 `json:"source_id"`
	StreamKey         string                 `json:"stream_key"`
	Cursor            string                 `json:"cursor,omitempty"`
	ETag              string                 `json:"etag,omitempty"`
	LastModified      string                 `json:"last_modified,omitempty"`
	LastSeenRemoteAt  *time.Time             `json:"last_seen_remote_at,omitempty"`
	LastProcessedAt   *time.Time             `json:"last_processed_at,omitempty"`
	LastVersionID     string                 `json:"last_version_id,omitempty"`
	OpaqueState       map[string]interface{} `json:"opaque_state,omitempty"`
	Revision          int64                  `json:"revision"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type MappingStatus string

const (
	MappingDraft    MappingStatus = "DRAFT"
	MappingActive   MappingStatus = "ACTIVE"
	MappingRetired  MappingStatus = "RETIRED"
	MappingRejected MappingStatus = "REJECTED"
)

type MappingVersion struct {
	ID                 string          `json:"id"`
	SourceID           string          `json:"source_id"`
	Name               string          `json:"name"`
	Version            string          `json:"version"`
	TargetResourceType string          `json:"target_resource_type"`
	ParserVersion      string          `json:"parser_version,omitempty"`
	InputSchemaHash    string          `json:"input_schema_hash,omitempty"`
	DefinitionHash     string          `json:"definition_hash"`
	Definition         json.RawMessage `json:"definition,omitempty"`
	Status             MappingStatus   `json:"status"`
	ActivatedAt        *time.Time      `json:"activated_at,omitempty"`
	RetiredAt          *time.Time      `json:"retired_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type IngestionMode string

const (
	IngestionModeDiscovery IngestionMode = "DISCOVERY"
	IngestionModeLive      IngestionMode = "LIVE"
	IngestionModeBackfill  IngestionMode = "BACKFILL"
	IngestionModeReplay    IngestionMode = "REPLAY"
)

type IngestionQueue string

const (
	IngestionQueueDiscovery          IngestionQueue = "DISCOVERY"
	IngestionQueueLive               IngestionQueue = "LIVE"
	IngestionQueueBackfill           IngestionQueue = "BACKFILL"
	IngestionQueueDocumentExtraction IngestionQueue = "DOCUMENT_EXTRACTION"
	IngestionQueueNormalization      IngestionQueue = "NORMALIZATION"
	IngestionQueueReconciliation     IngestionQueue = "RECONCILIATION"
)

const (
	IngestionPriorityP0 = 0
	IngestionPriorityP1 = 1
	IngestionPriorityP2 = 2
	IngestionPriorityP3 = 3
	IngestionPriorityP4 = 4
)

type IngestionJobStatus string

const (
	IngestionJobQueued     IngestionJobStatus = "QUEUED"
	IngestionJobRunning    IngestionJobStatus = "RUNNING"
	IngestionJobRetry      IngestionJobStatus = "RETRY"
	IngestionJobSucceeded  IngestionJobStatus = "SUCCEEDED"
	IngestionJobDeadLetter IngestionJobStatus = "DEAD_LETTER"
	IngestionJobCancelled  IngestionJobStatus = "CANCELLED"
)

type IngestionJob struct {
	ID                  string                 `json:"id"`
	SourceID            string                 `json:"source_id"`
	ResourceID          string                 `json:"resource_id,omitempty"`
	Queue               IngestionQueue         `json:"queue"`
	Mode                IngestionMode          `json:"mode"`
	Priority            int                    `json:"priority"`
	Status              IngestionJobStatus     `json:"status"`
	DedupeKey           string                 `json:"dedupe_key"`
	AvailableAt         time.Time              `json:"available_at"`
	LeaseOwner          string                 `json:"lease_owner,omitempty"`
	LeaseExpiresAt      *time.Time             `json:"lease_expires_at,omitempty"`
	AttemptCount        int                    `json:"attempt_count"`
	MaxAttempts         int                    `json:"max_attempts"`
	FailureStage        string                 `json:"failure_stage,omitempty"`
	LastError           string                 `json:"last_error,omitempty"`
	PipelineVersion     string                 `json:"pipeline_version"`
	Payload             map[string]interface{} `json:"payload,omitempty"`
	ReplayOfJobID       string                 `json:"replay_of_job_id,omitempty"`
	StartedAt           *time.Time             `json:"started_at,omitempty"`
	CompletedAt         *time.Time             `json:"completed_at,omitempty"`
	DeadLetteredAt      *time.Time             `json:"dead_lettered_at,omitempty"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

type OutboxStatus string

const (
	OutboxPending   OutboxStatus = "PENDING"
	OutboxPublishing OutboxStatus = "PUBLISHING"
	OutboxPublished OutboxStatus = "PUBLISHED"
	OutboxFailed    OutboxStatus = "FAILED"
)

// OutboxEvent is the durable internal event shape. It is not itself a CEGS
// economic event; consumers may project appropriate source changes into CEGS.
type OutboxEvent struct {
	ID                string                 `json:"id"`
	SourceID          string                 `json:"source_id"`
	ResourceID        string                 `json:"resource_id,omitempty"`
	EventType         string                 `json:"event_type"`
	ObservedAt        time.Time              `json:"observed_at"`
	EffectiveAt       *time.Time             `json:"effective_at,omitempty"`
	Payload           map[string]interface{} `json:"payload"`
	Hash              string                 `json:"hash"`
	PipelineVersion   string                 `json:"pipeline_version"`
	Status            OutboxStatus           `json:"status"`
	AttemptCount      int                    `json:"attempt_count"`
	AvailableAt       time.Time              `json:"available_at"`
	LeaseOwner        string                 `json:"lease_owner,omitempty"`
	LeaseExpiresAt    *time.Time             `json:"lease_expires_at,omitempty"`
	PublishedAt       *time.Time             `json:"published_at,omitempty"`
	LastError         string                 `json:"last_error,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type SourceFilter struct {
	PublisherID     string
	ParentSourceID  string
	Jurisdiction    string
	Sector          string
	Language        string
	Family          SourceFamily
	Kind            SourceKind
	LifecycleStatus SourceLifecycleStatus
	HealthStatus    SourceHealthStatus
	AuthorityTier   SourceAuthorityTier
	Search          string
	Limit           int
	Offset          int
}

type SourceCandidateFilter struct {
	SourceID        string
	DiscoveredByID  string
	LifecycleStatus SourceLifecycleStatus
	ReviewRequired  *bool
	Limit           int
	Offset          int
}

type IngestionJobFilter struct {
	SourceID string
	Queue    IngestionQueue
	Mode     IngestionMode
	Status   IngestionJobStatus
	Limit    int
	Offset   int
}

// SourceCoverageCount contains only observed counts. It intentionally avoids a
// guessed denominator or synthetic precision.
type SourceCoverageCount struct {
	Key        string `json:"key"`
	Registered int    `json:"registered"`
	Tested     int    `json:"tested"`
	Active     int    `json:"active"`
	Broken     int    `json:"broken"`
}

type SourceCoverage struct {
	GeneratedAt          time.Time             `json:"generated_at"`
	TotalSources         int                   `json:"total_sources"`
	Discovered           int                   `json:"discovered"`
	Classified           int                   `json:"classified"`
	Tested               int                   `json:"tested"`
	Approved             int                   `json:"approved"`
	Active               int                   `json:"active"`
	Rejected             int                   `json:"rejected"`
	Blocked              int                   `json:"blocked"`
	Retired              int                   `json:"retired"`
	Healthy              int                   `json:"healthy"`
	Stale                int                   `json:"stale"`
	Degraded             int                   `json:"degraded"`
	Broken               int                   `json:"broken"`
	Disabled             int                   `json:"disabled"`
	UnknownHealth        int                   `json:"unknown_health"`
	PrimaryActiveSources int                   `json:"primary_active_sources"`
	PrimarySourceRatio   *float64              `json:"primary_source_ratio,omitempty"`
	ByJurisdiction       []SourceCoverageCount `json:"by_jurisdiction"`
	BySector             []SourceCoverageCount `json:"by_sector"`
	ByFamily             []SourceCoverageCount `json:"by_family"`
}
