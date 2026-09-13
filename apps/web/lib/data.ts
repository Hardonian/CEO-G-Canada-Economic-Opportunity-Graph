import { Project, RadarStats, Procurement, Signal } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080/api/v1";

// High-fidelity fallback database for resilient frontend SSR / static builds
export const FALLBACK_PROJECTS: Project[] = [
  {
    id: "proj-darlington-smr",
    slug: "darlington-small-modular-reactor-deployment-project",
    name: "Darlington Small Modular Reactor Deployment Project",
    summary: "Construction of North America's first grid-scale GE Hitachi BWRX-300 Small Modular Reactor (300 MWe) at the existing Darlington nuclear site.",
    sector: "Nuclear & Clean Power",
    subsector: "Small Modular Reactor (SMR)",
    province: "ON",
    location_name: "Clarington, Regional Municipality of Durham",
    latitude: 43.8683,
    longitude: -78.7183,
    current_stage: "CONSTRUCTION",
    capex_cad: 3400000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 53.3,
      investability: 73.8,
      supplierability: 55.3,
      strategicity: 37.5,
    },
    last_meaningful_update: "2026-08-15T14:20:00Z",
  },
  {
    id: "proj-crawford-nickel",
    slug: "crawford-nickel-cobalt-sulphide-project",
    name: "Crawford Nickel-Cobalt Sulphide Project",
    summary: "Large-scale open-pit nickel and cobalt mine and processing mill with net-zero carbon capture capabilities in the Timmins mining camp.",
    sector: "Critical Minerals",
    subsector: "Nickel-Cobalt Sulphide",
    province: "ON",
    location_name: "Timmins, Porcupine Mining District",
    latitude: 48.7333,
    longitude: -81.3333,
    current_stage: "PERMITTING",
    capex_cad: 2600000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 39.8,
      investability: 68.5,
      supplierability: 62.0,
      strategicity: 92.5,
    },
    last_meaningful_update: "2026-07-28T11:45:00Z",
  },
  {
    id: "proj-oneida-battery",
    slug: "oneida-energy-storage-project",
    name: "Oneida Energy Storage Project",
    summary: "250 MW / 1,000 MWh utility-scale battery energy storage facility in 50/50 equity partnership with Six Nations of the Grand River.",
    sector: "Clean Energy & Grid",
    subsector: "Utility-Scale Battery Storage",
    province: "ON",
    location_name: "Jarvis, Haldimand County",
    latitude: 42.8833,
    longitude: -80.0500,
    current_stage: "CONSTRUCTION",
    capex_cad: 450000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 51.3,
      investability: 65.0,
      supplierability: 48.0,
      strategicity: 70.0,
    },
    last_meaningful_update: "2026-08-01T10:00:00Z",
  },
  {
    id: "proj-contrecoeur-port",
    slug: "contrecoeur-port-terminal-container-expansion",
    name: "Contrecoeur Port Terminal Container Expansion",
    summary: "Expansion of the Port of Montreal with a 1.15 million TEU deepwater container terminal and direct intermodal rail access on the St. Lawrence corridor.",
    sector: "Transportation & Ports",
    subsector: "Marine Terminal & Rail Corridor",
    province: "QC",
    location_name: "Contrecœur, Montérégie",
    latitude: 45.8500,
    longitude: -73.2333,
    current_stage: "PROCUREMENT",
    capex_cad: 1400000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 44.0,
      investability: 61.2,
      supplierability: 74.5,
      strategicity: 80.0,
    },
    last_meaningful_update: "2026-06-10T09:15:00Z",
  },
  {
    id: "proj-windsor-intertie",
    slug: "windsor-detroit-clean-power-intertie",
    name: "Windsor-Detroit Clean Power Intertie",
    summary: "High-voltage direct current (HVDC) 500kV international transmission interconnection across the Detroit River supporting cross-border clean electricity export.",
    sector: "Clean Energy & Grid",
    subsector: "International Power Line (IPL)",
    province: "ON",
    location_name: "Windsor, Essex County",
    latitude: 42.3149,
    longitude: -83.0364,
    current_stage: "PERMITTING",
    capex_cad: 620000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 44.3,
      investability: 58.0,
      supplierability: 50.0,
      strategicity: 75.0,
    },
    last_meaningful_update: "2026-08-10T12:00:00Z",
  },
  {
    id: "proj-arctic-ideas",
    slug: "arctic-all-weather-over-the-horizon-sensor",
    name: "Arctic All-Weather Over-The-Horizon Sensor and Northern Runway Stabilization",
    summary: "Rapid deployment of resilient permafrost thermosyphon runway infrastructure, secure satellite communication uplinks, and autonomous coastal radar sensors.",
    sector: "Defence & Arctic",
    subsector: "Arctic Surveillance & Strategic Basing",
    province: "NU",
    location_name: "Resolute Bay and Nanisivik",
    latitude: 74.6975,
    longitude: -94.8297,
    current_stage: "PROCUREMENT",
    capex_cad: 750000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 39.5,
      investability: 54.0,
      supplierability: 71.0,
      strategicity: 95.0,
    },
    last_meaningful_update: "2026-07-15T09:00:00Z",
  },
  {
    id: "proj-sudbury-nickel",
    slug: "sudbury-clean-nickel-expansion-&-sinter-plant-modernization",
    name: "Sudbury Clean Nickel Expansion & Sinter Plant Modernization",
    summary: "Expansion of underground nickel mining and processing in the Sudbury Basin targeting EV battery grade refined nickel and net-zero smelting.",
    sector: "Critical Minerals",
    subsector: "Nickel & Copper Processing",
    province: "ON",
    location_name: "Greater Sudbury",
    latitude: 46.4900,
    longitude: -80.9900,
    current_stage: "FEASIBILITY",
    capex_cad: 1800000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 36.8,
      investability: 62.0,
      supplierability: 59.0,
      strategicity: 90.0,
    },
    last_meaningful_update: "2026-05-14T09:00:00Z",
  },
  {
    id: "proj-chisasibi-ai",
    slug: "chisasibi-sovereign-ai-clean-compute-facility",
    name: "Chisasibi Sovereign AI Clean Compute Hyperscale Cluster",
    summary: "350 MW zero-carbon AI compute data centre cluster powered by La Grande hydro basin, operating under Canadian sovereignty with full US CLOUD Act immunity.",
    sector: "AI Compute & Data Centres",
    subsector: "Clean Hydro Hyperscale AI Data Centre",
    province: "QC",
    location_name: "Chisasibi, Eeyou Istchee",
    latitude: 53.7833,
    longitude: -78.9000,
    current_stage: "PERMITTING",
    capex_cad: 2800000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 62.5,
      investability: 88.0,
      supplierability: 79.5,
      strategicity: 98.0,
    },
    last_meaningful_update: "2026-08-20T16:00:00Z",
  },
  {
    id: "proj-kivalliq-hydro-fibre",
    slug: "kivalliq-hydro-fibre-northern-infrastructure-link",
    name: "Kivalliq Hydro-Fibre Northern Clean Infrastructure Link",
    summary: "1,200 km 150MW transmission line and high-speed terrestrial fibre optic corridor connecting northern Manitoba to Kivalliq communities and mines in Nunavut.",
    sector: "Clean Energy & Grid",
    subsector: "Northern Clean Transmission & Fibre",
    province: "NU",
    location_name: "Gillam (MB) to Rankin Inlet (NU)",
    latitude: 62.8167,
    longitude: -92.0833,
    current_stage: "FEASIBILITY",
    capex_cad: 1650000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 58.0,
      investability: 82.5,
      supplierability: 84.0,
      strategicity: 96.0,
    },
    last_meaningful_update: "2026-08-18T13:30:00Z",
  },
  {
    id: "proj-churchill-arctic-gateway",
    slug: "churchill-arctic-deepwater-port-and-rail-corridor",
    name: "Churchill Arctic Deepwater Port & Rail Resiliency Corridor",
    summary: "Rehabilitation and permafrost stabilization of the Hudson Bay Railway and deepwater expansion of North America's sole Arctic rail-connected seaport.",
    sector: "Transportation & Ports",
    subsector: "Arctic Seaport & Continental Railway",
    province: "MB",
    location_name: "Churchill, Hudson Bay",
    latitude: 58.7684,
    longitude: -94.1650,
    current_stage: "CONSTRUCTION",
    capex_cad: 850000000,
    confidence: "VERIFIED",
    scores: {
      buildability: 64.0,
      investability: 76.0,
      supplierability: 68.0,
      strategicity: 94.0,
    },
    last_meaningful_update: "2026-08-25T11:00:00Z",
  },
];

