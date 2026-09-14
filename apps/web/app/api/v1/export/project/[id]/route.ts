import { SNAPSHOT_PROJECTS } from "@/lib/data";
import { publicJSON, publicOptions, publicText } from "@/lib/public-api";

function markdownForProject(project: (typeof SNAPSHOT_PROJECTS)[number]): string {
  const evidence = project.evidence ?? [];
  return [
    `# ${project.name}`,
    "",
    project.summary,
    "",
    `- CEGS ID: cegs:project:ca:${project.province.toLowerCase()}:${project.slug}`,
    `- Sector: ${project.sector}`,
    `- Stage: ${project.current_stage}`,
    `- Location: ${project.location_name}, ${project.province}`,
    `- Reported CAPEX: ${project.capex_cad > 0 ? `${project.capex_cad} CAD` : "unknown"}`,
    `- Evidence status: ${project.confidence}`,
    `- Last sourced update: ${project.last_meaningful_update}`,
    "",
    "## Evidence",
    "",
    ...(evidence.length > 0
      ? evidence.flatMap((item) => [
          `- ${item.publisher} (Tier ${item.source_tier}): ${item.source_url}`,
          `  - SHA-256: ${item.content_hash}`,
        ])
      : ["- No evidence record is bundled for this project."]),
    "",
    "Generated from the CanadaOpportunityGraph reviewed snapshot. Verify the linked publisher records before making a material decision.",
    "",
  ].join("\n");
}

export async function GET(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const project = SNAPSHOT_PROJECTS.find((item) => item.slug === id || item.id === id);
  if (!project) return publicJSON({ error: "Project not found" }, { status: 404 });
  const format = new URL(request.url).searchParams.get("format")?.toLowerCase();
  if (format === "markdown" || format === "md") {
    return publicText(markdownForProject(project), "text/markdown; charset=utf-8", {
      headers: { "Content-Disposition": `attachment; filename="${project.slug}.md"` },
    });
  }
  if (!format || format === "cegs" || format === "json") {
    return publicJSON({
      cegs: "0.1",
      type: "project",
      id: `cegs:project:ca:${project.province.toLowerCase()}:${project.slug}`,
      source_mode: "BUNDLED_REVIEWED_SNAPSHOT",
      project,
    });
  }
  return publicJSON({ error: "Unsupported format", supported_formats: ["cegs", "json", "markdown"] }, { status: 400 });
}

export const OPTIONS = publicOptions;
