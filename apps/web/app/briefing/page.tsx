"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ArrowUpRight,
  Check,
  Copy,
  Database,
  FileText,
  Layers,
  Printer,
  ShieldCheck,
  Sliders,
} from "lucide-react";
import { FALLBACK_PROJECTS, SNAPSHOT_MANIFEST } from "@/lib/data";

const projects = FALLBACK_PROJECTS;
const totalCapexCad = projects.reduce((sum, project) => sum + project.capex_cad, 0);
const knownCapexProjects = projects.filter((project) => project.capex_cad > 0).length;
const evidenceLinkedProjects = projects.filter((project) => (project.evidence?.length ?? 0) > 0).length;
const scoredProjects = projects.filter((project) => (project.score_details?.length ?? 0) > 0).length;
const projectsChecksum = SNAPSHOT_MANIFEST.checksums_sha256["public/projects.jsonl"];
const generatedDate = new Intl.DateTimeFormat("en-CA", {
  dateStyle: "long",
  timeZone: "UTC",
}).format(new Date(SNAPSHOT_MANIFEST.generated_at));

const sectorPortfolios = [...new Set(projects.map((project) => project.sector))]
  .map((sector) => {
    const sectorProjects = projects.filter((project) => project.sector === sector);
    return {
      sector,
      projects: sectorProjects,
      capexCad: sectorProjects.reduce((sum, project) => sum + project.capex_cad, 0),
      unknownCapex: sectorProjects.filter((project) => project.capex_cad === 0).length,
      evidenceLinked: sectorProjects.filter((project) => (project.evidence?.length ?? 0) > 0).length,
    };
  })
  .sort((a, b) => b.capexCad - a.capexCad || a.sector.localeCompare(b.sector));

function formatCad(value: number): string {
  if (value >= 1e9) return `$${(value / 1e9).toFixed(1)}B`;
  if (value >= 1e6) return `$${(value / 1e6).toFixed(0)}M`;
  return value > 0 ? `$${value.toLocaleString("en-CA")}` : "Not reported";
}

