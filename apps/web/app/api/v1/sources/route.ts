import { VETTED_SOURCES } from "@/lib/source-data";
import { publicJSON, publicOptions } from "@/lib/public-api";

function boundedInteger(value: string | null, fallback: number, maximum: number): number {
  if (!value || !/^\d+$/.test(value)) return fallback;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) ? Math.min(maximum, parsed) : fallback;
}

export function GET(request: Request) {
  const search = new URL(request.url).searchParams;
  const limit = Math.max(1, boundedInteger(search.get("limit"), 24, 100));
  const offset = boundedInteger(search.get("offset"), 0, 1_000_000);
  const q = search.get("q")?.trim().toLocaleLowerCase("en-CA");
  const publisher = search.get("publisher")?.trim().toLocaleLowerCase("en-CA");
  const jurisdiction = search.get("jurisdiction")?.trim().toLocaleLowerCase("en-CA");
  const sector = search.get("sector")?.trim().toLocaleLowerCase("en-CA");
  const format = search.get("format")?.trim().toLocaleLowerCase("en-CA");
  const family = search.get("family")?.trim().toLocaleLowerCase("en-CA");
  const authority = search.get("authority")?.trim();
  const filtered = VETTED_SOURCES.filter((source) => {
    const searchable = `${source.name} ${source.publisher_name} ${source.description} ${source.subject_tags.join(" ")} ${source.sector_tags.join(" ")}`.toLocaleLowerCase("en-CA");
    if (q && !searchable.includes(q)) return false;
    if (publisher && !source.publisher_name.toLocaleLowerCase("en-CA").includes(publisher)) return false;
    if (jurisdiction && ![source.jurisdiction, ...source.geography].join(" ").toLocaleLowerCase("en-CA").includes(jurisdiction)) return false;
    if (sector && !source.sector_tags.join(" ").toLocaleLowerCase("en-CA").includes(sector)) return false;
    if (format && !source.content_type.toLocaleLowerCase("en-CA").includes(format)) return false;
    if (authority && String(source.authority_tier) !== authority) return false;
    if (family === "api" && !(source.access_method.includes("API") || source.source_family === "CKAN")) return false;
    return true;
  });
  return publicJSON({
    sources: filtered.slice(offset, offset + limit),
    total: filtered.length,
    limit,
    offset,
    source_mode: "BUNDLED_REVIEWED_SNAPSHOT",
  });
}

export const OPTIONS = publicOptions;
