import { BookOpen, ShieldCheck, Calculator, FileText, CheckCircle2, AlertTriangle, Layers, Radio } from "lucide-react";

export default function MethodologyPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
          <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
          OPEN SCORING & PROVENANCE METHODOLOGY
        </div>
        <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
          Public Methodology & <span className="text-aurora">Analytical Framework</span>
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-2 max-w-4xl leading-relaxed">
          Full mathematical transparency regarding deterministic score calculations, source tiering, change detection, 
          epistemic confidence invariants, and entity resolution in CanadaOpportunityGraph.
        </p>
      </div>

      {/* Grid: 5 Scoring Engines Breakdown */}
      <div className="space-y-6">
        <h2 className="text-base font-bold text-text-main font-mono uppercase tracking-wider flex items-center gap-2">
          <Calculator className="h-4 w-4 text-aurora" />
          <span>Deterministic Scoring Algorithms</span>
        </h2>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Buildability */}
          <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <span className="font-bold text-text-main text-base">Buildability Score (0–100)</span>
              <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-primary/10 text-aurora border border-primary/30 font-bold">
                buildability-v2.1
              </span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">
              A deterministic readiness index built from observed project attributes. It is not a calibrated probability of completion.
            </p>
            <div className="space-y-1.5 text-xs font-mono">
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Stage Progression Maturity</span>
                <span className="text-aurora font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Committed vs Required Financing</span>
                <span className="text-aurora font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Indigenous Agreements & Consensus</span>
                <span className="text-aurora font-bold">15%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Regulatory & Environmental Approvals</span>
                <span className="text-aurora font-bold">15%</span>
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
                <span className="text-text-muted">Observed Execution Evidence</span>
                <span className="text-text-main font-bold">10%</span>
              </div>
            </div>
          </div>

          {/* Investability */}
          <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <span className="font-bold text-text-main text-base">Investability Score (0–100)</span>
              <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-gold/10 text-gold border border-gold/30 font-bold">
                investability-v2.0
              </span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">
              Assesses the presence of a credible capital opportunity, financing momentum, and federal co-investment de-risking participation.
            </p>
            <div className="space-y-1.5 text-xs font-mono">
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Capital Scale & Financing Gap</span>
                <span className="text-gold font-bold">25%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Financing Velocity & Crown Participation</span>
                <span className="text-gold font-bold">25%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1 border-b border-borderSubtle">
                <span className="text-text-muted">Strategic National Demand / Macro Tailwinds</span>
                <span className="text-gold font-bold">25%</span>
              </div>
              <div className="flex justify-between text-[11px] py-1">
                <span className="text-text-muted">Execution Maturity & Downside Protection</span>
                <span className="text-gold font-bold">25%</span>
              </div>
            </div>
          </div>

          <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <span className="font-bold text-text-main text-base">Supplierability Score (0–100)</span>
              <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-primary/10 text-aurora border border-primary/30 font-bold">supplierability-v2.0</span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">Measures contracting opportunity from project stage, confirmed tenders, technical intensity, derived downstream needs, and Canada&apos;s official logistics context.</p>
            <div className="space-y-1.5 text-xs font-mono">
              {[["Procurement proximity", "25%"], ["Observed tender volume", "20%"], ["Technical intensity", "20%"], ["Downstream opportunity density", "20%"], ["World Bank logistics performance", "15%"]].map(([label, weight]) => (
                <div key={label} className="flex justify-between text-[11px] py-1 border-b border-borderSubtle last:border-0"><span className="text-text-muted">{label}</span><span className="text-aurora font-bold">{weight}</span></div>
              ))}
            </div>
          </div>

          <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <span className="font-bold text-text-main text-base">Strategicity Score (0–100)</span>
              <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-primary/10 text-aurora border border-primary/30 font-bold">strategicity-v2.0</span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">Classifies the sourced project sector and geography across critical minerals, energy security, sovereign infrastructure, and AI sovereignty.</p>
            <div className="grid grid-cols-2 gap-2 font-mono text-[11px] text-text-muted">
              {["Critical minerals · 25%", "Energy security · 25%", "Sovereign infrastructure · 25%", "AI sovereignty · 25%"].map((item) => <div key={item} className="rounded-lg border border-borderSubtle bg-surface p-2">{item}</div>)}
            </div>
          </div>

          <div className="glass-card p-6 rounded-2xl border border-gold/40 space-y-4 shadow-xl lg:col-span-2">
            <div className="flex flex-wrap items-center justify-between gap-2 border-b border-borderSubtle pb-3">
              <span className="font-bold text-text-main text-base">Trade Resilience Score (0–100)</span>
              <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-gold/10 text-gold border border-gold/30 font-bold">trade-resilience-v1.0</span>
            </div>
            <p className="text-xs text-text-muted leading-relaxed">A Canada-level operating-context index—not project-specific import exposure. Inputs are normalized World Bank observations with per-factor evidence IDs and reproducible SHA-256 input hashes.</p>
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4 font-mono text-[11px]">
              {[["Normalized LPI", "45%"], ["Trade / GDP", "20%"], ["Two-way merchandise balance", "20%"], ["High-tech export share", "15%"]].map(([label, weight]) => (
                <div key={label} className="rounded-lg border border-borderSubtle bg-surface p-3"><div className="text-text-muted">{label}</div><div className="mt-1 text-gold font-bold">{weight}</div></div>
              ))}
            </div>
            <p className="font-mono text-[10px] leading-relaxed text-text-subtle">Missing inputs are never replaced with synthetic values. Available weights are renormalized, and the published score reports coverage, unknown factors, factor-to-evidence mappings, calculation timestamp, and input hash.</p>
          </div>
        </div>
      </div>

      {/* Provenance & Epistemic Invariants */}
      <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
        <h2 className="text-base font-bold text-text-main font-mono uppercase tracking-wider flex items-center gap-2">
          <ShieldCheck className="h-4 w-4 text-aurora" />
          <span>Epistemic Integrity & Source Hierarchy</span>
        </h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs font-mono">
          <div className="p-4 rounded-xl bg-surface border border-borderSubtle space-y-2">
            <div className="text-aurora font-bold text-sm">Tier 1: Statutory Registries</div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Eligible examples include statutory registries and official contract or regulatory records. The current snapshot&apos;s actual coverage is listed in Sources and its release manifest.
            </p>
          </div>
          <div className="p-4 rounded-xl bg-surface border border-borderSubtle space-y-2">
            <div className="text-gold font-bold text-sm">Tier 2: Official Proponent Data</div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Quarterly investor decks, technical NI 43-101 reports, official press releases, and Indigenous co-development protocols.
            </p>
          </div>
          <div className="p-4 rounded-xl bg-surface border border-borderSubtle space-y-2">
            <div className="text-text-main font-bold text-sm">Tier 3: Secondary Media</div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Trade publications and general news feeds. Used solely for early discovery triggers and never as sole ground truth.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
