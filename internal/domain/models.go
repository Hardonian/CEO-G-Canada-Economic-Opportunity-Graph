package domain

import (
	"time"
)

// Evidence retains complete factual provenance. Every fact must point to an Evidence record.
type Evidence struct {
	ID                 string          `json:"id"`
	SourceURL          string          `json:"source_url"`
	Publisher          string          `json:"publisher"`
	SourceTier         SourceTier      `json:"source_tier"`
	RetrievalTimestamp time.Time       `json:"retrieval_timestamp"`
	PublicationDate    *time.Time      `json:"publication_date,omitempty"`
	EffectiveDate      *time.Time      `json:"effective_date,omitempty"`
	Confidence         ConfidenceLevel `json:"confidence"`
	ExtractionMethod   string          `json:"extraction_method"` // e.g., "deterministic_adapter", "official_api"
	ContentHash        string          `json:"content_hash"`      // SHA-256
	HashScope          string          `json:"hash_scope,omitempty"` // raw_document or normalized_source_record
	SourceClass        string          `json:"source_class,omitempty"`
	SourceRecordID     string          `json:"source_record_id,omitempty"`
	Locator            string          `json:"locator,omitempty"`
	PipelineVersion    string          `json:"pipeline_version,omitempty"`
	ParserVersion      string          `json:"parser_version,omitempty"`
	RawSnippet         string          `json:"raw_snippet,omitempty"`
}

