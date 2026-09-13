// Package ckan implements catalog-level discovery for the read-only CKAN
// Action API. Its outputs are deliberately neutral source candidates; graph
// normalization and activation belong to later control-plane stages.
package ckan

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/safefetch"
)

// Family identifies the ingestion protocol or document family most suitable
// for a CKAN resource.
type Family string

const (
	FamilyUnknown          Family = "unknown"
	FamilyRESTAPI          Family = "rest_api"
	FamilyGraphQL          Family = "graphql"
	FamilyOpenAPI          Family = "openapi"
	FamilyCKAN             Family = "ckan"
	FamilySocrata          Family = "socrata"
	FamilyArcGISFeature    Family = "arcgis_feature_server"
	FamilyArcGISMap        Family = "arcgis_map_server"
	FamilyGeoJSON          Family = "geojson"
	FamilyWFS              Family = "wfs"
	FamilyWMS              Family = "wms"
	FamilyWMTS             Family = "wmts"
	FamilySDMX             Family = "sdmx"
	FamilyCSV              Family = "csv"
	FamilyTSV              Family = "tsv"
	FamilyXLS              Family = "xls"
	FamilyXLSX             Family = "xlsx"
	FamilyJSON             Family = "json"
	FamilyJSONL            Family = "jsonl"
	FamilyXML              Family = "xml"
	FamilyRSS              Family = "rss"
	FamilyAtom             Family = "atom"
	FamilyDCAT             Family = "dcat"
	FamilyJSONLD           Family = "json_ld"
	FamilyRDF              Family = "rdf"
	FamilySPARQL           Family = "sparql"
	FamilyHTML             Family = "html"
	FamilyPDF              Family = "pdf"
	FamilyZIP              Family = "zip"
	FamilyKML              Family = "kml"
	FamilyKMZ              Family = "kmz"
	FamilyShapefile        Family = "shapefile"
	FamilyGeoPackage       Family = "geopackage"
	FamilyGit              Family = "git"
)

// LocalizedText retains Canada's English/French sibling representations.
// CKAN installations may encode the value as either a string or a language
// object, so UnmarshalJSON supports both without discarding the source form.
type LocalizedText struct {
	EN string `json:"en,omitempty"`
	FR string `json:"fr,omitempty"`
}

func (text *LocalizedText) UnmarshalJSON(data []byte) error {
	if text == nil {
		return fmt.Errorf("decode localized text into nil receiver")
	}
	if string(data) == "null" {
		*text = LocalizedText{}
		return nil
	}
	var scalar string
	if err := json.Unmarshal(data, &scalar); err == nil {
		text.EN = scalar
		return nil
	}
	var localized map[string]string
	if err := json.Unmarshal(data, &localized); err != nil {
		return fmt.Errorf("localized text must be a string or language object: %w", err)
	}
	text.EN = localized["en"]
	text.FR = localized["fr"]
	return nil
}

// Preferred returns the requested official language with a deterministic
// fallback to the sibling representation.
func (text LocalizedText) Preferred(language string) string {
	if strings.EqualFold(language, "fr") {
		if text.FR != "" {
			return text.FR
		}
		return text.EN
	}
	if text.EN != "" {
		return text.EN
	}
	return text.FR
}

// Publisher is CKAN organization metadata, not an inferred authority claim.
type Publisher struct {
	RemoteID string        `json:"remote_id,omitempty"`
	Name     string        `json:"name,omitempty"`
	Title    LocalizedText `json:"title,omitempty"`
}

// Relevance records deterministic economic-topic matches. It is a discovery
// priority hint, not evidence quality or publisher authority.
type Relevance struct {
	Score int      `json:"score"`
	Tags  []string `json:"tags,omitempty"`
}

