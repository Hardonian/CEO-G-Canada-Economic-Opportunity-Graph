package ckan

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode"
)

const MaxPayloadBytes = 8 << 20

type rawSearchEnvelope struct {
	Success bool `json:"success"`
	Result  struct {
		Count   int          `json:"count"`
		Sort    string       `json:"sort"`
		Results []rawDataset `json:"results"`
	} `json:"result"`
	Error json.RawMessage `json:"error"`
}

type rawDataset struct {
	ID                    string                    `json:"id"`
	Name                  string                    `json:"name"`
	Type                  string                    `json:"type"`
	State                 string                    `json:"state"`
	Title                 string                    `json:"title"`
	TitleTranslated       LocalizedText             `json:"title_translated"`
	Notes                 string                    `json:"notes"`
	NotesTranslated       LocalizedText             `json:"notes_translated"`
	Organization          rawOrganization           `json:"organization"`
	OrgTitleAtPublication LocalizedText             `json:"org_title_at_publication"`
	OwnerOrg              string                    `json:"owner_org"`
	Jurisdiction          string                    `json:"jurisdiction"`
	LicenseID             string                    `json:"license_id"`
	LicenseTitle          string                    `json:"license_title"`
	LicenseURL            string                    `json:"license_url"`
	Frequency             string                    `json:"frequency"`
	URL                   string                    `json:"url"`
	Private               bool                      `json:"private"`
	IsOpen                bool                      `json:"isopen"`
	Keywords              json.RawMessage           `json:"keywords"`
	Subject               stringList                `json:"subject"`
	MetadataCreated       string                    `json:"metadata_created"`
	MetadataModified      string                    `json:"metadata_modified"`
	Tags                  []rawTag                  `json:"tags"`
	Resources             []rawResource             `json:"resources"`
}

type rawOrganization struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

type rawTag struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type rawResource struct {
	ID                    string        `json:"id"`
	PackageID             string        `json:"package_id"`
	Name                  string        `json:"name"`
	NameTranslated        LocalizedText `json:"name_translated"`
	Description           string        `json:"description"`
	DescriptionTranslated LocalizedText `json:"description_translated"`
	URL                   string        `json:"url"`
	Format                string        `json:"format"`
	MIMEType              string        `json:"mimetype"`
	ResourceType          string        `json:"resource_type"`
	URLType               string        `json:"url_type"`
	Language              stringList    `json:"language"`
	State                 string        `json:"state"`
	Hash                  string        `json:"hash"`
	Position              int           `json:"position"`
	DataStoreActive       bool          `json:"datastore_active"`
	Created               string        `json:"created"`
	LastModified          string        `json:"last_modified"`
	MetadataModified      string        `json:"metadata_modified"`
}

type stringList []string

func (values *stringList) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*values = nil
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err == nil {
		*values = list
		return nil
	}
	var scalar string
	if err := json.Unmarshal(data, &scalar); err != nil {
		return fmt.Errorf("string list must be a string or array: %w", err)
	}
	if scalar == "" {
		*values = nil
	} else {
		*values = []string{scalar}
	}
	return nil
}

type rawChangeEnvelope struct {
	Success bool                `json:"success"`
	Result  []rawChangeActivity `json:"result"`
	Error   json.RawMessage     `json:"error"`
}

type rawChangeActivity struct {
	ID           string `json:"id"`
	Timestamp    string `json:"timestamp"`
	ObjectID     string `json:"object_id"`
	ActivityType string `json:"activity_type"`
	Data         struct {
		Package *rawDataset `json:"package"`
	} `json:"data"`
}

