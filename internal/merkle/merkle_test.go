package merkle

import (
	"context"
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func TestBuildRoot_EmptyStore(t *testing.T) {
	store := database.NewMemoryStore()
	root := BuildRoot(context.Background(), store)
	if root == nil {
		t.Fatal("expected non-nil root for empty store")
	}
	if root.LeafCount != 0 {
		t.Fatalf("expected 0 leaves, got %d", root.LeafCount)
	}
	if root.RootHash == "" {
		t.Fatal("expected deterministic empty-tree root")
	}
	// Empty-tree root must be stable across calls.
	root2 := BuildRoot(context.Background(), store)
	if root.RootHash != root2.RootHash {
		t.Fatalf("empty-tree root not deterministic: %s vs %s", root.RootHash, root2.RootHash)
	}
}

func TestBuildRoot_WithEvidence(t *testing.T) {
	store := database.NewMemoryStore()
	now := time.Now().UTC()
	proj := &domain.Project{
		ID:        "p1",
		Name:      "Test Project",
		Sector:    domain.SectorNuclearEnergy,
		Province:  "ON",
		CurrentStage: domain.StageFeasibility,
		CapexCAD:  1_000_000_000,
		EvidenceIDs: []string{"ev-1", "ev-2"},
		UpdatedAt: now,
	}
	if err := store.SaveProject(context.Background(), proj); err != nil {
		t.Fatal(err)
	}
	for _, ev := range []*domain.Evidence{
		{ID: "ev-1", ContentHash: "abc123", SourceURL: "https://example.com/1", Publisher: "Test", SourceTier: domain.SourceTier1, RetrievalTimestamp: now, Confidence: domain.ConfidenceVerified, ExtractionMethod: "test", Visibility: domain.VisibilityPublic, Publishable: true},
		{ID: "ev-2", ContentHash: "def456", SourceURL: "https://example.com/2", Publisher: "Test", SourceTier: domain.SourceTier2, RetrievalTimestamp: now, Confidence: domain.ConfidenceSupported, ExtractionMethod: "test", Visibility: domain.VisibilityPublic, Publishable: true},
	} {
		if err := store.SaveEvidence(context.Background(), ev); err != nil {
			t.Fatal(err)
		}
	}

	root := BuildRoot(context.Background(), store)
	if root.LeafCount != 2 {
		t.Fatalf("expected 2 leaves, got %d", root.LeafCount)
	}
	if root.RootHash == "" {
		t.Fatal("expected non-empty root hash")
	}
	// Same dataset must produce the same root.
	root2 := BuildRoot(context.Background(), store)
	if root.RootHash != root2.RootHash {
		t.Fatalf("root not deterministic: %s vs %s", root.RootHash, root2.RootHash)
	}
	// Different content must produce a different root.
	_ = store.SaveEvidence(context.Background(), &domain.Evidence{ID: "ev-3", ContentHash: "different", SourceURL: "https://example.com/3", Publisher: "Test", SourceTier: domain.SourceTier1, RetrievalTimestamp: now, Confidence: domain.ConfidenceVerified, ExtractionMethod: "test", Visibility: domain.VisibilityPublic, Publishable: true})
	proj.EvidenceIDs = append(proj.EvidenceIDs, "ev-3")
	_ = store.SaveProject(context.Background(), proj)
	root3 := BuildRoot(context.Background(), store)
	if root3.RootHash == root.RootHash {
		t.Fatal("expected different root after adding evidence")
	}
}

func TestPublishRoot_ReturnsHash(t *testing.T) {
	store := database.NewMemoryStore()
	hash := PublishRoot(context.Background(), store)
	if hash == "" {
		t.Fatal("expected non-empty hash from PublishRoot")
	}
}

func TestBuildRoot_DedupesEvidence(t *testing.T) {
	store := database.NewMemoryStore()
	now := time.Now().UTC()
	proj := &domain.Project{
		ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, Province: "ON",
		CurrentStage: domain.StageFeasibility, CapexCAD: 1_000_000_000,
		EvidenceIDs: []string{"ev-1", "ev-1"}, UpdatedAt: now,
	}
	if err := store.SaveProject(context.Background(), proj); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveEvidence(context.Background(), &domain.Evidence{ID: "ev-1", ContentHash: "abc", SourceURL: "https://example.com", Publisher: "T", SourceTier: domain.SourceTier1, RetrievalTimestamp: now, Confidence: domain.ConfidenceVerified, ExtractionMethod: "test", Visibility: domain.VisibilityPublic, Publishable: true}); err != nil {
		t.Fatal(err)
	}
	root := BuildRoot(context.Background(), store)
	if root.LeafCount != 1 {
		t.Fatalf("expected 1 leaf after dedup, got %d", root.LeafCount)
	}
}

func TestBuildRoot_SortedLeaves(t *testing.T) {
	store := database.NewMemoryStore()
	now := time.Now().UTC()
	proj := &domain.Project{
		ID: "p1", Name: "Test", Sector: domain.SectorNuclearEnergy, Province: "ON",
		CurrentStage: domain.StageFeasibility, CapexCAD: 1_000_000_000,
		EvidenceIDs: []string{"ev-b", "ev-a", "ev-c"}, UpdatedAt: now,
	}
	if err := store.SaveProject(context.Background(), proj); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"ev-a", "ev-b", "ev-c"} {
		if err := store.SaveEvidence(context.Background(), &domain.Evidence{ID: id, ContentHash: id + "-hash", SourceURL: "https://example.com/" + id, Publisher: "T", SourceTier: domain.SourceTier1, RetrievalTimestamp: now, Confidence: domain.ConfidenceVerified, ExtractionMethod: "test", Visibility: domain.VisibilityPublic, Publishable: true}); err != nil {
			t.Fatal(err)
		}
	}
	root := BuildRoot(context.Background(), store)
	for i := 1; i < len(root.Leaves); i++ {
		if root.Leaves[i-1].EvidenceID > root.Leaves[i].EvidenceID {
			t.Fatalf("leaves not sorted: %s > %s", root.Leaves[i-1].EvidenceID, root.Leaves[i].EvidenceID)
		}
	}
}