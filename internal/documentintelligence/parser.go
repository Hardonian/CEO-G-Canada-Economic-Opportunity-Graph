// Package documentintelligence extracts restricted candidate records without
// promoting them into the public project graph. It is intentionally a bounded,
// deterministic parser suitable for evaluation fixtures and human review.
package documentintelligence

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const ExtractorVersion = "portfolio-cards-v1.0"

var (
	cardBoundary = regexp.MustCompile(`(?im)^\s*---\s*PROJECT\s*---\s*$`)
	fieldLine    = regexp.MustCompile(`(?m)^([A-Za-z][A-Za-z ]{1,40}):\s*(.+)$`)
	amountToken  = regexp.MustCompile(`(?i)(US\$|USD\s*|C\$|CAD\s*|\$)\s*([0-9]+(?:\.[0-9]+)?)\s*([BM])`)
	phaseToken   = regexp.MustCompile(`(?i)\b(phase\s+[0-9A-Za-z]+)\b`)
)

type ExtractedCard struct {
	Candidate          *domain.CandidateProject     `json:"candidate"`
	CapitalRequirement *domain.CapitalRequirement  `json:"capital_requirement,omitempty"`
	CapitalNeed        *domain.CapitalNeed          `json:"capital_need,omitempty"`
}

func ExtractCards(text, sourceID string, visibility domain.VisibilityClass, observedAt time.Time) ([]ExtractedCard, error) {
	if !visibility.Valid() || visibility.Public() {
		return nil, fmt.Errorf("portfolio extraction requires an explicit private or restricted visibility")
	}
	parts := cardBoundary.Split(strings.TrimSpace(text), -1)
	results := make([]ExtractedCard, 0, len(parts))
	for _, part := range parts {
		fields := parseFields(part)
		name := strings.TrimSpace(first(fields, "project", "project name", "name"))
		if name == "" {
			continue
		}
		id := stableID(sourceID + "|" + name)
		candidate := &domain.CandidateProject{
			ID: "candidate-" + id, SourceID: sourceID, Name: name,
			Proponent: first(fields, "proponent", "developer"), Location: first(fields, "location", "geography"),
			OriginalStage: first(fields, "stage", "development stage"), Visibility: visibility,
			CreatedAt: observedAt.UTC(),
		}
		card := ExtractedCard{Candidate: candidate}
		if raw := first(fields, "capex", "capital cost", "project cost"); raw != "" {
			amount := ParseCapex(raw)
			phase := ""
			if match := phaseToken.FindStringSubmatch(raw); len(match) > 1 {
				phase = strings.TrimSpace(match[1])
			}
			card.CapitalRequirement = &domain.CapitalRequirement{
				ID: "capital-requirement-" + id, ProjectID: candidate.ID, Purpose: phase,
				Amount: amount, Status: domain.ConfidenceReported, Visibility: visibility,
				Publishable: false, CreatedAt: observedAt.UTC(),
			}
		}
		if raw := first(fields, "financing objective", "capital need", "partners sought"); raw != "" {
			card.CapitalNeed = &domain.CapitalNeed{
				ID: "capital-need-" + id, ProjectID: candidate.ID, Types: ClassifyCapitalNeeds(raw),
				Counterparties: ClassifyCounterparties(raw), Status: domain.CapitalSeeking,
				OriginalLanguage: raw, Visibility: visibility, Publishable: false,
				PublicationState: domain.PublicationPrivateOnly, CreatedAt: observedAt.UTC(), UpdatedAt: observedAt.UTC(),
			}
		}
		results = append(results, card)
	}
	return results, nil
}

func parseFields(text string) map[string]string {
	fields := map[string]string{}
	for _, match := range fieldLine.FindAllStringSubmatch(text, -1) {
		fields[strings.ToLower(strings.TrimSpace(match[1]))] = strings.TrimSpace(match[2])
	}
	return fields
}

func first(fields map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := fields[key]; value != "" {
			return value
		}
	}
	return ""
}

func ParseCapex(raw string) domain.MonetaryAmount {
	result := domain.MonetaryAmount{AmountType: domain.AmountNotAvailable, OriginalText: strings.TrimSpace(raw)}
	lower := strings.ToLower(strings.TrimSpace(raw))
	if lower == "" || strings.Contains(lower, "not available") || lower == "n/a" || lower == "unknown" {
		return result
	}
	matches := amountToken.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return result
	}
	currency := parseCurrency(matches[0][1])
	values := make([]int64, 0, len(matches))
	for _, match := range matches {
		if parseCurrency(match[1]) != currency {
			return result // mixed-currency text requires human review
		}
		value, _ := strconv.ParseFloat(match[2], 64)
		multiplier := float64(1_000_000)
		if strings.EqualFold(match[3], "B") {
			multiplier = 1_000_000_000
		}
		values = append(values, int64(value*multiplier+0.5))
	}
	result.Currency = currency
	if len(values) > 1 {
		min, max := values[0], values[1]
		if min > max {
			min, max = max, min
		}
		result.Minimum, result.Maximum, result.AmountType = &min, &max, domain.AmountRange
		return result
	}
	value := values[0]
	result.Amount = &value
	switch {
	case strings.Contains(lower, "+") || strings.Contains(lower, "at least"):
		result.AmountType = domain.AmountMinimum
		result.Minimum = &value
	case strings.Contains(lower, "approximately") || strings.Contains(lower, "approx.") || strings.Contains(lower, "about") || strings.Contains(lower, "~"):
		result.AmountType = domain.AmountApproximate
	case strings.Contains(lower, "up to"):
		result.AmountType = domain.AmountMaximum
		result.Maximum = &value
	default:
		result.AmountType = domain.AmountExact
	}
	return result
}