// ParsePackageSearch parses one CKAN package_search response and enumerates
// every resource without fetching any resource URL.
func ParsePackageSearch(catalogURL string, payload []byte, start int) (*SearchPage, error) {
	if start < 0 {
		return nil, fmt.Errorf("CKAN page start must be non-negative")
	}
	catalogURL, _, err := normalizeCatalogURL(catalogURL)
	if err != nil {
		return nil, err
	}
	var envelope rawSearchEnvelope
	if err := decodePayload(payload, &envelope); err != nil {
		return nil, fmt.Errorf("parse CKAN package_search response: %w", err)
	}
	if !envelope.Success {
		return nil, ckanFailure("package_search", envelope.Error)
	}
	if envelope.Result.Count < 0 || envelope.Result.Count < len(envelope.Result.Results) {
		return nil, fmt.Errorf("invalid CKAN package_search count %d for %d results", envelope.Result.Count, len(envelope.Result.Results))
	}
	if len(envelope.Result.Results) > MaxPageSize {
		return nil, fmt.Errorf("CKAN package_search returned %d rows, exceeding %d", len(envelope.Result.Results), MaxPageSize)
	}
	page := &SearchPage{Total: envelope.Result.Count, Start: start, Rows: len(envelope.Result.Results), Sort: envelope.Result.Sort}
	seen := make(map[string]struct{}, len(envelope.Result.Results))
	for index := range envelope.Result.Results {
		dataset, err := convertDataset(catalogURL, envelope.Result.Results[index])
		if err != nil {
			return nil, fmt.Errorf("CKAN dataset row %d: %w", index, err)
		}
		if _, duplicate := seen[dataset.ID]; duplicate {
			return nil, fmt.Errorf("duplicate CKAN dataset %q in page", dataset.RemoteID)
		}
		seen[dataset.ID] = struct{}{}
		page.Datasets = append(page.Datasets, dataset)
	}
	return page, nil
}

// ParseRecentlyChanged parses CKAN's changed-package activity stream.
func ParseRecentlyChanged(catalogURL string, payload []byte) ([]Change, error) {
	catalogURL, _, err := normalizeCatalogURL(catalogURL)
	if err != nil {
		return nil, err
	}
	var envelope rawChangeEnvelope
	if err := decodePayload(payload, &envelope); err != nil {
		return nil, fmt.Errorf("parse CKAN recently-changed response: %w", err)
	}
	if !envelope.Success {
		return nil, ckanFailure("recently_changed_packages_activity_list", envelope.Error)
	}
	changes := make([]Change, 0, len(envelope.Result))
	seen := make(map[string]struct{}, len(envelope.Result))
	for index, activity := range envelope.Result {
		remoteDatasetID := strings.TrimSpace(activity.ObjectID)
		var dataset *Dataset
		if activity.Data.Package != nil {
			converted, convertErr := convertDataset(catalogURL, *activity.Data.Package)
			if convertErr != nil {
				return nil, fmt.Errorf("CKAN activity row %d package: %w", index, convertErr)
			}
			dataset = &converted
			if remoteDatasetID == "" {
				remoteDatasetID = converted.RemoteID
			} else if remoteDatasetID != converted.RemoteID {
				return nil, fmt.Errorf("CKAN activity row %d object_id %q does not match package id %q", index, remoteDatasetID, converted.RemoteID)
			}
		}
		if remoteDatasetID == "" {
			return nil, fmt.Errorf("CKAN activity row %d has no dataset identifier", index)
		}
		timestamp, err := parseCKANTime(activity.Timestamp)
		if err != nil || timestamp == nil {
			return nil, fmt.Errorf("CKAN activity row %d timestamp: %w", index, err)
		}
		remoteActivityID := strings.TrimSpace(activity.ID)
		if remoteActivityID == "" {
			remoteActivityID = remoteDatasetID + ":" + activity.ActivityType + ":" + timestamp.Format(time.RFC3339Nano)
		}
		changeID := DeterministicID("change", catalogURL, remoteActivityID)
		if _, duplicate := seen[changeID]; duplicate {
			return nil, fmt.Errorf("duplicate CKAN activity %q", activity.ID)
		}
		seen[changeID] = struct{}{}
		changes = append(changes, Change{
			ID:               changeID,
			RemoteActivityID: strings.TrimSpace(activity.ID),
			DatasetID:        DeterministicID("dataset", catalogURL, remoteDatasetID),
			RemoteDatasetID:  remoteDatasetID,
			ActivityType:     activity.ActivityType,
			Kind:             normalizeChangeKind(activity.ActivityType),
			Timestamp:        timestamp.UTC(),
			Dataset:          dataset,
		})
	}
	return changes, nil
}

