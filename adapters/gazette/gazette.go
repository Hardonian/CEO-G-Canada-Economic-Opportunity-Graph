package gazette

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

const (
	adapterName         = "provincial_gazettes"
	pipelineVersion     = "gazette-v1"
	parserVersion       = "gazette-json-v1"
	maxResponseBytes    int64 = 8 << 20
	
	// Official gazette endpoints
	OntarioGazetteURL   = "https://www.ontario.ca/page/ontario-gazette"
	QuebecGazetteURL    = "https://gazetteofficielle.gouv.qc.ca/"
	BCGazetteURL        = "https://www.bcgazette.ca/"
	FederalGazetteURL   = "https://gazette.gc.ca/"
)

var (
	// Gazette notice types that indicate project relevance
	projectKeywords = []string{
		"environmental assessment", "major project", "infrastructure",
		"procurement", "tender", "contract award", "permit",
		"approval", "license", "authorization", "regulatory",
		"construction", "development", "investment",
	}
	
	// Regex patterns for capital amounts
	capexPattern = regexp.MustCompile(`(?i)(?:CAD|C\$|\$)\s*([\d,]+(?:\.\d+)?)\s*(?:million|billion|M|B)`)
)

type GazetteAdapter struct {
	province   string
	endpoints  []string
	client     *http.Client
	fixturePath string
	fetchedAt  time.Time
	health     adapters.SourceHealth
}

