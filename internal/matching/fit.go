// Package matching provides deterministic counterparty fit scoring and deal
// precedent search. It never generates personalized investment advice — it
// produces evidence-grounded compatibility assessments between projects and
// investor archetypes.
package matching

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const FitMethodologyVersion = "counterparty-fit-v1.0"

// FitScore is a decomposed compatibility assessment.
type FitScore struct {
	ProjectID   string             `json:"project_id"`
	EntityID    string             `json:"entity_id,omitempty"`
	Archetype   domain.CounterpartyType `json:"archetype"`
	Score       float64            `json:"score"` // 0-100
	Factors     map[string]float64 `json:"factors"`
	Explanation []string           `json:"explanation"`
	Methodology string             `json:"methodology"`
	CalculatedAt time.Time         `json:"calculated_at"`
}

// FitContext gathers the project-side data for matching.
type FitContext struct {
	Project      *domain.Project
	CapitalNeeds []*domain.CapitalNeed
	Milestones   []*domain.Milestone
	FID          *domain.FIDIntelligence
}

// CalculateArchetypeFit evaluates how well a project matches a given
// counterparty archetype. It does NOT score specific entities.
func CalculateArchetypeFit(ctx FitContext, archetype domain.CounterpartyType, asOf time.Time) *FitScore {
	if ctx.Project == nil {
		return nil
	}
	factors := map[string]float64{}
	explanation := []string{}

	// Factor 1: Does the project explicitly seek this counterparty type?
	needMatch := counterpartyNeedFactor(ctx.CapitalNeeds, archetype)
	factors["need_match"] = needMatch
	if needMatch > 0 {
		explanation = append(explanation, "Project explicitly seeks "+string(archetype))
	}

	// Factor 2: Sector alignment with archetype's typical mandate.
	sectorFit := sectorArchetypeFit(ctx.Project.Sector, archetype)
	factors["sector_fit"] = sectorFit

	// Factor 3: Stage appropriateness.
	stageFit := stageArchetypeFit(ctx.Project.CurrentStage, archetype)
	factors["stage_fit"] = stageFit

	// Factor 4: Instrument compatibility.
	instrumentFit := instrumentArchetypeFit(ctx.CapitalNeeds, archetype)
	factors["instrument_fit"] = instrumentFit

	// Factor 5: Scale fit (ticket size).
	scaleFit := scaleArchetypeFit(ctx.Project.CapexCAD, archetype)
	factors["scale_fit"] = scaleFit

	score := clamp(
		needMatch*0.30+
			sectorFit*0.15+
			stageFit*0.20+
			instrumentFit*0.20+
			scaleFit*0.15,
		0, 100)

	sort.Strings(explanation)
	return &FitScore{
		ProjectID:    ctx.Project.ID,
		Archetype:    archetype,
		Score:        math.Round(score*100) / 100,
		Factors:      factors,
		Explanation:  explanation,
		Methodology:  FitMethodologyVersion,
		CalculatedAt: asOf,
	}
}

// CalculateInvestorFit evaluates how well a project matches a specific
// investor profile. This is a refinement of archetype fit using the
// investor's stated preferences.
func CalculateInvestorFit(ctx FitContext, profile *domain.InvestorProfile, asOf time.Time) *FitScore {
	if ctx.Project == nil || profile == nil {
		return nil
	}
	factors := map[string]float64{}
	explanation := []string{}

	// Type alignment.
	typeScore := 0.0
	for _, t := range profile.InvestorTypes {
		archFit := CalculateArchetypeFit(ctx, t, asOf)
		if archFit != nil && archFit.Score > typeScore {
			typeScore = archFit.Score
		}
	}
	factors["type_alignment"] = typeScore

	// Sector alignment.
	sectorScore := 0.0
	for _, s := range profile.TargetSectors {
		if s == ctx.Project.Sector {
			sectorScore = 100
			break
		}
	}
	factors["sector_alignment"] = sectorScore

	// Geography alignment.
	geoScore := 0.0
	for _, g := range profile.TargetGeographies {
		if strings.EqualFold(g, ctx.Project.Province) {
			geoScore = 100
			break
		}
	}
	factors["geography_alignment"] = geoScore

	// Ticket size alignment.
	ticketScore := 100.0
	if profile.MinTicketCAD > 0 && ctx.Project.CapexCAD > 0 && ctx.Project.CapexCAD < profile.MinTicketCAD {
		ticketScore = 0
	}
	if profile.MaxTicketCAD > 0 && ctx.Project.CapexCAD > profile.MaxTicketCAD {
		ticketScore = 0
	}
	factors["ticket_alignment"] = ticketScore

	// Stage alignment.
	stageScore := 0.0
	for _, s := range profile.PreferredStages {
		if s == ctx.Project.CurrentStage {
			stageScore = 100
			break
		}
	}
	factors["stage_alignment"] = stageScore

	score := clamp(
		typeScore*0.25+
			sectorScore*0.20+
			geoScore*0.15+
			ticketScore*0.15+
			stageScore*0.25,
		0, 100)

	sort.Strings(explanation)
	return &FitScore{
		ProjectID:    ctx.Project.ID,
		EntityID:     profile.EntityID,
		Score:        math.Round(score*100) / 100,
		Factors:      factors,
		Explanation:  explanation,
		Methodology:  FitMethodologyVersion,
		CalculatedAt: asOf,
	}
}