func convertDataset(catalogURL string, raw rawDataset) (Dataset, error) {
	remoteID := strings.TrimSpace(raw.ID)
	if remoteID == "" {
		return Dataset{}, fmt.Errorf("dataset id is required")
	}
	created, err := parseCKANTime(raw.MetadataCreated)
	if err != nil {
		return Dataset{}, fmt.Errorf("metadata_created: %w", err)
	}
	modified, err := parseCKANTime(raw.MetadataModified)
	if err != nil {
		return Dataset{}, fmt.Errorf("metadata_modified: %w", err)
	}
	title := mergeLocalized(raw.TitleTranslated, raw.Title)
	description := mergeLocalized(raw.NotesTranslated, raw.Notes)
	publisherTitle := raw.OrgTitleAtPublication
	if publisherTitle.EN == "" && publisherTitle.FR == "" {
		publisherTitle = splitBilingual(raw.Organization.Title)
	}
	keywords, err := parseKeywords(raw.Keywords)
	if err != nil {
		return Dataset{}, fmt.Errorf("keywords: %w", err)
	}
	tags := make([]string, 0, len(raw.Tags))
	for _, tag := range raw.Tags {
		value := strings.TrimSpace(tag.Name)
		if value == "" {
			value = strings.TrimSpace(tag.DisplayName)
		}
		if value != "" {
			tags = append(tags, value)
		}
	}
	tags = uniqueSorted(tags)
	subjects := uniqueSorted([]string(raw.Subject))
	datasetID := DeterministicID("dataset", catalogURL, remoteID)
	dataset := Dataset{
		ID:               datasetID,
		RemoteID:         remoteID,
		CatalogURL:       catalogURL,
		Name:             raw.Name,
		Type:             raw.Type,
		State:            raw.State,
		Title:            title,
		Description:      description,
		Publisher:        Publisher{RemoteID: firstNonEmpty(raw.Organization.ID, raw.OwnerOrg), Name: raw.Organization.Name, Title: publisherTitle},
		Jurisdiction:     raw.Jurisdiction,
		LicenseID:        raw.LicenseID,
		LicenseTitle:     raw.LicenseTitle,
		LicenseURL:       raw.LicenseURL,
		UpdateFrequency:  raw.Frequency,
		URL:              raw.URL,
		Private:          raw.Private,
		Open:             raw.IsOpen,
		Keywords:         keywords,
		Tags:             tags,
		Subjects:         subjects,
		MetadataCreated:  created,
		MetadataModified: modified,
	}
	seenResources := make(map[string]struct{}, len(raw.Resources))
	for index, item := range raw.Resources {
		resource, err := convertResource(catalogURL, datasetID, remoteID, item, index)
		if err != nil {
			return Dataset{}, fmt.Errorf("resource row %d: %w", index, err)
		}
		if _, duplicate := seenResources[resource.ID]; duplicate {
			return Dataset{}, fmt.Errorf("duplicate resource %q", firstNonEmpty(item.ID, item.URL))
		}
		seenResources[resource.ID] = struct{}{}
		dataset.Resources = append(dataset.Resources, resource)
	}
	dataset.Relevance = scoreRelevance(dataset)
	return dataset, nil
}

func convertResource(catalogURL, datasetID, remoteDatasetID string, raw rawResource, index int) (Resource, error) {
	if raw.PackageID != "" && raw.PackageID != remoteDatasetID {
		return Resource{}, fmt.Errorf("package_id %q does not match dataset %q", raw.PackageID, remoteDatasetID)
	}
	remoteID := strings.TrimSpace(raw.ID)
	identityKey := remoteID
	if identityKey == "" {
		identityKey = strings.TrimSpace(raw.URL)
	}
	if identityKey == "" {
		return Resource{}, fmt.Errorf("resource id or URL is required")
	}
	created, err := parseCKANTime(raw.Created)
	if err != nil {
		return Resource{}, fmt.Errorf("created: %w", err)
	}
	lastModified, err := parseCKANTime(raw.LastModified)
	if err != nil {
		return Resource{}, fmt.Errorf("last_modified: %w", err)
	}
	metadataModified, err := parseCKANTime(raw.MetadataModified)
	if err != nil {
		return Resource{}, fmt.Errorf("metadata_modified: %w", err)
	}
	name := mergeLocalized(raw.NameTranslated, raw.Name)
	description := mergeLocalized(raw.DescriptionTranslated, raw.Description)
	family := ClassifyResource(raw.Format, raw.MIMEType, raw.ResourceType, raw.URL)
	archiveContains := FamilyUnknown
	if family == FamilyZIP {
		archiveContains = classifyArchiveContents(raw.Format, raw.MIMEType, raw.ResourceType, raw.URL)
	}
	position := raw.Position
	if position == 0 && index != 0 {
		// Some CKAN exports omit position. Preserve source order in that case.
		position = index
	}
	return Resource{
		ID:               DeterministicID("resource", catalogURL, remoteDatasetID+":"+identityKey),
		RemoteID:         remoteID,
		DatasetID:        datasetID,
		Name:             name,
		Description:      description,
		URL:              raw.URL,
		DeclaredFormat:   raw.Format,
		MIMEType:         raw.MIMEType,
		ResourceType:     raw.ResourceType,
		URLType:          raw.URLType,
		Family:           family,
		ArchiveContains:  archiveContains,
		Languages:        uniqueSorted([]string(raw.Language)),
		State:            raw.State,
		Hash:             raw.Hash,
		Position:         position,
		DataStoreActive:  raw.DataStoreActive,
		CreatedAt:        created,
		LastModified:     lastModified,
		MetadataModified: metadataModified,
	}, nil
}

