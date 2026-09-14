// Package releasecheck verifies the integrity and referential consistency of
// checked-in public and CEGS release artifacts.
package releasecheck

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/cegs"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

var expectedArtifacts = []string{
	"cegs/events.jsonl",
	"cegs/evidence.jsonl",
	"cegs/organizations.jsonl",
	"cegs/projects.jsonl",
	"public/entities.jsonl",
	"public/events.jsonl",
	"public/evidence.jsonl",
	"public/projects.csv",
	"public/projects.geojson",
	"public/projects.jsonl",
	"public/scores.jsonl",
	"public/trade_metrics.jsonl",
}

type manifest struct {
	CEGS            string            `json:"cegs"`
	DatasetVersion  string            `json:"dataset_version"`
	RecordCounts    map[string]int    `json:"record_counts"`
	ChecksumsSHA256 map[string]string `json:"checksums_sha256"`
}

// Verify checks hashes, counts, deterministic ordering, duplicate IDs,
// cross-resource references, CEGS conformance, and current/latest parity.
func Verify(repositoryRoot string) error {
	dataRoot := filepath.Join(repositoryRoot, "data")
	manifestPath := filepath.Join(dataRoot, "public", "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read public manifest: %w", err)
	}
	var releaseManifest manifest
	if err := json.Unmarshal(manifestBytes, &releaseManifest); err != nil {
		return fmt.Errorf("parse public manifest: %w", err)
	}
	if releaseManifest.CEGS != cegs.SpecVersion {
		return fmt.Errorf("manifest CEGS version %q does not match implementation %q", releaseManifest.CEGS, cegs.SpecVersion)
	}
	if strings.TrimSpace(releaseManifest.DatasetVersion) == "" {
		return errors.New("manifest dataset_version is empty")
	}
	if err := verifyArtifactSet(releaseManifest.ChecksumsSHA256); err != nil {
		return err
	}
	for _, name := range expectedArtifacts {
		artifact, err := os.ReadFile(filepath.Join(dataRoot, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		actual := sha256.Sum256(artifact)
		if hex.EncodeToString(actual[:]) != releaseManifest.ChecksumsSHA256[name] {
			return fmt.Errorf("checksum mismatch for %s", name)
		}
	}
	if err := verifyManifestCopies(dataRoot, manifestBytes); err != nil {
		return err
	}
	if err := verifyCurrentMatchesRelease(dataRoot, releaseManifest.DatasetVersion); err != nil {
		return err
	}
	if err := verifyCounts(dataRoot, releaseManifest.RecordCounts); err != nil {
		return err
	}
	return verifyReferences(repositoryRoot, dataRoot)
}

func verifyArtifactSet(checksums map[string]string) error {
	if len(checksums) != len(expectedArtifacts) {
		return fmt.Errorf("manifest lists %d checksums; expected %d", len(checksums), len(expectedArtifacts))
	}
	for _, name := range expectedArtifacts {
		hash, ok := checksums[name]
		if !ok {
			return fmt.Errorf("manifest does not checksum %s", name)
		}
		if len(hash) != sha256.Size*2 {
			return fmt.Errorf("manifest checksum for %s is not SHA-256", name)
		}
		if _, err := hex.DecodeString(hash); err != nil {
			return fmt.Errorf("manifest checksum for %s is invalid: %w", name, err)
		}
	}
	return nil
}

func verifyManifestCopies(dataRoot string, publicManifest []byte) error {
	cegsManifest, err := os.ReadFile(filepath.Join(dataRoot, "cegs", "manifest.json"))
	if err != nil {
		return fmt.Errorf("read CEGS manifest: %w", err)
	}
	if !bytes.Equal(publicManifest, cegsManifest) {
		return errors.New("public and CEGS manifests differ")
	}
	report, err := cegs.Validate(publicManifest)
	if err != nil {
		return fmt.Errorf("validate manifest: %w", err)
	}
	if !report.Valid {
		return fmt.Errorf("manifest is not CEGS-valid: %s", strings.Join(report.Errors, "; "))
	}
	return nil
}

func verifyCurrentMatchesRelease(dataRoot, version string) error {
	releasesRoot := filepath.Join(dataRoot, "releases")
	entries, err := os.ReadDir(releasesRoot)
	if err != nil {
		return fmt.Errorf("list releases: %w", err)
	}
	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	if len(versions) == 0 {
		return errors.New("no immutable releases found")
	}
	sort.Strings(versions)
	if versions[len(versions)-1] != version {
		return fmt.Errorf("current dataset version %s is not the latest immutable release %s", version, versions[len(versions)-1])
	}

	releaseRoot := filepath.Join(releasesRoot, version)
	for _, name := range append(append([]string{}, expectedArtifacts...), "cegs/manifest.json", "public/manifest.json") {
		current, err := os.ReadFile(filepath.Join(dataRoot, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Errorf("read current %s: %w", name, err)
		}
		released, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(name)))
		if err != nil {
			return fmt.Errorf("read release %s: %w", name, err)
		}
		if !bytes.Equal(current, released) {
			return fmt.Errorf("current %s differs from immutable release %s", name, version)
		}
	}
	return nil
}

