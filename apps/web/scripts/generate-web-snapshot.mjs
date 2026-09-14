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
  "World Bank": {
    name: "World Bank Indicators — Canada trade and logistics",
    publisherId: "publisher:intl:world-bank",
    jurisdiction: "INTL:WORLD_BANK",
    geography: ["GLOBAL", "CA"],
    family: "INDICATORS_API",
    accessMethod: "REST_API",
    contentType: "application/json",
    subjects: ["international trade", "logistics performance", "supply chain", "official statistics"],
    sectors: ["global trade and supply chains", "Transportation & Ports", "Industrial & Manufacturing"],
    languages: ["en"],
    frequency: "PERIODIC_ANNUAL",
    licence: "Creative Commons Attribution 4.0",
    coverage: "GLOBAL_TRADE_AND_LOGISTICS_SCORE_INPUTS",
  },
};

const registeredTradeSources = [
  {
    id: "source:ca:statcan:web-data-service",
    name: "Statistics Canada Web Data Service",
    publisher_id: "publisher:ca:statcan",
    publisher_name: "Statistics Canada",
    canonical_url: "https://www.statcan.gc.ca/en/developers/wds",
    jurisdiction: "CA",
    geography: ["CA"],
    source_family: "REST_API",
    access_method: "REST_API",
    content_type: "application/json, text/csv, application/vnd.sdmx",
    subject_tags: ["international trade", "merchandise trade", "supply chain", "official statistics"],
    sector_tags: ["global trade and supply chains"],
    languages: ["en-CA", "fr-CA"],
    update_frequency: "BUSINESS_DAILY",
    license: "Statistics Canada Open Licence",
    coverage_class: "CANADIAN_OFFICIAL_STATISTICS_API",
    description: "Official Statistics Canada API for released data and metadata, including Canadian international merchandise trade tables and daily change lists.",
  },
  {
    id: "source:ca:ised:trade-data-online",
    name: "Trade Data Online",
    publisher_id: "publisher:ca:ised",
    publisher_name: "Innovation, Science and Economic Development Canada",
    canonical_url: "https://ised-isde.canada.ca/site/trade-data-online/en?Open=1",
    jurisdiction: "CA",
    geography: ["CA", "US", "GLOBAL"],
    source_family: "TRADE_DATA_PORTAL",
    access_method: "WEB_APP_DOWNLOAD",
    content_type: "text/html, text/csv, application/vnd.ms-excel",
    subject_tags: ["international trade", "imports", "exports", "HS codes", "NAICS", "import replacement"],
    sector_tags: ["global trade and supply chains"],
    languages: ["en-CA", "fr-CA"],
    update_frequency: "MONTHLY",
    license: "Government of Canada terms apply",
    coverage_class: "CANADA_US_BILATERAL_TRADE",
    description: "Official Canadian trade portal for customized product- and industry-level reports covering Canada and United States trade with more than 200 economies.",
  },
  {
    id: "source:un:comtrade:api",
    name: "UN Comtrade API",
    publisher_id: "publisher:un:statistics-division",
    publisher_name: "United Nations Statistics Division",
    canonical_url: "https://comtradeapi.un.org/",
    jurisdiction: "INTL:UN",
    geography: ["GLOBAL"],
    source_family: "REST_API",
    access_method: "REST_API",
    content_type: "application/json, text/csv",
    subject_tags: ["international trade", "bilateral trade", "HS commodities", "imports", "exports"],
    sector_tags: ["global trade and supply chains", "Critical Minerals", "Industrial & Manufacturing"],
    languages: ["en"],
    update_frequency: "MONTHLY_ANNUAL",
    license: "UN Comtrade terms apply",
    coverage_class: "GLOBAL_BILATERAL_MERCHANDISE_TRADE",
    description: "United Nations API for global bilateral merchandise trade by reporting economy, partner, commodity classification, flow and period.",
  },
  {
    id: "source:oecd:tiva:sdmx",
    name: "OECD Trade in Value Added (TiVA) 2025",
    publisher_id: "publisher:intl:oecd",
    publisher_name: "Organisation for Economic Co-operation and Development",
    canonical_url: "https://data-explorer.oecd.org/vis?df%5Bag%5D=OECD.STI.PIE&df%5Bds%5D=dsDisseminateFinalDMZ&df%5Bid%5D=DSD_TIVA_MAINLV%40DF_MAINLV&df%5Bvs%5D=1.1",
    jurisdiction: "INTL:OECD",
    geography: ["GLOBAL", "CA"],
    source_family: "SDMX_API",
    access_method: "REST_API",
    content_type: "application/vnd.sdmx.data+csv, application/vnd.sdmx.data+json",
    subject_tags: ["international trade", "trade in value added", "global value chains", "domestic value added"],
    sector_tags: ["global trade and supply chains", "Industrial & Manufacturing"],
    languages: ["en", "fr"],
    update_frequency: "EDITION_BASED",
    license: "OECD terms and conditions apply",
    coverage_class: "GLOBAL_VALUE_CHAIN_INDICATORS",
    description: "Official OECD TiVA dataset and SDMX service for tracing domestic and foreign value added through global production and final demand.",
  },
  {
    id: "source:wto:timeseries-api",
    name: "WTO Timeseries API",
    publisher_id: "publisher:intl:wto",
    publisher_name: "World Trade Organization",
    canonical_url: "https://apiportal.wto.org/",
    jurisdiction: "INTL:WTO",
    geography: ["GLOBAL", "CA"],
    source_family: "REST_API",
    access_method: "REST_API_KEY",
    content_type: "application/json, text/csv",
    subject_tags: ["international trade", "services trade", "tariffs", "market access", "non-tariff measures"],
    sector_tags: ["global trade and supply chains"],
    languages: ["en", "fr", "es"],
    update_frequency: "MONTHLY_QUARTERLY_ANNUAL",
    license: "WTO API terms apply",
    coverage_class: "GLOBAL_TRADE_POLICY_AND_TIMESERIES",
    description: "Official WTO developer service for merchandise and services trade, tariffs, market access and non-tariff indicators. A free subscription key is required.",
    authentication_required: true,
  },
  {
    id: "source:worldbank:logistics-performance-index",
    name: "World Bank Logistics Performance Index",
    publisher_id: "publisher:intl:world-bank",
    publisher_name: "World Bank",
    canonical_url: "https://datacatalog.worldbank.org/search/dataset/0038649/logistics-performance-index",
    jurisdiction: "INTL:WORLD_BANK",
    geography: ["GLOBAL", "CA"],
    source_family: "INDICATORS_API",
    access_method: "REST_API_DOWNLOAD",
    content_type: "application/json, text/csv, application/zip",
    subject_tags: ["supply chain", "logistics performance", "customs", "trade infrastructure", "shipment reliability"],
    sector_tags: ["global trade and supply chains", "Transportation & Ports"],
    languages: ["en"],
    update_frequency: "PERIODIC",
    license: "Creative Commons Attribution 4.0",
    coverage_class: "GLOBAL_LOGISTICS_PERFORMANCE",
    description: "World Bank public dataset measuring customs, trade infrastructure, international shipments, logistics services, tracking and delivery timeliness.",
  },
  {
    id: "source:imf:portwatch-search-api",
    name: "IMF PortWatch Search API",
    publisher_id: "publisher:intl:imf",
    publisher_name: "International Monetary Fund",
    canonical_url: "https://portwatch.imf.org/api/search/definition/",
    jurisdiction: "INTL:IMF",
    geography: ["GLOBAL", "CA"],
    source_family: "OGC_API_RECORDS",
    access_method: "REST_API",
    content_type: "application/geo+json, application/json",
    subject_tags: ["supply chain", "ports", "shipping", "maritime disruption", "geospatial catalog"],
    sector_tags: ["global trade and supply chains", "Transportation & Ports"],
    languages: ["en"],
    update_frequency: "CONTINUAL",
    license: "IMF PortWatch terms apply",
    coverage_class: "GLOBAL_PORT_AND_SHIPPING_INTELLIGENCE",
    description: "Official IMF PortWatch OGC API catalogue for discovering port, shipping and maritime disruption datasets and geospatial services.",
  },
  {
    id: "source:unctad:data-centre",
    name: "UNCTADstat Data Centre",
    publisher_id: "publisher:un:unctad",
    publisher_name: "United Nations Trade and Development",
    canonical_url: "https://unctadstat.unctad.org/datacentre/",
    jurisdiction: "INTL:UNCTAD",
    geography: ["GLOBAL", "CA"],
    source_family: "STATISTICAL_DATA_PORTAL",
    access_method: "WEB_APP_DOWNLOAD",
    content_type: "text/html, text/csv",
    subject_tags: ["international trade", "critical minerals", "maritime connectivity", "port throughput", "trade and transport"],
    sector_tags: ["global trade and supply chains", "Critical Minerals", "Transportation & Ports"],
    languages: ["en", "fr"],
    update_frequency: "MONTHLY_QUARTERLY_ANNUAL",
    license: "Creative Commons Attribution 3.0 IGO",
    coverage_class: "GLOBAL_TRADE_AND_MARITIME_STATISTICS",
    description: "Official UNCTAD statistical centre covering merchandise and services trade, critical minerals, port throughput, maritime connectivity and trade-and-transport indicators.",
  },
];

