# Canada Economic Graph Schema (CEGS) — Specification v0.1

This document specifies the normative syntax, data structures, and semantic rules of **CEGS 0.1**.

---

## 1. The CEGS Canonical Envelope

Every root CEGS JSON resource must encapsulate its payload within the standard envelope:

```json
{
  "cegs": "0.1",
  "id": "cegs:project:ca:darlington-new-nuclear",
  "type": "project",
  "canonical_name": "Darlington Small Modular Reactor (SMR) Project",
  "jurisdiction": "CA:ON",
  "created_at": "2026-01-15T12:00:00Z",
  "updated_at": "2026-09-10T14:30:00Z",
  "provenance": ["cegs:evidence:ca:iaac-registry-darlington"],
  "extensions": {}
}
```

### Required Envelope Fields:
- `cegs` (string, required): Specification version identifier (e.g. `"0.1"`).
- `id` (string, required): Globally unique, stable CEGS identifier adhering to URI syntax.
- `type` (string, required): CEGS object type (`project`, `organization`, `event`, `relationship`, `evidence`, `source`, `capital_item`, `procurement`, `program`, `opportunity`, `score`, `signal`, `manifest`).
- `canonical_name` (string, required): Standard human-readable name in Canadian English (with optional French official alias).
- `jurisdiction` (string, required): ISO 3166-1/2 code (e.g. `"CA"`, `"CA:ON"`, `"CA:BC"`, `"CA:QC"`).
- `created_at` (string, required): RFC 3339 UTC timestamp of initial recording.
- `updated_at` (string, required): RFC 3339 UTC timestamp of last canonical update.
- `provenance` (array of string): List of referenced Evidence IDs.
- `extensions` (object): Optional vendor/domain namespaced dictionary.

---

## 2. Stable Identifier Syntax

All CEGS identifiers follow the canonical URI scheme:

```text
cegs:<resource_type>:<jurisdiction_code>:<slug_or_id>
```

Examples:
- `cegs:project:ca:on:darlington-new-nuclear`
- `cegs:org:ca:ontario-power-generation`
- `cegs:evidence:ca:iaac-registry-darlington`
- `cegs:event:ca:darlington-fid-approval`
- `cegs:rel:ca:opg-darlington-proponent`
- `cegs:proc:ca:canadabuys-w8486-2601`

Identifiers must:
- Consist exclusively of lowercase alphanumeric characters, colons, hyphens, and underscores (`^[a-z0-9:_-]+$`).
- Remain stable across entity name changes.
- Never expose volatile internal autoincrement primary keys.

---

## 3. Monetary Representation

Monetary figures must never strip currency or assume CAD by default. All amounts are strictly expressed via structured objects:

```json
{
  "amount": 1200000000,
  "currency": "CAD",
  "amount_type": "reported",
  "period": null,
  "range": null
}
```

For capital ranges:
```json
{
  "currency": "CAD",
  "amount_type": "range",
  "range": {
    "min": 900000000,
    "max": 1200000000
  }
}
```

Permitted `amount_type` values:
- `reported`: Sourced directly from official filings or press announcements.
- `estimated`: Calculated by engineers or authoritative industry research.
- `range`: Upper and lower bound provided in primary disclosures.
- `unknown`: Explicitly declared unknown capital requirement.

---

## 4. Temporal Semantics

CEGS preserves distinct temporal milestones to prevent chronological conflation:
- `occurred_at`: The actual real-world moment an event took place.
- `announced_at`: When the event was publicly disclosed.
- `effective_at`: When a legal agreement or regulatory ruling took statutory effect.
- `observed_at`: When an automated ingestion pipeline first captured the disclosure.
- `retrieved_at`: When a source document was fetched and hashed.

---

## 5. Claim Status & Confidence Invariants

