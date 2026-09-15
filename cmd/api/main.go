package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/global_trade"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/nrcan_major_projects"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/api"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/config"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/connector"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/merkle"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/reconciliation"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/sources"
)

func main() {
	cfg, err := config.LoadValidated()
	if err != nil {
		log.Fatalf("[FATAL] Invalid runtime configuration: %v", err)
	}
	processContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	store := database.NewMemoryStore()

	// Register authoritative adapters (raw, unwrapped).
	adapterList := []adapters.Adapter{
		nrcan_major_projects.NewNRCanAdapter("data/fixtures/nrcan_mpi_2025.json"),
		official.NewAdapter(""),
		global_trade.NewFromEnv(),
	}

	// Wrap with production middleware (retry, circuit-breaker, cache, dedup,
	// instrumentation) via the connector registry. The registry preserves
	// insertion order and exposes aggregate health.
	reg := connector.BuildRegistry(adapterList)
	pipelineAdapters := reg.BuildPipelineAdapters()
	pipeline := ingestion.NewPipeline(store, pipelineAdapters)
	ingestContext, cancelIngest := context.WithTimeout(processContext, cfg.InitialIngestTimeout)

	log.Println("[INFO] Bootstrapping initial ingestion from authoritative adapters...")
	report, err := pipeline.Run(ingestContext)
	cancelIngest()
	if err != nil {
		log.Printf("[WARN] Ingestion warning: %v\n", err)
	} else {
		log.Printf("[INFO] Ingestion complete: %d projects, %d entities, %d opportunities in %v\n",
			report.ProjectsIngested, report.EntitiesResolved, report.OpportunitiesDerived, report.Duration)
	}
	if processContext.Err() != nil {
		log.Println("[INFO] Shutdown signal received during initial ingestion")
		return
	}
	stats, statsErr := store.GetRadarStats(processContext)
	if statsErr != nil || stats == nil || stats.TotalProjects == 0 {
		log.Fatalf("[FATAL] Authoritative snapshot bootstrap produced no usable projects")
	}

	// Run multi-jurisdiction reconciliation across ingested records.
	reconReport := reconciliation.Reconcile(processContext, store)
	log.Printf("[INFO] Reconciliation: %d records, %d merges, %d links, %d conflicts\n",
		reconReport.TotalRecords, reconReport.Summary.Merged, reconReport.Summary.Linked, reconReport.Summary.Conflicts)

	// Publish daily Merkle root of all evidence hashes to the transparency log.
	merkleRoot := merkle.PublishRoot(processContext, store)
	log.Printf("[INFO] Merkle transparency root published: %s\n", merkleRoot)

	// Load public data mesh configuration (publishers, sources, policies).
	if meshConfig, meshErr := sources.LoadDirectory("sources"); meshErr != nil {
		log.Printf("[WARN] Data mesh configuration: %v\n", meshErr)
	} else if meshErr := sources.ApplyConfiguration(processContext, store, meshConfig, time.Now().UTC()); meshErr != nil {
		log.Printf("[WARN] Data mesh configuration: %v\n", meshErr)
	}

	log.Printf("[INFO] Runtime configuration: %s", cfg)

	server, err := api.NewServerWithOptions(store, api.Options{
		AllowedOrigins:      cfg.CORSOrigins,
		EnableHSTS:          cfg.EnableHSTS,
		TrustedProxyCIDRs:   cfg.TrustedProxyCIDRs,
		MaxRequestBodyBytes: cfg.MaxRequestBodyBytes,
		RateLimitPerMinute:  cfg.RateLimitPerMinute,
		RateLimitBurst:      cfg.RateLimitBurst,
		RequestTimeout:      cfg.RequestTimeout,
		ReadinessTimeout:    cfg.ReadinessTimeout,
		Logger:              log.Default(),
	})
	if err != nil {
		log.Fatalf("[FATAL] Invalid API security configuration: %v", err)
	}

	httpServer := &http.Server{
		Addr:              cfg.ListenAddress(),
		Handler:           server,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	log.Printf("[INFO] CanadaOpportunityGraph API running on %s\n", cfg.ListenAddress())
	log.Printf("[INFO] CEGS 0.1 Specification active at /api/v1/cegs/export\n")

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	select {
	case serveErr := <-serverErrors:
		if serveErr != nil && serveErr != http.ErrServerClosed {
			log.Fatalf("[FATAL] HTTP server error: %v", serveErr)
		}
		return
	case <-processContext.Done():
		log.Println("[INFO] Shutdown signal received; draining API requests...")
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancelShutdown()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		log.Printf("[ERROR] Graceful shutdown failed: %v", err)
		if closeErr := httpServer.Close(); closeErr != nil {
			log.Printf("[ERROR] Forced server close failed: %v", closeErr)
		}
	}
}
