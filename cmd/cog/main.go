package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/canadabuys"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/cer"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/iaac"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/ideas_defence"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/nrcan_major_projects"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/export"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "help", "-h", "--help":
		printUsage()
	case "cegs":
		handleCEGS(os.Args[2:])
	case "search":
		handleSearch(os.Args[2:])
	case "project":
		handleProject(os.Args[2:])
	case "changes":
		handleChanges(os.Args[2:])
	case "rankings":
		handleRankings(os.Args[2:])
	case "export":
		handleExport(os.Args[2:])
	case "demo":
		handleDemo()
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("CanadaOpportunityGraph CLI (cog) — Economic Intelligence & CEGS Reference Tool")
	fmt.Println("\nUsage:")
	fmt.Println("  cog search <query>                     Search projects and infrastructure assets")
	fmt.Println("  cog project show <id|slug>             Display investor-grade project profile")
	fmt.Println("  cog changes [--since 7d|30d]           List recent momentum signals and milestones")
	fmt.Println("  cog rankings <dimension>               Rank projects (buildability, investability, etc.)")
	fmt.Println("  cog export project <id> [--format json|md|cegs] Export dossier with provenance")
	fmt.Println("  cog cegs validate <file>               Validate document against CEGS 0.1 standard")
	fmt.Println("  cog cegs inspect <file>                Inspect CEGS document & evidence trust profile")
	fmt.Println("  cog cegs diff <old.json> <new.json>    Semantic diff between two CEGS states")
	fmt.Println("  cog demo                               Run instant offline demonstration")
}

func getSeededStore() database.Store {
	store := database.NewMemoryStore()
	adapterList := []adapters.Adapter{
		iaac.NewIAACAdapter(""),
		canadabuys.NewCanadaBuysAdapter(""),
		nrcan_major_projects.NewNRCanAdapter(""),
		ideas_defence.NewIDEaSAdapter(""),
		cer.NewCERAdapter(""),
	}
	p := ingestion.NewPipeline(store, adapterList)
	_, _ = p.Run(context.Background())
	return store
}

func handleCEGS(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cog cegs <validate|inspect|diff> [args...]")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "validate":
		if len(args) < 2 {
			fmt.Println("Usage: cog cegs validate <file.json>")
			return
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		rep, err := cegs.Validate(data)
		if err != nil {
			fmt.Printf("Validation error: %v\n", err)
			os.Exit(1)
		}
		if rep.Valid {
			fmt.Printf("✓ %s\n", rep.Summary)
			fmt.Printf("  ID:           %s\n", rep.ID)
			fmt.Printf("  Type:         %s\n", rep.ResourceType)
			fmt.Printf("  Conformance:  %s\n", rep.ConformanceLevel)
		} else {
			fmt.Printf("✗ %s\n", rep.Summary)
			for _, e := range rep.Errors {
				fmt.Printf("  - Error: %s\n", e)
			}
			os.Exit(1)
		}

	case "inspect":
		if len(args) < 2 {
			fmt.Println("Usage: cog cegs inspect <file.json>")
			return
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		insp, err := cegs.Inspect(data)
		if err != nil {
			fmt.Printf("Inspection error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CEGS %s | Type: %s\n", insp.CEGSVersion, insp.Type)
		fmt.Printf("Name:        %s\n", insp.CanonicalName)
		if insp.Stage != "" {
			fmt.Printf("Stage:       %s\n", insp.Stage)
		}
		fmt.Printf("Evidence:    %d reference(s)\n", insp.EvidenceCount)
		fmt.Printf("Conformance: %s\n", insp.ConformanceLevel)

	case "diff":
		if len(args) < 3 {
			fmt.Println("Usage: cog cegs diff <old.json> <new.json>")
			return
		}
		oldData, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("Error reading old file: %v\n", err)
			os.Exit(1)
		}
		newData, err := os.ReadFile(args[2])
		if err != nil {
			fmt.Printf("Error reading new file: %v\n", err)
			os.Exit(1)
		}
		diff, err := cegs.Diff(oldData, newData)
		if err != nil {
			fmt.Printf("Diff error: %v\n", err)
			os.Exit(1)
		}
		if !diff.HasChanges {
			fmt.Println("No semantic differences detected between states.")
			return
		}
		fmt.Printf("Detected %d semantic change(s) in %s:\n", len(diff.Changes), diff.ResourceID)
		for _, ch := range diff.Changes {
			fmt.Printf("  [%s] %s\n", ch.Code, ch.Description)
		}

	default:
		fmt.Printf("Unknown cegs command: %s\n", subCmd)
	}
}

func handleSearch(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: cog search <query>")
		return
	}
	query := strings.Join(args, " ")
	store := getSeededStore()

	projects, total, err := store.ListProjects(context.Background(), database.ProjectFilter{
		Search: query,
		Limit:  15,
	})
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		return
	}

	fmt.Printf("Found %d project(s) matching '%s':\n\n", total, query)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SLUG\tNAME\tSECTOR\tPROV\tSTAGE\tCAPEX (CAD)\tBUILDABILITY")
	for _, p := range projects {
		bScore := 0.0
		if p.Scores != nil {
			bScore = p.Scores["buildability"]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t$%d\t%.1f/100\n",
			p.Slug, p.Name, p.Sector, p.Province, p.CurrentStage, p.CapexCAD, bScore)
	}
	w.Flush()
}

