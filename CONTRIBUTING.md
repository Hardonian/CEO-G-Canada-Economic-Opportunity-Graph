# Contributing to CanadaOpportunityGraph & CEGS

Thank you for your interest in contributing to **CanadaOpportunityGraph (COG)** and the **Canada Economic Graph Schema (CEGS)**.

We welcome contributions from software engineers, data engineers, researchers, public policy analysts, infrastructure specialists, and institutional stakeholders across Canada and globally.

---

## 1. Principles of Contribution

1. **Anti-Theatre & Real Data**: Every contribution must adhere to strict provenance requirements. We do not accept synthetic data presented as real, hallucinated metrics, non-functional UI scaffolding, or mock APIs masquerading as features.
2. **Standard vs. Implementation Separation**:
   - Changes to the data specification belong in `spec/cegs/` and must follow the RFC process outlined in `spec/cegs/rfcs/RFC-TEMPLATE.md`.
   - Changes to the platform engine, adapters, scoring models, and CLI belong in `internal/`, `adapters/`, `cmd/`, and `apps/web/`.
3. **Deterministic Scoring**: Numerical scores (Buildability, Investability, Supplierability, Sovereignty) must be mathematically reproducible via Go algorithms (`internal/scoring/`). Do not submit non-deterministic LLM-generated scores as canonical.

---

## 2. Development Workflow

### Prerequisites
- **Go**: 1.22+ (tested with Go 1.26)
- **Node.js**: 20+ & **pnpm**: 9+
- **Git**

### Initial Setup
```bash
# Clone the repository
git clone https://github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph.git
cd CEO-G-Canada-Economic-Opportunity-Graph

# Validate Go toolchain and run tests
go test ./...

# Validate CEGS schemas and fixtures
go test -v ./internal/cegs

# Install web dependencies and test build
cd apps/web
pnpm install
pnpm build
cd ../..
```

### Automation Helper (`scripts/task.ps1` or `Makefile`)
- Windows PowerShell:
  - `.\scripts\task.ps1 test` — Run all Go tests
  - `.\scripts\task.ps1 cegs-validate` — Validate CEGS examples and datasets
  - `.\scripts\task.ps1 web-build` — Production build of the web app
  - `.\scripts\task.ps1 demo` — Run the CLI demonstration
- Linux / macOS:
  - `make test`
  - `make build`
  - `make web-build`

---

## 3. Contributing a Data Adapter

Adapters live in `adapters/<source_name>/` and implement the `adapters.SourceAdapter` interface:
```go
type SourceAdapter interface {
    Name() string
    SourceClass() string
    FetchLatest(ctx context.Context) ([]RawRecord, error)
    Transform(raw RawRecord) (*domain.Project, []domain.Event, []domain.Evidence, error)
}
```

Requirements for new adapters:
- **Statutory Authority**: Source must be a Tier 1 government registry or statutory public disclosure (e.g. IAAC, CER, CanadaBuys, NRCan, SEDAR+, provincial gazettes).
- **Cryptographic Hashing**: All raw records must be hashed using SHA-256 (`adapters.HashDocument`) to prevent duplicate writes.
- **Epistemic State**: Facts must be tagged as `VERIFIED` (statutory registry) or `SUPPORTED` (official proponent publication).
- **Deterministic Fixtures**: Provide a test fixture under `data/fixtures/<source_name>.json`.

---

## 4. Submitting Pull Requests

1. Create a feature branch: `git checkout -b feat/my-new-adapter`.
2. Ensure all tests pass: `go test -v ./...`.
3. Ensure the CEGS validator passes: `.\scripts\task.ps1 cegs-validate` (or `go test ./internal/cegs`).
4. Ensure the web application compiles without warnings: `pnpm --dir apps/web build`.
5. Keep commits atomic and descriptive following Conventional Commits (`feat:`, `fix:`, `docs:`, `spec:`).
6. Open a PR with detailed test notes and evidence sources.

---

## 5. Security & Responsible Disclosure

If you discover a vulnerability or security flaw, please review our [Security Policy](file:///docs/SECURITY.md) and report it to `security@canadaopportunitygraph.ca`. Do not open public issues for security exploits.
