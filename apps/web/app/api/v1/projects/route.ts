import { SNAPSHOT_MANIFEST, SNAPSHOT_PROJECTS } from "@/lib/data";
import { publicJSON, publicOptions } from "@/lib/public-api";

function boundedInteger(value: string | null, fallback: number, maximum: number): number {
  if (!value || !/^\d+$/.test(value)) return fallback;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) ? Math.min(maximum, parsed) : fallback;
}

export function GET(request: Request) {
  const search = new URL(request.url).searchParams;
  const limit = Math.max(1, boundedInteger(search.get("limit"), 100, 500));
  const offset = boundedInteger(search.get("offset"), 0, 1_000_000);
  const query = search.get("q")?.trim().toLocaleLowerCase("en-CA");
  const sector = search.get("sector")?.trim().toLocaleLowerCase("en-CA");
  const province = search.get("province")?.trim().toLocaleLowerCase("en-CA");
  const stage = search.get("stage")?.trim().toLocaleLowerCase("en-CA");
  const filtered = SNAPSHOT_PROJECTS.filter((project) => {
    if (query && !`${project.name} ${project.summary} ${project.proponent_name ?? ""}`.toLocaleLowerCase("en-CA").includes(query)) return false;
    if (sector && project.sector.toLocaleLowerCase("en-CA") !== sector) return false;
    if (province && project.province.toLocaleLowerCase("en-CA") !== province) return false;
    if (stage && project.current_stage.toLocaleLowerCase("en-CA") !== stage) return false;
    return true;
  });
  return publicJSON({
    projects: filtered.slice(offset, offset + limit),
    total: filtered.length,
    limit,
    offset,
    source_mode: "BUNDLED_REVIEWED_SNAPSHOT",
    generated_at: SNAPSHOT_MANIFEST.generated_at,
  });
}

export const OPTIONS = publicOptions;
