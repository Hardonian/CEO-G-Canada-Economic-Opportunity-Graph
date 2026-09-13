import type {
  DeadLetterSummary,
  MeasuredRatio,
  PublicSource,
  SignalLatencySummary,
  SourceCoverageReport,
  SourceHealthStatus,
  SourceLifecycle,
  SourceListResponse,
  SourceQuality,
} from "./types";

const API_BASE =
  process.env.COG_API_BASE ||
  process.env.NEXT_PUBLIC_API_BASE ||
  "http://localhost:8080/api/v1";

const SOURCE_LIFECYCLES = new Set<SourceLifecycle>([
  "DISCOVERED",
  "CLASSIFIED",
  "TESTED",
  "APPROVED",
  "ACTIVE",
  "REJECTED",
  "BLOCKED",
  "RETIRED",
]);

const SOURCE_HEALTH_STATUSES = new Set<SourceHealthStatus>([
  "CURRENT",
  "HEALTHY",
  "DELAYED",
  "STALE",
  "DEGRADED",
  "BROKEN",
  "UNAVAILABLE",
  "DISABLED",
  "UNKNOWN",
  "NOT_YET_CHECKED",
]);

const QUALITY_KEYS = new Set<keyof SourceQuality>([
  "structuredness",
  "freshness",
  "completeness",
  "stability",
  "authority",
  "historical_depth",
]);

const SOURCE_QUERY_KEYS = [
  "q",
  "publisher",
  "jurisdiction",
  "sector",
  "family",
  "format",
  "update_frequency",
  "authority",
  "lifecycle",
  "health",
] as const;

export type SourceQueryKey = (typeof SOURCE_QUERY_KEYS)[number];
export type SourceQuery = Partial<Record<SourceQueryKey, string>> & {
  limit?: number;
  offset?: number;
};

export type SourceDataResult<T> =
  | { status: "available"; data: T }
  | { status: "not_found"; data: null }
  | { status: "unavailable"; data: null };

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function boundedString(value: unknown, maxLength = 500): string | null {
  if (typeof value !== "string") return null;
  const normalized = value.trim();
  if (!normalized || normalized.length > maxLength) return null;
  return normalized;
}

function optionalTimestamp(value: unknown): string | undefined | null {
  if (value === undefined || value === null || value === "") return undefined;
  if (typeof value !== "string" || Number.isNaN(Date.parse(value))) return null;
  return value;
}

function stringArray(value: unknown, maxItems = 100): string[] | null {
  if (!Array.isArray(value) || value.length > maxItems) return null;
  const values: string[] = [];
  for (const item of value) {
    const normalized = boundedString(item, 160);
    if (!normalized) return null;
    values.push(normalized);
  }
  return values;
}

function nonNegativeInteger(value: unknown): number | null {
  return typeof value === "number" && Number.isSafeInteger(value) && value >= 0
    ? value
    : null;
}

function finiteNumber(value: unknown): number | null {
  return typeof value === "number" && Number.isFinite(value) ? value : null;
}

function publicHTTPURL(value: unknown): string | null {
  const raw = boundedString(value, 2_048);
  if (!raw) return null;
  try {
    const parsed = new URL(raw);
    if ((parsed.protocol !== "https:" && parsed.protocol !== "http:") || parsed.username || parsed.password) {
      return null;
    }
    return parsed.toString();
  } catch {
    return null;
  }
}

function normalizeQuality(value: unknown): SourceQuality | null {
  if (!isRecord(value)) return null;
  const quality: SourceQuality = {};
  for (const [key, raw] of Object.entries(value)) {
    if (!QUALITY_KEYS.has(key as keyof SourceQuality)) continue;
    const score = finiteNumber(raw);
    if (score === null || score < 0 || score > 100) return null;
    quality[key as keyof SourceQuality] = score;
  }
  return quality;
}

