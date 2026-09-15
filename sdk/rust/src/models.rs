use serde::{Deserialize, Serialize};

// ── Monetary ─────────────────────────────────────────────────────────────────

/// A CEGS monetary value in Canadian dollars.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Monetary {
    /// Amount in whole CAD.
    pub amount: Option<i64>,
    /// ISO 4217 currency code — always "CAD" in CEGS.
    pub currency: String,
    /// How the amount was determined: "reported", "estimated", "range", "unknown".
    pub amount_type: String,
    /// Optional range when amount_type is "range".
    pub range: Option<CapitalRange>,
}

/// A CAD range for capital estimates.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapitalRange {
    pub min: i64,
    pub max: i64,
}

// ── Location ─────────────────────────────────────────────────────────────────

/// Standardized geographic location.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Location {
    pub name: String,
    pub province: String,
    pub latitude: Option<f64>,
    pub longitude: Option<f64>,
}

// ── Project ──────────────────────────────────────────────────────────────────

/// A CanadaOpportunityGraph project.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Project {
    pub id: String,
    pub name: String,
    pub slug: String,
    pub summary: Option<String>,
    pub sector: String,
    pub province: String,
    pub current_stage: String,
    pub capex_cad: i64,
    pub proponent_id: Option<String>,
    pub scores: Option<std::collections::HashMap<String, f64>>,
    pub created_at: String,
    pub updated_at: String,
}

/// Wrapper returned by list-projects endpoints.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProjectList {
    pub projects: Vec<Project>,
    pub total: Option<i64>,
    pub limit: Option<i32>,
    pub offset: Option<i32>,
}

// ── ScoreBundle ──────────────────────────────────────────────────────────────

/// All current scores for a project.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScoreBundle {
    pub project_id: String,
    pub scores: std::collections::HashMap<String, ScoreDetail>,
}

/// A single scored dimension.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ScoreDetail {
    pub score: f64,
    pub confidence: f64,
    pub version: Option<String>,
    pub calculated_at: Option<String>,
}

// ── CapitalStack ─────────────────────────────────────────────────────────────

/// The resolved capital stack for a project.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapitalStack {
    pub project_id: String,
    pub total_available_cad: Option<i64>,
    pub programs: Vec<CapitalProgram>,
}

/// A single program in the capital stack.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CapitalProgram {
    pub program_name: String,
    pub program_id: Option<String>,
    pub category: String,
    pub max_amount_cad: Option<i64>,
    pub eligible: bool,
    pub notes: Option<String>,
}

// ── Signal ───────────────────────────────────────────────────────────────────

/// An economic momentum signal.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Signal {
    pub id: String,
    pub project_id: String,
    pub signal_type: String,
    pub magnitude: f64,
    pub confidence: f64,
    pub description: String,
    pub timestamp: String,
}

/// Wrapper returned by list-signals endpoints.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SignalList {
    pub signals: Vec<Signal>,
}

// ── CEGS Envelope ─────────────────────────────────────────────────────────────

/// A generic CEGS-compliant resource envelope.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CegsEnvelope {
    pub cegs: String,
    pub id: String,
    #[serde(rename = "type")]
    pub resource_type: String,
    pub canonical_name: Option<String>,
    pub jurisdiction: Option<String>,
    pub created_at: String,
    pub updated_at: String,
    pub provenance: Option<Vec<String>>,
    #[serde(flatten)]
    pub extra: std::collections::HashMap<String, serde_json::Value>,
}

// ── Organization ─────────────────────────────────────────────────────────────

/// A CanadaOpportunityGraph organization (company, government body, etc.).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Organization {
    pub id: String,
    pub slug: String,
    pub common_name: String,
    pub legal_name: Option<String>,
    pub entity_type: String,
    pub jurisdiction: Option<String>,
    pub website: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}
