-- CanadaOpportunityGraph Initial PostgreSQL Schema
-- Forward-only migration 001_initial_schema.sql

CREATE TABLE IF NOT EXISTS evidence (
    id VARCHAR(64) PRIMARY KEY,
    source_url TEXT NOT NULL,
    publisher VARCHAR(255) NOT NULL,
    source_tier INT NOT NULL CHECK (source_tier BETWEEN 1 AND 4),
    retrieval_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    publication_date TIMESTAMPTZ,
    effective_date TIMESTAMPTZ,
    confidence VARCHAR(32) NOT NULL,
    extraction_method VARCHAR(128) NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    raw_snippet TEXT
);

CREATE INDEX IF NOT EXISTS idx_evidence_hash ON evidence(content_hash);

CREATE TABLE IF NOT EXISTS entities (
    id VARCHAR(64) PRIMARY KEY,
    slug VARCHAR(128) UNIQUE NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    common_name VARCHAR(255) NOT NULL,
    aliases JSONB NOT NULL DEFAULT '[]'::jsonb,
    entity_type VARCHAR(64) NOT NULL,
    jurisdiction VARCHAR(32) NOT NULL,
    website TEXT,
    identifiers JSONB DEFAULT '{}'::jsonb,
    description TEXT,
    evidence_id VARCHAR(64) REFERENCES evidence(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_entities_slug ON entities(slug);
CREATE INDEX IF NOT EXISTS idx_entities_legal_name ON entities(legal_name);

CREATE TABLE IF NOT EXISTS projects (
    id VARCHAR(64) PRIMARY KEY,
    slug VARCHAR(128) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    summary TEXT NOT NULL,
    sector VARCHAR(64) NOT NULL,
    subsector VARCHAR(128) NOT NULL,
    province VARCHAR(32) NOT NULL,
    location_name VARCHAR(255) NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    current_stage VARCHAR(64) NOT NULL,
    capex_cad BIGINT NOT NULL,
    proponent_id VARCHAR(64) REFERENCES entities(id),
    confidence VARCHAR(32) NOT NULL,
    is_synthetic BOOLEAN NOT NULL DEFAULT FALSE,
    last_meaningful_update TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    scores JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_slug ON projects(slug);
CREATE INDEX IF NOT EXISTS idx_projects_sector ON projects(sector);
CREATE INDEX IF NOT EXISTS idx_projects_province ON projects(province);
CREATE INDEX IF NOT EXISTS idx_projects_stage ON projects(current_stage);
CREATE INDEX IF NOT EXISTS idx_projects_capex ON projects(capex_cad);

-- Full-text search index for projects
CREATE INDEX IF NOT EXISTS idx_projects_search ON projects USING GIN (to_tsvector('english', name || ' ' || summary || ' ' || location_name || ' ' || subsector));

CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_type VARCHAR(64) NOT NULL,
    event_date TIMESTAMPTZ NOT NULL,
    previous_stage VARCHAR(64),
    new_stage VARCHAR(64),
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    evidence_id VARCHAR(64) NOT NULL REFERENCES evidence(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_project ON events(project_id, event_date DESC);

CREATE TABLE IF NOT EXISTS relationships (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_entity_id VARCHAR(64) NOT NULL REFERENCES entities(id),
    target_entity_id VARCHAR(64) NOT NULL REFERENCES entities(id),
    relation_type VARCHAR(64) NOT NULL,
    confidence VARCHAR(32) NOT NULL,
    evidence_id VARCHAR(64) REFERENCES evidence(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS project_scores (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    score_type VARCHAR(64) NOT NULL,
    score_value DOUBLE PRECISION NOT NULL,
    score_version VARCHAR(32) NOT NULL,
    factors JSONB NOT NULL DEFAULT '{}'::jsonb,
    explanation TEXT NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_project_scores ON project_scores(project_id, score_type, calculated_at DESC);

CREATE TABLE IF NOT EXISTS capital_items (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category VARCHAR(64) NOT NULL,
    status VARCHAR(64) NOT NULL,
    amount_cad BIGINT NOT NULL,
    provider_entity_id VARCHAR(64) REFERENCES entities(id),
    provider_name VARCHAR(255) NOT NULL,
    notes TEXT,
    evidence_id VARCHAR(64) NOT NULL REFERENCES evidence(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS procurements (
    id VARCHAR(64) PRIMARY KEY,
    tender_id VARCHAR(128) UNIQUE NOT NULL,
    project_id VARCHAR(64) REFERENCES projects(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    stage VARCHAR(64) NOT NULL,
    closing_date TIMESTAMPTZ,
    estimated_cad BIGINT,
    buyer VARCHAR(255) NOT NULL,
    buyer_type VARCHAR(64) NOT NULL,
    source_url TEXT NOT NULL,
    categories JSONB NOT NULL DEFAULT '[]'::jsonb,
    requirement_class VARCHAR(32) NOT NULL,
    evidence_id VARCHAR(64) REFERENCES evidence(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS opportunities (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    sector VARCHAR(64) NOT NULL,
    requirement_class VARCHAR(32) NOT NULL,
    category VARCHAR(64) NOT NULL,
    estimated_cad BIGINT,
    description TEXT NOT NULL,
    trigger_milestone VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS signals (
    id VARCHAR(64) PRIMARY KEY,
    project_id VARCHAR(64) NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(64) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    magnitude DOUBLE PRECISION NOT NULL,
    confidence DOUBLE PRECISION NOT NULL,
    previous_state VARCHAR(64),
    new_state VARCHAR(64),
    description TEXT NOT NULL,
    evidence_id VARCHAR(64) REFERENCES evidence(id)
);