export default function ExecutiveBriefingPage() {
  const [activeSector, setActiveSector] = useState("ALL");
  const [targetCapex, setTargetCapex] = useState(Math.round(totalCapexCad / 1e8) / 10);
  const [publicFinancePct, setPublicFinancePct] = useState(15);
  const [taxSupportPct, setTaxSupportPct] = useState(10);
  const [indigenousEquityPct, setIndigenousEquityPct] = useState(5);
  const [copyStatus, setCopyStatus] = useState<"idle" | "copied" | "failed">("idle");

  const publicSupportPct = publicFinancePct + taxSupportPct;
  const availableForPrivatePct = Math.max(0, 100 - publicSupportPct - indigenousEquityPct);
  const commercialDebtPct = Math.min(35, availableForPrivatePct);
  const sponsorEquityPct = Math.max(0, availableForPrivatePct - commercialDebtPct);
  const targetCapexCad = targetCapex * 1e9;
  const privateCapitalPct = commercialDebtPct + sponsorEquityPct;
  const illustrativeRatio = publicSupportPct > 0 ? privateCapitalPct / publicSupportPct : null;
  const visiblePortfolios = activeSector === "ALL"
    ? sectorPortfolios
    : sectorPortfolios.filter((portfolio) => portfolio.sector === activeSector);

  const copyChecksum = async () => {
    try {
      await navigator.clipboard.writeText(projectsChecksum);
      setCopyStatus("copied");
    } catch {
      setCopyStatus("failed");
    }
    window.setTimeout(() => setCopyStatus("idle"), 3000);
  };

  return (
    <div className="mx-auto max-w-7xl space-y-10 px-4 py-8 sm:px-6 lg:px-8">
      <header className="relative overflow-hidden rounded-2xl border border-border/90 bg-gradient-to-b from-card to-surface p-6 shadow-2xl sm:p-8">
        <div aria-hidden="true" className="pointer-events-none absolute -right-16 -top-16 h-96 w-96 rounded-full bg-gold/10 blur-3xl" />
        <div className="relative z-10">
          <div className="flex flex-col gap-3 border-b border-borderSubtle pb-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="inline-flex items-center gap-2 font-mono text-xs font-bold text-gold">
              <FileText aria-hidden="true" className="h-4 w-4" />
              INDEPENDENT DECISION BRIEF · REVIEWED SNAPSHOT
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <button
                type="button"
                onClick={() => window.print()}
                className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-3 text-xs font-semibold text-text-main hover:border-primary hover:text-aurora"
              >
                <Printer aria-hidden="true" className="h-4 w-4" />
                Print brief
              </button>
              <span className="rounded-full border border-primary/40 bg-primary/10 px-3 py-1.5 font-mono text-[10px] font-semibold text-aurora">
                DATA THROUGH {generatedDate.toUpperCase()} UTC
              </span>
            </div>
          </div>

          <div className="mt-5 max-w-4xl space-y-3">
            <h1 className="text-3xl font-black tracking-tight text-text-main sm:text-5xl">
              Canadian capital <span className="text-gold">decision brief</span>
            </h1>
            <p className="text-sm leading-relaxed text-text-muted">
              A planning view of the bundled CEGS {SNAPSHOT_MANIFEST.cegs} dataset. It summarizes published records and exposes assumptions; it is not an official government document, a live market feed, or a recommendation to invest.
            </p>
          </div>

          <dl className="relative z-10 mt-8 grid grid-cols-2 gap-4 border-t border-borderSubtle pt-6 lg:grid-cols-4">
            <div className="glass-card rounded-xl p-4">
              <dt className="font-mono text-[10px] uppercase text-text-subtle">Snapshot projects</dt>
              <dd className="mt-1 text-3xl font-black tabular-nums text-aurora">{projects.length}</dd>
              <p className="mt-1 text-[11px] text-text-muted">Manifest reports {SNAPSHOT_MANIFEST.record_counts.projects}</p>
            </div>
            <div className="glass-card rounded-xl p-4">
              <dt className="font-mono text-[10px] uppercase text-text-subtle">Reported CAPEX total</dt>
              <dd className="mt-1 text-3xl font-black tabular-nums text-gold">{formatCad(totalCapexCad)}</dd>
              <p className="mt-1 text-[11px] text-text-muted">Known for {knownCapexProjects} of {projects.length} records</p>
            </div>
            <div className="glass-card rounded-xl p-4">
              <dt className="font-mono text-[10px] uppercase text-text-subtle">Evidence coverage</dt>
              <dd className="mt-1 text-3xl font-black tabular-nums text-text-main">{evidenceLinkedProjects}/{projects.length}</dd>
              <p className="mt-1 text-[11px] text-text-muted">{SNAPSHOT_MANIFEST.record_counts.evidence} evidence records in manifest</p>
            </div>
            <div className="glass-card rounded-xl p-4">
              <dt className="font-mono text-[10px] uppercase text-text-subtle">Model coverage</dt>
              <dd className="mt-1 text-3xl font-black tabular-nums text-text-main">{scoredProjects}/{projects.length}</dd>
              <p className="mt-1 text-[11px] text-text-muted">Versioned scores, not forecasts</p>
            </div>
          </dl>
        </div>
      </header>

      <section aria-labelledby="portfolio-heading" className="space-y-5">
        <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 id="portfolio-heading" className="text-xl font-black text-text-main">Observed sector portfolios</h2>
            <p className="mt-1 max-w-3xl text-xs leading-relaxed text-text-muted">
              These groupings and totals are calculated from the checked-in project snapshot. “Not reported” values are excluded from CAPEX totals, not estimated.
            </p>
          </div>
          <div className="flex max-w-full gap-1 overflow-x-auto rounded-xl border border-borderSubtle bg-surface p-1" aria-label="Filter sector portfolios">
            <button
              type="button"
              aria-pressed={activeSector === "ALL"}
              onClick={() => setActiveSector("ALL")}
              className={`min-h-9 shrink-0 rounded-lg px-3 text-xs font-semibold ${activeSector === "ALL" ? "bg-card text-aurora" : "text-text-muted hover:text-text-main"}`}
            >
              All sectors
            </button>
            {sectorPortfolios.map((portfolio) => (
              <button
                key={portfolio.sector}
                type="button"
                aria-pressed={activeSector === portfolio.sector}
                onClick={() => setActiveSector(portfolio.sector)}
                className={`min-h-9 shrink-0 rounded-lg px-3 text-xs font-semibold ${activeSector === portfolio.sector ? "bg-card text-aurora" : "text-text-muted hover:text-text-main"}`}
              >
                {portfolio.sector}
              </button>
            ))}
          </div>
        </div>

        <div className="grid gap-5 md:grid-cols-2">
          {visiblePortfolios.map((portfolio) => {
            const topProjects = [...portfolio.projects]
              .sort((a, b) => b.capex_cad - a.capex_cad || a.name.localeCompare(b.name))
              .slice(0, 3);
            return (
              <article key={portfolio.sector} className="glass-panel rounded-2xl p-6">
                <div className="flex items-start justify-between gap-4 border-b border-borderSubtle pb-4">
                  <div>
                    <p className="font-mono text-[10px] uppercase tracking-wider text-aurora">Observed sector</p>
                    <h3 className="mt-1 text-lg font-black text-text-main">{portfolio.sector}</h3>
                  </div>
                  <div className="text-right">
                    <p className="font-mono text-[10px] uppercase text-text-subtle">Reported CAPEX</p>
                    <p className="mt-1 text-xl font-black tabular-nums text-gold">{formatCad(portfolio.capexCad)}</p>
                  </div>
                </div>
                <dl className="mt-4 grid grid-cols-3 gap-3 text-xs">
                  <div><dt className="text-text-subtle">Projects</dt><dd className="mt-1 font-bold tabular-nums text-text-main">{portfolio.projects.length}</dd></div>
                  <div><dt className="text-text-subtle">Evidence-linked</dt><dd className="mt-1 font-bold tabular-nums text-text-main">{portfolio.evidenceLinked}</dd></div>
                  <div><dt className="text-text-subtle">CAPEX unknown</dt><dd className="mt-1 font-bold tabular-nums text-text-main">{portfolio.unknownCapex}</dd></div>
                </dl>
                <div className="mt-5 space-y-2">
                  <p className="font-mono text-[10px] uppercase text-text-subtle">Largest records by reported CAPEX</p>
                  {topProjects.map((project) => (
                    <Link key={project.id} href={`/projects/${project.slug}`} className="flex min-h-10 items-center justify-between gap-3 rounded-lg border border-borderSubtle bg-card px-3 py-2 text-xs hover:border-primary">
                      <span className="min-w-0 truncate font-semibold text-text-main">{project.name}</span>
                      <span className="shrink-0 font-mono font-bold text-aurora">{formatCad(project.capex_cad)}</span>
                    </Link>
                  ))}
                </div>
              </article>
            );
          })}
        </div>
      </section>

      <section aria-labelledby="scenario-heading" className="glass-card space-y-6 rounded-2xl p-6 shadow-2xl sm:p-8">
        <div className="flex flex-col gap-4 border-b border-borderSubtle pb-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <div className="inline-flex items-center gap-2 rounded-full border border-gold/40 bg-gold/10 px-3 py-1 font-mono text-[11px] text-gold">
              <Sliders aria-hidden="true" className="h-3.5 w-3.5" /> USER-DEFINED SCENARIO
            </div>
            <h2 id="scenario-heading" className="mt-2 text-2xl font-black text-text-main">Illustrative capital-stack calculator</h2>
            <p className="mt-1 max-w-3xl text-xs leading-relaxed text-text-muted">
              Change the assumptions to inspect arithmetic only. Percentages do not indicate program eligibility, committed financing, expected returns, or an approved policy package.
            </p>
          </div>
          <div className="rounded-xl border border-borderSubtle bg-surface p-3 text-right">
            <p className="font-mono text-[10px] uppercase text-text-subtle">Illustrative private / public ratio</p>
            <p className="mt-1 text-2xl font-black tabular-nums text-aurora">{illustrativeRatio === null ? "N/A" : `${illustrativeRatio.toFixed(2)}×`}</p>
          </div>
        </div>

        <div className="grid gap-5 rounded-xl border border-borderSubtle bg-surface/50 p-4 sm:grid-cols-2 lg:grid-cols-4">
          <label className="space-y-2 text-xs font-mono" htmlFor="scenario-capex">
            <span className="flex justify-between gap-3 text-text-subtle"><span>Portfolio CAPEX</span><output className="font-bold text-aurora">${targetCapex.toFixed(1)}B</output></span>
            <input id="scenario-capex" type="range" min="1" max="500" step="1" value={targetCapex} onChange={(event) => setTargetCapex(Number(event.target.value))} className="w-full" />
          </label>
          <label className="space-y-2 text-xs font-mono" htmlFor="scenario-public-finance">
            <span className="flex justify-between gap-3 text-text-subtle"><span>Public finance assumption</span><output className="font-bold text-gold">{publicFinancePct}%</output></span>
            <input id="scenario-public-finance" type="range" min="0" max="30" step="1" value={publicFinancePct} onChange={(event) => setPublicFinancePct(Number(event.target.value))} className="w-full" />
          </label>
          <label className="space-y-2 text-xs font-mono" htmlFor="scenario-tax-support">
            <span className="flex justify-between gap-3 text-text-subtle"><span>Tax-support assumption</span><output className="font-bold text-gold">{taxSupportPct}%</output></span>
            <input id="scenario-tax-support" type="range" min="0" max="30" step="1" value={taxSupportPct} onChange={(event) => setTaxSupportPct(Number(event.target.value))} className="w-full" />
          </label>
          <label className="space-y-2 text-xs font-mono" htmlFor="scenario-indigenous-equity">
            <span className="flex justify-between gap-3 text-text-subtle"><span>Indigenous-partner equity assumption</span><output className="font-bold text-text-main">{indigenousEquityPct}%</output></span>
            <input id="scenario-indigenous-equity" type="range" min="0" max="15" step="1" value={indigenousEquityPct} onChange={(event) => setIndigenousEquityPct(Number(event.target.value))} className="w-full" />
          </label>
        </div>

        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
          {[
            ["Sponsor equity", sponsorEquityPct, "bg-emerald-600"],
            ["Commercial debt (capped)", commercialDebtPct, "bg-teal-500"],
            ["Public finance", publicFinancePct, "bg-amber-500"],
            ["Tax support", taxSupportPct, "bg-yellow-400"],
            ["Indigenous-partner equity", indigenousEquityPct, "bg-orange-500"],
          ].map(([label, percentage, colour]) => {
            const pct = Number(percentage);
            return (
              <div key={String(label)} className="rounded-xl border border-borderSubtle bg-surface p-3 text-xs">
                <div className="flex items-center gap-2 text-text-main"><span aria-hidden="true" className={`h-2.5 w-2.5 rounded-full ${colour}`} /><span className="font-semibold">{label}</span></div>
                <p className="mt-2 font-mono font-bold text-aurora">{formatCad(targetCapexCad * pct / 100)} · {pct.toFixed(0)}%</p>
              </div>
            );
          })}
        </div>
      </section>

      <section aria-labelledby="decision-questions-heading" className="rounded-2xl border border-gold/40 bg-gold/10 p-6">
        <div className="flex items-center gap-2">
          <Layers aria-hidden="true" className="h-5 w-5 text-gold" />
          <h2 id="decision-questions-heading" className="text-lg font-black text-text-main">Questions this snapshot can support</h2>
        </div>
        <ul className="mt-4 grid gap-3 text-xs leading-relaxed text-text-muted md:grid-cols-3">
          <li className="rounded-xl border border-borderSubtle bg-surface/70 p-4">Where is reported project CAPEX concentrated by sector, and where are amounts still unknown?</li>
          <li className="rounded-xl border border-borderSubtle bg-surface/70 p-4">Which records have evidence and score coverage sufficient for deeper review?</li>
          <li className="rounded-xl border border-borderSubtle bg-surface/70 p-4">How would a user-defined financing mix divide a portfolio, before project-specific diligence?</li>
        </ul>
      </section>

      <section aria-labelledby="provenance-heading" className="glass-card flex flex-col gap-4 rounded-2xl p-5 text-xs sm:flex-row sm:items-center sm:justify-between">
        <div className="min-w-0">
          <div className="flex items-center gap-2 font-semibold text-text-main">
            <ShieldCheck aria-hidden="true" className="h-4 w-4 text-aurora" />
            <h2 id="provenance-heading">Snapshot provenance</h2>
          </div>
          <p className="mt-2 text-text-muted">Manifest <code>{SNAPSHOT_MANIFEST.dataset_version}</code> declares this SHA-256 checksum for <code>public/projects.jsonl</code>:</p>
          <code className="mt-2 block overflow-x-auto rounded border border-borderSubtle bg-surface px-2 py-1.5 text-[10px] text-aurora">{projectsChecksum}</code>
          <p aria-live="polite" className="mt-2 min-h-4 text-[11px] text-text-muted">
            {copyStatus === "failed" ? "Clipboard access was unavailable. Select the checksum above to copy it manually." : copyStatus === "copied" ? "Checksum copied." : "Copying a checksum does not verify a local file; compare it with a SHA-256 digest of the exported dataset."}
          </p>
        </div>
        <div className="flex shrink-0 flex-wrap gap-2">
          <button type="button" onClick={copyChecksum} className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-4 font-semibold text-text-main hover:border-primary hover:text-aurora">
            {copyStatus === "copied" ? <Check aria-hidden="true" className="h-4 w-4" /> : <Copy aria-hidden="true" className="h-4 w-4" />}
            Copy checksum
          </button>
          <a href="/api/v1/cegs/export" className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-primary px-4 font-black text-[#050b08] hover:bg-aurora-mint">
            <Database aria-hidden="true" className="h-4 w-4" /> Open export <ArrowUpRight aria-hidden="true" className="h-4 w-4" />
          </a>
        </div>
      </section>
    </div>
  );
}
