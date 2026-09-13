# Canadian Sovereign Deployment Profile

This profile defines a target operating posture for investors, Crown corporations and public agencies. It is a deployment blueprint, not a claim that the repository is certified for Protected A/B, classified information, personal information, or any departmental workload.

## Non-negotiable sovereignty controls

| Control objective | Target evidence | Repository support |
| :--- | :--- | :--- |
| Canadian data and backup residency | Contractual region list, data-flow diagram, restore test | Offline source snapshots and no required forecast egress |
| Canadian-controlled encryption | Customer-managed keys, HSM location, rotation and break-glass records | Deployment responsibility; no keys are embedded |
| Jurisdiction visibility | Processor/subprocessor register and foreign-law assessment | Sovereignty scoring model; legal review remains external |
| Least privilege and strong identity | Federated MFA, workload identities, quarterly access review | Containers are non-root; application IAM is not yet implemented |
| Evidence integrity | Source IDs, SHA-256 hashes, immutable releases, witnessed roots | Record hashes and create-only releases implemented; witnessed log is roadmap |
| AI containment | Model/data cards, approved endpoints, egress policy, prompt-injection tests | Core forecasts are deterministic, local and emit no raw evidence |
| Bilingual and accessible service | English/French content review and WCAG testing | UI accessibility baseline; full bilingual product coverage remains work |
| Operational authorization | Security plan, assessment findings, residual-risk decision, monitoring plan | Must be completed by the deploying organization |

## Recommended deployment zones

```text
Public Internet
    │
    ▼
Canadian WAF / rate-control tier
    │
    ▼
Stateless web tier ──► read-only public API
                           │
             ┌─────────────┴─────────────┐
             ▼                           ▼
   Canadian transactional store   append-only evidence vault
             │                           │
             └─────────────┬─────────────┘
                           ▼
                 offline analytics zone
                 (deny egress by default)
```

Keep administrative ingestion, evidence review and any future model tooling off the public API network. Use separate identities and keys for each zone, private service endpoints, encrypted backups in a second Canadian region, and tested recovery objectives. Export only the minimum evidence metadata required by public consumers; raw documents may carry separate licence, privacy or security constraints.

## Government adoption gate

Before production use, the deploying authority should at minimum:

1. categorize the information and business process;
2. select a current departmental security/privacy control profile;
3. complete threat, privacy, supply-chain and foreign-jurisdiction assessments;
4. test controls independently and document deficiencies;
5. obtain the responsible authorizer's explicit operating decision;
6. continuously monitor identity, configuration, vulnerabilities, egress, logs and evidence freshness.

Current Canadian references include the Cyber Centre's [ITSP.10.033 security and privacy controls catalogue](https://www.cyber.gc.ca/en/guidance/cyber-security-privacy-risk-management/itsp10033/foreword-overview-introduction), its [assessment, authorization and monitoring guidance](https://www.cyber.gc.ca/en/guidance/cyber-security-privacy-risk-management/itsp10033/assessment-authorization-monitoring), and the Treasury Board [Government of Canada Cloud Guardrails](https://www.tbs-sct.canada.ca/pol/doc-eng.aspx?id=32787). These sources guide an organizational assessment; this project does not self-attest compliance.
