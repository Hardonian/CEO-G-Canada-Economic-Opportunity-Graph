import Link from "next/link";
import { ArrowLeft, ArrowRight, ExternalLink, FileSearch, ShieldCheck } from "lucide-react";
import type { PublicSource } from "@/lib/types";

interface SourceResultsProps {
  sources: PublicSource[];
  total: number;
  limit: number;
  offset: number;
  pathname: string;
  query: Record<string, string | undefined>;
  emptyTitle: string;
  emptyBody: string;
}

function humanize(value: string): string {
  return value.replaceAll("_", " ").replace(/\b\w/g, (character) => character.toUpperCase());
}

function healthClass(health: PublicSource["health"]): string {
  if (health === "CURRENT" || health === "HEALTHY") {
    return "border-primary/40 bg-primary/10 text-aurora";
  }
  if (health === "DELAYED" || health === "STALE" || health === "DEGRADED") {
    return "border-gold/50 bg-gold/10 text-gold";
  }
  return "border-crimson/50 bg-crimson/10 text-crimson-light";
}

function pageHref(
  pathname: string,
  query: Record<string, string | undefined>,
  offset: number,
): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value) params.set(key, value);
  }
  if (offset > 0) params.set("offset", String(offset));
  const search = params.toString();
  return search ? `${pathname}?${search}` : pathname;
}

export default function SourceResults({
  sources,
  total,
  limit,
  offset,
  pathname,
  query,
  emptyTitle,
  emptyBody,
}: SourceResultsProps) {
  if (sources.length === 0) {
    return (
      <section className="glass-panel rounded-2xl p-8 text-center" aria-live="polite">
        <FileSearch aria-hidden="true" className="mx-auto h-8 w-8 text-text-subtle" />
        <h2 className="mt-3 text-lg font-bold text-text-main">{emptyTitle}</h2>
        <p className="mx-auto mt-2 max-w-xl text-sm text-text-muted">{emptyBody}</p>
      </section>
    );
  }

  const first = offset + 1;
  const last = Math.min(offset + sources.length, total);
  const hasPrevious = offset > 0;
  const hasNext = offset + sources.length < total;

  return (
    <section aria-labelledby="source-results-title" className="space-y-4">
      <div className="flex flex-col gap-2 border-b border-borderSubtle pb-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h2 id="source-results-title" className="text-lg font-black text-text-main">
            Source directory
          </h2>
          <p className="mt-1 text-xs text-text-muted" aria-live="polite">
            Showing <span className="font-tabular text-text-main">{first}–{last}</span> of{" "}
            <span className="font-tabular text-text-main">{total}</span> registered results.
          </p>
        </div>
        <p className="font-mono text-[10px] uppercase tracking-wider text-text-subtle">
          Registration does not imply active ingestion
        </p>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        {sources.map((source) => (
          <article key={source.id} className="glass-card flex h-full flex-col rounded-2xl p-5">
            <div className="flex flex-wrap items-center gap-2 font-mono text-[10px] uppercase tracking-wide">
              <span className={`rounded-full border px-2.5 py-1 font-bold ${healthClass(source.health)}`}>
                {humanize(source.health)}
              </span>
              <span className="rounded-full border border-borderSubtle bg-surface px-2.5 py-1 text-text-muted">
                {humanize(source.lifecycle)}
              </span>
              <span className="rounded-full border border-gold/40 bg-gold/10 px-2.5 py-1 text-gold">
                Authority tier {source.authority_tier}
              </span>
            </div>

            <div className="mt-4 flex-1">
              <p className="font-mono text-[10px] uppercase tracking-wider text-aurora">
                {source.publisher_name}
              </p>
              <h3 className="mt-1 text-lg font-black leading-snug text-text-main">
                <Link href={`/sources/${encodeURIComponent(source.id)}`} className="rounded-sm hover:text-aurora hover:underline">
                  {source.name}
                </Link>
              </h3>
              <p className="mt-2 line-clamp-3 text-xs leading-relaxed text-text-muted">{source.description}</p>
            </div>

            <dl className="mt-5 grid grid-cols-2 gap-x-4 gap-y-3 border-y border-borderSubtle py-4 text-xs">
              <div>
                <dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Jurisdiction</dt>
                <dd className="mt-0.5 text-text-main">{source.jurisdiction}</dd>
              </div>
              <div>
                <dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Family</dt>
                <dd className="mt-0.5 text-text-main">{humanize(source.source_family)}</dd>
              </div>
              <div>
                <dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Format</dt>
                <dd className="mt-0.5 text-text-main">{source.content_type}</dd>
              </div>
              <div>
                <dt className="font-mono text-[9px] uppercase tracking-wider text-text-subtle">Update pattern</dt>
                <dd className="mt-0.5 text-text-main">{humanize(source.update_frequency)}</dd>
              </div>
            </dl>

            <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-xs">
              <span className="inline-flex items-center gap-1.5 text-text-muted">
                <ShieldCheck aria-hidden="true" className="h-3.5 w-3.5 text-aurora" />
                Coverage: {humanize(source.coverage_class)}
              </span>
              <a
                href={source.canonical_url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center gap-1 rounded-sm font-semibold text-aurora underline decoration-border underline-offset-4 hover:decoration-aurora"
              >
                Publisher source
                <ExternalLink aria-hidden="true" className="h-3 w-3" />
                <span className="sr-only"> (opens in a new tab)</span>
              </a>
            </div>
          </article>
        ))}
      </div>

      {(hasPrevious || hasNext) && (
        <nav aria-label="Source directory pagination" className="flex items-center justify-between border-t border-borderSubtle pt-4">
          {hasPrevious ? (
            <Link
              href={pageHref(pathname, query, Math.max(0, offset - limit))}
              className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora"
            >
              <ArrowLeft aria-hidden="true" className="h-4 w-4" /> Previous
            </Link>
          ) : (
            <span />
          )}
          {hasNext && (
            <Link
              href={pageHref(pathname, query, offset + limit)}
              className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-primary/60 bg-primary/10 px-4 text-xs font-bold text-aurora hover:bg-primary/20"
            >
              Next <ArrowRight aria-hidden="true" className="h-4 w-4" />
            </Link>
          )}
        </nav>
      )}
    </section>
  );
}
