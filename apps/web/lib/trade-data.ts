import tradeMetricsSnapshot from "@/data/trade-metrics.snapshot.json";
import type { TradeMetric } from "./types";

function normalizeMetric(value: unknown): TradeMetric | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) return null;
  const item = value as Record<string, unknown>;
  if (
    typeof item.id !== "string" ||
    item.geography !== "CAN" ||
    typeof item.metric_code !== "string" ||
    typeof item.metric_name !== "string" ||
    typeof item.reference_period !== "string" ||
    !/^\d{4}$/.test(item.reference_period) ||
    typeof item.value !== "number" ||
    !Number.isFinite(item.value) ||
    typeof item.unit !== "string" ||
    typeof item.evidence_id !== "string" ||
    typeof item.observed_at !== "string" ||
    Number.isNaN(Date.parse(item.observed_at))
  ) return null;
  return {
    id: item.id,
    geography: item.geography,
    metric_code: item.metric_code,
    metric_name: item.metric_name,
    reference_period: item.reference_period,
    value: item.value,
    unit: item.unit,
    scale_min: typeof item.scale_min === "number" ? item.scale_min : undefined,
    scale_max: typeof item.scale_max === "number" ? item.scale_max : undefined,
    evidence_id: item.evidence_id,
    observed_at: item.observed_at,
  };
}

const normalized = (tradeMetricsSnapshot as unknown[]).flatMap((candidate) => {
  const metric = normalizeMetric(candidate);
  return metric ? [metric] : [];
});

if (normalized.length !== tradeMetricsSnapshot.length) {
  throw new Error("Bundled trade metric snapshot failed validation");
}

export const TRADE_METRICS: TradeMetric[] = normalized;
