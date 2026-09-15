package graphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	gql "github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/graphql"
)

// ─── stub store ───────────────────────────────────────────────────────────────

type stubStore struct{}

var _ database.Store = (*stubStore)(nil)

func (s *stubStore) SaveProject(_ context.Context, _ *domain.Project) error { return nil }
func (s *stubStore) GetProject(_ context.Context, id string) (*domain.Project, error) {
	if id == "proj-test-1" {
		return &domain.Project{
			ID:           "proj-test-1",
			Slug:         "test-project",
			Name:         "Test Nuclear Project",
			Sector:       domain.SectorNuclearEnergy,
			Province:     "ON",
			CurrentStage: domain.StageConstruction,
			CapexCAD:     4000000000,
			UpdatedAt:    time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("project %q not found", id)
}
func (s *stubStore) GetProjectBySlug(_ context.Context, slug string) (*domain.Project, error) {
	if slug == "test-project" {
		return &domain.Project{
			ID:           "proj-test-1",
			Slug:         "test-project",
			Name:         "Test Nuclear Project",
			Sector:       domain.SectorNuclearEnergy,
			Province:     "ON",
			CurrentStage: domain.StageConstruction,
			CapexCAD:     4000000000,
			UpdatedAt:    time.Now(),
		}, nil
	}
	return nil, fmt.Errorf("project slug %q not found", slug)
}
func (s *stubStore) ListProjects(_ context.Context, _ database.ProjectFilter) ([]*domain.Project, int, error) {
	return []*domain.Project{
		{ID: "proj-a", Slug: "project-a", Name: "Project A", Sector: domain.SectorCleanEnergy, Province: "BC", CurrentStage: domain.StageConcept, CapexCAD: 1000000000, UpdatedAt: time.Now()},
		{ID: "proj-b", Slug: "project-b", Name: "Project B", Sector: domain.SectorMiningMetals, Province: "AB", CurrentStage: domain.StagePermitting, CapexCAD: 500000000, UpdatedAt: time.Now()},
	}, 2, nil
}
func (s *stubStore) SaveEntity(_ context.Context, _ *domain.Entity) error { return nil }
func (s *stubStore) GetEntity(_ context.Context, id string) (*domain.Entity, error) {
	return nil, fmt.Errorf("entity %q not found", id)
}
func (s *stubStore) FindEntityByLegalOrAlias(_ context.Context, name string) (*domain.Entity, error) {
	return nil, fmt.Errorf("entity %q not found", name)
}
func (s *stubStore) ListEntities(_ context.Context) ([]*domain.Entity, error) {
	return []*domain.Entity{
		{ID: "org-1", Slug: "acme-energy", CommonName: "Acme Energy", LegalName: "Acme Energy Corp.", EntityType: "Corporation", UpdatedAt: time.Now()},
	}, nil
}
func (s *stubStore) SaveEvent(_ context.Context, _ *domain.Event) error { return nil }
func (s *stubStore) ListEventsByProject(_ context.Context, _ string) ([]*domain.Event, error) {
	return []*domain.Event{}, nil
}
func (s *stubStore) ListRecentEvents(_ context.Context, _ int) ([]*domain.Event, error) {
	return []*domain.Event{
		{ID: "evt-1", ProjectID: "proj-a", EventType: "stage_change", EventDate: time.Now(), Title: "FID reached", Description: "Financial Investment Decision"},
	}, nil
}
func (s *stubStore) SaveRelationship(_ context.Context, _ *domain.Relationship) error { return nil }
func (s *stubStore) ListRelationshipsByProject(_ context.Context, _ string) ([]*domain.Relationship, error) {
	return nil, nil
}
func (s *stubStore) SaveScore(_ context.Context, _ *domain.ProjectScore) error { return nil }
func (s *stubStore) GetLatestScores(_ context.Context, _ string) (map[string]*domain.ProjectScore, error) {
	return nil, nil
}
func (s *stubStore) ListScoreHistory(_ context.Context, _, _ string) ([]*domain.ProjectScore, error) {
	return nil, nil
}
func (s *stubStore) ListRankings(_ context.Context, _ string, _ int) ([]*domain.Project, error) {
	return nil, nil
}
func (s *stubStore) SaveProcurement(_ context.Context, _ *domain.Procurement) error { return nil }
func (s *stubStore) ListProcurements(_ context.Context, _, _ int) ([]*domain.Procurement, error) {
	return nil, nil
}
func (s *stubStore) ListProcurementsByProject(_ context.Context, _ string) ([]*domain.Procurement, error) {
	return nil, nil
}
func (s *stubStore) SaveCapitalItem(_ context.Context, _ *domain.CapitalItem) error { return nil }
func (s *stubStore) ListCapitalItemsByProject(_ context.Context, _ string) ([]*domain.CapitalItem, error) {
	return nil, nil
}
func (s *stubStore) ListAllCapitalItems(_ context.Context) ([]*domain.CapitalItem, error) {
	return nil, nil
}
func (s *stubStore) SaveSignal(_ context.Context, _ *domain.Signal) error { return nil }
func (s *stubStore) ListSignals(_ context.Context, _ time.Duration, _ int) ([]*domain.Signal, error) {
	return []*domain.Signal{
		{ID: "sig-1", ProjectID: "proj-a", Type: domain.SignalConstructionSignal, Magnitude: 0.85, Timestamp: time.Now(), Description: "Construction started"},
	}, nil
}
func (s *stubStore) SaveOpportunity(_ context.Context, _ *domain.Opportunity) error { return nil }
func (s *stubStore) GetOpportunity(_ context.Context, id string) (*domain.Opportunity, error) {
	return nil, fmt.Errorf("opportunity %q not found", id)
}
func (s *stubStore) ListOpportunitiesByProject(_ context.Context, _ string) ([]*domain.Opportunity, error) {
	return nil, nil
}
func (s *stubStore) ListAllOpportunities(_ context.Context) ([]*domain.Opportunity, error) {
	return nil, nil
}
func (s *stubStore) SaveClaim(_ context.Context, _ *domain.Claim) error { return nil }
func (s *stubStore) GetClaim(_ context.Context, id string) (*domain.Claim, error) {
	return nil, fmt.Errorf("claim %q not found", id)
}
func (s *stubStore) ListClaimsBySubject(_ context.Context, _ string) ([]*domain.Claim, error) {
	return nil, nil
}
func (s *stubStore) SaveCandidateProject(_ context.Context, _ *domain.CandidateProject) error {
	return nil
}
func (s *stubStore) GetCandidateProject(_ context.Context, id string) (*domain.CandidateProject, error) {
	return nil, fmt.Errorf("candidate project %q not found", id)
}
func (s *stubStore) SaveProjectPhase(_ context.Context, _ *domain.ProjectPhase) error { return nil }
func (s *stubStore) ListProjectPhases(_ context.Context, _ string) ([]*domain.ProjectPhase, error) {
	return nil, nil
}
func (s *stubStore) SaveCapitalRequirement(_ context.Context, _ *domain.CapitalRequirement) error {
	return nil
}
func (s *stubStore) ListCapitalRequirements(_ context.Context, _ string) ([]*domain.CapitalRequirement, error) {
	return nil, nil
}
func (s *stubStore) SaveCapitalNeed(_ context.Context, _ *domain.CapitalNeed) error { return nil }
func (s *stubStore) GetCapitalNeed(_ context.Context, id string) (*domain.CapitalNeed, error) {
	return nil, fmt.Errorf("capital need %q not found", id)
}
func (s *stubStore) ListCapitalNeeds(_ context.Context, _ string) ([]*domain.CapitalNeed, error) {
	return nil, nil
}
func (s *stubStore) SaveMilestone(_ context.Context, _ *domain.Milestone) error { return nil }
func (s *stubStore) ListMilestones(_ context.Context, _ string) ([]*domain.Milestone, error) {
	return nil, nil
}
func (s *stubStore) SaveReadinessAssessment(_ context.Context, _ *domain.ReadinessAssessment) error {
	return nil
}
func (s *stubStore) GetLatestReadinessAssessment(_ context.Context, _ string) (*domain.ReadinessAssessment, error) {
	return nil, fmt.Errorf("readiness assessment not found")
}
func (s *stubStore) SaveAuditEntry(_ context.Context, _ *domain.AuditEntry) error { return nil }
func (s *stubStore) ListAuditEntries(_ context.Context, _ string) ([]*domain.AuditEntry, error) {
	return nil, nil
}
func (s *stubStore) SaveEvidence(_ context.Context, _ *domain.Evidence) error { return nil }
func (s *stubStore) GetEvidence(_ context.Context, id string) (*domain.Evidence, error) {
	return nil, fmt.Errorf("evidence %q not found", id)
}
func (s *stubStore) SaveTradeMetric(_ context.Context, _ *domain.TradeMetric) error { return nil }
func (s *stubStore) ListTradeMetrics(_ context.Context, _ string) ([]*domain.TradeMetric, error) {
	return nil, nil
}
func (s *stubStore) GetRadarStats(_ context.Context) (*database.RadarStats, error) {
	return &database.RadarStats{
		TotalProjects: 2,
		TotalCapexCAD: 1500000000,
		DataStatus:    domain.StatusHealthy,
		GeneratedAt:   time.Now(),
	}, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	return gql.NewHandler(&stubStore{})
}

func doGET(t *testing.T, h http.Handler, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/graphql?query="+query, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func doPOST(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/graphql", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, w.Body.String())
	}
	return resp
}

