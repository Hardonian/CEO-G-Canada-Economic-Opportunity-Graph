// Package sources contains Public Data Mesh control-plane services that are
// independent of any particular catalog protocol or graph projection.
package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/database"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const (
	ConfigSchemaVersion = "1"
	maxConfigBytes      = 2 << 20
)

// Configuration is a versionable, non-secret set of reviewed discovery roots.
// Dynamically discovered sources belong in SourceStore and are not written
// back to these files automatically.
type Configuration struct {
	SchemaVersion     string                        `json:"schema_version"`
	Publishers        []*domain.Publisher           `json:"publishers"`
	PublisherPolicies []*domain.PublisherPolicy     `json:"publisher_policies,omitempty"`
	Sources           []*domain.Source              `json:"sources"`
	PrivateConfigs    []*domain.SourcePrivateConfig `json:"private_configs,omitempty"`
}

func DecodeConfiguration(reader io.Reader) (*Configuration, error) {
	limited := io.LimitReader(reader, maxConfigBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read source configuration: %w", err)
	}
	if len(data) > maxConfigBytes {
		return nil, fmt.Errorf("source configuration exceeds %d bytes", maxConfigBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var config Configuration
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("decode source configuration: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, fmt.Errorf("source configuration contains multiple JSON values")
	}
	if config.SchemaVersion != ConfigSchemaVersion {
		return nil, fmt.Errorf("source configuration schema_version %q is unsupported", config.SchemaVersion)
	}
	return &config, nil
}

// LoadDirectory reads every .json declaration beneath root in stable path
// order. A single malformed declaration fails the import before store writes.
func LoadDirectory(root string) (*Configuration, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(path), ".json") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk source configuration directory: %w", err)
	}
	sort.Strings(paths)
	combined := &Configuration{SchemaVersion: ConfigSchemaVersion}
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		config, decodeErr := DecodeConfiguration(file)
		closeErr := file.Close()
		if decodeErr != nil {
			return nil, fmt.Errorf("%s: %w", path, decodeErr)
		}
		if closeErr != nil {
			return nil, closeErr
		}
		combined.Publishers = append(combined.Publishers, config.Publishers...)
		combined.PublisherPolicies = append(combined.PublisherPolicies, config.PublisherPolicies...)
		combined.Sources = append(combined.Sources, config.Sources...)
		combined.PrivateConfigs = append(combined.PrivateConfigs, config.PrivateConfigs...)
	}
	return combined, nil
}

// ApplyConfiguration imports publishers before their policies and then sources
// in hierarchy order. importedAt is injected to keep fixtures deterministic.
func ApplyConfiguration(ctx context.Context, store database.SourceStore, config *Configuration, importedAt time.Time) error {
	if store == nil || config == nil {
		return fmt.Errorf("source store and configuration are required")
	}
	if importedAt.IsZero() {
		importedAt = time.Now().UTC()
	}
	for _, publisher := range config.Publishers {
		setPublisherTimes(publisher, importedAt)
		if err := store.UpsertPublisher(ctx, publisher); err != nil {
			return fmt.Errorf("upsert publisher %q: %w", publisher.ID, err)
		}
	}
	for _, policy := range config.PublisherPolicies {
		setPolicyTimes(policy, importedAt)
		if err := store.SavePublisherPolicy(ctx, policy); err != nil {
			return fmt.Errorf("save publisher policy %q: %w", policy.PublisherID, err)
		}
	}

	pending := append([]*domain.Source(nil), config.Sources...)
	known := make(map[string]struct{}, len(pending))
	for len(pending) > 0 {
		progress := false
		next := pending[:0]
		for _, source := range pending {
			if source.ParentSourceID != "" {
				if _, ok := known[source.ParentSourceID]; !ok {
					if _, err := store.GetSource(ctx, source.ParentSourceID); err != nil {
						next = append(next, source)
						continue
					}
				}
			}
			setSourceTimes(source, importedAt)
			if err := store.UpsertSource(ctx, source); err != nil {
				return fmt.Errorf("upsert source %q: %w", source.ID, err)
			}
			known[source.ID] = struct{}{}
			progress = true
		}
		if !progress {
			ids := make([]string, 0, len(next))
			for _, source := range next {
				ids = append(ids, source.ID)
			}
			sort.Strings(ids)
			return fmt.Errorf("source hierarchy has missing parents or a cycle: %s", strings.Join(ids, ", "))
		}
		pending = next
	}
	for _, privateConfig := range config.PrivateConfigs {
		setPrivateConfigTimes(privateConfig, importedAt)
		if err := store.SaveSourcePrivateConfig(ctx, privateConfig); err != nil {
			return fmt.Errorf("save private source config %q: %w", privateConfig.SourceID, err)
		}
	}
	return nil
}

func setPublisherTimes(value *domain.Publisher, now time.Time) {
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = now
	}
}

func setPolicyTimes(value *domain.PublisherPolicy, now time.Time) {
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = now
	}
}

func setSourceTimes(value *domain.Source, now time.Time) {
	if value.DiscoveredAt.IsZero() {
		value.DiscoveredAt = now
	}
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = now
	}
}

func setPrivateConfigTimes(value *domain.SourcePrivateConfig, now time.Time) {
	if value.CreatedAt.IsZero() {
		value.CreatedAt = now
	}
	if value.UpdatedAt.IsZero() {
		value.UpdatedAt = now
	}
}
