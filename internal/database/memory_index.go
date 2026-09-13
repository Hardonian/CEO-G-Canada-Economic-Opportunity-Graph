package database

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

// RebuildIndexes recomputes both indexes from the current project and entity
// maps. Call with the store write lock held.
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

func entityNames(entity *domain.Entity) []string {
	if entity == nil {
		return nil
	}
	names := make([]string, 0, 2+len(entity.Aliases))
	names = append(names, entity.LegalName, entity.CommonName)
	names = append(names, entity.Aliases...)
	return names
}

func normalizedName(value string) string {
	return domain.NormalizeLookupName(value)
}