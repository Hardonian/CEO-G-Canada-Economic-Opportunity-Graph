import projectsSnapshot from "@/data/projects.snapshot.json";
import manifestSnapshot from "@/data/manifest.snapshot.json";
import { Procurement, Project, ProjectEvidence, RadarStats, Signal } from "./types";

const SECTORS = new Set<Project["sector"]>([
  "Critical Minerals",
  "Nuclear & Clean Power",
  "Clean Energy & Grid",
  "AI Compute & Data Centres",
  "Defence & Arctic",
  "Transportation & Ports",
  "Industrial & Manufacturing",
  "Housing-Enabling Infrastructure",
  "Mining & Metals",
  "Energy & Fuels",
  "Forestry & Bioeconomy",
]);

const STAGES = new Set<Project["current_stage"]>([
  "UNKNOWN",
  "DISCOVERED",
  "ANNOUNCED",
  "REFERRED",
  "EARLY_DEVELOPMENT",
  "FEASIBILITY",
  "FINANCING",
  "ENVIRONMENTAL_REVIEW",
  "PERMITTING",
  "PROCUREMENT",
  "FID_LIKELY",
  "FID",
  "CONSTRUCTION",
  "COMMISSIONING",
  "OPERATING",
  "DELAYED",
  "PAUSED",
  "CANCELLED",
]);

const CONFIDENCE = new Set<Project["confidence"]>([
  "VERIFIED",
  "SUPPORTED",
  "REPORTED",
  "INFERRED",
  "CONFLICTED",
  "UNKNOWN",
  "STALE",
  "RETRACTED",
]);

const EXTERNAL_API_BASE = process.env.COG_API_BASE || process.env.NEXT_PUBLIC_API_BASE;

function configuredAPIBase(): string | null {
  if (!EXTERNAL_API_BASE) return null;
  try {
    const parsed = new URL(EXTERNAL_API_BASE);
    if (parsed.protocol !== "https:" && parsed.protocol !== "http:") return null;
    return parsed.toString().replace(/\/$/, "");
  } catch {
    console.error("[data] Ignoring invalid API base URL", { configured: true });
    return null;
  }
}

const API_BASE = configuredAPIBase();

function isRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function validPublicURL(value: unknown): value is string {
  if (typeof value !== "string") return false;
  try {
    const parsed = new URL(value);
    return (parsed.protocol === "https:" || parsed.protocol === "http:") && !parsed.username && !parsed.password;
  } catch {
    return false;
  }
}

function normalizeEvidence(value: unknown): ProjectEvidence | null {
  if (!isRecord(value)) return null;
  if (
    typeof value.id !== "string" ||
    !validPublicURL(value.source_url) ||
    typeof value.publisher !== "string" ||
    typeof value.source_tier !== "number" ||
    !Number.isInteger(value.source_tier) ||
    value.source_tier < 1 ||
    value.source_tier > 5 ||
    typeof value.retrieval_timestamp !== "string" ||
    Number.isNaN(Date.parse(value.retrieval_timestamp)) ||
    typeof value.confidence !== "string" ||
    typeof value.content_hash !== "string" ||
    !/^[a-f0-9]{64}$/i.test(value.content_hash)
  ) {
    return null;
  }
  return {
    id: value.id,
    source_url: value.source_url,
    publisher: value.publisher,
    source_tier: value.source_tier,
    retrieval_timestamp: value.retrieval_timestamp,
    effective_date: typeof value.effective_date === "string" ? value.effective_date : undefined,
    confidence: value.confidence,
    content_hash: value.content_hash,
    locator: typeof value.locator === "string" ? value.locator : undefined,
    source_record_id: typeof value.source_record_id === "string" ? value.source_record_id : undefined,
    pipeline_version: typeof value.pipeline_version === "string" ? value.pipeline_version : undefined,
  };
}

