// Package nrcan_major_projects ingests Natural Resources Canada's public
// Major Projects Inventory (MPI). The adapter accepts either a pinned ArcGIS
// snapshot or the legacy hand-authored fixture shape.
package nrcan_major_projects

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

const (
	adapterName = "nrcan_major_projects_inventory"

	// DefaultEndpoint is deliberately fixed to the official Government of
	// Canada ArcGIS service. Runtime URL overrides are validated against the
	// same host and path prefix before any request is made.
	DefaultEndpoint        = "https://maps-cartes.services.geo.ca/server_serveur/rest/services/NRCan/major_projects_inventory_en/MapServer/0/query?where=1%3D1&outFields=*&returnGeometry=true&outSR=4326&resultRecordCount=1000&f=json"
	datasetPageURL         = "https://open.canada.ca/data/en/dataset/f5f2db55-31e4-42fb-8c73-23e1c44de9b2"
	maxResponseBytes int64 = 8 << 20
)

type rawNRCanRecord struct {
	NRCanID          string  `json:"nrcan_id"`
	ProjectName      string  `json:"project_name"`
	ProponentName    string  `json:"proponent_name"`
	Province         string  `json:"province"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Sector           string  `json:"sector"`
	Subsector        string  `json:"subsector"`
	Stage            string  `json:"stage"`
	CapexCAD         int64   `json:"capex_cad"`
	CIBFinancingCAD  int64   `json:"cib_financing_cad"`
	NRCanGrantCAD    int64   `json:"nrcan_grant_cad"`
	AnnouncementDate string  `json:"announcement_date"`
	Summary          string  `json:"summary"`
}

type snapshotEnvelope struct {
	Source         string       `json:"source"`
	SourceURL      string       `json:"source_url"`
	DatasetVintage string       `json:"dataset_vintage"`
	EffectiveAt    string       `json:"effective_at"`
	RetrievedAt    string       `json:"retrieved_at"`
	License        string       `json:"license"`
	Features       []arcFeature `json:"features"`
}

type arcGISResponse struct {
	Features []arcFeature `json:"features"`
	Error    *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type arcFeature struct {
	Attributes arcAttributes `json:"attributes"`
	Geometry   arcGeometry   `json:"geometry"`
}

type arcAttributes struct {
	ID                  string `json:"id"`
	Company             string `json:"company"`
	ProjectName         string `json:"project_name"`
	Province            string `json:"province"`
	Location            string `json:"location"`
	CapitalCost         string `json:"capital_cost"`
	CapitalCostRange    string `json:"capital_cost_range"`
	Sector              string `json:"sector"`
	Status              string `json:"status"`
	CleanTechnology     string `json:"clean_technology"`
	CleanTechnologyType string `json:"clean_technology_type"`
	ObjectID            int    `json:"OBJECTID"`
}

type arcGeometry struct {
	Longitude float64 `json:"x"`
	Latitude  float64 `json:"y"`
}

type NRCanAdapter struct {
	fixturePath string
	endpoint    string
	client      *http.Client
	fetchedAt   time.Time
	health      adapters.SourceHealth
}

// NewNRCanAdapter constructs a pinned-snapshot adapter. An empty path retains
// compatibility with the original fixture location.
func NewNRCanAdapter(fixturePath string) *NRCanAdapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/nrcan_major_projects.json"
	}
	return &NRCanAdapter{
		fixturePath: fixturePath,
		health: adapters.SourceHealth{
			AdapterName:    adapterName,
			Tier:           domain.SourceTier1,
			Status:         string(domain.StatusHealthy),
			RateLimitState: "NOT_APPLICABLE",
			Mode:           "CURATED_SNAPSHOT",
		},
	}
}

// NewLiveNRCanAdapter constructs an opt-in live adapter. The endpoint must
// remain on the official services.geo.ca MPI API. A nil client receives a
// bounded, redirect-safe default.
func NewLiveNRCanAdapter(client *http.Client, endpoint string) (*NRCanAdapter, error) {
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	if err := validateEndpoint(endpoint); err != nil {
		return nil, err
	}
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				return validateEndpoint(req.URL.String())
			},
		}
	}
	return &NRCanAdapter{
		endpoint: endpoint,
		client:   client,
		health: adapters.SourceHealth{
			AdapterName:    adapterName,
			Tier:           domain.SourceTier1,
			Status:         string(domain.StatusHealthy),
			RateLimitState: "AVAILABLE",
			Mode:           "LIVE",
		},
	}, nil
}

func (a *NRCanAdapter) Name() string                   { return adapterName }
func (a *NRCanAdapter) Tier() domain.SourceTier        { return domain.SourceTier1 }
func (a *NRCanAdapter) Health() *adapters.SourceHealth { return &a.health }

func (a *NRCanAdapter) Fetch(ctx context.Context) ([]byte, error) {
	a.health.LastAttempt = time.Now().UTC()
	if a.endpoint == "" {
		data, err := adapters.ReadBoundedFile(a.fixturePath, maxResponseBytes)
		if err != nil {
			a.fail(err)
			return nil, fmt.Errorf("read NRCan MPI snapshot: %w", err)
		}
		a.health.LastSuccess = a.health.LastAttempt
		a.health.Status = string(domain.StatusHealthy)
		a.health.LastError = ""
		return data, nil
	}

	if err := validateEndpoint(a.endpoint); err != nil {
		a.fail(err)
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpoint, nil)
	if err != nil {
		a.fail(err)
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CanadaOpportunityGraph/1.0 (+https://github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph)")
	resp, err := a.client.Do(req)
	if err != nil {
		a.fail(err)
		return nil, fmt.Errorf("fetch NRCan MPI: %w", err)
	}
	defer resp.Body.Close()
	if err := validateEndpoint(resp.Request.URL.String()); err != nil {
		a.fail(err)
		return nil, fmt.Errorf("NRCan MPI redirect rejected: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("NRCan MPI returned HTTP %d", resp.StatusCode)
		a.fail(err)
		return nil, err
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		a.fail(err)
		return nil, fmt.Errorf("read NRCan MPI response: %w", err)
	}
	if int64(len(body)) > maxResponseBytes {
		err := fmt.Errorf("NRCan MPI response exceeds %d byte limit", maxResponseBytes)
		a.fail(err)
		return nil, err
	}
	a.fetchedAt = time.Now().UTC()
	a.health.LastSuccess = a.fetchedAt
	a.health.Status = string(domain.StatusHealthy)
	a.health.LastError = ""
	return body, nil
}

func (a *NRCanAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, a.parseError(fmt.Errorf("empty NRCan MPI document"))
	}

	// Legacy fixtures are still accepted so downstream adopters are not broken.
	if strings.HasPrefix(trimmed, "[") {
		var records []rawNRCanRecord
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, a.parseError(fmt.Errorf("parse legacy NRCan fixture: %w", err))
		}
		return a.parseLegacy(records)
	}

	var snapshot snapshotEnvelope
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, a.parseError(fmt.Errorf("parse NRCan MPI JSON: %w", err))
	}
	if snapshot.Source != "" {
		return a.parseFeatures(snapshot.Features, snapshot.RetrievedAt, snapshot.EffectiveAt, snapshot.DatasetVintage, snapshot.SourceURL)
	}

	var response arcGISResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, a.parseError(fmt.Errorf("parse NRCan ArcGIS response: %w", err))
	}
	if response.Error != nil {
		return nil, a.parseError(fmt.Errorf("NRCan ArcGIS error %d: %s", response.Error.Code, response.Error.Message))
	}
	retrieved := a.fetchedAt
	if retrieved.IsZero() {
		retrieved = time.Now().UTC()
	}
	return a.parseFeatures(response.Features, retrieved.Format(time.RFC3339), retrieved.Format(time.RFC3339), "live", datasetPageURL)
}

func (a *NRCanAdapter) parseFeatures(features []arcFeature, retrievedRaw, effectiveRaw, vintage, sourceURL string) (*adapters.IngestionResult, error) {
	if len(features) == 0 {
		return nil, a.parseError(fmt.Errorf("NRCan MPI contains no features"))
	}
	if len(features) > 1000 {
		return nil, a.parseError(fmt.Errorf("NRCan MPI feature count exceeds 1000-record safety limit"))
	}
	retrieved, err := parseTimestamp(retrievedRaw)
	if err != nil {
		return nil, a.parseError(fmt.Errorf("invalid snapshot retrieved_at: %w", err))
	}
	effective, err := parseTimestamp(effectiveRaw)
	if err != nil {
		return nil, a.parseError(fmt.Errorf("invalid snapshot effective_at: %w", err))
	}
	if sourceURL == "" {
		sourceURL = datasetPageURL
	}
	if vintage == "" {
		vintage = "unknown"
	}

	result := &adapters.IngestionResult{}
	seen := make(map[string]struct{}, len(features))
	for index, feature := range features {
		attrs := feature.Attributes
		if strings.TrimSpace(attrs.ID) == "" || strings.TrimSpace(attrs.ProjectName) == "" {
			return nil, a.parseError(fmt.Errorf("feature %d lacks id or project_name", index))
		}
		if _, exists := seen[attrs.ID]; exists {
			return nil, a.parseError(fmt.Errorf("duplicate NRCan project id %q", attrs.ID))
		}
		seen[attrs.ID] = struct{}{}

		province := provinceCode(attrs.Province)
		sector, subsector := classifySector(attrs)
		stage := mapStage(attrs.Status)
		capex, capexStatus := parseCapitalCost(attrs.CapitalCost)
		featureHash, hashErr := adapters.HashRecord(feature)
		if hashErr != nil {
			return nil, a.parseError(hashErr)
		}
		projectID := identity.StableID("project", adapterName, attrs.ID)
		evidenceID := identity.StableID("evidence", adapterName, attrs.ID+":"+featureHash)
		entityID := ""
		var entity *domain.Entity
		if strings.TrimSpace(attrs.Company) != "" {
			entityID = identity.StableID("entity", "ca", attrs.Company)
			entity = &domain.Entity{
				ID:           entityID,
				Slug:         identity.Slug(attrs.Company),
				LegalName:    attrs.Company,
				CommonName:   attrs.Company,
				Aliases:      []string{},
				EntityType:   "ProjectProponent",
				Jurisdiction: jurisdiction(province),
				EvidenceID:   evidenceID,
				CreatedAt:    effective,
				UpdatedAt:    effective,
			}
		}

		evidence := &domain.Evidence{
			ID:                 evidenceID,
			SourceURL:          sourceURL,
			Publisher:          "Natural Resources Canada",
			SourceTier:         domain.SourceTier1,
			RetrievalTimestamp: retrieved,
			EffectiveDate:      &effective,
			Confidence:         domain.ConfidenceReported,
			ExtractionMethod:   "official_arcgis_snapshot",
			ContentHash:        featureHash,
			HashScope:          "normalized_source_record",
			SourceClass:        "federal_open_data_registry",
			SourceRecordID:     attrs.ID,
			Locator:            fmt.Sprintf("MPI %s project %s", vintage, attrs.ID),
			PipelineVersion:    "nrcan-mpi-v2",
			ParserVersion:      "arcgis-feature-v1",
			RawSnippet:         sourceSummary(attrs, capex),
		}
		if entity != nil {
			entity.Evidence = evidence
			result.Entities = append(result.Entities, entity)
		}

		metadata := map[string]interface{}{
			"source_dataset":              "NRCan Major Projects Inventory",
			"source_project_id":           attrs.ID,
			"dataset_vintage":             vintage,
			"source_sector":               attrs.Sector,
			"source_status":               attrs.Status,
			"capital_cost_range_millions": attrs.CapitalCostRange,
			"clean_technology":            strings.EqualFold(attrs.CleanTechnology, "yes"),
		}
		project := &domain.Project{
			ID:                   projectID,
			Slug:                 identity.Slug(attrs.ProjectName),
			Name:                 attrs.ProjectName,
			Summary:              sourceSummary(attrs, capex),
			Sector:               sector,
			Subsector:            subsector,
			Province:             province,
			LocationName:         attrs.Location,
			Latitude:             feature.Geometry.Latitude,
			Longitude:            feature.Geometry.Longitude,
			CurrentStage:         stage,
			CapexCAD:             capex,
			CapexStatus:          capexStatus,
			ProponentID:          entityID,
			Proponent:            entity,
			Confidence:           domain.ConfidenceReported,
			EvidenceIDs:          []string{evidenceID},
			ExternalIDs:          map[string]string{"nrcan_mpi": attrs.ID},
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
				ID:             identity.StableID("relationship", adapterName, attrs.ID+":"+entityID+":develops"),
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

	a.health.DocumentsSeen = len(features)
	a.health.DocumentsChanged = len(features)
	a.health.LastChange = effective
	return result, nil
}

func (a *NRCanAdapter) parseLegacy(records []rawNRCanRecord) (*adapters.IngestionResult, error) {
	result := &adapters.IngestionResult{}
	for _, rec := range records {
		if strings.TrimSpace(rec.NRCanID) == "" || strings.TrimSpace(rec.ProjectName) == "" {
			return nil, a.parseError(fmt.Errorf("legacy NRCan record lacks nrcan_id or project_name"))
		}
		when := time.Now().UTC()
		if rec.AnnouncementDate != "" {
			if parsed, err := time.Parse(time.RFC3339, rec.AnnouncementDate); err == nil {
				when = parsed
			}
		}
		feature := arcFeature{
			Attributes: arcAttributes{
				ID: rec.NRCanID, Company: rec.ProponentName, ProjectName: rec.ProjectName,
				Province: rec.Province, Location: rec.Province,
				CapitalCost: strconv.FormatFloat(float64(rec.CapexCAD)/1_000_000, 'f', 2, 64),
				Sector:      rec.Sector, Status: rec.Stage, CleanTechnologyType: rec.Subsector,
			},
			Geometry: arcGeometry{Longitude: rec.Longitude, Latitude: rec.Latitude},
		}
		parsed, err := a.parseFeatures([]arcFeature{feature}, when.Format(time.RFC3339), when.Format(time.RFC3339), "legacy", datasetPageURL)
		if err != nil {
			return nil, err
		}
		if len(parsed.Projects) == 1 && rec.Summary != "" {
			parsed.Projects[0].Summary = rec.Summary
		}
		result.Projects = append(result.Projects, parsed.Projects...)
		result.Entities = append(result.Entities, parsed.Entities...)
		result.Evidence = append(result.Evidence, parsed.Evidence...)
		result.Relationships = append(result.Relationships, parsed.Relationships...)
	}
	a.health.DocumentsSeen = len(records)
	a.health.DocumentsChanged = len(records)
	return result, nil
}

func validateEndpoint(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid NRCan MPI endpoint: %w", err)
	}
	if parsed.Scheme != "https" || parsed.Hostname() != "maps-cartes.services.geo.ca" || parsed.Port() != "" || parsed.User != nil {
		return fmt.Errorf("NRCan MPI endpoint must use the official HTTPS host")
	}
	const prefix = "/server_serveur/rest/services/NRCan/major_projects_inventory_en/MapServer/"
	if !strings.HasPrefix(parsed.EscapedPath(), prefix) {
		return fmt.Errorf("NRCan MPI endpoint path is outside the official service")
	}
	return nil
}

func parseTimestamp(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, fmt.Errorf("timestamp is required")
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func parseCapitalCost(raw string) (int64, domain.ConfidenceLevel) {
	cleaned := strings.NewReplacer(",", "", "$", "", " ", "").Replace(strings.TrimSpace(raw))
	if cleaned == "" || strings.EqualFold(cleaned, "N/A") || strings.EqualFold(cleaned, "unknown") {
		return 0, domain.ConfidenceUnknown
	}
	millions, err := strconv.ParseFloat(cleaned, 64)
	if err != nil || millions < 0 || millions > 10_000_000 {
		return 0, domain.ConfidenceUnknown
	}
	return int64(millions * 1_000_000), domain.ConfidenceReported
}

func mapStage(raw string) domain.LifecycleStage {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "UNDER CONSTRUCTION", "CONSTRUCTION":
		return domain.StageConstruction
	case "PLANNED", "ANNOUNCED":
		return domain.StageAnnounced
	case "IN REVIEW":
		return domain.StageEnvironmentalReview
	case "APPROVED":
		return domain.StagePermitting
	case "OPERATING", "COMPLETED":
		return domain.StageOperating
	case "ON HOLD", "PAUSED":
		return domain.StagePaused
	case "CANCELLED", "CANCELED":
		return domain.StageCancelled
	default:
		return domain.StageUnknown
	}
}

func classifySector(attrs arcAttributes) (domain.Sector, string) {
	sourceSector := strings.TrimSpace(attrs.Sector)
	technology := strings.TrimSpace(attrs.CleanTechnologyType)
	switch strings.ToUpper(sourceSector) {
	case "MINING":
		return domain.SectorMiningMetals, "Mining & mineral development"
	case "FOREST", "FORESTRY":
		if !strings.EqualFold(technology, "N/A") && technology != "" {
			return domain.SectorForestryBioeconomy, technology
		}
		return domain.SectorForestryBioeconomy, "Forest products"
	case "ENERGY":
		switch strings.ToUpper(technology) {
		case "NUCLEAR":
			return domain.SectorNuclearEnergy, technology
		case "HYDRO", "WIND", "SOLAR", "TIDAL", "ENERGY STORAGE", "GEOTHERMAL", "BIOENERGY", "CARBON CAPTURE AND STORAGE":
			return domain.SectorCleanEnergy, technology
		default:
			return domain.SectorEnergyFuels, "Energy infrastructure"
		}
	default:
		return domain.SectorIndustrial, sourceSector
	}
}

func provinceCode(raw string) string {
	codes := map[string]string{
		"alberta": "AB", "british columbia": "BC", "manitoba": "MB",
		"new brunswick": "NB", "newfoundland and labrador": "NL", "nova scotia": "NS",
		"northwest territories": "NT", "nunavut": "NU", "ontario": "ON",
		"prince edward island": "PE", "quebec": "QC", "saskatchewan": "SK", "yukon": "YT",
	}
	if code, ok := codes[strings.ToLower(strings.TrimSpace(raw))]; ok {
		return code
	}
	upper := strings.ToUpper(strings.TrimSpace(raw))
	if len(upper) == 2 {
		return upper
	}
	return "Federal"
}

func jurisdiction(province string) string {
	if province == "" || province == "Federal" {
		return "CA"
	}
	return "CA:" + province
}

func sourceSummary(attrs arcAttributes, capex int64) string {
	parts := []string{attrs.ProjectName}
	if attrs.Company != "" {
		parts = append(parts, "led by "+attrs.Company)
	}
	if attrs.Location != "" || attrs.Province != "" {
		location := strings.Trim(strings.Join([]string{attrs.Location, attrs.Province}, ", "), ", ")
		parts = append(parts, "located in "+location)
	}
	if attrs.Status != "" {
		parts = append(parts, "is listed by NRCan as "+strings.ToLower(attrs.Status))
	}
	if capex > 0 {
		parts = append(parts, fmt.Sprintf("with reported capital cost of CAD %.1f million", float64(capex)/1_000_000))
	} else if attrs.CapitalCostRange != "" && !strings.EqualFold(attrs.CapitalCostRange, "N/A") {
		parts = append(parts, "with reported capital-cost range of CAD "+attrs.CapitalCostRange+" million")
	}
	return strings.Join(parts, "; ") + "."
}

func (a *NRCanAdapter) fail(err error) {
	a.health.Status = string(domain.StatusDegraded)
	a.health.LastError = err.Error()
}

func (a *NRCanAdapter) parseError(err error) error {
	a.health.ParseFailures++
	a.fail(err)
	return err
}