func verifyCounts(dataRoot string, counts map[string]int) error {
	files := map[string]string{
		"projects":      "public/projects.jsonl",
		"organizations": "public/entities.jsonl",
		"events":        "public/events.jsonl",
		"evidence":      "public/evidence.jsonl",
		"scores":        "public/scores.jsonl",
		"trade_metrics": "public/trade_metrics.jsonl",
	}
	if len(counts) != len(files) {
		return fmt.Errorf("manifest has %d record counts; expected %d", len(counts), len(files))
	}
	for key, name := range files {
		actual, err := countJSONLLines(filepath.Join(dataRoot, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		if counts[key] != actual {
			return fmt.Errorf("record count mismatch for %s: manifest=%d actual=%d", key, counts[key], actual)
		}
	}
	for key, name := range map[string]string{
		"projects":      "cegs/projects.jsonl",
		"organizations": "cegs/organizations.jsonl",
		"events":        "cegs/events.jsonl",
		"evidence":      "cegs/evidence.jsonl",
	} {
		actual, err := countJSONLLines(filepath.Join(dataRoot, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		if counts[key] != actual {
			return fmt.Errorf("CEGS record count mismatch for %s: manifest=%d actual=%d", key, counts[key], actual)
		}
	}
	return nil
}

func countJSONLLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)
	count := 0
	for scanner.Scan() {
		if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
			return 0, fmt.Errorf("%s contains an empty JSONL line at %d", path, count+1)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("scan %s: %w", path, err)
	}
	return count, nil
}

func verifyReferences(repositoryRoot, dataRoot string) error {
	publicEvidence, err := readJSONL[domain.Evidence](filepath.Join(dataRoot, "public", "evidence.jsonl"), false)
	if err != nil {
		return err
	}
	entities, err := readJSONL[domain.Entity](filepath.Join(dataRoot, "public", "entities.jsonl"), false)
	if err != nil {
		return err
	}
	projects, err := readJSONL[domain.Project](filepath.Join(dataRoot, "public", "projects.jsonl"), false)
	if err != nil {
		return err
	}
	events, err := readJSONL[domain.Event](filepath.Join(dataRoot, "public", "events.jsonl"), false)
	if err != nil {
		return err
	}
	scores, err := readJSONL[domain.ProjectScore](filepath.Join(dataRoot, "public", "scores.jsonl"), false)
	if err != nil {
		return err
	}
	tradeMetrics, err := readJSONL[domain.TradeMetric](filepath.Join(dataRoot, "public", "trade_metrics.jsonl"), false)
	if err != nil {
		return err
	}

	evidenceIDs := ids(publicEvidence, func(value domain.Evidence) string { return value.ID })
	entityIDs := ids(entities, func(value domain.Entity) string { return value.ID })
	projectIDs := ids(projects, func(value domain.Project) string { return value.ID })
	for _, entity := range entities {
		if err := requireReference("entity "+entity.ID+" evidence", entity.EvidenceID, evidenceIDs, true); err != nil {
			return err
		}
	}
	for _, project := range projects {
		if err := requireReference("project "+project.ID+" proponent", project.ProponentID, entityIDs, true); err != nil {
			return err
		}
		for _, evidenceID := range project.EvidenceIDs {
			if err := requireReference("project "+project.ID+" evidence", evidenceID, evidenceIDs, false); err != nil {
				return err
			}
		}
		if err := requireUnique("project "+project.ID+" evidence", project.EvidenceIDs); err != nil {
			return err
		}
	}
	for _, event := range events {
		if err := requireReference("event "+event.ID+" project", event.ProjectID, projectIDs, false); err != nil {
			return err
		}
		if err := requireReference("event "+event.ID+" evidence", event.EvidenceID, evidenceIDs, false); err != nil {
			return err
		}
	}
	for _, score := range scores {
		if err := requireReference("score "+score.ID+" project", score.ProjectID, projectIDs, false); err != nil {
			return err
		}
		for _, evidenceID := range score.EvidenceIDs {
			if err := requireReference("score "+score.ID+" evidence", evidenceID, evidenceIDs, false); err != nil {
				return err
			}
		}
		if err := requireUnique("score "+score.ID+" evidence", score.EvidenceIDs); err != nil {
			return err
		}
	}
	for _, metric := range tradeMetrics {
		if err := requireReference("trade metric "+metric.ID+" evidence", metric.EvidenceID, evidenceIDs, false); err != nil {
			return err
		}
	}

	cegsEvidence, err := readJSONL[cegs.Evidence](filepath.Join(dataRoot, "cegs", "evidence.jsonl"), true)
	if err != nil {
		return err
	}
	cegsOrganizations, err := readJSONL[cegs.Organization](filepath.Join(dataRoot, "cegs", "organizations.jsonl"), true)
	if err != nil {
		return err
	}
	cegsProjects, err := readJSONL[cegs.Project](filepath.Join(dataRoot, "cegs", "projects.jsonl"), true)
	if err != nil {
		return err
	}
	cegsEvents, err := readJSONL[cegs.Event](filepath.Join(dataRoot, "cegs", "events.jsonl"), true)
	if err != nil {
		return err
	}
	cegsEvidenceIDs := ids(cegsEvidence, func(value cegs.Evidence) string { return value.ID })
	cegsOrganizationIDs := ids(cegsOrganizations, func(value cegs.Organization) string { return value.ID })
	cegsProjectIDs := ids(cegsProjects, func(value cegs.Project) string { return value.ID })
	for _, organization := range cegsOrganizations {
		for _, evidenceID := range organization.Provenance {
			if err := requireReference("CEGS organization "+organization.ID+" provenance", evidenceID, cegsEvidenceIDs, false); err != nil {
				return err
			}
		}
	}
	for _, project := range cegsProjects {
		for _, evidenceID := range project.Provenance {
			if err := requireReference("CEGS project "+project.ID+" provenance", evidenceID, cegsEvidenceIDs, false); err != nil {
				return err
			}
		}
		for _, proponentID := range project.Proponents {
			if err := requireReference("CEGS project "+project.ID+" proponent", proponentID, cegsOrganizationIDs, false); err != nil {
				return err
			}
		}
	}
	for _, event := range cegsEvents {
		if err := requireReference("CEGS event "+event.ID+" subject", event.Subject, cegsProjectIDs, false); err != nil {
			return err
		}
		for _, evidenceID := range event.Evidence {
			if err := requireReference("CEGS event "+event.ID+" evidence", evidenceID, cegsEvidenceIDs, false); err != nil {
				return err
			}
		}
	}
	return verifyWebSnapshots(repositoryRoot, projectIDs, evidenceIDs)
}

func verifyWebSnapshots(repositoryRoot string, projectIDs, evidenceIDs map[string]struct{}) error {
	type webEvidence struct {
		ID string `json:"id"`
	}
	type webProject struct {
		ID       string        `json:"id"`
		Evidence []webEvidence `json:"evidence"`
	}
	path := filepath.Join(repositoryRoot, "apps", "web", "data", "projects.snapshot.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read web project snapshot: %w", err)
	}
	var projects []webProject
	if err := json.Unmarshal(data, &projects); err != nil {
		return fmt.Errorf("parse web project snapshot: %w", err)
	}
	if len(projects) != len(projectIDs) {
		return fmt.Errorf("web project snapshot has %d projects; expected %d", len(projects), len(projectIDs))
	}
	seenProjects := make(map[string]struct{}, len(projects))
	for _, project := range projects {
		if err := requireReference("web snapshot project", project.ID, projectIDs, false); err != nil {
			return err
		}
		if _, duplicate := seenProjects[project.ID]; duplicate {
			return fmt.Errorf("web project snapshot contains duplicate project %s", project.ID)
		}
		seenProjects[project.ID] = struct{}{}
		seenEvidence := make(map[string]struct{}, len(project.Evidence))
		for _, evidence := range project.Evidence {
			if err := requireReference("web snapshot project "+project.ID+" evidence", evidence.ID, evidenceIDs, false); err != nil {
				return err
			}
			if _, duplicate := seenEvidence[evidence.ID]; duplicate {
				return fmt.Errorf("web snapshot project %s embeds duplicate evidence %s", project.ID, evidence.ID)
			}
			seenEvidence[evidence.ID] = struct{}{}
		}
	}

	webManifestPath := filepath.Join(repositoryRoot, "apps", "web", "data", "manifest.snapshot.json")
	webManifest, err := os.ReadFile(webManifestPath)
	if err != nil {
		return fmt.Errorf("read web manifest snapshot: %w", err)
	}
	publicManifest, err := os.ReadFile(filepath.Join(repositoryRoot, "data", "public", "manifest.json"))
	if err != nil {
		return fmt.Errorf("read public manifest for web comparison: %w", err)
	}
	var webValue, publicValue any
	if err := json.Unmarshal(webManifest, &webValue); err != nil {
		return fmt.Errorf("parse web manifest snapshot: %w", err)
	}
	if err := json.Unmarshal(publicManifest, &publicValue); err != nil {
		return fmt.Errorf("parse public manifest for web comparison: %w", err)
	}
	if !deepJSONEqual(webValue, publicValue) {
		return errors.New("web manifest snapshot differs from the public release manifest")
	}
	return nil
}

func deepJSONEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func readJSONL[T any](path string, validateCEGS bool) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	items := make([]T, 0)
	seen := make(map[string]struct{})
	previousID := ""
	lineNumber := 0
	for {
		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			lineNumber++
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				return nil, fmt.Errorf("%s contains an empty JSONL line at %d", path, lineNumber)
			}
			var envelope struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(line, &envelope); err != nil {
				return nil, fmt.Errorf("parse %s line %d: %w", path, lineNumber, err)
			}
			if envelope.ID == "" {
				return nil, fmt.Errorf("%s line %d has no id", path, lineNumber)
			}
			if _, duplicate := seen[envelope.ID]; duplicate {
				return nil, fmt.Errorf("%s contains duplicate id %s", path, envelope.ID)
			}
			if previousID != "" && envelope.ID <= previousID {
				return nil, fmt.Errorf("%s is not strictly sorted by id at %s", path, envelope.ID)
			}
			seen[envelope.ID] = struct{}{}
			previousID = envelope.ID
			if validateCEGS {
				report, err := cegs.Validate(line)
				if err != nil {
					return nil, fmt.Errorf("validate %s line %d: %w", path, lineNumber, err)
				}
				if !report.Valid {
					return nil, fmt.Errorf("%s line %d is not CEGS-valid: %s", path, lineNumber, strings.Join(report.Errors, "; "))
				}
			}
			var item T
			if err := json.Unmarshal(line, &item); err != nil {
				return nil, fmt.Errorf("decode %s line %d: %w", path, lineNumber, err)
			}
			items = append(items, item)
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read %s: %w", path, readErr)
		}
	}
	return items, nil
}

func ids[T any](values []T, id func(T) string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[id(value)] = struct{}{}
	}
	return result
}

func requireReference(label, reference string, available map[string]struct{}, optional bool) error {
	if reference == "" && optional {
		return nil
	}
	if reference == "" {
		return fmt.Errorf("%s is empty", label)
	}
	if _, ok := available[reference]; !ok {
		return fmt.Errorf("%s references missing id %s", label, reference)
	}
	return nil
}

func requireUnique(label string, references []string) error {
	seen := make(map[string]struct{}, len(references))
	for _, reference := range references {
		if _, duplicate := seen[reference]; duplicate {
			return fmt.Errorf("%s contains duplicate id %s", label, reference)
		}
		seen[reference] = struct{}{}
	}
	return nil
}
