package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

func (m *MemoryStore) UpsertPublisher(ctx context.Context, publisher *domain.Publisher) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if publisher == nil || strings.TrimSpace(publisher.ID) == "" || strings.TrimSpace(publisher.Name) == "" || !publisher.AuthorityTier.Valid() {
		return fmt.Errorf("publisher id, name, and authority tier are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := clonePublisher(publisher)
	if current, ok := m.publishers[copy.ID]; ok && copy.CreatedAt.IsZero() {
		copy.CreatedAt = current.CreatedAt
	}
	m.publishers[copy.ID] = copy
	return nil
}

func (m *MemoryStore) GetPublisher(ctx context.Context, id string) (*domain.Publisher, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	publisher, ok := m.publishers[id]
	if !ok {
		return nil, ErrNotFound
	}
	return clonePublisher(publisher), nil
}

func (m *MemoryStore) ListPublishers(ctx context.Context) ([]*domain.Publisher, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Publisher, 0, len(m.publishers))
	for _, publisher := range m.publishers {
		result = append(result, clonePublisher(publisher))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (m *MemoryStore) SavePublisherPolicy(ctx context.Context, policy *domain.PublisherPolicy) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if policy == nil || policy.PublisherID == "" || policy.MaxConcurrency < 0 || policy.RequestsPerMinute < 0 || policy.Burst < 0 || policy.MinimumPollIntervalSeconds < 0 {
		return fmt.Errorf("valid publisher policy is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.publishers[policy.PublisherID]; !ok {
		return ErrNotFound
	}
	m.publisherPolicies[policy.PublisherID] = clonePublisherPolicy(policy)
	return nil
}

func (m *MemoryStore) GetPublisherPolicy(ctx context.Context, publisherID string) (*domain.PublisherPolicy, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	policy, ok := m.publisherPolicies[publisherID]
	if !ok {
		return nil, ErrNotFound
	}
	return clonePublisherPolicy(policy), nil
}

func (m *MemoryStore) UpsertSource(ctx context.Context, source *domain.Source) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := validateSource(source); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.publishers[source.PublisherID]; !ok {
		return fmt.Errorf("publisher %q: %w", source.PublisherID, ErrNotFound)
	}
	if source.ParentSourceID != "" {
		if source.ParentSourceID == source.ID {
			return fmt.Errorf("source cannot be its own parent")
		}
		if _, ok := m.sources[source.ParentSourceID]; !ok {
			return fmt.Errorf("parent source %q: %w", source.ParentSourceID, ErrNotFound)
		}
	}
	urlKey, err := canonicalSourceURLKey(source.CanonicalURL)
	if err != nil {
		return err
	}
	if owner, ok := m.sourceURLIndex[urlKey]; ok && owner != source.ID {
		return fmt.Errorf("canonical URL is already owned by source %q: %w", owner, ErrSourceConflict)
	}
	copy := cloneSource(source)
	if current, ok := m.sources[copy.ID]; ok {
		oldKey, _ := canonicalSourceURLKey(current.CanonicalURL)
		if oldKey != urlKey {
			delete(m.sourceURLIndex, oldKey)
		}
		if copy.CreatedAt.IsZero() {
			copy.CreatedAt = current.CreatedAt
		}
		if copy.DiscoveredAt.IsZero() {
			copy.DiscoveredAt = current.DiscoveredAt
		}
	}
	m.sources[copy.ID] = copy
	m.sourceURLIndex[urlKey] = copy.ID
	return nil
}

func (m *MemoryStore) GetSource(ctx context.Context, id string) (*domain.Source, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	source, ok := m.sources[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSource(source), nil
}

func (m *MemoryStore) ListSources(ctx context.Context, filter domain.SourceFilter) ([]*domain.Source, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.Source, 0)
	for _, source := range m.sources {
		if !sourceMatchesFilter(source, filter) {
			continue
		}
		result = append(result, cloneSource(source))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) SaveSourcePrivateConfig(ctx context.Context, config *domain.SourcePrivateConfig) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if config == nil || config.SourceID == "" {
		return fmt.Errorf("source private config requires source id")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sources[config.SourceID]; !ok {
		return ErrNotFound
	}
	m.sourcePrivateConfigs[config.SourceID] = cloneSourcePrivateConfig(config)
	return nil
}

func (m *MemoryStore) GetSourcePrivateConfig(ctx context.Context, sourceID string) (*domain.SourcePrivateConfig, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	config, ok := m.sourcePrivateConfigs[sourceID]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSourcePrivateConfig(config), nil
}

func (m *MemoryStore) UpsertSourceCandidate(ctx context.Context, candidate *domain.SourceCandidate) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if candidate == nil || candidate.ID == "" || candidate.SourceID == "" || candidate.DiscoveryURL == "" || !candidate.LifecycleStatus.Valid() {
		return fmt.Errorf("candidate id, source id, discovery URL, and valid lifecycle are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sources[candidate.SourceID]; !ok {
		return fmt.Errorf("candidate source %q: %w", candidate.SourceID, ErrNotFound)
	}
	if candidate.DiscoveredBySourceID != "" {
		if _, ok := m.sources[candidate.DiscoveredBySourceID]; !ok {
			return fmt.Errorf("discovering source %q: %w", candidate.DiscoveredBySourceID, ErrNotFound)
		}
	}
	copy := cloneSourceCandidate(candidate)
	if current, ok := m.sourceCandidates[copy.ID]; ok && copy.CreatedAt.IsZero() {
		copy.CreatedAt = current.CreatedAt
	}
	m.sourceCandidates[copy.ID] = copy
	if source := m.sources[copy.SourceID]; source != nil {
		source.LifecycleStatus = copy.LifecycleStatus
		if copy.UpdatedAt.After(source.UpdatedAt) {
			source.UpdatedAt = copy.UpdatedAt
		}
	}
	return nil
}

func (m *MemoryStore) GetSourceCandidate(ctx context.Context, id string) (*domain.SourceCandidate, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	candidate, ok := m.sourceCandidates[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSourceCandidate(candidate), nil
}

func (m *MemoryStore) ListSourceCandidates(ctx context.Context, filter domain.SourceCandidateFilter) ([]*domain.SourceCandidate, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.SourceCandidate, 0)
	for _, candidate := range m.sourceCandidates {
		if filter.SourceID != "" && candidate.SourceID != filter.SourceID {
			continue
		}
		if filter.DiscoveredByID != "" && candidate.DiscoveredBySourceID != filter.DiscoveredByID {
			continue
		}
		if filter.LifecycleStatus != "" && candidate.LifecycleStatus != filter.LifecycleStatus {
			continue
		}
		if filter.ReviewRequired != nil && candidate.ReviewRequired != *filter.ReviewRequired {
			continue
		}
		result = append(result, cloneSourceCandidate(candidate))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) TransitionSourceCandidate(ctx context.Context, request CandidateTransitionRequest) (*domain.SourceCandidate, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	if request.CandidateID == "" || !request.To.Valid() || request.At.IsZero() {
		return nil, fmt.Errorf("candidate id, valid target lifecycle, and transition time are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	candidate, ok := m.sourceCandidates[request.CandidateID]
	if !ok {
		return nil, ErrNotFound
	}
	if !domain.CanTransitionSourceLifecycle(candidate.LifecycleStatus, request.To) {
		return nil, fmt.Errorf("candidate %s cannot transition from %s to %s: %w", candidate.ID, candidate.LifecycleStatus, request.To, ErrInvalidTransition)
	}
	candidate.LifecycleStatus = request.To
	candidate.UpdatedAt = request.At.UTC()
	if request.ReviewedBy != "" {
		candidate.ReviewedBy = request.ReviewedBy
		reviewedAt := request.At.UTC()
		candidate.ReviewedAt = &reviewedAt
	}
	if request.Reason != "" {
		candidate.DecisionReason = request.Reason
	}
	if source := m.sources[candidate.SourceID]; source != nil {
		source.LifecycleStatus = request.To
		source.UpdatedAt = request.At.UTC()
	}
	return cloneSourceCandidate(candidate), nil
}

func (m *MemoryStore) SaveSourceRelationship(ctx context.Context, relationship *domain.SourceRelationship) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if relationship == nil || relationship.ID == "" || relationship.FromSourceID == "" || relationship.ToSourceID == "" || relationship.RelationshipType == "" {
		return fmt.Errorf("source relationship id, endpoints, and type are required")
	}
	if relationship.FromSourceID == relationship.ToSourceID {
		return fmt.Errorf("source relationship cannot be self-referential")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sources[relationship.FromSourceID]; !ok {
		return fmt.Errorf("from source %q: %w", relationship.FromSourceID, ErrNotFound)
	}
	if _, ok := m.sources[relationship.ToSourceID]; !ok {
		return fmt.Errorf("to source %q: %w", relationship.ToSourceID, ErrNotFound)
	}
	key := sourceRelationshipKey(relationship)
	if owner, ok := m.sourceRelationshipIndex[key]; ok && owner != relationship.ID {
		return fmt.Errorf("source relationship already exists as %q: %w", owner, ErrSourceConflict)
	}
	if existing, ok := m.sourceRelationships[relationship.ID]; ok {
		existingKey := sourceRelationshipKey(existing)
		if existingKey != key {
			return fmt.Errorf("immutable relationship id reused: %w", ErrSourceConflict)
		}
		return nil
	}
	m.sourceRelationships[relationship.ID] = cloneSourceRelationship(relationship)
	m.sourceRelationshipIndex[key] = relationship.ID
	return nil
}

func (m *MemoryStore) ListSourceRelationships(ctx context.Context, filter SourceRelationshipFilter) ([]*domain.SourceRelationship, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.SourceRelationship, 0)
	for _, relationship := range m.sourceRelationships {
		if filter.SourceID != "" && relationship.FromSourceID != filter.SourceID && relationship.ToSourceID != filter.SourceID {
			continue
		}
		if filter.FromSourceID != "" && relationship.FromSourceID != filter.FromSourceID {
			continue
		}
		if filter.ToSourceID != "" && relationship.ToSourceID != filter.ToSourceID {
			continue
		}
		if filter.RelationshipType != "" && relationship.RelationshipType != filter.RelationshipType {
			continue
		}
		result = append(result, cloneSourceRelationship(relationship))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) SaveSourceVersion(ctx context.Context, version *domain.SourceVersion) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := validateSourceVersion(version); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveSourceVersionLocked(version)
}

func (m *MemoryStore) saveSourceVersionLocked(version *domain.SourceVersion) error {
	if _, ok := m.sources[version.SourceID]; !ok {
		return fmt.Errorf("source %q: %w", version.SourceID, ErrNotFound)
	}
	if _, ok := m.sourceVersions[version.ID]; ok {
		return nil
	}
	hashKey := sourceVersionHashKey(version.SourceID, version.ContentHash)
	if owner, ok := m.sourceVersionHashIndex[hashKey]; ok && owner != version.ID {
		return fmt.Errorf("source content version already exists as %q: %w", owner, ErrSourceConflict)
	}
	for _, existing := range m.sourceVersions {
		if existing.SourceID == version.SourceID && existing.Sequence == version.Sequence && existing.ID != version.ID {
			return fmt.Errorf("source sequence %d already exists: %w", version.Sequence, ErrSourceConflict)
		}
	}
	copy := cloneSourceVersion(version)
	m.sourceVersions[copy.ID] = copy
	m.sourceVersionHashIndex[hashKey] = copy.ID
	source := m.sources[copy.SourceID]
	if source.LastChangeAt == nil || copy.ObservedAt.After(*source.LastChangeAt) {
		changedAt := copy.ObservedAt.UTC()
		source.LastChangeAt = &changedAt
	}
	if copy.CreatedAt.After(source.UpdatedAt) {
		source.UpdatedAt = copy.CreatedAt
	}
	return nil
}

func (m *MemoryStore) GetSourceVersion(ctx context.Context, id string) (*domain.SourceVersion, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	version, ok := m.sourceVersions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSourceVersion(version), nil
}

func (m *MemoryStore) ListSourceVersions(ctx context.Context, filter SourceVersionFilter) ([]*domain.SourceVersion, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.SourceVersion, 0)
	for _, version := range m.sourceVersions {
		if filter.SourceID != "" && version.SourceID != filter.SourceID {
			continue
		}
		if filter.LastKnownGood != nil && version.LastKnownGood != *filter.LastKnownGood {
			continue
		}
		result = append(result, cloneSourceVersion(version))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Sequence != result[j].Sequence {
			return result[i].Sequence < result[j].Sequence
		}
		return result[i].ID < result[j].ID
	})
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) GetLastKnownGoodSourceVersion(ctx context.Context, sourceID string) (*domain.SourceVersion, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var latest *domain.SourceVersion
	for _, version := range m.sourceVersions {
		if version.SourceID != sourceID || !version.LastKnownGood {
			continue
		}
		if latest == nil || version.Sequence > latest.Sequence ||
			(version.Sequence == latest.Sequence && version.ObservedAt.After(latest.ObservedAt)) ||
			(version.Sequence == latest.Sequence && version.ObservedAt.Equal(latest.ObservedAt) && version.ID > latest.ID) {
			latest = version
		}
	}
	if latest == nil {
		return nil, ErrNotFound
	}
	return cloneSourceVersion(latest), nil
}

func (m *MemoryStore) SaveSourceChange(ctx context.Context, change *domain.SourceChange) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := validateSourceChange(change); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveSourceChangeLocked(change)
}

func (m *MemoryStore) saveSourceChangeLocked(change *domain.SourceChange) error {
	if _, ok := m.sources[change.SourceID]; !ok {
		return fmt.Errorf("source %q: %w", change.SourceID, ErrNotFound)
	}
	toVersion, ok := m.sourceVersions[change.ToVersionID]
	if !ok || toVersion.SourceID != change.SourceID {
		return fmt.Errorf("to version %q: %w", change.ToVersionID, ErrNotFound)
	}
	if change.FromVersionID != "" {
		fromVersion, ok := m.sourceVersions[change.FromVersionID]
		if !ok || fromVersion.SourceID != change.SourceID {
			return fmt.Errorf("from version %q: %w", change.FromVersionID, ErrNotFound)
		}
	}
	if _, ok := m.sourceChanges[change.ID]; ok {
		return nil
	}
	if owner, ok := m.sourceChangeIndex[change.Fingerprint]; ok && owner != change.ID {
		return fmt.Errorf("source change fingerprint already exists as %q: %w", owner, ErrSourceConflict)
	}
	copy := cloneSourceChange(change)
	m.sourceChanges[copy.ID] = copy
	m.sourceChangeIndex[copy.Fingerprint] = copy.ID
	source := m.sources[copy.SourceID]
	if source.LastChangeAt == nil || copy.ObservedAt.After(*source.LastChangeAt) {
		changedAt := copy.ObservedAt.UTC()
		source.LastChangeAt = &changedAt
	}
	return nil
}

func (m *MemoryStore) ListSourceChanges(ctx context.Context, filter SourceChangeFilter) ([]*domain.SourceChange, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.SourceChange, 0)
	for _, change := range m.sourceChanges {
		if filter.SourceID != "" && change.SourceID != filter.SourceID {
			continue
		}
		if filter.ChangeType != "" && change.ChangeType != filter.ChangeType {
			continue
		}
		if filter.Materiality != "" && change.Materiality != filter.Materiality {
			continue
		}
		if filter.Since != nil && change.ObservedAt.Before(*filter.Since) {
			continue
		}
		result = append(result, cloneSourceChange(change))
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].ObservedAt.Equal(result[j].ObservedAt) {
			return result[i].ObservedAt.Before(result[j].ObservedAt)
		}
		return result[i].ID < result[j].ID
	})
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) SaveSourceHealthCheck(ctx context.Context, check *domain.SourceHealthCheck) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := validateSourceHealthCheck(check); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveSourceHealthCheckLocked(check)
}

