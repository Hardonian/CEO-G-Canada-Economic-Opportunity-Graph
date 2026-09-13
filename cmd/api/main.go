package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/nrcan_major_projects"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/official"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/api"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/config"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/ingestion"
)

func main() {
	cfg, err := config.LoadValidated()
	if err != nil {
		log.Fatalf("[FATAL] Invalid runtime configuration: %v", err)
	}
	processContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	store := database.NewMemoryStore()

	// Register authoritative adapters
	adapterList := []adapters.Adapter{
		nrcan_major_projects.NewNRCanAdapter("data/fixtures/nrcan_mpi_2025.json"),
		official.NewAdapter(""),
	}

	pipeline := ingestion.NewPipeline(store, adapterList)
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