// --- Internal helpers ---

func counterpartyNeedFactor(needs []*domain.CapitalNeed, archetype domain.CounterpartyType) float64 {
	for _, need := range needs {
		for _, cp := range need.Counterparties {
			if cp == archetype {
				return 100
			}
		}
	}
	return 0
}

func sectorArchetypeFit(sector domain.Sector, archetype domain.CounterpartyType) float64 {
	// Infrastructure funds typically invest in infra sectors.
	infraSectors := map[domain.Sector]bool{
		domain.SectorCleanEnergy:   true,
		domain.SectorNuclearEnergy: true,
		domain.SectorTransportation: true,
		domain.SectorAICompute:     true,
	}
	switch archetype {
	case domain.CounterpartyInfrastructureFund, domain.CounterpartyPensionFund:
		if infraSectors[sector] {
			return 80
		}
		return 30
	case domain.CounterpartyECA:
		return 60 // ECAs are sector-agnostic for Canadian exports
	default:
		return 50
	}
}

func stageArchetypeFit(stage domain.LifecycleStage, archetype domain.CounterpartyType) float64 {
	switch archetype {
	case domain.CounterpartyPrivateEquity:
		// PE prefers development stage.
		switch stage {
		case domain.StageFeasibility, domain.StagePreFEED, domain.StageFEED:
			return 90
		case domain.StageDetailedEngineering, domain.StageFinancing:
			return 70
		default:
			return 30
		}
	case domain.CounterpartyInfrastructureFund, domain.CounterpartyPensionFund:
		// Infra funds prefer construction-ready or operating.
		switch stage {
		case domain.StageConstructionReady, domain.StageConstruction, domain.StageCommissioning, domain.StageOperating:
			return 90
		case domain.StageFID, domain.StageFIDLikely:
			return 70
		default:
			return 20
		}
	case domain.CounterpartyBank:
		// Banks prefer post-FID.
		switch stage {
		case domain.StageFID, domain.StageConstruction, domain.StageConstructionReady:
			return 90
		case domain.StageFinancing, domain.StageFIDLikely:
			return 70
		default:
			return 20
		}
	default:
		return 50
	}
}

func instrumentArchetypeFit(needs []*domain.CapitalNeed, archetype domain.CounterpartyType) float64 {
	archetypeInstruments := map[domain.CounterpartyType][]domain.CapitalNeedType{
		domain.CounterpartyInfrastructureFund: {domain.NeedInfrastructureEquity, domain.NeedEquity},
		domain.CounterpartyPensionFund:        {domain.NeedPensionCapital, domain.NeedInfrastructureEquity, domain.NeedEquity},
		domain.CounterpartyBank:               {domain.NeedSeniorDebt, domain.NeedDebt, domain.NeedProjectFinance},
		domain.CounterpartyPrivateCredit:      {domain.NeedPrivateCredit, domain.NeedSubordinatedDebt},
		domain.CounterpartyECA:                {domain.NeedExportCredit},
		domain.CounterpartySovereignFund:      {domain.NeedSovereignCapital, domain.NeedEquity},
		domain.CounterpartyPrivateEquity:      {domain.NeedEquity, domain.NeedDevelopmentCapital},
	}
	preferred, ok := archetypeInstruments[archetype]
	if !ok {
		return 50
	}
	for _, need := range needs {
		for _, nt := range need.Types {
			for _, pi := range preferred {
				if nt == pi {
					return 90
				}
			}
		}
	}
	return 20
}

func scaleArchetypeFit(capexCAD int64, archetype domain.CounterpartyType) float64 {
	if capexCAD == 0 {
		return 30 // unknown scale
	}
	switch archetype {
	case domain.CounterpartyInfrastructureFund, domain.CounterpartyPensionFund:
		if capexCAD >= 500_000_000 { // $500M+
			return 90
		}
		if capexCAD >= 100_000_000 { // $100M+
			return 60
		}
		return 20
	case domain.CounterpartyPrivateEquity:
		if capexCAD >= 50_000_000 && capexCAD <= 2_000_000_000 { // $50M-$2B sweet spot
			return 80
		}
		return 40
	default:
		return 50
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
