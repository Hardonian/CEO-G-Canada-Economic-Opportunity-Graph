package database

import (
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

// ensureSlugIndexLocked builds the slug -> project ID index while the store
// write lock is held.
func (m *MemoryStore) ensureSlugIndexLocked() {
	if m.slugIndex != nil {
		return
	}
	m.slugIndex = make(map[string]string, len(m.projects))
	for id, p := range m.projects {
		if p.Slug != "" {
			m.slugIndex[p.Slug] = id
		}
	}
}

// ensureEntityNameIndexLocked builds the normalized legal/common/alias name
// -> entity ID index while the store write lock is held.
func (m *MemoryStore) ensureEntityNameIndexLocked() {
	if m.entityNameIndex != nil {
		return
	}
	m.entityNameIndex = make(map[string]string, len(m.entities))
	for id, entity := range m.entities {
		for _, name := range entityNames(entity) {
			key := normalizedName(name)
			if key != "" {
				m.entityNameIndex[key] = id
			}
		}
	}
}

// ensureSignalIndexLocked builds the projectID -> signal IDs index. Safe to
// call with the write lock held. No-op if the index is already populated.
func (m *MemoryStore) ensureSignalIndexLocked() {
	if m.signalsByProject != nil {
		return
	}
	m.signalsByProject = make(map[string][]string, len(m.signals))
	for id, s := range m.signals {
		if s.ProjectID != "" {
			m.signalsByProject[s.ProjectID] = append(m.signalsByProject[s.ProjectID], id)
		}
	}
}

// ensureEventIndexLocked builds the projectID -> event IDs index. Safe to
// call with the write lock held. No-op if the index is already populated.
func (m *MemoryStore) ensureEventIndexLocked() {
	if m.eventsByProject != nil {
		return
	}
	m.eventsByProject = make(map[string][]string, len(m.events))
	for id, ev := range m.events {
		if ev.ProjectID != "" {
			m.eventsByProject[ev.ProjectID] = append(m.eventsByProject[ev.ProjectID], id)
		}
	}
}

// RebuildIndexes recomputes all indexes from the current maps.
// Call with the store write lock held.
func (m *MemoryStore) RebuildIndexes() {
	m.slugIndex = make(map[string]string, len(m.projects))
	m.entityNameIndex = make(map[string]string, len(m.entities))
	for id, project := range m.projects {
		if project.Slug != "" {
			m.slugIndex[project.Slug] = id
		}
	}
	for id, entity := range m.entities {
		for _, name := range entityNames(entity) {
			key := normalizedName(name)
			if key != "" {
				m.entityNameIndex[key] = id
			}
		}
	}
}

// RebuildSignalIndex rebuilds the signals-by-project secondary index from
// the current signals map. Call with the store write lock held.
func (m *MemoryStore) RebuildSignalIndex() {
	m.signalsByProject = make(map[string][]string, len(m.signals))
	for id, s := range m.signals {
		if s.ProjectID != "" {
			m.signalsByProject[s.ProjectID] = append(m.signalsByProject[s.ProjectID], id)
		}
	}
}

// RebuildEventIndex rebuilds the events-by-project secondary index from
// the current events map. Call with the store write lock held.
func (m *MemoryStore) RebuildEventIndex() {
	m.eventsByProject = make(map[string][]string, len(m.events))
	for id, ev := range m.events {
		if ev.ProjectID != "" {
			m.eventsByProject[ev.ProjectID] = append(m.eventsByProject[ev.ProjectID], id)
		}
	}
}

func normalizedName(value string) string {
	return domain.NormalizeLookupName(value)
}