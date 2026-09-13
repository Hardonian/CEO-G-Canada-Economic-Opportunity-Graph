# CanadaOpportunityGraph: Official CEGS Reference Implementation

This document details the exact mapping between the internal storage and domain models of **CanadaOpportunityGraph (COG)** and the **Canada Economic Graph Schema (CEGS 0.1)**.

---

## 1. Architectural Relationship

```text
       ┌────────────────────────────────────────────────────────┐
       │             Canada Economic Graph Schema               │
       │                       (CEGS)                           │
       │    Open specification, JSON Schemas, Vocabularies      │
       └───────────────────────────┬────────────────────────────┘
                                   │
                                   ▼
       ┌────────────────────────────────────────────────────────┐
       │             CanadaOpportunityGraph (COG)               │
       │                Reference Implementation                │
       │                                                        │
       │  ┌───────────────────────┐  ┌───────────────────────┐  │
       │  │ Ingestion & Adapters  │  │ Deterministic Scoring │  │
       │  │ (IAAC, CanadaBuys...) │  │ (Buildability v1.0)   │  │
       │  └───────────┬───────────┘  └───────────┬───────────┘  │
       │              ▼                          ▼              │
       │  ┌──────────────────────────────────────────────────┐  │
       │  │        Internal Storage & Event Ledger           │  │
       │  │          (PostgreSQL / MemoryStore)              │  │
       │  └───────────────────┬──────────────────────────────┘  │
       │                      │                                 │
       │                      ▼                                 │
       │  ┌──────────────────────────────────────────────────┐  │
       │  │            CEGS Transformer & Serializer         │  │
       │  │    (internal/cegs: ToCEGS / FromCEGS)            │  │
       │  └───────────────────┬──────────────────────────────┘  │
       └──────────────────────┼─────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
    CEGS REST API         CLI Engine        Public Snapshot
  /api/v1/cegs/...        cog cegs ...       /data/cegs/...
```

---

## 2. Resource Mapping Table

| Internal Domain Entity (`internal/domain`) | CEGS Canonical Resource (`spec/cegs/schemas`) | Canonical ID Prefix |
| :--- | :--- | :--- |
| `domain.Project` | `project.schema.json` | `cegs:project:ca:<province>:<slug>` |
| `domain.Entity` | `organization.schema.json` | `cegs:org:ca:<slug>` |
| `domain.Event` | `event.schema.json` | `cegs:event:ca:<slug>` |
| `domain.Relationship` | `relationship.schema.json` | `cegs:rel:ca:<slug>` |
| `domain.Evidence` | `evidence.schema.json` | `cegs:evidence:ca:<slug>` |
| `domain.CapitalItem` | `capital-event.schema.json` | `cegs:capital:ca:<slug>` |
| `domain.Procurement` | `procurement.schema.json` | `cegs:proc:ca:<slug>` |
| `domain.Opportunity` | `opportunity.schema.json` | `cegs:opp:ca:<slug>` |
| `domain.ProjectScore` | `score.schema.json` | `cegs:score:ca:<slug>` |
| `domain.Signal` | `signal.schema.json` | `cegs:signal:ca:<slug>` |

---

## 3. Transformation Guarantees

1. **Deterministic ID Generation**:
   The Go translation package (`internal/cegs`) maps UUIDs and slugs to canonical URI schemes deterministically:
   - If a project slug is `darlington-smr` in Ontario, the canonical CEGS ID is `cegs:project:ca:on:darlington-smr`.
2. **Monetary Standardization**:
   CAD integers in cents or whole dollars are mapped strictly to the CEGS currency envelope:
   ```json
   {
     "amount": 3400000000,
     "currency": "CAD",
     "amount_type": "reported"
   }
   ```
3. **Epistemic Invariants**:
   Internal confidence enums (`VERIFIED`, `SUPPORTED`, `INFERRED`, `CONFLICTED`, `UNKNOWN`, `STALE`) map 1:1 without semantic degradation.
4. **Extension Isolation**:
   Platform-specific calculation details (e.g. Buildability factor sub-weights) are isolated into `extensions["ca.opengraph.buildability"]`.
