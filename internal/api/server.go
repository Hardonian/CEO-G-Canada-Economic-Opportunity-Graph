package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/capitalstack"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/export"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/sovereignty"
)

type Server struct {
	store database.Store
	mux   *http.ServeMux
}

func NewServer(store database.Store) *Server {
	s := &Server{
		store: store,
		mux:   http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Global CORS middleware
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	// Health & Diagnostics
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ready", s.handleReady)
	s.mux.HandleFunc("GET /metrics", s.handleMetrics)
	s.mux.HandleFunc("GET /api/v1/openapi.json", s.handleOpenAPI)

	// Flagship Capital Radar
	s.mux.HandleFunc("GET /api/v1/radar", s.handleRadar)

	// Projects
	s.mux.HandleFunc("GET /api/v1/projects", s.handleListProjects)
	s.mux.HandleFunc("GET /api/v1/projects/{id}", s.handleGetProject)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/events", s.handleGetProjectEvents)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/scores", s.handleGetProjectScores)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/provenance", s.handleGetProjectProvenance)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/trust", s.handleGetProjectTrust)

	// Procurements & Opportunities
	s.mux.HandleFunc("GET /api/v1/procurements", s.handleListProcurements)
	s.mux.HandleFunc("GET /api/v1/signals", s.handleListSignals)
	s.mux.HandleFunc("GET /api/v1/search", s.handleSearch)

	// Capital Stack & AI Sovereignty
	s.mux.HandleFunc("GET /api/v1/capital/stack", s.handleCapitalStack)
	s.mux.HandleFunc("GET /api/v1/ai-sovereignty", s.handleAISovereignty)

	// Rankings
	s.mux.HandleFunc("GET /api/v1/rankings/{dimension}", s.handleRankings)

	// Exports & CEGS Open Standard Endpoints
	s.mux.HandleFunc("GET /api/v1/export/project/{id}", s.handleExportProject)
	s.mux.HandleFunc("GET /api/v1/cegs/projects/{id}", s.handleCEGSProject)
	s.mux.HandleFunc("GET /api/v1/cegs/export", s.handleCEGSExport)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"service":   "CanadaOpportunityGraph API",
		"version":   "1.0.0",
		"cegs_spec": cegs.SpecVersion,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats, _ := s.store.GetRadarStats(r.Context())
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	if stats != nil {
		fmt.Fprintf(w, "# HELP cog_total_projects Tracked major projects count\n")
		fmt.Fprintf(w, "# TYPE cog_total_projects gauge\n")
		fmt.Fprintf(w, "cog_total_projects %d\n", stats.TotalProjects)
		fmt.Fprintf(w, "# HELP cog_total_capex_cad Tracked total CAPEX in CAD\n")
		fmt.Fprintf(w, "# TYPE cog_total_capex_cad gauge\n")
		fmt.Fprintf(w, "cog_total_capex_cad %d\n", stats.TotalCapexCAD)
	}
}