// ─── tests ────────────────────────────────────────────────────────────────────

func TestGraphQL_GET_Projects(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, `{projects{id+name}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T: %v", resp["data"], resp)
	}
	projects, ok := data["projects"].([]interface{})
	if !ok {
		t.Fatalf("expected projects array, got %T", data["projects"])
	}
	if len(projects) < 2 {
		t.Errorf("expected at least 2 projects, got %d", len(projects))
	}
}

func TestGraphQL_POST_Project(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ project(id: \"proj-test-1\") { id name stage } }"}`
	w := doPOST(t, h, body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %v", resp)
	}
	proj, ok := data["project"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected project object, got %T", data["project"])
	}
	if proj["id"] != "proj-test-1" {
		t.Errorf("expected project id 'proj-test-1', got %v", proj["id"])
	}
}

func TestGraphQL_POST_ProjectNotFound(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ project(id: \"does-not-exist\") { id } }"}`
	w := doPOST(t, h, body)
	// Not-found returns 200 with data.project = null
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object")
	}
	if data["project"] != nil {
		t.Errorf("expected project=null for not-found, got %v", data["project"])
	}
}

func TestGraphQL_GET_Organizations(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, `{organizations{id+commonName}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	data := resp["data"].(map[string]interface{})
	orgs, ok := data["organizations"].([]interface{})
	if !ok || len(orgs) == 0 {
		t.Errorf("expected at least 1 organization, got %v", data["organizations"])
	}
}

func TestGraphQL_GET_Events(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, `{events{id+title}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	data := resp["data"].(map[string]interface{})
	if _, ok := data["events"]; !ok {
		t.Errorf("expected events field in response")
	}
}

func TestGraphQL_GET_Signals(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, `{signals{id+signalType+strength}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	data := resp["data"].(map[string]interface{})
	signals, ok := data["signals"].([]interface{})
	if !ok || len(signals) == 0 {
		t.Errorf("expected at least 1 signal, got %v", data["signals"])
	}
}

func TestGraphQL_GET_Reconciliation(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, `{reconciliation{totalRecords+merged+linked+conflicts}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	data := resp["data"].(map[string]interface{})
	rec, ok := data["reconciliation"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected reconciliation object, got %T", data["reconciliation"])
	}
	total, _ := rec["totalRecords"].(float64)
	if total < 1 {
		t.Errorf("expected totalRecords >= 1, got %v", total)
	}
}

func TestGraphQL_Introspection_Schema(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, `{__schema{sdl}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object")
	}
	schema, ok := data["__schema"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected __schema object")
	}
	sdl, _ := schema["sdl"].(string)
	if !strings.Contains(sdl, "type Query") {
		t.Errorf("__schema.sdl does not contain 'type Query', got: %s", sdl[:min(200, len(sdl))])
	}
}

func TestGraphQL_EmptyQuery(t *testing.T) {
	h := newTestHandler(t)
	w := doGET(t, h, "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty query, got %d", w.Code)
	}
}

func TestGraphQL_InvalidMethod(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/graphql", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestGraphQL_UnknownField(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ notARealField { id } }"}`
	w := doPOST(t, h, body)
	// Unknown field returns errors.
	resp := decodeResponse(t, w)
	if _, hasErrors := resp["errors"]; !hasErrors {
		t.Errorf("expected errors for unknown field, got: %v", resp)
	}
}

func TestGraphQL_POST_BadContentType(t *testing.T) {
	h := newTestHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/graphql", strings.NewReader(`{"query":"{projects{id}}"}`))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", w.Code)
	}
}

// ─── Field Projection ────────────────────────────────────────────────────────

func TestGraphQL_FieldProjection_Projects(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ projects { id name } }"}`
	w := doPOST(t, h, body)
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	projects, ok := data["projects"].([]interface{})
	if !ok || len(projects) == 0 {
		t.Fatalf("expected non-empty projects list, got %v", data["projects"])
	}
	first, ok := projects[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected project object, got %T", projects[0])
	}
	// Only id and name should be present.
	if _, hasSlug := first["slug"]; hasSlug {
		t.Errorf("field projection failed: slug present but not requested: %v", first)
	}
	if _, hasSector := first["sector"]; hasSector {
		t.Errorf("field projection failed: sector present but not requested: %v", first)
	}
	if first["id"] == nil {
		t.Error("id should be present")
	}
	if first["name"] == nil {
		t.Error("name should be present")
	}
}

