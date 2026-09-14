import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(scriptDirectory, "../../..");
const outputDirectory = resolve(scriptDirectory, "../data");

async function readJsonLines(path) {
  const value = await readFile(path, "utf8");
  return value
    .split(/\r?\n/)
    .filter(Boolean)
    .map((line) => JSON.parse(line));
}

function stableSourceId(publisher, url) {
  const digest = createHash("sha256").update(`${publisher}\0${url}`).digest("hex").slice(0, 20);
  return `snapshot-source:${digest}`;
}

const sourceProfiles = {
  "Natural Resources Canada": {
    name: "NRCan Major Projects Inventory 2025–2035",
    publisherId: "publisher:ca:nrcan",
    jurisdiction: "CA",
    geography: ["CA"],
    family: "CKAN",
    accessMethod: "REST_API",
    contentType: "application/json, text/csv, application/geo+json",
    subjects: ["major projects", "infrastructure", "capital investment", "open data"],
    sectors: ["all tracked economic sectors"],
    languages: ["en-CA", "fr-CA"],
    frequency: "ANNUAL",
    licence: "Open Government Licence - Canada",
    coverage: "NATIONAL_PROJECT_INVENTORY",
  },
  "Canadian Nuclear Safety Commission": {
    name: "Darlington New Nuclear Project regulatory record",
    publisherId: "publisher:ca:cnsc",
    jurisdiction: "CA",
    geography: ["CA:ON"],
    family: "REGULATORY_WEB_RECORD",
    accessMethod: "WEB_PAGE",
    contentType: "text/html",
    subjects: ["nuclear regulation", "construction licence", "project milestone"],
    sectors: ["Nuclear & Clean Power"],
    languages: ["en-CA", "fr-CA"],
    frequency: "EVENT_DRIVEN",
    licence: "Publisher terms apply",
    coverage: "PROJECT_REGULATORY_RECORD",
  },
  "Impact Assessment Agency of Canada": {
    name: "Impact Assessment Registry project record",
    publisherId: "publisher:ca:iaac",
    jurisdiction: "CA",
    geography: ["CA"],
    family: "REGULATORY_WEB_RECORD",
    accessMethod: "WEB_PAGE",
    contentType: "text/html",
    subjects: ["impact assessment", "regulatory review", "project milestone"],
    sectors: ["Critical Minerals", "Transportation & Ports"],
    languages: ["en-CA", "fr-CA"],
    frequency: "EVENT_DRIVEN",
    licence: "Publisher terms apply",
    coverage: "PROJECT_REGULATORY_RECORD",
  },
  "Ontario Power Generation Inc.": {
    name: "Ontario Power Generation audited financial disclosure",
    publisherId: "publisher:ca:on:opg",
    jurisdiction: "CA:ON",
    geography: ["CA:ON"],
    family: "ISSUER_DISCLOSURE",
    accessMethod: "DOCUMENT",
    contentType: "application/pdf",
    subjects: ["audited financial statements", "capital cost", "nuclear project"],
    sectors: ["Nuclear & Clean Power"],
    languages: ["en-CA"],
    frequency: "QUARTERLY",
    licence: "Publisher terms apply",
    coverage: "PRIMARY_ISSUER_DISCLOSURE",
  },
  "Canada Infrastructure Bank": {
    name: "Oneida Energy Storage investment announcement",
    publisherId: "publisher:ca:cib",
    jurisdiction: "CA",
    geography: ["CA:ON"],
    family: "CROWN_CORPORATION_DISCLOSURE",
    accessMethod: "WEB_PAGE",
    contentType: "text/html",
    subjects: ["project finance", "infrastructure investment", "energy storage"],
    sectors: ["Clean Energy & Grid"],
    languages: ["en-CA", "fr-CA"],
    frequency: "EVENT_DRIVEN",
    licence: "Publisher terms apply",
    coverage: "PROJECT_FINANCING_RECORD",
  },
  "Northland Power Inc.": {
    name: "Oneida Energy Storage commercial operations announcement",
    publisherId: "publisher:ca:northland-power",
    jurisdiction: "CA",
    geography: ["CA:ON"],
    family: "ISSUER_DISCLOSURE",
    accessMethod: "WEB_PAGE",
    contentType: "text/html",
    subjects: ["commercial operations", "energy storage", "project delivery"],
    sectors: ["Clean Energy & Grid"],
    languages: ["en-CA"],
    frequency: "EVENT_DRIVEN",
    licence: "Publisher terms apply",
    coverage: "PRIMARY_ISSUER_DISCLOSURE",
  },
};

