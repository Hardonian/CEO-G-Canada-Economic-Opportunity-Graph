# CanadaOpportunityGraph (COG)

> **Know what Canada is building before everyone else does.**  
> **Follow Canadian capital before it becomes Canadian construction.**

[![CEGS Standard](https://img.shields.io/badge/CEGS-0.1%20Compliant-00F2FE?style=flat-square&logo=json)](spec/cegs/README.md)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg?style=flat-square)](LICENSE)
[![Data Terms](https://img.shields.io/badge/data-source--specific-4ade80?style=flat-square)](DATA_SOURCES.md)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat-square&logo=go)](cmd/)
[![Next.js](https://img.shields.io/badge/Frontend-Next.js%2015-black?style=flat-square&logo=next.js)](apps/web/)

Canada is entering a multi-trillion dollar nation-building and infrastructure cycle spanning critical minerals, nuclear Small Modular Reactors (SMRs), clean electricity grids, sovereign AI compute, defence and Arctic modernization, port expansions, and industrial gigafactories.

The essential data describing this buildout is fragmented across hundreds of federal registries, provincial gazettes, utility regulators, environmental assessment portals, corporate disclosures, and procurement tender boards.

**CanadaOpportunityGraph** turns that fragmented public evidence into an open, continuously evolving, mathematically rigorous economic graph.

---

### Built on CEGS

CanadaOpportunityGraph is the official **reference implementation and flagship application** of the **[Canada Economic Graph Schema (CEGS)](spec/cegs/README.md)**.

CEGS is an open, implementation-neutral data specification for representing Canadian capital projects, corporate entities, procurement tenders, regulatory milestones, Indigenous partnerships, and cryptographic evidence provenance.

**CEGS is strictly decoupled from CanadaOpportunityGraph:** Other publishers, government programs, consultancies, investigative journalists, and software developers can produce and consume valid CEGS datasets without operating this platform.

---

## Three-Tier Architecture

```text
       ┌────────────────────────────────────────────────────────┐
       │             Canada Economic Graph Schema               │
       │                       (CEGS)                           │
       │     Open Data Standard, JSON Schemas, Vocabularies     │
       └───────────────────────────┬────────────────────────────┘
                                   │
                                   ▼
       ┌────────────────────────────────────────────────────────┐
       │             CanadaOpportunityGraph (COG)               │
       │                Reference Implementation                │
       │  ┌───────────────────────┐  ┌───────────────────────┐  │
       │  │ Ingestion & Adapters  │  │ Deterministic Scoring │  │
       │  │ (IAAC, CanadaBuys...) │  │ (Buildability v1.0)   │  │
       │  └───────────┬───────────┘  └───────────┬───────────┘  │
       │              ▼                          ▼              │
       │  ┌──────────────────────────────────────────────────┐  │
       │  │       Evidence-linked Snapshot & Event Graph     │  │
       │  │         (deterministic in-memory projection)     │  │
       │  └──────────────────────────────────────────────────┘  │
       └───────────────────────────┬────────────────────────────┘
                                   │
                                   ▼
       ┌────────────────────────────────────────────────────────┐
       │                 Canada Capital Radar                   │
       │                 Decision Intelligence                  │
       │     Capital Radar • Procurement • AI Sovereignty       │
       └────────────────────────────────────────────────────────┘
```

---

## Key Differentiating Capabilities

### 1. Deterministic Scoring Engines (No Hallucinated Scores)
Authoritative scores are computed using versioned, deterministic algorithms:
- **Buildability Score (0–100)**: Execution likelihood based on regulatory progress, Indigenous agreements, financing readiness, and site control.
- **Investability Score (0–100)**: Capital opportunity scale, public co-investment de-risking, and offtake strength.
- **Supplierability Score (0–100)**: Procurement proximity, engineering complexity, and domestic supply-chain intensity.
- **Strategicity Score (0–100)**: Alignment with Canada's 34 critical minerals list, energy security, and Arctic sovereignty.
- **Canadian AI Sovereignty Index (0–100)**: Transparent scorecard on Canadian data/compute residency, US CLOUD Act immunity, bilingual capability, and Quebec Law 25 compliance.

### 2. Opportunity Propagation Engine
When an upstream project is announced (e.g. an open-pit nickel mine or SMR nuclear facility), the engine deterministically derives downstream industrial demand (e.g. 230kV high-voltage substation interconnect, EPCM engineering packages, modular remote workforce camps, continuous aquatic monitoring).
- Strictly distinguishes `CONFIRMED` procurement from `DERIVED` requirements and `SPECULATIVE` dependencies.

### 3. Sourced Provenance & Cryptographic Data Moat
- Every material factual assertion points to an `Evidence` record with source URL, publisher, retrieval timestamp, and raw document SHA-256 content hash.
- History is immutable: state transitions are append-only.
- Missing data remains explicitly `UNKNOWN`. Conflicting evidence remains `CONFLICTED`.

---

## Developer Quickstart

The repository provides identical targets for Unix (`make`) and native Windows PowerShell (`scripts/task.ps1`).

### Option A: Windows PowerShell
```powershell
# Run deterministic offline demo
.\scripts\task.ps1 demo

# Run the full test suite with race detector
.\scripts\task.ps1 test

# Validate CEGS specification schemas & public manifests
.\scripts\task.ps1 cegs-validate

# Launch the API server (:8080)
.\scripts\task.ps1 api
```

### Option B: Unix / macOS / CI
```bash
# Build CLI and server binaries
make build

# Run deterministic demo
make demo

# Run all tests
make test

# Validate CEGS schemas and public dataset
make cegs-validate

# Run Next.js frontend (:3000)
make web-dev
```

---

## The Official CLI: `cog`

```bash
# Search major projects across sectors
$ cog search "nuclear ontario"

# Display investor-grade project dossier
$ cog project show darlington-small-modular-reactor-deployment-project

# Inspect recent economic momentum signals (last 30 days)
$ cog changes --since 30d

# View rankings by Buildability or Investability
$ cog rankings buildability

# Export dossier in CEGS 0.1 canonical format
$ cog export project darlington-smr --format cegs

# Validate any external document against the CEGS 0.1 standard
$ cog cegs validate dataset.json

# Inspect CEGS trust profile and evidence count
$ cog cegs inspect project.json

# Semantic diff between two project versions
$ cog cegs diff v1.json v2.json
```

---

## Public CEGS Datasets

Public release snapshots are maintained under [`data/cegs/`](data/cegs/) and [`data/public/`](data/public/):
- `data/cegs/manifest.json`: Cryptographic dataset manifest.
- `data/cegs/projects.jsonl`: Verified major infrastructure projects.
- `data/cegs/organizations.jsonl`: Proponents, Crown corporations, and regulators.
- `data/cegs/events.jsonl`: Append-only transition events.
- `data/public/trade_metrics.jsonl`: Official normalized trade and logistics score inputs with evidence IDs.
- [`docs/DATA_CREDENTIALS.md`](docs/DATA_CREDENTIALS.md): Credential-free sources, required account keys, and secret-handling rules.

---

## REST API Overview

Base URL: `http://localhost:8080/api/v1`

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/radar` | Flagship macro stats, moving capital, accelerating assets |
| `GET` | `/projects` | Query major projects with sector/province filters |
| `GET` | `/projects/{id}` | Complete project dossier with scores and events |
| `GET` | `/projects/{id}/provenance` | Cryptographic evidence graph |
| `GET` | `/projects/{id}/trust` | Evidence Quality and primary source coverage |
| `GET` | `/capital/stack` | Canadian Capital Stack programs & stacking rules |
| `GET` | `/ai-sovereignty` | Canadian AI Sovereignty Index benchmarks |
| `GET` | `/rankings/{dim}` | Ranked projects by Buildability, Investability, etc. |
| `GET` | `/cegs/projects/{id}` | Export project conforming to CEGS 0.1 |
| `GET` | `/cegs/export` | Full CEGS canonical JSON dataset export |

---

## Documentation Index

- [CEGS Specification (v0.1)](spec/cegs/README.md)
- [CEGS Reference Implementation Mapping](docs/cegs/reference-implementation.md)
- [CEGS Adoption Guide](apps/web/app/cegs/adopt/page.tsx)
- [Scoring Methodology & Mathematical Formulas](docs/SCORING.md)
- [Security Threat Model & SSRF Defenses](docs/SECURITY.md)
- [Commercialization & Entitlement Architecture](docs/COMMERCIAL.md)

---

## License

- **Software**: [Apache License 2.0](LICENSE)
- **Specification (CEGS)**: [Creative Commons Attribution 4.0 International (CC BY 4.0)](spec/cegs/LICENSE)
- **Public Datasets**: source-specific terms are retained. Government of Canada records use the Open Government Licence – Canada; World Bank observations generally use CC BY 4.0 plus the publisher's dataset terms. See [DATA_SOURCES.md](DATA_SOURCES.md) and each evidence record before redistribution.
