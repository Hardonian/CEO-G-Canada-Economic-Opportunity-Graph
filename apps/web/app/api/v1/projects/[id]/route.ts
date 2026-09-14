import { SNAPSHOT_PROJECTS } from "@/lib/data";
import { publicJSON, publicOptions } from "@/lib/public-api";

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const project = SNAPSHOT_PROJECTS.find((item) => item.slug === id || item.id === id);
  if (!project) return publicJSON({ error: "Project not found" }, { status: 404 });
  return publicJSON({ project, source_mode: "BUNDLED_REVIEWED_SNAPSHOT" });
}

export const OPTIONS = publicOptions;
