package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/adaptersandbox"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/capitalstack"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/export"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/forecast"
	graphqlhandler "github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/graphql"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/matching"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/publication"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/readiness"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/reconciliation"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/sovereignty"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/trust"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/verifier"
	"github.com/google/uuid"
	)

const (
	maxRequestIDLength     = 64
	maxRequestTargetBytes  = 8 << 10
	maxQueryStringBytes    = 4 << 10
	maxResourceIDLength    = 200
	maxRateLimitIdentities = 10_000
)

type Logger interface {
	Printf(format string, values ...any)
}

// Options controls transport-adjacent API protections. The zero value is not
// accepted by NewServerWithOptions; use DefaultOptions as a safe baseline.
type Options struct {
	AllowedOrigins      []string
	EnableHSTS          bool
	TrustedProxyCIDRs   []string
	MaxRequestBodyBytes int64
	RateLimitPerMinute  int
	RateLimitBurst      int
	RequestTimeout      time.Duration
	ReadinessTimeout    time.Duration
	Logger              Logger
	// AdapterRegistry exposes the community adapter sandbox via REST endpoints.
	// When nil, the sandbox endpoints respond with 503.
	AdapterRegistry *adaptersandbox.Registry
	// VerifierStore persists verifier attestations and is surfaced by
	// GET /api/v1/verifier/attestations. When nil, the endpoint responds 503.
	VerifierStore *verifier.AttestationStore
	// VerifierNetwork is used to compute quorum vote results for
	// GET /api/v1/verifier/milestones/{id}. When nil, the endpoint responds 503.
	VerifierNetwork *verifier.Network
}

func DefaultOptions() Options {
	return Options{
		AllowedOrigins:      []string{"*"},
		MaxRequestBodyBytes: 4 << 10,
		RateLimitPerMinute:  600,
		RateLimitBurst:      100,
		RequestTimeout:      15 * time.Second,
		ReadinessTimeout:    2 * time.Second,
	}
}

type Server struct {
	store            database.Store
	mux              *http.ServeMux
	allowedOrigins   map[string]struct{}
	allowAnyOrigin   bool
	enableHSTS       bool
	trustedProxies   []*net.IPNet
	maxRequestBody   int64
	requestTimeout   time.Duration
	readinessTimeout time.Duration
	limiter          *rateLimiter
	logger           Logger
	adapterRegistry  *adaptersandbox.Registry
	attStore         *verifier.AttestationStore
	verifierNet      *verifier.Network
}

func NewServer(store database.Store) *Server {
	server, err := NewServerWithOptions(store, DefaultOptions())
	if err != nil {
		panic(err)
	}
	return server
}

func NewServerWithOptions(store database.Store, options Options) (*Server, error) {
	if options.MaxRequestBodyBytes <= 0 {
		return nil, fmt.Errorf("max request body bytes must be positive")
	}
	if options.RequestTimeout <= 0 || options.ReadinessTimeout <= 0 {
		return nil, fmt.Errorf("request and readiness timeouts must be positive")
	}
	if options.RateLimitPerMinute < 0 || options.RateLimitBurst <= 0 {
		return nil, fmt.Errorf("rate limit must be non-negative and burst must be positive")
	}

	s := &Server{
		store:            store,
		mux:              http.NewServeMux(),
		allowedOrigins:   make(map[string]struct{}, len(options.AllowedOrigins)),
		enableHSTS:       options.EnableHSTS,
		maxRequestBody:   options.MaxRequestBodyBytes,
		requestTimeout:   options.RequestTimeout,
		readinessTimeout: options.ReadinessTimeout,
		logger:           options.Logger,
		adapterRegistry:  options.AdapterRegistry,
		attStore:         options.VerifierStore,
		verifierNet:      options.VerifierNetwork,
	}
	for _, origin := range options.AllowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			if len(options.AllowedOrigins) != 1 {
				return nil, fmt.Errorf("wildcard CORS origin cannot be combined with explicit origins")
			}
			s.allowAnyOrigin = true
			continue
		}
		var err error
		origin, err = normalizeAllowedOrigin(origin)
		if err != nil {
			return nil, err
		}
		s.allowedOrigins[origin] = struct{}{}
	}
	for _, rawCIDR := range options.TrustedProxyCIDRs {
		_, cidr, err := net.ParseCIDR(rawCIDR)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q", rawCIDR)
		}
		s.trustedProxies = append(s.trustedProxies, cidr)
	}
	if options.RateLimitPerMinute > 0 {
		s.limiter = newRateLimiter(options.RateLimitPerMinute, options.RateLimitBurst)
	}
	s.registerRoutes()
	return s, nil
}

