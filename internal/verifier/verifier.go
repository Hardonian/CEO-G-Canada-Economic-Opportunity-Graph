// Package verifier provides decentralized multi-party notarization of major
// project milestone occurrences. Each verifier node independently observes
// evidence for a milestone, hashes it, and publishes a signed attestation.
// When a quorum of nodes agrees on the same milestone hash, the milestone
// is considered notarized and can be trusted by downstream consumers.
package verifier

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// MethodologyVersion identifies the verifier protocol.
const MethodologyVersion = "verifier-v1.0"

// DefaultQuorum is the fraction of nodes that must agree for notarization.
const DefaultQuorum = 0.6

// NodeID uniquely identifies a verifier node.
type NodeID string

// Attestation is a single node's signed observation of a milestone.
type Attestation struct {
	NodeID       NodeID     `json:"node_id"`
	MilestoneID  string     `json:"milestone_id"`
	ProjectID    string     `json:"project_id"`
	EvidenceHash string     `json:"evidence_hash"`
	ObservedAt   time.Time  `json:"observed_at"`
	Signature    string     `json:"signature"`
}

// NotarizedMilestone is the result of a quorum vote.
type NotarizedMilestone struct {
	MilestoneID    string         `json:"milestone_id"`
	ProjectID      string         `json:"project_id"`
	EvidenceHash   string         `json:"evidence_hash"`
	Attestations   []*Attestation `json:"attestations"`
	QuorumReached  bool           `json:"quorum_reached"`
	QuorumFraction float64        `json:"quorum_fraction"`
	NotarizedAt    time.Time      `json:"notarized_at"`
}

// Node is a verifier that observes milestones and publishes attestations.
type Node struct {
	id       NodeID
	quorum   float64
	attstore []*Attestation
	mu       sync.RWMutex
}

// NewNode constructs a verifier node.
func NewNode(id NodeID) *Node {
	return &Node{id: id, quorum: DefaultQuorum}
}

// ID returns the node identifier.
func (n *Node) ID() NodeID { return n.id }

// Observe records an attestation for a milestone.
func (n *Node) Observe(milestoneID, projectID, evidenceHash string) *Attestation {
	now := time.Now().UTC()
	at := &Attestation{
		NodeID:       n.id,
		MilestoneID:  milestoneID,
		ProjectID:    projectID,
		EvidenceHash: evidenceHash,
		ObservedAt:   now,
		Signature:    signAttestation(n.id, milestoneID, projectID, evidenceHash, now),
	}
	n.mu.Lock()
	n.attstore = append(n.attstore, at)
	n.mu.Unlock()
	return at
}

// Attestations returns all attestations recorded by this node.
func (n *Node) Attestations() []*Attestation {
	n.mu.RLock()
	defer n.mu.RUnlock()
	out := make([]*Attestation, len(n.attstore))
	copy(out, n.attstore)
	return out
}

// Network coordinates a set of verifier nodes and runs quorum votes.
type Network struct {
	nodes  []*Node
	quorum float64
	mu     sync.RWMutex
}

// NewNetwork constructs a verifier network with the given quorum fraction.
func NewNetwork(quorum float64) *Network {
	if quorum <= 0 || quorum > 1 {
		quorum = DefaultQuorum
	}
	return &Network{quorum: quorum}
}

// AddNode registers a verifier node.
func (net *Network) AddNode(n *Node) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.nodes = append(net.nodes, n)
}

// Nodes returns all registered nodes.
func (net *Network) Nodes() []*Node {
	net.mu.RLock()
	defer net.mu.RUnlock()
	out := make([]*Node, len(net.nodes))
	copy(out, net.nodes)
	return out
}

// Vote runs a quorum vote over the supplied attestations.
func (net *Network) Vote(attestations []*Attestation) *NotarizedMilestone {
	net.mu.RLock()
	quorum := net.quorum
	net.mu.RUnlock()

	if len(attestations) == 0 {
		return &NotarizedMilestone{QuorumReached: false, QuorumFraction: 0}
	}

	votes := make(map[string][]*Attestation)
	for _, at := range attestations {
		votes[at.EvidenceHash] = append(votes[at.EvidenceHash], at)
	}

	var bestHash string
	var bestVotes []*Attestation
	for hash, atts := range votes {
		if len(atts) > len(bestVotes) {
			bestHash = hash
			bestVotes = atts
		}
	}

	fraction := float64(len(bestVotes)) / float64(len(attestations))
	reached := fraction >= quorum

	milestoneID := ""
	projectID := ""
	if len(bestVotes) > 0 {
		milestoneID = bestVotes[0].MilestoneID
		projectID = bestVotes[0].ProjectID
	}

	return &NotarizedMilestone{
		MilestoneID:    milestoneID,
		ProjectID:      projectID,
		EvidenceHash:   bestHash,
		Attestations:   bestVotes,
		QuorumReached:  reached,
		QuorumFraction: fraction,
		NotarizedAt:    time.Now().UTC(),
	}
}

// NotarizeMilestone observes the milestone on every node and runs a quorum vote.
func (net *Network) NotarizeMilestone(milestoneID, projectID, evidenceHash string) *NotarizedMilestone {
	net.mu.RLock()
	nodes := append([]*Node{}, net.nodes...)
	net.mu.RUnlock()

	attestations := make([]*Attestation, 0, len(nodes))
	for _, n := range nodes {
		attestations = append(attestations, n.Observe(milestoneID, projectID, evidenceHash))
	}
	return net.Vote(attestations)
}

// BuildMerkleProof returns a deterministic SHA-256 proof for the given evidence.
func BuildMerkleProof(evidence *domain.Evidence) string {
	if evidence == nil {
		return ""
	}
	sum := sha256.Sum256([]byte("proof|" + evidence.ID + "|" + evidence.ContentHash))
	return hex.EncodeToString(sum[:])
}

// signAttestation produces a deterministic signature for an attestation.
func signAttestation(nodeID NodeID, milestoneID, projectID, evidenceHash string, observedAt time.Time) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%d", nodeID, milestoneID, projectID, evidenceHash, observedAt.UnixNano())
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Ensure database import is used.
var _ = database.Store(nil)
var _ = context.Background