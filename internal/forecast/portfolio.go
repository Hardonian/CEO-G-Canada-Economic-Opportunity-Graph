package forecast

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// EvaluatePortfolio applies one canonical request to a bounded set of projects
// and returns both the underlying reports and transparent aggregates.
func EvaluatePortfolio(contexts []Context, request Request) (*PortfolioReport, error) {
	if len(contexts) == 0 {
		return nil, fmt.Errorf("%w: at least one project context is required", ErrInvalidInput)
	}
	if len(contexts) > DefaultMaxProjects {
		return nil, fmt.Errorf("%w: portfolio exceeds %d projects", ErrInvalidInput, DefaultMaxProjects)
	}
	req, err := normalizeRequest(request)
	if err != nil {
		return nil, err
	}
	canonicalRequest := Request{AsOf: req.asOf, HorizonsMonths: append([]int(nil), req.horizons...), Scenarios: append([]Scenario(nil), req.scenarios...)}
	ordered := append([]Context(nil), contexts...)
	for i, ctx := range ordered {
		if ctx.Project == nil {
			return nil, fmt.Errorf("%w: project context %d has no project", ErrInvalidInput, i)
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Project.ID < ordered[j].Project.ID })
	for i := 1; i < len(ordered); i++ {
		if ordered[i-1].Project.ID == ordered[i].Project.ID {
			return nil, fmt.Errorf("%w: duplicate portfolio project id %q", ErrInvalidInput, ordered[i].Project.ID)
		}
	}

	reports := make([]*Report, 0, len(ordered))
	for _, ctx := range ordered {
		report, evaluateErr := Evaluate(ctx, canonicalRequest)
		if evaluateErr != nil {
			return nil, fmt.Errorf("project %q: %w", ctx.Project.ID, evaluateErr)
		}
		reports = append(reports, report)
	}

	aggregates, err := aggregateScenarios(reports, req)
	if err != nil {
		return nil, err
	}
	concentrations, totalKnownCapex, unknownCapex, err := calculateConcentrations(ordered)
	if err != nil {
		return nil, err
	}
	quality, status, confidence := aggregateQuality(reports, unknownCapex)
	warnings := portfolioWarnings(reports)
	watch := portfolioWatchItems(quality, concentrations, totalKnownCapex)
	inputHash, err := portfolioInputHash(reports, req)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot hash portfolio input: %v", ErrInvalidInput, err)
	}
	assurance := SovereignAssurance{
		ProcessingMode:          "LOCAL_DETERMINISTIC_RULES",
		ExternalNetworkRequired: false,
		ExternalTelemetry:       false,
		RawEvidenceEmitted:      false,
		Deterministic:           true,
		Statement:               "These are package execution properties, not a certification of the host, deployment, data classification, or legal compliance.",
	}
	return &PortfolioReport{
		ID: "portfolio_forecast_" + inputHash[:24], MethodologyVersion: MethodologyVersion, AsOf: req.asOf, CalculatedAt: req.asOf,
		Status: status, Confidence: confidence, InputHash: inputHash, ProjectCount: len(reports), DataQuality: quality,
		ScenarioAggregates: aggregates, Concentrations: concentrations, ProjectForecasts: reports, WatchItems: watch, Warnings: warnings,
		Limitations: []string{
			"Portfolio figures are arithmetic aggregates of project-level deterministic planning indices, not statistically calibrated probabilities or investment advice.",
			"Capex-weighted indices exclude projects without reported capex; the excluded count is explicit in data_quality.",
			"Project dependencies, common-shock correlation, double-counted supply-chain opportunities, inflation, and portfolio diversification are not modelled.",
			"Concentration shares use known reported capex, not market value, economic impact, or risk-adjusted exposure.",
		},
		Assurance: assurance,
	}, nil
}