function normalizeSource(value: unknown): PublicSource | null {
  if (!isRecord(value)) return null;

  const id = boundedString(value.id, 200);
  const name = boundedString(value.name, 500);
  const publisherID = boundedString(value.publisher_id, 200);
  const publisherName = boundedString(value.publisher_name, 500);
  const canonicalURL = publicHTTPURL(value.canonical_url);
  const jurisdiction = boundedString(value.jurisdiction, 80);
  const geography = stringArray(value.geography);
  const sourceFamily = boundedString(value.source_family, 100);
  const accessMethod = boundedString(value.access_method, 100);
  const contentType = boundedString(value.content_type, 160);
  const authorityTier = nonNegativeInteger(value.authority_tier);
  const subjectTags = stringArray(value.subject_tags);
  const sectorTags = stringArray(value.sector_tags);
  const languages = stringArray(value.languages, 20);
  const updateFrequency = boundedString(value.update_frequency, 100);
  const lifecycle = boundedString(value.lifecycle, 40) as SourceLifecycle | null;
  const health = boundedString(value.health, 40) as SourceHealthStatus | null;
  const lastCheckedAt = optionalTimestamp(value.last_checked_at);
  const lastSuccessAt = optionalTimestamp(value.last_success_at);
  const lastChangeAt = optionalTimestamp(value.last_change_at);
  const license = boundedString(value.license, 500);
  const quality = normalizeQuality(value.quality);
  const coverageClass = boundedString(value.coverage_class, 100);
  const description = boundedString(value.description, 4_000);

  if (
    !id ||
    !name ||
    !publisherID ||
    !publisherName ||
    !canonicalURL ||
    !jurisdiction ||
    !geography ||
    !sourceFamily ||
    !accessMethod ||
    !contentType ||
    authorityTier === null ||
    authorityTier < 1 ||
    authorityTier > 5 ||
    !subjectTags ||
    !sectorTags ||
    !languages ||
    !updateFrequency ||
    !lifecycle ||
    !SOURCE_LIFECYCLES.has(lifecycle) ||
    !health ||
    !SOURCE_HEALTH_STATUSES.has(health) ||
    lastCheckedAt === null ||
    lastSuccessAt === null ||
    lastChangeAt === null ||
    !license ||
    !quality ||
    !coverageClass ||
    !description
  ) {
    return null;
  }

  return {
    id,
    name,
    publisher_id: publisherID,
    publisher_name: publisherName,
    canonical_url: canonicalURL,
    jurisdiction,
    geography,
    source_family: sourceFamily,
    access_method: accessMethod,
    content_type: contentType,
    authority_tier: authorityTier,
    subject_tags: subjectTags,
    sector_tags: sectorTags,
    languages,
    update_frequency: updateFrequency,
    lifecycle,
    health,
    last_checked_at: lastCheckedAt,
    last_success_at: lastSuccessAt,
    last_change_at: lastChangeAt,
    license,
    quality,
    coverage_class: coverageClass,
    description,
  };
}

function normalizeSourceList(value: unknown): SourceListResponse | null {
  if (!isRecord(value) || !Array.isArray(value.sources)) return null;
  const total = nonNegativeInteger(value.total);
  const limit = nonNegativeInteger(value.limit);
  const offset = nonNegativeInteger(value.offset);
  if (total === null || limit === null || limit < 1 || offset === null) return null;

  const sources = value.sources.flatMap((candidate) => {
    const source = normalizeSource(candidate);
    return source ? [source] : [];
  });

  // Reject a partially malformed page rather than presenting incomplete data
  // as a complete API response.
  if (sources.length !== value.sources.length) return null;
  return { sources, total, limit, offset };
}

function normalizeCountMap(value: unknown): Record<string, number> | null {
  if (!isRecord(value)) return null;
  const result: Record<string, number> = {};
  for (const [key, raw] of Object.entries(value)) {
    const normalizedKey = boundedString(key, 160);
    const count = nonNegativeInteger(raw);
    if (!normalizedKey || count === null) return null;
    result[normalizedKey] = count;
  }
  return result;
}

function normalizeMeasuredRatio(value: unknown): MeasuredRatio | null {
  if (!isRecord(value)) return null;
  const status = boundedString(value.status, 80);
  const ratio = value.value === undefined || value.value === null ? undefined : finiteNumber(value.value);
  if (!status || ratio === null || (ratio !== undefined && (ratio < 0 || ratio > 1))) return null;
  return { status, value: ratio };
}

function normalizeSignalLatency(value: unknown): SignalLatencySummary | null {
  if (!isRecord(value)) return null;
  const status = boundedString(value.status, 80);
  if (!status) return null;
  const result: SignalLatencySummary = { status };
  for (const key of ["p50_ms", "p95_ms", "sample_count"] as const) {
    if (value[key] === undefined || value[key] === null) continue;
    const measurement = nonNegativeInteger(value[key]);
    if (measurement === null) return null;
    result[key] = measurement;
  }
  return result;
}

function normalizeDeadLetters(value: unknown): DeadLetterSummary | null {
  if (!isRecord(value)) return null;
  const status = boundedString(value.status, 80);
  if (!status) return null;
  if (value.count === undefined || value.count === null) return { status };
  const count = nonNegativeInteger(value.count);
  return count === null ? null : { status, count };
}

