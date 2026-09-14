import { SNAPSHOT_MANIFEST, SNAPSHOT_PROJECTS } from "@/lib/data";
import { publicJSON, publicOptions } from "@/lib/public-api";

export const dynamic = "force-static";

export function GET() {
  return publicJSON({
    cegs: "0.1",
    id: SNAPSHOT_MANIFEST.id,
    type: "dataset",
    generated_at: SNAPSHOT_MANIFEST.generated_at,
    dataset_version: SNAPSHOT_MANIFEST.dataset_version,
    source_mode: SNAPSHOT_MANIFEST.source_mode,
    record_counts: SNAPSHOT_MANIFEST.record_counts,
    projects: SNAPSHOT_PROJECTS.map((project) => ({
      cegs: "0.1",
      id: `cegs:project:ca:${project.province.toLowerCase()}:${project.slug}`,
      type: "project",
      canonical_name: project.name,
      jurisdiction: project.province === "Federal" ? "CA" : `CA:${project.province}`,
      description: project.summary,
      sector: project.sector,
      subsector: project.subsector,
      stage: project.current_stage,
      capex: { amount: project.capex_cad, currency: "CAD", amount_type: project.capex_status?.toLowerCase() ?? "unknown" },
      location: { name: project.location_name, province: project.province, latitude: project.latitude, longitude: project.longitude },
      source_status: project.confidence,
      provenance: (project.evidence ?? []).map((evidence) => `cegs:evidence:ca:${evidence.id}`),
      extensions: { "ca.opengraph.scores": project.scores, "ca.opengraph.score_details": project.score_details },
      updated_at: project.last_meaningful_update,
    })),
  });
}

export const OPTIONS = publicOptions;
