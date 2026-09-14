import type { Metadata } from "next";
import Link from "next/link";
import { ArrowUpRight, BarChart3, Boxes, Globe2, Network, PackageSearch, Ship, ShieldCheck } from "lucide-react";
import { SNAPSHOT_PROJECTS } from "@/lib/data";
import { VETTED_SOURCES } from "@/lib/source-data";
import { TRADE_METRICS } from "@/lib/trade-data";

export const metadata: Metadata = {
  title: "Global Trade & Supply Chains",
  description: "Official Canadian and international trade, logistics, port, tariff, and global value-chain sources registered for CanadaOpportunityGraph.",
};

const dimensions = [
  {
    title: "Commodity trade flows",
    description: "Imports and exports by partner, province, HS commodity and period.",
    coverage: "Statistics Canada · ISED · UN Comtrade",
    icon: PackageSearch,
  },
  {
    title: "Global value chains",
    description: "Domestic and foreign value added embodied in production and final demand.",
    coverage: "OECD TiVA",
    icon: Network,
  },
  {
    title: "Tariffs and market access",
    description: "Applied and preferential tariffs, services trade, non-tariff indicators and trade policy.",
    coverage: "WTO",
    icon: ShieldCheck,
  },
  {
    title: "Logistics and port resilience",
    description: "Customs efficiency, shipment reliability, ports, maritime connectivity and disruption catalogues.",
    coverage: "World Bank · IMF PortWatch · UNCTAD",
    icon: Ship,
  },
];

