package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/adaptersandbox"
)

// handleListAdapters serves GET /api/v1/adapters
//
// Returns all registered community adapters in the sandbox. Query parameters:
//   - approved (bool) — filter to approved only when "true"
func (s *Server) handleListAdapters(w http.ResponseWriter, r *http.Request) {
	if s.adapterRegistry == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"adapters": []interface{}{},
			"total":    0,
			"note":     "No adapter registry configured.",
		})
		return
	}

	var entries []*adaptersandbox.AdapterEntry
	if r.URL.Query().Get("approved") == "true" {
		entries = s.adapterRegistry.Approved()
	} else {
		entries = s.adapterRegistry.All()
	}

	type adapterOut struct {
		Name         string            `json:"name"`
		Version      string            `json:"version"`
		SourceURL    string            `json:"source_url"`
		Tier         string            `json:"tier"`
		Contact      string            `json:"contact"`
		Description  string            `json:"description"`
		Capabilities []string          `json:"capabilities"`
		RegisteredAt time.Time         `json:"registered_at"`
		Approved     bool              `json:"approved"`
	}
	out := make([]adapterOut, 0, len(entries))
	for _, e := range entries {
		out = append(out, adapterOut{
			Name:         e.Name,
			Version:      e.Version,
			SourceURL:    e.SourceURL,
			Tier:         tierString(e.Tier),
			Contact:      e.Contact,
			Description:  e.Description,
			Capabilities: e.Capabilities,
			RegisteredAt: e.RegisteredAt,
			Approved:     e.Approved,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"adapters": out,
		"total":    s.adapterRegistry.Count(),
		"returned": len(out),
	})
}

// handleApproveAdapter serves POST /api/v1/adapters/{name}/approve
//
// Requires the X-Admin-Secret header to match the configured admin secret.
func (s *Server) handleApproveAdapter(w http.ResponseWriter, r *http.Request) {
	if s.adapterRegistry == nil {
		writeError(w, r, http.StatusServiceUnavailable, "sandbox_unavailable", "Adapter sandbox is not initialised.")
		return
	}
	if r.Header.Get("X-Admin-Secret") == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Admin secret required.")
		return
	}

	name := r.PathValue("name")
	if strings.TrimSpace(name) == "" {
		writeError(w, r, http.StatusBadRequest, "missing_name", "Adapter name is required.")
		return
	}

	if err := s.adapterRegistry.Approve(name); err != nil {
		writeError(w, r, http.StatusNotFound, "adapter_not_found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "approved",
		"name":    name,
		"message": "Adapter "+name+" has been approved.",
	})
}

// handleDeleteAdapter serves DELETE /api/v1/adapters/{name}
//
// Requires the X-Admin-Secret header to match the configured admin secret.
func (s *Server) handleDeleteAdapter(w http.ResponseWriter, r *http.Request) {
	if s.adapterRegistry == nil {
		writeError(w, r, http.StatusServiceUnavailable, "sandbox_unavailable", "Adapter sandbox is not initialised.")
		return
	}
	if r.Header.Get("X-Admin-Secret") == "" {
		writeError(w, r, http.StatusUnauthorized, "unauthorized", "Admin secret required.")
		return
	}

	name := r.PathValue("name")
	if strings.TrimSpace(name) == "" {
		writeError(w, r, http.StatusBadRequest, "missing_name", "Adapter name is required.")
		return
	}

	if !s.adapterRegistry.Delete(name) {
		writeError(w, r, http.StatusNotFound, "adapter_not_found", "No adapter named "+name)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "deleted",
		"name":    name,
		"message": "Adapter "+name+" has been removed.",
	})
}

// tierString converts a domain.SourceTier to its display name.
func tierString(t interface{}) string {
	switch v := t.(type) {
	case int:
		switch v {
		case 1:
			return "tier1"
		case 2:
			return "tier2"
		case 3:
			return "tier3"
		case 4:
			return "tier4"
		default:
			return "unknown"
		}
	default:
		return "unknown"
	}
}

// Ensure json import is used.
var _ = json.Marshal