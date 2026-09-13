# CEGS-RFC-0005: Typed Relationship Semantics

- **RFC Number:** CEGS-RFC-0005
- **Title:** Explicit First-Class Relationship Entities
- **Status:** Accepted
- **Created Date:** 2026-09-12
- **Specification Version:** CEGS 0.1

---

## 1. Summary
Establishes typed, directional relationship resources (`owns`, `finances`, `regulates`, `supplies`, `indigenous_partner`) as discrete first-class objects rather than embedded foreign key arrays.

## 2. Motivation
Relationships in infrastructure projects are rarely simple binary links. They possess temporal validity windows (`valid_from` / `valid_to`), independent evidentiary backing, and verification status that cannot be captured in nested ID arrays.

## 3. Detailed Specification
Every edge in the graph is represented by a `Relationship` document specifying `from`, `to`, `relationship_type`, `status`, temporal bounds, and provenance.
