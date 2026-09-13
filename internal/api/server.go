package api

import (
	"context"
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
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/trust"
	"github.com/google/uuid"
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
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" || len(requestID) > 128 { requestID = uuid.NewString() }
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	// This service is a public read-only API. Mutation origins are not allowed.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, X-Request-ID")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			writeError(w, r, http.StatusInternalServerError, "internal_error", "The request could not be completed.")
		}
	}()
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
	s.mux.HandleFunc("GET /api/v1/projects/{id}/scores/history", s.handleGetProjectScoreHistory)
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
		writeError(w, r, http.StatusServiceUnavailable, "radar_unavailable", "Radar aggregates are temporarily unavailable.")
		return
	}

	accelerating, err := s.store.ListRankings(ctx, "buildability", 5); if err != nil { accelerating = nil }
	signalList, err := s.store.ListSignals(ctx, 30*24*time.Hour, 10); if err != nil { signalList = nil }
	recentEvents, err := s.store.ListRecentEvents(ctx, 5); if err != nil { recentEvents = nil }

	resp := map[string]interface{}{
		"stats":                 stats,
		"accelerating_projects": accelerating,
		"recent_signals":        signalList,
		"recent_events":         recentEvents,
		"cegs_version":          cegs.SpecVersion,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := boundedInt(q.Get("limit"), 50, 1, 500)
	if err != nil { writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error()); return }
	offset, err := boundedInt(q.Get("offset"), 0, 0, 1_000_000)
	if err != nil { writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error()); return }
	minCapex, err := optionalNonNegativeInt64(q.Get("min_capex"))
	if err != nil { writeError(w, r, http.StatusBadRequest, "invalid_min_capex", err.Error()); return }
	if q.Get("sort_dir") != "" && q.Get("sort_dir") != "asc" && q.Get("sort_dir") != "desc" { writeError(w, r, http.StatusBadRequest, "invalid_sort_dir", "sort_dir must be asc or desc"); return }

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
		writeError(w, r, http.StatusServiceUnavailable, "projects_unavailable", "Projects are temporarily unavailable.")
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
	proj, err := s.resolveProject(ctx, id)
	if err != nil { writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found."); return }

	scores, err := s.store.GetLatestScores(ctx, proj.ID); if err != nil { writeError(w, r, 503, "scores_unavailable", "Scores are temporarily unavailable."); return }
	events, err := s.store.ListEventsByProject(ctx, proj.ID); if err != nil { writeError(w, r, 503, "events_unavailable", "Events are temporarily unavailable."); return }
	relationships, err := s.store.ListRelationshipsByProject(ctx, proj.ID); if err != nil { writeError(w, r, 503, "relationships_unavailable", "Relationships are temporarily unavailable."); return }
	capital, err := s.store.ListCapitalItemsByProject(ctx, proj.ID); if err != nil { writeError(w, r, 503, "capital_unavailable", "Capital records are temporarily unavailable."); return }
	opps, err := s.store.ListOpportunitiesByProject(ctx, proj.ID); if err != nil { writeError(w, r, 503, "opportunities_unavailable", "Opportunities are temporarily unavailable."); return }

	resp := map[string]interface{}{
		"project":       proj,
		"scores":        scores,
		"events":        events,
		"relationships": relationships,
		"capital_items": capital,
		"opportunities": opps,
		"status":        domain.StatusHealthy,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetProjectEvents(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil { writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found."); return }
	events, err := s.store.ListEventsByProject(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "events_unavailable", "Events are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleGetProjectScores(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil { writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found."); return }
	scores, err := s.store.GetLatestScores(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "scores_unavailable", "Scores are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, scores)
}

func (s *Server) handleGetProjectScoreHistory(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil { writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found."); return }
	history, err := s.store.ListScoreHistory(r.Context(), project.ID, r.URL.Query().Get("type"))
	if err != nil { writeError(w, r, http.StatusServiceUnavailable, "score_history_unavailable", "Score history is temporarily unavailable."); return }
	writeJSON(w, http.StatusOK, map[string]any{"project_id": project.ID, "history": history})
}

func (s *Server) handleGetProjectProvenance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()
	bundle, err := export.ExportProjectBundle(ctx, s.store, id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
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
		"project_id":   bundle.Project.ID,
		"project_name": bundle.Project.Name,
		"provenance":   nodes,
	})
}

func (s *Server) handleGetProjectTrust(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()
	bundle, err := export.ExportProjectBundle(ctx, s.store, id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	report := trust.Evaluate(bundle.Project, bundle.Evidence, bundle.Relationships, time.Now().UTC())
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleListProcurements(w http.ResponseWriter, r *http.Request) {
	limit, err := boundedInt(r.URL.Query().Get("limit"), 50, 1, 500)
	if err != nil { writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error()); return }
	offset, err := boundedInt(r.URL.Query().Get("offset"), 0, 0, 1_000_000)
	if err != nil { writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error()); return }
	procs, err := s.store.ListProcurements(r.Context(), limit, offset)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "procurements_unavailable", "Procurements are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"procurements": procs, "limit": limit, "offset": offset, "status": domain.StatusHealthy})
}

func (s *Server) handleListSignals(w http.ResponseWriter, r *http.Request) {
	sigs, err := s.store.ListSignals(r.Context(), 90*24*time.Hour, 50)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "signals_unavailable", "Signals are temporarily unavailable.")
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
		writeError(w, r, http.StatusServiceUnavailable, "search_unavailable", "Search is temporarily unavailable.")
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
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"methodology": sovereignty.VersionAISovereignty,
		"status":      domain.StatusUnavailable,
		"benchmarks":  []interface{}{},
		"reason":      "No reviewed evidence-backed provider scorecards are published in the current dataset.",
	})
}

