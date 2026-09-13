# Sources & Adapter Ingestion Details

This document provides engineering details on how external government APIs and document sources are ingested and parsed.

---

## 1. Adapter Implementation Matrix

### IAAC Adapter (`adapters/iaac/`)
- **Upstream Registry**: Impact Assessment Agency of Canada API & open data catalog.
- **Entity Extracted**: Major physical projects subject to federal impact assessment.
- **Key Fields**: Project name, registry ID, province, stage (Planning, Impact Statement, Decision), proponent name, environmental coordinates.
- **Cryptographic Hash**: Computed over raw JSON registry dump or HTML decision summary.

### CanadaBuys Adapter (`adapters/canadabuys/`)
- **Upstream Registry**: CanadaBuys API (PSPC / Buyandsell Open Data).
- **Entity Extracted**: Active procurement tenders, solicitations, GSIN / UNSPSC classifications.
- **Key Fields**: Solicitation number, closing date, procuring organization, value estimate, buyer contact, status.

### NRCan Major Projects Adapter (`adapters/nrcan_major_projects/`)
- **Upstream Registry**: Natural Resources Canada Major Projects Inventory (Open Government Portal).
- **Entity Extracted**: Energy, mining, clean tech, and infrastructure projects over $50M CAD.
- **Key Fields**: Project name, commodity, estimated capital ($M CAD), construction start/end year, jobs created.

### IDEaS Defence Adapter (`adapters/ideas_defence/`)
- **Upstream Registry**: Innovation for Defence Excellence and Security (Department of National Defence).
- **Entity Extracted**: Defence procurement challenges, sandbox demonstrations, targeted defense innovation funding.
- **Key Fields**: Challenge ID, challenge title, funding tier, defense requirement focus, closing date.

### Canada Energy Regulator Adapter (`adapters/cer/`)
- **Upstream Registry**: CER Facility & Pipeline Registry.
- **Entity Extracted**: Interprovincial/international pipelines, electrical transmission facilities, offshore infrastructure.
- **Key Fields**: Facility ID, regulatory status, commodity, capacity, route length, proponent name.

---

## 2. Ingestion Resilience & Error Handling

- **Exponential Backoff**: Upstream rate limits or temporary downtime trigger exponential backoff with jitter (max 5 retries).
- **Schema Drift Detection**: Raw records failing structural validation are sequestered into a dead-letter log with full error traceback.
- **Local Hermetic Fixtures**: Offline deterministic fixtures under `data/fixtures/` allow 100% full-featured offline test execution without live internet dependencies.
