package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/gazette"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/global_trade"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/indigenous"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/nrcan_major_projects"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/connector"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/gazettepoll"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/indigenouslinker"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/merkle"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/reconciliation"
)

func main() {
	log.Println("[INFO] CanadaOpportunityGraph Worker daemon started...")
	store := database.NewMemoryStore()

	adapterList := []adapters.Adapter{
		nrcan_major_projects.NewNRCanAdapter("data/fixtures/nrcan_mpi_2025.json"),
		official.NewAdapter(""),
		global_trade.NewFromEnv(),
		gazette.NewGazetteAdapter("ON", "data/fixtures/gazette_on.json"),
		gazette.NewGazetteAdapter("QC", "data/fixtures/gazette_qc.json"),
		gazette.NewGazetteAdapter("BC", "data/fixtures/gazette_bc.json"),
		indigenous.NewIndigenousAdapter("data/fixtures/isc_indigenous_business_directory.json"),
	}

	// Wrap adapters with production middleware via the connector registry.
	reg := connector.BuildRegistry(adapterList)
	pipelineAdapters := reg.BuildPipelineAdapters()
	pipeline := ingestion.NewPipeline(store, pipelineAdapters)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Scheduled gazette polling worker — runs on its own interval (5 minutes)
	// with change detection so unchanged gazettes are not re-parsed.
	gazetteWorker := gazettepoll.NewWorker(gazettepoll.DefaultConfig())
	go gazetteWorker.Run(ctx, func(results []gazettepoll.PollResult) {
		changed := 0
		for _, r := range results {
			if r.Err != nil {
				log.Printf("[GAZETTE-POLL] %s error: %v\n", r.Province, r.Err)
				continue
			}
			if r.Changed {
				changed++
			}
		}
		log.Printf("[GAZETTE-POLL] Cycle complete: %d/%d provinces changed\n", changed, len(results))
	})

	// Main ingestion pipeline — runs every 60 seconds.
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial run
	runIngestion(ctx, pipeline, store)

	for {
		select {
		case <-sigChan:
			log.Println("[INFO] Worker shutting down cleanly...")
			cancel()
			return
		case <-ticker.C:
			runIngestion(ctx, pipeline, store)
		}
	}
}

func runIngestion(ctx context.Context, p *ingestion.Pipeline, store database.Store) {
	log.Println("[INFO] Executing scheduled adapter ingestion cycle...")
	rep, err := p.Run(ctx)
	if err != nil {
		log.Printf("[ERROR] Ingestion cycle error: %v\n", err)
	} else {
		log.Printf("[INFO] Completed cycle in %v: %d projects updated, %d opportunities derived\n",
			rep.Duration, rep.ProjectsIngested, rep.OpportunitiesDerived)
	}

	// Re-run reconciliation and publish a fresh Merkle root each cycle so the
	// transparency log stays current with the ingested evidence.
	if reconReport := reconciliation.Reconcile(ctx, store); reconReport != nil {
		log.Printf("[INFO] Reconciliation: %d records, %d merges, %d conflicts\n",
			reconReport.TotalRecords, reconReport.Summary.Merged, reconReport.Summary.Conflicts)
	}
	if root := merkle.PublishRoot(ctx, store); root != "" {
		log.Printf("[INFO] Merkle root: %s\n", root)
	}

	// Cross-reference Indigenous businesses for set-asides and partnerships.
	if linkReport := indigenouslinker.Link(ctx, store); linkReport != nil {
		log.Printf("[INFO] Indigenous linker: %d projects, %d procurements scanned, %d relationships created\n",
			linkReport.ProjectsScanned, linkReport.ProcurementsScanned, linkReport.RelationshipsCreated)
	}
}

// Ensure adapters import is used by the package.
var _ = adapters.Adapter(nil)