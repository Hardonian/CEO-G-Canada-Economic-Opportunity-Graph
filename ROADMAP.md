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

* [ ] **Automated Gazette Polling**: Recurring scheduled workers scraping provincial gazettes (Ontario Gazette, Gazette officielle du Québec, BC Gazette).
* [ ] **Cryptographic Merkle Proofs**: Publish daily Merkle tree roots of all ingested evidence hashes to an append-only transparency log.
* [ ] **Expanded Capital Stack Intelligence**: Integrate detailed provincial financing programs (Emissions Reduction Alberta, Investissement Québec, BC InBC).
* [ ] **Indigenous Business Directory Cross-Referencing**: Automated linkage to Indigenous Services Canada (ISC) Indigenous Business Directory for procurement set-asides.
* [ ] **Multi-Jurisdiction Reconciliation**: Automated resolution of overlapping federal-provincial regulatory filings.

---

## Phase 3: CEGS 1.0 Stability & Ecosystem Adoption (v1.0) — Q3/Q4

* [ ] **CEGS 1.0 Final Standardization**: Lock core vocabulary and release formal migration toolkits.
* [ ] **Python & Rust SDKs**: Standalone CEGS client libraries for data science, notebooks, and backend pipelines.
* [ ] **Enterprise GraphQL API**: High-throughput graph query API with fine-grained subscription webhooks for moving projects.
* [ ] **Community-Contributed Adapters**: Sandbox registry for community-maintained regional and municipal adapters.
* [ ] **Decentralized Verifier Nodes**: Multi-party notarization of major project milestone occurrences.
