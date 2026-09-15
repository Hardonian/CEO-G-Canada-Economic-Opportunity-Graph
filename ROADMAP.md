# CanadaOpportunityGraph & CEGS Strategic Roadmap

This document outlines the phased milestone roadmap for CanadaOpportunityGraph and the Canada Economic Graph Schema (CEGS).

---

## Phase 1: Foundation & Baseline (v0.1) — Current Baseline (Completed)

* [x] **Core Ontology**: Project, Organization, Event, Relationship, Evidence, CapitalItem, Procurement, Opportunity, Signal.
* [x] **CEGS Specification 0.1**: JSON Schemas (Draft 2020-12), normative vocabularies, reference examples, RFC process.
* [x] **Reference Go Engine**: Ingestion pipeline, deterministic scoring engines (Buildability, Investability, Supplierability, AI Sovereignty).
* [x] **Authoritative Adapters**: IAAC, CanadaBuys, NRCan Major Projects, IDEaS Defence, CER Facilities.
* [x] **Deterministic Scoring & Propagation**: Downstream requirements graph generator and momentum signal tracking.
* [x] **Zero-Dependency CLI & REST API**: `cog` binary, OpenAPI 3.0 documentation, JSONL public datasets.
* [x] **Interactive Web Platform**: Next.js 15 institutional interface with Capital Radar, Projects Directory, Map, and Sovereignty Radar.

---

## Phase 2: Live Expansion & Provable Provenance (v0.2) — Next 90 Days

* [x] **Automated Gazette Polling**: Recurring scheduled workers scraping provincial gazettes (Ontario Gazette, Gazette officielle du Québec, BC Gazette). Implemented in `internal/gazettepoll/` with change detection, wired into `cmd/worker/main.go`.
* [x] **Cryptographic Merkle Proofs**: Publish daily Merkle tree roots of all ingested evidence hashes to an append-only transparency log. Implemented in `internal/merkle/` with deterministic empty-tree root, wired into `cmd/worker/main.go` and `cmd/api/main.go`.
* [x] **Expanded Capital Stack Intelligence**: Detailed provincial financing programs integrated (Emissions Reduction Alberta, Investissement Québec, BC InBC, and 30+ more). Implemented in `internal/capitalstack/engine.go` with stacking evaluation, mutual exclusivity checks, and CAPEX conflict detection.
* [x] **Indigenous Business Directory Cross-Referencing**: Automated linkage to Indigenous Services Canada (ISC) Indigenous Business Directory for procurement set-asides. Implemented in `adapters/indigenous/` and `internal/indigenouslinker/`, wired into `cmd/worker/main.go`.
* [x] **Multi-Jurisdiction Reconciliation**: Automated resolution of overlapping federal-provincial regulatory filings. Implemented in `internal/reconciliation/` with merge/link/conflict actions, wired into `cmd/worker/main.go` and `cmd/api/main.go`.

---

## Phase 3: CEGS 1.0 Stability & Ecosystem Adoption (v1.0) — Q3/Q4

* [x] **CEGS 1.0 Migration Toolkit**: Migration toolkit for converting legacy JSON Schemas to canonical CEGS 1.0 vocabulary. Implemented in `internal/cegs/migration.go`.
* [ ] **CEGS 1.0 Final Standardization**: Lock core vocabulary and release formal migration toolkits. Add conformance test suite for all CEGS resource types.
* [ ] **Python & Rust SDKs**: Standalone CEGS client libraries for data science, notebooks, and backend pipelines.
* [ ] **Enterprise GraphQL API**: High-throughput graph query API with fine-grained subscription webhooks for moving projects. Foundation in `internal/api/server.go`; needs `internal/graphql/` package.
* [x] **Community-Contributed Adapters**: Sandbox registry for community-maintained regional and municipal adapters. Implemented in `internal/adaptersandbox/registry.go` with validation.
* [x] **Decentralized Verifier Nodes**: Multi-party notarization of major project milestone occurrences. Implemented in `internal/verifier/verifier.go` with quorum-based attestation.
