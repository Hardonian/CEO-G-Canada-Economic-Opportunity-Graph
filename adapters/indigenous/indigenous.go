package indigenous

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

const (
	adapterName         = "isc_indigenous_business_directory"
	pipelineVersion     = "indigenous-business-v1"
	parserVersion       = "indigenous-json-v1"
	maxResponseBytes    int64 = 8 << 20
	
	// Indigenous Services Canada Business Directory
	ISCBusinessDirectoryURL = "https://www.sac-isc.gc.ca/eng/1100100032800/1558373938872"
)

type IndigenousAdapter struct {
	fixturePath string
	client      *http.Client
	fetchedAt   time.Time
	health      adapters.SourceHealth
}

type IndigenousBusinessRecord struct {
	BusinessID        string                 `json:"business_id"`
	BusinessName      string                 `json:"business_name"`
	LegalName         string                 `json:"legal_name"`
	OperatingName     string                 `json:"operating_name"`
	IndigenousGroup   string                 `json:"indigenous_group"` // First Nation, Inuit, Metis
	CommunityName     string                 `json:"community_name"`
	Province          string                 `json:"province"`
	City              string                 `json:"city"`
	PostalCode        string                 `json:"postal_code"`
	NAICSCode         string                 `json:"naics_code"`
	NAICSDescription  string                 `json:"naics_description"`
	BusinessType      string                 `json:"business_type"`
	OwnershipPercent  int                    `json:"ownership_percent"`
	CertificationDate string                 `json:"certification_date"`
	ExpiryDate        string                 `json:"expiry_date"`
	Status            string                 `json:"status"` // Active, Expired, Suspended
	ContactEmail      string                 `json:"contact_email"`
	ContactPhone      string                 `json:"contact_phone"`
	Website           string                 `json:"website"`
	Address           string                 `json:"address"`
	Description       string                 `json:"description"`
	Capabilities      []string               `json:"capabilities"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type IndigenousFixture struct {
	Source        string                    `json:"source"`
	SourceURL     string                    `json:"source_url"`
	RetrievedAt   string                    `json:"retrieved_at"`
	EffectiveAt   string                    `json:"effective_at"`
	DatasetVintage string                   `json:"dataset_vintage"`
	Businesses    []IndigenousBusinessRecord `json:"businesses"`
}

// NewIndigenousAdapter creates a curated snapshot adapter for the ISC Indigenous Business Directory.
func NewIndigenousAdapter(fixturePath string) *IndigenousAdapter {
	if fixturePath == "" {
		fixturePath = "data/fixtures/isc_indigenous_business_directory.json"
	}
	return &IndigenousAdapter{
		fixturePath: fixturePath,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		health: adapters.SourceHealth{
			AdapterName:    adapterName,
			Tier:           domain.SourceTier1,
			Status:         string(domain.StatusHealthy),
			RateLimitState: "NOT_APPLICABLE",
			Mode:           "CURATED_SNAPSHOT",
		},
	}
}

func (a *IndigenousAdapter) Name() string                   { return adapterName }
func (a *IndigenousAdapter) Tier() domain.SourceTier        { return domain.SourceTier1 }
func (a *IndigenousAdapter) Health() *adapters.SourceHealth { return &a.health }

func (a *IndigenousAdapter) Fetch(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	
	a.health.LastAttempt = time.Now().UTC()
	data, err := adapters.ReadBoundedFile(a.fixturePath, maxResponseBytes)
	if err != nil {
		a.fail(err)
		return nil, fmt.Errorf("read ISC Indigenous Business Directory snapshot: %w", err)
	}
	
	a.fetchedAt = time.Now().UTC()
	a.health.LastSuccess = a.fetchedAt
	a.health.Status = string(domain.StatusHealthy)
	a.health.LastError = ""
	return data, nil
}

func (a *IndigenousAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, a.parseError(fmt.Errorf("empty Indigenous Business Directory document"))
	}
	
	var fixture IndigenousFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return nil, a.parseError(fmt.Errorf("parse ISC Indigenous Business Directory JSON: %w", err))
	}
	
	if len(fixture.Businesses) == 0 {
		return nil, a.parseError(fmt.Errorf("Indigenous Business Directory contains no records"))
	}
	
	if len(fixture.Businesses) > 10000 {
		return nil, a.parseError(fmt.Errorf("ISC business directory exceeds 10000-record safety limit"))
	}
	
	result := &adapters.IngestionResult{}
	seen := make(map[string]struct{}, len(fixture.Businesses))
	
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
		sourceURL = ISCBusinessDirectoryURL
	}
	
	for index, biz := range fixture.Businesses {
		if strings.TrimSpace(biz.BusinessID) == "" || strings.TrimSpace(biz.BusinessName) == "" {
			continue
		}
		
		if _, exists := seen[biz.BusinessID]; exists {
			continue
		}
		seen[biz.BusinessID] = struct{}{}
		
		featureHash, hashErr := adapters.HashRecord(biz)
		if hashErr != nil {
			return nil, a.parseError(hashErr)
		}
		
		entityID := identity.StableID("entity", "ca:indigenous", biz.BusinessID)
		evidenceID := identity.StableID("evidence", adapterName, biz.BusinessID+":"+featureHash)
		
		entity := &domain.Entity{
			ID:           entityID,
			Slug:         identity.Slug(biz.BusinessName),
			LegalName:    biz.LegalName,
			CommonName:   biz.BusinessName,
			Aliases:      []string{biz.OperatingName},
			EntityType:   "IndigenousBusiness",
			Jurisdiction: "CA:" + biz.Province,
			Website:      biz.Website,
			Identifiers: map[string]string{
				"isc_business_id": biz.BusinessID,
				"naics_code":      biz.NAICSCode,
			},
			Description: buildBusinessDescription(biz),
			EvidenceID:   evidenceID,
			CreatedAt:    effective,
			UpdatedAt:    effective,
		}
		
		evidence := &domain.Evidence{
			ID:                 evidenceID,
			SourceURL:          sourceURL,
			Publisher:          "Indigenous Services Canada",
			SourceTier:         domain.SourceTier1,
			RetrievalTimestamp: retrieved,
			EffectiveDate:      &effective,
			Confidence:         domain.ConfidenceVerified,
			ExtractionMethod:   "official_indigenous_business_directory",
			ContentHash:        featureHash,
			HashScope:          "normalized_source_record",
			SourceClass:        "federal_indigenous_registry",
			SourceRecordID:     biz.BusinessID,
			Locator:            fmt.Sprintf("ISC Indigenous Business Directory %s", biz.BusinessID),
			PipelineVersion:    pipelineVersion,
			ParserVersion:      parserVersion,
			RawSnippet:         truncate(buildBusinessDescription(biz), 500),
		}
		entity.Evidence = evidence
		
		result.Entities = append(result.Entities, entity)
		result.Evidence = append(result.Evidence, evidence)
	}
	
	a.health.DocumentsSeen = len(fixture.Businesses)
	a.health.DocumentsChanged = len(result.Entities)
	a.health.LastChange = effective
	return result, nil
}

// CrossReferenceProcurement finds Indigenous businesses eligible for specific procurement opportunities.
func CrossReferenceProcurement(businesses []*domain.Entity, procurement *domain.Procurement) []*domain.Entity {
	var matches []*domain.Entity
	
	procurementNAICS := extractNAICSFromProcurement(procurement)
	procurementRegion := extractRegionFromProcurement(procurement)
	
	for _, biz := range businesses {
		if biz.EntityType != "IndigenousBusiness" {
			continue
		}
		
		// Check if business is active
		if !isBusinessActive(biz) {
			continue
		}
		
		// NAICS match
		bizNAICS := biz.Identifiers["naics_code"]
		if procurementNAICS != "" && bizNAICS != "" {
			if !naicsMatch(procurementNAICS, bizNAICS) {
				continue
			}
		}
		
		// Regional match
		if procurementRegion != "" {
			bizRegion := extractProvinceFromJurisdiction(biz.Jurisdiction)
			if !regionMatch(procurementRegion, bizRegion) {
				continue
			}
		}
		
		// Ownership threshold for set-asides
		ownership := getOwnershipPercent(biz.Metadata)
		if ownership < 51 {
			continue
		}
		
		matches = append(matches, biz)
	}
	
	return matches
}

// FindIndigenousPartnersForProject finds Indigenous businesses that could partner on a project.
func FindIndigenousPartnersForProject(businesses []*domain.Entity, project *domain.Project) []*domain.Entity {
	var partners []*domain.Entity
	
	projectSector := mapSectorToNAICS(project.Sector)
	projectProvince := project.Province
	
	for _, biz := range businesses {
		if biz.EntityType != "IndigenousBusiness" {
			continue
		}
		
		if !isBusinessActive(biz) {
			continue
		}
		
		// Sector capability match
		if !hasRelevantCapability(biz, projectSector, project.Subsector) {
			continue
		}
		
		// Geographic proximity
		bizProvince := extractProvinceFromJurisdiction(biz.Jurisdiction)
		if bizProvince != "" && projectProvince != "" && !provinceAdjacent(bizProvince, projectProvince) {
			continue
		}
		
		partners = append(partners, biz)
	}
	
	return partners
}

func isBusinessActive(biz *domain.Entity) bool {
	if biz.Metadata == nil {
		return true // Default to active if no metadata
	}
	
	if status, ok := biz.Metadata["status"].(string); ok {
		return strings.EqualFold(status, "active")
	}
	if expiry, ok := biz.Metadata["expiry_date"].(string); ok {
		if parsed, err := time.Parse(time.RFC3339, expiry); err == nil {
			return time.Now().Before(parsed)
		}
	}
	return true
}

func extractNAICSFromProcurement(proc *domain.Procurement) string {
	if proc.Metadata == nil {
		return ""
	}
	if naics, ok := proc.Metadata["naics_code"].(string); ok {
		return naics
	}
	// Try to infer from categories
	for _, cat := range proc.Categories {
		if naics := categoryToNAICS(cat); naics != "" {
			return naics
		}
	}
	return ""
}

func extractRegionFromProcurement(proc *domain.Procurement) string {
	if proc.Metadata == nil {
		return ""
	}
	if region, ok := proc.Metadata["region"].(string); ok {
		return region
	}
	if province, ok := proc.Metadata["province"].(string); ok {
		return province
	}
	return ""
}

func naicsMatch(procNAICS, bizNAICS string) bool {
	// Match at 4-digit level (industry group)
	if len(procNAICS) >= 4 && len(bizNAICS) >= 4 {
		return procNAICS[:4] == bizNAICS[:4]
	}
	return procNAICS == bizNAICS
}

func regionMatch(procRegion, bizRegion string) bool {
	proc := strings.ToUpper(strings.TrimSpace(procRegion))
	biz := strings.ToUpper(strings.TrimSpace(bizRegion))
	
	if proc == biz {
		return true
	}
	
	// Check if same province
	provinceMap := map[string]string{
		"ON": "ONTARIO", "QC": "QUEBEC", "BC": "BRITISH COLUMBIA",
		"AB": "ALBERTA", "MB": "MANITOBA", "SK": "SASKATCHEWAN",
		"NS": "NOVA SCOTIA", "NB": "NEW BRUNSWICK", "NL": "NEWFOUNDLAND AND LABRADOR",
		"PE": "PRINCE EDWARD ISLAND", "NT": "NORTHWEST TERRITORIES",
		"NU": "NUNAVUT", "YT": "YUKON",
	}
	
	if full, ok := provinceMap[biz]; ok && strings.Contains(strings.ToUpper(full), proc) {
		return true
	}
	if full, ok := provinceMap[proc]; ok && strings.Contains(strings.ToUpper(full), biz) {
		return true
	}
	
	return false
}

func getOwnershipPercent(biz *domain.Entity) int {
	if biz.Metadata == nil {
		return 100
	}
	if ownership, ok := biz.Metadata["ownership_percent"].(float64); ok {
		return int(ownership)
	}
	if ownership, ok := biz.Metadata["ownership_percent"].(int); ok {
		return ownership
	}
	return 100
}

func hasRelevantCapability(biz *domain.Entity, projectSector, projectSubsector domain.Sector) bool {
	capabilities := []string{}
	if caps, ok := biz.Metadata["capabilities"].([]interface{}); ok {
		for _, c := range caps {
			if s, ok := c.(string); ok {
				capabilities = append(capabilities, strings.ToLower(s))
			}
		}
	}
	
	// Also check NAICS
	bizNAICS := biz.Identifiers["naics_code"]
	
	sectorKeywords := sectorKeywords(projectSector, projectSubsector)
	
	for _, cap := range capabilities {
		for _, keyword := range sectorKeywords {
			if strings.Contains(cap, strings.ToLower(keyword)) {
				return true
			}
		}
	}
	
	if bizNAICS != "" {
		for _, keyword := range sectorKeywords {
			if strings.Contains(bizNAICS, keyword) {
				return true
			}
		}
	}
	
	return false
}

func sectorKeywords(sector, subsector domain.Sector) []string {
	keywords := []string{}
	
	switch sector {
	case domain.SectorCriticalMinerals:
		keywords = append(keywords, "mining", "mineral", "exploration", "processing", "nickel", "lithium", "cobalt", "copper")
	case domain.SectorNuclearEnergy:
		keywords = append(keywords, "nuclear", "reactor", "smr", "radiation", "uranium")
	case domain.SectorCleanEnergy:
		keywords = append(keywords, "renewable", "wind", "solar", "hydro", "transmission", "grid", "storage", "battery")
	case domain.SectorAICompute:
		keywords = append(keywords, "data centre", "data center", "compute", "cloud", "ai", "artificial intelligence")
	case domain.SectorDefenceArctic:
		keywords = append(keywords, "defence", "defense", "arctic", "military", "security")
	case domain.SectorTransportation:
		keywords = append(keywords, "transport", "rail", "port", "highway", "transit", "logistics")
	case domain.SectorHousingInfra:
		keywords = append(keywords, "construction", "housing", "infrastructure", "civil", "site prep")
	case domain.SectorMiningMetals:
		keywords = append(keywords, "mining", "metal", "smelting", "refining", "quarry")
	case domain.SectorEnergyFuels:
		keywords = append(keywords, "oil", "gas", "pipeline", "refinery", "petrochemical")
	case domain.SectorForestryBioeconomy:
		keywords = append(keywords, "forest", "lumber", "pulp", "paper", "bioeconomy", "biomass")
	}
	
	return keywords
}

func mapSectorToNAICS(sector domain.Sector) string {
	switch sector {
	case domain.SectorCriticalMinerals:
		return "212"
	case domain.SectorNuclearEnergy:
		return "221113"
	case domain.SectorCleanEnergy:
		return "2211"
	case domain.SectorAICompute:
		return "518210"
	case domain.SectorDefenceArctic:
		return "928110"
	case domain.SectorTransportation:
		return "48"
	case domain.SectorHousingInfra:
		return "23"
	case domain.SectorMiningMetals:
		return "212"
	case domain.SectorEnergyFuels:
		return "211"
	case domain.SectorForestryBioeconomy:
		return "113"
	default:
		return "23"
	}
}

func categoryToNAICS(category string) string {
	cat := strings.ToLower(category)
	switch {
	case strings.Contains(cat, "construction"):
		return "23"
	case strings.Contains(cat, "engineering"):
		return "541330"
	case strings.Contains(cat, "environmental"):
		return "541620"
	case strings.Contains(cat, "it") || strings.Contains(cat, "software") || strings.Contains(cat, "technology"):
		return "54151"
	case strings.Contains(cat, "consulting"):
		return "5416"
	case strings.Contains(cat, "transport") || strings.Contains(cat, "logistics"):
		return "48"
	case strings.Contains(cat, "manufacturing"):
		return "31-33"
	case strings.Contains(cat, "mining"):
		return "212"
	}
	return ""
}

func extractProvinceFromJurisdiction(jurisdiction string) string {
	// Format: "CA:ON" or "CA"
	parts := strings.Split(jurisdiction, ":")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func provinceAdjacent(a, b string) bool {
	// Same province
	if a == b {
		return true
	}
	
	// Adjacent provinces
	adjacent := map[string][]string{
		"ON": {"QC", "MB"},
		"QC": {"ON", "NB", "NL"},
		"BC": {"AB", "YT", "NT"},
		"AB": {"BC", "SK", "NT"},
		"MB": {"ON", "SK", "NU"},
		"SK": {"AB", "MB", "NT"},
		"NS": {"NB", "PE"},
		"NB": {"QC", "NS", "PE"},
		"NL": {"QC"},
		"PE": {"NB", "NS"},
		"NT": {"BC", "AB", "SK", "MB", "NU", "YT"},
		"NU": {"MB", "NT"},
		"YT": {"BC", "NT"},
	}
	
	if list, ok := adjacent[a]; ok {
		for _, adj := range list {
			if adj == b {
				return true
			}
		}
	}
	return false
}

func buildBusinessDescription(biz IndigenousBusinessRecord) string {
	parts := []string{biz.BusinessName}
	if biz.LegalName != "" && biz.LegalName != biz.BusinessName {
		parts = append(parts, "("+biz.LegalName+")")
	}
	if biz.IndigenousGroup != "" {
		parts = append(parts, biz.IndigenousGroup+"-owned")
	}
	if biz.CommunityName != "" {
		parts = append(parts, "community: "+biz.CommunityName)
	}
	if biz.NAICSDescription != "" {
		parts = append(parts, biz.NAICSDescription)
	}
	if len(biz.Capabilities) > 0 {
		parts = append(parts, "capabilities: "+strings.Join(biz.Capabilities, ", "))
	}
	if biz.Description != "" {
		parts = append(parts, biz.Description)
	}
	return strings.Join(parts, "; ")
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

func (a *IndigenousAdapter) fail(err error) {
	a.health.Status = string(domain.StatusDegraded)
	a.health.LastError = err.Error()
}

func (a *IndigenousAdapter) parseError(err error) error {
	a.health.ParseFailures++
	a.fail(err)
	return err
}