func (s *Server) handleRankings(w http.ResponseWriter, r *http.Request) {
	dim := r.PathValue("dimension")
	if dim != "buildability" { writeError(w, r, http.StatusBadRequest, "unsupported_ranking", "Only buildability-v2.0 is currently published."); return }
	list, err := s.store.ListRankings(r.Context(), dim, 25)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "rankings_unavailable", "Rankings are temporarily unavailable.")
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
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}

	switch strings.ToLower(format) {
	case "markdown", "md":
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = w.Write([]byte(bundle.ToMarkdown()))
	case "cegs":
		cegsProj, err := bundle.ToCEGSExport(); if err != nil { writeError(w, r, 500, "export_failed", "CEGS export failed."); return }
		writeJSON(w, http.StatusOK, cegsProj)
	default:
		writeJSON(w, http.StatusOK, bundle)
	}
}

func (s *Server) handleCEGSProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	cegsProj, err := bundle.ToCEGSExport(); if err != nil { writeError(w, r, 500, "export_failed", "CEGS export failed."); return }
	writeJSON(w, http.StatusOK, cegsProj)
}

func (s *Server) handleCEGSExport(w http.ResponseWriter, r *http.Request) {
	projects, _, err := s.store.ListProjects(r.Context(), database.ProjectFilter{Limit: 500})
	if err != nil { writeError(w, r, 503, "export_unavailable", "CEGS export is temporarily unavailable."); return }
	var cegsList []*cegs.Project
	for _, p := range projects {
		cegsList = append(cegsList, cegs.ToCEGSProject(p, p.EvidenceIDs))
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
					"parameters": []map[string]any{
						{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer", "minimum": 1, "maximum": 500}},
						{"name": "offset", "in": "query", "schema": map[string]any{"type": "integer", "minimum": 0}},
					},
				},
			},
			"/api/v1/projects/{id}/trust": map[string]interface{}{"get": map[string]interface{}{"summary": "Fetch deterministic evidence-quality assessment"}},
			"/api/v1/projects/{id}/scores/history": map[string]interface{}{"get": map[string]interface{}{"summary": "Fetch append-only score history"}},
			"/api/v1/cegs/projects/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Fetch project conforming to CEGS 0.1 standard",
				},
			},
		},
	}
	writeJSON(w, http.StatusOK, spec)
}

func (s *Server) resolveProject(ctx context.Context, id string) (*domain.Project, error) {
	project, err := s.store.GetProject(ctx, id)
	if err == nil { return project, nil }
	return s.store.GetProjectBySlug(ctx, id)
}

func boundedInt(raw string, defaultValue, min, max int) (int, error) {
	if raw == "" { return defaultValue, nil }
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max { return 0, fmt.Errorf("value must be an integer between %d and %d", min, max) }
	return value, nil
}

func optionalNonNegativeInt64(raw string) (int64, error) {
	if raw == "" { return 0, nil }
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 { return 0, fmt.Errorf("value must be a non-negative integer") }
	return value, nil
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{"code": code, "message": message},
		"request_id": w.Header().Get("X-Request-ID"),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
