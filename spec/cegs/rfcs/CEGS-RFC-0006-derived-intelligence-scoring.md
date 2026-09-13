# CEGS-RFC-0006: Derived Intelligence, Scores, and Signals

- **RFC Number:** CEGS-RFC-0006
- **Title:** Representation of Algorithmic Scores and Inflection Signals
- **Status:** Accepted
- **Created Date:** 2026-09-12
- **Specification Version:** CEGS 0.1

---

## 1. Summary
Defines normative structures for analytical scores and temporal inflection signals while explicitly decoupling CEGS from any proprietary or platform-specific scoring algorithm.

## 2. Motivation
Implementations such as CanadaOpportunityGraph compute proprietary or open-source scores (Buildability, Investability, Supplierability, Strategicity). CEGS must support exchanging these scores without dictating that CanadaOpportunityGraph's algorithm is the only valid scoring methodology.

## 3. Detailed Specification
- `score`: Contains numerical value, normalized scale bounds (`min`, `max`), calculation timestamp, and a required `methodology` URI (e.g. `"cog:buildability:v1.0"`).
- `signal`: Contains signal type, magnitude, directionality, and previous/new state diffs.