func handleProject(args []string) {
	if len(args) < 2 || args[0] != "show" {
		fmt.Println("Usage: cog project show <id|slug>")
		return
	}
	id := args[1]
	store := getSeededStore()
	ctx := context.Background()

	bundle, err := export.ExportProjectBundle(ctx, store, id)
	if err != nil {
		proj, err2 := store.GetProjectBySlug(ctx, id)
		if err2 != nil {
			fmt.Printf("Project not found: %s\n", id)
			return
		}
		bundle, err = export.ExportProjectBundle(ctx, store, proj.ID)
		if err != nil {
			fmt.Printf("Error exporting project: %v\n", err)
			return
		}
	}

	fmt.Println(bundle.ToMarkdown())
}

func handleChanges(args []string) {
	store := getSeededStore()
	signals, _ := store.ListSignals(context.Background(), 30*24*time.Hour, 15)

	fmt.Printf("Recent Capital and Milestone Signals (Last 30 Days):\n\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tPROJECT\tSIGNAL TYPE\tDESCRIPTION")
	for _, s := range signals {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			s.Timestamp.Format("2006-01-02"), s.ProjectName, s.Type, s.Description)
	}
	w.Flush()
}

func handleRankings(args []string) {
	dim := "buildability"
	if len(args) > 0 {
		dim = strings.ToLower(args[0])
	}
	store := getSeededStore()
	projects, _ := store.ListRankings(context.Background(), dim, 10)

	fmt.Printf("Top Projects by %s:\n\n", strings.Title(dim))
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "RANK\tNAME\tSECTOR\tPROVINCE\tSCORE\tCAPEX (CAD)")
	for i, p := range projects {
		score := 0.0
		if p.Scores != nil {
			score = p.Scores[dim]
		}
		fmt.Fprintf(w, "#%d\t%s\t%s\t%s\t%.1f\t$%d\n",
			i+1, p.Name, p.Sector, p.Province, score, p.CapexCAD)
	}
	w.Flush()
}

func handleExport(args []string) {
	if len(args) < 2 || args[0] != "project" {
		fmt.Println("Usage: cog export project <id|slug> [--format json|md|cegs]")
		return
	}
	id := args[1]
	format := "json"
	for i, arg := range args {
		if arg == "--format" && i+1 < len(args) {
			format = args[i+1]
		}
	}

	store := getSeededStore()
	ctx := context.Background()

	proj, err := store.GetProject(ctx, id)
	if err != nil {
		proj, err = store.GetProjectBySlug(ctx, id)
		if err != nil {
			fmt.Printf("Project not found: %s\n", id)
			return
		}
	}

	bundle, err := export.ExportProjectBundle(ctx, store, proj.ID)
	if err != nil {
		fmt.Printf("Export error: %v\n", err)
		return
	}

	switch format {
	case "md", "markdown":
		fmt.Println(bundle.ToMarkdown())
	case "cegs":
		cegsProj, _ := bundle.ToCEGSExport()
		out, _ := json.MarshalIndent(cegsProj, "", "  ")
		fmt.Println(string(out))
	default:
		out, _ := json.MarshalIndent(bundle, "", "  ")
		fmt.Println(string(out))
	}
}

func handleDemo() {
	fmt.Println("=== CanadaOpportunityGraph Deterministic Demo Mode ===")
	store := getSeededStore()
	ctx := context.Background()

	stats, _ := store.GetRadarStats(ctx)
	fmt.Printf("Tracked Projects:     %d\n", stats.TotalProjects)
	fmt.Printf("Total Tracked CAPEX:  $%.2f Billion CAD\n", float64(stats.TotalCapexCAD)/1e9)
	fmt.Printf("Capital Moving Week:  $%.2f Million CAD\n", float64(stats.CapitalMovingWeekCAD)/1e6)
	fmt.Printf("Accelerating Assets:  %d\n", stats.AcceleratingProjectsCount)
	fmt.Printf("Active Procurements:  %d\n\n", stats.ActiveProcurementsCount)

	fmt.Println("Top Ranked Projects (Buildability):")
	rankings, _ := store.ListRankings(ctx, "buildability", 3)
	for i, p := range rankings {
		fmt.Printf("  %d. %s (%s, %s) — Buildability: %.1f/100, CAPEX: $%.1fB\n",
			i+1, p.Name, p.Province, p.Sector, p.Scores["buildability"], float64(p.CapexCAD)/1e9)
	}

	fmt.Println("\nCEGS Standard Specification: 0.1 | Reference Implementation Verified.")
}