func (m *MemoryStore) saveSourceHealthCheckLocked(check *domain.SourceHealthCheck) error {
	if _, ok := m.sources[check.SourceID]; !ok {
		return fmt.Errorf("source %q: %w", check.SourceID, ErrNotFound)
	}
	if _, ok := m.sourceHealthChecks[check.ID]; ok {
		return nil
	}
	copy := cloneSourceHealthCheck(check)
	m.sourceHealthChecks[copy.ID] = copy
	source := m.sources[copy.SourceID]
	checkedAt := copy.CheckedAt.UTC()
	if source.LastCheckedAt == nil || checkedAt.After(*source.LastCheckedAt) {
		source.LastCheckedAt = &checkedAt
		source.HealthStatus = copy.Status
		if copy.Status == domain.SourceHealthHealthy {
			successAt := checkedAt
			source.LastSuccessAt = &successAt
		}
	}
	return nil
}

func (m *MemoryStore) ListSourceHealthChecks(ctx context.Context, filter SourceHealthFilter) ([]*domain.SourceHealthCheck, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.SourceHealthCheck, 0)
	for _, check := range m.sourceHealthChecks {
		if filter.SourceID != "" && check.SourceID != filter.SourceID {
			continue
		}
		if filter.Status != "" && check.Status != filter.Status {
			continue
		}
		if filter.Since != nil && check.CheckedAt.Before(*filter.Since) {
			continue
		}
		result = append(result, cloneSourceHealthCheck(check))
	}
	sort.Slice(result, func(i, j int) bool {
		if !result[i].CheckedAt.Equal(result[j].CheckedAt) {
			return result[i].CheckedAt.Before(result[j].CheckedAt)
		}
		return result[i].ID < result[j].ID
	})
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) GetLatestSourceHealthCheck(ctx context.Context, sourceID string) (*domain.SourceHealthCheck, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	var latest *domain.SourceHealthCheck
	for _, check := range m.sourceHealthChecks {
		if check.SourceID != sourceID {
			continue
		}
		if latest == nil || check.CheckedAt.After(latest.CheckedAt) ||
			(check.CheckedAt.Equal(latest.CheckedAt) && check.ID > latest.ID) {
			latest = check
		}
	}
	if latest == nil {
		return nil, ErrNotFound
	}
	return cloneSourceHealthCheck(latest), nil
}

