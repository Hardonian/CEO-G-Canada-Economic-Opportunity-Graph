// Package global_trade ingests official, international trade and logistics
// observations without requiring a commercial data credential.
package global_trade

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

const (
	adapterName     = "world_bank_global_trade"
	pipelineVersion = "world-bank-trade-v1"
	parserVersion   = "world-bank-indicators-json-v1"
	mappingVersion  = "canada-trade-score-input-v1"
	defaultFixture  = "data/fixtures/world_bank_trade_canada.json"
	maxResponseSize = int64(2 << 20)
)

var indicatorUnits = map[string]string{
	"LP.LPI.OVRL.XQ":    "index_1_to_5",
	"LP.LPI.CUST.XQ":    "index_1_to_5",
	"LP.LPI.INFR.XQ":    "index_1_to_5",
	"LP.LPI.ITRN.XQ":    "index_1_to_5",
	"LP.LPI.LOGS.XQ":    "index_1_to_5",
	"LP.LPI.TRAC.XQ":    "index_1_to_5",
	"LP.LPI.TIME.XQ":    "index_1_to_5",
	"NE.TRD.GNFS.ZS":    "percent",
	"TX.VAL.MRCH.CD.WT": "current_USD",
	"TM.VAL.MRCH.CD.WT": "current_USD",
	"TX.VAL.TECH.MF.ZS": "percent",
}

var indicatorCodes = []string{
	"LP.LPI.OVRL.XQ", "LP.LPI.CUST.XQ", "LP.LPI.INFR.XQ", "LP.LPI.ITRN.XQ",
	"LP.LPI.LOGS.XQ", "LP.LPI.TRAC.XQ", "LP.LPI.TIME.XQ", "NE.TRD.GNFS.ZS",
	"TX.VAL.MRCH.CD.WT", "TM.VAL.MRCH.CD.WT", "TX.VAL.TECH.MF.ZS",
}

type snapshot struct {
	SchemaVersion        string        `json:"schema_version"`
	Publisher            string        `json:"publisher"`
	SourceID             string        `json:"source_id"`
	License              string        `json:"license"`
	RetrievedAt          string        `json:"retrieved_at"`
	PublisherLastUpdated string        `json:"publisher_last_updated"`
	Geography            string        `json:"geography"`
	Observations         []observation `json:"observations"`
}

type observation struct {
	MetricCode      string   `json:"metric_code"`
	MetricName      string   `json:"metric_name"`
	ReferencePeriod string   `json:"reference_period"`
	Value           float64  `json:"value"`
	Unit            string   `json:"unit"`
	ScaleMin        *float64 `json:"scale_min,omitempty"`
	ScaleMax        *float64 `json:"scale_max,omitempty"`
	SourceURL       string   `json:"source_url"`
	Locator         string   `json:"locator"`
}

type Adapter struct {
	fixturePath string
	live        bool
	client      *http.Client
	health      adapters.SourceHealth
}

func NewAdapter(fixturePath string) *Adapter {
	if fixturePath == "" {
		fixturePath = defaultFixture
	}
	return &Adapter{fixturePath: fixturePath, health: adapters.SourceHealth{
		AdapterName: adapterName, Tier: domain.SourceTier1,
		Status: string(domain.StatusHealthy), RateLimitState: "NOT_APPLICABLE", Mode: "CURATED_SNAPSHOT",
	}}
}

func NewLiveAdapter(client *http.Client) *Adapter {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	copyClient := *client
	copyClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" || !strings.EqualFold(req.URL.Hostname(), "api.worldbank.org") {
			return fmt.Errorf("blocked World Bank redirect")
		}
		return nil
	}
	return &Adapter{live: true, client: &copyClient, health: adapters.SourceHealth{
		AdapterName: adapterName, Tier: domain.SourceTier1,
		Status: string(domain.StatusHealthy), RateLimitState: "PUBLIC_NO_KEY", Mode: "LIVE",
	}}
}

// NewFromEnv opts into network ingestion only when explicitly configured.
// World Bank Indicators is a public, credential-free API.
func NewFromEnv() *Adapter {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("GLOBAL_TRADE_MODE")), "live") {
		return NewLiveAdapter(nil)
	}
	return NewAdapter(strings.TrimSpace(os.Getenv("GLOBAL_TRADE_FIXTURE")))
}

func (a *Adapter) Name() string                   { return adapterName }
func (a *Adapter) Tier() domain.SourceTier        { return domain.SourceTier1 }
func (a *Adapter) Health() *adapters.SourceHealth { return &a.health }