export function normalizeProject(value: unknown): Project | null {
  if (!isRecord(value)) return null;
  const raw = value;
  if (
    typeof raw.id !== "string" ||
    typeof raw.slug !== "string" ||
    typeof raw.name !== "string" ||
    typeof raw.summary !== "string" ||
    typeof raw.subsector !== "string" ||
    typeof raw.province !== "string" ||
    typeof raw.location_name !== "string" ||
    typeof raw.last_meaningful_update !== "string" ||
    Number.isNaN(Date.parse(raw.last_meaningful_update)) ||
    typeof raw.sector !== "string" ||
    !SECTORS.has(raw.sector as Project["sector"]) ||
    typeof raw.current_stage !== "string" ||
    !STAGES.has(raw.current_stage as Project["current_stage"]) ||
    typeof raw.confidence !== "string" ||
    !CONFIDENCE.has(raw.confidence as Project["confidence"]) ||
    typeof raw.capex_cad !== "number" ||
    !Number.isSafeInteger(raw.capex_cad) ||
    raw.capex_cad < 0
  ) {
    return null;
  }

  const scores = isRecord(raw.scores)
    ? (Object.fromEntries(
        Object.entries(raw.scores).filter(
          ([, score]) => typeof score === "number" && Number.isFinite(score) && score >= 0 && score <= 100,
        ),
      ) as Record<string, number>)
    : undefined;
  const coordinate = (candidate: unknown, min: number, max: number) =>
    typeof candidate === "number" && Number.isFinite(candidate) && candidate >= min && candidate <= max
      ? candidate
      : null;
  const capexStatus =
    typeof raw.capex_status === "string" && CONFIDENCE.has(raw.capex_status as Project["confidence"])
      ? (raw.capex_status as Project["confidence"])
      : undefined;
  const evidence = Array.isArray(raw.evidence)
    ? raw.evidence.flatMap((candidate) => {
        const item = normalizeEvidence(candidate);
        return item ? [item] : [];
      })
    : undefined;

  return {
    id: raw.id,
    slug: raw.slug,
    name: raw.name,
    summary: raw.summary,
    sector: raw.sector as Project["sector"],
    subsector: raw.subsector,
    province: raw.province,
    location_name: raw.location_name,
    latitude: coordinate(raw.latitude, 40, 84),
    longitude: coordinate(raw.longitude, -142, -50),
    current_stage: raw.current_stage as Project["current_stage"],
    capex_cad: raw.capex_cad,
    proponent_id: typeof raw.proponent_id === "string" ? raw.proponent_id : undefined,
    proponent_name: typeof raw.proponent_name === "string" ? raw.proponent_name : undefined,
    capex_status: capexStatus,
    confidence: raw.confidence as Project["confidence"],
    scores,
    last_meaningful_update: raw.last_meaningful_update,
    evidence,
  };
}

const normalizedSnapshotProjects = (projectsSnapshot as unknown[]).flatMap((candidate) => {
  const project = normalizeProject(candidate);
  return project ? [project] : [];
});

if (normalizedSnapshotProjects.length !== projectsSnapshot.length) {
  throw new Error("Bundled project snapshot failed validation");
}

export const SNAPSHOT_PROJECTS: Project[] = normalizedSnapshotProjects;
// Kept as a compatibility alias for client components that consume the bundled snapshot.
export const FALLBACK_PROJECTS = SNAPSHOT_PROJECTS;
export const SNAPSHOT_MANIFEST = manifestSnapshot;

function buildSnapshotStats(projects: Project[]): RadarStats {
  const sectorBreakdown: Record<string, number> = {};
  const provinceBreakdown: Record<string, number> = {};
  let totalCapex = 0;
  let unknownCapex = 0;
  for (const project of projects) {
    totalCapex += project.capex_cad;
    if (project.capex_cad === 0 || project.capex_status === "UNKNOWN") unknownCapex += 1;
    sectorBreakdown[project.sector] = (sectorBreakdown[project.sector] || 0) + project.capex_cad;
    provinceBreakdown[project.province] = (provinceBreakdown[project.province] || 0) + project.capex_cad;
  }
  return {
    total_projects: projects.length,
    total_capex_cad: totalCapex,
    capital_moving_week_cad: 0,
    accelerating_projects_count: projects.filter((project) => (project.scores?.buildability ?? 0) >= 50).length,
    stalled_projects_count: projects.filter((project) => project.current_stage === "DELAYED" || project.current_stage === "PAUSED").length,
    active_procurements_count: 0,
    unknown_capex_projects: unknownCapex,
    data_status: "HEALTHY",
    sector_breakdown: sectorBreakdown,
    province_breakdown: provinceBreakdown,
  };
}

export const SNAPSHOT_RADAR_STATS = buildSnapshotStats(SNAPSHOT_PROJECTS);
export const FALLBACK_RADAR_STATS = SNAPSHOT_RADAR_STATS;
export const FALLBACK_PROCUREMENTS: Procurement[] = [];
export const FALLBACK_SIGNALS: Signal[] = [];

export interface RadarData {
  stats: RadarStats;
  accelerating_projects: Project[];
  recent_signals: Signal[];
  cegs_version: string;
  source_mode: "LIVE_UPSTREAM_API" | "BUNDLED_REVIEWED_SNAPSHOT";
  generated_at?: string;
}

function normalizeNumberMap(value: unknown): Record<string, number> | null {
  if (!isRecord(value)) return null;
  const result: Record<string, number> = {};
  for (const [key, raw] of Object.entries(value)) {
    if (typeof raw !== "number" || !Number.isFinite(raw) || raw < 0) return null;
    result[key] = raw;
  }
  return result;
}

