import { publicJSON, publicOptions } from "@/lib/public-api";
import { TRADE_METRICS } from "@/lib/trade-data";

export const dynamic = "force-static";

export async function GET() {
  return publicJSON({
    geography: "CAN",
    source_mode: "BUNDLED_REVIEWED_SNAPSHOT",
    metric_count: TRADE_METRICS.length,
    metrics: TRADE_METRICS,
  });
}

export async function OPTIONS() {
  return publicOptions();
}
