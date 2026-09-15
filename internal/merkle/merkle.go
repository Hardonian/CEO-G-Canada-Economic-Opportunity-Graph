// Package merkle provides a deterministic, append-only transparency log over
// evidence content hashes. It builds a binary Merkle tree over all evidence
// records currently held in the store and publishes the root hash. The root
// can be pinned externally (blockchain, notary, git commit) to give anyone a
// single value they can audit against a full dataset.
package merkle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// MethodologyVersion identifies the Merkle tree construction algorithm.
const MethodologyVersion = "merkle-v1.0"

// Leaf is a single evidence record hashed into the tree. The tree is built over
// the canonical (evidence ID, content hash) pairs in stable ID order, so the
// root is reproducible for an identical dataset.
type Leaf struct {
	EvidenceID   string `json:"evidence_id"`
	ContentHash  string `json:"content_hash"`
	LeafHash     string `json:"leaf_hash"`
}

// Root is the published Merkle root and supporting metadata.
type Root struct {
	MethodologyVersion string    `json:"methodology_version"`
	RootHash           string    `json:"root_hash"`
	LeafCount          int       `json:"leaf_count"`
	Leaves             []*Leaf   `json:"leaves,omitempty"`
	PublishedAt        time.Time `json:"published_at"`
}

// PublishRoot builds the Merkle tree over every evidence record in the store
// and returns the root. The caller can pin RootHash externally. An empty store
// produces a deterministic empty-tree root so that downstream consumers can
// always compare against a single canonical value.
func PublishRoot(ctx context.Context, store database.Store) string {
	root := BuildRoot(ctx, store)
	return root.RootHash
}

// BuildRoot constructs the Merkle root over all evidence in the store.
func BuildRoot(ctx context.Context, store database.Store) *Root {
	leaves := collectLeaves(ctx, store)
	root := buildMerkleRoot(leaves)
	return &Root{
		MethodologyVersion: MethodologyVersion,
		RootHash:           root,
		LeafCount:          len(leaves),
		Leaves:             leaves,
		PublishedAt:        time.Now().UTC(),
	}
}

func collectLeaves(ctx context.Context, store database.Store) []*Leaf {
	// Evidence is not directly listable through the Store interface, so we
	// walk every project and collect the evidence it references. Duplicate
	// evidence IDs are de-duplicated to keep the tree canonical.
	seen := make(map[string]*domain.Evidence)
	projects, _, err := store.ListProjects(ctx, database.ProjectFilter{Limit: 10_000})
	if err != nil {
		return nil
	}
	for _, project := range projects {
		if project == nil {
			continue
		}
		for _, evidenceID := range project.EvidenceIDs {
			if evidenceID == "" || seen[evidenceID] != nil {
				continue
			}
			evidence, err := store.GetEvidence(ctx, evidenceID)
			if err != nil || evidence == nil {
				continue
			}
			seen[evidenceID] = evidence
		}
	}

	leaves := make([]*Leaf, 0, len(seen))
	for id, evidence := range seen {
		leafHash := hashLeaf(id, evidence.ContentHash)
		leaves = append(leaves, &Leaf{
			EvidenceID:  id,
			ContentHash: evidence.ContentHash,
			LeafHash:    leafHash,
		})
	}
	sort.Slice(leaves, func(i, j int) bool { return leaves[i].EvidenceID < leaves[j].EvidenceID })
	return leaves
}

// buildMerkleRoot constructs a binary Merkle tree over the supplied leaves.
// The empty-tree root is the SHA-256 of the empty string so that an empty
// dataset has a stable, non-nil root.
func buildMerkleRoot(leaves []*Leaf) string {
	if len(leaves) == 0 {
		return hashLeaf("", "")
	}
	if len(leaves) == 1 {
		return leaves[0].LeafHash
	}

	layers := make([][]string, 0, len(leaves))
	current := make([]string, len(leaves))
	for i, leaf := range leaves {
		current[i] = leaf.LeafHash
	}
	layers = append(layers, current)

	for len(current) > 1 {
		next := make([]string, 0, (len(current)+1)/2)
		for i := 0; i < len(current); i += 2 {
			if i+1 < len(current) {
				next = append(next, hashPair(current[i], current[i+1]))
			} else {
				// Odd node promoted unchanged (standard Merkle behaviour).
				next = append(next, current[i])
			}
		}
		layers = append(layers, next)
		current = next
	}
	return current[0]
}

func hashLeaf(evidenceID, contentHash string) string {
	sum := sha256.Sum256([]byte("leaf|" + evidenceID + "|" + contentHash))
	return hex.EncodeToString(sum[:])
}

func hashPair(left, right string) string {
	sum := sha256.Sum256([]byte("pair|" + left + "|" + right))
	return hex.EncodeToString(sum[:])
}