function normalizeRadarStats(value: unknown): RadarStats | null {
  if (!isRecord(value)) return null;
  const numericKeys = [
    "total_projects",
    "total_capex_cad",
    "capital_moving_week_cad",
    "accelerating_projects_count",
    "stalled_projects_count",
    "active_procurements_count",
  ] as const;
  for (const key of numericKeys) {
    if (typeof value[key] !== "number" || !Number.isFinite(value[key]) || value[key] < 0) return null;
  }
  const sectorBreakdown = normalizeNumberMap(value.sector_breakdown);
  const provinceBreakdown = normalizeNumberMap(value.province_breakdown);
  if (!sectorBreakdown || !provinceBreakdown) return null;
  return {
    total_projects: value.total_projects as number,
    total_capex_cad: value.total_capex_cad as number,
    capital_moving_week_cad: value.capital_moving_week_cad as number,
    accelerating_projects_count: value.accelerating_projects_count as number,
    stalled_projects_count: value.stalled_projects_count as number,
    active_procurements_count: value.active_procurements_count as number,
    unknown_capex_projects: typeof value.unknown_capex_projects === "number" ? value.unknown_capex_projects : undefined,
    data_status: typeof value.data_status === "string" ? value.data_status as RadarStats["data_status"] : undefined,
    sector_breakdown: sectorBreakdown,
    province_breakdown: provinceBreakdown,
  };
}

async function fetchExternalAPI(path: string): Promise<Response | null> {
  if (!API_BASE) return null;
  try {
    const response = await fetch(`${API_BASE}${path}`, {
      next: { revalidate: 300 },
      signal: AbortSignal.timeout(5_000),
      headers: { Accept: "application/json" },
    });
    if (!response.ok) {
      console.warn("[data] Upstream API returned a non-success response", { path, status: response.status });
    }
    return response;
  } catch (error) {
    console.error("[data] Upstream API request failed; serving vetted snapshot", {
      path,
      error: error instanceof Error ? error.message : String(error),
    });
    return null;
  }
}

export async function getRadarData(): Promise<RadarData> {
  const response = await fetchExternalAPI("/radar");
  if (response?.ok) {
    try {
      const data: unknown = await response.json();
      const stats = isRecord(data) ? normalizeRadarStats(data.stats) : null;
      if (stats) {
        return {
          stats,
          accelerating_projects: [],
          recent_signals: [],
          cegs_version: typeof data.cegs_version === "string" ? data.cegs_version : "0.1",
          source_mode: "LIVE_UPSTREAM_API",
          generated_at: typeof data.generated_at === "string" ? data.generated_at : undefined,
        };
      }
    } catch (error) {
      console.error("[data] Invalid radar response; serving vetted snapshot", {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }
  return {
    stats: SNAPSHOT_RADAR_STATS,
    accelerating_projects: SNAPSHOT_PROJECTS.filter((project) => (project.scores?.buildability ?? 0) >= 50),
    recent_signals: FALLBACK_SIGNALS,
    cegs_version: String(manifestSnapshot.cegs),
    source_mode: "BUNDLED_REVIEWED_SNAPSHOT",
    generated_at: String(manifestSnapshot.generated_at),
  };
}

export async function getProjects(): Promise<Project[]> {
  const response = await fetchExternalAPI("/projects?limit=500");
  if (response?.ok) {
    try {
      const data: unknown = await response.json();
      const candidates = isRecord(data) && Array.isArray(data.projects) ? data.projects : [];
      const projects = candidates.flatMap((candidate) => {
        const normalized = normalizeProject(candidate);
        return normalized ? [normalized] : [];
      });
      if (projects.length > 0) return projects;
    } catch (error) {
      console.error("[data] Invalid project response; serving vetted snapshot", {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }
  return SNAPSHOT_PROJECTS;
}

export function findSnapshotProject(idOrSlug: string): Project | null {
  return SNAPSHOT_PROJECTS.find((project) => project.slug === idOrSlug || project.id === idOrSlug) || null;
}

export async function getProjectBySlug(slug: string): Promise<Project | null> {
  const response = await fetchExternalAPI(`/projects/${encodeURIComponent(slug)}`);
  if (response?.ok) {
    try {
      const data: unknown = await response.json();
      const normalized = normalizeProject(isRecord(data) ? data.project : null);
      if (normalized) return normalized;
    } catch (error) {
      console.error("[data] Invalid project detail response; serving vetted snapshot", {
        slug,
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }
  return findSnapshotProject(slug);
}
