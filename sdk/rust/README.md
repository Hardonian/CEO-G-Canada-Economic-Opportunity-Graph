# cog-sdk — Rust Client for CanadaOpportunityGraph

[![Crates.io](https://img.shields.io/crates/v/cog-sdk.svg)](https://crates.io/crates/cog-sdk)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache--2.0-blue.svg)](../../LICENSE)

A zero-async, blocking Rust client library for the [CanadaOpportunityGraph](https://github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph) API, implementing the **CEGS 1.0** open standard.

---

## Installation

Add to your `Cargo.toml`:

```toml
[dependencies]
cog-sdk = "0.2"
```

Or with a path reference during development:

```toml
[dependencies]
cog-sdk = { path = "../../sdk/rust" }
```

---

## Quick Start

```rust
use cog_sdk::{CogClient, ProjectFilter};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = CogClient::new("https://api.canadaopportunitygraph.ca")?;

    // List major clean energy projects in British Columbia
    let result = client.list_projects(ProjectFilter {
        sector: Some("Clean Energy & Grid".into()),
        province: Some("BC".into()),
        limit: Some(20),
        ..Default::default()
    })?;

    for project in &result.projects {
        println!(
            "[{}] {} — {} CAD — Stage: {}",
            project.province, project.name, project.capex_cad, project.current_stage
        );
    }

    Ok(())
}
```

---

## Features

| Method | Description |
|--------|-------------|
| `list_projects(filter)` | List projects with optional filters (sector, province, stage, capex, search) |
| `get_project(id)` | Retrieve a project by ID or slug |
| `get_project_scores(id)` | Get buildability / investability / supplierability scores |
| `get_capital_stack(id)` | Resolve provincial and federal financing programs |
| `list_signals(since, limit)` | Economic momentum signals (CONSTRUCTION_SIGNAL, FINANCING_ACCELERATION, …) |
| `list_organizations(limit)` | List organizations / entities |
| `get_cegs_project(id)` | Retrieve a project as a canonical CEGS 1.0 envelope |
| `graphql(query, vars)` | Execute any GraphQL query against the enterprise API |

---

## Authentication

Set `COG_API_KEY` in your environment, or chain `with_api_key`:

```rust
let client = CogClient::new("https://api.canadaopportunitygraph.ca")?
    .with_api_key("my-secret-key")?;
```

---

## GraphQL Example

```rust
let data = client.graphql(
    r#"{ projects(sector: "Nuclear & Clean Power", limit: 5) {
        id name stage capexCAD
    }}"#,
    None,
)?;
println!("{}", serde_json::to_string_pretty(&data)?);
```

---

## Error Handling

All methods return `CogResult<T>` — a type alias for `Result<T, CogError>`.

```rust
use cog_sdk::CogError;

match client.get_project("unknown-id") {
    Ok(proj) => println!("Found: {}", proj.name),
    Err(CogError::Api { status, code, message }) => {
        eprintln!("API error {status} [{code}]: {message}");
    }
    Err(e) => eprintln!("Other error: {e}"),
}
```

---

## CEGS 1.0 Compliance

The `get_cegs_project` method returns a [`CegsEnvelope`] that conforms to the
[Canada Economic Graph Schema (CEGS) 1.0](../../spec/cegs/SPECIFICATION.md).
The `cegs` field will always be `"0.1"` until the canonical 1.0 version tag is
published.

---

## License

Apache-2.0 — see [LICENSE](../../LICENSE).
