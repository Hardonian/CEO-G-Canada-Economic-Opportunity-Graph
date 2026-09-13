import { Procurement, Project, RadarStats, Signal } from "./types";

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

const API_BASE =
  process.env.COG_API_BASE ||
  process.env.NEXT_PUBLIC_API_BASE ||
  "http://localhost:8080/api/v1";

// Offline fallback records are intentionally limited to the human-reviewed
// primary-source snapshot. They must never contain demo projects or pretend
// that a generated procurement is confirmed. The running API enriches these
// records with the nationwide NRCan MPI snapshot.
export const FALLBACK_PROJECTS: Project[] = [
  {
    id: "snapshot:cnsc:darlington-new-nuclear-project",
    slug: "darlington-new-nuclear-project-unit-1",
    name: "Darlington New Nuclear Project — Unit 1",
    summary:
      "Ontario Power Generation is constructing one BWRX-300 reactor at the Darlington New Nuclear Project site in Clarington, Ontario.",
    sector: "Nuclear & Clean Power",
    subsector: "Small Modular Reactor",
    province: "ON",
    location_name: "Darlington site, Clarington",
    latitude: null,
    longitude: null,
    current_stage: "CONSTRUCTION",
    capex_cad: 7_700_000_000,
    capex_status: "REPORTED",
    confidence: "SUPPORTED",
    scores: { buildability: 96.3 },
    last_meaningful_update: "2026-03-30T00:00:00Z",
  },
  {
    id: "snapshot:iaac:83857",
    slug: "crawford-nickel-project",
    name: "Crawford Nickel Project",
    summary:
      "Canada Nickel Company proposes an open-pit nickel-cobalt mine and on-site metal mill approximately 42 kilometres north of Timmins, Ontario.",
    sector: "Critical Minerals",
    subsector: "Nickel-Cobalt Mine and Mill",
    province: "ON",
    location_name: "Approximately 42 km north of Timmins",
    latitude: null,
    longitude: null,
    current_stage: "PERMITTING",
    capex_cad: 0,
    capex_status: "UNKNOWN",
    confidence: "VERIFIED",
    scores: { buildability: 72.5 },
    last_meaningful_update: "2026-07-31T00:00:00Z",
  },
  {
    id: "snapshot:iaac:80116",
    slug: "contrecoeur-port-terminal-expansion-project",
    name: "Contrecœur Port Terminal Expansion Project",
    summary:
      "The Montreal Port Authority proposes a container terminal in Contrecœur with a maximum annual capacity of 1.15 million twenty-foot-equivalent containers.",
    sector: "Transportation & Ports",
    subsector: "Container Port Terminal",
    province: "QC",
    location_name: "Contrecœur, approximately 40 km northeast of Montreal",
    latitude: null,
    longitude: null,
    current_stage: "UNKNOWN",
    capex_cad: 0,
    capex_status: "UNKNOWN",
    confidence: "VERIFIED",
    scores: { buildability: 0 },
    last_meaningful_update: "2026-09-04T00:00:00Z",
  },
  {
    id: "snapshot:oneida-energy-storage",
    slug: "oneida-energy-storage-project",
    name: "Oneida Energy Storage Project",
    summary:
      "Oneida is a 250 MW / 1,000 MWh battery energy storage facility in Haldimand County, Ontario that entered commercial operations in 2025.",
    sector: "Clean Energy & Grid",
    subsector: "Battery Energy Storage",
    province: "ON",
    location_name: "Haldimand County",
    latitude: null,
    longitude: null,
    current_stage: "OPERATING",
    capex_cad: 0,
    capex_status: "UNKNOWN",
    confidence: "SUPPORTED",
    scores: { buildability: 100 },
    last_meaningful_update: "2025-05-07T00:00:00Z",
  },
];

export const FALLBACK_RADAR_STATS: RadarStats = {
  total_projects: FALLBACK_PROJECTS.length,
  total_capex_cad: 7_700_000_000,
  capital_moving_week_cad: 0,
  accelerating_projects_count: 0,
  stalled_projects_count: 0,
  active_procurements_count: 0,
  unknown_capex_projects: 3,
  data_status: "PARTIAL",
  sector_breakdown: { "Nuclear & Clean Power": 7_700_000_000 },
  province_breakdown: { ON: 7_700_000_000 },
};

// No checked-in tender is currently part of the reviewed snapshot. Keeping
// this empty is safer than presenting illustrative identifiers as confirmed.
export const FALLBACK_PROCUREMENTS: Procurement[] = [];
export const FALLBACK_SIGNALS: Signal[] = [];

async function fetchAPI(path: string): Promise<Response | null> {
  try {
    return await fetch(`${API_BASE}${path}`, {
      next: { revalidate: 300 },
      signal: AbortSignal.timeout(5_000),
      headers: { Accept: "application/json" },
    });
  } catch {
    return null;
  }
}

function normalizeProject(value: unknown): Project | null {
  if (!value || typeof value !== "object") return null;
  const raw = value as Record<string, unknown>;
  if (
    typeof raw.id !== "string" ||
    typeof raw.slug !== "string" ||
    typeof raw.name !== "string" ||
    typeof raw.summary !== "string" ||
    typeof raw.subsector !== "string" ||
    typeof raw.province !== "string" ||
    typeof raw.location_name !== "string" ||
    typeof raw.last_meaningful_update !== "string" ||
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

  const scores = raw.scores && typeof raw.scores === "object"
    ? Object.fromEntries(
        Object.entries(raw.scores as Record<string, unknown>).filter(
          ([, score]) => typeof score === "number" && Number.isFinite(score) && score >= 0 && score <= 100,
        ),
      )
    : undefined;
  const coordinate = (candidate: unknown, min: number, max: number) =>
    typeof candidate === "number" && Number.isFinite(candidate) && candidate >= min && candidate <= max
      ? candidate
      : null;
  const capexStatus =
    typeof raw.capex_status === "string" && CONFIDENCE.has(raw.capex_status as Project["confidence"])
      ? (raw.capex_status as Project["confidence"])
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
    capex_status: capexStatus,
    confidence: raw.confidence as Project["confidence"],
    scores,
    last_meaningful_update: raw.last_meaningful_update,
  };
}

export async function getRadarData() {
  const response = await fetchAPI("/radar");
  if (response?.ok) {
    const data = await response.json();
    if (data && typeof data === "object" && data.stats) return data;
  }
  return {
    stats: FALLBACK_RADAR_STATS,
    accelerating_projects: [],
    recent_signals: FALLBACK_SIGNALS,
    cegs_version: "0.1",
    source_mode: "REVIEWED_OFFLINE_FALLBACK",
  };
}

export async function getProjects(): Promise<Project[]> {
  const response = await fetchAPI("/projects?limit=500");
  if (response?.ok) {
    const data = await response.json();
    if (Array.isArray(data?.projects)) {
      const projects = data.projects.flatMap((project: unknown) => {
        const normalized = normalizeProject(project);
        return normalized ? [normalized] : [];
      });
      if (projects.length > 0) return projects;
    }
  }
  return FALLBACK_PROJECTS;
}

export async function getProjectBySlug(slug: string): Promise<Project | null> {
  const safeSlug = encodeURIComponent(slug);
  const response = await fetchAPI(`/projects/${safeSlug}`);
  if (response?.ok) {
    const data = await response.json();
    const project = normalizeProject(data?.project);
    if (project) return project;
  }
  return (
    FALLBACK_PROJECTS.find(
      (project) => project.slug === slug || project.id === slug,
    ) || null
  );
}
