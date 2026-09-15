package matching

import (
	"sort"
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// PrecedentMatch records a historical deal that is similar to a given project,
// with an explanation of which dimensions matched.
type PrecedentMatch struct {
	Deal          domain.DealPrecedent `json:"deal"`
	SimilarityPct float64             `json:"similarity_pct"` // 0-100
	MatchedOn     []string            `json:"matched_on"`     // sector, province, stage, instrument, scale
}

// FindDealPrecedents searches investor profiles for historical transactions
// that resemble the target project along key dimensions. This answers:
// "Who has financed projects like this before?" — it explains similarity,
// never claims future intent.
func FindDealPrecedents(project *domain.Project, profiles []*domain.InvestorProfile, topN int) []PrecedentMatch {
	if project == nil || len(profiles) == 0 {
		return nil
	}
	if topN <= 0 {
		topN = 10
	}

	var matches []PrecedentMatch
	for _, profile := range profiles {
		for _, deal := range profile.PublicDealHistory {
			similarity, matchedOn := dealSimilarity(project, deal)
			if similarity < 20 {
				continue
			}
			matches = append(matches, PrecedentMatch{
				Deal:          deal,
				SimilarityPct: similarity,
				MatchedOn:     matchedOn,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].SimilarityPct > matches[j].SimilarityPct
	})

	if len(matches) > topN {
		matches = matches[:topN]
	}
	return matches
}

func dealSimilarity(project *domain.Project, deal domain.DealPrecedent) (float64, []string) {
	var score float64
	var matched []string

	// Sector match (strongest signal).
	if deal.Sector == project.Sector {
		score += 35
		matched = append(matched, "sector")
	}

	// Province match.
	if deal.Province != "" && strings.EqualFold(deal.Province, project.Province) {
		score += 15
		matched = append(matched, "province")
	}

	// Stage match.
	if deal.Stage == project.CurrentStage {
		score += 20
		matched = append(matched, "stage")
	}

	// Scale similarity (within 5x).
	if deal.AmountCAD > 0 && project.CapexCAD > 0 {
		ratio := float64(deal.AmountCAD) / float64(project.CapexCAD)
		if ratio < 1 {
			ratio = 1 / ratio
		}
		if ratio <= 5 {
			score += 20
			matched = append(matched, "scale")
		} else if ratio <= 10 {
			score += 10
			matched = append(matched, "scale_approximate")
		}
	}

	// Recent deal bonus.
	if deal.Year > 0 && deal.Year >= 2022 {
		score += 10
		matched = append(matched, "recent")
	}

	return score, matched
}