func (m *MemoryStore) SaveSourceCheckpoint(ctx context.Context, checkpoint *domain.SourceCheckpoint) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := validateSourceCheckpoint(checkpoint); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveSourceCheckpointLocked(checkpoint)
}

func (m *MemoryStore) saveSourceCheckpointLocked(checkpoint *domain.SourceCheckpoint) error {
	if _, ok := m.sources[checkpoint.SourceID]; !ok {
		return fmt.Errorf("source %q: %w", checkpoint.SourceID, ErrNotFound)
	}
	if checkpoint.LastVersionID != "" {
		version, ok := m.sourceVersions[checkpoint.LastVersionID]
		if !ok || version.SourceID != checkpoint.SourceID {
			return fmt.Errorf("last source version %q: %w", checkpoint.LastVersionID, ErrNotFound)
		}
	}
	key := sourceCheckpointKey(checkpoint.SourceID, checkpoint.StreamKey)
	if current, ok := m.sourceCheckpoints[key]; ok {
		if checkpoint.Revision < current.Revision {
			return fmt.Errorf("checkpoint revision moved backwards: %w", ErrSourceConflict)
		}
		if checkpoint.Revision == current.Revision {
			if reflect.DeepEqual(current, checkpoint) {
				return nil
			}
			return fmt.Errorf("checkpoint revision %d already has different state: %w", checkpoint.Revision, ErrSourceConflict)
		}
	}
	m.sourceCheckpoints[key] = cloneSourceCheckpoint(checkpoint)
	return nil
}