const [projects, evidence, tradeMetrics, manifest] = await Promise.all([
  readJsonLines(resolve(repositoryRoot, "data/public/projects.jsonl")),
  readJsonLines(resolve(repositoryRoot, "data/public/evidence.jsonl")),
  readJsonLines(resolve(repositoryRoot, "data/public/trade_metrics.jsonl")),
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
  score_details: (project.score_details || []).map((score) => ({
    score_type: score.score_type,
    score_value: score.score_value,
    score_version: score.score_version,
    factors: score.factors,
    factor_evidence: score.factor_evidence,
    evidence_ids: score.evidence_ids,
    unknown_factors: score.unknown_factors,
    coverage: score.coverage,
    confidence: score.confidence,
    input_hash: score.input_hash,
    explanation: score.explanation,
    calculated_at: score.calculated_at,
  })),
  last_meaningful_update: project.last_meaningful_update,
  evidence: [...new Set([
    ...(project.evidence_ids || []),
    ...(project.score_details || []).flatMap((score) => score.evidence_ids || []),
  ])]
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
  const sourceURL = item.publisher === "World Bank"
    ? "https://api.worldbank.org/v2/country/CAN"
    : item.source_url;
  const key = `${item.publisher}\0${sourceURL}`;
  const group = sourceGroups.get(key) || { publisher: item.publisher, url: sourceURL, records: [] };
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
    integration_status: "EVIDENCE_LINKED",
    authentication_required: false,
  };
});

