package verifier

import (
	"sync"
	"time"
)

// AttestationStore is a thread-safe, in-memory store for attestations.
// It provides fast lookup by milestone ID and supports TTL-based pruning
// to bound memory growth in long-running processes.
type AttestationStore struct {
	mu          sync.RWMutex
	byMilestone map[string][]*Attestation // milestoneID → attestations
	all         []*Attestation            // insertion-ordered full list
}

// NewAttestationStore constructs an empty AttestationStore.
func NewAttestationStore() *AttestationStore {
	return &AttestationStore{
		byMilestone: make(map[string][]*Attestation),
	}
}

// Save records an attestation. Duplicate (same NodeID + MilestoneID) entries
// are appended — deduplication is the caller's responsibility.
func (s *AttestationStore) Save(at *Attestation) {
	if at == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byMilestone[at.MilestoneID] = append(s.byMilestone[at.MilestoneID], at)
	s.all = append(s.all, at)
}

// GetByMilestone returns all attestations for the given milestone ID.
// Returns an empty (not nil) slice when none are found.
func (s *AttestationStore) GetByMilestone(milestoneID string) []*Attestation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.byMilestone[milestoneID]
	out := make([]*Attestation, len(src))
	copy(out, src)
	return out
}

// GetAll returns all stored attestations in insertion order.
func (s *AttestationStore) GetAll() []*Attestation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Attestation, len(s.all))
	copy(out, s.all)
	return out
}

// Count returns the total number of stored attestations.
func (s *AttestationStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.all)
}

// Prune removes attestations observed before `before` and returns the count
// of removed entries. This operation is O(N) and should be called
// infrequently (e.g. on a maintenance timer).
func (s *AttestationStore) Prune(before time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	removed := 0
	kept := s.all[:0]
	for _, at := range s.all {
		if at.ObservedAt.Before(before) {
			removed++
		} else {
			kept = append(kept, at)
		}
	}
	s.all = kept

	// Rebuild milestone index from kept entries.
	s.byMilestone = make(map[string][]*Attestation, len(s.byMilestone))
	for _, at := range kept {
		s.byMilestone[at.MilestoneID] = append(s.byMilestone[at.MilestoneID], at)
	}
	return removed
}

// MilestoneIDs returns a deduplicated list of all milestone IDs that have
// at least one attestation in the store.
func (s *AttestationStore) MilestoneIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.byMilestone))
	for id := range s.byMilestone {
		ids = append(ids, id)
	}
	return ids
}
