import { SNAPSHOT_MANIFEST, SNAPSHOT_PROJECTS, SNAPSHOT_RADAR_STATS } from "@/lib/data";
import { publicJSON, publicOptions } from "@/lib/public-api";

export const dynamic = "force-static";

export function GET() {
  return publicJSON({
    stats: SNAPSHOT_RADAR_STATS,
    accelerating_projects: SNAPSHOT_PROJECTS.filter((project) => (project.scores?.buildability ?? 0) >= 50).slice(0, 25),
    recent_signals: [],
    cegs_version: "0.1",
    source_mode: "BUNDLED_REVIEWED_SNAPSHOT",
    generated_at: SNAPSHOT_MANIFEST.generated_at,
  });
}

export const OPTIONS = publicOptions;
