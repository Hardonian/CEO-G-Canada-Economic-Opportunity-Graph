import type { Metadata } from "next";
import Link from "next/link";
import { AlertTriangle, ArrowLeft, Braces, Filter, Search } from "lucide-react";
import SourceResults from "@/components/SourceResults";
import { getSources, type SourceQuery } from "@/lib/source-data";

export const metadata: Metadata = {
  title: "Canadian Public API Directory",
  description: "Discover registered public APIs and machine-readable services relevant to Canadian economic development.",
};

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function firstValue(value: string | string[] | undefined, maxLength = 200): string | undefined {
  const candidate = Array.isArray(value) ? value[0] : value;
  if (!candidate) return undefined;
  const normalized = candidate.trim();
  return normalized && normalized.length <= maxLength ? normalized : undefined;
}

function offsetValue(value: string | string[] | undefined): number {
  const raw = firstValue(value, 12);
  if (!raw || !/^\d+$/.test(raw)) return 0;
  const parsed = Number(raw);
  return Number.isSafeInteger(parsed) && parsed >= 0 && parsed <= 1_000_000 ? parsed : 0;
}

export default async function APIDirectoryPage({ searchParams }: { searchParams: SearchParams }) {
  const rawParams = await searchParams;
  const query: SourceQuery = {
    family: "api",
    q: firstValue(rawParams.q),
    jurisdiction: firstValue(rawParams.jurisdiction),
    format: firstValue(rawParams.format),
    authority: firstValue(rawParams.authority),
    limit: 24,
    offset: offsetValue(rawParams.offset),
  };
  const queryForLinks = {
    q: query.q,
    jurisdiction: query.jurisdiction,
    format: query.format,
    authority: query.authority,
  };
  const sourceResult = await getSources(query);

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <header className="relative overflow-hidden rounded-2xl border border-border/80 bg-gradient-to-b from-card/90 to-surface/95 p-6 shadow-2xl sm:p-8">
        <div aria-hidden="true" className="pointer-events-none absolute -right-12 -top-16 h-64 w-64 rounded-full bg-gold/10 blur-3xl" />
        <div className="relative z-10 max-w-4xl">
          <Link href="/sources" className="inline-flex items-center gap-2 rounded-sm text-xs font-semibold text-text-muted hover:text-aurora">
            <ArrowLeft aria-hidden="true" className="h-4 w-4" /> All public data sources
          </Link>
          <div className="mt-5 inline-flex items-center gap-2 rounded-full border border-gold/40 bg-gold/10 px-3 py-1 font-mono text-[11px] text-gold">
            <Braces aria-hidden="true" className="h-3.5 w-3.5" /> UPSTREAM MACHINE-READABLE SERVICES
          </div>
          <h1 className="mt-3 text-3xl font-black tracking-tight text-text-main sm:text-5xl">
            Canadian Public <span className="text-gold">API Directory</span>
          </h1>
          <p className="mt-3 max-w-3xl text-sm leading-relaxed text-text-muted">
            Browse public REST, GraphQL, OpenAPI-described, CKAN, Socrata, ArcGIS, and other API-family sources. Entries point to publisher-controlled services; listing does not guarantee uptime, unrestricted use, or active ingestion by CanadaOpportunityGraph.
          </p>
        </div>
      </header>

      <section aria-labelledby="api-filter-title" className="glass-panel rounded-2xl p-5">
        <div className="mb-4 flex items-center gap-2">
          <Filter aria-hidden="true" className="h-4 w-4 text-gold" />
          <h2 id="api-filter-title" className="text-sm font-bold text-text-main">Filter public APIs</h2>
        </div>
        <form action="/apis" method="get" className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <label className="space-y-1.5 sm:col-span-2">
            <span className="text-xs font-semibold text-text-muted">Search</span>
            <span className="relative block">
              <Search aria-hidden="true" className="absolute left-3 top-3 h-4 w-4 text-text-subtle" />
              <input name="q" defaultValue={query.q} maxLength={200} placeholder="API, dataset, publisher…" className="w-full rounded-lg border border-border bg-background py-2.5 pl-9 pr-3 text-sm text-text-main placeholder:text-text-subtle" />
            </span>
          </label>
          <FilterInput name="jurisdiction" label="Jurisdiction" value={query.jurisdiction} placeholder="CA or CA:BC" />
          <FilterInput name="format" label="Format" value={query.format} placeholder="JSON, XML, GeoJSON…" />
          <label className="space-y-1.5">
            <span className="text-xs font-semibold text-text-muted">Authority tier</span>
            <select name="authority" defaultValue={query.authority ?? ""} className="w-full rounded-lg border border-border bg-background px-3 py-2.5 text-sm text-text-main">
              <option value="">Any authority</option>
              <option value="1">Tier 1 · public authority</option>
              <option value="2">Tier 2 · primary issuer</option>
              <option value="3">Tier 3</option>
              <option value="4">Tier 4</option>
              <option value="5">Tier 5 · unverified lead</option>
            </select>
          </label>
          <div className="flex items-end gap-2 sm:col-span-2 lg:col-span-3">
            <button type="submit" className="inline-flex min-h-10 items-center rounded-lg bg-primary px-5 text-xs font-black text-[#050b08] hover:bg-aurora-mint">Apply filters</button>
            <Link href="/apis" className="inline-flex min-h-10 items-center rounded-lg border border-border bg-surface px-4 text-xs font-semibold text-text-muted hover:border-primary hover:text-text-main">Clear</Link>
          </div>
        </form>
      </section>

      {sourceResult.status === "available" ? (
        <SourceResults
          sources={sourceResult.data.sources}
          total={sourceResult.data.total}
          limit={sourceResult.data.limit}
          offset={sourceResult.data.offset}
          pathname="/apis"
          query={queryForLinks}
          emptyTitle="No registered public APIs match these filters"
          emptyBody="Try broadening a filter. The directory does not add placeholder APIs when the registry returns no matches."
        />
      ) : (
        <section role="alert" className="rounded-2xl border border-crimson/50 bg-crimson/10 p-6">
          <div className="flex items-start gap-3">
            <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-crimson-light" />
            <div>
              <h2 className="font-bold text-text-main">API directory unavailable</h2>
              <p className="mt-1 text-sm text-text-muted">The source registry did not return a valid API-family response. No fallback API listings are being displayed.</p>
            </div>
          </div>
        </section>
      )}
    </div>
  );
}

function FilterInput({ name, label, value, placeholder }: { name: string; label: string; value?: string; placeholder: string }) {
  return (
    <label className="space-y-1.5">
      <span className="text-xs font-semibold text-text-muted">{label}</span>
      <input name={name} defaultValue={value} maxLength={200} placeholder={placeholder} className="w-full rounded-lg border border-border bg-background px-3 py-2.5 text-sm text-text-main placeholder:text-text-subtle" />
    </label>
  );
}
