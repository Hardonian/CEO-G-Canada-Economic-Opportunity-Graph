# Authoritative Data Sources & Provenance Policy

CanadaOpportunityGraph maintains an uncompromising **Anti-Theatre Provenance Policy**. All primary project data, capital figures, tenders, and regulatory milestones must originate from statutory, audit-grade public disclosures.

---

## 1. Primary Statutory Sources

| Source Registry | Agency / Authority | Jurisdiction | Data Coverage | Ingestion Frequency |
| :--- | :--- | :--- | :--- | :--- |
| **Canadian Impact Assessment Registry (CIAR)** | Impact Assessment Agency of Canada (IAAC) | Federal (All Provinces/Territories) | Major project environmental assessments, public comments, project descriptions, regulatory decisions | Daily batch |
| **CanadaBuys (Ariba / Buyandsell)** | Public Services and Procurement Canada (PSPC) | Federal | Active tenders, RFP notices, standing offers, awarded contracts | Real-time & 6h polling |
| **Canada Energy Regulator (CER)** | CER / Régie de l'énergie du Canada | Federal / Interprovincial | Pipelines, power lines, offshore energy, import/export permits | Weekly update |
| **NRCan Major Projects Inventory** | Natural Resources Canada (NRCan) | Federal / National | Resource & clean energy projects >$50M CAPEX | Quarterly release reconciliation |
| **IDEaS Defence Innovation** | Department of National Defence (DND) / CAF | Federal | Defence procurement challenges, sandbox calls, competitive project awards | Bi-weekly check |
| **SEDAR+ Regulatory Filings** | Canadian Securities Administrators (CSA) | National Capital Markets | Technical reports (NI 43-101, NI 51-101), prospectus disclosures, material changes | On-demand / Material events |

---

## 2. Epistemic Verification States

Every factual claim in CanadaOpportunityGraph is assigned an explicit verification confidence tier:

* **`VERIFIED`**: Certified directly against a primary statutory government registry or regulatory filing (e.g. IAAC decision, CanadaBuys tender notice).
* **`SUPPORTED`**: Confirmed by formal proponent publication (investor presentation, audited financial report, official press release).
* **`INFERRED`**: Derived mathematically or logically through deterministic domain ontology rules (e.g., downstream electrical substation requirement derived from a 300MW data centre build).
* **`CONFLICTED`**: Multiple authoritative sources report contradictory figures (e.g., competing CAPEX estimates between provincial regulator and proponent).
* **`UNKNOWN`**: Data is unavailable in public registries. Missing data is never synthetically generated or hallucinated.
* **`STALE`**: No official filing or status change detected within the last 180 days.

---

## 3. Cryptographic Immutability

1. Every ingested document is hashed using SHA-256 upon initial retrieval.
2. The hash, retrieval timestamp, canonical source URI, and HTTP payload headers are recorded in `domain.Evidence`.
3. If an upstream record changes, a new discrete immutable `domain.Event` is appended. Historic records are never overwritten.
