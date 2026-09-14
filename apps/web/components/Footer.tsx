import Link from "next/link";
import { Database, ShieldCheck } from "lucide-react";

const footerLinkClass =
  "rounded-sm text-text-muted underline decoration-border underline-offset-4 transition-colors hover:text-text-main hover:decoration-aurora focus:outline-none focus-visible:ring-2 focus-visible:ring-aurora focus-visible:ring-offset-2 focus-visible:ring-offset-background";

export default function Footer() {
  return (
    <footer aria-labelledby="footer-title" className="border-t-2 border-border bg-[#040806] text-xs text-text-muted">
      <div className="mx-auto max-w-7xl px-4 py-10 sm:px-6 lg:px-8 lg:py-12">
        <div className="legal-rule mb-9 border-y border-borderSubtle bg-surface/70 py-3 pl-5 pr-4 sm:flex sm:items-center sm:justify-between sm:gap-6">
          <div>
            <p className="font-bold text-text-main">Independent research platform</p>
            <p className="mt-0.5 text-[11px] leading-relaxed text-text-muted">
              CanadaOpportunityGraph is not a Government of Canada service and is not affiliated with or endorsed by any federal, provincial, territorial, municipal, or Indigenous government.
            </p>
          </div>
          <span className="mt-2 inline-flex shrink-0 items-center gap-1.5 rounded border border-primary/50 bg-primary/10 px-2.5 py-1 font-mono text-[10px] font-bold uppercase tracking-wider text-aurora sm:mt-0">
            <ShieldCheck aria-hidden="true" className="h-3.5 w-3.5" />
            Evidence-linked research
          </span>
        </div>

        <div className="mb-9 grid grid-cols-1 gap-9 sm:grid-cols-2 lg:grid-cols-4">
          <section aria-labelledby="footer-title" className="space-y-3">
            <div className="flex items-center gap-2.5">
              <span aria-hidden="true" className="relative flex h-8 w-8 items-center justify-center overflow-hidden rounded-md border border-border bg-card font-mono text-[10px] font-black text-text-main">
                <span className="absolute inset-x-0 top-0 h-0.5 bg-crimson" />
                CA
              </span>
              <h2 id="footer-title" className="text-sm font-black tracking-tight text-text-main">
                CanadaOpportunityGraph
              </h2>
            </div>
            <p className="max-w-sm leading-relaxed text-text-muted">
              An open, machine-readable intelligence layer for Canadian economic development, infrastructure capital, and sovereign procurement planning.
            </p>
            <div className="pt-1">
              <span className="inline-flex items-center gap-1.5 rounded-md border border-primary/40 bg-card px-2.5 py-1 font-mono text-[10px] font-semibold text-aurora">
                <Database aria-hidden="true" className="h-3.5 w-3.5" />
                CEGS 0.1 reference implementation
              </span>
            </div>
          </section>

          <nav aria-labelledby="intelligence-links-title">
            <h2 id="intelligence-links-title" className="mb-3 text-[11px] font-bold uppercase tracking-[0.14em] text-text-main">
              Intelligence lenses
            </h2>
            <ul className="space-y-2.5" role="list">
              <li><Link href="/" className={footerLinkClass}>Capital Radar</Link></li>
              <li><Link href="/planning" className={footerLinkClass}>Decision Planning</Link></li>
              <li><Link href="/projects" className={footerLinkClass}>Major Projects Directory</Link></li>
              <li><Link href="/map" className={footerLinkClass}>Geospatial Infrastructure Map</Link></li>
              <li><Link href="/capital" className={footerLinkClass}>Canadian Capital Stack</Link></li>
              <li><Link href="/procurement" className={footerLinkClass}>Procurement Pipeline</Link></li>
              <li><Link href="/sources" className={footerLinkClass}>Public Data Explorer</Link></li>
              <li><Link href="/apis" className={footerLinkClass}>Public API Directory</Link></li>
              <li><Link href="/trade" className={footerLinkClass}>Global Trade & Supply Chains</Link></li>
              <li><Link href="/ai-sovereignty" className={footerLinkClass}>AI Sovereignty Index</Link></li>
            </ul>
          </nav>

          <nav aria-labelledby="standards-links-title">
            <h2 id="standards-links-title" className="mb-3 text-[11px] font-bold uppercase tracking-[0.14em] text-text-main">
              Standards & assurance
            </h2>
            <ul className="space-y-2.5" role="list">
              <li><Link href="/cegs" className={footerLinkClass}>CEGS specification v0.1</Link></li>
              <li><Link href="/cegs/adopt" className={footerLinkClass}>Adoption Guide</Link></li>
              <li><Link href="/methodology" className={footerLinkClass}>Scoring Methodology</Link></li>
              <li><Link href="/apis" className={footerLinkClass}>Registered Public APIs</Link></li>
              <li><Link href="/admin" className={footerLinkClass}>Public Data Mesh Status</Link></li>
            </ul>
          </nav>

          <section aria-labelledby="developer-tools-title">
            <h2 id="developer-tools-title" className="mb-3 text-[11px] font-bold uppercase tracking-[0.14em] text-text-main">
              Developer tools
            </h2>
            <div className="space-y-3 font-mono text-[11px]">
              <div className="select-all rounded-md border border-border bg-surface p-3 leading-5 text-text-muted">
                <span aria-hidden="true" className="text-aurora">$</span> cog search "nuclear ontario"<br />
                <span aria-hidden="true" className="text-aurora">$</span> cog cegs validate file.json
              </div>
              <p className="font-sans leading-relaxed text-text-muted">
                Software is released under Apache-2.0; public datasets use CC BY 4.0 with attribution.
              </p>
            </div>
          </section>
        </div>

        <section aria-labelledby="legal-notice-title" className="legal-rule border-y border-borderSubtle bg-card/45 py-5 pl-5 pr-4 text-[11px] leading-relaxed text-text-muted">
          <h2 id="legal-notice-title" className="font-bold uppercase tracking-[0.12em] text-text-main">
            Research, regulatory & investment-use notice
          </h2>
          <div className="mt-2 grid gap-2 lg:grid-cols-2 lg:gap-8">
            <p>
              Information is compiled from public records and attributed sources and may be incomplete, delayed, revised, or contain errors. Verify dates, filings, approvals, counterparties, and source records before relying on any item for a material decision.
            </p>
            <p>
              Scores, scenarios, forecasts, and AI-assisted outputs are research indicators—not guarantees of outcome, investment recommendations, legal or tax advice, an offer or solicitation, or a substitute for regulatory, procurement, financial, or Indigenous-rights due diligence.
            </p>
          </div>
        </section>

        <div className="mt-6 flex flex-col gap-2 border-t border-borderSubtle pt-5 text-[10px] leading-relaxed text-text-subtle sm:flex-row sm:items-center sm:justify-between">
          <p>© 2026 CanadaOpportunityGraph Open Consortium. Apache-2.0 software.</p>
          <p className="font-mono uppercase tracking-wider">English/French bilingual interface · WCAG 2.2 AA design target</p>
        </div>
      </div>
    </footer>
  );
}
