package api

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/google/uuid"
)

func TestSecurityHeadersAndRequestIDs(t *testing.T) {
	options := testOptions()
	options.EnableHSTS = true
	server := mustServer(t, database.NewMemoryStore(), options)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	request.Header.Set("X-Request-ID", "caller id with spaces")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	requestID := response.Header().Get("X-Request-ID")
	if requestID == "caller id with spaces" {
		t.Fatal("unsafe caller request ID was reflected")
	}
	if _, err := uuid.Parse(requestID); err != nil {
		t.Fatalf("generated request ID %q is not a UUID: %v", requestID, err)
	}

	expectedHeaders := map[string]string{
		"Cache-Control":                     "no-store",
		"Content-Security-Policy":           "default-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'",
		"Cross-Origin-Opener-Policy":        "same-origin",
		"Cross-Origin-Resource-Policy":      "cross-origin",
		"Permissions-Policy":                "camera=(), geolocation=(), microphone=(), payment=(), usb=()",
		"Referrer-Policy":                   "no-referrer",
		"Strict-Transport-Security":         "max-age=31536000; includeSubDomains",
		"X-Content-Type-Options":            "nosniff",
		"X-Frame-Options":                   "DENY",
		"X-Permitted-Cross-Domain-Policies": "none",
		"X-XSS-Protection":                  "0",
	}
	for name, expected := range expectedHeaders {
		if actual := response.Header().Get(name); actual != expected {
			t.Errorf("%s = %q, want %q", name, actual, expected)
		}
	}

	validRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	validRequest.Header.Set("X-Request-ID", "agency-trace_2026.09")
	validResponse := httptest.NewRecorder()
	server.ServeHTTP(validResponse, validRequest)
	if actual := validResponse.Header().Get("X-Request-ID"); actual != "agency-trace_2026.09" {
		t.Fatalf("valid request ID = %q", actual)
	}
}

func TestCORSAllowlistAndReadOnlyPreflight(t *testing.T) {
	options := testOptions()
	options.AllowedOrigins = []string{"https://planner.gc.ca", "https://investor.example"}
	server := mustServer(t, database.NewMemoryStore(), options)

	t.Run("allowed origin", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		request.Header.Set("Origin", "https://planner.gc.ca")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if actual := response.Header().Get("Access-Control-Allow-Origin"); actual != "https://planner.gc.ca" {
			t.Fatalf("allow origin = %q", actual)
		}
		if !strings.Contains(response.Header().Get("Vary"), "Origin") {
			t.Fatalf("Vary = %q, want Origin", response.Header().Get("Vary"))
		}
	})

	t.Run("unlisted origin receives no grant", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/health", nil)
		request.Header.Set("Origin", "https://attacker.example")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want normal server-to-server response", response.Code)
		}
		if actual := response.Header().Get("Access-Control-Allow-Origin"); actual != "" {
			t.Fatalf("unexpected CORS grant %q", actual)
		}
	})

	t.Run("allowed read preflight", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
		request.Header.Set("Origin", "https://planner.gc.ca")
		request.Header.Set("Access-Control-Request-Method", http.MethodGet)
		request.Header.Set("Access-Control-Request-Headers", "Accept, X-Request-ID")
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
		if actual := response.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(actual, http.MethodGet) {
			t.Fatalf("allow methods = %q", actual)
		}
	})

	for name, test := range map[string][2]string{
		"unknown origin":  {"https://attacker.example", http.MethodGet},
		"mutation method": {"https://planner.gc.ca", http.MethodPost},
	} {
		t.Run(name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
			request.Header.Set("Origin", test[0])
			request.Header.Set("Access-Control-Request-Method", test[1])
			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
			}
		})
	}
}

