import type { Metadata } from "next";
import Link from "next/link";
import {
  AlertTriangle,
  ArrowUpRight,
  Database,
  Gauge,
  MapPinned,
  Network,
  Radio,
  ShieldAlert,
} from "lucide-react";
import { getSourceCoverage } from "@/lib/source-data";

export const metadata: Metadata = {
  title: "Public Data Mesh Status",
  description: "Read-only measured coverage and operational status for the CanadaOpportunityGraph public-source census.",
};

function humanize(value: string): string {
  return value.replaceAll("_", " ").replace(/\b\w/g, (character) => character.toUpperCase());
}

function formatCount(value: number): string {
  return new Intl.NumberFormat("en-CA").format(value);
}

function Breakdown({ title, values }: { title: string; values: Record<string, number> }) {
  const entries = Object.entries(values).sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]));
  return (
    <section className="glass-card rounded-2xl p-5">
      <h2 className="text-xs font-bold uppercase tracking-wider text-text-main">{title}</h2>
      {entries.length > 0 ? (
        <dl className="mt-4 max-h-80 space-y-2 overflow-y-auto pr-1 text-xs">
          {entries.map(([label, count]) => (
            <div key={label} className="flex items-center justify-between gap-4 border-b border-borderSubtle pb-2 last:border-0">
              <dt className="text-text-muted">{humanize(label)}</dt>
              <dd className="font-tabular font-bold text-text-main">{formatCount(count)}</dd>
            </div>
          ))}
        </dl>
      ) : (
        <p className="mt-4 text-sm text-text-muted">No measured breakdown is available.</p>
      )}
    </section>
  );
}

