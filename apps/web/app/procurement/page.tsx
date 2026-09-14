import Link from "next/link";
import { ArrowUpRight, DatabaseZap, FileSearch, ShieldAlert } from "lucide-react";

export default function ProcurementPage() {
  return (
    <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <header className="border-b border-border/80 pb-6">
        <div className="mb-3 inline-flex items-center gap-1.5 rounded-full border border-gold/40 bg-card px-3 py-1 font-mono text-[11px] text-gold">
          <DatabaseZap className="h-3.5 w-3.5" />
          CONNECTOR STATUS
        </div>
        <h1 className="text-2xl font-black tracking-tight text-text-main sm:text-4xl">
          Canadian <span className="text-gold">Procurement Radar</span>
        </h1>
        <p className="mt-2 max-w-3xl text-sm leading-relaxed text-text-muted">
          The procurement model and API contract are present, but this release does not ship an active CanadaBuys or defence-procurement feed.
        </p>
      </header>

      <section className="glass-card rounded-2xl border border-gold/30 p-6 shadow-xl sm:p-8">
        <div className="flex flex-col gap-5 sm:flex-row sm:items-start">
          <div className="rounded-xl border border-gold/30 bg-gold/10 p-3 text-gold">
            <ShieldAlert className="h-6 w-6" />
          </div>
          <div className="space-y-3">
            <h2 className="text-lg font-bold text-text-main">No procurement records published</h2>
            <p className="max-w-3xl text-sm leading-relaxed text-text-muted">
              This empty state is intentional. The application will not fabricate tender identifiers, closing dates, buyers, values, or verification claims. Add an attributed connector and release-tested records before this page reports active solicitations.
            </p>
            <div className="flex flex-wrap gap-3 pt-2">
              <Link
                href="/sources"
                className="inline-flex items-center gap-1.5 rounded-xl border border-primary/40 bg-primary/10 px-4 py-2 text-xs font-bold text-aurora"
              >
                Inspect source registry <FileSearch className="h-3.5 w-3.5" />
              </Link>
              <Link
                href="/cegs/adopt"
                className="inline-flex items-center gap-1.5 rounded-xl border border-border bg-card px-4 py-2 text-xs font-medium text-text-main"
              >
                Implement a CEGS feed <ArrowUpRight className="h-3.5 w-3.5" />
              </Link>
            </div>
          </div>
        </div>
      </section>

      <p className="rounded-xl border border-borderSubtle bg-surface p-4 text-xs leading-relaxed text-text-subtle">
        For a material procurement decision, consult the issuing authority&apos;s official notice and amendments directly. A future connector must preserve source URLs, retrieval times, evidence hashes, and notice status before records appear here.
      </p>
    </div>
  );
}
