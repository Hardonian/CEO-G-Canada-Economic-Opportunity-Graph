export type Sector =
  | "Critical Minerals"
  | "Nuclear & Clean Power"
  | "Clean Energy & Grid"
  | "AI Compute & Data Centres"
  | "Defence & Arctic"
  | "Transportation & Ports"
  | "Industrial & Manufacturing"
  | "Housing-Enabling Infrastructure"
  | "Mining & Metals"
  | "Energy & Fuels"
  | "Forestry & Bioeconomy";

export type LifecycleStage =
  | "DISCOVERED"
  | "ANNOUNCED"
  | "REFERRED"
  | "EARLY_DEVELOPMENT"
  | "FEASIBILITY"
  | "FINANCING"
  | "ENVIRONMENTAL_REVIEW"
  | "PERMITTING"
  | "PROCUREMENT"
  | "FID_LIKELY"
  | "FID"
  | "CONSTRUCTION"
  | "COMMISSIONING"
  | "OPERATING"
  | "DELAYED"
  | "PAUSED"
  | "CANCELLED";

export interface Project {
  id: string;
  slug: string;
  name: string;
  summary: string;
  sector: Sector;
  subsector: string;
  province: string;
  location_name: string;
  latitude: number | null;
  longitude: number | null;
  current_stage: LifecycleStage;
  capex_cad: number;
  proponent_id?: string;
  capex_status?: "VERIFIED" | "SUPPORTED" | "REPORTED" | "INFERRED" | "CONFLICTED" | "UNKNOWN" | "STALE" | "RETRACTED";
  confidence: "VERIFIED" | "SUPPORTED" | "REPORTED" | "INFERRED" | "CONFLICTED" | "UNKNOWN" | "STALE" | "RETRACTED";
  scores?: Record<string, number>;
  last_meaningful_update: string;
}

export interface ProjectScore {
  score_type: string;
  score_value: number;
  score_version: string;
  factors: Record<string, number>;
  explanation: string;
}

export interface CapitalItem {
  id: string;
  category: string;
  status: string;
  amount_cad: number;
  provider_name: string;
  notes?: string;
}

export interface Event {
  id: string;
  event_type: string;
  event_date: string;
  title: string;
  description: string;
  evidence_id?: string;
}

export interface Opportunity {
  id: string;
  title: string;
  sector: Sector;
  requirement_class: "CONFIRMED" | "DERIVED" | "SPECULATIVE";
  category: string;
  estimated_cad: number;
  description: string;
  trigger_milestone: string;
}

export interface Procurement {
  id: string;
  tender_id: string;
  title: string;
  stage: string;
  buyer: string;
  buyer_type: string;
  estimated_cad?: number;
  closing_date?: string;
  source_url: string;
  categories: string[];
  requirement_class: string;
}

export interface RadarStats {
  total_projects: number;
  total_capex_cad: number;
  capital_moving_week_cad: number;
  accelerating_projects_count: number;
  stalled_projects_count: number;
  active_procurements_count: number;
  unknown_capex_projects?: number;
  data_status?: "HEALTHY" | "STALE" | "PARTIAL" | "DEGRADED" | "UNAVAILABLE";
  sector_breakdown: Record<string, number>;
  province_breakdown: Record<string, number>;
}

export interface Signal {
  id: string;
  project_id: string;
  project_name: string;
  type: string;
  timestamp: string;
  magnitude: number;
  confidence: number;
  description: string;
}