func TestRequestAndQueryBounds(t *testing.T) {
	options := testOptions()
	options.MaxRequestBodyBytes = 8
	server := mustServer(t, database.NewMemoryStore(), options)

	tests := []struct {
		name       string
		request    *http.Request
		wantStatus int
		wantCode   string
	}{
		{
			name:       "query string ceiling",
			request:    httptest.NewRequest(http.MethodGet, "/api/v1/search?q="+strings.Repeat("a", maxQueryStringBytes+1), nil),
			wantStatus: http.StatusRequestURITooLong,
			wantCode:   "query_too_long",
		},
		{
			name:       "bounded search value",
			request:    httptest.NewRequest(http.MethodGet, "/api/v1/search?q="+strings.Repeat("a", 201), nil),
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_query",
		},
		{
			name:       "sort allowlist",
			request:    httptest.NewRequest(http.MethodGet, "/api/v1/projects?sort_by=arbitrary", nil),
			wantStatus: http.StatusBadRequest,
			wantCode:   "invalid_sort",
		},
		{
			name:       "body forbidden",
			request:    httptest.NewRequest(http.MethodGet, "/api/v1/projects", strings.NewReader("x")),
			wantStatus: http.StatusBadRequest,
			wantCode:   "request_body_not_allowed",
		},
	}

	tooLarge := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	tooLarge.ContentLength = 9
	tests = append(tests, struct {
		name       string
		request    *http.Request
		wantStatus int
		wantCode   string
	}{"declared body too large", tooLarge, http.StatusRequestEntityTooLarge, "request_body_too_large"})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			server.ServeHTTP(response, test.request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), `"code":"`+test.wantCode+`"`) {
				t.Fatalf("body = %s, want error code %q", response.Body.String(), test.wantCode)
			}
			if response.Header().Get("X-Request-ID") == "" {
				t.Fatal("bounded request response omitted request ID")
			}
		})
	}
}

func TestRateLimitUsesConnectedPeerAndExemptsHealth(t *testing.T) {
	options := testOptions()
	options.RateLimitPerMinute = 60
	options.RateLimitBurst = 2
	server := mustServer(t, database.NewMemoryStore(), options)

	for attempt, wantStatus := range []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests} {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		request.RemoteAddr = "192.0.2.10:4000"
		request.Header.Set("X-Forwarded-For", "198.51.100."+strconv.Itoa(attempt+1))
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != wantStatus {
			t.Fatalf("attempt %d status = %d, want %d", attempt+1, response.Code, wantStatus)
		}
		if response.Header().Get("RateLimit-Limit") != "2" {
			t.Fatalf("attempt %d missing rate limit headers", attempt+1)
		}
		if wantStatus == http.StatusTooManyRequests && response.Header().Get("Retry-After") == "" {
			t.Fatal("rate-limited response omitted Retry-After")
		}
	}

	healthRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRequest.RemoteAddr = "192.0.2.10:4000"
	healthResponse := httptest.NewRecorder()
	server.ServeHTTP(healthResponse, healthRequest)
	if healthResponse.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", healthResponse.Code, http.StatusOK)
	}

	otherClient := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	otherClient.RemoteAddr = "192.0.2.11:4000"
	otherResponse := httptest.NewRecorder()
	server.ServeHTTP(otherResponse, otherClient)
	if otherResponse.Code != http.StatusOK {
		t.Fatalf("independent client status = %d, want %d", otherResponse.Code, http.StatusOK)
	}
}

func TestTrustedProxyIdentityWalksForwardedChainFromRight(t *testing.T) {
	options := testOptions()
	options.TrustedProxyCIDRs = []string{"10.0.0.0/8"}
	server := mustServer(t, database.NewMemoryStore(), options)

	trustedRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	trustedRequest.RemoteAddr = "10.0.0.3:443"
	trustedRequest.Header.Set("X-Forwarded-For", "192.0.2.99, 203.0.113.8, 10.0.0.2")
	if actual := server.clientIdentity(trustedRequest); actual != "203.0.113.8" {
		t.Fatalf("trusted proxy identity = %q, want %q", actual, "203.0.113.8")
	}

	untrustedRequest := httptest.NewRequest(http.MethodGet, "/health", nil)
	untrustedRequest.RemoteAddr = "198.51.100.7:443"
	untrustedRequest.Header.Set("X-Forwarded-For", "203.0.113.9")
	if actual := server.clientIdentity(untrustedRequest); actual != "198.51.100.7" {
		t.Fatalf("untrusted peer identity = %q", actual)
	}
}

