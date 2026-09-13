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
  | "UNKNOWN"
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

export type SourceLifecycle =
  | "DISCOVERED"
  | "CLASSIFIED"
  | "TESTED"
  | "APPROVED"
  | "ACTIVE"
  | "REJECTED"
  | "BLOCKED"
  | "RETIRED";

export type SourceHealthStatus =
  | "CURRENT"
  | "HEALTHY"
  | "DELAYED"
  | "STALE"
  | "DEGRADED"
  | "BROKEN"
  | "UNAVAILABLE"
  | "DISABLED"
  | "UNKNOWN"
  | "NOT_YET_CHECKED";

export interface SourceQuality {
  structuredness?: number;
  freshness?: number;
  completeness?: number;
  stability?: number;
  authority?: number;
  historical_depth?: number;
}

export interface PublicSource {
  id: string;
  name: string;
  publisher_id: string;
  publisher_name: string;
  canonical_url: string;
  jurisdiction: string;
  geography: string[];
  source_family: string;
  access_method: string;
  content_type: string;
  authority_tier: number;
  subject_tags: string[];
  sector_tags: string[];
  languages: string[];
  update_frequency: string;
  lifecycle: SourceLifecycle;
  health: SourceHealthStatus;
  last_checked_at?: string;
  last_success_at?: string;
  last_change_at?: string;
  license: string;
  quality: SourceQuality;
  coverage_class: string;
  description: string;
}

export interface SourceListResponse {
  sources: PublicSource[];
  total: number;
  limit: number;
  offset: number;
}

export interface MeasuredRatio {
  status: string;
  value?: number;
}

export interface SignalLatencySummary {
  status: string;
  p50_ms?: number;
  p95_ms?: number;
  sample_count?: number;
}

export interface DeadLetterSummary {
  status: string;
  count?: number;
}

export interface SourceCoverageReport {
  generated_at: string;
  lifecycle_counts: {
    discovered: number;
    registered: number;
    tested: number;
    active: number;
    broken: number;
  };
  by_jurisdiction: Record<string, number>;
  by_sector: Record<string, number>;
  by_family: Record<string, number>;
  primary_source_ratio: MeasuredRatio;
  signal_latency: SignalLatencySummary;
  dead_letters: DeadLetterSummary;
  blind_spots: string[];
}
