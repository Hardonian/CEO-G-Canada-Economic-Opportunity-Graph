# CEGS Open Governance & RFC Process

CEGS is developed as a public good for the Canadian economic and data ecosystem. Its governance is transparent, meritocratic, and community-driven.

---

## 1. Decision Making & Roles

- **Maintainers**: Responsible for reviewing schema changes, certifying validator conformance, and stewarding the repository.
- **Contributors**: Anyone who submits an issue, proposes vocabulary terms, or authors a Request for Comments (RFC).
- **Advisory Implementers**: External data publishers (industry associations, government programs, research institutions) who test draft specifications against real data.

---

## 2. The Request for Comments (RFC) Process

Any substantial change to CEGS must proceed through the RFC process:

```text
Draft RFC ──► Community Review ──► Maintainer Consensus ──► Accepted ──► Released in Spec
```

### When an RFC is Required:
- Modifying the canonical envelope structure
- Adding, renaming, or removing core entity types
- Changing stable identifier syntax
- Altering claim status or confidence invariants
- Modifying monetary or temporal representation rules

### Proposing an RFC:
1. Copy `spec/cegs/rfcs/RFC-TEMPLATE.md` to `spec/cegs/rfcs/CEGS-RFC-XXXX-<title>.md`.
2. Detail motivation, proposal, compatibility impact, and alternate designs considered.
3. Open a Pull Request for public discussion.
4. After a minimum 14-day review period, maintainers record the decision.
