package ckan

import (
	"net/url"
	"path"
	"strings"
	"unicode"
)

// ClassifyResource selects the most specific reusable protocol family from
// CKAN format metadata, MIME type, resource type, and URL structure.
func ClassifyResource(format, mimeType, resourceType, rawURL string) Family {
	declared := normalizeClassifierText(format + " " + mimeType + " " + resourceType)
	lowerURL := strings.ToLower(rawURL)
	parsed, _ := url.Parse(rawURL)
	lowerPath := strings.ToLower(parsed.Path)
	query := parsed.Query()
	service := strings.ToLower(query.Get("service"))

	switch {
	case strings.Contains(lowerURL, "/featureserver"):
		return FamilyArcGISFeature
	case strings.Contains(lowerURL, "/mapserver"):
		return FamilyArcGISMap
	case strings.Contains(lowerURL, "openapi") || strings.Contains(lowerURL, "swagger") || hasToken(declared, "openapi") || hasToken(declared, "swagger"):
		return FamilyOpenAPI
	case strings.Contains(lowerURL, "graphql") || hasToken(declared, "graphql"):
		return FamilyGraphQL
	case strings.Contains(lowerURL, "/api/3/action/") || hasToken(declared, "ckan"):
		return FamilyCKAN
	case service == "wfs" || hasToken(declared, "wfs"):
		return FamilyWFS
	case service == "wms" || hasToken(declared, "wms"):
		return FamilyWMS
	case service == "wmts" || hasToken(declared, "wmts"):
		return FamilyWMTS
	case strings.Contains(lowerURL, "sparql") || hasToken(declared, "sparql"):
		return FamilySPARQL
	case strings.Contains(lowerURL, "sdmx") || hasToken(declared, "sdmx"):
		return FamilySDMX
	case strings.Contains(lowerURL, "socrata") || strings.Contains(lowerURL, "/resource/") && strings.Contains(lowerURL, ".json") || hasToken(declared, "socrata"):
		return FamilySocrata
	case path.Ext(lowerPath) == ".zip" || hasToken(declared, "zip"):
		return FamilyZIP
	case path.Ext(lowerPath) == ".kmz" || hasToken(declared, "kmz"):
		return FamilyKMZ
	case path.Ext(lowerPath) == ".gpkg" || hasToken(declared, "geopackage") || hasToken(declared, "gpkg"):
		return FamilyGeoPackage
	case path.Ext(lowerPath) == ".shp" || hasToken(declared, "shapefile") || hasToken(declared, "shp"):
		return FamilyShapefile
	case path.Ext(lowerPath) == ".kml" || hasToken(declared, "kml"):
		return FamilyKML
	case hasToken(declared, "geojson") || strings.HasSuffix(lowerPath, ".geojson"):
		return FamilyGeoJSON
	case hasToken(declared, "jsonld") || hasToken(declared, "json ld") || strings.HasSuffix(lowerPath, ".jsonld"):
		return FamilyJSONLD
	case hasToken(declared, "jsonl") || hasToken(declared, "ndjson") || strings.HasSuffix(lowerPath, ".jsonl") || strings.HasSuffix(lowerPath, ".ndjson"):
		return FamilyJSONL
	case hasToken(declared, "dcat"):
		return FamilyDCAT
	case hasToken(declared, "rdf") || strings.HasSuffix(lowerPath, ".rdf") || strings.HasSuffix(lowerPath, ".ttl"):
		return FamilyRDF
	case hasToken(declared, "rss") || strings.HasSuffix(lowerPath, ".rss"):
		return FamilyRSS
	case hasToken(declared, "atom") || strings.HasSuffix(lowerPath, ".atom"):
		return FamilyAtom
	case hasToken(declared, "csv") || strings.HasSuffix(lowerPath, ".csv"):
		return FamilyCSV
	case hasToken(declared, "tsv") || hasToken(declared, "tab separated") || strings.HasSuffix(lowerPath, ".tsv"):
		return FamilyTSV
	case hasToken(declared, "xlsx") || strings.HasSuffix(lowerPath, ".xlsx"):
		return FamilyXLSX
	case hasToken(declared, "xls") || strings.HasSuffix(lowerPath, ".xls"):
		return FamilyXLS
	case hasToken(declared, "pdf") || strings.HasSuffix(lowerPath, ".pdf"):
		return FamilyPDF
	case hasToken(declared, "xml") || strings.HasSuffix(lowerPath, ".xml"):
		return FamilyXML
	case hasToken(declared, "json") || strings.HasSuffix(lowerPath, ".json"):
		return FamilyJSON
	case hasToken(declared, "git") || strings.Contains(lowerURL, "github.com/") || strings.Contains(lowerURL, "gitlab.com/"):
		return FamilyGit
	case hasToken(declared, "html") || strings.HasSuffix(lowerPath, ".html") || strings.HasSuffix(lowerPath, ".htm"):
		return FamilyHTML
	case hasToken(declared, "api"):
		return FamilyRESTAPI
	default:
		return FamilyUnknown
	}
}

func classifyArchiveContents(format, mimeType, resourceType, rawURL string) Family {
	if ClassifyResource(format, mimeType, resourceType, rawURL) != FamilyZIP {
		return FamilyUnknown
	}
	// Ignore the .zip suffix so the declared format or an inner extension can
	// identify what a bulk archive contains.
	innerURL := strings.TrimSuffix(rawURL, path.Ext(rawURL))
	declared := normalizeClassifierText(format)
	if hasToken(declared, "zip") {
		declared = strings.TrimSpace(strings.ReplaceAll(declared, "zip", ""))
	}
	return ClassifyResource(declared, mimeType, resourceType, innerURL)
}

func normalizeClassifierText(value string) string {
	fields := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(value)), func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	return strings.Join(fields, " ")
}

func hasToken(normalized, token string) bool {
	token = normalizeClassifierText(token)
	if normalized == token {
		return true
	}
	return strings.Contains(" "+normalized+" ", " "+token+" ")
}
