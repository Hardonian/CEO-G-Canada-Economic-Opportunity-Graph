package cegs

import (
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// ProjectID returns the canonical CEGS identifier for a domain project. CEGS
// references must use this helper rather than reconstructing IDs from slugs or
// raw UUIDs, because the project jurisdiction is part of the identifier.
func ProjectID(p *domain.Project) string {
	jurisdiction := "ca"
	if p != nil && p.Province != "" && p.Province != "Federal" {
		jurisdiction = "ca:" + strings.ToLower(p.Province)
	}
	if p == nil {
		return ""
	}
	return FormatID("project", jurisdiction, p.ID)
}

// OrganizationID returns the canonical CEGS identifier for a domain entity.
// Domain UUIDs, rather than display slugs, keep references unique when two
// legal entities normalize to the same human-readable slug.
func OrganizationID(e *domain.Entity) string {
	if e == nil {
		return ""
	}
	return FormatID("org", "ca", e.ID)
}

// EvidenceID returns the canonical CEGS identifier for a domain evidence ID.
func EvidenceID(id string) string {
	if id == "" || strings.HasPrefix(id, "cegs:") {
		return id
	}
	return FormatID("evidence", "ca", id)
}

// ToCEGSProject converts an internal domain Project into a canonical CEGS Project resource.
func ToCEGSProject(p *domain.Project, evidenceIDs []string) *Project {
	jur := "ca"
	if p.Province != "" && p.Province != "Federal" {
		jur = "ca:" + strings.ToLower(p.Province)
	}

	cegsID := ProjectID(p)

	var proponents []string
	if p.ProponentID != "" {
		proponents = append(proponents, FormatID("org", "ca", p.ProponentID))
	} else if p.Proponent != nil {
		proponents = append(proponents, OrganizationID(p.Proponent))
	}
	provenance := make([]string, 0, len(evidenceIDs))
	for _, evidenceID := range evidenceIDs {
		if normalized := EvidenceID(evidenceID); normalized != "" {
			provenance = append(provenance, normalized)
		}
	}

	ext := make(map[string]interface{})
	if p.Scores != nil {
		ext["ca.opengraph.scores"] = p.Scores
	}

	amountType := "unknown"
	if p.CapexCAD > 0 {
		if p.CapexStatus == domain.ConfidenceReported || p.CapexStatus == domain.ConfidenceVerified || p.CapexStatus == domain.ConfidenceSupported {
			amountType = "reported"
		}
		if p.CapexStatus == domain.ConfidenceInferred {
			amountType = "estimated"
		}
	}
	location := Location{Name: p.LocationName, Province: p.Province}
	if p.Latitude != 0 || p.Longitude != 0 {
		lat, lon := p.Latitude, p.Longitude
		location.Latitude, location.Longitude = &lat, &lon
	}

	return &Project{
		Envelope: Envelope{
			CEGS:          SpecVersion,
			ID:            cegsID,
			Type:          "project",
			CanonicalName: p.Name,
			Jurisdiction:  strings.ToUpper(jur),
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
			Provenance:    provenance,
			Extensions:    ext,
		},
		Description: p.Summary,
		Sector:      string(p.Sector),
		Subsector:   p.Subsector,
		Stage:       string(p.CurrentStage),
		Capex: Monetary{
			Amount:     p.CapexCAD,
			Currency:   "CAD",
			AmountType: amountType,
		},
		Proponents:   proponents,
		Location:     location,
		SourceStatus: string(p.Confidence),
	}
}

// ToCEGSOrganization converts an internal Entity into a CEGS Organization.
func ToCEGSOrganization(e *domain.Entity) *Organization {
	jur := "CA"
	if e.Jurisdiction != "" {
		jur = strings.ToUpper(e.Jurisdiction)
		if !strings.HasPrefix(jur, "CA") {
			jur = "CA:" + jur
		}
	}

	cegsID := OrganizationID(e)

	var prov []string
	if e.EvidenceID != "" {
		prov = append(prov, EvidenceID(e.EvidenceID))
	}

	return &Organization{
		Envelope: Envelope{
			CEGS:          SpecVersion,
			ID:            cegsID,
			Type:          "organization",
			CanonicalName: e.CommonName,
			Jurisdiction:  jur,
			CreatedAt:     e.CreatedAt,
			UpdatedAt:     e.UpdatedAt,
			Provenance:    prov,
			Extensions:    make(map[string]interface{}),
		},
		LegalName:   e.LegalName,
		Aliases:     e.Aliases,
		EntityType:  e.EntityType,
		Website:     e.Website,
		Identifiers: e.Identifiers,
		Description: e.Description,
	}
}

// ToCEGSEvent converts an internal Event into a CEGS Event whose subject is
// guaranteed to match the canonical ID emitted for the supplied project.
func ToCEGSEvent(ev *domain.Event, project *domain.Project) *Event {
	subj := ProjectID(project)
	evID := FormatID("event", "ca", ev.ID)

	var prov []string
	if ev.EvidenceID != "" {
		prov = append(prov, EvidenceID(ev.EvidenceID))
	}

	attrs := make(map[string]interface{})
	if ev.PreviousStage != nil {
		attrs["previous_stage"] = *ev.PreviousStage
	}
	if ev.NewStage != nil {
		attrs["new_stage"] = *ev.NewStage
	}

	return &Event{
		CEGS:        SpecVersion,
		ID:          evID,
		Type:        "event",
		EventType:   ev.EventType,
		Subject:     subj,
		OccurredAt:  ev.EventDate,
		Title:       ev.Title,
		Description: ev.Description,
		Evidence:    prov,
		Attributes:  attrs,
		CreatedAt:   ev.CreatedAt,
	}
}

// ToCEGSEvidence converts an internal Evidence into a CEGS Evidence.
func ToCEGSEvidence(ev *domain.Evidence) *Evidence {
	evID := EvidenceID(ev.ID)
	lineage := make(map[string]interface{})
	for key, value := range map[string]string{
		"source_id":         ev.SourceID,
		"source_version_id": ev.SourceVersionID,
		"source_record_id":  ev.SourceRecordID,
		"locator":           ev.Locator,
		"parser_version":    ev.ParserVersion,
		"mapping_version":   ev.MappingVersion,
		"pipeline_version":  ev.PipelineVersion,
	} {
		if value != "" {
			lineage[key] = value
		}
	}
	extensions := make(map[string]interface{})
	if len(lineage) > 0 {
		extensions["ca.opengraph.public_data_lineage"] = lineage
	}

	return &Evidence{
		CEGS:               SpecVersion,
		ID:                 evID,
		Type:               "evidence",
		SourceURL:          ev.SourceURL,
		Publisher:          ev.Publisher,
		SourceTier:         int(ev.SourceTier),
		ContentHash:        ev.ContentHash,
		RetrievalTimestamp: ev.RetrievalTimestamp,
		PublicationDate:    ev.PublicationDate,
		EffectiveDate:      ev.EffectiveDate,
		Confidence:         string(ev.Confidence),
		ExtractionMethod:   ev.ExtractionMethod,
		RawSnippet:         ev.RawSnippet,
		Extensions:         extensions,
	}
}

// ToCEGSRelationship converts an internal Relationship into a CEGS
// Relationship using the same canonical entity and project IDs as their
// standalone resources.
func ToCEGSRelationship(r *domain.Relationship, source *domain.Entity, target *domain.Project) *Relationship {
	relID := FormatID("rel", "ca", r.ID)
	from := OrganizationID(source)
	to := ProjectID(target)

	var prov []string
	if r.EvidenceID != "" {
		prov = append(prov, EvidenceID(r.EvidenceID))
	}

	return &Relationship{
		CEGS:             SpecVersion,
		ID:               relID,
		Type:             "relationship",
		RelationshipType: r.RelationType,
		From:             from,
		To:               to,
		Status:           string(r.Confidence),
		ValidFrom:        r.CreatedAt,
		Evidence:         prov,
		CreatedAt:        r.CreatedAt,
	}
}
