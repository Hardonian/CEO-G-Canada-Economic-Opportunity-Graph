# Sources & Adapter Engineering

## Implemented production path

The default API, CLI and dataset generator ingest local, reproducible inputs:

1. `data/fixtures/nrcan_mpi_2025.json` — a pinned copy of the official NRCan Major Projects Inventory point layer, including source URL, dataset vintage, retrieval time, licence and 295 features.
2. `data/fixtures/official_records.json` — a narrow human-reviewed record set with assertion-level evidence for selected regulatory, construction and capital milestones.

The curated records run after NRCan. When stable project slugs match, the store preserves complementary coordinates, CAPEX lineage, external identifiers and evidence instead of erasing a known value when a later source is silent.

## NRCan normalization

`adapters/nrcan_major_projects/` currently provides the national data path:

- stable UUIDs derived from source ID and namespace;
- record-scoped SHA-256 evidence hashes;
- full province/territory normalization;
- exact CAPEX conversion from reported millions, with unparseable/range-only values left `UNKNOWN`;
- explicit source-sector preservation plus CEGS sector classification;
- deterministic lifecycle mapping;
- readable French-character slug folding;
- an allowlisted live transport with HTTPS, redirect, response-size, record-count and schema constraints.

The source inventory is broad but is not a complete census. Its values are `REPORTED`, not promoted to `VERIFIED` merely because the publisher is authoritative.

## Scaffolded adapters

The IAAC, CER, CanadaBuys, IDEaS and legacy NRCan adapters demonstrate source-specific parsing contracts. Their checked-in fixtures are empty and they are not enabled by default. Claims of automated retry, dead-letter queues, or polling schedules should not be made until those operational controls exist and are tested.

## Release artifacts

`scripts/generate_datasets.go` deterministically creates current and immutable versioned releases:

- domain JSONL for projects, organizations, events, evidence and scores;
- CEGS JSONL equivalents;
- CSV and GeoJSON project exports;
- a SHA-256 checksum manifest with dynamic jurisdiction and record counts.

Historical release directories are create-only: the generator refuses to overwrite an existing version with different bytes.