func normalizeAllowedOrigin(origin string) (string, error) {
	if origin == "" || len(origin) > 2048 || strings.ContainsAny(origin, "\r\n\t") {
		return "", fmt.Errorf("invalid CORS origin")
	}
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", fmt.Errorf("CORS origin %q must be an absolute HTTP(S) origin without credentials, path, query, or fragment", origin)
	}
	return strings.TrimSuffix(origin, "/"), nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := normalizedRequestID(r.Header.Get("X-Request-ID"))
	ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
	r = r.WithContext(ctx)

	w.Header().Set("X-Request-ID", requestID)
	s.setSecurityHeaders(w.Header())
	originAllowed := s.setCORSHeaders(w.Header(), r.Header.Get("Origin"))

	defer func() {
		if recover() != nil {
			if s.logger != nil {
				s.logger.Printf("[ERROR] recovered API panic request_id=%s", requestID)
			}
			w.Header().Set("Connection", "close")
			writeError(w, r, http.StatusInternalServerError, "internal_error", "The request could not be completed.")
		}
	}()

	if !s.validateRequest(w, r) {
		return
	}
	if r.Method == http.MethodOptions {
		s.handlePreflight(w, r, originAllowed)
		return
	}
	if s.limiter != nil && (r.Method == http.MethodGet || r.Method == http.MethodHead) &&
		(strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/metrics") {
		allowed, remaining, retryAfter := s.limiter.Allow(s.clientIdentity(r), time.Now())
		w.Header().Set("RateLimit-Limit", strconv.Itoa(s.limiter.burst))
		w.Header().Set("RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("RateLimit-Policy", fmt.Sprintf("%d;w=60;burst=%d", s.limiter.perMinute, s.limiter.burst))
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			writeError(w, r, http.StatusTooManyRequests, "rate_limit_exceeded", "Request rate limit exceeded; retry later.")
			return
		}
	}

	requestContext, cancel := context.WithTimeout(r.Context(), s.requestTimeout)
	defer cancel()
	r = r.WithContext(requestContext)
	s.mux.ServeHTTP(w, r)
}

type requestIDContextKey struct{}

// RequestIDFromContext returns the validated request identifier assigned by
// the API boundary, allowing downstream logs to correlate without raw input.
func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func normalizedRequestID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxRequestIDLength {
		return uuid.NewString()
	}
	for _, char := range raw {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || strings.ContainsRune("-_.:", char) {
			continue
		}
		return uuid.NewString()
	}
	return raw
}

