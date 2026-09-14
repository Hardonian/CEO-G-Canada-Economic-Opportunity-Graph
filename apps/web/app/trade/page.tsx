import type { Metadata } from "next";
import Link from "next/link";
import { ArrowUpRight, Boxes, Globe2, Network, PackageSearch, Ship, ShieldCheck } from "lucide-react";
import { VETTED_SOURCES } from "@/lib/source-data";

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
            A vetted registry of official Canadian and multilateral services for commodity flows, value-added trade, tariffs, logistics performance, ports and maritime connectivity. Registration expands the evidence map without implying that a feed already affects project scores.
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
            <p className="mt-1 text-xs text-text-muted">{canadian.length} Canadian services · {international.length} multilateral services · all publisher links checked</p>
          </div>
          <span className="font-mono text-[10px] font-bold uppercase tracking-wider text-gold">Registered · not yet joined to scores</span>
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          {sources.map((source) => (
            <article key={source.id} className="glass-card flex flex-col rounded-2xl p-5">
              <div className="flex flex-wrap items-center gap-2 font-mono text-[9px] font-bold uppercase tracking-wide">
                <span className="rounded-full border border-primary/40 bg-primary/10 px-2.5 py-1 text-aurora">Official publisher</span>
                <span className="rounded-full border border-gold/40 bg-gold/10 px-2.5 py-1 text-gold">Registered, not ingested</span>
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
              These services are registered, classified and link-checked. They are not yet active ingestion inputs and do not alter project facts, scores or opportunities. Promotion to active requires a versioned adapter, licence review, schema mapping, freshness monitoring, reconciliation tests and immutable evidence lineage.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}
