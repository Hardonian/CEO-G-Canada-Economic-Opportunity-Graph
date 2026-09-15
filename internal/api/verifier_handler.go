package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/verifier"
)

// handleVerifierAttestations serves GET /api/v1/verifier/attestations
//
// Query parameters:
//   - milestone (string) — filter to a specific milestone ID
//   - limit     (int)    — max results (default 100, max 500)
func (s *Server) handleVerifierAttestations(w http.ResponseWriter, r *http.Request) {
	if s.attStore == nil {
		writeError(w, r, http.StatusServiceUnavailable, "verifier_unavailable", "Verifier attestation store is not initialised.")
		return
	}

	milestoneFilter := r.URL.Query().Get("milestone")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > 500 {
		limit = 500
	}

	var attestations []*verifier.Attestation
	if milestoneFilter != "" {
		attestations = s.attStore.GetByMilestone(milestoneFilter)
	} else {
		attestations = s.attStore.GetAll()
	}

	// Apply limit.
	if len(attestations) > limit {
		attestations = attestations[len(attestations)-limit:]
	}

	type attestationOut struct {
		NodeID       string    `json:"node_id"`
		MilestoneID  string    `json:"milestone_id"`
		ProjectID    string    `json:"project_id"`
		EvidenceHash string    `json:"evidence_hash"`
		ObservedAt   time.Time `json:"observed_at"`
		Signature    string    `json:"signature"`
	}
	out := make([]attestationOut, 0, len(attestations))
	for _, a := range attestations {
		out = append(out, attestationOut{
			NodeID:       string(a.NodeID),
			MilestoneID:  a.MilestoneID,
			ProjectID:    a.ProjectID,
			EvidenceHash: a.EvidenceHash,
			ObservedAt:   a.ObservedAt,
			Signature:    a.Signature,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"attestations": out,
		"total":        s.attStore.Count(),
		"returned":     len(out),
	})
}

// handleVerifierMilestone serves GET /api/v1/verifier/milestones/{id}
//
// Runs a quorum vote over all stored attestations for the given milestone
// and returns the NotarizedMilestone result.
func (s *Server) handleVerifierMilestone(w http.ResponseWriter, r *http.Request) {
	if s.attStore == nil || s.verifierNet == nil {
		writeError(w, r, http.StatusServiceUnavailable, "verifier_unavailable", "Verifier network is not initialised.")
		return
	}

	milestoneID := r.PathValue("id")
	if milestoneID == "" {
		writeError(w, r, http.StatusBadRequest, "missing_id", "Milestone ID is required.")
		return
	}

	attestations := s.attStore.GetByMilestone(milestoneID)
	if len(attestations) == 0 {
		writeError(w, r, http.StatusNotFound, "milestone_not_found",
			"No attestations found for milestone "+milestoneID)
		return
	}

	result := s.verifierNet.Vote(attestations)

	type notarizedOut struct {
		MilestoneID    string    `json:"milestone_id"`
		ProjectID      string    `json:"project_id"`
		EvidenceHash   string    `json:"evidence_hash"`
		QuorumReached  bool      `json:"quorum_reached"`
		QuorumFraction float64   `json:"quorum_fraction"`
		AttestationCnt int       `json:"attestation_count"`
		NotarizedAt    time.Time `json:"notarized_at"`
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(notarizedOut{
		MilestoneID:    result.MilestoneID,
		ProjectID:      result.ProjectID,
		EvidenceHash:   result.EvidenceHash,
		QuorumReached:  result.QuorumReached,
		QuorumFraction: result.QuorumFraction,
		AttestationCnt: len(result.Attestations),
		NotarizedAt:    result.NotarizedAt,
	})
}