Every status assertion must use one of the eight standard states:
- `VERIFIED`: Directly corroborated by a Tier 1 authoritative publisher (Government, Regulator).
- `SUPPORTED`: Corroborated by a Tier 2 issuer disclosure or multiple credible Tier 3 sources.
- `REPORTED`: Stated by a single source without independent regulatory corroboration.
- `INFERRED`: Derived algorithmically via explicit dependency rules.
- `CONFLICTED`: Multiple active source documents assert mutually irreconcilable facts.
- `UNKNOWN`: The fact is undetermined.
- `STALE`: Information has exceeded its validity threshold without recent reconfirmation.
- `RETRACTED`: Sourced document or claim was formally withdrawn by publisher.

**Validation Rule**: An assertion marked `VERIFIED` with a numerical confidence below 0.70 fails semantic validation unless accompanied by an explicit override rationale.

---

## 6. Core Resource Models

### 6.1 Project Resource
Represents a Canadian capital project or infrastructure asset:
- `sector`: Controlled vocabulary (Critical Minerals, Nuclear & Clean Power, AI Compute & Data Centres, Defence & Arctic, etc.).
- `subsector`: Specific technology or resource (e.g. "Small Modular Reactor", "Lithium Hydroxide Processing").
- `stage`: Current lifecycle position (from `DISCOVERED` to `OPERATING` or `CANCELLED`).
- `capex`: Structured monetary object.
- `proponents`: Array of Organization IDs.
- `location`: Structured location object (coordinates, municipality, province).

### 6.2 Organization Resource
Represents corporate issuers, Crown corporations, First Nations, government agencies, utilities, and suppliers:
- `legal_name`: Formal legal registration name.
- `aliases`: Recognized trade names and bilingual acronyms.
- `entity_type`: Controlled vocabulary (`Corporation`, `CrownCorp`, `FirstNation`, `GovernmentAgency`, `Utility`, `Investor`, `Supplier`).
- `identifiers`: External cross-references (`LEI`, `NEBN`, `Ticker`, `SEDAR_ID`).

### 6.3 Event Resource
Append-only log of real-world economic occurrences:
- `event_type`: Namespaced event taxonomy (`project.announced`, `capital.commitment`, `regulatory.approved`, `indigenous.agreement_signed`, etc.).
- `subject`: The primary entity or project ID.
- `occurred_at`: Real-world timestamp.
- `evidence`: Array of Evidence IDs corroborating the event.

### 6.4 Relationship Resource
Typed directional graph edge connecting two entities:
- `relationship_type`: Controlled vocabulary (`owns`, `operates`, `develops`, `finances`, `regulates`, `supplies`, `partners_with`, `indigenous_partner`).
- `from`: Source entity ID.
- `to`: Target entity or project ID.
- `status`: Verification status.
- `valid_from` / `valid_to`: Temporal validity bounds.

### 6.5 Evidence Resource
Cryptographic provenance record:
- `source_url`: Canonical URL.
- `publisher`: Responsible organization.
- `source_tier`: 1 (Authoritative), 2 (Issuer), 3 (Credible News), 4 (Industry).
- `content_hash`: Cryptographic SHA-256 hash of the raw document snapshot.
- `retrieval_timestamp`: UTC timestamp of fetch.
- `confidence`: Semantic confidence rating.

### 6.6 Source Resource

Represents public metadata for a catalogue, dataset, distribution, API, feed, repository, document, or web page:

- `source_kind`: Position in the publisher → catalogue → dataset → resource hierarchy.
- `canonical_url`: Public HTTP(S) identity; credentials and private request configuration are never exported.
- `source_family` and `access_method`: Protocol and retrieval semantics.
- `authority_tier`: Publisher/source authority from 1–5. Tier 5 is a discovery lead and cannot establish canonical Evidence.
- `lifecycle`: Review/activation state, kept distinct from `health`.
- `health`: Latest measured operating condition; `UNKNOWN` until an access check has occurred.
- `incremental_capabilities`: Public protocol capabilities such as ETag, Last-Modified, cursor, delta, or change feed.

See CEGS-RFC-0007. Registration, testing, approval, activation, and health MUST NOT be inferred from one another.
