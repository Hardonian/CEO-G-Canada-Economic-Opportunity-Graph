package cegs

import (
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// ToCEGSProject converts an internal domain Project into a canonical CEGS Project resource.
func ToCEGSProject(p *domain.Project, evidenceIDs []string) *Project {
	jur := "ca"
	if p.Province != "" && p.Province != "Federal" {
		jur = "ca:" + strings.ToLower(p.Province)
	}

	cegsID := FormatID("project", jur, p.ID)

	var proponents []string
	if p.Proponent != nil {
		proponents = append(proponents, FormatID("org", "ca", p.Proponent.ID))
	}
	for i, evidenceID := range evidenceIDs {
		if !strings.HasPrefix(evidenceID, "cegs:") { evidenceIDs[i] = FormatID("evidence", "ca", evidenceID) }
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
			Provenance:    evidenceIDs,
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

	cegsID := FormatID("org", "ca", e.Slug)

	var prov []string
	if e.EvidenceID != "" {
		prov = append(prov, FormatID("evidence", "ca", e.EvidenceID))
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

// ToCEGSEvent converts an internal Event into a CEGS Event.
func ToCEGSEvent(ev *domain.Event, projectSlug string) *Event {
	subj := FormatID("project", "ca", projectSlug)
	evID := FormatID("event", "ca", ev.ID)

	var prov []string
	if ev.EvidenceID != "" {
		prov = append(prov, FormatID("evidence", "ca", ev.EvidenceID))
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
	evID := FormatID("evidence", "ca", ev.ID)

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
	}
}

// ToCEGSRelationship converts an internal Relationship into a CEGS Relationship.
func ToCEGSRelationship(r *domain.Relationship, sourceSlug, targetSlug string) *Relationship {
	relID := FormatID("rel", "ca", r.ID)
	from := FormatID("org", "ca", sourceSlug)
	to := FormatID("project", "ca", targetSlug)

	var prov []string
	if r.EvidenceID != "" {
		prov = append(prov, FormatID("evidence", "ca", r.EvidenceID))
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