func (a *Adapter) Fetch(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	a.health.LastAttempt = time.Now().UTC()
	var (
		data []byte
		err  error
	)
	if a.live {
		data, err = a.fetchLive(ctx)
	} else {
		data, err = adapters.ReadBoundedFile(a.fixturePath, adapters.MaxFixtureBytes)
	}
	if err != nil {
		a.health.Status = string(domain.StatusDegraded)
		a.health.LastError = err.Error()
		return nil, err
	}
	a.health.LastSuccess = time.Now().UTC()
	a.health.Status = string(domain.StatusHealthy)
	a.health.LastError = ""
	return data, nil
}

func (a *Adapter) fetchLive(ctx context.Context) ([]byte, error) {
	result := snapshot{
		SchemaVersion: "world-bank-canada-trade-v1", Publisher: "World Bank",
		SourceID: "world-bank-indicators-api", License: "Creative Commons Attribution 4.0",
		RetrievedAt: time.Now().UTC().Format(time.RFC3339), Geography: "CAN",
	}
	for _, code := range indicatorCodes {
		unit := indicatorUnits[code]
		endpoint := "https://api.worldbank.org/v2/country/CAN/indicator/" + url.PathEscape(code) + "?format=json&per_page=100"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "CanadaOpportunityGraph/1.0 (+https://github.com/Hardonian/Canada-Economic-Opportunity-Graph)")
		response, err := a.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch World Bank indicator %s: %w", code, err)
		}
		body, readErr := readBounded(response.Body, maxResponseSize)
		response.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read World Bank indicator %s: %w", code, readErr)
		}
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("World Bank indicator %s returned HTTP %d", code, response.StatusCode)
		}
		obs, updated, err := parseWorldBankResponse(body, code, unit, endpoint)
		if err != nil {
			return nil, err
		}
		if updated > result.PublisherLastUpdated {
			result.PublisherLastUpdated = updated
		}
		result.Observations = append(result.Observations, obs)
	}
	if result.PublisherLastUpdated != "" {
		result.PublisherLastUpdated += "T00:00:00Z"
	}
	return json.Marshal(result)
}

func parseWorldBankResponse(data []byte, code, unit, sourceURL string) (observation, string, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(data, &parts); err != nil || len(parts) != 2 {
		return observation{}, "", fmt.Errorf("parse World Bank indicator %s response envelope", code)
	}
	var header struct {
		LastUpdated string `json:"lastupdated"`
	}
	var rows []struct {
		Indicator struct {
			ID    string `json:"id"`
			Value string `json:"value"`
		} `json:"indicator"`
		CountryISO3Code string   `json:"countryiso3code"`
		Date            string   `json:"date"`
		Value           *float64 `json:"value"`
	}
	if err := json.Unmarshal(parts[0], &header); err != nil {
		return observation{}, "", err
	}
	if err := json.Unmarshal(parts[1], &rows); err != nil {
		return observation{}, "", err
	}
	for _, row := range rows {
		if row.Value == nil || row.Indicator.ID != code || row.CountryISO3Code != "CAN" {
			continue
		}
		obs := observation{MetricCode: code, MetricName: row.Indicator.Value, ReferencePeriod: row.Date, Value: *row.Value, Unit: unit, SourceURL: sourceURL,
			Locator: "country=CAN; indicator=" + code + "; date=" + row.Date}
		if unit == "index_1_to_5" {
			min, max := 1.0, 5.0
			obs.ScaleMin, obs.ScaleMax = &min, &max
		} else if unit == "percent" {
			min, max := 0.0, 100.0
			obs.ScaleMin, obs.ScaleMax = &min, &max
		}
		return obs, header.LastUpdated, nil
	}
	return observation{}, "", fmt.Errorf("World Bank indicator %s has no published Canada value", code)
}

