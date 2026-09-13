package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/api"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/config"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
)

func main() {
	cfg := config.Load()
	store := database.NewMemoryStore()

	// Register authoritative adapters
	adapterList := []adapters.Adapter{
		official.NewAdapter(""),
	}

	pipeline := ingestion.NewPipeline(store, adapterList)
	ctx := context.Background()

	log.Println("[INFO] Bootstrapping initial ingestion from authoritative adapters...")
	report, err := pipeline.Run(ctx)
	if err != nil {
		log.Printf("[WARN] Ingestion warning: %v\n", err)
	} else {
		log.Printf("[INFO] Ingestion complete: %d projects, %d entities, %d opportunities in %v\n",
			report.ProjectsIngested, report.EntitiesResolved, report.OpportunitiesDerived, report.Duration)
	}

	server := api.NewServer(store)
	addr := fmt.Sprintf(":%d", cfg.Port)

	log.Printf("[INFO] CanadaOpportunityGraph API running on %s\n", addr)
	log.Printf("[INFO] CEGS 0.1 Specification active at /api/v1/cegs/export\n")

	if err := http.ListenAndServe(addr, server); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[FATAL] HTTP server error: %v\n", err)
		os.Exit(1)
	}
}