// Entity represents an organization (company, government body, First Nation, regulator, investor, supplier).
type Entity struct {
	ID             string            `json:"id"`
	Slug           string            `json:"slug"`
	LegalName      string            `json:"legal_name"`
	CommonName     string            `json:"common_name"`
	Aliases        []string          `json:"aliases"`
	EntityType     string            `json:"entity_type"` // Corporation, CrownCorp, FirstNation, GovernmentAgency, Investor, Utility, Supplier
	Jurisdiction   string            `json:"jurisdiction"` // CA, ON, BC, QC, AB, etc.
	Website        string            `json:"website,omitempty"`
	Identifiers    map[string]string `json:"identifiers,omitempty"` // NEBN, Ticker, LEI
	Description    string            `json:"description,omitempty"`
	AISovereignty  *AISovereignty    `json:"ai_sovereignty,omitempty"`
	EvidenceID     string            `json:"evidence_id,omitempty"`
	Evidence       *Evidence         `json:"evidence,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// Project is the central economic graph object.
type Project struct {
	ID                   string                 `json:"id"`
	Slug                 string                 `json:"slug"`
	Name                 string                 `json:"name"`
	Summary              string                 `json:"summary"`
	Sector               Sector                 `json:"sector"`
	Subsector            string                 `json:"subsector"`
	Province             string                 `json:"province"` // e.g., "ON", "BC", "AB", "QC", "SK", "MB", "NL", "NS", "NB", "PE", "YT", "NT", "NU", "Federal"
	LocationName         string                 `json:"location_name"`
	Latitude             float64                `json:"latitude"`
	Longitude            float64                `json:"longitude"`
	CurrentStage         LifecycleStage         `json:"current_stage"`
	CapexCAD             int64                  `json:"capex_cad"` // In CAD cents or whole dollars; we use whole CAD
	CapexStatus          ConfidenceLevel        `json:"capex_status"`
	ProponentID          string                 `json:"proponent_id"`
	Proponent            *Entity                `json:"proponent,omitempty"`
	Confidence           ConfidenceLevel        `json:"confidence"`
	EvidenceIDs          []string               `json:"evidence_ids,omitempty"`
	ExternalIDs          map[string]string      `json:"external_ids,omitempty"`
	IsSynthetic          bool                   `json:"is_synthetic"` // Clearly tags DEMO data
	LastMeaningfulUpdate time.Time              `json:"last_meaningful_update"`
	Scores               map[string]float64     `json:"scores,omitempty"` // buildability, investability, supplierability, strategicity
	ScoreDetails         []*ProjectScore        `json:"score_details,omitempty"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// Event records append-only historical transitions and milestones.
type Event struct {
	ID            string          `json:"id"`
	ProjectID     string          `json:"project_id"`
	EventType     string          `json:"event_type"` // stage_change, financing_announced, regulatory_filing, indigenous_agreement, contract_awarded
	EventDate     time.Time       `json:"event_date"`
	PreviousStage *LifecycleStage `json:"previous_stage,omitempty"`
	NewStage      *LifecycleStage `json:"new_stage,omitempty"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	EvidenceID    string          `json:"evidence_id"`
	Evidence      *Evidence       `json:"evidence,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// Relationship models typed connections between entities and projects.
type Relationship struct {
	ID             string          `json:"id"`
	ProjectID      string          `json:"project_id"`
	SourceEntityID string          `json:"source_entity_id"`
	TargetEntityID string          `json:"target_entity_id"`
	SourceEntity   *Entity         `json:"source_entity,omitempty"`
	TargetEntity   *Entity         `json:"target_entity,omitempty"`
	RelationType   string          `json:"relation_type"` // proponent, investor, lender, regulator, indigenous_partner, supplier, offtaker
	Confidence     ConfidenceLevel `json:"confidence"`
	EvidenceID     string          `json:"evidence_id"`
	Evidence       *Evidence       `json:"evidence,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	ValidFrom      *time.Time      `json:"valid_from,omitempty"`
	ValidTo        *time.Time      `json:"valid_to,omitempty"`
}

// ProjectScore represents a versioned, deterministic score with factor breakdown.
type ProjectScore struct {
	ID           string             `json:"id"`
	ProjectID    string             `json:"project_id"`
	ScoreType    string             `json:"score_type"` // buildability, investability, supplierability, strategicity
	ScoreValue   float64            `json:"score_value"` // 0-100
	ScoreVersion string             `json:"score_version"` // e.g. "buildability-v1.0"
	Factors      map[string]float64 `json:"factors"`
	UnknownFactors []string         `json:"unknown_factors,omitempty"`
	Coverage       float64          `json:"coverage"`
	Confidence     ConfidenceLevel  `json:"confidence"`
	InputHash      string           `json:"input_hash"`
	PreviousValue  *float64         `json:"previous_value,omitempty"`
	Movement       *float64         `json:"movement,omitempty"`
	MovementReasons []string        `json:"movement_reasons,omitempty"`
	Explanation  string             `json:"explanation"`
	CalculatedAt time.Time          `json:"calculated_at"`
}

// CapitalItem records categorized, non-blended capital events.
type CapitalItem struct {
	ID               string          `json:"id"`
	ProjectID        string          `json:"project_id"`
	Category         CapitalCategory `json:"category"`
	Status           CapitalStatus   `json:"status"`
	AmountCAD        int64           `json:"amount_cad"`
	AmountType       string          `json:"amount_type"` // exact, maximum, estimated, unknown
	ProviderEntityID string          `json:"provider_entity_id,omitempty"`
	ProviderName     string          `json:"provider_name"`
	Notes            string          `json:"notes,omitempty"`
	EvidenceID       string          `json:"evidence_id"`
	Evidence         *Evidence       `json:"evidence,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

// Procurement records public and private tenders and requests.
type Procurement struct {
	ID               string           `json:"id"`
	TenderID         string           `json:"tender_id"`
	ProjectID        string           `json:"project_id,omitempty"` // Linked parent project if verified/derived
	ProjectName      string           `json:"project_name,omitempty"`
	Title            string           `json:"title"`
	Stage            string           `json:"stage"` // RFI, RFQ, RFP, Standing Offer, Awarded
	ClosingDate      *time.Time       `json:"closing_date,omitempty"`
	EstimatedCAD     int64            `json:"estimated_cad,omitempty"`
	Buyer            string           `json:"buyer"`
	BuyerType        string           `json:"buyer_type"` // Federal, Provincial, Crown, Municipal, Private
	SourceURL        string           `json:"source_url"`
	Categories       []string         `json:"categories"`
	RequirementClass RequirementClass `json:"requirement_class"` // CONFIRMED, DERIVED
	EvidenceID       string           `json:"evidence_id"`
	Evidence         *Evidence        `json:"evidence,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

// Opportunity represents an inferred or confirmed downstream demand.
type Opportunity struct {
	ID               string           `json:"id"`
	ProjectID        string           `json:"project_id"`
	ProjectName      string           `json:"project_name"`
	Title            string           `json:"title"`
	Sector           Sector           `json:"sector"`
	RequirementClass RequirementClass `json:"requirement_class"` // CONFIRMED, DERIVED, SPECULATIVE
	Category         string           `json:"category"` // engineering, electrical, environmental, etc.
	EstimatedCAD     int64            `json:"estimated_cad,omitempty"`
	EstimateStatus   ConfidenceLevel  `json:"estimate_status"`
	Description      string           `json:"description"`
	TriggerMilestone string           `json:"trigger_milestone"` // e.g., "FID", "ENVIRONMENTAL_APPROVAL"
	CreatedAt        time.Time        `json:"created_at"`
}

// Signal records calculated inflection points.
type Signal struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	ProjectName   string     `json:"project_name"`
	Type          SignalType `json:"type"`
	Timestamp     time.Time  `json:"timestamp"`
	Magnitude     float64    `json:"magnitude"` // 0.0 - 1.0
	Confidence    float64    `json:"confidence"` // 0.0 - 1.0
	PreviousState string     `json:"previous_state,omitempty"`
	NewState      string     `json:"new_state,omitempty"`
	Description   string     `json:"description"`
	EvidenceID    string     `json:"evidence_id,omitempty"`
}

// AISovereignty models evaluation criteria for the Canadian AI Sovereignty Index.
type AISovereignty struct {
	ScoreVersion      string             `json:"score_version"`
	OverallScore      float64            `json:"overall_score"` // 0-100
	DataResidency     float64            `json:"data_residency"` // 0-10
	ComputeResidency  float64            `json:"compute_residency"`
	CanadianOwnership float64            `json:"canadian_ownership"`
	ForeignLegalRisk  float64            `json:"foreign_legal_risk"` // Higher means lower legal vulnerability (e.g. CLOUD Act protection)
	LocalDeployment   float64            `json:"local_deployment"`
	BilingualCapacity float64            `json:"bilingual_capacity"`
	QuebecLaw25       float64            `json:"quebec_law_25"`
	CleanEnergy       float64            `json:"clean_energy"`
	Dimensions        map[string]float64 `json:"dimensions,omitempty"`
	Notes             string             `json:"notes"`
	CalculatedAt      time.Time          `json:"calculated_at"`
}