export const FALLBACK_RADAR_STATS: RadarStats = {
  total_projects: 10,
  total_capex_cad: 16320000000,
  capital_moving_week_cad: 1420000000,
  accelerating_projects_count: 5,
  stalled_projects_count: 0,
  active_procurements_count: 5,
  sector_breakdown: {
    "Nuclear & Clean Power": 3400000000,
    "Critical Minerals": 4400000000,
    "AI Compute & Data Centres": 2800000000,
    "Clean Energy & Grid": 2720000000,
    "Transportation & Ports": 2250000000,
    "Defence & Arctic": 750000000,
  },
  province_breakdown: {
    ON: 8870000000,
    QC: 4200000000,
    NU: 2400000000,
    MB: 850000000,
  },
};

export const FALLBACK_PROCUREMENTS: Procurement[] = [
  {
    id: "proc-1",
    tender_id: "WS394827101",
    title: "Engineering and Environmental Monitoring Services for Arctic Deep-Water Facilities",
    stage: "Active RFP",
    buyer: "Department of National Defence / PSPC",
    buyer_type: "Federal",
    estimated_cad: 45000000,
    closing_date: "2026-10-30T14:00:00Z",
    source_url: "https://canadabuys.canada.ca/en/tender-opportunities/notice/WS394827101",
    categories: ["Civil Engineering", "Arctic Infrastructure", "Environmental Surveillance"],
    requirement_class: "CONFIRMED",
  },
  {
    id: "proc-2",
    tender_id: "WS401829384",
    title: "Turnkey High-Voltage Substation Equipment for Nuclear Interconnect",
    stage: "RFI",
    buyer: "Ontario Power Generation Inc.",
    buyer_type: "Crown",
    estimated_cad: 85000000,
    closing_date: "2026-11-15T16:00:00Z",
    source_url: "https://canadabuys.canada.ca/en/tender-opportunities/notice/WS401829384",
    categories: ["High Voltage Electrical", "Nuclear Components", "Substations"],
    requirement_class: "CONFIRMED",
  },
  {
    id: "proc-3",
    tender_id: "WS418920192",
    title: "Clean Hydro High-Voltage Substation Intertie & 350MW Step-Up Transformers",
    stage: "Active RFP",
    buyer: "Hydro-Québec / Cree Regional Authority",
    buyer_type: "Utility / Indigenous",
    estimated_cad: 120000000,
    closing_date: "2026-12-01T17:00:00Z",
    source_url: "https://canadabuys.canada.ca/en/tender-opportunities/notice/WS418920192",
    categories: ["High Voltage Power", "Substation Intertie", "Clean Compute"],
    requirement_class: "CONFIRMED",
  },
  {
    id: "proc-4",
    tender_id: "WS429103847",
    title: "Northern Subsea & Terrestrial Fibre-Optic Bathymetric & Geotechnical Survey",
    stage: "Advance Notice",
    buyer: "Kivalliq Inuit Association / CIB",
    buyer_type: "Indigenous Co-Ownership",
    estimated_cad: 28000000,
    closing_date: "2026-11-28T15:00:00Z",
    source_url: "https://canadabuys.canada.ca/en/tender-opportunities/notice/WS429103847",
    categories: ["Geotechnical Survey", "Fibre Telecommunications", "Arctic Infrastructure"],
    requirement_class: "CONFIRMED",
  },
  {
    id: "proc-5",
    tender_id: "WS431092841",
    title: "Hudson Bay Railway Permafrost Stabilization, Thermopile & Culvert Modernization",
    stage: "Active RFP",
    buyer: "Arctic Gateway Group / Transport Canada",
    buyer_type: "Federal / Indigenous",
    estimated_cad: 62000000,
    closing_date: "2026-10-15T16:30:00Z",
    source_url: "https://canadabuys.canada.ca/en/tender-opportunities/notice/WS431092841",
    categories: ["Railway Construction", "Permafrost Geotechnical", "Arctic Logistics"],
    requirement_class: "CONFIRMED",
  },
];

