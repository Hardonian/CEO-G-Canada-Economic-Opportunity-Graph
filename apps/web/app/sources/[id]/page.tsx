import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import {
  AlertTriangle,
  ArrowLeft,
  Building2,
  CalendarClock,
  ExternalLink,
  FileCode2,
  Gauge,
  Globe2,
  ShieldCheck,
} from "lucide-react";
import { getSource } from "@/lib/source-data";

interface SourcePageProps {
  params: Promise<{ id: string }>;
}

function humanize(value: string): string {
  return value.replaceAll("_", " ").replace(/\b\w/g, (character) => character.toUpperCase());
}

function displayDate(value?: string): string {
  if (!value) return "Not yet recorded";
  return new Intl.DateTimeFormat("en-CA", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(new Date(value)) + " UTC";
}

export async function generateMetadata({ params }: SourcePageProps): Promise<Metadata> {
  const { id } = await params;
  const result = await getSource(id);
  if (result.status !== "available") return { title: "Public data source" };
  return {
    title: result.data.name,
    description: result.data.description,
  };
}

export default async function SourceDetailPage({ params }: SourcePageProps) {
  const { id } = await params;
  const result = await getSource(id);
  if (result.status === "not_found") notFound();

  if (result.status === "unavailable") {
    return (
      <div className="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
        <Link href="/sources" className="inline-flex items-center gap-2 rounded-sm text-xs font-semibold text-text-muted hover:text-aurora">
          <ArrowLeft aria-hidden="true" className="h-4 w-4" /> Back to public data
        </Link>
        <section role="alert" className="mt-6 rounded-2xl border border-gold/50 bg-gold/10 p-6">
          <div className="flex items-start gap-3">
            <AlertTriangle aria-hidden="true" className="mt-0.5 h-5 w-5 shrink-0 text-gold" />
            <div>
              <h1 className="text-lg font-bold text-text-main">Source record temporarily unavailable</h1>
              <p className="mt-1 text-sm text-text-muted">The source API could not be reached or returned an invalid record. This is distinct from a source that does not exist.</p>
            </div>
          </div>
        </section>
      </div>
    );
  }

  const source = result.data;
  const qualityEntries = Object.entries(source.quality).filter((entry): entry is [string, number] => typeof entry[1] === "number");
  const useStatement = source.integration_status === "EVIDENCE_LINKED"
    ? "This source is explicitly marked active in the registry. Activity does not by itself establish complete or current downstream coverage."
    : `This source is registered at the ${humanize(source.lifecycle)} lifecycle stage and is not an active ingestion input. It does not affect project facts or scores.`;

  return (
    <div className="mx-auto max-w-6xl space-y-7 px-4 py-8 sm:px-6 lg:px-8">
      <nav aria-label="Breadcrumb">
        <Link href="/sources" className="inline-flex items-center gap-2 rounded-sm text-xs font-semibold text-text-muted hover:text-aurora">
          <ArrowLeft aria-hidden="true" className="h-4 w-4" /> Explore Canadian Public Data
        </Link>
      </nav>

      <header className="relative overflow-hidden rounded-2xl border border-border/80 bg-gradient-to-b from-card/95 to-surface/95 p-6 shadow-2xl sm:p-8">
        <div aria-hidden="true" className="pointer-events-none absolute -right-12 -top-16 h-64 w-64 rounded-full bg-aurora/10 blur-3xl" />
        <div className="relative z-10">
          <div className="flex flex-wrap gap-2 font-mono text-[10px] uppercase tracking-wider">
            <span className="rounded-full border border-primary/40 bg-primary/10 px-2.5 py-1 font-bold text-aurora">{humanize(source.health)}</span>
            <span className="rounded-full border border-borderSubtle bg-background px-2.5 py-1 text-text-muted">{humanize(source.lifecycle)}</span>
            <span className="rounded-full border border-gold/40 bg-gold/10 px-2.5 py-1 text-gold">Authority tier {source.authority_tier}</span>
            <span className="rounded-full border border-borderSubtle bg-background px-2.5 py-1 text-text-muted">{source.integration_status === "EVIDENCE_LINKED" ? "Evidence linked" : "Registered · not ingested"}</span>
          </div>
          <p className="mt-5 font-mono text-xs uppercase tracking-wider text-aurora">{source.publisher_name}</p>
          <h1 className="mt-1 max-w-4xl text-3xl font-black tracking-tight text-text-main sm:text-5xl">{source.name}</h1>
          <p className="mt-4 max-w-4xl text-sm leading-relaxed text-text-muted">{source.description}</p>
          <div className="mt-6 flex flex-wrap gap-3">
            <a href={source.canonical_url} target="_blank" rel="noopener noreferrer" className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-primary px-4 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
              Open publisher source <ExternalLink aria-hidden="true" className="h-4 w-4" />
              <span className="sr-only"> (opens in a new tab)</span>
            </a>
            <Link href={{ pathname: "/sources", query: { publisher: source.publisher_name } }} className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora">
              More from this publisher
            </Link>
          </div>
        </div>
      </header>

      <div className="grid gap-6 lg:grid-cols-[1.25fr_0.75fr]">
        <div className="space-y-6">
          <section aria-labelledby="source-profile-title" className="glass-panel rounded-2xl p-6">
            <h2 id="source-profile-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main">
              <FileCode2 aria-hidden="true" className="h-4 w-4 text-aurora" /> Source profile
            </h2>
            <dl className="mt-5 grid gap-5 text-sm sm:grid-cols-2">
              <Detail label="Source family" value={humanize(source.source_family)} />
              <Detail label="Access method" value={humanize(source.access_method)} />
              <Detail label="Content type" value={source.content_type} />
              <Detail label="Update frequency" value={humanize(source.update_frequency)} />
              <Detail label="Jurisdiction" value={source.jurisdiction} />
              <Detail label="Geography" value={source.geography.length > 0 ? source.geography.join(", ") : "Not specified"} />
              <Detail label="Languages" value={source.languages.length > 0 ? source.languages.join(", ") : "Not specified"} />
              <Detail label="Licence" value={source.license} />
              <Detail label="Coverage class" value={humanize(source.coverage_class)} />
              <Detail label="Registry ID" value={source.id} mono />
              {source.authentication_required && <Detail label="Authentication" value="Free API key required" />}
              {source.evidence_record_count !== undefined && <Detail label="Evidence records" value={String(source.evidence_record_count)} />}
            </dl>
          </section>

          <section aria-labelledby="cog-use-title" className="legal-rule rounded-2xl border border-border bg-card/70 p-6 pl-8">
            <h2 id="cog-use-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main">
              <ShieldCheck aria-hidden="true" className="h-4 w-4 text-aurora" /> How COG represents this source
            </h2>
            <p className="mt-3 text-sm leading-relaxed text-text-muted">{useStatement}</p>
          </section>

          <section aria-labelledby="topics-title" className="glass-panel rounded-2xl p-6">
            <h2 id="topics-title" className="text-sm font-bold uppercase tracking-wider text-text-main">Subjects and sectors</h2>
            <div className="mt-4 space-y-4">
              <TagList label="Subject tags" values={source.subject_tags} />
              <TagList label="Sector tags" values={source.sector_tags} />
            </div>
          </section>
        </div>

        <aside className="space-y-6">
          <section aria-labelledby="freshness-title" className="glass-card rounded-2xl p-6">
            <h2 id="freshness-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main">
              <CalendarClock aria-hidden="true" className="h-4 w-4 text-aurora" /> Observed freshness
            </h2>
            <dl className="mt-5 space-y-4 text-sm">
              <Detail label="Last checked" value={displayDate(source.last_checked_at)} />
              <Detail label="Last successful ingest" value={displayDate(source.last_success_at)} />
              <Detail label="Last material change" value={displayDate(source.last_change_at)} />
            </dl>
          </section>

          <section aria-labelledby="quality-title" className="glass-card rounded-2xl p-6">
            <h2 id="quality-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main">
              <Gauge aria-hidden="true" className="h-4 w-4 text-gold" /> Dataset quality
            </h2>
            {qualityEntries.length > 0 ? (
              <dl className="mt-5 space-y-4">
                {qualityEntries.map(([label, value]) => (
                  <div key={label}>
                    <div className="flex items-center justify-between gap-4 text-xs">
                      <dt className="text-text-muted">{humanize(label)}</dt>
                      <dd className="font-tabular font-bold text-text-main">{value.toFixed(0)}/100</dd>
                    </div>
                    <div aria-hidden="true" className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-background">
                      <div className="h-full rounded-full bg-primary" style={{ width: `${value}%` }} />
                    </div>
                  </div>
                ))}
              </dl>
            ) : (
              <p className="mt-4 text-sm text-text-muted">Quality dimensions have not been measured for this source.</p>
            )}
            <p className="mt-5 border-t border-borderSubtle pt-4 text-[10px] leading-relaxed text-text-subtle">
              Dataset quality and publisher authority are separate measures. Neither is a guarantee that every record is complete or correct.
            </p>
          </section>

          <section aria-labelledby="publisher-title" className="glass-card rounded-2xl p-6">
            <h2 id="publisher-title" className="flex items-center gap-2 text-sm font-bold uppercase tracking-wider text-text-main">
              <Building2 aria-hidden="true" className="h-4 w-4 text-aurora" /> Publisher identity
            </h2>
            <p className="mt-4 text-sm font-bold text-text-main">{source.publisher_name}</p>
            <p className="mt-1 break-all font-mono text-[10px] text-text-subtle">{source.publisher_id}</p>
          </section>
        </aside>
      </div>
    </div>
  );
}

function Detail({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">{label}</dt>
      <dd className={`mt-1 break-words text-text-main ${mono ? "font-mono text-xs" : ""}`}>{value}</dd>
    </div>
  );
}

function TagList({ label, values }: { label: string; values: string[] }) {
  return (
    <div>
      <p className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">{label}</p>
      {values.length > 0 ? (
        <ul className="mt-2 flex list-none flex-wrap gap-2" role="list">
          {values.map((value) => <li key={value} className="rounded-full border border-borderSubtle bg-background px-2.5 py-1 text-xs text-text-muted">{value}</li>)}
        </ul>
      ) : (
        <p className="mt-1 text-sm text-text-muted">None recorded</p>
      )}
    </div>
  );
}