func parseCurrency(prefix string) string {
	upper := strings.ToUpper(strings.TrimSpace(prefix))
	switch upper {
	case "US$", "USD":
		return "USD"
	case "C$", "CAD":
		return "CAD"
	default:
		return "UNSPECIFIED"
	}
}

func ClassifyCapitalNeeds(raw string) []domain.CapitalNeedType {
	lower := strings.ToLower(raw)
	mapping := []struct {
		phrases []string
		kind    domain.CapitalNeedType
	}{
		{[]string{"infrastructure equity"}, domain.NeedInfrastructureEquity},
		{[]string{"project finance", "project financing"}, domain.NeedProjectFinance},
		{[]string{"private credit"}, domain.NeedPrivateCredit},
		{[]string{"joint venture", " jv "}, domain.NeedJointVenture},
		{[]string{"strategic invest"}, domain.NeedStrategicInvestment},
		{[]string{"government co-invest", "government support"}, domain.NeedGovernmentSupport},
		{[]string{"loan guarantee"}, domain.NeedLoanGuarantee},
		{[]string{"export credit", "eca financing"}, domain.NeedExportCredit},
		{[]string{"indigenous equity"}, domain.NeedIndigenousEquity},
		{[]string{"pension capital"}, domain.NeedPensionCapital},
		{[]string{"sovereign capital"}, domain.NeedSovereignCapital},
		{[]string{"offtake"}, domain.NeedOfftake},
		{[]string{"anchor tenant"}, domain.NeedAnchorTenant},
		{[]string{"prepayment"}, domain.NeedPrepayment},
		{[]string{"streaming"}, domain.NeedStreaming},
		{[]string{"royalty"}, domain.NeedRoyalty},
		{[]string{"senior debt"}, domain.NeedSeniorDebt},
		{[]string{"subordinated debt", "mezzanine"}, domain.NeedSubordinatedDebt},
		{[]string{"development capital"}, domain.NeedDevelopmentCapital},
		{[]string{"equity"}, domain.NeedEquity},
		{[]string{"debt"}, domain.NeedDebt},
	}
	seen := map[domain.CapitalNeedType]bool{}
	result := []domain.CapitalNeedType{}
	for _, item := range mapping {
		for _, phrase := range item.phrases {
			if strings.Contains(" "+lower+" ", phrase) && !seen[item.kind] {
				seen[item.kind] = true
				result = append(result, item.kind)
				break
			}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func ClassifyCounterparties(raw string) []domain.CounterpartyType {
	lower := strings.ToLower(raw)
	mapping := map[string]domain.CounterpartyType{
		"infrastructure fund": domain.CounterpartyInfrastructureFund,
		"pension": domain.CounterpartyPensionFund,
		"private equity": domain.CounterpartyPrivateEquity,
		"bank": domain.CounterpartyBank,
		"private credit": domain.CounterpartyPrivateCredit,
		"export credit": domain.CounterpartyECA,
		"sovereign": domain.CounterpartySovereignFund,
		"strategic partner": domain.CounterpartyStrategicCorporate,
		"epc": domain.CounterpartyEPC,
		"offtaker": domain.CounterpartyOfftaker,
		"anchor tenant": domain.CounterpartyAnchorTenant,
		"indigenous partner": domain.CounterpartyIndigenousPartner,
		"joint venture": domain.CounterpartyJVPartner,
	}
	seen := map[domain.CounterpartyType]bool{}
	result := []domain.CounterpartyType{}
	for phrase, kind := range mapping {
		if strings.Contains(lower, phrase) && !seen[kind] {
			seen[kind] = true
			result = append(result, kind)
		}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func NormalizeStages(raw string) []domain.LifecycleStage {
	lower := strings.ToLower(raw)
	mapping := []struct {
		phrase string
		stage  domain.LifecycleStage
	}{
		{"under construction", domain.StageConstruction}, {"construction-ready", domain.StageConstructionReady},
		{"shovel-ready", domain.StageConstructionReady}, {"pre-fid", domain.StageFIDLikely},
		{"detailed engineering", domain.StageDetailedEngineering}, {"pre-feed", domain.StagePreFEED},
		{"feed", domain.StageFEED}, {"pre-feasibility", domain.StagePreDevelopment},
		{"feasibility", domain.StageFeasibility}, {"permitting", domain.StagePermitting},
		{"financing", domain.StageFinancing}, {"commissioning", domain.StageCommissioning},
		{"operational", domain.StageOperating}, {"operating", domain.StageOperating},
		{"expansion", domain.StageExpansion}, {"concept", domain.StageConcept},
	}
	seen := map[domain.LifecycleStage]bool{}
	result := []domain.LifecycleStage{}
	for _, item := range mapping {
		if strings.Contains(lower, item.phrase) && !seen[item.stage] {
			seen[item.stage] = true
			result = append(result, item.stage)
		}
	}
	return result
}

func stableID(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:8])
}