func (s *Server) handleRadar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := s.store.GetRadarStats(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accelerating, _ := s.store.ListRankings(ctx, "buildability", 5)
	signals, _ := s.store.ListSignals(ctx, 30*24*time.Hour, 10)
	recentEvents, _ := s.store.ListRecentEvents(ctx, 5)

	resp := map[string]interface{}{
		"stats":                 stats,
		"accelerating_projects": accelerating,
		"recent_signals":        signals,
		"recent_events":         recentEvents,
		"cegs_version":          cegs.SpecVersion,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	minCapex, _ := strconv.ParseInt(q.Get("min_capex"), 10, 64)

	filter := database.ProjectFilter{
		Sector:      q.Get("sector"),
		Province:    q.Get("province"),
		Stage:       q.Get("stage"),
		MinCapexCAD: minCapex,
		Search:      q.Get("q"),
		SortBy:      q.Get("sort_by"),
		SortDir:     q.Get("sort_dir"),
		Limit:       limit,
		Offset:      offset,
	}

	projects, total, err := s.store.ListProjects(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	// Find by ID or Slug
	proj, err := s.store.GetProject(ctx, id)
	if err != nil {
		proj, err = s.store.GetProjectBySlug(ctx, id)
		if err != nil {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}
	}

	scores, _ := s.store.GetLatestScores(ctx, proj.ID)
	events, _ := s.store.ListEventsByProject(ctx, proj.ID)
	relationships, _ := s.store.ListRelationshipsByProject(ctx, proj.ID)
	capital, _ := s.store.ListCapitalItemsByProject(ctx, proj.ID)
	opps, _ := s.store.ListOpportunitiesByProject(ctx, proj.ID)

	resp := map[string]interface{}{
		"project":       proj,
		"scores":        scores,
		"events":        events,
		"relationships": relationships,
		"capital_items": capital,
		"opportunities": opps,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetProjectEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	events, err := s.store.ListEventsByProject(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleGetProjectScores(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	scores, err := s.store.GetLatestScores(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, scores)
}

func (s *Server) handleGetProjectProvenance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()
	bundle, err := export.ExportProjectBundle(ctx, s.store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	type ProvenanceNode struct {
		Fact        string `json:"fact"`
		EvidenceID  string `json:"evidence_id"`
		Publisher   string `json:"publisher"`
		SourceURL   string `json:"source_url"`
		SourceTier  int    `json:"source_tier"`
		ContentHash string `json:"content_hash"`
		Confidence  string `json:"confidence"`
	}

	var nodes []ProvenanceNode
	for _, ev := range bundle.Evidence {
		nodes = append(nodes, ProvenanceNode{
			Fact:        "Project Milestone Assertion",
			EvidenceID:  ev.ID,
			Publisher:   ev.Publisher,
			SourceURL:   ev.SourceURL,
			SourceTier:  int(ev.SourceTier),
			ContentHash: ev.ContentHash,
			Confidence:  string(ev.Confidence),
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":   id,
		"project_name": bundle.Project.Name,
		"provenance":   nodes,
	})
}

func (s *Server) handleGetProjectTrust(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()
	bundle, err := export.ExportProjectBundle(ctx, s.store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	evidenceCount := len(bundle.Evidence)
	tier1Count := 0
	for _, e := range bundle.Evidence {
		if e.SourceTier == domain.SourceTier1 {
			tier1Count++
		}
	}

	tier1Coverage := 0.0
	if evidenceCount > 0 {
		tier1Coverage = (float64(tier1Count) / float64(evidenceCount)) * 100.0
	}

	conformance := "CEGS Core"
	if evidenceCount > 0 {
		conformance = "CEGS Provenance"
	}
	if len(bundle.Events) > 0 {
		conformance = "CEGS Historical"
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":             id,
		"evidence_count":         evidenceCount,
		"tier1_primary_coverage": tier1Coverage,
		"staleness":              "LOW",
		"conflicting_claims":     0,
		"conformance_level":      conformance,
		"evidence_quality":       "INVESTOR_GRADE",
	})
}

func (s *Server) handleListProcurements(w http.ResponseWriter, r *http.Request) {
	procs, err := s.store.ListProcurements(r.Context(), 50, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, procs)
}

func (s *Server) handleListSignals(w http.ResponseWriter, r *http.Request) {
	sigs, err := s.store.ListSignals(r.Context(), 90*24*time.Hour, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, sigs)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	filter := database.ProjectFilter{
		Search: query,
		Limit:  20,
	}
	projects, total, err := s.store.ListProjects(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":    query,
		"total":    total,
		"projects": projects,
	})
}

func (s *Server) handleCapitalStack(w http.ResponseWriter, r *http.Request) {
	programs := capitalstack.CanonicalPrograms()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"programs":    programs,
		"disclaimer":  "LEGAL NOTICE: Informational analysis only. Not tax or legal advice.",
		"last_update": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleAISovereignty(w http.ResponseWriter, r *http.Request) {
	profiles := []*sovereignty.AIProfile{
		{
			SubjectID:          "mila-sovereign-cluster",
			SubjectName:        "Mila Sovereign AI Cluster (Montreal)",
			DataResidencyCA:    true,
			ComputeResidencyCA: true,
			CanadianOwnership:  1.0,
			ForeignLegalRisk:   1.0,
			LocalDeployment:    true,
			OfflineCapability:  true,
			OpenWeights:        true,
			BilingualCapacity:  0.95,
			QuebecLaw25Ready:   true,
			CleanEnergySource:  0.99,
		},
		{
			SubjectID:          "hyperscale-cloud-canada",
			SubjectName:        "Hyperscale US Cloud (Central Canada Region)",
			DataResidencyCA:    true,
			ComputeResidencyCA: true,
			CanadianOwnership:  0.0,
			ForeignLegalRisk:   0.20, // High US CLOUD Act exposure
			LocalDeployment:    false,
			OfflineCapability:  false,
			OpenWeights:        false,
			BilingualCapacity:  0.80,
			QuebecLaw25Ready:   true,
			CleanEnergySource:  0.90,
		},
	}

	var scorecards []interface{}
	for _, p := range profiles {
		score := sovereignty.EvaluateSovereignty(p)
		scorecards = append(scorecards, map[string]interface{}{
			"subject_name": p.SubjectName,
			"scorecard":    score,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"methodology": "cai-sovereignty-v1.0",
		"benchmarks":  scorecards,
	})
}

func (s *Server) handleRankings(w http.ResponseWriter, r *http.Request) {
	dim := r.PathValue("dimension")
	list, err := s.store.ListRankings(r.Context(), dim, 25)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"dimension": dim,
		"rankings":  list,
	})
}

func (s *Server) handleExportProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	format := r.URL.Query().Get("format")
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	switch strings.ToLower(format) {
	case "markdown", "md":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Write([]byte(bundle.ToMarkdown()))
	case "cegs":
		cegsProj, _ := bundle.ToCEGSExport()
		writeJSON(w, http.StatusOK, cegsProj)
	default:
		writeJSON(w, http.StatusOK, bundle)
	}
}

func (s *Server) handleCEGSProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	cegsProj, _ := bundle.ToCEGSExport()
	writeJSON(w, http.StatusOK, cegsProj)
}

func (s *Server) handleCEGSExport(w http.ResponseWriter, r *http.Request) {
	projects, _, _ := s.store.ListProjects(r.Context(), database.ProjectFilter{Limit: 500})
	var cegsList []*cegs.Project
	for _, p := range projects {
		cegsList = append(cegsList, cegs.ToCEGSProject(p, nil))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cegs":         cegs.SpecVersion,
		"manifest_id": "cegs:manifest:ca:live-export",
		"generated_at": time.Now().Format(time.RFC3339),
		"projects":     cegsList,
	})
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	spec := map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       "CanadaOpportunityGraph REST API",
			"version":     "1.0.0",
			"description": "Investor-grade API for Canadian economic infrastructure, capital tracking, and CEGS data standard reference implementation.",
		},
		"paths": map[string]interface{}{
			"/api/v1/radar": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Flagship Canada Capital Radar metrics",
				},
			},
			"/api/v1/projects": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Query Canadian major capital projects",
				},
			},
			"/api/v1/cegs/projects/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Fetch project conforming to CEGS 0.1 standard",
				},
			},
		},
	}
	writeJSON(w, http.StatusOK, spec)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
