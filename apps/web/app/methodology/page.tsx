import { BookOpen, ShieldCheck, Calculator, FileText, CheckCircle2, AlertTriangle, Layers } from "lucide-react";

export default function MethodologyPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border pb-6">
        <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-card border border-border text-[11px] font-mono text-accent-cyan mb-2">
          <BookOpen className="h-3 w-3" />
          OPEN SCORING & PROVENANCE METHODOLOGY
        </div>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
          Public Methodology & Analytical Framework
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-1 max-w-4xl leading-relaxed">
          Full mathematical transparency regarding deterministic score calculations, source tiering, change detection, 
          epistemic confidence invariants, and entity resolution in CanadaOpportunityGraph.
        </p>
      </div>

      {/* Grid: 4 Scoring Engines Breakdown */}
      <div className="space-y-6">
        <h2 className="text-base font-bold text-text-main">Deterministic Scoring Algorithms</h2>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Buildability */}
          <div className="bg-card p-5 rounded border border-border space-y-3">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-2">
              <span className="font-bold text-text-main text-sm">Buildability Score (0–100)</span>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface text-accent-green border border-borderSubtle">
                buildability-v1.0
              </span>
            </div>
            <p className="text-xs text-text-subtle leading-relaxed">
              Measures the objective probability that an announced major project will successfully reach operational completion.
            </p>
            <div className="space-y-1.5 text-xs font-mono">
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Stage Progression Maturity</span>
                <span className="text-text-main font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Committed vs Required Financing</span>
                <span className="text-text-main font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Indigenous Agreements & Consensus</span>
                <span className="text-text-main font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Regulatory & Environmental Approvals</span>
                <span className="text-text-main font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Land & Site Control / Clear Title</span>
                <span className="text-text-main font-bold">10%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Commercial Offtake & PPAs</span>
                <span className="text-text-main font-bold">10%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Grid / Infrastructure Readiness</span>
                <span className="text-text-main font-bold">10%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1">
                <span className="text-text-muted">Proponent Institutional Credibility</span>
                <span className="text-text-main font-bold">10%</span>
              </div>
            </div>
          </div>

          {/* Investability */}
          <div className="bg-card p-5 rounded border border-border space-y-3">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-2">
              <span className="font-bold text-text-main text-sm">Investability Score (0–100)</span>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface text-accent-cyan border border-borderSubtle">
                investability-v1.0
              </span>
            </div>
            <p className="text-xs text-text-subtle leading-relaxed">
              Assesses the presence of a credible capital opportunity, financing momentum, and government co-investment de-risking.
            </p>
            <div className="space-y-1.5 text-xs font-mono">
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Capital Scale & Financing Gap</span>
                <span className="text-text-main font-bold">25%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Financing Velocity & Crown Participation</span>
                <span className="text-text-main font-bold">25%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Strategic National Demand / Macro Tailwinds</span>
                <span className="text-text-main font-bold">25%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1">
                <span className="text-text-muted">Execution Maturity & Downside Protection</span>
                <span className="text-text-main font-bold">25%</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Sourcing Tiering Framework */}
      <div className="bg-card p-6 rounded border border-border space-y-4">
        <h2 className="text-base font-bold text-text-main">Authoritative Source Tiering Hierarchy</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-xs">
          <div className="p-3.5 rounded bg-surface border border-borderSubtle space-y-1">
            <div className="font-bold text-accent-green font-mono">Tier 1: Primary & Statutory</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Official federal/provincial registries: Impact Assessment Agency of Canada (IAAC), Canadian Energy Regulator (CER), CanadaBuys, NRCan.
            </p>
          </div>
          <div className="p-3.5 rounded bg-surface border border-borderSubtle space-y-1">
            <div className="font-bold text-accent-cyan font-mono">Tier 2: Corporate Issuers</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              SEDAR+ regulatory filings, audited annual reports, TSX market disclosures, and official proponent presentations.
            </p>
          </div>
          <div className="p-3.5 rounded bg-surface border border-borderSubtle space-y-1">
            <div className="font-bold text-accent-gold font-mono">Tier 3: Credible News</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Major financial and business press (The Globe and Mail, Bloomberg, Reuters, Canadian Press). Sourced only for breaking signals.
            </p>
          </div>
          <div className="p-3.5 rounded bg-surface border border-borderSubtle space-y-1">
            <div className="font-bold text-primary font-mono">Tier 4: Secondary Research</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Industry association bulletins, consultancy whitepapers, academic working papers. Never treated as primary evidence.
            </p>
          </div>
        </div>
      </div>

      {/* Epistemic Invariants */}
      <div className="p-5 rounded bg-surface border border-borderSubtle text-xs space-y-2 text-text-subtle leading-relaxed">
        <div className="flex items-center gap-1.5 font-bold text-text-main">
          <ShieldCheck className="h-4 w-4 text-accent-cyan" />
          <span>EPISTEMIC INTEGRITY INVARIANTS</span>
        </div>
        <p className="text-[11px]">
          1. <strong>Missing is UNKNOWN:</strong> Missing values are never silently coerced to 0, false, or complete.<br />
          2. <strong>Immutable Event Ledger:</strong> History is never overwritten. Every stage transition is recorded as a distinct event.<br />
          3. <strong>Anti-Theatre Guarantee:</strong> LLMs are strictly forbidden from fabricating authoritative numerical scores. All scores are computed deterministically.
        </p>
      </div>
    </div>
  );
}
