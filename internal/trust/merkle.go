package trust

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

const (
	MerkleMethodologyVersion = "merkle-proof-v1.0"
	DefaultMerkleLeafLimit   = 10000
)

// MerkleLeaf represents a single evidence item in the Merkle tree.
type MerkleLeaf struct {
	Index       int       `json:"index"`
	EvidenceID  string    `json:"evidence_id"`
	ContentHash string    `json:"content_hash"`
	SourceURL   string    `json:"source_url,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// MerkleNode represents a node in the Merkle tree.
type MerkleNode struct {
	Hash        string `json:"hash"`
	LeftHash    string `json:"left_hash,omitempty"`
	RightHash   string `json:"right_hash,omitempty"`
	LeafIndex   *int   `json:"leaf_index,omitempty"`
	IsLeaf      bool   `json:"is_leaf"`
	ParentHash  string `json:"parent_hash,omitempty"`
}

// MerkleTree represents a complete Merkle tree of evidence hashes.
type MerkleTree struct {
	RootHash    string        `json:"root_hash"`
	LeafCount   int           `json:"leaf_count"`
	LevelCount  int           `json:"level_count"`
	GeneratedAt time.Time     `json:"generated_at"`
	Leaves      []MerkleLeaf  `json:"leaves"`
	Levels      [][]MerkleNode `json:"levels"`
}

// MerkleProof represents a cryptographic proof for a specific evidence item.
type MerkleProof struct {
	MethodologyVersion string      `json:"methodology_version"`
	RootHash           string      `json:"root_hash"`
	Leaf               MerkleLeaf  `json:"leaf"`
	Path               []MerkleNode `json:"path"`
	Verified           bool        `json:"verified"`
	VerifiedAt         time.Time   `json:"verified_at"`
}

// NewMerkleTree constructs a Merkle tree from evidence hashes.
func NewMerkleTree(leaves []MerkleLeaf) *MerkleTree {
	if len(leaves) == 0 {
		leaves = []MerkleLeaf{}
	}
	
	// Sort by evidence ID for deterministic ordering
	sortedLeaves := make([]MerkleLeaf, len(leaves))
	copy(sortedLeaves, leaves)
	sort.Slice(sortedLeaves, func(i, j int) bool {
		return sortedLeaves[i].EvidenceID < sortedLeaves[j].EvidenceID
	})
	
	levels := make([][]MerkleNode, 0)
	currentLevel := make([]MerkleNode, 0, len(sortedLeaves))
	
	for i, leaf := range sortedLeaves {
		hashInput := fmt.Sprintf("%s:%s:%d", leaf.EvidenceID, leaf.ContentHash, leaf.Index)
		hash := hashString(hashInput)
		index := i
		currentLevel = append(currentLevel, MerkleNode{
			Hash:      hash,
			LeafIndex: &index,
			IsLeaf:    true,
		})
	}
	
	levels = append(levels, currentLevel)
	
	for len(currentLevel) > 1 {
		nextLevel := make([]MerkleNode, 0, (len(currentLevel)+1)/2)
		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			right := currentLevel[i+1]
			if i+1 >= len(currentLevel) {
				right = left
			}
			combined := left.Hash + right.Hash
			hash := hashString(combined)
			nextLevel = append(nextLevel, MerkleNode{
				Hash:       hash,
				LeftHash:   left.Hash,
				RightHash:  right.Hash,
				ParentHash: "",
			})
		}
		levels = append(levels, nextLevel)
		currentLevel = nextLevel
	}
	
	rootHash := ""
	if len(levels) > 0 && len(levels[len(levels)-1]) > 0 {
		rootHash = levels[len(levels)-1][0].Hash
	}
	
	return &MerkleTree{
		RootHash:    rootHash,
		LeafCount:   len(sortedLeaves),
		LevelCount:  len(levels),
		GeneratedAt: time.Now().UTC(),
		Leaves:      sortedLeaves,
		Levels:      levels,
	}
}

// GenerateProof creates a Merkle proof for a specific evidence item.
func (m *MerkleTree) GenerateProof(evidenceID string) (*MerkleProof, error) {
	if m == nil || len(m.Leaves) == 0 {
		return nil, fmt.Errorf("merkle tree is empty")
	}
	
	leafIndex := -1
	var targetLeaf MerkleLeaf
	for i, leaf := range m.Leaves {
		if leaf.EvidenceID == evidenceID {
			leafIndex = i
			targetLeaf = leaf
			break
		}
	}
	
	if leafIndex == -1 {
		return nil, fmt.Errorf("evidence %q not found in merkle tree", evidenceID)
	}
	
	path := make([]MerkleNode, 0)
	currentIndex := leafIndex
	for level := 0; level < len(m.Levels)-1; level++ {
		currentLevel := m.Levels[level]
		if currentIndex >= len(currentLevel) {
			return nil, fmt.Errorf("invalid merkle tree structure")
		}
		
		siblingIndex := currentIndex
		if currentIndex%2 == 0 {
			if currentIndex+1 < len(currentLevel) {
				siblingIndex = currentIndex + 1
			}
		} else {
			siblingIndex = currentIndex - 1
		}
		
		if siblingIndex >= 0 && siblingIndex < len(currentLevel) {
			path = append(path, currentLevel[siblingIndex])
		}
		currentIndex = currentIndex / 2
	}
	
	return &MerkleProof{
		MethodologyVersion: MerkleMethodologyVersion,
		RootHash:           m.RootHash,
		Leaf:               targetLeaf,
		Path:               path,
		Verified:           false,
		VerifiedAt:         time.Now().UTC(),
	}, nil
}

// VerifyProof validates a Merkle proof against a root hash.
func VerifyProof(proof *MerkleProof) bool {
	if proof == nil || proof.RootHash == "" || proof.Leaf.ContentHash == "" {
		return false
	}
	
	currentHash := hashString(fmt.Sprintf("%s:%s:%d", proof.Leaf.EvidenceID, proof.Leaf.ContentHash, proof.Leaf.Index))
	for _, node := range proof.Path {
		if node.LeftHash != "" && node.RightHash != "" {
			if node.LeftHash == currentHash {
				currentHash = hashString(currentHash + node.RightHash)
			} else if node.RightHash == currentHash {
				currentHash = hashString(node.LeftHash + currentHash)
			} else {
				return false
			}
		} else if node.Hash != "" && node.Hash != currentHash {
			currentHash = hashString(currentHash + node.Hash)
		}
	}
	
	return currentHash == proof.RootHash
}

// BuildMerkleTreeFromEvidence creates a Merkle tree from evidence records.
func BuildMerkleTreeFromEvidence(evidence []*domain.Evidence) *MerkleTree {
	leaves := make([]MerkleLeaf, 0, len(evidence))
	for i, ev := range evidence {
		leaves = append(leaves, MerkleLeaf{
			Index:       i,
			EvidenceID:  ev.ID,
			ContentHash: ev.ContentHash,
			SourceURL:   ev.SourceURL,
			Timestamp:   ev.RetrievalTimestamp,
		})
	}
	return NewMerkleTree(leaves)
}

// PublishMerkleRoot simulates publishing a daily Merkle root to an append-only log.
func PublishMerkleRoot(tree *MerkleTree) string {
	if tree == nil {
		return ""
	}
	
	payload := fmt.Sprintf("%s:%d:%s", 
		tree.RootHash, 
		tree.LeafCount, 
		tree.GeneratedAt.Format(time.RFC3339Nano))
	
	return hashString(payload)
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