func (s *Server) setSecurityHeaders(header http.Header) {
	header.Set("Cache-Control", "no-store")
	header.Set("Content-Security-Policy", "default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	header.Set("Cross-Origin-Opener-Policy", "same-origin")
	header.Set("Cross-Origin-Resource-Policy", "cross-origin")
	header.Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(), payment=(), usb=()")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
	header.Set("X-Permitted-Cross-Domain-Policies", "none")
	header.Set("X-XSS-Protection", "0")
	if s.enableHSTS {
		header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
}

func (s *Server) setCORSHeaders(header http.Header, origin string) bool {
	if origin == "" {
		return true
	}
	if len(origin) > 2048 || strings.ContainsAny(origin, "\r\n\t") {
		return false
	}
	allowed := s.allowAnyOrigin
	if !allowed {
		_, allowed = s.allowedOrigins[origin]
		appendVary(header, "Origin")
	}
	if !allowed {
		return false
	}
	if s.allowAnyOrigin {
		header.Set("Access-Control-Allow-Origin", "*")
	} else {
		header.Set("Access-Control-Allow-Origin", origin)
	}
	header.Set("Access-Control-Expose-Headers", "X-Request-ID, RateLimit-Limit, RateLimit-Remaining, RateLimit-Policy")
	return true
}

func (s *Server) handlePreflight(w http.ResponseWriter, r *http.Request, originAllowed bool) {
	if r.Header.Get("Origin") != "" && !originAllowed {
		writeError(w, r, http.StatusForbidden, "origin_not_allowed", "The request origin is not allowed.")
		return
	}
	requestedMethod := strings.ToUpper(strings.TrimSpace(r.Header.Get("Access-Control-Request-Method")))
	if requestedMethod != "" && requestedMethod != http.MethodGet && requestedMethod != http.MethodHead {
		writeError(w, r, http.StatusForbidden, "method_not_allowed", "Only read-only cross-origin requests are allowed.")
		return
	}
	allowedHeaders := map[string]struct{}{
		"accept":       {},
		"content-type": {},
		"x-request-id": {},
	}
	for _, rawHeader := range strings.Split(r.Header.Get("Access-Control-Request-Headers"), ",") {
		name := strings.ToLower(strings.TrimSpace(rawHeader))
		if name == "" {
			continue
		}
		if _, ok := allowedHeaders[name]; !ok {
			writeError(w, r, http.StatusForbidden, "header_not_allowed", "A requested cross-origin header is not allowed.")
			return
		}
	}
	appendVary(w.Header(), "Access-Control-Request-Method")
	appendVary(w.Header(), "Access-Control-Request-Headers")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-Request-ID")
	w.Header().Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) validateRequest(w http.ResponseWriter, r *http.Request) bool {
	target := r.RequestURI
	if target == "" {
		target = r.URL.EscapedPath()
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
	}
	if len(target) > maxRequestTargetBytes {
		writeError(w, r, http.StatusRequestURITooLong, "request_target_too_long", "The request target is too long.")
		return false
	}
	if len(r.URL.RawQuery) > maxQueryStringBytes {
		writeError(w, r, http.StatusRequestURITooLong, "query_too_long", "The query string is too long.")
		return false
	}
	if r.ContentLength > s.maxRequestBody {
		writeError(w, r, http.StatusRequestEntityTooLarge, "request_body_too_large", "The request body is too large.")
		return false
	}
	if r.ContentLength != 0 || len(r.TransferEncoding) > 0 {
		writeError(w, r, http.StatusBadRequest, "request_body_not_allowed", "Request bodies are not accepted by this read-only API.")
		return false
	}
	return true
}

func appendVary(header http.Header, value string) {
	for _, existing := range header.Values("Vary") {
		for _, item := range strings.Split(existing, ",") {
			if strings.EqualFold(strings.TrimSpace(item), value) {
				return
			}
		}
	}
	header.Add("Vary", value)
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
	s.mux.HandleFunc("GET /api/v1/projects/{id}/capital-stack", s.handleProjectCapitalStack)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/opportunities", s.handleProjectOpportunities)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/readiness", s.handleProjectReadiness)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/corroboration", s.handleProjectCorroboration)

	// Procurements & Opportunities
	s.mux.HandleFunc("GET /api/v1/procurements", s.handleListProcurements)
	s.mux.HandleFunc("GET /api/v1/opportunities", s.handleListOpportunities)
	s.mux.HandleFunc("GET /api/v1/opportunities/{id}", s.handleGetOpportunity)
	s.mux.HandleFunc("GET /api/v1/capital-needs", s.handleListCapitalNeeds)
	s.mux.HandleFunc("GET /api/v1/milestones", s.handleListMilestones)
	s.mux.HandleFunc("GET /api/v1/signals", s.handleListSignals)
	s.mux.HandleFunc("GET /api/v1/search", s.handleSearch)

	// Capital Stack & AI Sovereignty
	s.mux.HandleFunc("GET /api/v1/capital/stack", s.handleCapitalStack)
	s.mux.HandleFunc("GET /api/v1/ai-sovereignty", s.handleAISovereignty)
	s.mux.HandleFunc("GET /api/v1/ai-sovereignty/{id}", s.handleAISovereigntyEntity)

	// Rankings
	s.mux.HandleFunc("GET /api/v1/rankings/{dimension}", s.handleRankings)

	// Forecasting & Scenario Planning
	s.mux.HandleFunc("GET /api/v1/forecast/projects/{id}", s.handleForecastProject)
	s.mux.HandleFunc("GET /api/v1/forecast/portfolio", s.handleForecastPortfolio)

	// Counterparty Matching & Deal Precedents
	s.mux.HandleFunc("GET /api/v1/projects/{id}/fit/{archetype}", s.handleProjectArchetypeFit)
	s.mux.HandleFunc("GET /api/v1/projects/{id}/precedents", s.handleProjectPrecedents)

	// Multi-Jurisdiction Reconciliation
	s.mux.HandleFunc("GET /api/v1/reconciliation", s.handleReconciliation)

	// Enterprise GraphQL API
	gqlHandler := graphqlhandler.NewHandler(s.store)
	s.mux.Handle("GET /api/v1/graphql", gqlHandler)
	s.mux.Handle("POST /api/v1/graphql", gqlHandler)

	// Decentralized Verifier Endpoints
	s.mux.HandleFunc("GET /api/v1/verifier/attestations", s.handleVerifierAttestations)
	s.mux.HandleFunc("GET /api/v1/verifier/milestones/{id}", s.handleVerifierMilestone)

	// Community Adapter Sandbox
	s.mux.HandleFunc("GET /api/v1/adapters", s.handleListAdapters)
	s.mux.HandleFunc("POST /api/v1/adapters/{name}/approve", s.handleApproveAdapter)
	s.mux.HandleFunc("DELETE /api/v1/adapters/{name}", s.handleDeleteAdapter)

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
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), s.readinessTimeout)
	defer cancel()
	stats, err := s.store.GetRadarStats(ctx)
	if err != nil || stats == nil || stats.TotalProjects == 0 || stats.DataStatus == domain.StatusUnavailable {
		writeError(w, r, http.StatusServiceUnavailable, "not_ready", "The API data store is not ready.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats, err := s.store.GetRadarStats(r.Context())
	if err != nil || stats == nil {
		writeError(w, r, http.StatusServiceUnavailable, "metrics_unavailable", "Metrics are temporarily unavailable.")
		return
	}
	entities, _ := s.store.ListEntities(r.Context())
	signals, _ := s.store.ListSignals(r.Context(), 30*24*time.Hour, 10000)
	procurements, _ := s.store.ListProcurements(r.Context(), 0, 10000)
	capitalItems, _ := s.store.ListAllCapitalItems(r.Context())

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP cog_total_projects Tracked major projects count\n")
	fmt.Fprintf(w, "# TYPE cog_total_projects gauge\n")
	fmt.Fprintf(w, "cog_total_projects %d\n", stats.TotalProjects)
	fmt.Fprintf(w, "# HELP cog_total_capex_cad Tracked total CAPEX in CAD\n")
	fmt.Fprintf(w, "# TYPE cog_total_capex_cad gauge\n")
	fmt.Fprintf(w, "cog_total_capex_cad %d\n", stats.TotalCapexCAD)
	fmt.Fprintf(w, "# HELP cog_total_entities Tracked organizations count\n")
	fmt.Fprintf(w, "# TYPE cog_total_entities gauge\n")
	fmt.Fprintf(w, "cog_total_entities %d\n", len(entities))
	fmt.Fprintf(w, "# HELP cog_total_signals Economic momentum signals count\n")
	fmt.Fprintf(w, "# TYPE cog_total_signals gauge\n")
	fmt.Fprintf(w, "cog_total_signals %d\n", len(signals))
	fmt.Fprintf(w, "# HELP cog_total_procurements Tracked procurement records count\n")
	fmt.Fprintf(w, "# TYPE cog_total_procurements gauge\n")
	fmt.Fprintf(w, "cog_total_procurements %d\n", len(procurements))
	fmt.Fprintf(w, "# HELP cog_total_capital_items Capital financing items count\n")
	fmt.Fprintf(w, "# TYPE cog_total_capital_items gauge\n")
	fmt.Fprintf(w, "cog_total_capital_items %d\n", len(capitalItems))
	var attCount int
	if s.attStore != nil {
		attCount = s.attStore.Count()
	}
	fmt.Fprintf(w, "# HELP cog_verifier_attestations Total verifier attestations recorded\n")
	fmt.Fprintf(w, "# TYPE cog_verifier_attestations counter\n")
	fmt.Fprintf(w, "cog_verifier_attestations %d\n", attCount)
}

func (s *Server) handleRadar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := s.store.GetRadarStats(ctx)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "radar_unavailable", "Radar aggregates are temporarily unavailable.")
		return
	}

	accelerating, err := s.store.ListRankings(ctx, "buildability", 5)
	if err != nil {
		accelerating = nil
	}
	signalList, err := s.publicSignals(ctx, 30*24*time.Hour, 10)
	if err != nil {
		signalList = nil
	}
	recentEvents, err := s.store.ListRecentEvents(ctx, 25)
	if err != nil {
		recentEvents = nil
	}
	filteredEvents := make([]*domain.Event, 0, 5)
	for _, event := range recentEvents {
		if s.publicEvidence(ctx, event.EvidenceID) { filteredEvents = append(filteredEvents, event) }
		if len(filteredEvents) == 5 { break }
	}

	resp := map[string]interface{}{
		"stats":                 stats,
		"accelerating_projects": accelerating,
		"recent_signals":        signalList,
		"recent_events":         filteredEvents,
		"cegs_version":          cegs.SpecVersion,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := boundedInt(q.Get("limit"), 50, 1, 500)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}
	offset, err := boundedInt(q.Get("offset"), 0, 0, 1_000_000)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error())
		return
	}
	minCapex, err := optionalNonNegativeInt64(q.Get("min_capex"))
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_min_capex", err.Error())
		return
	}
	sector, err := boundedText(q.Get("sector"), "sector", 80)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_sector", err.Error())
		return
	}
	province, err := boundedText(q.Get("province"), "province", 32)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_province", err.Error())
		return
	}
	stage, err := boundedText(q.Get("stage"), "stage", 64)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_stage", err.Error())
		return
	}
	search, err := boundedText(q.Get("q"), "q", 200)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	sortBy, err := boundedText(q.Get("sort_by"), "sort_by", 32)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_sort", err.Error())
		return
	}
	if !oneOf(sortBy, "", "updated", "capex", "name", "buildability", "investability", "supplierability", "strategicity", "trade_resilience") {
		writeError(w, r, http.StatusBadRequest, "invalid_sort", "sort_by is not supported")
		return
	}
	if q.Get("sort_dir") != "" && q.Get("sort_dir") != "asc" && q.Get("sort_dir") != "desc" {
		writeError(w, r, http.StatusBadRequest, "invalid_sort_dir", "sort_dir must be asc or desc")
		return
	}

	filter := database.ProjectFilter{
		Sector:      sector,
		Province:    province,
		Stage:       stage,
		MinCapexCAD: minCapex,
		Search:      search,
		SortBy:      sortBy,
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
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}

	bundle, err := export.ExportProjectBundle(ctx, s.store, proj.ID)
	if err != nil {
		writeError(w, r, 503, "project_bundle_unavailable", "Project intelligence is temporarily unavailable.")
		return
	}

	resp := map[string]interface{}{
		"project":       bundle.Project,
		"scores":        bundle.Scores,
		"events":        bundle.Events,
		"relationships": bundle.Relationships,
		"capital_items": bundle.CapitalItems,
		"capital_needs": bundle.CapitalNeeds,
		"capital_requirements": bundle.CapitalRequirements,
		"milestones":    bundle.Milestones,
		"readiness":     bundle.Readiness,
		"opportunities": bundle.Opportunities,
		"status":        domain.StatusHealthy,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetProjectEvents(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "events_unavailable", "Events are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, bundle.Events)
}

func (s *Server) handleGetProjectScores(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "scores_unavailable", "Scores are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, bundle.Scores)
}

func (s *Server) handleGetProjectScoreHistory(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	scoreType, err := boundedText(r.URL.Query().Get("type"), "type", 32)
	if err != nil || !oneOf(scoreType, "", "buildability", "investability", "supplierability", "strategicity", "trade_resilience") {
		writeError(w, r, http.StatusBadRequest, "invalid_score_type", "Score type is not supported.")
		return
	}
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "score_history_unavailable", "Score history is temporarily unavailable.")
		return
	}
	history := make([]*domain.ProjectScore, 0)
	for _, score := range bundle.ScoreHistory { if scoreType == "" || score.ScoreType == scoreType { history = append(history, score) } }
	writeJSON(w, http.StatusOK, map[string]any{"project_id": project.ID, "history": history})
}

