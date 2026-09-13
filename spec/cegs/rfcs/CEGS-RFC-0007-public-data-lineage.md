# CEGS-RFC-0007: Public Sources and Dataset Lineage

- **RFC Number:** CEGS-RFC-0007
- **Title:** Public Source Resources and Versioned Data Lineage
- **Status:** Accepted (additive)
- **Created Date:** 2026-09-13
- **Specification Version:** CEGS 0.1

---

## 1. Summary

This RFC adds one implementation-neutral `source` resource for describing public catalogues, datasets, distributions, APIs, feeds, repositories, documents, and web pages. It also defines optional evidence-lineage identifiers for the source, source version, parser, mapping, and normalization pipeline that produced a claim.

## 2. Motivation

Evidence URLs alone cannot answer which catalogue discovered a resource, which version was normalized, or which mapping produced a claim. Implementations need portable lineage without standardizing their database, queue, crawler, or object-storage architecture.

## 3. Source resource

`source` uses the normal CEGS envelope and the fields defined by `schemas/source.schema.json`. `source_kind` captures its position in the usual publisher → catalogue → dataset → resource hierarchy. `parent_source` and typed CEGS relationships may express hierarchy, mirrors, supersession, references, and bilingual sibling representations.

Lifecycle and health are independent:

- lifecycle describes review and activation (`DISCOVERED` through `ACTIVE`, `BLOCKED`, or `RETIRED`);
- health describes the latest measured operating condition and remains `UNKNOWN` until checked.

Registration, testing, approval, and activation MUST NOT be inferred from one another.

## 4. Authority

Source metadata permits authority tiers 1–5. Tier 5 is a discovery lead only and MUST NOT establish a canonical CEGS claim. The Evidence resource continues to permit only tiers 1–4.

Publisher authority and dataset quality are distinct. A Tier 1 publisher may release an incomplete or stale dataset, and quality measures MUST NOT be represented as publisher authority.

## 5. Evidence lineage extension

Until a future major version makes lineage fields normative, implementations SHOULD place these optional keys under `extensions.ca.opengraph.public_data_lineage`:

```json
{
  "source_id": "cegs:source:ca:example-resource",
  "source_version_id": "sha256:...",
  "source_record_id": "upstream-record-42",
  "locator": "row=42",
  "parser_version": "csv-v2",
  "mapping_version": "project-map-v3",
  "pipeline_version": "cog-ingest-v4"
}
```

Values describe reproducibility and do not prove factual correctness or publisher authenticity.

## 6. Temporal semantics

Source publication/effective time and system observation time MUST remain separate. A source change MAY be represented as an Event whose `subject` is a Source ID. It MUST NOT be silently collapsed into the real-world occurrence date of an economic event.

## 7. Security and compliance

The resource contains public metadata only. Credentials, request headers, private checkpoints, internal object locations, unredacted failures, and bypass instructions MUST NOT be exported. Public accessibility MUST NOT be interpreted as permission to redistribute raw content.

## 8. Compatibility

This is additive. Existing CEGS 0.1 resources remain valid. Implementations that do not support `source` may preserve source identifiers in namespaced extensions.
