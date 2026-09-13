# Security Architecture & Threat Model

CanadaOpportunityGraph ingests external public documents, parses structured feeds, and serves an open API. This document details the threat model and defenses.

---

## 1. Threat Model & Mitigations

### A. SSRF (Server-Side Request Forgery) in Ingestion Adapters

- **Risk**: Malicious source URLs could target internal cloud metadata services (`169.254.169.254`) or local loopback interfaces (`127.0.0.1`).
- **Mitigation**: Adapters must strictly whitelist permitted statutory hostnames (`iaac-aeic.gc.ca`, `apps.cer-rec.gc.ca`, `canadabuys.canada.ca`, `natural-resources.canada.ca`). Private IP ranges (RFC 1918) and link-local ranges are blocked.

### B. Prompt Injection from Ingested Documents

- **Risk**: External PDF, HTML, or filings containing adversarial prompt injection intended to subvert entity extraction or scoring.
- **Mitigation**:
  1. Authoritative scores are calculated strictly deterministically in Go (`internal/scoring`), completely bypassing LLMs.
  2. Document text is treated as untrusted data input with bounded length limits and strict escaping before parsing.

### C. Stored XSS & HTML Injection

- **Risk**: Malicious project summaries or tender descriptions containing `<script>` or event handlers.
- **Mitigation**: Next.js automatically escapes React JSX outputs. API endpoints serve strict `Content-Type: application/json; charset=utf-8`.

### D. Resource Attribution & Rate Limiting

- **Risk**: Denial-of-service via unbounded queries or scraping loops.
- **Mitigation**: API endpoints enforce default pagination ceilings (`limit=50`, max 500), and requests are tracked via request IDs and Prometheus metrics.

---

## 2. Privacy & Personal Data Minimization

- The platform exclusively tracks **projects, organizations, public programs, and official corporate positions**.
- Personal dossiers on individuals are strictly prohibited.
- Public officials and corporate officers are referenced solely in their professional, public-record capacities.