func (m *MemoryStore) GetSourceCheckpoint(ctx context.Context, sourceID, streamKey string) (*domain.SourceCheckpoint, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	checkpoint, ok := m.sourceCheckpoints[sourceCheckpointKey(sourceID, streamKey)]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneSourceCheckpoint(checkpoint), nil
}

func (m *MemoryStore) SaveMappingVersion(ctx context.Context, mapping *domain.MappingVersion) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if mapping == nil || mapping.ID == "" || mapping.SourceID == "" || mapping.Name == "" || mapping.Version == "" || mapping.DefinitionHash == "" || mapping.Status == "" {
		return fmt.Errorf("mapping id, source, name, version, definition hash, and status are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.sources[mapping.SourceID]; !ok {
		return fmt.Errorf("source %q: %w", mapping.SourceID, ErrNotFound)
	}
	key := mappingVersionKey(mapping.SourceID, mapping.Name, mapping.Version)
	if owner, ok := m.mappingVersionIndex[key]; ok && owner != mapping.ID {
		return fmt.Errorf("mapping version already exists as %q: %w", owner, ErrSourceConflict)
	}
	if mapping.Status == domain.MappingActive {
		for _, current := range m.mappingVersions {
			if current.SourceID == mapping.SourceID && current.Name == mapping.Name && current.Status == domain.MappingActive && current.ID != mapping.ID {
				return fmt.Errorf("mapping %q already has active version %q: %w", mapping.Name, current.ID, ErrSourceConflict)
			}
		}
	}
	copy := cloneMappingVersion(mapping)
	if current, ok := m.mappingVersions[copy.ID]; ok && copy.CreatedAt.IsZero() {
		copy.CreatedAt = current.CreatedAt
	}
	m.mappingVersions[copy.ID] = copy
	m.mappingVersionIndex[key] = copy.ID
	return nil
}