func TestNewServerRejectsUnsafeOptions(t *testing.T) {
	tests := map[string]func(*Options){
		"non HTTP CORS origin": func(options *Options) {
			options.AllowedOrigins = []string{"javascript:alert(1)"}
		},
		"mixed wildcard origin": func(options *Options) {
			options.AllowedOrigins = []string{"*", "https://planner.gc.ca"}
		},
		"invalid proxy CIDR": func(options *Options) {
			options.TrustedProxyCIDRs = []string{"10.0.0.1"}
		},
		"missing timeout": func(options *Options) {
			options.RequestTimeout = 0
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			options := testOptions()
			mutate(&options)
			if _, err := NewServerWithOptions(database.NewMemoryStore(), options); err == nil {
				t.Fatal("NewServerWithOptions accepted unsafe options")
			}
		})
	}
}

func TestReadinessChecksStoreAndMetricsFailClosed(t *testing.T) {
	server := mustServer(t, &radarStore{err: errors.New("database offline")}, testOptions())

	for _, path := range []string{"/ready", "/metrics"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status = %d, want %d", path, response.Code, http.StatusServiceUnavailable)
		}
		if strings.Contains(response.Body.String(), "database offline") {
			t.Fatalf("%s leaked internal error: %s", path, response.Body.String())
		}
	}

	readyServer := mustServer(t, database.NewMemoryStore(), testOptions())
	response := httptest.NewRecorder()
	readyServer.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("ready status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestReadinessTimeoutCancelsDependencyCheck(t *testing.T) {
	options := testOptions()
	options.ReadinessTimeout = 10 * time.Millisecond
	server := mustServer(t, waitingRadarStore{}, options)

	started := time.Now()
	response := httptest.NewRecorder()
	server.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("readiness dependency check was not bounded: %s", elapsed)
	}
}

func TestRequestIDIsAvailableToStoreContext(t *testing.T) {
	store := &requestIDStore{}
	server := mustServer(t, store, testOptions())
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	request.Header.Set("X-Request-ID", "agency-correlation-42")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if store.requestID != "agency-correlation-42" {
		t.Fatalf("store context request ID = %q", store.requestID)
	}
}

func TestPanicResponseAndLogsRedactRecoveredValue(t *testing.T) {
	var logs bytes.Buffer
	options := testOptions()
	options.Logger = log.New(&logs, "", 0)
	server := mustServer(t, panicListStore{}, options)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	request.Header.Set("X-Request-ID", "safe-correlation-id")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	for location, value := range map[string]string{"response": response.Body.String(), "logs": logs.String()} {
		if strings.Contains(value, "super-secret-password") {
			t.Fatalf("%s leaked recovered panic value: %s", location, value)
		}
	}
	if !strings.Contains(logs.String(), "safe-correlation-id") {
		t.Fatalf("log omitted safe request ID: %s", logs.String())
	}
}

func testOptions() Options {
	options := DefaultOptions()
	options.RateLimitPerMinute = 0
	return options
}

func mustServer(t *testing.T, store database.Store, options Options) *Server {
	t.Helper()
	server, err := NewServerWithOptions(store, options)
	if err != nil {
		t.Fatalf("NewServerWithOptions: %v", err)
	}
	return server
}

type radarStore struct {
	database.Store
	err error
}

func (s *radarStore) GetRadarStats(context.Context) (*database.RadarStats, error) {
	return nil, s.err
}

type panicListStore struct {
	database.Store
}

func (panicListStore) ListProjects(context.Context, database.ProjectFilter) ([]*domain.Project, int, error) {
	panic("super-secret-password")
}

type waitingRadarStore struct {
	database.Store
}

func (waitingRadarStore) GetRadarStats(ctx context.Context) (*database.RadarStats, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

type requestIDStore struct {
	database.Store
	requestID string
}

func (s *requestIDStore) ListProjects(ctx context.Context, _ database.ProjectFilter) ([]*domain.Project, int, error) {
	s.requestID = RequestIDFromContext(ctx)
	return nil, 0, nil
}