// DeterministicID produces a stable protocol-scoped identifier without
// depending on database keys or catalog display names.
func DeterministicID(kind, catalogURL, remoteID string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(kind) + "\x00" + strings.TrimSpace(catalogURL) + "\x00" + strings.TrimSpace(remoteID)))
	return "ckan:" + strings.ToLower(strings.TrimSpace(kind)) + ":" + hex.EncodeToString(sum[:])
}

func decodePayload(payload []byte, destination any) error {
	if len(payload) == 0 {
		return fmt.Errorf("empty payload")
	}
	if len(payload) > MaxPayloadBytes {
		return fmt.Errorf("payload exceeds %d-byte CKAN parser limit", MaxPayloadBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return fmt.Errorf("invalid trailing JSON: %w", err)
	}
	return nil
}

func ckanFailure(action string, raw json.RawMessage) error {
	message := strings.TrimSpace(string(raw))
	if message == "" || message == "null" {
		message = "unspecified CKAN error"
	}
	if len(message) > 512 {
		message = message[:512] + "..."
	}
	return fmt.Errorf("CKAN %s reported success=false: %s", action, message)
}

func parseCKANTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	layoutsWithZone := []string{time.RFC3339Nano, time.RFC3339}
	for _, layout := range layoutsWithZone {
		if parsed, err := time.Parse(layout, raw); err == nil {
			parsed = parsed.UTC()
			return &parsed, nil
		}
	}
	layoutsUTC := []string{
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layoutsUTC {
		if parsed, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("unsupported CKAN timestamp %q", raw)
}

func parseKeywords(raw json.RawMessage) (map[string][]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var localized map[string]stringList
	if err := json.Unmarshal(raw, &localized); err == nil {
		result := make(map[string][]string, len(localized))
		for language, values := range localized {
			result[strings.ToLower(language)] = uniqueSorted([]string(values))
		}
		return result, nil
	}
	var list stringList
	if err := json.Unmarshal(raw, &list); err == nil {
		return map[string][]string{"und": uniqueSorted([]string(list))}, nil
	}
	return nil, fmt.Errorf("expected language object, string, or string array")
}

func mergeLocalized(localized LocalizedText, fallback string) LocalizedText {
	if localized.EN == "" && fallback != "" {
		localized.EN = fallback
	}
	return localized
}

func splitBilingual(value string) LocalizedText {
	parts := strings.SplitN(value, " | ", 2)
	if len(parts) == 2 {
		return LocalizedText{EN: strings.TrimSpace(parts[0]), FR: strings.TrimSpace(parts[1])}
	}
	return LocalizedText{EN: strings.TrimSpace(value)}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]string, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; !exists {
			seen[key] = value
		}
	}
	result := make([]string, 0, len(seen))
	for _, value := range seen {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return strings.ToLower(result[i]) < strings.ToLower(result[j]) })
	return result
}

func normalizeChangeKind(activityType string) ChangeKind {
	value := strings.ToLower(strings.TrimSpace(activityType))
	switch {
	case strings.Contains(value, "new") || strings.Contains(value, "create"):
		return ChangeCreated
	case strings.Contains(value, "delete") || strings.Contains(value, "purge"):
		return ChangeDeleted
	case strings.Contains(value, "change") || strings.Contains(value, "update"):
		return ChangeUpdated
	default:
		return ChangeUnknown
	}
}