func (s *Server) handleGetProjectProvenance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := validateResourceID(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_project_id", "Project identifier is invalid.")
		return
	}
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
	if err := validateResourceID(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_project_id", "Project identifier is invalid.")
		return
	}
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
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}
	offset, err := boundedInt(r.URL.Query().Get("offset"), 0, 0, 1_000_000)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error())
		return
	}
	procs, err := s.store.ListProcurements(r.Context(), 500, 0)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "procurements_unavailable", "Procurements are temporarily unavailable.")
		return
	}
	public := make([]*domain.Procurement, 0, len(procs))
	for _, procurement := range procs { if s.publicEvidence(r.Context(), procurement.EvidenceID) { public = append(public, procurement) } }
	start := offset; if start > len(public) { start = len(public) }; end := start + limit; if end > len(public) { end = len(public) }
	writeJSON(w, http.StatusOK, map[string]any{"procurements": public[start:end], "limit": limit, "offset": offset, "status": domain.StatusHealthy})
}

func (s *Server) handleListSignals(w http.ResponseWriter, r *http.Request) {
	sigs, err := s.publicSignals(r.Context(), 90*24*time.Hour, 50)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "signals_unavailable", "Signals are temporarily unavailable.")
		return
	}
	writeJSON(w, http.StatusOK, sigs)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query, err := boundedText(r.URL.Query().Get("q"), "q", 200)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
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
	profiles := s.collectAISovereigntyProfiles()
	if len(profiles) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"methodology": sovereignty.VersionAISovereignty,
			"status":      domain.StatusUnavailable,
			"benchmarks":  []interface{}{},
			"reason":      "No reviewed evidence-backed provider scorecards are published in the current dataset.",
		})
		return
	}
	benchmarks := make([]*domain.AISovereignty, 0, len(profiles))
	for _, profile := range profiles {
		benchmarks = append(benchmarks, sovereignty.EvaluateSovereignty(profile))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"methodology": sovereignty.VersionAISovereignty,
		"status":      domain.StatusHealthy,
		"benchmarks":  benchmarks,
	})
}

