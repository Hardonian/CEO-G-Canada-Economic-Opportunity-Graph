# Source Visibility & Publication Policy

## Overview

CanadaOpportunityGraph enforces a strict, one-way boundary between private
workspace intelligence and public products. This document defines the
visibility classes, publication rules, and the private→public corroboration
workflow.

## Visibility Classes

Every source, evidence record, claim, opportunity, capital need, and milestone
carries a `VisibilityClass`:

| Class | Can Publish? | Description |
|-------|-------------|-------------|
| `PUBLIC` | ✅ Yes | Freely available government/public data |
| `PUBLIC_WITH_ATTRIBUTION` | ✅ Yes | Public data requiring source attribution |
| `LICENSED_PRIVATE` | ❌ No | Data obtained under license/NDA |
| `USER_PRIVATE` | ❌ No | User-supplied private data |
| `INTERNAL_RESTRICTED` | ❌ No | Internal workspace intelligence |

## Publication Gate Rules

The `publication` package owns the boundary. Public-facing surfaces (API,
frontend, CEGS export) **must** use the gate predicates:

```go
publication.PublicEvidence(e)       // Evidence can appear in public products
publication.PublicClaim(c)          // Claim can be a public canonical record
publication.PublicOpportunity(o)    // Opportunity visible in public API
publication.PublicCapitalNeed(n)    // Capital need visible in public API
publication.PublicMilestone(m)      // Milestone visible in public API
publication.PublicCapitalItem(c)    // Capital item visible in public API
```

A record passes the gate **only if all three conditions are met**:
1. `Visibility` is `PUBLIC` or `PUBLIC_WITH_ATTRIBUTION`
2. `Publishable` is `true`
3. `PublicationState` is `PUBLIC_CANONICAL` (for claims, opportunities, capital needs)

## Private→Public Corroboration Workflow

```
┌─────────────────┐
│ Private Source   │ e.g., prospectus, NDA materials
│ (USER_PRIVATE)   │
└────────┬────────┘
         │ ExtractCards()
         ▼
┌─────────────────┐
│ CandidateProject│ Visibility: PRIVATE
│ + Claims        │ Publishable: false
│ + CapitalNeeds  │ PublicationState: PRIVATE_ONLY
└────────┬────────┘
         │ ReconcileClaims()
         ▼
┌─────────────────────────────────────────────┐
│ Search public corpus for matching claims     │
│ (IAAC, CER, CanadaBuys, NRCan, etc.)       │
└────────┬────────────────────────┬───────────┘
         │ No match              │ Match found
         ▼                      ▼
┌─────────────────┐   ┌──────────────────────┐
│ Stays PRIVATE   │   │ Public claim promoted │
│ No publication  │   │ to PUBLIC_CANONICAL   │
│ Original stays  │   │ Private stays PRIVATE │
│ PRIVATE_ONLY    │   │ Audit trail created   │
└─────────────────┘   └──────────────────────┘
```

**Critical invariants:**
1. A private claim is **never** directly promoted to public
2. Only the matching public-source claim gets promoted
3. The private claim remains `PRIVATE_ONLY` even after corroboration
4. An `AuditEntry` records every promotion decision

## Contact Data Policy

The system **never** stores, processes, or publishes:
- Personal email addresses
- Personal phone numbers
- Personal social media profiles
- CRM subscription data
- Lead scoring or personalization data

Entity records use only public corporate identifiers (LEI, NEBN, ticker
symbols, public websites).

## DLP Testing

Every release must pass the DLP test suite:
- `publication/gate_test.go` — Unit tests for visibility predicates
- API-level tests verify no private data appears in any public endpoint

```bash
go test -v ./internal/publication/...
```