export default async function AdminPage() {
  const coverageResult = await getSourceCoverage();

  if (coverageResult.status !== "available") {
    return (
      <div className="mx-auto max-w-5xl space-y-6 px-4 py-10 sm:px-6 lg:px-8">
        <header>
          <div className="inline-flex items-center gap-2 rounded-full border border-gold/50 bg-gold/10 px-3 py-1 font-mono text-[11px] text-gold">
            <Radio aria-hidden="true" className="h-3.5 w-3.5" /> READ-ONLY DATA MESH STATUS
          </div>
          <h1 className="mt-3 text-3xl font-black tracking-tight text-text-main sm:text-5xl">Public Data Mesh <span className="text-aurora">Status</span></h1>
        </header>
        <section role="alert" className="rounded-2xl border border-gold/50 bg-gold/10 p-6">
          <div className="flex items-start gap-3">
            <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-gold" />
            <div>
              <h2 className="font-bold text-text-main">Coverage report unavailable</h2>
              <p className="mt-1 text-sm text-text-muted">The public coverage endpoint did not return a valid measured report. Source counts, health, latency, and dead-letter status are intentionally not estimated.</p>
              <Link href="/sources" className="mt-4 inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora">Browse source directory <ArrowUpRight aria-hidden="true" className="h-4 w-4" /></Link>
            </div>
          </div>
        </section>
      </div>
    );
  }

  const coverage = coverageResult.data;
  const ratio = coverage.primary_source_ratio.value;
  const primaryRatio = ratio === undefined
    ? humanize(coverage.primary_source_ratio.status)
    : new Intl.NumberFormat("en-CA", { style: "percent", maximumFractionDigits: 1 }).format(ratio);
  const latency = coverage.signal_latency.p50_ms === undefined
    ? humanize(coverage.signal_latency.status)
    : `${formatCount(coverage.signal_latency.p50_ms)} ms p50`;
  const deadLetters = coverage.dead_letters.count === undefined
    ? humanize(coverage.dead_letters.status)
    : formatCount(coverage.dead_letters.count);

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <header className="flex flex-col gap-5 border-b border-border pb-6 lg:flex-row lg:items-end lg:justify-between">
        <div className="max-w-4xl">
          <div className="inline-flex items-center gap-2 rounded-full border border-primary/40 bg-primary/10 px-3 py-1 font-mono text-[11px] text-aurora">
            <Radio aria-hidden="true" className="h-3.5 w-3.5" /> READ-ONLY · MEASURED COVERAGE
          </div>
          <h1 className="mt-3 text-3xl font-black tracking-tight text-text-main sm:text-5xl">Public Data Mesh <span className="text-aurora">Status</span></h1>
          <p className="mt-3 max-w-3xl text-sm leading-relaxed text-text-muted">A public-safe snapshot of source lifecycle, coverage, freshness instrumentation, and known blind spots. This page does not expose credentials, parser configuration, raw errors, or operator controls.</p>
        </div>
        <div className="shrink-0 text-left lg:text-right">
          <p className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Report generated</p>
          <p className="mt-1 text-xs text-text-main">{new Intl.DateTimeFormat("en-CA", { dateStyle: "medium", timeStyle: "short", timeZone: "UTC" }).format(new Date(coverage.generated_at))} UTC</p>
        </div>
      </header>

      <section aria-labelledby="lifecycle-title" className="space-y-3">
        <div className="flex flex-col gap-1 sm:flex-row sm:items-end sm:justify-between">
          <h2 id="lifecycle-title" className="text-sm font-bold uppercase tracking-wider text-text-main">Source lifecycle census</h2>
          <p className="font-mono text-[10px] uppercase tracking-wider text-text-subtle">Counts remain intentionally distinct</p>
        </div>
        <dl className="grid grid-cols-2 gap-3 sm:grid-cols-5">
          {([
            ["Discovered", coverage.lifecycle_counts.discovered, Network],
            ["Registered", coverage.lifecycle_counts.registered, Database],
            ["Tested", coverage.lifecycle_counts.tested, Gauge],
            ["Active", coverage.lifecycle_counts.active, Radio],
            ["Broken", coverage.lifecycle_counts.broken, ShieldAlert],
          ] as const).map(([label, value, Icon]) => (
            <div key={label} className="glass-card rounded-xl p-4">
              <dt className="flex items-center gap-1.5 font-mono text-[9px] uppercase tracking-wider text-text-subtle"><Icon aria-hidden="true" className="h-3.5 w-3.5 text-aurora" /> {label}</dt>
              <dd className="mt-2 font-tabular text-2xl font-black text-text-main">{formatCount(value)}</dd>
            </div>
          ))}
        </dl>
        <p className="text-[10px] leading-relaxed text-text-subtle">Discovered candidates, persisted registrations, completed access tests, approved active sources, and currently broken active sources are different populations.</p>
      </section>

      <section aria-labelledby="operations-title" className="space-y-3">
        <h2 id="operations-title" className="text-sm font-bold uppercase tracking-wider text-text-main">Measured operations</h2>
        <dl className="grid gap-3 sm:grid-cols-3">
          <div className="glass-card rounded-xl p-5"><dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Primary-source ratio</dt><dd className="mt-2 text-xl font-black text-text-main">{primaryRatio}</dd><p className="mt-1 text-[10px] text-text-subtle">Status: {humanize(coverage.primary_source_ratio.status)}</p></div>
          <div className="glass-card rounded-xl p-5"><dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Signal latency</dt><dd className="mt-2 text-xl font-black text-text-main">{latency}</dd><p className="mt-1 text-[10px] text-text-subtle">Status: {humanize(coverage.signal_latency.status)}{coverage.signal_latency.sample_count !== undefined ? ` · ${formatCount(coverage.signal_latency.sample_count)} samples` : ""}</p></div>
          <div className="glass-card rounded-xl p-5"><dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Dead letters</dt><dd className="mt-2 text-xl font-black text-text-main">{deadLetters}</dd><p className="mt-1 text-[10px] text-text-subtle">Status: {humanize(coverage.dead_letters.status)}</p></div>
        </dl>
      </section>

      <section aria-labelledby="coverage-title" className="space-y-3">
        <h2 id="coverage-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main"><MapPinned aria-hidden="true" className="h-4 w-4 text-aurora" /> Registered-source coverage</h2>
        <div className="grid gap-4 lg:grid-cols-3">
          <Breakdown title="By jurisdiction" values={coverage.by_jurisdiction} />
          <Breakdown title="By sector" values={coverage.by_sector} />
          <Breakdown title="By source family" values={coverage.by_family} />
        </div>
      </section>

      <section aria-labelledby="blind-spots-title" className="rounded-2xl border border-gold/50 bg-gold/10 p-6">
        <h2 id="blind-spots-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main"><AlertTriangle aria-hidden="true" className="h-4 w-4 text-gold" /> Declared blind spots</h2>
        {coverage.blind_spots.length > 0 ? (
          <ul className="mt-4 list-disc space-y-2 pl-5 text-sm text-text-muted">
            {coverage.blind_spots.map((blindSpot) => <li key={blindSpot}>{blindSpot}</li>)}
          </ul>
        ) : (
          <p className="mt-3 text-sm text-text-muted">No blind spots were declared in this report. This does not establish comprehensive coverage.</p>
        )}
      </section>

      <div className="flex flex-wrap gap-3 border-t border-borderSubtle pt-5">
        <Link href="/sources" className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-primary px-4 text-xs font-black text-[#050b08] hover:bg-aurora-mint">Explore public sources <ArrowUpRight aria-hidden="true" className="h-4 w-4" /></Link>
        <Link href="/apis" className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora">Public API directory</Link>
      </div>
    </div>
  );
}
