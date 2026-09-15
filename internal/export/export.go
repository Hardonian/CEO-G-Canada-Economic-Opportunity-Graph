package export

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/publication"
)

// ProjectExportBundle aggregates a project with its full provenance, events, and scores.
type ProjectExportBundle struct {
	Project       *domain.Project        `json:"project"`
	Scores        map[string]float64     `json:"scores"`
	Events        []*domain.Event        `json:"events"`
	CapitalItems  []*domain.CapitalItem  `json:"capital_items"`
	Procurements  []*domain.Procurement  `json:"procurements"`
	Relationships []*domain.Relationship `json:"relationships"`
	ScoreHistory  []*domain.ProjectScore `json:"score_history"`
	Opportunities []*domain.Opportunity  `json:"opportunities"`
	CapitalNeeds  []*domain.CapitalNeed  `json:"capital_needs"`
	CapitalRequirements []*domain.CapitalRequirement `json:"capital_requirements"`
	Milestones    []*domain.Milestone    `json:"milestones"`
	Readiness     *domain.ReadinessAssessment `json:"readiness,omitempty"`
	Claims        []*domain.Claim        `json:"claims"`
	Evidence      []*domain.Evidence     `json:"evidence"`
	ExportedAt    time.Time              `json:"exported_at"`
}

// ExportProjectBundle creates an investor-grade dossier with full evidence backing.
func ExportProjectBundle(ctx context.Context, store database.Store, projectID string) (*ProjectExportBundle, error) {
	proj, err := store.GetProject(ctx, projectID)
	if err != nil {
		proj, err = store.GetProjectBySlug(ctx, projectID)
		if err != nil {
			return nil, err
		}
	}
	projectID = proj.ID

	events, err := store.ListEventsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	capital, err := store.ListCapitalItemsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	procs, err := store.ListProcurementsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	relationships, err := store.ListRelationshipsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	opps, err := store.ListOpportunitiesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	capitalNeeds, err := store.ListCapitalNeeds(ctx, projectID)
	if err != nil { return nil, err }
	capitalRequirements, err := store.ListCapitalRequirements(ctx, projectID)
	if err != nil { return nil, err }
	milestones, err := store.ListMilestones(ctx, projectID)
	if err != nil { return nil, err }
	claims, err := store.ListClaimsBySubject(ctx, projectID)
	if err != nil { return nil, err }
	readiness, _ := store.GetLatestReadinessAssessment(ctx, projectID)
	scores, err := store.GetLatestScores(ctx, projectID)
	if err != nil {
		return nil, err
	}
	scoreHistory, err := store.ListScoreHistory(ctx, projectID, "")
	if err != nil {
		return nil, err
	}

	scoreMap := make(map[string]float64)
	for k, v := range scores {
		scoreMap[k] = v.ScoreValue
	}

	evidenceMap := make(map[string]*domain.Evidence)
	publicEvidence := func(evidenceID string) bool {
		if evidenceID == "" { return false }
		e, getErr := store.GetEvidence(ctx, evidenceID)
		if getErr != nil || !publication.PublicEvidence(e) { return false }
		evidenceMap[e.ID] = e
		return true
	}
	for _, evidenceID := range proj.EvidenceIDs {
		publicEvidence(evidenceID)
	}
	publicEvents := make([]*domain.Event, 0, len(events))
	for _, ev := range events {
		if publicEvidence(ev.EvidenceID) { publicEvents = append(publicEvents, ev) }
	}
	publicCapital := make([]*domain.CapitalItem, 0, len(capital))
	for _, c := range capital {
		if publication.PublicCapitalItem(c) && publicEvidence(c.EvidenceID) { publicCapital = append(publicCapital, c) }
	}
	publicProcs := make([]*domain.Procurement, 0, len(procs))
	for _, procurement := range procs {
		if publicEvidence(procurement.EvidenceID) { publicProcs = append(publicProcs, procurement) }
	}
	publicRelationships := make([]*domain.Relationship, 0, len(relationships))
	for _, relationship := range relationships {
		if publicEvidence(relationship.EvidenceID) { publicRelationships = append(publicRelationships, relationship) }
	}
	publicScoreHistory := make([]*domain.ProjectScore, 0, len(scoreHistory))
	for _, score := range scoreHistory {
		publishScore := len(score.EvidenceIDs) > 0
		for _, evidenceID := range score.EvidenceIDs {
			if !publicEvidence(evidenceID) { publishScore = false }
		}
		if publishScore { publicScoreHistory = append(publicScoreHistory, score) }
	}
	publicOpps := make([]*domain.Opportunity, 0, len(opps))
	for _, opportunity := range opps {
		if !publication.PublicOpportunity(opportunity) { continue }
		valid := true
		for _, evidenceID := range opportunity.EvidenceIDs { if !publicEvidence(evidenceID) { valid = false } }
		// Ontology-derived opportunities can be public without their own evidence,
		// provided they are explicitly labelled DERIVED rather than a tender.
		if valid && (len(opportunity.EvidenceIDs) > 0 || opportunity.RequirementClass == domain.RequirementDerived) { publicOpps = append(publicOpps, opportunity) }
	}
	publicNeeds := make([]*domain.CapitalNeed, 0, len(capitalNeeds))
	for _, need := range capitalNeeds {
		if !publication.PublicCapitalNeed(need) { continue }
		valid := len(need.EvidenceIDs) > 0
		for _, evidenceID := range need.EvidenceIDs { if !publicEvidence(evidenceID) { valid = false } }
		if valid { publicNeeds = append(publicNeeds, need) }
	}
	publicRequirements := make([]*domain.CapitalRequirement, 0, len(capitalRequirements))
	for _, requirement := range capitalRequirements {
		if !publication.PublicCapitalRequirement(requirement) { continue }
		valid := len(requirement.EvidenceIDs) > 0
		for _, evidenceID := range requirement.EvidenceIDs { if !publicEvidence(evidenceID) { valid = false } }
		if valid { publicRequirements = append(publicRequirements, requirement) }
	}
	publicMilestones := make([]*domain.Milestone, 0, len(milestones))
	for _, milestone := range milestones {
		if !publication.PublicMilestone(milestone) { continue }
		valid := len(milestone.EvidenceIDs) > 0
		for _, evidenceID := range milestone.EvidenceIDs { if !publicEvidence(evidenceID) { valid = false } }
		if valid { publicMilestones = append(publicMilestones, milestone) }
	}
	publicClaims := make([]*domain.Claim, 0, len(claims))
	for _, claim := range claims {
		if publication.PublicClaim(claim) && publicEvidence(claim.EvidenceID) { publicClaims = append(publicClaims, claim) }
	}

	var evidenceList []*domain.Evidence
	for _, e := range evidenceMap {
		evidenceList = append(evidenceList, e)
	}
	sort.Slice(evidenceList, func(i, j int) bool { return evidenceList[i].ID < evidenceList[j].ID })

	projectCopy := *proj
	projectCopy.EvidenceIDs = make([]string, 0, len(evidenceMap))
	for evidenceID := range evidenceMap { projectCopy.EvidenceIDs = append(projectCopy.EvidenceIDs, evidenceID) }
	sort.Strings(projectCopy.EvidenceIDs)
	return &ProjectExportBundle{
		Project:       &projectCopy,
		Scores:        scoreMap,
		Events:        publicEvents,
		CapitalItems:  publicCapital,
		Procurements:  publicProcs,
		Relationships: publicRelationships,
		ScoreHistory:  publicScoreHistory,
		Opportunities: publicOpps,
		CapitalNeeds:  publicNeeds,
		CapitalRequirements: publicRequirements,
		Milestones:    publicMilestones,
		Readiness:     readiness,
		Claims:        publicClaims,
		Evidence:      evidenceList,
		ExportedAt:    time.Now(),
	}, nil
}

