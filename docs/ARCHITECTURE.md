# CanadaOpportunityGraph Architecture

## 1. System Overview

CanadaOpportunityGraph is architected around three discrete layers:

1. **The Open Standard (CEGS)**: `spec/cegs/` defines the formal schema, vocabularies, and normative constraints for Canadian economic data.
2. **The Reference Implementation (COG Core)**: `internal/` and `adapters/` implement ingestion, cryptographic hashing, entity resolution, append-only event sourcing, deterministic scoring, and opportunity propagation.
3. **The Presentation & Intelligence Surfaces**: `cmd/api` (REST API with OpenAPI spec), `cmd/cog` (CLI tool), and `apps/web` (Next.js institutional application).

```text
                                CEGS Specification (0.1)
                                      │
                 ┌────────────────────┼────────────────────┐
                 ▼                    ▼                    ▼
             JSON Schemas        Vocabularies         Validators
                 │                    │                    │
                 └────────────────────┼────────────────────┘
                                      │
                                      ▼
                      CanadaOpportunityGraph Reference Engine
               ┌──────────────────────────────────────────────┐
               │ Adapters: IAAC, CanadaBuys, NRCan, IDEaS...  │
               │ Fetcher -> SHA-256 Hashing -> Change Detector│
               │ Entity Resolution & Canonical Deduplication  │
               │ Append-Only Event Ledger & Temporal Graph    │
               │ Deterministic Scoring Engine (Buildability)  │
               │ Opportunity Propagation Engine (Downstream)  │
               │ Immutable snapshots -> MemoryStore projection│
               └──────────────────────┬───────────────────────┘
                                      │
         ┌────────────────────────────┼────────────────────────────┐
         ▼                            ▼                            ▼
  REST API (:8080)             CLI Binary (cog)           Next.js Web (:3000)
  /api/v1/radar                cog search ...             Capital Radar
  /api/v1/projects             cog project show ...       Projects Directory
  /api/v1/cegs/export          cog cegs validate ...      Geospatial Map
  OpenAPI 3.0                  cog demo                   Capital Stack
```

## 2. Ingestion Pipeline & Anti-Hallucination Guardrails

- **Deterministic Adapters**: Ingest data from Tier 1 authoritative sources (IAAC, CER, CanadaBuys, NRCan).
- **Cryptographic Moat**: Every raw payload is hashed via SHA-256 (`adapters.HashDocument`). Identical documents are bypassed.
- **Append-Only History**: Entity lifecycle transitions create discrete `domain.Event` records. Historical records are never mutated or deleted.
- **Epistemic States**: Factual confidence is explicitly marked as `VERIFIED`, `SUPPORTED`, `INFERRED`, `CONFLICTED`, `UNKNOWN`, or `STALE`. Missing data remains `UNKNOWN`.

## 3. Runtime Data Layer

- **Shipped runtime**: The API deterministically loads pinned, evidence-linked source snapshots into a thread-safe `MemoryStore` at startup. It is a stateless, read-only projection that can be rebuilt and independently verified on every instance.
- **Future durable store**: `migrations/001_initial_schema.sql` is a design draft, not a wired production backend. `DATABASE_URL` is rejected at startup so operators cannot mistake discarded writes for persistence.
- **Release boundary**: `cmd/releasecheck` validates hashes, counts, ordering, duplicate IDs, graph references, CEGS conformance, web snapshot parity, and current/latest immutable release equality.
