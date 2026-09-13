import { Procurement, Project, RadarStats, Signal } from "./types";

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
    if (Array.isArray(data?.projects)) return data.projects;
  }
  return FALLBACK_PROJECTS;
}

export async function getProjectBySlug(slug: string): Promise<Project | null> {
  const safeSlug = encodeURIComponent(slug);
  const response = await fetchAPI(`/projects/${safeSlug}`);
  if (response?.ok) {
    const data = await response.json();
    if (data?.project) return data.project;
  }
  return (
    FALLBACK_PROJECTS.find(
      (project) => project.slug === slug || project.id === slug,
    ) || null
  );
}
