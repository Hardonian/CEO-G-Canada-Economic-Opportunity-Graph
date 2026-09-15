# COG Project Status & Next Steps

## Quick Summary

All Phase 1 and Phase 2 roadmap items are **complete and wired into the entry points**.
`go build ./...`, `go vet ./...`, and `go test ./...` all pass across 40+ packages.

---

## What Was Done This Session

### New Packages Created

| Package | Path | Purpose |
|---------|------|---------|
| Connector Factory | `internal/connector/factory.go` | Wraps adapters with retry/circuit-breaker/cache/dedup/instrumented middleware |
| Merkle Log | `internal/merkle/merkle.go` | Deterministic transparency log over evidence hashes |
| Reconciliation | `internal/reconciliation/` | Multi-jurisdiction record matching with merge/link/conflict actions |
| Gazette Polling | `internal/gazettepoll/poll.go` | Scheduled polling worker with change detection |
| Indigenous Linker | `internal/indigenouslinker/linker.go` | Cross-references ISC Business Directory with projects/procurements |
| CEGS Migration | `internal/cegs/migration.go` | Converts legacy JSON Schemas to CEGS 1.0 canonical form |
| Adapter Sandbox | `internal/adaptersandbox/registry.go` | Community adapter registry with validation |
| Verifier Nodes | `internal/verifier/verifier.go` | Decentralized multi-party milestone notarization |

### Entry Points Wired

- **`cmd/api/main.go`** — connector registry, reconciliation, merkle, sources config
- **`cmd/worker/main.go`** — gazette polling (5-min interval), reconciliation, merkle, indigenous linker
- **`cmd/cog/main.go`** — `extract` subcommand for documentintelligence

### API Endpoints Added (`internal/api/server.go`)

- `GET /api/v1/forecast/projects/{id}` — project forecast
- `GET /api/v1/forecast/portfolio` — portfolio forecast
- `GET /api/v1/projects/{id}/fit/{archetype}` — archetype fit scoring
- `GET /api/v1/projects/{id}/precedents` — deal precedents
- `GET /api/v1/reconciliation` — multi-jurisdiction reconciliation report
- `GET /api/v1/ai-sovereignty` — AI sovereignty benchmarks
- `GET /api/v1/rankings/{dimension}` — buildability rankings

---

## Remaining Roadmap (Phase 3)

### CEGS 1.0 Finalization
- [ ] Lock core vocabulary and release formal migration toolkits
- [ ] Add migration tests for all CEGS resource types (project, organization, event, relationship, evidence, capital, procurement, opportunity, signal)
- [ ] Publish CEGS 1.0 conformance test suite

### Python & Rust SDKs
- [ ] Create `sdks/python/` — CEGS client library for data science and notebooks
- [ ] Create `sdks/rust/` — standalone CEGS client for backend pipelines
- [ ] Both should use the OpenAPI spec at `api/v1/openapi.json` as the source of truth

### Enterprise GraphQL API
- [ ] Add `internal/graphql/` package with graph query support
- [ ] Add subscription webhooks for moving projects (stage changes, FID, construction start)
- [ ] Wire into `cmd/api/main.go` as `/api/v1/graphql` endpoint
- [ ] Fine-grained authZ based on entity visibility

### Community-Contributed Adapters
- [ ] Build adapter registry UI or CLI for submitting community adapters
- [ ] Add adapter validation pipeline (source-tier, URL safety, schema conformance)
- [ ] `internal/adaptersandbox/registry.go` is the foundation — extend it with submission workflow

### Decentralized Verifier Nodes
- [ ] `internal/verifier/verifier.go` is the foundation — add:
  - [ ] Persistent attestation store (DB-backed)
  - [ ] Quorum-gated milestone notarization service
  - [ ] REST endpoint for querying notarized milestones
  - [ ] CLI tool for running a verifier node

---

## Key Conventions to Follow

- **Middleware signatures**: `middleware.Retry(RetryConfig)`, `middleware.CircuitBreaker(CircuitBreakerConfig)`, `middleware.Cache(CacheConfig)`, `middleware.Dedupe()`, `middleware.Instrumented()`
- **Store interface**: `s.store.ListEntities()` returns 2 values (entities, err). `s.store.ListProjects(ctx, filter)` returns 3 values (projects, count, err).
- **Domain types**: `domain.RequirementConfidence` (not `RequirementClass`) for `Requirement.Confidence`. `domain.SourceTier` is `int` (1-4), not string.
- **Forecast Context**: fields are `Project`, `Events`, `CapitalItems`, `Relationships`, `Procurements`, `Opportunities` (not `Capital`).
- **Reconciliation**: `Reconcile(ctx, store)` returns `*ReconciliationReport` with `TotalRecords` and `Summary` (containing `Merged`, `Linked`, `Conflicts`).
- **CEGS**: `SpecVersion = "0.1"` in `internal/cegs/types.go`. Migration toolkit uses `MigrationVersion = "cegs-migration-v1.0"`.

## Verification Commands

```bash
go build ./...              # all packages compile
go vet ./...                # static analysis
go test ./...               # full test suite
go test -race ./...         # race detector
make verify                 # full pipeline (build, seed, test, release-check, cegs-validate, demo, web-build)
```

## Relevant Files

- `ROADMAP.md` — strategic roadmap (phases 1-3)
- `Makefile` — build/test/verify targets
- `internal/api/server.go` — API routes and handlers (1573 lines)
- `cmd/worker/main.go` — worker daemon with polling, reconciliation, merkle, indigenous linker
- `cmd/api/main.go` — API server wiring
- `cmd/cog/main.go` — CLI with `extract` subcommand