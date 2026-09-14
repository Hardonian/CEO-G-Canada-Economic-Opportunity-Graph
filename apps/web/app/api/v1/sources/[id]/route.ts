import { getSnapshotSource } from "@/lib/source-data";
import { publicJSON, publicOptions } from "@/lib/public-api";

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const source = getSnapshotSource(id);
  if (!source) return publicJSON({ error: "Source not found" }, { status: 404 });
  return publicJSON({ source, source_mode: "BUNDLED_REVIEWED_SNAPSHOT" });
}

export const OPTIONS = publicOptions;