func (s *Server) handleRankings(w http.ResponseWriter, r *http.Request) {
	dim := r.PathValue("dimension")
	if dim != "buildability" {
		writeError(w, r, http.StatusBadRequest, "unsupported_ranking", "Only buildability-v2.0 is currently published.")
		return
	}
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
	if err := validateResourceID(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_project_id", "Project identifier is invalid.")
		return
	}
	format, err := boundedText(r.URL.Query().Get("format"), "format", 16)
	if err != nil || !oneOf(strings.ToLower(format), "", "json", "markdown", "md", "cegs") {
		writeError(w, r, http.StatusBadRequest, "invalid_format", "Export format is not supported.")
		return
	}
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
		cegsProj, err := bundle.ToCEGSExport()
		if err != nil {
			writeError(w, r, 500, "export_failed", "CEGS export failed.")
			return
		}
		writeJSON(w, http.StatusOK, cegsProj)
	default:
		writeJSON(w, http.StatusOK, bundle)
	}
}

func (s *Server) handleCEGSProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := validateResourceID(id); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_project_id", "Project identifier is invalid.")
		return
	}
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	cegsProj, err := bundle.ToCEGSExport()
	if err != nil {
		writeError(w, r, 500, "export_failed", "CEGS export failed.")
		return
	}
	writeJSON(w, http.StatusOK, cegsProj)
}

func (s *Server) handleCEGSExport(w http.ResponseWriter, r *http.Request) {
	projects, _, err := s.store.ListProjects(r.Context(), database.ProjectFilter{Limit: 500})
	if err != nil {
		writeError(w, r, 503, "export_unavailable", "CEGS export is temporarily unavailable.")
		return
	}
	var cegsList []*cegs.Project
	for _, p := range projects {
		bundle, bundleErr := export.ExportProjectBundle(r.Context(), s.store, p.ID)
		if bundleErr != nil { continue }
		cegsProject, convertErr := bundle.ToCEGSExport()
		if convertErr == nil { cegsList = append(cegsList, cegsProject) }
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"cegs":         cegs.SpecVersion,
		"manifest_id":  "cegs:manifest:ca:live-export",
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
			"/api/v1/projects/{id}/trust":          map[string]interface{}{"get": map[string]interface{}{"summary": "Fetch deterministic evidence-quality assessment"}},
			"/api/v1/opportunities": map[string]interface{}{"get": map[string]interface{}{"summary": "Query publication-safe investment and counterparty opportunities"}},
			"/api/v1/capital-needs": map[string]interface{}{"get": map[string]interface{}{"summary": "Query publicly corroborated capital needs"}},
			"/api/v1/milestones": map[string]interface{}{"get": map[string]interface{}{"summary": "Query sourced project milestones"}},
			"/api/v1/projects/{id}/readiness": map[string]interface{}{"get": map[string]interface{}{"summary": "Fetch deterministic investment-readiness decomposition"}},
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

func (s *Server) publicEvidence(ctx context.Context, evidenceID string) bool {
	if evidenceID == "" {
		return false
	}
	e, err := s.store.GetEvidence(ctx, evidenceID)
	if err != nil || e == nil {
		return false
	}
	return publication.PublicEvidence(e)
}

func (s *Server) publicSignals(ctx context.Context, since time.Duration, limit int) ([]*domain.Signal, error) {
	sigs, err := s.store.ListSignals(ctx, since, limit)
	if err != nil {
		return nil, err
	}
	filtered := make([]*domain.Signal, 0, len(sigs))
	for _, sig := range sigs {
		if sig.EvidenceID != "" && !s.publicEvidence(ctx, sig.EvidenceID) {
			continue
		}
		filtered = append(filtered, sig)
	}
	return filtered, nil
}

func (s *Server) handleProjectCapitalStack(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	items, err := s.store.ListCapitalItemsByProject(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "capital_stack_unavailable", "Capital stack is temporarily unavailable.")
		return
	}
	public := make([]*domain.CapitalItem, 0, len(items))
	for _, item := range items {
		if publication.PublicCapitalItem(item) && s.publicEvidence(r.Context(), item.EvidenceID) {
			public = append(public, item)
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":   project.ID,
		"capital_items": public,
		"status":       domain.StatusHealthy,
	})
}

func (s *Server) handleProjectOpportunities(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	opps, err := s.store.ListOpportunitiesByProject(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "opportunities_unavailable", "Opportunities are temporarily unavailable.")
		return
	}
	public := make([]*domain.Opportunity, 0, len(opps))
	for _, opp := range opps {
		if !publication.PublicOpportunity(opp) {
			continue
		}
		valid := true
		for _, eid := range opp.EvidenceIDs {
			if !s.publicEvidence(r.Context(), eid) {
				valid = false
				break
			}
		}
		if valid {
			public = append(public, opp)
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":    project.ID,
		"opportunities": public,
		"status":        domain.StatusHealthy,
	})
}

