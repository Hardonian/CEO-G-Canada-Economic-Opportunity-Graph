# CEGS Design Principles

These eight normative design principles govern the development, evolution, and validation of the Canada Economic Graph Schema (CEGS).

---

## Principle 1 — Evidence Before Assertion
Every material factual assertion (e.g., project stage, capital expenditure amount, regulatory filing, Indigenous agreement) must be capable of referencing an explicit `Evidence` record containing:
- Source publisher and canonical URL
- Cryptographic raw snapshot hash (SHA-256)
- Retrieval and publication timestamps
- Source reliability classification (Tier 1 through 4)

Unsubstantiated claims are rejected by CEGS Provenance conformance profiles.

---

## Principle 2 — Unknown is Legitimate Data
Incomplete information is a natural condition of economic intelligence. Missing information must **never** be silently converted to `false`, `0`, `HEALTHY`, `APPROVED`, or `COMPLETE`.

CEGS explicitly reserves and mandates the use of `UNKNOWN`, `CONFLICTED`, and `STALE` to preserve epistemic honesty.

---

## Principle 3 — History is Immutable
Economic reality progresses through append-only historical occurrences. CEGS entities do not overwrite their past.

Lifecycle transitions, capital revisions, and regulatory milestones are recorded as discrete, immutable `Event` objects. Corrections create subsequent reconciling events rather than rewriting or erasing prior recorded states.

---

## Principle 4 — Fact and Inference Are Different Objects
CEGS maintains a rigorous ontological wall between:
1. **Reported Facts**: Sourced directly from primary documents (e.g., a formal Impact Assessment Agency filing).
2. **Derived Relationships**: Inferred through explicit semantic rules (e.g., parent company ownership).
3. **Speculative Dependencies**: Downstream market opportunities inferred through sector dependency ontologies.

Inferences must never be masqueraded as verified contractual facts.

---

## Principle 5 — Sources Remain Visible
Consumers of CEGS data must always be able to inspect where information originated. Field-level provenance attributes ensure that aggregators and analytical systems do not strip source visibility when compiling macro graphs.

---

## Principle 6 — Determinism
Canonicalization, identifier generation, validation, and semantic diffing must produce identical results across independent runtimes. No probabilistic or heuristic ambiguity is permitted in schema validation or core graph traversal.

---

## Principle 7 — Implementation Neutrality
CEGS does not require or assume:
- PostgreSQL, Neo4j, SQLite, or any specific database engine
- Go, Python, TypeScript, or any specific programming runtime
- Next.js, React, or any specific frontend framework
- CanadaOpportunityGraph or any commercial application

A researcher with a static JSON file or a government agency with an internal relational warehouse can produce and consume valid CEGS data without running CanadaOpportunityGraph.

---

## Principle 8 — Extension Without Fragmentation
CEGS permits vendor and domain extensions via the namespaced `extensions` envelope property (e.g. `extensions["ca.opengraph.buildability"]`).

Extensions must adhere to strict rules:
- Unknown extension namespaces must not invalidate base CEGS documents.
- Extensions cannot redefine or contradict required core semantics.
- Extensions must not override or bypass provenance constraints.
