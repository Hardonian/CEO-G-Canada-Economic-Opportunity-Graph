package verifier_test

import (
	"testing"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/verifier"
)

// ─── AttestationStore ─────────────────────────────────────────────────────────

func TestAttestationStore_SaveAndGet(t *testing.T) {
	s := verifier.NewAttestationStore()
	node := verifier.NewNodeWithStore("node-1", s)

	at := node.Observe("milestone-A", "proj-1", "hash-abc")
	if at == nil {
		t.Fatal("expected attestation, got nil")
	}
	if s.Count() != 1 {
		t.Errorf("expected 1 attestation in store, got %d", s.Count())
	}

	byM := s.GetByMilestone("milestone-A")
	if len(byM) != 1 {
		t.Fatalf("expected 1 attestation for milestone-A, got %d", len(byM))
	}
	if byM[0].NodeID != "node-1" {
		t.Errorf("expected nodeID node-1, got %s", byM[0].NodeID)
	}
}

func TestAttestationStore_MultiNode(t *testing.T) {
	s := verifier.NewAttestationStore()
	n1 := verifier.NewNodeWithStore("alpha", s)
	n2 := verifier.NewNodeWithStore("beta", s)
	n3 := verifier.NewNodeWithStore("gamma", s)

	n1.Observe("ms-1", "proj-x", "hash-1")
	n2.Observe("ms-1", "proj-x", "hash-1")
	n3.Observe("ms-1", "proj-x", "hash-1")

	if s.Count() != 3 {
		t.Errorf("expected 3 attestations, got %d", s.Count())
	}
	byM := s.GetByMilestone("ms-1")
	if len(byM) != 3 {
		t.Errorf("expected 3 attestations for ms-1, got %d", len(byM))
	}
}

func TestAttestationStore_GetAll(t *testing.T) {
	s := verifier.NewAttestationStore()
	n := verifier.NewNodeWithStore("node-a", s)
	n.Observe("ms-1", "proj-1", "h1")
	n.Observe("ms-2", "proj-2", "h2")

	all := s.GetAll()
	if len(all) != 2 {
		t.Errorf("expected 2 attestations, got %d", len(all))
	}
}

func TestAttestationStore_MilestoneIDs(t *testing.T) {
	s := verifier.NewAttestationStore()
	n := verifier.NewNodeWithStore("n", s)
	n.Observe("ms-X", "p", "h")
	n.Observe("ms-Y", "p", "h")

	ids := s.MilestoneIDs()
	if len(ids) != 2 {
		t.Errorf("expected 2 milestone IDs, got %d: %v", len(ids), ids)
	}
}

func TestAttestationStore_Prune(t *testing.T) {
	s := verifier.NewAttestationStore()
	n := verifier.NewNodeWithStore("n", s)

	// Record an old attestation then two recent ones.
	n.Observe("ms-old", "p", "h")
	time.Sleep(2 * time.Millisecond) // ensure time ordering
	cutoff := time.Now()
	time.Sleep(2 * time.Millisecond)
	n.Observe("ms-new-1", "p", "h")
	n.Observe("ms-new-2", "p", "h")

	removed := s.Prune(cutoff)
	if removed != 1 {
		t.Errorf("expected 1 pruned, got %d", removed)
	}
	if s.Count() != 2 {
		t.Errorf("expected 2 remaining, got %d", s.Count())
	}
	if len(s.GetByMilestone("ms-old")) != 0 {
		t.Error("ms-old should have been pruned")
	}
}

func TestAttestationStore_NilSave(t *testing.T) {
	s := verifier.NewAttestationStore()
	s.Save(nil) // must not panic
	if s.Count() != 0 {
		t.Error("nil save should be a no-op")
	}
}

// ─── Node without store (backward compat) ─────────────────────────────────────

func TestNode_WithoutStore(t *testing.T) {
	n := verifier.NewNode("standalone")
	at := n.Observe("ms-1", "proj-z", "hashZ")
	if at == nil {
		t.Fatal("expected attestation")
	}
	if at.MilestoneID != "ms-1" {
		t.Errorf("expected ms-1, got %s", at.MilestoneID)
	}
	all := n.Attestations()
	if len(all) != 1 {
		t.Errorf("expected 1 in-process attestation, got %d", len(all))
	}
}

// ─── Network quorum voting ────────────────────────────────────────────────────

func TestNetwork_QuorumReached(t *testing.T) {
	store := verifier.NewAttestationStore()
	net := verifier.NewNetwork(0.6)
	for i := 0; i < 5; i++ {
		n := verifier.NewNodeWithStore(verifier.NodeID("node-"+string(rune('A'+i))), store)
		net.AddNode(n)
	}

	result := net.NotarizeMilestone("ms-quorum", "proj-quorum", "hash-q")
	if !result.QuorumReached {
		t.Errorf("expected quorum reached, fraction=%.2f", result.QuorumFraction)
	}
	if result.QuorumFraction != 1.0 {
		t.Errorf("expected 1.0 fraction (all agree), got %.2f", result.QuorumFraction)
	}
	// Attestations should also be in the store.
	if store.Count() != 5 {
		t.Errorf("expected 5 attestations in store, got %d", store.Count())
	}
}

func TestNetwork_QuorumNotReached(t *testing.T) {
	net := verifier.NewNetwork(0.8)
	n1 := verifier.NewNode("n1")
	n2 := verifier.NewNode("n2")
	n3 := verifier.NewNode("n3")
	net.AddNode(n1)
	net.AddNode(n2)
	net.AddNode(n3)

	// 2 out of 3 agree (0.667) — below 0.8 quorum.
	attestations := []*verifier.Attestation{
		n1.Observe("ms-split", "p", "hash-A"),
		n2.Observe("ms-split", "p", "hash-A"),
		n3.Observe("ms-split", "p", "hash-B"),
	}
	result := net.Vote(attestations)
	if result.QuorumReached {
		t.Errorf("expected quorum NOT reached, fraction=%.2f", result.QuorumFraction)
	}
}

func TestNetwork_EmptyVote(t *testing.T) {
	net := verifier.NewNetwork(0.6)
	result := net.Vote(nil)
	if result.QuorumReached {
		t.Error("empty vote should not reach quorum")
	}
}

func TestBuildMerkleProof(t *testing.T) {
	// nil input must not panic.
	if got := verifier.BuildMerkleProof(nil); got != "" {
		t.Errorf("expected empty string for nil evidence, got %q", got)
	}
}