// ToMarkdown converts a project bundle into investor-grade Markdown.
func (b *ProjectExportBundle) ToMarkdown() string {
	var sb strings.Builder
	p := b.Project

	sb.WriteString(fmt.Sprintf("# %s\n\n", p.Name))
	sb.WriteString(fmt.Sprintf("**Sector:** %s | **Subsector:** %s | **Province:** %s\n", p.Sector, p.Subsector, p.Province))
	capex := "UNKNOWN"
	if p.CapexCAD > 0 && p.CapexStatus != domain.ConfidenceUnknown {
		capex = fmt.Sprintf("$%d CAD (%s)", p.CapexCAD, p.CapexStatus)
	}
	sb.WriteString(fmt.Sprintf("**Current Stage:** `%s` | **CAPEX:** %s\n", p.CurrentStage, capex))
	sb.WriteString(fmt.Sprintf("**Location:** %s\n\n", p.LocationName))

	sb.WriteString("## Executive Summary\n\n")
	sb.WriteString(p.Summary + "\n\n")

	sb.WriteString("## Deterministic Scores\n\n")
	sb.WriteString("| Dimension | Score (0-100) |\n| :--- | :--- |\n")
	for k, v := range b.Scores {
		sb.WriteString(fmt.Sprintf("| %s | %.1f |\n", title(k), v))
	}
	sb.WriteString("\n")

	if len(b.CapitalItems) > 0 {
		sb.WriteString("## Tracked Capital Events\n\n")
		sb.WriteString("| Category | Status | Amount (CAD) | Provider |\n| :--- | :--- | :--- | :--- |\n")
		for _, c := range b.CapitalItems {
			sb.WriteString(fmt.Sprintf("| %s | %s | $%d | %s |\n", c.Category, c.Status, c.AmountCAD, c.ProviderName))
		}
		sb.WriteString("\n")
	}

	if len(b.Opportunities) > 0 {
		sb.WriteString("## Downstream Procurement & Supply-Chain Opportunities\n\n")
		sb.WriteString("| Requirement Class | Category | Title | Est. CAD |\n| :--- | :--- | :--- | :--- |\n")
		for _, o := range b.Opportunities {
			estimate := "UNKNOWN"
			if o.EstimateStatus != domain.ConfidenceUnknown && o.EstimatedCAD > 0 {
				estimate = fmt.Sprintf("$%d", o.EstimatedCAD)
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", o.RequirementClass, o.Category, o.Title, estimate))
		}
		sb.WriteString("\n")
	}

	if len(b.Evidence) > 0 {
		sb.WriteString("## Provenance & Sourced Evidence\n\n")
		sb.WriteString("| Publisher | Tier | Source URL | Content Hash (SHA-256) |\n| :--- | :--- | :--- | :--- |\n")
		for _, ev := range b.Evidence {
			shortHash := ev.ContentHash
			if len(shortHash) > 16 {
				shortHash = shortHash[:16] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | Tier %d | [%s](%s) | `%s` |\n", ev.Publisher, ev.SourceTier, ev.SourceURL, ev.SourceURL, shortHash))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("---\n*Export generated by CanadaOpportunityGraph on %s. CEGS Standard v0.1 compliant.*\n", b.ExportedAt.Format(time.RFC3339)))
	return sb.String()
}

// ToGeoJSON converts a project list into GeoJSON FeatureCollection.
func ToGeoJSON(projects []*domain.Project) ([]byte, error) {
	type Feature struct {
		Type       string                 `json:"type"`
		Geometry   map[string]interface{} `json:"geometry"`
		Properties map[string]interface{} `json:"properties"`
	}
	type FeatureCollection struct {
		Type     string    `json:"type"`
		Features []Feature `json:"features"`
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]Feature, 0, len(projects)),
	}

	for _, p := range projects {
		if p.Latitude == 0 && p.Longitude == 0 {
			continue
		}
		feat := Feature{
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{p.Longitude, p.Latitude},
			},
			Properties: map[string]interface{}{
				"id":            p.ID,
				"slug":          p.Slug,
				"name":          p.Name,
				"sector":        p.Sector,
				"province":      p.Province,
				"current_stage": p.CurrentStage,
				"capex_cad":     p.CapexCAD,
				"scores":        p.Scores,
			},
		}
		fc.Features = append(fc.Features, feat)
	}

	return json.MarshalIndent(fc, "", "  ")
}

// ToCSV converts a project list into standard CSV format.
func ToCSV(projects []*domain.Project) (string, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)

	headers := []string{"ID", "Slug", "Name", "Sector", "Subsector", "Province", "Location", "Stage", "CapexCAD", "CapexStatus", "Buildability", "BuildabilityCoverage", "Confidence"}
	if err := w.Write(headers); err != nil {
		return "", err
	}

	for _, p := range projects {
		bScore := 0.0
		if p.Scores != nil {
			bScore = p.Scores["buildability"]
		}
		capex := ""
		if p.CapexCAD > 0 && p.CapexStatus != domain.ConfidenceUnknown {
			capex = fmt.Sprintf("%d", p.CapexCAD)
		}
		coverage := ""
		for _, detail := range p.ScoreDetails {
			if detail.ScoreType == "buildability" {
				coverage = fmt.Sprintf("%.1f", detail.Coverage)
			}
		}

		row := []string{
			p.ID,
			p.Slug,
			p.Name,
			string(p.Sector),
			p.Subsector,
			p.Province,
			p.LocationName,
			string(p.CurrentStage),
			capex,
			string(p.CapexStatus),
			fmt.Sprintf("%.1f", bScore),
			coverage,
			string(p.Confidence),
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	w.Flush()
	return sb.String(), nil
}

func title(value string) string {
	parts := strings.Fields(strings.ReplaceAll(value, "_", " "))
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

// ToCEGSExport converts a project bundle to canonical CEGS Project format.
func (b *ProjectExportBundle) ToCEGSExport() (*cegs.Project, error) {
	var evidenceIDs []string
	for _, ev := range b.Evidence {
		evidenceIDs = append(evidenceIDs, cegs.FormatID("evidence", "ca", ev.ID))
	}
	return cegs.ToCEGSProject(b.Project, evidenceIDs), nil
}
