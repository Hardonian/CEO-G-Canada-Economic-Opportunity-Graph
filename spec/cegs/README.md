# Canada Economic Graph Schema (CEGS)

**Specification Version: 0.1**  
**Target Stable Specification: 1.0**  
**Canonical Identifier:** `CEGS` (pronounced informally as *"segs"*)  
**Canonical Expansion:** *Canada Economic Graph Schema*  

---

## What is CEGS?

The **Canada Economic Graph Schema (CEGS)** is an open, implementation-neutral, machine-readable data standard for representing economic infrastructure, capital flows, public procurements, regulatory approvals, Indigenous partnerships, and industrial supply-chain opportunities across Canada.

CEGS defines **economic truth and provenance**. It is intentionally decoupled from specific database engines, cloud runtimes, programming languages, and commercial application platforms.

```text
                    CEGS
         Canada Economic Graph Schema
        open specification + validators
                       │
              ┌────────┼────────┐
              │        │        │
              ▼        ▼        ▼
           Dataset    APIs    Other
          Producers          Implementers
              │
              └────────┬────────┘
                       ▼
            CanadaOpportunityGraph
             Reference Implementation
                       │
          ┌────────────┼─────────────┐
          ▼            ▼             ▼
      Capital      Procurement    Investor
       Radar          Graph         Intel
```

---

## Core Specification Documents

* [SPECIFICATION.md](SPECIFICATION.md): Comprehensive normative specification for all CEGS object types, envelope structures, and semantic constraints.
* [DESIGN_PRINCIPLES.md](DESIGN_PRINCIPLES.md): Core axioms governing CEGS design (evidence before assertion, immutable history, fact vs inference separation).
* [VERSIONING.md](VERSIONING.md): Version lifecycle, schema compatibility guarantees, and deprecation policies.
* [GOVERNANCE.md](GOVERNANCE.md): Open-source governance model and Request for Comments (RFC) process.
* [CHANGELOG.md](CHANGELOG.md): History of revisions from CEGS 0.1 onward.

## Schemas & Vocabularies

* [`/schemas`](schemas/): Normative JSON Schemas (Draft 2020-12) for `entity`, `project`, `organization`, `location`, `event`, `relationship`, `evidence`, `capital-event`, `procurement`, `program`, `opportunity`, `score`, `signal`, and `dataset-manifest`.
* [`/vocab`](vocab/): Machine-readable authoritative vocabularies for sectors, project stages, relationship types, event types, evidence states, and source classifications.
* [`/examples`](examples/): Valid canonical JSON examples illustrating representative Canadian projects, organizations, events, and dataset manifests.
* [`/rfcs`](rfcs/): Architectural Decision Records and formal Requests for Comments governing the standard.

---

## Conformance Levels

1. **CEGS Core**: Schema-valid entities, locations, and relationships.
2. **CEGS Provenance**: Core conformance plus mandatory cryptographic content hashing (SHA-256) and source locators for every material assertion.
3. **CEGS Historical**: Provenance conformance plus an immutable, append-only event ledger capturing all lifecycle transitions.
4. **CEGS Intelligence**: Historical conformance plus derived opportunities, momentum signals, and versioned scoring models.

---

## Reference Implementation

[CanadaOpportunityGraph](../../README.md) is the official reference implementation of CEGS, demonstrating ingestion, deterministic scoring, geospatial exploration, and capital momentum tracking on top of CEGS-compliant data.