func TestGraphQL_FieldProjection_SingleProject(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ project(id: \"proj-test-1\") { id name stage } }"}`
	w := doPOST(t, h, body)
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	proj, ok := data["project"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected project object, got %T", data["project"])
	}
	if _, hasSlug := proj["slug"]; hasSlug {
		t.Errorf("field projection failed: slug present but not requested: %v", proj)
	}
	if _, hasCapex := proj["capexCAD"]; hasCapex {
		t.Errorf("field projection failed: capexCAD present but not requested: %v", proj)
	}
	if proj["id"] == nil {
		t.Error("id should be present")
	}
	if proj["name"] == nil {
		t.Error("name should be present")
	}
	if proj["stage"] == nil {
		t.Error("stage should be present")
	}
}

func TestGraphQL_FieldProjection_Organizations(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ organizations { id commonName } }"}`
	w := doPOST(t, h, body)
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	orgs, ok := data["organizations"].([]interface{})
	if !ok || len(orgs) == 0 {
		t.Fatalf("expected non-empty organizations list, got %v", data["organizations"])
	}
	first, ok := orgs[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected org object, got %T", orgs[0])
	}
	if _, hasLegal := first["legalName"]; hasLegal {
		t.Errorf("field projection failed: legalName present but not requested: %v", first)
	}
	if first["id"] == nil {
		t.Error("id should be present")
	}
	if first["commonName"] == nil {
		t.Error("commonName should be present")
	}
}

