package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/nrcan_major_projects"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	graphExport "github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/export"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
)

const (
	datasetVersion = "2026-09-13.1"
	datasetTime    = "2026-09-13T00:00:00Z"
)

func main() {
	fmt.Println("[INFO] Generating reviewed public datasets and CEGS snapshots...")
	store := database.NewMemoryStore()
	pipeline := ingestion.NewPipeline(store, []adapters.Adapter{
		nrcan_major_projects.NewNRCanAdapter("data/fixtures/nrcan_mpi_2025.json"),
		official.NewAdapter("data/fixtures/official_records.json"),
	})
	ctx := context.Background()
	report, err := pipeline.Run(ctx)
	must(err)

	projects, _, err := store.ListProjects(ctx, database.ProjectFilter{Limit: 1000})
	must(err)
	entities, err := store.ListEntities(ctx)
	must(err)
	events, err := store.ListRecentEvents(ctx, 1000)
	must(err)
	sort.Slice(projects, func(i, j int) bool { return projects[i].ID < projects[j].ID })
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID })
	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })

	evidenceByID := make(map[string]*domain.Evidence)
	var scores []*domain.ProjectScore
	for _, project := range projects {
		for _, evidenceID := range project.EvidenceIDs {
			evidence, err := store.GetEvidence(ctx, evidenceID)
			must(err)
			evidenceByID[evidence.ID] = evidence
		}
		history, err := store.ListScoreHistory(ctx, project.ID, "")
		must(err)
		scores = append(scores, history...)
	}
	var evidence []*domain.Evidence
	for _, item := range evidenceByID {
		evidence = append(evidence, item)
	}
	sort.Slice(evidence, func(i, j int) bool { return evidence[i].ID < evidence[j].ID })
	sort.Slice(scores, func(i, j int) bool { return scores[i].ID < scores[j].ID })

	var cegsProjects []*cegs.Project
	for _, project := range projects {
		cegsProjects = append(cegsProjects, cegs.ToCEGSProject(project, project.EvidenceIDs))
	}
	var cegsOrgs []*cegs.Organization
	for _, entity := range entities {
		cegsOrgs = append(cegsOrgs, cegs.ToCEGSOrganization(entity))
	}
	var cegsEvents []*cegs.Event
	for _, event := range events {
		cegsEvents = append(cegsEvents, cegs.ToCEGSEvent(event, event.ProjectID))
	}
	var cegsEvidence []*cegs.Evidence
	for _, item := range evidence {
		cegsEvidence = append(cegsEvidence, cegs.ToCEGSEvidence(item))
	}

	files := map[string][]byte{}
	files["public/projects.jsonl"] = mustJSONL(projects)
	files["public/entities.jsonl"] = mustJSONL(entities)
	files["public/events.jsonl"] = mustJSONL(events)
	files["public/evidence.jsonl"] = mustJSONL(evidence)
	files["public/scores.jsonl"] = mustJSONL(scores)
	files["cegs/projects.jsonl"] = mustJSONL(cegsProjects)
	files["cegs/organizations.jsonl"] = mustJSONL(cegsOrgs)
	files["cegs/events.jsonl"] = mustJSONL(cegsEvents)
	files["cegs/evidence.jsonl"] = mustJSONL(cegsEvidence)
	csvData, err := graphExport.ToCSV(projects)
	must(err)
	files["public/projects.csv"] = []byte(csvData)
	geoJSON, err := graphExport.ToGeoJSON(projects)
	must(err)
	files["public/projects.geojson"] = append(geoJSON, '\n')

	checksums := make(map[string]string, len(files))
	for name, data := range files {
		checksums[name] = checksum(data)
	}
	generatedAt, err := time.Parse(time.RFC3339, datasetTime)
	must(err)
	jurisdictionSet := map[string]struct{}{"CA": {}}
	for _, project := range projects {
		if project.Province != "" && project.Province != "Federal" {
			jurisdictionSet["CA:"+project.Province] = struct{}{}
		}
	}
	jurisdictions := make([]string, 0, len(jurisdictionSet))
	for value := range jurisdictionSet {
		jurisdictions = append(jurisdictions, value)
	}
	sort.Strings(jurisdictions)

	manifest := map[string]any{
		"cegs":             cegs.SpecVersion,
		"id":               "cegs:manifest:ca:" + datasetVersion,
		"type":             "manifest",
		"dataset_id":       "cegs-canada-reviewed-primary-sources",
		"dataset_version":  datasetVersion,
		"title":            "CEGS reviewed Canadian economic project snapshot",
		"description":      "A nationwide planning snapshot combining the NRCan Major Projects Inventory with individually reviewed primary-source records; it is not a complete census of Canadian projects.",
		"publisher":        "CanadaOpportunityGraph",
		"license":          "LicenseRef-COG-Generated-Data",
		"generated_at":     generatedAt.Format(time.RFC3339),
		"record_counts":    map[string]int{"projects": len(projects), "organizations": len(entities), "events": len(events), "evidence": len(evidence), "scores": len(scores)},
		"jurisdictions":    jurisdictions,
		"checksums_sha256": checksums,
		"coverage_note":    "Coverage includes the 2025-2035 NRCan Major Projects Inventory point layer plus curated primary-source records. It remains incomplete and source-reported values are not independently audited.",
		"source_mode":      "OFFICIAL_SNAPSHOT_ENSEMBLE",
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	must(err)
	manifestBytes = append(manifestBytes, '\n')
	files["cegs/manifest.json"] = manifestBytes
	files["public/manifest.json"] = manifestBytes

	for name, data := range files {
		must(writeCurrent(filepath.Join("data", filepath.FromSlash(name)), data))
	}
	releaseRoot := filepath.Join("data", "releases", datasetVersion)
	for name, data := range files {
		must(writeHistorical(filepath.Join(releaseRoot, filepath.FromSlash(name)), data))
	}

	fmt.Printf("[INFO] Snapshot %s created: %d projects, %d evidence records, %d scores (status %s, duration %v)\n", datasetVersion, len(projects), len(evidence), len(scores), report.Status, report.Duration)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR]", err)
		os.Exit(1)
	}
}

func mustJSONL[T any](items []T) []byte {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	for _, item := range items {
		must(encoder.Encode(item))
	}
	return out.Bytes()
}

func writeCurrent(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func writeHistorical(path string, data []byte) error {
	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, data) {
			return nil
		}
		return fmt.Errorf("refusing to overwrite historical release %s", path)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func checksum(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