compactSources.push(...registeredTradeSources
  .filter((source) => source.id !== "source:worldbank:logistics-performance-index")
  .map((source) => ({
  ...source,
  authority_tier: 1,
  lifecycle: "APPROVED",
  health: "CURRENT",
  last_checked_at: "2026-09-13T21:00:00Z",
  last_success_at: "2026-09-13T21:00:00Z",
  quality: {},
  evidence_record_count: 0,
  integration_status: "REGISTERED_NOT_INGESTED",
  authentication_required: source.authentication_required ?? false,
})));

compactSources.sort((a, b) => a.publisher_name.localeCompare(b.publisher_name) || a.name.localeCompare(b.name));

await mkdir(outputDirectory, { recursive: true });
await Promise.all([
  writeFile(resolve(outputDirectory, "projects.snapshot.json"), `${JSON.stringify(compactProjects)}\n`),
  writeFile(resolve(outputDirectory, "sources.snapshot.json"), `${JSON.stringify(compactSources)}\n`),
  writeFile(resolve(outputDirectory, "trade-metrics.snapshot.json"), `${JSON.stringify(tradeMetrics)}\n`),
  writeFile(resolve(outputDirectory, "manifest.snapshot.json"), `${JSON.stringify(manifest)}\n`),
]);

console.log(`Generated web snapshot: ${compactProjects.length} projects, ${evidence.length} evidence records, ${tradeMetrics.length} trade metrics, ${compactSources.length} canonical source records.`);