const [projects, evidence, manifest] = await Promise.all([
  readJsonLines(resolve(repositoryRoot, "data/public/projects.jsonl")),
  readJsonLines(resolve(repositoryRoot, "data/public/evidence.jsonl")),
  readFile(resolve(repositoryRoot, "data/public/manifest.json"), "utf8").then(JSON.parse),
]);

const evidenceById = new Map(evidence.map((item) => [item.id, item]));
const compactProjects = projects.map((project) => ({
  id: project.id,
  slug: project.slug,
  name: project.name,
  summary: project.summary,
  sector: project.sector,
  subsector: project.subsector,
  province: project.province,
  location_name: project.location_name,
  latitude: project.latitude,
  longitude: project.longitude,
  current_stage: project.current_stage,
  capex_cad: project.capex_cad,
  capex_status: project.capex_status,
  proponent_id: project.proponent_id,
  proponent_name: project.proponent?.common_name || project.proponent?.legal_name,
  confidence: project.confidence,
  scores: project.scores,
  last_meaningful_update: project.last_meaningful_update,
  evidence: (project.evidence_ids || [])
    .map((id) => evidenceById.get(String(id).replace(/^cegs:evidence:ca:/, "")))
    .filter(Boolean)
    .map((item) => ({
      id: item.id,
      source_url: item.source_url,
      publisher: item.publisher,
      source_tier: item.source_tier,
      retrieval_timestamp: item.retrieval_timestamp,
      effective_date: item.effective_date,
      confidence: item.confidence,
      content_hash: item.content_hash,
      locator: item.locator,
      source_record_id: item.source_record_id,
      pipeline_version: item.pipeline_version,
    })),
}));

const sourceGroups = new Map();
for (const item of evidence) {
  const key = `${item.publisher}\0${item.source_url}`;
  const group = sourceGroups.get(key) || { publisher: item.publisher, url: item.source_url, records: [] };
  group.records.push(item);
  sourceGroups.set(key, group);
}

const compactSources = [...sourceGroups.values()].map((group) => {
  const profile = sourceProfiles[group.publisher];
  if (!profile) throw new Error(`Missing source profile for ${group.publisher}`);
  const lastRetrieved = group.records.map((item) => item.retrieval_timestamp).filter(Boolean).sort().at(-1);
  const lastChanged = group.records.map((item) => item.effective_date).filter(Boolean).sort().at(-1);
  return {
    id: stableSourceId(group.publisher, group.url),
    name: profile.name,
    publisher_id: profile.publisherId,
    publisher_name: group.publisher,
    canonical_url: group.url,
    jurisdiction: profile.jurisdiction,
    geography: profile.geography,
    source_family: profile.family,
    access_method: profile.accessMethod,
    content_type: profile.contentType,
    authority_tier: group.records.every((item) => item.source_tier === 1) ? 1 : 2,
    subject_tags: profile.subjects,
    sector_tags: profile.sectors,
    languages: profile.languages,
    update_frequency: profile.frequency,
    lifecycle: "ACTIVE",
    health: "CURRENT",
    last_checked_at: lastRetrieved,
    last_success_at: lastRetrieved,
    last_change_at: lastChanged,
    license: profile.licence,
    quality: {},
    coverage_class: profile.coverage,
    description: `Primary-source record published by ${group.publisher}. This bundled, reviewed snapshot consolidates ${group.records.length} evidence ${group.records.length === 1 ? "record" : "records"} and links directly to the publisher-controlled source.`,
    evidence_record_count: group.records.length,
  };
});

compactSources.sort((a, b) => a.publisher_name.localeCompare(b.publisher_name) || a.name.localeCompare(b.name));

await mkdir(outputDirectory, { recursive: true });
await Promise.all([
  writeFile(resolve(outputDirectory, "projects.snapshot.json"), `${JSON.stringify(compactProjects)}\n`),
  writeFile(resolve(outputDirectory, "sources.snapshot.json"), `${JSON.stringify(compactSources)}\n`),
  writeFile(resolve(outputDirectory, "manifest.snapshot.json"), `${JSON.stringify(manifest)}\n`),
]);

console.log(`Generated web snapshot: ${compactProjects.length} projects, ${evidence.length} evidence records, ${compactSources.length} canonical source records.`);
