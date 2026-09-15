package reconciliation

import (
	"fmt"
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const (
	MethodologyVersion = "reconciliation-v1.0"
)

type JurisdictionLevel string

const (
	JurisdictionFederal    JurisdictionLevel = "FEDERAL"
	JurisdictionProvincial JurisdictionLevel = "PROVINCIAL"
	JurisdictionMunicipal  JurisdictionLevel = "MUNICIPAL"
	JurisdictionIndigenous JurisdictionLevel = "INDIGENOUS"
)

type ReconciliationAction string

const (
	ActionMerge        ReconciliationAction = "MERGE"
	ActionLink         ReconciliationAction = "LINK"
	ActionConflict     ReconciliationAction = "CONFLICT"
	ActionNoMatch      ReconciliationAction = "NO_MATCH"
	ActionPending      ReconciliationAction = "PENDING"
)

type MatchConfidence string

const (
	ConfidenceHigh   MatchConfidence = "HIGH"
	ConfidenceMedium MatchConfidence = "MEDIUM"
	ConfidenceLow    MatchConfidence = "LOW"
)

type JurisdictionRecord struct {
	ProjectID       string            `json:"project_id"`
	Jurisdiction    string            `json:"jurisdiction"`
	Level           JurisdictionLevel `json:"level"`
	ExternalID      string            `json:"external_id"`
	SourceURL       string            `json:"source_url"`
	Stage           domain.LifecycleStage `json:"stage"`
	Title           string            `json:"title"`
	Summary         string            `json:"summary"`
	Location        string            `json:"location"`
	CapexCAD        int64             `json:"capex_cad"`
	EffectiveDate   string            `json:"effective_date"`
	Confidence      domain.ConfidenceLevel `json:"confidence"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

type ReconciliationMatch struct {
	FederalRecord     *JurisdictionRecord `json:"federal_record,omitempty"`
	ProvincialRecord  *JurisdictionRecord `json:"provincial_record,omitempty"`
	MunicipalRecord   *JurisdictionRecord `json:"municipal_record,omitempty"`
	IndigenousRecord  *JurisdictionRecord `json:"indigenous_record,omitempty"`
	Action            ReconciliationAction `json:"action"`
	Confidence        MatchConfidence      `json:"confidence"`
	Rationale         string              `json:"rationale"`
	Conflicts         []string            `json:"conflicts,omitempty"`
	MergedProjectID   string              `json:"merged_project_id,omitempty"`
}

type ReconciliationReport struct {
	MethodologyVersion string                   `json:"methodology_version"`
	TotalRecords       int                      `json:"total_records"`
	Matches            []*ReconciliationMatch   `json:"matches"`
	UnmatchedFederal   []*JurisdictionRecord    `json:"unmatched_federal,omitempty"`
	UnmatchedProvincial []*JurisdictionRecord   `json:"unmatched_provincial,omitempty"`
	UnmatchedMunicipal  []*JurisdictionRecord   `json:"unmatched_municipal,omitempty"`
	UnmatchedIndigenous []*JurisdictionRecord   `json:"unmatched_indigenous,omitempty"`
	Summary            ReconciliationSummary    `json:"summary"`
}

type ReconciliationSummary struct {
	Merged      int `json:"merged"`
	Linked      int `json:"linked"`
	Conflicts   int `json:"conflicts"`
	NoMatch     int `json:"no_match"`
	Pending     int `json:"pending"`
}

// ReconcileJurisdictions performs multi-jurisdiction reconciliation of project records.
func ReconcileJurisdictions(records []*JurisdictionRecord) *ReconciliationReport {
	report := &ReconciliationReport{
		MethodologyVersion: MethodologyVersion,
		TotalRecords:       len(records),
		Matches:            make([]*ReconciliationMatch, 0),
	}
	
	// Group records by level
	federal := make([]*JurisdictionRecord, 0)
	provincial := make([]*JurisdictionRecord, 0)
	municipal := make([]*JurisdictionRecord, 0)
	indigenous := make([]*JurisdictionRecord, 0)
	
	for _, r := range records {
		switch r.Level {
		case JurisdictionFederal:
			federal = append(federal, r)
		case JurisdictionProvincial:
			provincial = append(provincial, r)
		case JurisdictionMunicipal:
			municipal = append(municipal, r)
		case JurisdictionIndigenous:
			indigenous = append(indigenous, r)
		}
	}
	
	matchedProvincial := make(map[string]bool)
	matchedMunicipal := make(map[string]bool)
	matchedIndigenous := make(map[string]bool)
	
	// For each federal record, find best matches
	for _, fed := range federal {
		match := &ReconciliationMatch{
			FederalRecord: fed,
			Action:        ActionNoMatch,
			Confidence:    ConfidenceLow,
		}
		
		// Find provincial match
		bestProv := findBestMatch(fed, provincial, matchedProvincial)
		if bestProv != nil {
			match.ProvincialRecord = bestProv
			matchedProvincial[bestProv.ProjectID] = true
		}
		
		// Find municipal match
		bestMun := findBestMatch(fed, municipal, matchedMunicipal)
		if bestMun != nil {
			match.MunicipalRecord = bestMun
			matchedMunicipal[bestMun.ProjectID] = true
		}
		
		// Find indigenous match
		bestInd := findBestMatch(fed, indigenous, matchedIndigenous)
		if bestInd != nil {
			match.IndigenousRecord = bestInd
			matchedIndigenous[bestInd.ProjectID] = true
		}
		
		// Determine action and confidence
		determineAction(match)
		
		report.Matches = append(report.Matches, match)
	}
	
	// Collect unmatched
	for _, r := range provincial {
		if !matchedProvincial[r.ProjectID] {
			report.UnmatchedProvincial = append(report.UnmatchedProvincial, r)
		}
	}
	for _, r := range municipal {
		if !matchedMunicipal[r.ProjectID] {
			report.UnmatchedMunicipal = append(report.UnmatchedMunicipal, r)
		}
	}
	for _, r := range indigenous {
		if !matchedIndigenous[r.ProjectID] {
			report.UnmatchedIndigenous = append(report.UnmatchedIndigenous, r)
		}
	}
	
	// Build summary
	for _, m := range report.Matches {
		switch m.Action {
		case ActionMerge:
			report.Summary.Merged++
		case ActionLink:
			report.Summary.Linked++
		case ActionConflict:
			report.Summary.Conflicts++
		case ActionNoMatch:
			report.Summary.NoMatch++
		case ActionPending:
			report.Summary.Pending++
		}
	}
	// Count unmatched records as NoMatch so the summary is complete.
	report.Summary.NoMatch += len(report.UnmatchedProvincial) +
		len(report.UnmatchedMunicipal) + len(report.UnmatchedIndigenous)

	return report
}

func findBestMatch(base *JurisdictionRecord, candidates []*JurisdictionRecord, matched map[string]bool) *JurisdictionRecord {
	var best *JurisdictionRecord
	bestScore := 0.0
	
	for _, cand := range candidates {
		if matched[cand.ProjectID] {
			continue
		}
		
		score := calculateMatchScore(base, cand)
		if score > bestScore && score >= 0.6 {
			bestScore = score
			best = cand
		}
	}
	
	return best
}

func calculateMatchScore(a, b *JurisdictionRecord) float64 {
	score := 0.0
	
	// Name similarity (highest weight)
	nameScore := stringSimilarity(strings.ToLower(a.Title), strings.ToLower(b.Title))
	score += nameScore * 0.4
	
	// Location similarity
	locScore := stringSimilarity(strings.ToLower(a.Location), strings.ToLower(b.Location))
	score += locScore * 0.25
	
	// Proponent/Entity similarity
	proponentA := getProponent(a.Metadata)
	proponentB := getProponent(b.Metadata)
	if proponentA != "" && proponentB != "" {
		propScore := stringSimilarity(strings.ToLower(proponentA), strings.ToLower(proponentB))
		score += propScore * 0.2
	}
	
	// Stage compatibility
	if a.Stage == b.Stage {
		score += 0.1
	} else if isStageCompatible(a.Stage, b.Stage) {
		score += 0.05
	}
	
	// CAPEX similarity
	if a.CapexCAD > 0 && b.CapexCAD > 0 {
		capexRatio := float64(min(a.CapexCAD, b.CapexCAD)) / float64(max(a.CapexCAD, b.CapexCAD))
		score += capexRatio * 0.1
	}
	
	return score
}

func stringSimilarity(a, b string) float64 {
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	
	// Simple Jaccard similarity on words
	wordsA := strings.Fields(a)
	wordsB := strings.Fields(b)
	
	setA := make(map[string]bool)
	for _, w := range wordsA {
		if len(w) > 2 {
			setA[w] = true
		}
	}
	
	intersection := 0
	union := len(setA)
	for _, w := range wordsB {
		if len(w) > 2 {
			if setA[w] {
				intersection++
			} else {
				union++
			}
		}
	}
	
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

func getProponent(metadata map[string]interface{}) string {
	if metadata == nil {
		return ""
	}
	if val, ok := metadata["proponent_name"].(string); ok {
		return val
	}
	if val, ok := metadata["proponent"].(string); ok {
		return val
	}
	if val, ok := metadata["company"].(string); ok {
		return val
	}
	return ""
}

func isStageCompatible(a, b domain.LifecycleStage) bool {
	// Define compatible stage pairs
	earlyStages := map[domain.LifecycleStage]bool{
		domain.StageUnknown: true,
		domain.StageDiscovered: true,
		domain.StageAnnounced: true,
		domain.StageReferred: true,
		domain.StageEarlyDevelopment: true,
	}
	
	midStages := map[domain.LifecycleStage]bool{
		domain.StageFeasibility: true,
		domain.StageFinancing: true,
		domain.StageEnvironmentalReview: true,
		domain.StagePermitting: true,
		domain.StageProcurement: true,
		domain.StageFIDLikely: true,
		domain.StageFID: true,
	}
	
	lateStages := map[domain.LifecycleStage]bool{
		domain.StageConstruction: true,
		domain.StageCommissioning: true,
		domain.StageOperating: true,
	}
	
	riskStages := map[domain.LifecycleStage]bool{
		domain.StageDelayed: true,
		domain.StagePaused: true,
		domain.StageCancelled: true,
	}
	
	aEarly := earlyStages[a]
	bEarly := earlyStages[b]
	aMid := midStages[a]
	bMid := midStages[b]
	aLate := lateStages[a]
	bLate := lateStages[b]
	aRisk := riskStages[a]
	bRisk := riskStages[b]
	
	if aEarly && bEarly { return true }
	if aMid && bMid { return true }
	if aLate && bLate { return true }
	if aRisk && bRisk { return true }
	
	return false
}

func determineAction(match *ReconciliationMatch) {
	hasProvincial := match.ProvincialRecord != nil
	hasMunicipal := match.MunicipalRecord != nil
	hasIndigenous := match.IndigenousRecord != nil
	
	conflicts := []string{}
	
	// Check for conflicts
	if hasProvincial && hasMunicipal {
		if match.ProvincialRecord.Stage != match.MunicipalRecord.Stage {
			conflicts = append(conflicts, fmt.Sprintf("Stage mismatch: provincial=%s municipal=%s", 
				match.ProvincialRecord.Stage, match.MunicipalRecord.Stage))
		}
	}
	
	if hasFederal := match.FederalRecord != nil; hasFederal {
		if hasProvincial && match.FederalRecord.Stage != match.ProvincialRecord.Stage {
			conflicts = append(conflicts, fmt.Sprintf("Stage mismatch: federal=%s provincial=%s", 
				match.FederalRecord.Stage, match.ProvincialRecord.Stage))
		}
		if hasIndigenous && match.FederalRecord.Stage != match.IndigenousRecord.Stage {
			conflicts = append(conflicts, fmt.Sprintf("Stage mismatch: federal=%s indigenous=%s", 
				match.FederalRecord.Stage, match.IndigenousRecord.Stage))
		}
	}
	
	// CAPEX conflicts
	if hasProvincial && match.FederalRecord.CapexCAD > 0 && match.ProvincialRecord.CapexCAD > 0 {
		diff := float64(max(match.FederalRecord.CapexCAD, match.ProvincialRecord.CapexCAD) - min(match.FederalRecord.CapexCAD, match.ProvincialRecord.CapexCAD))
		ratio := diff / float64(max(match.FederalRecord.CapexCAD, match.ProvincialRecord.CapexCAD))
		if ratio > 0.25 {
			conflicts = append(conflicts, fmt.Sprintf("CAPEX discrepancy >25%%: federal=%d provincial=%d", 
				match.FederalRecord.CapexCAD, match.ProvincialRecord.CapexCAD))
		}
	}
	
	if len(conflicts) > 0 {
		match.Conflicts = conflicts
		match.Action = ActionConflict
		match.Confidence = ConfidenceLow
		match.Rationale = "Conflicting information across jurisdictions requires manual review"
		return
	}
	
	// Determine best action
	matchCount := 0
	if hasProvincial { matchCount++ }
	if hasMunicipal { matchCount++ }
	if hasIndigenous { matchCount++ }
	
	if matchCount >= 2 {
		match.Action = ActionMerge
		match.Confidence = ConfidenceHigh
		match.Rationale = fmt.Sprintf("Strong multi-jurisdiction alignment (%d matches)", matchCount)
		match.MergedProjectID = generateMergedID(match)
	} else if matchCount == 1 {
		match.Action = ActionLink
		match.Confidence = ConfidenceMedium
		match.Rationale = "Single jurisdiction match; linked for cross-reference"
	} else {
		match.Action = ActionNoMatch
		match.Confidence = ConfidenceLow
		match.Rationale = "No jurisdictional matches found"
	}
}

func generateMergedID(match *ReconciliationMatch) string {
	ids := []string{}
	if match.FederalRecord != nil {
		ids = append(ids, "fed:"+match.FederalRecord.ExternalID)
	}
	if match.ProvincialRecord != nil {
		ids = append(ids, "prov:"+match.ProvincialRecord.ExternalID)
	}
	if match.MunicipalRecord != nil {
		ids = append(ids, "mun:"+match.MunicipalRecord.ExternalID)
	}
	if match.IndigenousRecord != nil {
		ids = append(ids, "ind:"+match.IndigenousRecord.ExternalID)
	}
	return "merged:" + strings.Join(ids, "|")
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}