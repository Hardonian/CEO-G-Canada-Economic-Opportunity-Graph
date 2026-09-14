import { SNAPSHOT_MANIFEST, SNAPSHOT_PROJECTS } from "@/lib/data";
import { VETTED_SOURCES } from "@/lib/source-data";
import { publicJSON, publicOptions } from "@/lib/public-api";

export const dynamic = "force-static";

export function GET() {
  return publicJSON({
    status: "ok",
    mode: "BUNDLED_REVIEWED_SNAPSHOT",
    generated_at: SNAPSHOT_MANIFEST.generated_at,
    dataset_version: SNAPSHOT_MANIFEST.dataset_version,
    projects: SNAPSHOT_PROJECTS.length,
    canonical_sources: VETTED_SOURCES.length,
    evidence_records: SNAPSHOT_MANIFEST.record_counts.evidence,
  });
}

export const OPTIONS = publicOptions;
