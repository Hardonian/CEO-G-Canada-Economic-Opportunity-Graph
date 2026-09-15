//! # cog-sdk
//!
//! Rust client SDK for the **CanadaOpportunityGraph (COG)** API — CEGS 1.0.
//!
//! ## Quick Start
//!
//! ```no_run
//! use cog_sdk::{CogClient, ProjectFilter};
//!
//! let client = CogClient::new("https://api.canadaopportunitygraph.ca").unwrap();
//!
//! // List projects in the LNG sector in BC
//! let projects = client.list_projects(ProjectFilter {
//!     sector: Some("Energy & Fuels".into()),
//!     province: Some("BC".into()),
//!     limit: Some(10),
//!     ..Default::default()
//! }).unwrap();
//!
//! for p in &projects.projects {
//!     println!("{}: {} CAD", p.name, p.capex_cad);
//! }
//!
//! // Retrieve a CEGS-formatted project record
//! let cegs = client.get_cegs_project(&projects.projects[0].id).unwrap();
//! println!("CEGS ID: {}", cegs.id);
//!
//! // Run a GraphQL query
//! let data = client.graphql(
//!     "{ projects(limit: 5) { id name stage capexCAD } }",
//!     None,
//! ).unwrap();
//! println!("{}", serde_json::to_string_pretty(&data).unwrap());
//! ```
//!
//! ## Authentication
//!
//! Set the `COG_API_KEY` environment variable, or chain [`CogClient::with_api_key`]:
//!
//! ```no_run
//! use cog_sdk::CogClient;
//!
//! let client = CogClient::new("https://api.canadaopportunitygraph.ca")
//!     .unwrap()
//!     .with_api_key("my-api-key")
//!     .unwrap();
//! ```

pub mod client;
pub mod error;
pub mod models;

pub use client::{CogClient, ProjectFilter};
pub use error::{CogError, CogResult};
pub use models::{
    CapitalProgram, CapitalRange, CapitalStack, CegsEnvelope, Location, Monetary, Organization,
    Project, ProjectList, ScoreBundle, ScoreDetail, Signal, SignalList,
};
