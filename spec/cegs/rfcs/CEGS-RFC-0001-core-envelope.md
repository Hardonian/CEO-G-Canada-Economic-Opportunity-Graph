# CEGS-RFC-0001: Core Resource Envelope

- **RFC Number:** CEGS-RFC-0001
- **Title:** Core Canonical Resource Envelope
- **Status:** Accepted
- **Created Date:** 2026-09-12
- **Specification Version:** CEGS 0.1

---

## 1. Summary
Establishes the normative top-level JSON envelope required for every primary CEGS document, standardizing metadata (`cegs`, `id`, `type`, `canonical_name`, `jurisdiction`, `created_at`, `updated_at`, `provenance`, and `extensions`).

## 2. Motivation
Independent data producers (federal registries, provincial databases, corporate disclosures) represent economic records using vastly disparate schemas. A shared envelope ensures any JSON consumer can inspect identity, version, jurisdiction, and provenance without bespoke ingestion logic.

## 3. Detailed Specification
Every root CEGS object must include:
- `cegs`: String version marker (`"0.1"`).
- `id`: Canonical URI identifier.
- `type`: Lowercase entity classifier.
- `canonical_name`: Primary human-readable identifier.
- `jurisdiction`: ISO 3166 territory code.
- `created_at` & `updated_at`: RFC 3339 UTC timestamps.
- `provenance`: Array of evidence reference IDs.
- `extensions`: Object reserved for non-standard vendor fields.

## 4. Rationale & Alternatives Considered
We evaluated JSON-LD `@context` envelopes. While semantically expressive, strict JSON-LD creates excessive implementation complexity for standard data science and Go/TypeScript runtimes. The CEGS envelope provides immediate clarity in standard JSON while allowing optional JSON-LD context mapping.