func (s *Server) handleProjectReadiness(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	milestones, err := s.store.ListMilestones(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "readiness_unavailable", "Readiness data is temporarily unavailable.")
		return
	}
	capitalNeeds, err := s.store.ListCapitalNeeds(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "readiness_unavailable", "Readiness data is temporarily unavailable.")
		return
	}
	capitalItems, err := s.store.ListCapitalItemsByProject(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "readiness_unavailable", "Readiness data is temporarily unavailable.")
		return
	}
	score := readiness.CalculateCapitalReadiness(readiness.CapitalReadinessInputs{
		Project:      project,
		Milestones:   milestones,
		CapitalNeeds: capitalNeeds,
		CapitalItems: capitalItems,
	}, time.Now().UTC())
	writeJSON(w, http.StatusOK, score)
}

func (s *Server) handleProjectCorroboration(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	claims, err := s.store.ListClaimsBySubject(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "corroboration_unavailable", "Corroboration data is temporarily unavailable.")
		return
	}
	audits, err := s.store.ListAuditEntries(r.Context(), project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "corroboration_unavailable", "Corroboration data is temporarily unavailable.")
		return
	}
	publicClaims := make([]*domain.Claim, 0, len(claims))
	for _, c := range claims {
		if publication.PublicClaim(c) && s.publicEvidence(r.Context(), c.EvidenceID) {
			publicClaims = append(publicClaims, c)
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":   project.ID,
		"claims":       publicClaims,
		"audit_entries": audits,
		"status":       domain.StatusHealthy,
	})
}

func (s *Server) handleListOpportunities(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := boundedInt(q.Get("limit"), 50, 1, 500)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}
	offset, err := boundedInt(q.Get("offset"), 0, 0, 1_000_000)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error())
		return
	}
	opps, err := s.store.ListAllOpportunities(r.Context())
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "opportunities_unavailable", "Opportunities are temporarily unavailable.")
		return
	}
	public := make([]*domain.Opportunity, 0, len(opps))
	for _, opp := range opps {
		if !publication.PublicOpportunity(opp) {
			continue
		}
		valid := true
		for _, eid := range opp.EvidenceIDs {
			if !s.publicEvidence(r.Context(), eid) {
				valid = false
				break
			}
		}
		if valid {
			public = append(public, opp)
		}
	}
	start := offset
	if start > len(public) {
		start = len(public)
	}
	end := start + limit
	if end > len(public) {
		end = len(public)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"opportunities": public[start:end],
		"total":         len(public),
		"limit":         limit,
		"offset":        offset,
		"status":        domain.StatusHealthy,
	})
}

func (s *Server) handleGetOpportunity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()
	opp, err := s.store.GetOpportunity(ctx, id)
	if err != nil || opp == nil {
		writeError(w, r, http.StatusNotFound, "opportunity_not_found", "Opportunity not found.")
		return
	}
	if !publication.PublicOpportunity(opp) {
		writeError(w, r, http.StatusNotFound, "opportunity_not_found", "Opportunity not found.")
		return
	}
	valid := true
	for _, eid := range opp.EvidenceIDs {
		if !s.publicEvidence(ctx, eid) {
			valid = false
			break
		}
	}
	if !valid {
		writeError(w, r, http.StatusNotFound, "opportunity_not_found", "Opportunity not found.")
		return
	}
	writeJSON(w, http.StatusOK, opp)
}

func (s *Server) handleListCapitalNeeds(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := boundedInt(q.Get("limit"), 50, 1, 500)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}
	offset, err := boundedInt(q.Get("offset"), 0, 0, 1_000_000)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error())
		return
	}
	projects, _, err := s.store.ListProjects(r.Context(), database.ProjectFilter{Limit: 500})
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "capital_needs_unavailable", "Capital needs are temporarily unavailable.")
		return
	}
	public := make([]*domain.CapitalNeed, 0)
	for _, p := range projects {
		needs, err := s.store.ListCapitalNeeds(r.Context(), p.ID)
		if err != nil {
			continue
		}
		for _, need := range needs {
			if !publication.PublicCapitalNeed(need) {
				continue
			}
			valid := len(need.EvidenceIDs) > 0
			for _, eid := range need.EvidenceIDs {
				if !s.publicEvidence(r.Context(), eid) {
					valid = false
					break
				}
			}
			if valid {
				public = append(public, need)
			}
		}
	}
	start := offset
	if start > len(public) {
		start = len(public)
	}
	end := start + limit
	if end > len(public) {
		end = len(public)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"capital_needs": public[start:end],
		"total":         len(public),
		"limit":         limit,
		"offset":        offset,
		"status":        domain.StatusHealthy,
	})
}