func aggregateScenarios(reports []*Report, req normalizedRequest) ([]PortfolioScenario, error) {
	result := make([]PortfolioScenario, 0, len(req.scenarios))
	for _, scenario := range req.scenarios {
		aggregate := PortfolioScenario{ScenarioID: scenario.ID, ScenarioName: scenario.Name}
		for _, report := range reports {
			projectScenario, ok := findScenario(report, scenario.ID)
			if !ok {
				return nil, fmt.Errorf("%w: project %q is missing scenario %q", ErrInvalidInput, report.ProjectID, scenario.ID)
			}
			var err error
			aggregate.ScenarioAdjustedCapexCAD, err = addMoneyRange(aggregate.ScenarioAdjustedCapexCAD, projectScenario.Capital.ScenarioAdjustedCapexCAD)
			if err != nil {
				return nil, portfolioOverflow(scenario.ID, "adjusted capex")
			}
			aggregate.IndicativeFundingGapCAD, err = addMoneyRange(aggregate.IndicativeFundingGapCAD, projectScenario.Capital.IndicativeFundingGapCAD)
			if err != nil {
				return nil, portfolioOverflow(scenario.ID, "funding gap")
			}
			for destination, value := range map[*int64]int64{
				&aggregate.EvidencedCommittedCAD:     projectScenario.Capital.EvidencedCommittedCAD,
				&aggregate.ConditionallyCommittedCAD: projectScenario.Capital.ConditionallyCommittedCAD,
				&aggregate.ConfirmedProcurementCAD:   projectScenario.Capital.ConfirmedProcurementCAD,
				&aggregate.ConfirmedOpportunityCAD:   projectScenario.Capital.ConfirmedOpportunityCAD,
				&aggregate.DerivedOpportunityCAD:     projectScenario.Capital.DerivedOpportunityCAD,
			} {
				*destination, err = safeAdd(*destination, value)
				if err != nil {
					return nil, portfolioOverflow(scenario.ID, "capital pipeline")
				}
			}
		}
		for _, stage := range []domain.LifecycleStage{domain.StageFID, domain.StageConstruction, domain.StageOperating} {
			for _, horizon := range req.horizons {
				summary := PortfolioMilestoneSummary{Stage: stage, HorizonMonths: horizon, HorizonDate: addCalendarMonths(req.asOf, horizon)}
				baseTotal, weightedTotal, weight := 0.0, 0.0, float64(0)
				for _, report := range reports {
					projectScenario, _ := findScenario(report, scenario.ID)
					likelihood, ok := findLikelihood(projectScenario, stage, horizon)
					if !ok {
						return nil, fmt.Errorf("%w: project %q is missing %s/%d-month output", ErrInvalidInput, report.ProjectID, stage, horizon)
					}
					base := likelihood.CompletionLikelihoodIndexPct.Base
					baseTotal += base
					if base >= 50 {
						summary.ProjectsAtOrAbove50Index++
					}
					capex := float64(projectScenario.Capital.ScenarioAdjustedCapexCAD.Base)
					if capex > 0 {
						weightedTotal += base * capex
						weight += capex
					}
				}
				summary.MeanBaseLikelihoodIndexPct = round1(baseTotal / float64(len(reports)))
				if weight > 0 {
					weighted := round1(weightedTotal / weight)
					summary.CapexWeightedBaseIndexPct = &weighted
				}
				aggregate.MilestoneSummaries = append(aggregate.MilestoneSummaries, summary)
			}
		}
		result = append(result, aggregate)
	}
	return result, nil
}

func calculateConcentrations(contexts []Context) ([]Concentration, int64, int, error) {
	type bucket struct {
		count int
		capex int64
	}
	provinces, sectors := make(map[string]bucket), make(map[string]bucket)
	var total int64
	unknown := 0
	for _, ctx := range contexts {
		capex := ctx.Project.CapexCAD
		if capex == 0 {
			unknown++
		}
		var err error
		total, err = safeAdd(total, capex)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("%w: portfolio reported capex overflows CAD range", ErrInvalidInput)
		}
		province := ctx.Project.Province
		if province == "" {
			province = "UNKNOWN"
		}
		sector := string(ctx.Project.Sector)
		if sector == "" {
			sector = "UNKNOWN"
		}
		provinceBucket := provinces[province]
		provinceBucket.count++
		provinceBucket.capex, err = safeAdd(provinceBucket.capex, capex)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("%w: concentration capex overflows CAD range", ErrInvalidInput)
		}
		provinces[province] = provinceBucket
		sectorBucket := sectors[sector]
		sectorBucket.count++
		sectorBucket.capex, err = safeAdd(sectorBucket.capex, capex)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("%w: concentration capex overflows CAD range", ErrInvalidInput)
		}
		sectors[sector] = sectorBucket
	}
	result := make([]Concentration, 0, len(provinces)+len(sectors))
	for dimension, values := range map[string]map[string]bucket{"province": provinces, "sector": sectors} {
		for value, item := range values {
			share := 0.0
			if total > 0 {
				share = round1(float64(item.capex) / float64(total) * 100)
			}
			result = append(result, Concentration{Dimension: dimension, Value: value, ProjectCount: item.count, ReportedCapexCAD: item.capex, KnownCapexSharePct: share})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Dimension != result[j].Dimension {
			return result[i].Dimension < result[j].Dimension
		}
		if result[i].KnownCapexSharePct != result[j].KnownCapexSharePct {
			return result[i].KnownCapexSharePct > result[j].KnownCapexSharePct
		}
		return result[i].Value < result[j].Value
	})
	return result, total, unknown, nil
}

