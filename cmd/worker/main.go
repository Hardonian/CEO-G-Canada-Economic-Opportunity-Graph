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
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/indigenous"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/nrcan_major_projects"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
)

func main() {
	log.Println("[INFO] CanadaOpportunityGraph Worker daemon started...")
	store := database.NewMemoryStore()

	adapterList := []adapters.Adapter{
		nrcan_major_projects.NewNRCanAdapter("data/fixtures/nrcan_mpi_2025.json"),
		official.NewAdapter(""),
		gazette.NewGazetteAdapter("ON", "data/fixtures/gazette_on.json"),
		gazette.NewGazetteAdapter("QC", "data/fixtures/gazette_qc.json"),
		gazette.NewGazetteAdapter("BC", "data/fixtures/gazette_bc.json"),
		indigenous.NewIndigenousAdapter("data/fixtures/isc_indigenous_business_directory.json"),
	}

	pipeline := ingestion.NewPipeline(store, adapterList)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Initial run
	runIngestion(ctx, pipeline)

	for {
		select {
		case <-sigChan:
			log.Println("[INFO] Worker shutting down cleanly...")
			return
		case <-ticker.C:
			runIngestion(ctx, pipeline)
		}
	}
}

func runIngestion(ctx context.Context, p *ingestion.Pipeline) {
	log.Println("[INFO] Executing scheduled adapter ingestion cycle...")
	rep, err := p.Run(ctx)
	if err != nil {
		log.Printf("[ERROR] Ingestion cycle error: %v\n", err)
	} else {
		log.Printf("[INFO] Completed cycle in %v: %d projects updated, %d opportunities derived\n",
			rep.Duration, rep.ProjectsIngested, rep.OpportunitiesDerived)
	}
}
