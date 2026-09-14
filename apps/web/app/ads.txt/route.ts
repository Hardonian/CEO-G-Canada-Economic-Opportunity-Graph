import { publicText } from "@/lib/public-api";

export const dynamic = "force-static";

export function GET() {
  const publisherId = process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID;
  if (!publisherId || !/^ca-pub-\d{16}$/.test(publisherId)) {
    return publicText("Advertising is not configured.\n", "text/plain; charset=utf-8", { status: 404 });
  }
  const account = publisherId.replace(/^ca-/, "");
  return publicText(`google.com, ${account}, DIRECT, f08c47fec0942fa0\n`, "text/plain; charset=utf-8");
}
