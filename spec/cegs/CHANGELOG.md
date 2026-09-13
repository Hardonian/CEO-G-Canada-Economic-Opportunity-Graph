# CEGS Changelog

All notable changes to the Canada Economic Graph Schema specification will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [0.1.1] - 2026-09-13

### Added
- Additive sector terms for `Mining & Metals`, `Energy & Fuels`, and `Forestry & Bioeconomy` so nationwide source inventories can be represented without falsely classifying all mining as critical minerals or all energy as clean power.
- Nationwide NRCan Major Projects Inventory snapshot support with record-level hashes and explicit `REPORTED` epistemic status.

## [0.1.0] - 2026-09-12

### Added
- Initial formal release of the **Canada Economic Graph Schema (CEGS 0.1)**.
- Normative JSON Schemas for `entity`, `project`, `organization`, `location`, `event`, `relationship`, `evidence`, `capital-event`, `procurement`, `program`, `opportunity`, `score`, `signal`, and `dataset-manifest`.
- Standard vocabularies for sectors, project stages, event types, relationship types, evidence states, and source classifications.
- Formal CEGS envelope (`cegs`, `id`, `type`, `canonical_name`, `jurisdiction`, `created_at`, `updated_at`, `provenance`, `extensions`).
- Stable URI-based identifier syntax (`cegs:<type>:<jurisdiction>:<slug>`).
- Epistemic invariants (`VERIFIED`, `SUPPORTED`, `REPORTED`, `INFERRED`, `CONFLICTED`, `UNKNOWN`, `STALE`, `RETRACTED`).
- Four conformance levels: Core, Provenance, Historical, Intelligence.
- Initial foundational RFCs: CEGS-RFC-0001 through CEGS-RFC-0006.
- CanadaOpportunityGraph established as the primary reference implementation.
