# Investment Ontology

CanadaOpportunityGraph uses a structured ontology to classify capital needs,
counterparty types, opportunity kinds, and project requirements. Every
classification is deterministic, evidence-linked, and versioned.

## Capital Need Taxonomy

Capital needs classify what a project is seeking. A project may carry multiple
simultaneous capital needs.

| Need Type | Description |
|-----------|-------------|
| `EQUITY` | General equity investment |
| `DEBT` | General debt financing |
| `SENIOR_DEBT` | Senior secured debt |
| `SUBORDINATED_DEBT` | Mezzanine or subordinated debt |
| `PROJECT_FINANCE` | Non-recourse project financing |
| `INFRASTRUCTURE_EQUITY` | Infrastructure fund equity |
| `PRIVATE_CREDIT` | Private credit facilities |
| `JOINT_VENTURE` | Joint venture partnership |
| `STRATEGIC_INVESTMENT` | Strategic corporate investment |
| `GOVERNMENT_SUPPORT` | Government co-investment or support |
| `GRANT` | Non-repayable government grants |
| `LOAN_GUARANTEE` | Government loan guarantees |
| `EXPORT_CREDIT` | Export credit agency financing |
| `INDIGENOUS_EQUITY` | Indigenous community equity participation |
| `PENSION_CAPITAL` | Pension fund investment |
| `SOVEREIGN_CAPITAL` | Sovereign wealth fund investment |
| `OFFTAKE` | Long-term purchase agreements |
| `ANCHOR_TENANT` | Anchor tenant commitment |
| `PREPAYMENT` | Pre-payment or advance purchase |
| `STREAMING` | Streaming royalty agreements |
| `ROYALTY` | Royalty financing |
| `CONCESSION` | Government concession arrangements |
| `LEASE` | Equipment or land lease financing |
| `TOLLING` | Tolling arrangements |
| `SERVICE_AGREEMENT` | Service or capacity agreements |
| `FIRM_TRANSPORTATION` | Firm transportation capacity |
| `COMMERCIAL_PARTNERSHIP` | Commercial partnership agreements |
| `TECHNOLOGY_PARTNERSHIP` | Technology licensing or partnership |
| `DEVELOPMENT_CAPITAL` | Early-stage development funding |

## Counterparty Taxonomy

Counterparty types classify who might fill a capital need.

| Type | Typical Mandate |
|------|----------------|
| `INFRASTRUCTURE_FUND` | Long-duration, contracted infra assets |
| `PENSION_FUND` | Liability-matching, inflation-linked returns |
| `PRIVATE_EQUITY` | Development-stage, value-add, growth |
| `BANK` | Senior debt, project finance, working capital |
| `PRIVATE_CREDIT` | Subordinated debt, mezzanine, unitranche |
| `EXPORT_CREDIT_AGENCY` | Cross-border export facilitation |
| `SOVEREIGN_WEALTH_FUND` | Strategic diversification, long-duration |
| `STRATEGIC_CORPORATE` | Vertical integration, supply security |
| `EPC_CONTRACTOR` | Design-build, turnkey execution |
| `OEM` | Equipment manufacture and supply |
| `OPERATOR` | Operations and maintenance |
| `OFFTAKER` | Long-term purchase commitment |
| `UTILITY` | Grid integration, power purchase |
| `ANCHOR_TENANT` | First or primary facility user |
| `LOGISTICS_OPERATOR` | Supply chain and transportation |
| `TERMINAL_OPERATOR` | Port and terminal operations |
| `TECHNOLOGY_PROVIDER` | Technology licensing, R&D partnership |
| `INDIGENOUS_PARTNER` | Indigenous community partnership |
| `GOVERNMENT` | Government program participation |
| `JOINT_VENTURE_PARTNER` | Co-development and shared risk |

## Opportunity Types

| Kind | Description |
|------|-------------|
| `CAPITAL` | Direct capital deployment opportunity |
| `OFFTAKE` | Product or service purchase opportunity |
| `TENANCY` | Anchor tenant or capacity user |
| `PARTNERSHIP` | Infrastructure co-development |
| `PROCUREMENT` | Supply chain or service provision |

## Opportunity Funnel States

```
DISCOVERED → QUALIFYING → RESEARCHING → CORROBORATED → ACTIVE_OPPORTUNITY → FINANCING_IN_PROGRESS → CLOSED
                                                                         ↘ STALE
                                                                         ↘ DEAD
```

## Investment Readiness Methodology

The Investment Readiness Vector assesses a project across 8 dimensions:

1. **Engineering** — Has FEED/detailed engineering been completed?
2. **Regulatory** — Environmental approval, permits received?
3. **Financing** — Capital committed, financial close?
4. **Commercial** — Offtake signed, anchor customer?
5. **Site** — Land secured, site selected?
6. **Infrastructure** — Grid connection, road/rail access?
7. **Offtake** — Long-term purchase agreements?
8. **Execution** — EPC awarded, construction started?

Each factor is scored 0-100 based on milestone completion. The overall
readiness score is the unweighted mean of all 8 factors.

## Capital Readiness Methodology

Capital Readiness answers a distinct question: "How prepared is this project
to absorb the capital it says it needs?"

Factors (weighted):
- Engineering maturity (18%)
- Regulatory readiness (15%)
- Commercial foundation (15%)
- Site control (10%)
- Committed capital / gap closure (17%)
- FID proximity (10%)
- Financing specificity (15%)

## Velocity Scoring

Velocity scores measure the *rate* of progress, not the absolute level:

- **Capital Velocity**: Time-decayed financing events, gap closure, milestone progress
- **Execution Velocity**: Milestone completion rate, physical progress, stage advancement

Trends: `ACCELERATING` | `STEADY` | `DECELERATING` | `STALLED`

## Source Policy

See [SOURCE_POLICY.md](SOURCE_POLICY.md) for visibility class definitions and
publication boundary rules.
