package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
)

func main() {
	fmt.Println("[INFO] Generating public datasets and CEGS snapshots...")
	store := database.NewMemoryStore()

	adapterList := []adapters.Adapter{
		official.NewAdapter("data/fixtures/official_records.json"),
	}

	pipeline := ingestion.NewPipeline(store, adapterList)
	ctx := context.Background()
	rep, err := pipeline.Run(ctx)
	if err != nil {
		fmt.Printf("[ERROR] Pipeline run error: %v\n", err)
		os.Exit(1)
	}

	_ = os.MkdirAll("data/public", 0755)
	_ = os.MkdirAll("data/cegs", 0755)

	projects, _, _ := store.ListProjects(ctx, database.ProjectFilter{Limit: 1000})
	entities, _ := store.ListEntities(ctx)
	events, _ := store.ListRecentEvents(ctx, 1000)

	// 1. Write data/public JSONL files
	writeJSONL("data/public/projects.jsonl", projects)
	writeJSONL("data/public/entities.jsonl", entities)
	writeJSONL("data/public/events.jsonl", events)

	// 2. Write data/cegs JSONL files (CEGS-compliant standard)
	var cegsProjects []*cegs.Project
	for _, p := range projects {
		cegsProjects = append(cegsProjects, cegs.ToCEGSProject(p, nil))
	}
	writeJSONL("data/cegs/projects.jsonl", cegsProjects)

	var cegsOrgs []*cegs.Organization
	for _, e := range entities {
		cegsOrgs = append(cegsOrgs, cegs.ToCEGSOrganization(e))
	}
	writeJSONL("data/cegs/organizations.jsonl", cegsOrgs)

	var cegsEvents []*cegs.Event
	for _, ev := range events {
		cegsEvents = append(cegsEvents, cegs.ToCEGSEvent(ev, "major-project"))
	}
	writeJSONL("data/cegs/events.jsonl", cegsEvents)

	// 3. Write data/cegs/manifest.json
	manifest := map[string]interface{}{
		"cegs":        cegs.SpecVersion,
		"id":          "cegs:manifest:ca:v0-1-snapshot",
		"type":        "manifest",
		"dataset_id":  "cegs-canada-national-v0-1",
		"title":       "CEGS Canonical National Economic Snapshot — v0.1",
		"publisher":   "CanadaOpportunityGraph Consortium",
		"license":     "CC-BY-4.0",
		"generated_at": time.Now().Format(time.RFC3339),
		"record_counts": map[string]int{
			"projects":      len(projects),
			"organizations": len(entities),
			"events":        len(events),
		},
		"jurisdictions": []string{"CA", "CA:ON", "CA:QC", "CA:BC", "CA:NU"},
	}
	manBytes, _ := json.MarshalIndent(manifest, "", "  ")
	_ = os.WriteFile("data/cegs/manifest.json", manBytes, 0644)

	fmt.Printf("[INFO] Snapshots created successfully: %d projects, %d entities in data/public & data/cegs (duration: %v)\n",
		rep.ProjectsIngested, rep.EntitiesResolved, rep.Duration)
}

func writeJSONL(filePath string, items interface{}) {
	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	valBytes, _ := json.Marshal(items)
	var list []interface{}
	_ = json.Unmarshal(valBytes, &list)

	for _, item := range list {
		line, _ := json.Marshal(item)
		_, _ = file.Write(line)
		_, _ = file.WriteString("\n")
	}
}
