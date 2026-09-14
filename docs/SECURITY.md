# Security Architecture & Threat Model

This is an implementation threat model, not a certification, Privacy Impact Assessment, Threat and Risk Assessment, or Authority to Operate.

## Trust boundaries and implemented controls

### Source ingestion

- Live NRCan retrieval is opt-in and restricted to HTTPS on the official `maps-cartes.services.geo.ca` service path.
- Redirects, body size, record count and required schema fields are bounded before normalization.
- Checked-in snapshots make default execution hermetic and reviewable.
- Source text is untrusted data. Deterministic scoring and forecasting do not execute document instructions or send source content to a model provider.

### Public API

- GET/HEAD/OPTIONS-only routes, strict method handling, bounded targets/query strings/bodies, allowlisted filters/sorts and resource identifier validation.
- Configurable exact-origin CORS, with wildcard rejected for hardened production configuration.
- Trusted-proxy-aware, bounded-memory token-bucket limiting; forwarded client addresses are ignored unless the connected peer is trusted.
- Request deadlines, readiness deadlines, server read/header/write/idle timeouts and graceful shutdown.
- UUID request IDs are generated when incoming values are malformed; panics are logged only with the request ID and return a redacted error.
- `nosniff`, anti-framing CSP, referrer policy, permissions policy, optional HSTS and no-store API responses.
- Readiness returns failure when its dependency check cannot complete; metrics do not silently convert dependency errors into healthy output.

### Web and container runtime

- Next.js emits a standalone image with CSP and browser security headers and hides framework identification.
- API and web containers run as non-root with dropped capabilities, `no-new-privileges`, bounded PIDs and read-only roots plus constrained temporary filesystems.
- Docker Compose binds public development ports to loopback and explicitly selects the stateless snapshot storage mode.
- JavaScript dependency versions are locked; the current production audit reports no known vulnerabilities.

## AI and forecast safety

The forecast engine is local, deterministic and versioned. It emits assumptions, input hash, data coverage, evidence references, sensitivity ranges, warnings and limitations. Likelihood values are planning indices—not calibrated probabilities, investment advice, cabinet advice, procurement authority, credit ratings or model-generated facts.

Any future generative layer must be downstream of this evidence contract, treat retrieved text as hostile, cite only allowlisted evidence IDs, never upgrade `INFERRED` output to observed fact, and be deployable with outbound network access disabled.

## Residual risks

- The shipped server is a read-only snapshot projection. It rejects `DATABASE_URL`; durable mutations and cross-instance ingestion remain unsupported until a transactional store, migrations, backup/restore and row-level authorization are implemented.
- Rate limiting is process-local and must move to a shared sovereign service for horizontally scaled enforcement.
- CSP permits inline scripts/styles required by the present Next.js build. A nonce-based policy is a future hardening item.
- Source hashes are not publisher signatures. Independent verification and an externally witnessed transparency log remain roadmap items.
- Government deployment still requires organizational security categorization, privacy review, control assessment and explicit authorization by the responsible authority.