func TestGraphQL_FieldProjection_Signals(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ signals { id signalType } }"}`
	w := doPOST(t, h, body)
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	signals, ok := data["signals"].([]interface{})
	if !ok || len(signals) == 0 {
		t.Fatalf("expected non-empty signals list, got %v", data["signals"])
	}
	first, ok := signals[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected signal object, got %T", signals[0])
	}
	if _, hasDesc := first["description"]; hasDesc {
		t.Errorf("field projection failed: description present but not requested: %v", first)
	}
	if first["id"] == nil {
		t.Error("id should be present")
	}
	if first["signalType"] == nil {
		t.Error("signalType should be present")
	}
}

func TestGraphQL_FieldProjection_EmptySelection(t *testing.T) {
	// Empty sub-selection should return all fields (backward compatible).
	h := newTestHandler(t)
	body := `{"query":"{ projects { } }"}`
	w := doPOST(t, h, body)
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	projects, ok := data["projects"].([]interface{})
	if !ok || len(projects) == 0 {
		t.Fatalf("expected non-empty projects list, got %v", data["projects"])
	}
	first, ok := projects[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected project object, got %T", projects[0])
	}
	// With empty selection, all fields should be present.
	if first["id"] == nil {
		t.Error("id should be present with empty selection")
	}
	if first["name"] == nil {
		t.Error("name should be present with empty selection")
	}
}

func TestGraphQL_FieldProjection_Events(t *testing.T) {
	h := newTestHandler(t)
	body := `{"query":"{ events { id eventType } }"}`
	w := doPOST(t, h, body)
	resp := decodeResponse(t, w)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	events, ok := data["events"].([]interface{})
	if !ok || len(events) == 0 {
		t.Fatalf("expected non-empty events list, got %v", data["events"])
	}
	first, ok := events[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected event object, got %T", events[0])
	}
	if _, hasTitle := first["title"]; hasTitle {
		t.Errorf("field projection failed: title present but not requested: %v", first)
	}
	if first["id"] == nil {
		t.Error("id should be present")
	}
	if first["eventType"] == nil {
		t.Error("eventType should be present")
	}
}
