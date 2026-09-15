package connector

import (
	"context"
	"testing"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/adapters/middleware"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// stubAdapter implements adapters.Adapter for testing.
type stubAdapter struct {
	name string
	tier domain.SourceTier
}

func (s *stubAdapter) Name() string                 { return s.name }
func (s *stubAdapter) Tier() domain.SourceTier      { return s.tier }
func (s *stubAdapter) Fetch(ctx context.Context) ([]byte, error) {
	return nil, nil
}
func (s *stubAdapter) Parse(data []byte) (*adapters.IngestionResult, error) {
	return &adapters.IngestionResult{}, nil
}
func (s *stubAdapter) Health() *adapters.SourceHealth {
	return &adapters.SourceHealth{AdapterName: s.name, Status: "HEALTHY"}
}
func (s *stubAdapter) PollHealth() *adapters.SourceHealth {
	return s.Health()
}

// Ensure stubAdapter satisfies the interface.
var _ adapters.Adapter = (*stubAdapter)(nil)

func newStub(name string) *stubAdapter {
	return &stubAdapter{name: name, tier: domain.SourceTier1}
}

func TestBuildRegistry_PreservesInsertionOrder(t *testing.T) {
	list := []adapters.Adapter{
		newStub("alpha"),
		newStub("beta"),
		newStub("gamma"),
	}
	reg := BuildRegistry(list)
	names := reg.Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 connectors, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" || names[2] != "gamma" {
		t.Errorf("order = %v, want [alpha beta gamma]", names)
	}
}

func TestBuildRegistry_GetByName(t *testing.T) {
	list := []adapters.Adapter{
		newStub("alpha"),
		newStub("beta"),
	}
	reg := BuildRegistry(list)
	conn := reg.Get("beta")
	if conn == nil {
		t.Fatal("expected non-nil connector for beta")
	}
	if conn.Name() != "beta" {
		t.Errorf("name = %q, want beta", conn.Name())
	}
	if got := reg.Get("nonexistent"); got != nil {
		t.Errorf("expected nil for absent name, got %v", got)
	}
}

func TestBuildRegistry_DuplicateNameReplaces(t *testing.T) {
	list := []adapters.Adapter{
		newStub("dup"),
		newStub("dup"),
	}
	reg := BuildRegistry(list)
	if reg.Get("dup") == nil {
		t.Fatal("expected connector for dup")
	}
	names := reg.Names()
	if len(names) != 1 || names[0] != "dup" {
		t.Errorf("names = %v, want [dup]", names)
	}
}

func TestBuildPipelineAdapters(t *testing.T) {
	list := []adapters.Adapter{
		newStub("alpha"),
		newStub("beta"),
	}
	adapters := BuildPipelineAdapters(list)
	if len(adapters) != 2 {
		t.Fatalf("expected 2 adapters, got %d", len(adapters))
	}
	if adapters[0].Name() != "alpha" || adapters[1].Name() != "beta" {
		t.Errorf("order = [%s, %s], want [alpha, beta]", adapters[0].Name(), adapters[1].Name())
	}
}

func TestDefaultMiddlewares(t *testing.T) {
	mw := DefaultMiddlewares()
	if len(mw) != 5 {
		t.Fatalf("expected 5 middlewares, got %d", len(mw))
	}
}

func TestRegistry_All(t *testing.T) {
	list := []adapters.Adapter{
		newStub("alpha"),
		newStub("beta"),
	}
	reg := BuildRegistry(list)
	all := reg.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 connectors, got %d", len(all))
	}
}

func TestRegistry_AggregateHealth(t *testing.T) {
	list := []adapters.Adapter{
		newStub("alpha"),
		newStub("beta"),
	}
	reg := BuildRegistry(list)
	snap := reg.AggregateHealth()
	if snap == nil {
		t.Fatal("expected non-nil health snapshot")
	}
	if snap.Healthy != 2 {
		t.Errorf("expected 2 healthy, got %d", snap.Healthy)
	}
	if snap.Degraded != 0 || snap.Broken != 0 {
		t.Errorf("expected 0 degraded/broken, got %d/%d", snap.Degraded, snap.Broken)
	}
}

func TestRegistry_AggregateHealth_Empty(t *testing.T) {
	reg := NewRegistry()
	snap := reg.AggregateHealth()
	if snap == nil {
		t.Fatal("expected non-nil health snapshot for empty registry")
	}
	if snap.Healthy != 0 {
		t.Errorf("expected 0 healthy, got %d", snap.Healthy)
	}
}

func TestAdapterNames(t *testing.T) {
	list := []adapters.Adapter{
		newStub("alpha"),
		newStub("beta"),
	}
	names := AdapterNames(list)
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("names = %v, want [alpha beta]", names)
	}
}

func TestNewRegistry_Empty(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(reg.Names()) != 0 {
		t.Error("expected empty names")
	}
	if len(reg.All()) != 0 {
		t.Error("expected empty All()")
	}
	if reg.Get("anything") != nil {
		t.Error("expected nil for absent name")
	}
}

// Ensure middleware import is used.
var _ = middleware.DefaultRetryConfig()