// Resource is a source candidate enumerated from a CKAN dataset.
type Resource struct {
	ID                string        `json:"id"`
	RemoteID          string        `json:"remote_id,omitempty"`
	DatasetID         string        `json:"dataset_id"`
	Name              LocalizedText `json:"name,omitempty"`
	Description       LocalizedText `json:"description,omitempty"`
	URL               string        `json:"url"`
	DeclaredFormat    string        `json:"declared_format,omitempty"`
	MIMEType          string        `json:"mime_type,omitempty"`
	ResourceType      string        `json:"resource_type,omitempty"`
	URLType           string        `json:"url_type,omitempty"`
	Family            Family        `json:"family"`
	ArchiveContains   Family        `json:"archive_contains,omitempty"`
	Languages         []string      `json:"languages,omitempty"`
	State             string        `json:"state,omitempty"`
	Hash              string        `json:"hash,omitempty"`
	Position          int           `json:"position"`
	DataStoreActive   bool          `json:"datastore_active"`
	CreatedAt         *time.Time    `json:"created_at,omitempty"`
	LastModified      *time.Time    `json:"last_modified,omitempty"`
	MetadataModified  *time.Time    `json:"metadata_modified,omitempty"`
}

// Dataset is a CKAN package plus all automatically enumerated resources.
type Dataset struct {
	ID               string              `json:"id"`
	RemoteID         string              `json:"remote_id"`
	CatalogURL       string              `json:"catalog_url"`
	Name             string              `json:"name,omitempty"`
	Type             string              `json:"type,omitempty"`
	State            string              `json:"state,omitempty"`
	Title            LocalizedText       `json:"title,omitempty"`
	Description      LocalizedText       `json:"description,omitempty"`
	Publisher        Publisher           `json:"publisher,omitempty"`
	Jurisdiction     string              `json:"jurisdiction,omitempty"`
	LicenseID        string              `json:"license_id,omitempty"`
	LicenseTitle     string              `json:"license_title,omitempty"`
	LicenseURL       string              `json:"license_url,omitempty"`
	UpdateFrequency  string              `json:"update_frequency,omitempty"`
	URL              string              `json:"url,omitempty"`
	Private          bool                `json:"private"`
	Open             bool                `json:"open"`
	Keywords         map[string][]string `json:"keywords,omitempty"`
	Tags             []string            `json:"tags,omitempty"`
	Subjects         []string            `json:"subjects,omitempty"`
	MetadataCreated  *time.Time          `json:"metadata_created,omitempty"`
	MetadataModified *time.Time          `json:"metadata_modified,omitempty"`
	Resources        []Resource          `json:"resources,omitempty"`
	Relevance        Relevance           `json:"relevance"`
}

// SearchPage represents one package_search response.
type SearchPage struct {
	Total        int                      `json:"total"`
	Start        int                      `json:"start"`
	Rows         int                      `json:"rows"`
	Sort         string                   `json:"sort,omitempty"`
	Datasets     []Dataset                `json:"datasets,omitempty"`
	Checkpoint   safefetch.Checkpoint     `json:"checkpoint,omitempty"`
	NotModified  bool                     `json:"not_modified"`
}

// DiscoveryResult combines bounded package_search pages.
type DiscoveryResult struct {
	TotalAvailable int                              `json:"total_available"`
	Pages          int                              `json:"pages"`
	Datasets       []Dataset                        `json:"datasets,omitempty"`
	Checkpoints    map[int]safefetch.Checkpoint     `json:"checkpoints,omitempty"`
}

// ChangeKind normalizes common CKAN activity names without treating them as
// graph events yet.
type ChangeKind string

const (
	ChangeUnknown  ChangeKind = "UNKNOWN"
	ChangeCreated  ChangeKind = "DATASET_CREATED"
	ChangeUpdated  ChangeKind = "DATASET_CHANGED"
	ChangeDeleted  ChangeKind = "DATASET_DELETED"
)

// Change is one recently_changed_packages_activity_list entry.
type Change struct {
	ID               string      `json:"id"`
	RemoteActivityID string      `json:"remote_activity_id"`
	DatasetID        string      `json:"dataset_id"`
	RemoteDatasetID  string      `json:"remote_dataset_id"`
	ActivityType     string      `json:"activity_type"`
	Kind             ChangeKind  `json:"kind"`
	Timestamp        time.Time   `json:"timestamp"`
	Dataset          *Dataset    `json:"dataset,omitempty"`
}

// ChangePage carries a changed-activity checkpoint independently of catalog
// activation state.
type ChangePage struct {
	Offset       int                  `json:"offset"`
	Limit        int                  `json:"limit"`
	Changes      []Change             `json:"changes,omitempty"`
	Checkpoint   safefetch.Checkpoint `json:"checkpoint,omitempty"`
	NotModified  bool                 `json:"not_modified"`
}
