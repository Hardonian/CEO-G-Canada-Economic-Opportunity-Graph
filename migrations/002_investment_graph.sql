-- Migration 002: Investment Opportunity Graph
-- Extends the base schema with the domain types needed for the investment
-- intelligence layer: opportunities funnel, capital needs, capacity metrics,
-- FID tracking, project requirements, and investor profiles.

-- Strategic themes and multi-sector support on projects
ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS secondary_sectors JSONB DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS strategic_themes JSONB DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS finance_type TEXT DEFAULT '',
  ADD COLUMN IF NOT EXISTS revenue_model TEXT DEFAULT '',
  ADD COLUMN IF NOT EXISTS tech_maturity TEXT DEFAULT '';

-- Opportunity funnel enrichment
ALTER TABLE opportunities
  ADD COLUMN IF NOT EXISTS funnel_state TEXT DEFAULT 'DISCOVERED',
  ADD COLUMN IF NOT EXISTS estimated_amount JSONB,
  ADD COLUMN IF NOT EXISTS evidence_quality TEXT DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_verified TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();

-- Capacity metrics: typed, evidence-linked measurements of project output
CREATE TABLE IF NOT EXISTS capacity_metrics (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id),
  phase_id TEXT DEFAULT '',
  metric_type TEXT NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  unit TEXT NOT NULL,
  time_basis TEXT DEFAULT '',
  status TEXT NOT NULL DEFAULT 'PLANNED',
  confidence TEXT NOT NULL DEFAULT 'UNKNOWN',
  evidence_ids JSONB DEFAULT '[]'::jsonb,
  visibility TEXT NOT NULL DEFAULT 'PUBLIC',
  publishable BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_capacity_metrics_project ON capacity_metrics(project_id);

-- Economic output observations
CREATE TABLE IF NOT EXISTS economic_outputs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id),
  phase_id TEXT DEFAULT '',
  output_type TEXT NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  unit TEXT NOT NULL DEFAULT '',
  period TEXT DEFAULT '',
  status TEXT NOT NULL DEFAULT 'UNKNOWN',
  evidence_ids JSONB DEFAULT '[]'::jsonb,
  visibility TEXT NOT NULL DEFAULT 'PUBLIC',
  publishable BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_economic_outputs_project ON economic_outputs(project_id);

-- FID intelligence tracking
CREATE TABLE IF NOT EXISTS fid_intelligence (
  project_id TEXT PRIMARY KEY REFERENCES projects(id),
  status TEXT NOT NULL DEFAULT 'UNKNOWN',
  target_date TIMESTAMPTZ,
  earliest_date TIMESTAMPTZ,
  latest_date TIMESTAMPTZ,
  actual_date TIMESTAMPTZ,
  confidence TEXT NOT NULL DEFAULT 'UNKNOWN',
  revisions JSONB DEFAULT '[]'::jsonb,
  evidence_ids JSONB DEFAULT '[]'::jsonb,
  visibility TEXT NOT NULL DEFAULT 'PUBLIC',
  publishable BOOLEAN NOT NULL DEFAULT true,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Project requirements (infrastructure dependency graph)
CREATE TABLE IF NOT EXISTS project_requirements (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id),
  type TEXT NOT NULL,
  description TEXT DEFAULT '',
  capacity_needed JSONB,
  confidence TEXT NOT NULL DEFAULT 'ONTOLOGY_DERIVED',
  satisfied_by TEXT DEFAULT '',
  evidence_ids JSONB DEFAULT '[]'::jsonb,
  visibility TEXT NOT NULL DEFAULT 'PUBLIC',
  publishable BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_project_requirements_project ON project_requirements(project_id);
CREATE INDEX IF NOT EXISTS idx_project_requirements_type ON project_requirements(type);

-- Investor profiles (entity extension)
CREATE TABLE IF NOT EXISTS investor_profiles (
  entity_id TEXT PRIMARY KEY REFERENCES entities(id),
  investor_types JSONB DEFAULT '[]'::jsonb,
  target_sectors JSONB DEFAULT '[]'::jsonb,
  target_geographies JSONB DEFAULT '[]'::jsonb,
  min_ticket_cad BIGINT DEFAULT 0,
  max_ticket_cad BIGINT DEFAULT 0,
  preferred_instruments JSONB DEFAULT '[]'::jsonb,
  preferred_stages JSONB DEFAULT '[]'::jsonb,
  canadian_exposure_cad BIGINT DEFAULT 0,
  active_investments INTEGER DEFAULT 0,
  public_deal_history JSONB DEFAULT '[]'::jsonb,
  evidence_ids JSONB DEFAULT '[]'::jsonb,
  visibility TEXT NOT NULL DEFAULT 'PUBLIC',
  publishable BOOLEAN NOT NULL DEFAULT true,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Velocity scores
CREATE TABLE IF NOT EXISTS velocity_scores (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id),
  velocity_type TEXT NOT NULL,
  score DOUBLE PRECISION NOT NULL DEFAULT 0,
  trend TEXT NOT NULL DEFAULT 'STALLED',
  methodology_version TEXT NOT NULL,
  factors JSONB DEFAULT '{}'::jsonb,
  factor_evidence JSONB DEFAULT '{}'::jsonb,
  explanation JSONB DEFAULT '[]'::jsonb,
  input_hash TEXT NOT NULL DEFAULT '',
  calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_velocity_scores_project ON velocity_scores(project_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_velocity_scores_unique ON velocity_scores(project_id, velocity_type);
