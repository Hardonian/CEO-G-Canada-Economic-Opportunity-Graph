# CEGS-RFC-0002: Stable URI-Based Identifiers

- **RFC Number:** CEGS-RFC-0002
- **Title:** Stable URI-Based Identifiers for Canadian Economic Entities
- **Status:** Accepted
- **Created Date:** 2026-09-12
- **Specification Version:** CEGS 0.1

---

## 1. Summary
Defines a predictable, human-readable, and database-independent URI syntax (`cegs:<type>:<jurisdiction>:<slug>`) for referencing projects, organizations, events, procurements, and evidence.

## 2. Motivation
Database-specific integer keys (`14829`) or opaque UUIDs (`a8b2...`) prevent data exchange between federation partners and obscure entity semantics in logs and diffs.

## 3. Detailed Specification
Format: `cegs:<resource_type>:<jurisdiction_code>:<slug>`
- `resource_type`: `project`, `org`, `event`, `rel`, `evidence`, `proc`, `opp`, `score`, `signal`.
- `jurisdiction_code`: `ca` (federal) or `ca:<province_code>` (e.g. `ca:on`, `ca:bc`, `ca:qc`).
- `slug`: kebab-case alphanumeric string derived deterministically from canonical legal name or statutory project code.

## 4. Rationale & Alternatives Considered
URIs provide natural namespace collision avoidance, support decentralized publishing, and allow instant offline debugging.