func aggregateQuality(reports []*Report, unknownCapex int) (PortfolioDataQuality, domain.IntelligenceStatus, ConfidenceRating) {
	quality := PortfolioDataQuality{UnknownCapexProjects: unknownCapex}
	allHealthy := true
	for _, report := range reports {
		quality.AverageScore += report.DataQuality.Score
		switch report.Confidence {
		case ConfidenceHigh:
			quality.HighProjects++
		case ConfidenceModerate:
			quality.ModerateProjects++
		case ConfidenceLow:
			quality.LowProjects++
		case ConfidenceInsufficient:
			quality.InsufficientProjects++
		}
		if report.Status == domain.StatusStale {
			quality.StaleProjects++
		}
		if report.Status != domain.StatusHealthy {
			allHealthy = false
		}
	}
	quality.AverageScore = round1(quality.AverageScore / float64(len(reports)))
	confidence := ConfidenceHigh
	switch {
	case quality.InsufficientProjects > 0:
		confidence = ConfidenceInsufficient
	case quality.LowProjects > 0:
		confidence = ConfidenceLow
	case quality.ModerateProjects > 0:
		confidence = ConfidenceModerate
	}
	status := domain.StatusPartial
	if allHealthy {
		status = domain.StatusHealthy
	} else if quality.StaleProjects == len(reports) {
		status = domain.StatusStale
	} else if quality.InsufficientProjects == len(reports) {
		status = domain.StatusDegraded
	}
	return quality, status, confidence
}

func portfolioWarnings(reports []*Report) []string {
	values := make([]string, 0)
	for _, report := range reports {
		for _, warning := range report.Warnings {
			values = append(values, report.ProjectID+":"+warning)
		}
	}
	return sortedUnique(values)
}

func portfolioWatchItems(quality PortfolioDataQuality, concentrations []Concentration, totalKnownCapex int64) []WatchItem {
	var result []WatchItem
	if quality.LowProjects+quality.InsufficientProjects > 0 {
		result = append(result, WatchItem{Code: "PORTFOLIO_LOW_SUPPORT", Severity: "HIGH", Message: fmt.Sprintf("%d project forecasts have low or insufficient support.", quality.LowProjects+quality.InsufficientProjects)})
	}
	if quality.UnknownCapexProjects > 0 {
		result = append(result, WatchItem{Code: "PORTFOLIO_UNKNOWN_CAPEX", Severity: "HIGH", Message: fmt.Sprintf("%d projects are excluded from capex-weighted metrics because reported capex is unknown.", quality.UnknownCapexProjects)})
	}
	if totalKnownCapex > 0 {
		for _, item := range concentrations {
			if item.KnownCapexSharePct > 50 {
				result = append(result, WatchItem{Code: "CAPEX_CONCENTRATION_" + stringsForCode(item.Dimension, item.Value), Severity: "MEDIUM", Message: fmt.Sprintf("%s %q represents %.1f%% of known reported portfolio capex; this is a screening flag, not a risk conclusion.", item.Dimension, item.Value, item.KnownCapexSharePct)})
			}
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Code < result[j].Code })
	return result
}

func portfolioInputHash(reports []*Report, req normalizedRequest) (string, error) {
	type item struct{ ProjectID, InputHash string }
	value := struct {
		Version   string
		AsOf      interface{}
		Horizons  []int
		Scenarios []Scenario
		Projects  []item
	}{Version: MethodologyVersion, AsOf: req.asOf, Horizons: req.horizons, Scenarios: req.scenarios}
	for _, report := range reports {
		value.Projects = append(value.Projects, item{report.ProjectID, report.InputHash})
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func findScenario(report *Report, id string) (ScenarioForecast, bool) {
	for _, scenario := range report.Scenarios {
		if scenario.Scenario.ID == id {
			return scenario, true
		}
	}
	return ScenarioForecast{}, false
}

func findLikelihood(scenario ScenarioForecast, stage domain.LifecycleStage, horizon int) (HorizonLikelihood, bool) {
	for _, milestone := range scenario.Milestones {
		if milestone.Stage != stage {
			continue
		}
		for _, value := range milestone.ByHorizon {
			if value.HorizonMonths == horizon {
				return value, true
			}
		}
	}
	return HorizonLikelihood{}, false
}

func addMoneyRange(left, right MoneyRange) (MoneyRange, error) {
	low, err := safeAdd(left.Low, right.Low)
	if err != nil {
		return MoneyRange{}, err
	}
	base, err := safeAdd(left.Base, right.Base)
	if err != nil {
		return MoneyRange{}, err
	}
	high, err := safeAdd(left.High, right.High)
	if err != nil {
		return MoneyRange{}, err
	}
	return MoneyRange{Low: low, Base: base, High: high}, nil
}

func portfolioOverflow(scenarioID, field string) error {
	return fmt.Errorf("%w: portfolio scenario %q %s overflows CAD range", ErrInvalidInput, scenarioID, field)
}

func stringsForCode(values ...string) string {
	result := ""
	for _, value := range values {
		for _, character := range value {
			if character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9' {
				result += string(character)
			} else if len(result) > 0 && result[len(result)-1] != '_' {
				result += "_"
			}
		}
		if len(result) > 0 && result[len(result)-1] != '_' {
			result += "_"
		}
	}
	for len(result) > 0 && result[len(result)-1] == '_' {
		result = result[:len(result)-1]
	}
	return result
}
