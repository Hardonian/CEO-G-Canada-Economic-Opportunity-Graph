# COG Project Status & Next Steps

## Quick Summary

All Phase 1, 2, and 3 roadmap items are **complete**. All 5 optimization areas are implemented.
`go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` all pass across 40+ packages.

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
| GraphQL API | `internal/graphql/` | Zero-dependency hand-rolled GraphQL executor |
| CEGS Conformance | `internal/cegs/conformance_test.go` | 36 tests covering all 9 resource types |
| Rust SDK | `sdk/rust/` | CogClient crate with reqwest/serde/thiserror |

### Optimization Areas Completed

| Area | Work | Status |
|------|------|--------|
| 1 — Verifier Persistent Store + REST | `internal/verifier/store.go`, `internal/api/verifier_handler.go`, `internal/verifier/verifier_test.go` | ✅ |
| 2 — GraphQL Field Projection | `internal/graphql/server.go` (parseQuery captures sub-fields), `internal/graphql/resolver.go` (marshalX accept `[]string`), 6 new projection tests | ✅ |
| 3 — MemoryStore Secondary Indexes | `signalsByProject`, `eventsByProject` indexes; `cachedTotalCapex` + `capexDirty` cache; `recomputeCapexLocked()` | ✅ |
| 4 — Adapter Sandbox Workflow + API | `Approve()`, `Reject()`, `Delete()` methods; `internal/api/sandbox_handler.go` with 3 REST endpoints; 10 registry tests | ✅ |
| 5 — Metrics + Makefile | 5 more Prometheus gauges; `bench`, `vet`, `lint` Makefile targets; CEGS 1.0 startup log | ✅ |

### Entry Points Wired

- **`cmd/api/main.go`** — connector registry, reconciliation, merkle, sources config, verifier endpoints, sandbox endpoints, CEGS 1.0 log
- **`cmd/worker/main.go`** — gazette polling (5-min interval, race-safe), reconciliation, merkle, indigenous linker
- **`cmd/cog/main.go`** — `extract` subcommand for documentintelligence

### API Endpoints Added (`internal/api/server.go`)

- `GET /api/v1/forecast/projects/{id}` — project forecast
- `GET /api/v1/forecast/portfolio` — portfolio forecast
- `GET /api/v1/projects/{id}/fit/{archetype}` — archetype fit scoring
- `GET /api/v1/projects/{id}/precedents` — deal precedents
- `GET /api/v1/reconciliation` — multi-jurisdiction reconciliation report
- `GET /api/v1/ai-sovereignty` — AI sovereignty benchmarks
- `GET /api/v1/rankings/{dimension}` — buildability rankings
- `GET /api/v1/verifier/attestations` — paginated attestation list
- `GET /api/v1/verifier/milestones/{id}` — quorum vote result
- `GET /api/v1/adapters` — sandbox adapter catalog
- `POST /api/v1/adapters/{name}/approve` — approve adapter (admin secret)
- `DELETE /api/v1/adapters/{name}` — remove adapter (admin secret)
- `GET/POST /api/v1/graphql` — enterprise GraphQL API

---

## Key Conventions to Follow

- **Middleware signatures**: `middleware.Retry(RetryConfig)`, `middleware.CircuitBreaker(CircuitBreakerConfig)`, `middleware.Cache(CacheConfig)`, `middleware.Dedupe()`, `middleware.Instrumented()`
- **Store interface**: `s.store.ListEntities()` returns 2 values (entities, err). `s.store.ListProjects(ctx, filter)` returns 3 values (projects, count, err).
- **Domain types**: `domain.RequirementConfidence` (not `RequirementClass`) for `Requirement.Confidence`. `domain.SourceTier` is `int` (1-4), not string.
- **Forecast Context**: fields are `Project`, `Events`, `CapitalItems`, `Relationships`, `Procurements`, `Opportunities` (not `Capital`).
- **Reconciliation**: `Reconcile(ctx, store)` returns `*ReconciliationReport` with `TotalRecords` and `Summary` (containing `Merged`, `Linked`, `Conflicts`).
- **CEGS**: `SpecVersion = "0.1"` in `internal/cegs/types.go`. Migration toolkit uses `MigrationVersion = "cegs-migration-v1.0"`.
- **RadarStats**: fields are `AcceleratingProjectsCount` and `StalledProjectsCount` (not `AcceleratingProjects`/`StalledProjects`).

## Verification Commands

```bash
go build ./...              # all packages compile
go vet ./...                # static analysis
go test ./...               # full test suite
go test -race ./...         # race detector
go test -bench=. ./...      # benchmarks
make verify                 # full pipeline (build, vet, test, bench, release-check, cegs-validate, demo, web-build)
```

## Relevant Files

- `ROADMAP.md` — strategic roadmap (phases 1-3, all [x])
- `Makefile` — build/test/verify/bench/vet/lint targets
- `HANDOFF.md` — this file
- `internal/api/server.go` — API routes and handlers
- `cmd/worker/main.go` — worker daemon with polling, reconciliation, merkle, indigenous linker
- `cmd/api/main.go` — API server wiring
- `cmd/cog/main.go` — CLI with `extract` subcommand
- `internal/graphql/` — GraphQL executor, resolver, schema, tests
- `internal/verifier/` — verifier nodes, attestation store, tests
- `internal/adaptersandbox/` — community adapter registry, tests
- `internal/cegs/` — migration toolkit, conformance suite
- `sdk/rust/` — Rust SDK (CogClient crate)