type GazetteRecord struct {
	NoticeID      string                 `json:"notice_id"`
	Title         string                 `json:"title"`
	Content       string                 `json:"content"`
	PublishDate   string                 `json:"publish_date"`
	SourceURL     string                 `json:"source_url"`
	Category      string                 `json:"category"`
	ProponentName string                 `json:"proponent_name,omitempty"`
	ProjectName   string                 `json:"project_name,omitempty"`
	Location      string                 `json:"location,omitempty"`
	AmountText    string                 `json:"amount_text,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type GazetteFixture struct {
	Source        string         `json:"source"`
	SourceURL     string         `json:"source_url"`
	RetrievedAt   string         `json:"retrieved_at"`
	EffectiveAt   string         `json:"effective_at"`
	DatasetVintage string        `json:"dataset_vintage"`
	Notices       []GazetteRecord `json:"notices"`
}

// NewGazetteAdapter creates a curated snapshot adapter for provincial gazettes.
func NewGazetteAdapter(province, fixturePath string) *GazetteAdapter {
	if fixturePath == "" {
		fixturePath = fmt.Sprintf("data/fixtures/gazette_%s.json", strings.ToLower(province))
	}
	
	var endpoints []string
	var sourceTier domain.SourceTier = domain.SourceTier1
	
	switch strings.ToLower(province) {
	case "on", "ontario":
		endpoints = []string{OntarioGazetteURL}
	case "qc", "quebec":
		endpoints = []string{QuebecGazetteURL}
	case "bc", "british columbia":
		endpoints = []string{BCGazetteURL}
	case "ca", "federal", "canada":
		endpoints = []string{FederalGazetteURL}
		sourceTier = domain.SourceTier1
	default:
		endpoints = []string{}
	}
	
	return &GazetteAdapter{
		province:    strings.ToUpper(province),
		endpoints:   endpoints,
		fixturePath: fixturePath,
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		health: adapters.SourceHealth{
			AdapterName:    fmt.Sprintf("%s_%s", adapterName, strings.ToLower(province)),
			Tier:           sourceTier,
			Status:         string(domain.StatusHealthy),
			RateLimitState: "NOT_APPLICABLE",
			Mode:           "CURATED_SNAPSHOT",
		},
	}
}

// NewLiveGazetteAdapter creates an opt-in live adapter (future implementation).
func NewLiveGazetteAdapter(province string, client *http.Client) (*GazetteAdapter, error) {
	return &GazetteAdapter{
		province: strings.ToUpper(province),
		client:   client,
		health: adapters.SourceHealth{
			AdapterName:    fmt.Sprintf("%s_%s", adapterName, strings.ToLower(province)),
			Tier:           domain.SourceTier1,
			Status:         string(domain.StatusHealthy),
			RateLimitState: "AVAILABLE",
			Mode:           "LIVE",
		},
	}, nil
}

func (a *GazetteAdapter) Name() string                   { return fmt.Sprintf("%s_%s", adapterName, strings.ToLower(a.province)) }
func (a *GazetteAdapter) Tier() domain.SourceTier        { return a.health.Tier }
func (a *GazetteAdapter) Health() *adapters.SourceHealth { return &a.health }

func (a *GazetteAdapter) Fetch(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	
	a.health.LastAttempt = time.Now().UTC()
	
	if a.endpoints != nil && len(a.endpoints) > 0 && a.fixturePath == "" {
		// Live mode - not fully implemented, return fixture instead
		// TODO: Implement live scraping for gazette APIs
	}
	
	data, err := adapters.ReadBoundedFile(a.fixturePath, maxResponseBytes)
	if err != nil {
		a.fail(err)
		return nil, fmt.Errorf("read gazette snapshot: %w", err)
	}
	
	a.fetchedAt = time.Now().UTC()
	a.health.LastSuccess = a.fetchedAt
	a.health.Status = string(domain.StatusHealthy)
	a.health.LastError = ""
	return data, nil
}

func (a *GazetteAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, a.parseError(fmt.Errorf("empty gazette document"))
	}
	
	var fixture GazetteFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, a.parseError(fmt.Errorf("parse gazette JSON: %w", err))
	}
	
	if len(fixture.Notices) == 0 {
		return nil, a.parseError(fmt.Errorf("gazette contains no notices"))
	}
	
	if len(fixture.Notices) > 5000 {
		return nil, a.parseError(fmt.Errorf("gazette notice count exceeds 5000-record safety limit"))
	}
	
	result := &adapters.IngestionResult{}
	seen := make(map[string]struct{}, len(fixture.Notices))
	
	retrieved, err := parseTimestamp(fixture.RetrievedAt)
	if err != nil {
		retrieved = time.Now().UTC()
	}
	
	effective, err := parseTimestamp(fixture.EffectiveAt)
	if err != nil {
		effective = retrieved
	}
	
	sourceURL := fixture.SourceURL
	if sourceURL == "" {
		if len(a.endpoints) > 0 {
			sourceURL = a.endpoints[0]
		}
	}
	
	publisher := gazettePublisher(a.province)
	
	for _, notice := range fixture.Notices {
		if strings.TrimSpace(notice.NoticeID) == "" || strings.TrimSpace(notice.Title) == "" {
			continue
		}
		
		if _, exists := seen[notice.NoticeID]; exists {
			continue
		}
		seen[notice.NoticeID] = struct{}{}
		
		// Filter for project-relevant notices
		if !isProjectRelevant(notice) {
			continue
		}
		
		featureHash, hashErr := adapters.HashRecord(notice)
		if hashErr != nil {
			return nil, a.parseError(hashErr)
		}
		
		noticeID := fmt.Sprintf("gazette_%s_%s", strings.ToLower(a.province), notice.NoticeID)
		projectID := identity.StableID("project", adapterName, noticeID)
		evidenceID := identity.StableID("evidence", adapterName, noticeID+":"+featureHash)
		
		entityID := ""
		var entity *domain.Entity
		if strings.TrimSpace(notice.ProponentName) != "" {
			entityID = identity.StableID("entity", "ca", notice.ProponentName)
			entity = &domain.Entity{
				ID:           entityID,
				Slug:         identity.Slug(notice.ProponentName),
				LegalName:    notice.ProponentName,
				CommonName:   notice.ProponentName,
				Aliases:      []string{},
				EntityType:   "ProjectProponent",
				Jurisdiction: "CA:" + a.province,
				EvidenceID:   evidenceID,
				CreatedAt:    effective,
				UpdatedAt:    effective,
			}
		}
		
		evidence := &domain.Evidence{
			ID:                 evidenceID,
			SourceURL:          notice.SourceURL,
			Publisher:          publisher,
			SourceTier:         domain.SourceTier1,
			Visibility:         domain.VisibilityPublicAttribution,
			Publishable:        true,
			RetrievalTimestamp: retrieved,
			EffectiveDate:      &effective,
			Confidence:         domain.ConfidenceReported,
			ExtractionMethod:   "gazette_notice_ingestion",
			ContentHash:        featureHash,
			HashScope:          "normalized_source_record",
			SourceClass:        "provincial_gazette",
			SourceRecordID:     notice.NoticeID,
			Locator:            fmt.Sprintf("%s Gazette notice %s", a.province, notice.NoticeID),
			PipelineVersion:    pipelineVersion,
			ParserVersion:      parserVersion,
			RawSnippet:         notice.Title + ": " + truncate(notice.Content, 500),
		}
		if entity != nil {
			entity.Evidence = evidence
			result.Entities = append(result.Entities, entity)
		}
		
		// Extract sector from notice content
		sector, subsector := classifySectorFromNotice(notice)
		
		// Extract capex from amount text
		capex, capexStatus := parseCapexFromNotice(notice)
		
		// Determine stage from notice category
		stage := mapGazetteCategory(notice.Category)
		
		metadata := map[string]interface{}{
			"source_dataset":      fmt.Sprintf("%s Gazette", a.province),
			"source_notice_id":    notice.NoticeID,
			"dataset_vintage":     fixture.DatasetVintage,
			"source_category":     notice.Category,
			"source_amount_text":  notice.AmountText,
			"gazette_province":    a.province,
		}
		
		project := &domain.Project{
			ID:                   projectID,
			Slug:                 identity.Slug(notice.ProjectName),
			Name:                 notice.ProjectName,
			Summary:              truncate(notice.Content, 1000),
			Sector:               sector,
			Subsector:            subsector,
			Province:             a.province,
			LocationName:         notice.Location,
			CurrentStage:         stage,
			CapexCAD:             capex,
			CapexStatus:          capexStatus,
			ProponentID:          entityID,
			Proponent:            entity,
			Confidence:           domain.ConfidenceReported,
			EvidenceIDs:          []string{evidenceID},
			ExternalIDs:          map[string]string{fmt.Sprintf("gazette_%s", strings.ToLower(a.province)): notice.NoticeID},
			IsSynthetic:          false,
			LastMeaningfulUpdate: effective,
			Metadata:             metadata,
			CreatedAt:            effective,
			UpdatedAt:            effective,
		}
		
		result.Evidence = append(result.Evidence, evidence)
		result.Projects = append(result.Projects, project)
		
		if entity != nil {
			result.Relationships = append(result.Relationships, &domain.Relationship{
				ID:             identity.StableID("relationship", adapterName, noticeID+":"+entityID+":develops"),
				ProjectID:      projectID,
				SourceEntityID: entityID,
				RelationType:   "develops",
				Confidence:     domain.ConfidenceReported,
				EvidenceID:     evidenceID,
				Evidence:       evidence,
				CreatedAt:      effective,
			})
		}
	}
	
	a.health.DocumentsSeen = len(fixture.Notices)
	a.health.DocumentsChanged = len(result.Projects)
	a.health.LastChange = effective
	return result, nil
}

func gazettePublisher(province string) string {
	switch strings.ToLower(province) {
	case "on", "ontario":
		return "Ontario Gazette"
	case "qc", "quebec":
		return "Gazette officielle du Québec"
	case "bc", "british columbia":
		return "BC Gazette"
	case "ca", "federal", "canada":
		return "Canada Gazette"
	default:
		return province + " Gazette"
	}
}

func isProjectRelevant(notice GazetteRecord) bool {
	searchText := strings.ToLower(notice.Title + " " + notice.Content + " " + notice.Category)
	for _, keyword := range projectKeywords {
		if strings.Contains(searchText, keyword) {
			return true
		}
	}
	return false
}

func classifySectorFromNotice(notice GazetteRecord) (domain.Sector, string) {
	searchText := strings.ToLower(notice.Title + " " + notice.Content + " " + notice.Category)
	
	if strings.Contains(searchText, "nuclear") || strings.Contains(searchText, "smr") || strings.Contains(searchText, "reactor") {
		return domain.SectorNuclearEnergy, "Nuclear Power"
	}
	if strings.Contains(searchText, "critical mineral") || strings.Contains(searchText, "nickel") || strings.Contains(searchText, "lithium") || strings.Contains(searchText, "cobalt") {
		return domain.SectorCriticalMinerals, "Critical Minerals"
	}
	if strings.Contains(searchText, "transmission") || strings.Contains(searchText, "grid") || strings.Contains(searchText, "wind") || strings.Contains(searchText, "solar") || strings.Contains(searchText, "battery") || strings.Contains(searchText, "storage") || strings.Contains(searchText, "hydro") {
		return domain.SectorCleanEnergy, "Clean Energy & Grid"
	}
	if strings.Contains(searchText, "data centre") || strings.Contains(searchText, "data center") || strings.Contains(searchText, "ai compute") {
		return domain.SectorAICompute, "AI Compute & Data Centres"
	}
	if strings.Contains(searchText, "defence") || strings.Contains(searchText, "arctic") || strings.Contains(searchText, "military") {
		return domain.SectorDefenceArctic, "Defence & Arctic"
	}
	if strings.Contains(searchText, "port") || strings.Contains(searchText, "rail") || strings.Contains(searchText, "highway") || strings.Contains(searchText, "transit") {
		return domain.SectorTransportation, "Transportation & Ports"
	}
	if strings.Contains(searchText, "housing") || strings.Contains(searchText, "infrastructure") {
		return domain.SectorHousingEnabling, "Housing-Enabling Infrastructure"
	}
	if strings.Contains(searchText, "mining") || strings.Contains(searchText, "quarry") {
		return domain.SectorMiningMetals, "Mining & Metals"
	}
	if strings.Contains(searchText, "oil") || strings.Contains(searchText, "gas") || strings.Contains(searchText, "pipeline") || strings.Contains(searchText, "refinery") {
		return domain.SectorEnergyFuels, "Energy & Fuels"
	}
	if strings.Contains(searchText, "forest") || strings.Contains(searchText, "lumber") || strings.Contains(searchText, "pulp") {
		return domain.SectorForestryBioeconomy, "Forestry & Bioeconomy"
	}
	
	return domain.SectorIndustrialMfg, "Industrial & Manufacturing"
}

func mapGazetteCategory(category string) domain.LifecycleStage {
	cat := strings.ToUpper(strings.TrimSpace(category))
	switch {
	case strings.Contains(cat, "ENVIRONMENTAL") || strings.Contains(cat, "ASSESSMENT"):
		return domain.StageEnvironmentalReview
	case strings.Contains(cat, "PERMIT") || strings.Contains(cat, "LICENSE") || strings.Contains(cat, "AUTHORIZATION"):
		return domain.StagePermitting
	case strings.Contains(cat, "PROCUREMENT") || strings.Contains(cat, "TENDER") || strings.Contains(cat, "RFQ") || strings.Contains(cat, "RFP"):
		return domain.StageProcurement
	case strings.Contains(cat, "CONSTRUCTION") || strings.Contains(cat, "BUILD"):
		return domain.StageConstruction
	case strings.Contains(cat, "FINANCING") || strings.Contains(cat, "FUNDING") || strings.Contains(cat, "INVESTMENT"):
		return domain.StageFinancing
	case strings.Contains(cat, "APPROVAL") || strings.Contains(cat, "DECISION"):
		return domain.StageFID
	case strings.Contains(cat, "OPERATING") || strings.Contains(cat, "COMPLETION"):
		return domain.StageOperating
	case strings.Contains(cat, "ANNOUNCE") || strings.Contains(cat, "PROPOSE"):
		return domain.StageAnnounced
	case strings.Contains(cat, "DELAY") || strings.Contains(cat, "HOLD") || strings.Contains(cat, "PAUSE"):
		return domain.StageDelayed
	case strings.Contains(cat, "CANCEL") || strings.Contains(cat, "TERMINAT"):
		return domain.StageCancelled
	default:
		return domain.StageUnknown
	}
}

func parseCapexFromNotice(notice GazetteRecord) (int64, domain.ConfidenceLevel) {
	if notice.AmountText == "" {
		return 0, domain.ConfidenceUnknown
	}
	
	matches := capexPattern.FindStringSubmatch(notice.AmountText)
	if len(matches) < 2 {
		return 0, domain.ConfidenceUnknown
	}
	
	amountStr := strings.ReplaceAll(matches[1], ",", "")
	amount, err := parseAmount(amountStr, notice.AmountText)
	if err != nil {
		return 0, domain.ConfidenceUnknown
	}
	
	return amount, domain.ConfidenceReported
}

func parseAmount(value, original string) (int64, error) {
	lower := strings.ToLower(original)
	multiplier := int64(1)
	
	if strings.Contains(lower, "billion") || strings.Contains(lower, " b") {
		multiplier = 1_000_000_000
	} else if strings.Contains(lower, "million") || strings.Contains(lower, " m") {
		multiplier = 1_000_000
	}
	
	num, err := parseFloat(value)
	if err != nil {
		return 0, err
	}
	
	return int64(num * float64(multiplier)), nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func parseTimestamp(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, fmt.Errorf("timestamp is required")
	}
	
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"02/01/2006",
		"January 2, 2006",
	}
	
	for _, format := range formats {
		if parsed, err := time.Parse(format, raw); err == nil {
			return parsed.UTC(), nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", raw)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func (a *GazetteAdapter) fail(err error) {
	a.health.Status = string(domain.StatusDegraded)
	a.health.LastError = err.Error()
}

func (a *GazetteAdapter) parseError(err error) error {
	a.health.ParseFailures++
	a.fail(err)
	return err
}