func (m *MemoryStore) GetMappingVersion(ctx context.Context, id string) (*domain.MappingVersion, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	mapping, ok := m.mappingVersions[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneMappingVersion(mapping), nil
}

func (m *MemoryStore) ListMappingVersions(ctx context.Context, filter MappingVersionFilter) ([]*domain.MappingVersion, int, error) {
	if err := contextError(ctx); err != nil {
		return nil, 0, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*domain.MappingVersion, 0)
	for _, mapping := range m.mappingVersions {
		if filter.SourceID != "" && mapping.SourceID != filter.SourceID {
			continue
		}
		if filter.Name != "" && mapping.Name != filter.Name {
			continue
		}
		if filter.Status != "" && mapping.Status != filter.Status {
			continue
		}
		result = append(result, cloneMappingVersion(mapping))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		if result[i].Version != result[j].Version {
			return result[i].Version < result[j].Version
		}
		return result[i].ID < result[j].ID
	})
	total := len(result)
	start, end := pageBounds(total, filter.Offset, filter.Limit)
	return result[start:end], total, nil
}

func (m *MemoryStore) GetActiveMappingVersion(ctx context.Context, sourceID, name string) (*domain.MappingVersion, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, mapping := range m.mappingVersions {
		if mapping.SourceID == sourceID && mapping.Name == name && mapping.Status == domain.MappingActive {
			return cloneMappingVersion(mapping), nil
		}
	}
	return nil, ErrNotFound
}