export default function TradePage() {
  const sources = VETTED_SOURCES.filter((source) =>
    source.sector_tags.some((tag) => tag.toLocaleLowerCase("en-CA") === "global trade and supply chains"),
  );
  const canadian = sources.filter((source) => source.jurisdiction === "CA");
  const international = sources.filter((source) => source.jurisdiction !== "CA");
  const active = sources.filter((source) => source.integration_status === "EVIDENCE_LINKED");
  const registered = sources.filter((source) => source.integration_status !== "EVIDENCE_LINKED");
  const tradeScore = SNAPSHOT_PROJECTS[0]?.scores?.trade_resilience;
  const featuredMetrics = ["LP.LPI.OVRL.XQ", "NE.TRD.GNFS.ZS", "TX.VAL.MRCH.CD.WT", "TM.VAL.MRCH.CD.WT"]
    .map((code) => TRADE_METRICS.find((metric) => metric.metric_code === code))
    .filter((metric): metric is (typeof TRADE_METRICS)[number] => Boolean(metric));
  const formatMetric = (value: number, unit: string) => unit === "current_USD"
    ? `$${(value / 1_000_000_000).toFixed(1)}B USD`
    : unit === "percent" ? `${value.toFixed(1)}%` : `${value.toFixed(1)}/5`;

  return (
    <div className="mx-auto max-w-7xl space-y-9 px-4 py-8 sm:px-6 lg:px-8">
      <header className="relative overflow-hidden rounded-2xl border border-border/80 bg-gradient-to-b from-card/95 to-surface/95 p-6 shadow-2xl sm:p-8">
        <div aria-hidden="true" className="pointer-events-none absolute -right-16 -top-20 h-72 w-72 rounded-full bg-gold/10 blur-3xl" />
        <div className="relative z-10 max-w-4xl">
          <div className="inline-flex items-center gap-2 rounded-full border border-gold/40 bg-gold/10 px-3 py-1 font-mono text-[11px] text-gold">
            <Globe2 aria-hidden="true" className="h-3.5 w-3.5" /> GLOBAL TRADE & SUPPLY-CHAIN SOURCE LAYER
          </div>
          <h1 className="mt-3 text-3xl font-black tracking-tight text-text-main sm:text-5xl">
            Trade & <span className="text-gold">Supply Chains</span>
          </h1>
          <p className="mt-4 max-w-3xl text-sm leading-relaxed text-text-muted">
            Official Canadian and multilateral intelligence for commodity flows, value-added trade, tariffs, logistics performance, ports and maritime connectivity. World Bank observations now feed a reproducible trade-resilience score; all other services remain explicitly staged until their adapters and lineage gates pass.
          </p>
          <div className="mt-5 flex flex-wrap gap-3">
            <Link href="/api/v1/sources?q=trade" className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-primary px-4 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
              Open machine-readable registry <ArrowUpRight aria-hidden="true" className="h-4 w-4" />
            </Link>
            <Link href="/sources?sector=global+trade+and+supply+chains" className="inline-flex min-h-10 items-center rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora">
              Inspect source profiles
            </Link>
          </div>
        </div>
      </header>

      <section aria-labelledby="active-trade-intelligence-title" className="glass-card rounded-2xl border border-primary/30 p-6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <div className="flex items-center gap-2 font-mono text-[10px] font-bold uppercase tracking-wider text-aurora">
              <BarChart3 aria-hidden="true" className="h-4 w-4" /> Evidence-linked scoring input
            </div>
            <h2 id="active-trade-intelligence-title" className="mt-2 text-xl font-black text-text-main">Canada trade-resilience context</h2>
            <p className="mt-2 max-w-3xl text-xs leading-relaxed text-text-muted">
              {TRADE_METRICS.length} normalized official observations with record hashes, source locators and immutable evidence IDs. The score is national context—not a claim about a specific project&apos;s supplier exposure.
            </p>
          </div>
          <div className="rounded-xl border border-primary/40 bg-primary/10 px-5 py-3 text-right">
            <div className="font-mono text-[9px] uppercase tracking-wider text-text-muted">trade-resilience-v1.0</div>
            <div className="mt-1 text-2xl font-black tabular-nums text-aurora">{tradeScore == null ? "NOT SCORED" : `${tradeScore.toFixed(1)}/100`}</div>
          </div>
        </div>
        <div className="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {featuredMetrics.map((metric) => (
            <div key={metric.id} className="rounded-xl border border-borderSubtle bg-background/60 p-4">
              <div className="font-mono text-[9px] text-aurora">{metric.metric_code} · {metric.reference_period}</div>
              <div className="mt-2 text-xl font-black tabular-nums text-text-main">{formatMetric(metric.value, metric.unit)}</div>
              <div className="mt-1 line-clamp-2 text-[10px] leading-relaxed text-text-muted">{metric.metric_name}</div>
            </div>
          ))}
        </div>
        <Link href="/api/v1/trade/metrics" className="mt-5 inline-flex items-center gap-1 text-xs font-bold text-aurora underline decoration-border underline-offset-4">
          Open all metrics and lineage IDs <ArrowUpRight aria-hidden="true" className="h-3 w-3" />
        </Link>
      </section>

      <section aria-labelledby="trade-dimensions-title" className="space-y-4">
        <div>
          <h2 id="trade-dimensions-title" className="text-lg font-black text-text-main">Coverage architecture</h2>
          <p className="mt-1 text-xs text-text-muted">Four complementary lenses connect Canadian projects to external demand, input dependencies and transport risk.</p>
        </div>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {dimensions.map(({ title, description, coverage, icon: Icon }) => (
            <article key={title} className="glass-card rounded-2xl p-5">
              <Icon aria-hidden="true" className="h-5 w-5 text-gold" />
              <h3 className="mt-3 font-bold text-text-main">{title}</h3>
              <p className="mt-2 text-xs leading-relaxed text-text-muted">{description}</p>
              <p className="mt-4 border-t border-borderSubtle pt-3 font-mono text-[10px] text-aurora">{coverage}</p>
            </article>
          ))}
        </div>
      </section>

      <section aria-labelledby="trade-sources-title" className="space-y-5">
        <div className="flex flex-col gap-2 border-b border-borderSubtle pb-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 id="trade-sources-title" className="text-lg font-black text-text-main">Registered official sources</h2>
            <p className="mt-1 text-xs text-text-muted">{canadian.length} Canadian services · {international.length} multilateral services · {active.length} evidence-linked · {registered.length} staged</p>
          </div>
          <span className="font-mono text-[10px] font-bold uppercase tracking-wider text-gold">Integration status shown per source</span>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          {sources.map((source) => (
            <article key={source.id} className="glass-card flex flex-col rounded-2xl p-5">
              <div className="flex flex-wrap items-center gap-2 font-mono text-[9px] font-bold uppercase tracking-wide">
                <span className="rounded-full border border-primary/40 bg-primary/10 px-2.5 py-1 text-aurora">Official publisher</span>
                <span className={`rounded-full border px-2.5 py-1 ${source.integration_status === "EVIDENCE_LINKED" ? "border-primary/40 bg-primary/10 text-aurora" : "border-gold/40 bg-gold/10 text-gold"}`}>
                  {source.integration_status === "EVIDENCE_LINKED" ? `${source.evidence_record_count ?? 0} evidence-linked records` : "Registered, not ingested"}
                </span>
                {source.authentication_required && <span className="rounded-full border border-borderSubtle bg-background px-2.5 py-1 text-text-muted">Free API key required</span>}
              </div>
              <p className="mt-4 font-mono text-[10px] uppercase tracking-wider text-aurora">{source.publisher_name}</p>
              <h3 className="mt-1 text-lg font-black text-text-main">{source.name}</h3>
              <p className="mt-2 flex-1 text-xs leading-relaxed text-text-muted">{source.description}</p>
              <div className="mt-5 flex flex-wrap gap-3 border-t border-borderSubtle pt-4 text-xs">
                <Link href={`/sources/${encodeURIComponent(source.id)}`} className="font-bold text-text-main underline decoration-border underline-offset-4 hover:text-aurora">Source profile</Link>
                <a href={source.canonical_url} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-1 font-bold text-aurora underline decoration-border underline-offset-4">
                  Open official service <ArrowUpRight aria-hidden="true" className="h-3 w-3" />
                </a>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="legal-rule rounded-2xl border border-border bg-card/70 p-6 pl-8" aria-labelledby="integration-boundary-title">
        <div className="flex items-start gap-3">
          <Boxes aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-aurora" />
          <div>
            <h2 id="integration-boundary-title" className="font-bold text-text-main">Integration boundary</h2>
            <p className="mt-2 text-sm leading-relaxed text-text-muted">
              World Bank Indicators is active through a versioned adapter and contributes only to the disclosed trade-resilience and supplierability factors. Every other service is registered, classified and link-checked but does not yet alter scores. Promotion requires licence review, schema mapping, freshness monitoring, reconciliation tests and immutable evidence lineage.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
