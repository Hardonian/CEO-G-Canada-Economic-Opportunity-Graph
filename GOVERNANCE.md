# CanadaOpportunityGraph & CEGS Governance Model

## 1. Dual Governance Architecture

CanadaOpportunityGraph operates a **dual-layer governance model** that strictly decouples the data standard from the reference platform implementation:

```text
┌─────────────────────────────────────────────────────────────┐
│                 CEGS Standards Committee                   │
│  - Neutral data ontology for Canadian economic entities    │
│  - Governed by formal RFC process (spec/cegs/rfcs/)         │
│  - Consensus-driven, backwards-compatible versioning        │
└──────────────────────────────┬──────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│             CanadaOpportunityGraph Core Team                │
│  - Reference implementation engine (Go, PostgreSQL, Next.js)│
│  - Real-world adapters, ingestors, scoring models, and UI   │
│  - Open-source stewardship under Apache 2.0                 │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Roles & Responsibilities

### Maintainers
Maintainers have write and release access across the repository. They are responsible for:
- Reviewing and approving RFCs for CEGS.
- Reviewing pull requests and ensuring strict test coverage and anti-hallucination standards.
- Managing security advisories and dependency updates.
- Tagging semantic releases (`v0.1.0`, `v0.2.0`, `v1.0.0`).

### Domain Reviewers
Domain specialists in Canadian infrastructure finance, regulatory law (IAAC, CER), Indigenous economic development, and defence procurement provide domain validation on schema fields and scoring methodologies.

### Contributors
Anyone who submits issues, adapter code, documentation improvements, or RFC proposals.

---

## 3. The CEGS RFC Process

Any breaking change, new entity type, vocabulary addition, or scoring modification must follow the RFC lifecycle:

1. **Draft**: Author copies `spec/cegs/rfcs/RFC-TEMPLATE.md` and opens a PR titled `RFC: <Feature Name>`.
2. **Review & Discussion**: Community and maintainers debate the proposal for a minimum 14-day consultation window.
3. **Implementation & Conformance**: A working prototype or validator update must be demonstrated.
4. **Final Call for Comments (FCP)**: 7-day period for final objections.
5. **Accepted / Rejected**: Merged into `spec/cegs/rfcs/` with an assigned number.

---

## 4. Decision Making & Escalation

Decisions are made by consensus whenever possible. When consensus cannot be reached, the Lead Maintainer has ultimate architectural responsibility for project integrity and security.
