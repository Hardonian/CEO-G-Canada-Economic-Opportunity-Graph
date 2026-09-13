# CEGS-RFC-0004: Event Lifecycle & Append-Only History

- **RFC Number:** CEGS-RFC-0004
- **Title:** Event-Sourced Lifecycle Transitions and Append-Only History
- **Status:** Accepted
- **Created Date:** 2026-09-12
- **Specification Version:** CEGS 0.1

---

## 1. Summary
Mandates an append-only event ledger for all economic entity mutations (stage transitions, financing commitments, permit submissions, Indigenous agreements).

## 2. Motivation
Traditional relational database patterns overwrite prior rows (e.g. updating `stage = 'CONSTRUCTION'`), destroying the velocity, delay patterns, and historical trajectory of the project. CEGS treats history as immutable.

## 3. Detailed Specification
All state transitions are realized as discrete `Event` objects referencing their parent entity/project and corroborated by an `Evidence` ID. Entity views are projections over the cumulative event stream.
