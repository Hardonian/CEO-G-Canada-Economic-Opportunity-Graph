# CEGS Versioning Policy

CEGS adheres to a strict specification lifecycle designed to guarantee predictability for data publishers, researchers, and software implementers.

---

## 1. Specification Lifecycle

```text
Draft (0.x) ──► Candidate (1.0-RC) ──► Stable (1.x) ──► Major Evolution (2.0)
```

- **CEGS 0.1 – 0.9 (Pre-1.0 Development)**: The specification is functional and used in production by CanadaOpportunityGraph. Breaking structural enhancements are permitted across minor versions (e.g. 0.1 to 0.2), but all changes must be documented via formal RFCs and accompanied by data migration scripts.
- **CEGS 1.0 (Target Stable Release)**: Marks the first frozen standard with strict backward-compatibility guarantees.
- **CEGS 1.x (Post-1.0 Minor Releases)**: Additive fields, new vocabulary terms, and non-breaking schema expansions only.
- **CEGS 2.0 (Major Breaking Evolutions)**: Reserved for architectural paradigm shifts requiring breaking changes to the core envelope.

---

## 2. Backward Compatibility Rules (Post-1.0)

A revision is **backward-compatible** if and only if:
1. Every dataset conforming to version `1.x` remains valid under version `1.x+1`.
2. Existing required fields are never removed or renamed.
3. Existing vocabulary keys retain their exact semantic meaning.
4. Added fields are strictly optional (`required: false`).
5. New vocabulary items are added additively without altering previous interpretations.

---

## 3. Schema Hashes

Every release of the CEGS schema files generates a canonical SHA-256 manifest hash recorded in `spec/cegs/SCHEMA_HASHES.json`. CI will fail if schema files are altered without a corresponding version tag update.