func (a *Adapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var doc snapshot
	if err := decoder.Decode(&doc); err != nil {
		return a.parseFailure(fmt.Errorf("parse global trade snapshot: %w", err))
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return a.parseFailure(fmt.Errorf("parse global trade snapshot: trailing JSON content"))
	}
	retrievedAt, err := time.Parse(time.RFC3339, doc.RetrievedAt)
	if err != nil {
		return a.parseFailure(fmt.Errorf("invalid retrieved_at: %w", err))
	}
	publisherUpdated, err := time.Parse(time.RFC3339, doc.PublisherLastUpdated)
	if err != nil {
		return a.parseFailure(fmt.Errorf("invalid publisher_last_updated: %w", err))
	}
	if doc.SchemaVersion != "world-bank-canada-trade-v1" || doc.Publisher != "World Bank" || doc.SourceID == "" || doc.Geography != "CAN" || len(doc.Observations) == 0 {
		return a.parseFailure(fmt.Errorf("global trade snapshot metadata is incomplete or unsupported"))
	}

	result := &adapters.IngestionResult{}
	seen := make(map[string]struct{}, len(doc.Observations))
	for _, record := range doc.Observations {
		expectedUnit, allowed := indicatorUnits[record.MetricCode]
		if !allowed || record.Unit != expectedUnit || record.MetricName == "" || record.SourceURL == "" || record.Locator == "" || !isYear(record.ReferencePeriod) || math.IsNaN(record.Value) || math.IsInf(record.Value, 0) {
			return a.parseFailure(fmt.Errorf("invalid global trade observation %q", record.MetricCode))
		}
		key := record.MetricCode + ":" + record.ReferencePeriod
		if _, duplicate := seen[key]; duplicate {
			return a.parseFailure(fmt.Errorf("duplicate global trade observation %q", key))
		}
		seen[key] = struct{}{}
		if record.ScaleMin != nil && record.ScaleMax != nil && (*record.ScaleMax <= *record.ScaleMin || record.Value < *record.ScaleMin || record.Value > *record.ScaleMax) {
			return a.parseFailure(fmt.Errorf("global trade observation %q is outside its declared scale", key))
		}
		observedAt, err := time.Parse("2006-01-02", record.ReferencePeriod+"-12-31")
		if err != nil {
			return a.parseFailure(err)
		}
		hash, err := adapters.HashRecord(record)
		if err != nil {
			return a.parseFailure(err)
		}
		evidenceID := identity.StableID("evidence", adapterName, key+":"+hash)
		evidence := &domain.Evidence{
			ID: evidenceID, SourceURL: record.SourceURL, Publisher: doc.Publisher, SourceTier: domain.SourceTier1,
			RetrievalTimestamp: retrievedAt, PublicationDate: &publisherUpdated, EffectiveDate: &observedAt,
			Confidence: domain.ConfidenceVerified, ExtractionMethod: "official_api_normalized_observation",
			ContentHash: hash, HashScope: "normalized_source_record", SourceClass: "multilateral_official_statistics",
			SourceID: doc.SourceID, SourceVersionID: doc.PublisherLastUpdated, SourceRecordID: key,
			Locator: record.Locator, PipelineVersion: pipelineVersion, ParserVersion: parserVersion, MappingVersion: mappingVersion,
			RawSnippet: fmt.Sprintf("Canada %s: %s = %s %s", record.ReferencePeriod, record.MetricName, strconv.FormatFloat(record.Value, 'f', -1, 64), record.Unit),
		}
		metric := &domain.TradeMetric{
			ID: identity.StableID("trade-metric", adapterName, key), Geography: doc.Geography,
			MetricCode: record.MetricCode, MetricName: record.MetricName, ReferencePeriod: record.ReferencePeriod,
			Value: record.Value, Unit: record.Unit, ScaleMin: record.ScaleMin, ScaleMax: record.ScaleMax,
			EvidenceID: evidenceID, Evidence: evidence, ObservedAt: observedAt.UTC(),
		}
		result.Evidence = append(result.Evidence, evidence)
		result.TradeMetrics = append(result.TradeMetrics, metric)
		if metric.ObservedAt.After(a.health.LastChange) {
			a.health.LastChange = metric.ObservedAt
		}
	}
	if err := adapters.ValidateLineage(result); err != nil {
		return a.parseFailure(err)
	}
	a.health.DocumentsSeen = len(result.TradeMetrics)
	a.health.DocumentsChanged = len(result.TradeMetrics)
	return result, nil
}

func (a *Adapter) parseFailure(err error) (*adapters.IngestionResult, error) {
	a.health.ParseFailures++
	a.health.Status = string(domain.StatusDegraded)
	a.health.LastError = err.Error()
	return nil, err
}

func isYear(value string) bool {
	if len(value) != 4 {
		return false
	}
	_, err := strconv.Atoi(value)
	return err == nil
}

func readBounded(body io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response exceeds %d byte limit", limit)
	}
	return data, nil
}
