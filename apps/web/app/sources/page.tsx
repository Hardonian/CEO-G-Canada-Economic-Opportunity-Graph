import type { Metadata } from "next";
import Link from "next/link";
import { AlertTriangle, Database, Filter, Gauge, Network, Search } from "lucide-react";
import SourceResults from "@/components/SourceResults";
import { getSourceCoverage, getSources, type SourceQuery } from "@/lib/source-data";

export const metadata: Metadata = {
  title: "Explore Canadian Public Data",
  description: "Browse the measured public-source census behind CanadaOpportunityGraph by publisher, jurisdiction, sector, format, frequency, and authority.",
};

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

const FILTER_KEYS = [
  "q",
  "publisher",
  "jurisdiction",
  "sector",
  "format",
  "update_frequency",
  "authority",
] as const;

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

function metricValue(value: number): string {
  return new Intl.NumberFormat("en-CA").format(value);
}

export default async function SourcesPage({ searchParams }: { searchParams: SearchParams }) {
  const rawParams = await searchParams;
  const query: SourceQuery = { limit: 24, offset: offsetValue(rawParams.offset) };
  const queryForLinks: Record<string, string | undefined> = {};
  for (const key of FILTER_KEYS) {
    const value = firstValue(rawParams[key]);
    if (value) {
      query[key] = value;
      queryForLinks[key] = value;
    }
  }

  const [sourceResult, coverageResult] = await Promise.all([
    getSources(query),
    getSourceCoverage(),
  ]);

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <header className="relative overflow-hidden rounded-2xl border border-border/80 bg-gradient-to-b from-card/90 to-surface/95 p-6 shadow-2xl sm:p-8">
        <div aria-hidden="true" className="pointer-events-none absolute -right-16 -top-20 h-72 w-72 rounded-full bg-aurora/10 blur-3xl" />
        <div className="relative z-10 max-w-4xl">
          <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-primary/40 bg-background px-3 py-1 font-mono text-[11px] text-aurora">
            <Network aria-hidden="true" className="h-3.5 w-3.5" />
            PUBLIC DATA MESH · MEASURED SOURCE CENSUS
          </div>
          <h1 className="text-3xl font-black tracking-tight text-text-main sm:text-5xl">
            Explore Canadian <span className="text-aurora">Public Data</span>
          </h1>
          <p className="mt-3 max-w-3xl text-sm leading-relaxed text-text-muted">
            Find public datasets, feeds, repositories, maps, and APIs relevant to Canadian economic development. Lifecycle and health are reported separately so a discovered source is never mistaken for an active one.
          </p>
          <div className="mt-5 flex flex-wrap gap-3">
            <Link href="/apis" className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-primary/60 bg-primary/10 px-4 text-xs font-bold text-aurora hover:bg-primary/20">
              <Database aria-hidden="true" className="h-4 w-4" /> Public API directory
            </Link>
            <Link href="/admin" className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora">
              <Gauge aria-hidden="true" className="h-4 w-4" /> Data mesh status
            </Link>
          </div>
        </div>
      </header>

      {coverageResult.status === "available" ? (
        <section aria-labelledby="coverage-snapshot-title" className="space-y-3">
          <div className="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
            <h2 id="coverage-snapshot-title" className="text-sm font-bold uppercase tracking-wider text-text-main">
              Coverage snapshot
            </h2>
            <p className="font-mono text-[10px] text-text-subtle">
              Generated {new Intl.DateTimeFormat("en-CA", { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" }).format(new Date(coverageResult.data.generated_at))} UTC
            </p>
          </div>
          <dl className="grid grid-cols-2 gap-3 sm:grid-cols-5">
            {([
              ["Discovered", coverageResult.data.lifecycle_counts.discovered],
              ["Registered", coverageResult.data.lifecycle_counts.registered],
              ["Tested", coverageResult.data.lifecycle_counts.tested],
              ["Active", coverageResult.data.lifecycle_counts.active],
              ["Broken", coverageResult.data.lifecycle_counts.broken],
            ] as const).map(([label, value]) => (
              <div key={label} className="glass-card rounded-xl p-4">
                <dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">{label}</dt>
                <dd className="mt-1 font-tabular text-2xl font-black text-text-main">{metricValue(value)}</dd>
              </div>
            ))}
          </dl>
        </section>
      ) : (
        <div role="status" className="flex items-start gap-3 rounded-xl border border-gold/50 bg-gold/10 p-4 text-sm text-text-muted">
          <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-gold" />
          Coverage metrics are temporarily unavailable. No source counts have been substituted.
        </div>
      )}

      <section aria-labelledby="source-filter-title" className="glass-panel rounded-2xl p-5">
        <div className="mb-4 flex items-center gap-2">
          <Filter aria-hidden="true" className="h-4 w-4 text-aurora" />
          <h2 id="source-filter-title" className="text-sm font-bold text-text-main">Filter the directory</h2>
        </div>
        <form action="/sources" method="get" className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <label className="space-y-1.5 lg:col-span-2">
            <span className="text-xs font-semibold text-text-muted">Search</span>
            <span className="relative block">
              <Search aria-hidden="true" className="absolute left-3 top-3 h-4 w-4 text-text-subtle" />
              <input name="q" defaultValue={query.q} maxLength={200} placeholder="Dataset, publisher, subject…" className="w-full rounded-lg border border-border bg-background py-2.5 pl-9 pr-3 text-sm text-text-main placeholder:text-text-subtle" />
            </span>
          </label>
          <FilterInput name="publisher" label="Publisher" value={query.publisher} placeholder="Statistics Canada" />
          <FilterInput name="jurisdiction" label="Jurisdiction" value={query.jurisdiction} placeholder="CA:ON" />
          <FilterInput name="sector" label="Sector" value={query.sector} placeholder="Clean Energy & Grid" />
          <FilterInput name="format" label="Format" value={query.format} placeholder="CSV, JSON, SDMX…" />
          <FilterInput name="update_frequency" label="Update frequency" value={query.update_frequency} placeholder="Daily, monthly…" />
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
          <div className="flex items-end gap-2 lg:col-span-4">
            <button type="submit" className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-primary px-5 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
              Apply filters
            </button>
            <Link href="/sources" className="inline-flex min-h-10 items-center rounded-lg border border-border bg-surface px-4 text-xs font-semibold text-text-muted hover:border-primary hover:text-text-main">
              Clear
            </Link>
          </div>
        </form>
      </section>

      {sourceResult.status === "available" ? (
        <SourceResults
          sources={sourceResult.data.sources}
          total={sourceResult.data.total}
          limit={sourceResult.data.limit}
          offset={sourceResult.data.offset}
          pathname="/sources"
          query={queryForLinks}
          emptyTitle="No registered sources match these filters"
          emptyBody="Try broadening a filter. This result is an actual empty API response; it is not an offline fallback."
        />
      ) : (
        <section role="alert" className="rounded-2xl border border-crimson/50 bg-crimson/10 p-6">
          <div className="flex items-start gap-3">
            <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-crimson-light" />
            <div>
              <h2 className="font-bold text-text-main">Source directory unavailable</h2>
              <p className="mt-1 text-sm text-text-muted">The public source API did not return a valid response. No cached or fabricated source records are being shown.</p>
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