function normalizeCoverage(value: unknown): SourceCoverageReport | null {
  if (!isRecord(value) || !isRecord(value.lifecycle_counts)) return null;
  const generatedAt = optionalTimestamp(value.generated_at);
  if (!generatedAt) return null;

  const lifecycleCounts = {
    discovered: nonNegativeInteger(value.lifecycle_counts.discovered),
    registered: nonNegativeInteger(value.lifecycle_counts.registered),
    tested: nonNegativeInteger(value.lifecycle_counts.tested),
    active: nonNegativeInteger(value.lifecycle_counts.active),
    broken: nonNegativeInteger(value.lifecycle_counts.broken),
  };
  if (Object.values(lifecycleCounts).some((count) => count === null)) return null;

  const byJurisdiction = normalizeCountMap(value.by_jurisdiction);
  const bySector = normalizeCountMap(value.by_sector);
  const byFamily = normalizeCountMap(value.by_family);
  const primarySourceRatio = normalizeMeasuredRatio(value.primary_source_ratio);
  const signalLatency = normalizeSignalLatency(value.signal_latency);
  const deadLetters = normalizeDeadLetters(value.dead_letters);
  const blindSpots = stringArray(value.blind_spots, 200);
  if (!byJurisdiction || !bySector || !byFamily || !primarySourceRatio || !signalLatency || !deadLetters || !blindSpots) {
    return null;
  }

  return {
    generated_at: generatedAt,
    lifecycle_counts: lifecycleCounts as SourceCoverageReport["lifecycle_counts"],
    by_jurisdiction: byJurisdiction,
    by_sector: bySector,
    by_family: byFamily,
    primary_source_ratio: primarySourceRatio,
    signal_latency: signalLatency,
    dead_letters: deadLetters,
    blind_spots: blindSpots,
  };
}

async function fetchSourceAPI(path: string, revalidate: number): Promise<Response | null> {
  try {
    return await fetch(`${API_BASE}${path}`, {
      headers: { Accept: "application/json" },
      next: { revalidate },
      signal: AbortSignal.timeout(5_000),
    });
  } catch {
    return null;
  }
}

function boundedQueryValue(value: string | undefined): string | null {
  if (!value) return null;
  const normalized = value.trim();
  return normalized && normalized.length <= 200 ? normalized : null;
}

export async function getSources(query: SourceQuery = {}): Promise<SourceDataResult<SourceListResponse>> {
  const params = new URLSearchParams();
  for (const key of SOURCE_QUERY_KEYS) {
    const value = boundedQueryValue(query[key]);
    if (value) params.set(key, value);
  }
  const limit = Number.isSafeInteger(query.limit) ? Math.min(100, Math.max(1, query.limit ?? 24)) : 24;
  const offset = Number.isSafeInteger(query.offset) ? Math.max(0, query.offset ?? 0) : 0;
  params.set("limit", String(limit));
  params.set("offset", String(offset));

  const response = await fetchSourceAPI(`/sources?${params.toString()}`, 60);
  if (!response?.ok) return { status: "unavailable", data: null };
  try {
    const sources = normalizeSourceList(await response.json());
    return sources ? { status: "available", data: sources } : { status: "unavailable", data: null };
  } catch {
    return { status: "unavailable", data: null };
  }
}

export async function getSource(id: string): Promise<SourceDataResult<PublicSource>> {
  const normalizedID = boundedQueryValue(id);
  if (!normalizedID) return { status: "not_found", data: null };
  const response = await fetchSourceAPI(`/sources/${encodeURIComponent(normalizedID)}`, 60);
  if (response?.status === 404) return { status: "not_found", data: null };
  if (!response?.ok) return { status: "unavailable", data: null };
  try {
    const payload: unknown = await response.json();
    const source = normalizeSource(isRecord(payload) && payload.source ? payload.source : payload);
    return source ? { status: "available", data: source } : { status: "unavailable", data: null };
  } catch {
    return { status: "unavailable", data: null };
  }
}

export async function getSourceCoverage(): Promise<SourceDataResult<SourceCoverageReport>> {
  const response = await fetchSourceAPI("/sources/coverage", 60);
  if (!response?.ok) return { status: "unavailable", data: null };
  try {
    const coverage = normalizeCoverage(await response.json());
    return coverage ? { status: "available", data: coverage } : { status: "unavailable", data: null };
  } catch {
    return { status: "unavailable", data: null };
  }
}