func (s *Server) handleListMilestones(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, err := boundedInt(q.Get("limit"), 50, 1, 500)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_limit", err.Error())
		return
	}
	offset, err := boundedInt(q.Get("offset"), 0, 0, 1_000_000)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_offset", err.Error())
		return
	}
	projects, _, err := s.store.ListProjects(r.Context(), database.ProjectFilter{Limit: 500})
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "milestones_unavailable", "Milestones are temporarily unavailable.")
		return
	}
	public := make([]*domain.Milestone, 0)
	for _, p := range projects {
		milestones, err := s.store.ListMilestones(r.Context(), p.ID)
		if err != nil {
			continue
		}
		for _, m := range milestones {
			if !publication.PublicMilestone(m) {
				continue
			}
			valid := len(m.EvidenceIDs) > 0
			for _, eid := range m.EvidenceIDs {
				if !s.publicEvidence(r.Context(), eid) {
					valid = false
					break
				}
			}
			if valid {
				public = append(public, m)
			}
		}
	}
	start := offset
	if start > len(public) {
		start = len(public)
	}
	end := start + limit
	if end > len(public) {
		end = len(public)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"milestones": public[start:end],
		"total":      len(public),
		"limit":      limit,
		"offset":     offset,
		"status":     domain.StatusHealthy,
	})
}

func (s *Server) resolveProject(ctx context.Context, id string) (*domain.Project, error) {
	if err := validateResourceID(id); err != nil {
		return nil, err
	}
	project, err := s.store.GetProject(ctx, id)
	if err == nil {
		return project, nil
	}
	return s.store.GetProjectBySlug(ctx, id)
}

type rateLimitBucket struct {
	tokens   float64
	lastSeen time.Time
}

type rateLimiter struct {
	mu          sync.Mutex
	perMinute   int
	perSecond   float64
	burst       int
	buckets     map[string]*rateLimitBucket
	lastCleanup time.Time
}

func newRateLimiter(perMinute, burst int) *rateLimiter {
	return &rateLimiter{
		perMinute: perMinute,
		perSecond: float64(perMinute) / 60,
		burst:     burst,
		buckets:   make(map[string]*rateLimitBucket),
	}
}

func (l *rateLimiter) Allow(identity string, now time.Time) (allowed bool, remaining, retryAfter int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= time.Minute {
		for key, bucket := range l.buckets {
			if now.Sub(bucket.lastSeen) > 10*time.Minute {
				delete(l.buckets, key)
			}
		}
		l.lastCleanup = now
	}

	if _, exists := l.buckets[identity]; !exists && len(l.buckets) >= maxRateLimitIdentities {
		identity = "rate-limit-overflow"
	}
	bucket, exists := l.buckets[identity]
	if !exists {
		bucket = &rateLimitBucket{tokens: float64(l.burst), lastSeen: now}
		l.buckets[identity] = bucket
	}
	elapsed := now.Sub(bucket.lastSeen).Seconds()
	if elapsed > 0 {
		bucket.tokens = math.Min(float64(l.burst), bucket.tokens+elapsed*l.perSecond)
	}
	bucket.lastSeen = now
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true, int(math.Floor(bucket.tokens)), 0
	}
	waitSeconds := int(math.Ceil((1 - bucket.tokens) / l.perSecond))
	if waitSeconds < 1 {
		waitSeconds = 1
	}
	return false, 0, waitSeconds
}

func (s *Server) clientIdentity(r *http.Request) string {
	remoteIP := remoteIP(r.RemoteAddr)
	if remoteIP == nil {
		return "unknown"
	}
	if len(s.trustedProxies) == 0 || !ipInNetworks(remoteIP, s.trustedProxies) {
		return remoteIP.String()
	}

	forwarded := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	chain := make([]net.IP, 0, len(forwarded))
	for _, rawIP := range forwarded {
		ip := net.ParseIP(strings.TrimSpace(rawIP))
		if ip == nil {
			// A malformed chain is not trustworthy; fall back to the connected
			// peer instead of accepting a caller-controlled identity.
			return remoteIP.String()
		}
		chain = append(chain, ip)
	}
	for i := len(chain) - 1; i >= 0; i-- {
		if !ipInNetworks(chain[i], s.trustedProxies) {
			return chain[i].String()
		}
	}
	if len(chain) > 0 {
		return chain[0].String()
	}
	return remoteIP.String()
}

func remoteIP(remoteAddress string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = strings.Trim(remoteAddress, "[]")
	}
	return net.ParseIP(host)
}

