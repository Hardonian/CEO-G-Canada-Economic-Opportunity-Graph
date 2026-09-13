# CEGS-RFC-0003: Evidence and Cryptographic Provenance

- **RFC Number:** CEGS-RFC-0003
- **Title:** Evidence Objects and Cryptographic Content Hashing
- **Status:** Accepted
- **Created Date:** 2026-09-12
- **Specification Version:** CEGS 0.1

---

## 1. Summary
Establishes the `Evidence` object model requiring cryptographic SHA-256 snapshots, publisher attribution, source tiering, and retrieval timestamps for all factual assertions.

## 2. Motivation
Economic data feeds often assert project status or capital figures without providing audit trails. To prevent hallucination and detect source document updates, every fact must be grounded in an immutable evidence record.

## 3. Detailed Specification
An `Evidence` record contains:
- `source_url`: Verifiable HTTP(S) location.
- `publisher`: Named legal entity responsible for the content.
- `source_tier`: Tier 1 (statutory/primary) through Tier 4 (secondary).
- `content_hash`: SHA-256 hash of the exact retrieved document byte stream.
- `confidence`: Standard epistemic classification.

## 4. Copyright & Privacy Safe Harbor
Where source copyright prevents caching full document text, CEGS strictly retains the cryptographic hash, metadata, and locator without replicating restricted third-party prose.