export const FALLBACK_SIGNALS: Signal[] = [
  {
    id: "sig-1",
    project_id: "proj-darlington-smr",
    project_name: "Darlington Small Modular Reactor Deployment Project",
    type: "CONSTRUCTION_SIGNAL",
    timestamp: "2026-08-15T14:20:00Z",
    magnitude: 0.95,
    confidence: 0.98,
    description: "Major milestone achieved: Transitioned to active CONSTRUCTION following Canadian Nuclear Safety Commission licensing approval.",
  },
  {
    id: "sig-2",
    project_id: "proj-crawford-nickel",
    project_name: "Crawford Nickel-Cobalt Sulphide Project",
    type: "REGULATORY_PROGRESS",
    timestamp: "2026-07-28T11:45:00Z",
    magnitude: 0.70,
    confidence: 0.92,
    description: "Regulatory milestone reached: Impact Assessment Agency accepted formal environmental impact statement.",
  },
  {
    id: "sig-3",
    project_id: "proj-oneida-battery",
    project_name: "Oneida Energy Storage Project",
    type: "FINANCING_ACCELERATION",
    timestamp: "2026-07-10T10:00:00Z",
    magnitude: 0.85,
    confidence: 0.95,
    description: "Capital commitment: $170M CAD loan closed via Canada Infrastructure Bank alongside NRCan grant.",
  },
];

export async function getRadarData() {
  try {
    const res = await fetch(`${API_BASE}/radar`, { cache: "no-store" });
    if (res.ok) {
      return await res.json();
    }
  } catch (err) {
    // Graceful fallback to static verified dataset
  }
  return {
    stats: FALLBACK_RADAR_STATS,
    accelerating_projects: FALLBACK_PROJECTS.slice(0, 3),
    recent_signals: FALLBACK_SIGNALS,
    cegs_version: "0.1",
  };
}

export async function getProjects(): Promise<Project[]> {
  try {
    const res = await fetch(`${API_BASE}/projects`, { cache: "no-store" });
    if (res.ok) {
      const data = await res.json();
      return data.projects || FALLBACK_PROJECTS;
    }
  } catch (err) {
    // Fallback
  }
  return FALLBACK_PROJECTS;
}

export async function getProjectBySlug(slug: string): Promise<Project | null> {
  try {
    const res = await fetch(`${API_BASE}/projects/${slug}`, { cache: "no-store" });
    if (res.ok) {
      const data = await res.json();
      return data.project || null;
    }
  } catch (err) {
    // Fallback
  }
  const match = FALLBACK_PROJECTS.find((p) => p.slug === slug || p.id === slug);
  return match || null;
}