func ipInNetworks(ip net.IP, networks []*net.IPNet) bool {
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func boundedText(raw, name string, maxRunes int) (string, error) {
	value := strings.TrimSpace(raw)
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxRunes {
		return "", fmt.Errorf("%s must be at most %d characters", name, maxRunes)
	}
	for _, char := range value {
		if unicode.IsControl(char) {
			return "", fmt.Errorf("%s contains an invalid control character", name)
		}
	}
	return value, nil
}

func validateResourceID(id string) error {
	if id == "" || len(id) > maxResourceIDLength {
		return fmt.Errorf("resource identifier must be between 1 and %d characters", maxResourceIDLength)
	}
	for _, char := range id {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || strings.ContainsRune("-_.:", char) {
			continue
		}
		return fmt.Errorf("resource identifier contains invalid characters")
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func boundedInt(raw string, defaultValue, min, max int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("value must be an integer between %d and %d", min, max)
	}
	return value, nil
}

func optionalNonNegativeInt64(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("value must be a non-negative integer")
	}
	return value, nil
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error":      map[string]string{"code": code, "message": message},
		"request_id": w.Header().Get("X-Request-ID"),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// --- Forecasting, sovereignty, matching, and reconciliation handlers ---

func (s *Server) handleForecastProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	project, err := s.resolveProject(r.Context(), id)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "forecast_unavailable", "Forecast is temporarily unavailable.")
		return
	}
	ctx := forecast.Context{
		Project:        bundle.Project,
		Events:         bundle.Events,
		CapitalItems:   bundle.CapitalItems,
		Relationships:  bundle.Relationships,
		Procurements:   bundle.Procurements,
		Opportunities:  bundle.Opportunities,
	}
	req := forecast.Request{AsOf: time.Now().UTC()}
	report, err := forecast.Evaluate(ctx, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_forecast", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleForecastPortfolio(w http.ResponseWriter, r *http.Request) {
	projects, _, err := s.store.ListProjects(r.Context(), database.ProjectFilter{Limit: 100})
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "portfolio_unavailable", "Portfolio forecast is temporarily unavailable.")
		return
	}
	contexts := make([]forecast.Context, 0, len(projects))
	for _, p := range projects {
		bundle, bundleErr := export.ExportProjectBundle(r.Context(), s.store, p.ID)
		if bundleErr != nil {
			continue
		}
		contexts = append(contexts, forecast.Context{
			Project:        bundle.Project,
			Events:         bundle.Events,
			CapitalItems:   bundle.CapitalItems,
			Relationships:  bundle.Relationships,
			Procurements:   bundle.Procurements,
			Opportunities:  bundle.Opportunities,
		})
	}
	if len(contexts) == 0 {
		writeError(w, r, http.StatusServiceUnavailable, "portfolio_unavailable", "No projects available for portfolio forecast.")
		return
	}
	req := forecast.Request{AsOf: time.Now().UTC()}
	report, err := forecast.EvaluatePortfolio(contexts, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid_portfolio", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleAISovereigntyEntity(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	entity, err := s.store.GetEntity(r.Context(), id)
	if err != nil || entity == nil || entity.AISovereignty == nil {
		writeError(w, r, http.StatusNotFound, "ai_sovereignty_not_found", "No AI sovereignty scorecard is published for this entity.")
		return
	}
	writeJSON(w, http.StatusOK, entity.AISovereignty)
}

func (s *Server) handleProjectArchetypeFit(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	bundle, err := export.ExportProjectBundle(r.Context(), s.store, project.ID)
	if err != nil {
		writeError(w, r, http.StatusServiceUnavailable, "fit_unavailable", "Fit scoring is temporarily unavailable.")
		return
	}
	archetype := domain.CounterpartyType(strings.TrimSpace(r.PathValue("archetype")))
	if archetype == "" {
		writeError(w, r, http.StatusBadRequest, "invalid_archetype", "Archetype is required.")
		return
	}
	ctx := matching.FitContext{
		Project:      bundle.Project,
		CapitalNeeds: bundle.CapitalNeeds,
		Milestones:   bundle.Milestones,
	}
	writeJSON(w, http.StatusOK, matching.CalculateArchetypeFit(ctx, archetype, time.Now().UTC()))
}

func (s *Server) handleProjectPrecedents(w http.ResponseWriter, r *http.Request) {
	project, err := s.resolveProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, http.StatusNotFound, "project_not_found", "Project not found.")
		return
	}
	profiles := s.collectInvestorProfiles()
	matches := matching.FindDealPrecedents(project, profiles, 10)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"project_id":  project.ID,
		"precedents":  matches,
		"status":      domain.StatusHealthy,
	})
}

func (s *Server) handleReconciliation(w http.ResponseWriter, r *http.Request) {
	report := reconciliation.Reconcile(r.Context(), s.store)
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) collectInvestorProfiles() []*domain.InvestorProfile {
	// Investor profiles are derived from entity metadata; the store does not
	// yet expose a separate profile index, so we surface an empty list
	// rather than hallucinating matches.
	_, _ = s.store.ListEntities(context.Background())
	return nil
}

func (s *Server) collectAISovereigntyProfiles() []*sovereignty.AIProfile {
	entities, err := s.store.ListEntities(context.Background())
	if err != nil {
		return nil
	}
	profiles := make([]*sovereignty.AIProfile, 0, len(entities))
	for _, entity := range entities {
		if entity == nil || entity.AISovereignty == nil {
			continue
		}
		profiles = append(profiles, &sovereignty.AIProfile{
			SubjectID:          entity.ID,
			SubjectName:        entity.LegalName,
			DataResidencyCA:    entity.AISovereignty.DataResidency >= 8,
			ComputeResidencyCA: entity.AISovereignty.ComputeResidency >= 8,
			CanadianOwnership:  entity.AISovereignty.CanadianOwnership / 10.0,
			ForeignLegalRisk:   entity.AISovereignty.ForeignLegalRisk / 10.0,
			LocalDeployment:    entity.AISovereignty.LocalDeployment >= 8,
			OfflineCapability:  entity.AISovereignty.LocalDeployment >= 9,
			BilingualCapacity:  entity.AISovereignty.BilingualCapacity / 10.0,
			QuebecLaw25Ready:   entity.AISovereignty.QuebecLaw25 >= 8,
			CleanEnergySource:  entity.AISovereignty.CleanEnergy / 10.0,
		})
	}
	return profiles
}