var relevanceTerms = map[string][]string{
	"ai_compute":            {"artificial intelligence", "intelligence artificielle", "data centre", "data center", "centre de données", "compute", "calcul informatique"},
	"capital":               {"capital expenditure", "capital investment", "capex", "financing", "funding", "grant", "loan", "investissement", "financement", "subvention", "prêt"},
	"construction":          {"construction", "building permit", "development application", "permis de construction", "demande d'aménagement"},
	"corporate":             {"corporation", "company", "business", "issuer", "société", "entreprise", "émetteur"},
	"critical_minerals":     {"critical mineral", "mining", "mine", "mineral processing", "minéraux critiques", "exploitation minière", "traitement des minéraux"},
	"defence_arctic":        {"defence", "defense", "military", "arctic", "norad", "défense", "militaire", "arctique"},
	"energy":                {"energy", "electricity", "power grid", "generation", "transmission", "pipeline", "énergie", "électricité", "réseau électrique", "production", "transport d'électricité"},
	"environment_regulatory": {"environmental assessment", "regulator", "permit", "approval", "impact assessment", "évaluation environnementale", "organisme de réglementation", "permis", "approbation"},
	"housing":               {"housing", "residential development", "affordable housing", "logement", "développement résidentiel", "logement abordable"},
	"indigenous_economy":    {"indigenous", "first nation", "inuit", "métis", "autochtone", "première nation"},
	"infrastructure":        {"infrastructure", "public works", "utility", "utilities", "travaux publics", "service public"},
	"procurement":           {"procurement", "tender", "contract award", "standing offer", "request for proposal", "approvisionnement", "appel d'offres", "attribution de contrat", "offre à commandes"},
	"telecom":               {"telecommunications", "broadband", "fibre", "fiber", "connectivity", "télécommunications", "large bande", "connectivité"},
	"trade":                 {"trade", "export", "import", "supply chain", "commerce", "exportation", "importation", "chaîne d'approvisionnement"},
	"transport":             {"transportation", "transit", "rail", "port", "airport", "highway", "freight", "transport", "ferroviaire", "aéroport", "autoroute", "marchandises"},
	"workforce":             {"employment", "labour", "workforce", "job", "emploi", "main-d'œuvre", "travailleur"},
}

func scoreRelevance(dataset Dataset) Relevance {
	titles := strings.ToLower(dataset.Title.EN + " " + dataset.Title.FR)
	descriptions := strings.ToLower(dataset.Description.EN + " " + dataset.Description.FR)
	var keywordParts []string
	for _, values := range dataset.Keywords {
		keywordParts = append(keywordParts, values...)
	}
	keywordParts = append(keywordParts, dataset.Tags...)
	keywordParts = append(keywordParts, dataset.Subjects...)
	keywords := strings.ToLower(strings.Join(keywordParts, " "))

	result := Relevance{}
	for tag, terms := range relevanceTerms {
		titleMatch := containsAnyTerm(titles, terms)
		keywordMatch := containsAnyTerm(keywords, terms)
		descriptionMatch := containsAnyTerm(descriptions, terms)
		if !titleMatch && !keywordMatch && !descriptionMatch {
			continue
		}
		result.Tags = append(result.Tags, tag)
		if titleMatch {
			result.Score += 15
		}
		if keywordMatch {
			result.Score += 10
		}
		if descriptionMatch {
			result.Score += 5
		}
	}
	if result.Score > 0 {
		for _, resource := range dataset.Resources {
			if resource.Family != FamilyUnknown && resource.Family != FamilyHTML && resource.Family != FamilyPDF {
				result.Score += 5
				break
			}
		}
	}
	if result.Score > 100 {
		result.Score = 100
	}
	sort.Strings(result.Tags)
	return result
}

func containsAnyTerm(text string, terms []string) bool {
	for _, term := range terms {
		if containsTerm(text, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func containsTerm(text, term string) bool {
	if term == "" {
		return false
	}
	if strings.Contains(term, " ") || strings.ContainsAny(term, "-'’") || len([]rune(term)) > 3 {
		return strings.Contains(text, term)
	}
	for _, token := range strings.FieldsFunc(text, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		if token == term {
			return true
		}
	}
	return false
}
