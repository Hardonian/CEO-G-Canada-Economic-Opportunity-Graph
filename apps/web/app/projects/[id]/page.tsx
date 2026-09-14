import Link from "next/link";
import { notFound } from "next/navigation";
import { 
  ShieldCheck, 
  MapPin, 
  Calendar, 
  Download, 
  ExternalLink, 
  Layers, 
  Pickaxe, 
  Coins, 
  FileText, 
  GitBranch, 
  CheckCircle2, 
  Clock, 
  ArrowLeft,
  ChevronRight,
  TrendingUp,
  Cpu,
  Radio
} from "lucide-react";
import { getProjectBySlug } from "@/lib/data";

interface Props {
  params: Promise<{ id: string }>;
}

export default async function ProjectProfilePage({ params }: Props) {
  const { id } = await params;
  const project = await getProjectBySlug(id);

  if (!project) {
    notFound();
  }

  // Scores are sparse by design. Never substitute an unevidenced score: doing
  // so both caused the production crash and overstated analytical coverage.
  const score = (name: string) => {
    const value = project.scores?.[name];
    return typeof value === "number" && Number.isFinite(value) ? value : null;
  };
  const buildability = score("buildability");
  const investability = score("investability");
  const supplierability = score("supplierability");
  const strategicity = score("strategicity");
  const tradeResilience = score("trade_resilience");
  const evidence = project.evidence ?? [];
  const scoreDetails = project.score_details ?? [];

  const stages = [
    "ANNOUNCED",
    "FEASIBILITY",
    "ENVIRONMENTAL_REVIEW",
    "PERMITTING",
    "PROCUREMENT",
    "FID",
    "CONSTRUCTION",
    "COMMISSIONING",
    "OPERATING"
  ];

  const currentStageIndex = stages.indexOf(project.current_stage);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Breadcrumbs & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/80 pb-4">
        <Link href="/projects" className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-aurora transition-colors font-mono">
          <ArrowLeft className="h-3.5 w-3.5" /> Back to Major Projects Directory
        </Link>
        <div className="flex items-center gap-2">
          <a
            href={`/api/v1/export/project/${project.slug}?format=markdown`}
            target="_blank"
            rel="noreferrer"
            className="px-3.5 py-1.5 rounded-xl bg-surface border border-border hover:border-aurora text-text-muted hover:text-aurora text-xs transition-all flex items-center gap-1.5 font-mono shadow-sm"
          >
            <Download className="h-3.5 w-3.5" /> Markdown
          </a>
          <a
            href={`/api/v1/export/project/${project.slug}?format=cegs`}
            target="_blank"
            rel="noreferrer"
            className="px-3.5 py-1.5 rounded-xl bg-card border border-primary/40 text-aurora text-xs transition-all flex items-center gap-1.5 font-mono hover:bg-cardHover shadow-sm"
          >
            <Download className="h-3.5 w-3.5" /> CEGS 0.1 JSON
          </a>
        </div>
      </div>

      {/* Profile Header Dossier */}
      <div className="glass-card p-6 sm:p-8 rounded-2xl border border-border/80 space-y-5 shadow-2xl relative overflow-hidden">
        {/* Subtle Glow */}
        <div className="absolute top-0 right-0 -mt-12 -mr-12 w-80 h-80 rounded-full bg-aurora/10 blur-3xl pointer-events-none"></div>

        <div className="flex flex-wrap items-center gap-2 text-xs font-mono relative z-10">
          <span className="px-2.5 py-0.5 rounded-full bg-surface border border-borderSubtle text-aurora font-bold">
            {project.province}
          </span>
          <span className="px-2.5 py-0.5 rounded-full bg-surface border border-borderSubtle text-text-main">
            {project.sector}
          </span>
          <span className="px-2.5 py-0.5 rounded-md bg-gold/10 border border-gold/30 text-gold font-semibold">
            {project.current_stage}
          </span>
          <span className="px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-aurora flex items-center gap-1 font-semibold">
            <ShieldCheck className="h-3.5 w-3.5" /> {project.confidence} EVIDENCE
          </span>
        </div>

        <div className="space-y-1 relative z-10">
          <h1 className="text-2xl sm:text-4xl font-black text-text-main tracking-tight">
            {project.name}
          </h1>
          <div className="flex items-center gap-2 text-xs text-text-subtle pt-1">
            <MapPin className="h-3.5 w-3.5 text-aurora" />
            <span>
              {project.location_name}
              {project.latitude != null && project.longitude != null
                ? ` (Lat: ${project.latitude.toFixed(4)}°, Long: ${project.longitude.toFixed(4)}°)`
                : " (coordinates not published)"}
            </span>
            <span>•</span>
            <span className="text-text-muted">Subsector: {project.subsector}</span>
          </div>
        </div>

        <p className="text-sm text-text-muted leading-relaxed max-w-4xl relative z-10">
          {project.summary}
        </p>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-4 border-t border-borderSubtle text-xs relative z-10">
          <div>
            <div className="text-[10px] font-mono text-text-subtle uppercase">Reported CAPEX</div>
            <div className="text-xl font-black font-tabular text-aurora mt-0.5">
              {project.capex_cad > 0 ? `$${(project.capex_cad / 1e9).toFixed(2)}B CAD` : "UNKNOWN"}
            </div>
          </div>
          <div>
            <div className="text-[10px] font-mono text-text-subtle uppercase">Lifecycle Phase</div>
            <div className="text-sm font-semibold text-gold mt-1 font-mono">
              {project.current_stage}
            </div>
          </div>
          <div>
            <div className="text-[10px] font-mono text-text-subtle uppercase">Last Sourced Update</div>
            <div className="text-sm text-text-muted mt-1 font-mono">
              {new Date(project.last_meaningful_update).toLocaleDateString("en-CA")}
            </div>
          </div>
          <div>
            <div className="text-[10px] font-mono text-text-subtle uppercase">Canonical CEGS ID</div>
            <div className="text-xs font-mono text-aurora mt-1 truncate">
              cegs:project:ca:{project.province.toLowerCase()}:{project.slug}
            </div>
          </div>
        </div>
      </div>

      {/* 5-D Deterministic Scorecards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-5 gap-4">
        {/* Buildability */}
        <div className="glass-card p-5 rounded-2xl border border-border/80 space-y-2.5 shadow-xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-mono uppercase text-text-subtle">Buildability</span>
            <span className="text-xs font-bold font-tabular text-aurora">
              {buildability == null ? "NOT SCORED" : `${buildability.toFixed(1)}/100`}
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-primary rounded-full" style={{ width: `${buildability ?? 0}%` }}></div>
          </div>
          <p className="text-[11px] text-text-subtle leading-tight pt-1">
            Execution likelihood based on regulatory clearance, site control, and Indigenous consensus.
          </p>
        </div>

        {/* Investability */}
        <div className="glass-card p-5 rounded-2xl border border-border/80 space-y-2.5 shadow-xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-mono uppercase text-text-subtle">Investability</span>
            <span className="text-xs font-bold font-tabular text-gold">
              {investability == null ? "NOT SCORED" : `${investability.toFixed(1)}/100`}
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-gold rounded-full" style={{ width: `${investability ?? 0}%` }}></div>
          </div>
          <p className="text-[11px] text-text-subtle leading-tight pt-1">
            Capital opportunity appeal, federal de-risking participation, and offtake strength.
          </p>
        </div>

        {/* Supplierability */}
        <div className="glass-card p-5 rounded-2xl border border-border/80 space-y-2.5 shadow-xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-mono uppercase text-text-subtle">Supplierability</span>
            <span className="text-xs font-bold font-tabular text-aurora-mint">
              {supplierability == null ? "NOT SCORED" : `${supplierability.toFixed(1)}/100`}
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-aurora-mint rounded-full" style={{ width: `${supplierability ?? 0}%` }}></div>
          </div>
          <p className="text-[11px] text-text-subtle leading-tight pt-1">
            Downstream tender density and specialized equipment/engineering requirements.
          </p>
        </div>

        {/* Strategicity */}
        <div className="glass-card p-5 rounded-2xl border border-border/80 space-y-2.5 shadow-xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-mono uppercase text-text-subtle">Strategicity</span>
            <span className="text-xs font-bold font-tabular text-aurora">
              {strategicity == null ? "NOT SCORED" : `${strategicity.toFixed(1)}/100`}
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-primary rounded-full shadow-[0_0_8px_#00F5A0]" style={{ width: `${strategicity ?? 0}%` }}></div>
          </div>
          <p className="text-[11px] text-text-subtle leading-tight pt-1">
            Importance to Canadian critical minerals, clean baseload power, and Arctic sovereignty.
          </p>
        </div>

        {/* Trade resilience */}
        <div className="glass-card p-5 rounded-2xl border border-border/80 space-y-2.5 shadow-xl">
          <div className="flex items-center justify-between gap-2">
            <span className="text-xs font-mono uppercase text-text-subtle">Trade resilience</span>
            <span className="text-xs font-bold font-tabular text-gold">
              {tradeResilience == null ? "NOT SCORED" : `${tradeResilience.toFixed(1)}/100`}
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-gold rounded-full" style={{ width: `${tradeResilience ?? 0}%` }}></div>
          </div>
          <p className="text-[11px] text-text-subtle leading-tight pt-1">
            Canada-level logistics, trade participation, two-way flows, and high-technology exports.
          </p>
        </div>
      </div>

      {scoreDetails.length > 0 && (
        <section className="glass-card rounded-2xl border border-border/80 p-6 shadow-xl" aria-labelledby="score-lineage-title">
          <div className="flex flex-col gap-2 border-b border-borderSubtle pb-4 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 id="score-lineage-title" className="font-mono text-sm font-bold uppercase tracking-wider text-text-main">Scoring lineage</h2>
              <p className="mt-1 text-xs text-text-muted">Version, factor coverage, input hash, and resolvable evidence for every published score.</p>
            </div>
            <Link href="/methodology" className="text-xs font-bold text-aurora underline decoration-border underline-offset-4">Open formulas</Link>
          </div>
          <div className="mt-4 grid gap-3 lg:grid-cols-2">
            {scoreDetails.map((detail) => (
              <details key={detail.score_type} className="rounded-xl border border-borderSubtle bg-background/50 p-4 open:border-primary/30">
                <summary className="cursor-pointer list-none">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <span className="font-mono text-xs font-bold uppercase text-text-main">{detail.score_type.replaceAll("_", " ")}</span>
                    <span className="font-mono text-[10px] text-aurora">{detail.score_version} · {detail.coverage?.toFixed(0) ?? "—"}% covered</span>
                  </div>
                </summary>
                <p className="mt-3 text-xs leading-relaxed text-text-muted">{detail.explanation}</p>
                <div className="mt-3 grid grid-cols-2 gap-2">
                  {Object.entries(detail.factors).map(([factor, value]) => (
                    <div key={factor} className="rounded-lg border border-borderSubtle bg-surface px-3 py-2">
                      <div className="font-mono text-[9px] uppercase text-text-subtle">{factor.replaceAll("_", " ")}</div>
                      <div className="mt-1 font-mono text-xs font-bold tabular-nums text-text-main">{value.toFixed(1)}/100</div>
                    </div>
                  ))}
                </div>
                <div className="mt-3 break-all font-mono text-[9px] text-text-subtle">SHA-256 input: {detail.input_hash ?? "not available"}</div>
                <div className="mt-1 font-mono text-[9px] text-text-subtle">Evidence records: {detail.evidence_ids?.length ?? 0}</div>
              </details>
            ))}
          </div>
        </section>
      )}

      {/* Lifecycle Stage Progression Tracker */}
      <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
        <h3 className="font-bold text-xs uppercase tracking-wider text-text-main flex items-center gap-2 font-mono">
          <Clock className="h-4 w-4 text-aurora" />
          Lifecycle Stage Progression
        </h3>
        <div className="grid grid-cols-3 sm:grid-cols-9 gap-2 text-center text-[10px] font-mono">
          {stages.map((st, idx) => {
            const isCompleted = currentStageIndex >= idx;
            const isCurrent = currentStageIndex === idx;
            return (
              <div
                key={st}
                className={`p-2.5 rounded-xl border transition-colors ${
                  isCurrent
                    ? "bg-primary/20 border-primary text-aurora font-bold shadow-sm"
                    : isCompleted
                    ? "bg-surface border-borderSubtle text-text-main"
                    : "bg-surface/30 border-borderSubtle text-text-subtle opacity-40"
                }`}
              >
                <div className="truncate">{st}</div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Grid: Downstream Opportunities & Evidence Trust Profile */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Downstream Derived Opportunities */}
        <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
          <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
            <h3 className="font-bold text-xs uppercase tracking-wider text-text-main flex items-center gap-2 font-mono">
              <Layers className="h-4 w-4 text-gold" />
              Downstream Opportunities (Propagation Engine)
            </h3>
            <span className="text-[10px] font-mono text-text-subtle">CEGS Ontology</span>
          </div>

          <div className="rounded-xl border border-gold/40 bg-gold/10 p-4 text-xs leading-relaxed text-text-muted">
            No project-specific procurement opportunity is asserted by the reviewed snapshot. Derived opportunities require disclosed rules and project evidence before publication; use the source dossier beside this panel for primary-record diligence.
          </div>
        </div>

        {/* Evidence Quality & Trust Profile */}
        <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
          <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
            <h3 className="font-bold text-xs uppercase tracking-wider text-text-main flex items-center gap-2 font-mono">
              <ShieldCheck className="h-4 w-4 text-aurora" />
              Evidence Quality & Trust Profile
            </h3>
            <span className="text-[10px] font-mono text-aurora">CEGS Provenance</span>
          </div>

          <div className="space-y-3 text-xs">
            <div className="grid grid-cols-2 gap-3 text-center">
              <div className="p-3 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Linked evidence records</div>
                <div className="text-xl font-black text-aurora mt-0.5 font-tabular">{evidence.length}</div>
                <div className="text-[10px] text-text-muted">Bundled in the reviewed release</div>
              </div>
              <div className="p-3 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Tier 1 records</div>
                <div className="text-xl font-black text-text-main mt-0.5 font-tabular">{evidence.filter((item) => item.source_tier === 1).length}</div>
                <div className="text-[10px] text-aurora">Publisher-controlled sources</div>
              </div>
            </div>

            <div className="space-y-2 pt-2 border-t border-borderSubtle">
              <div className="text-[11px] font-semibold text-text-main">Sourced Evidence References:</div>
              {evidence.length > 0 ? evidence.map((item) => (
                <div key={item.id} className="p-3 rounded-xl bg-surface border border-borderSubtle space-y-1 font-mono text-[11px]">
                  <div className="flex items-center justify-between gap-3 text-text-main">
                    <span>{item.publisher}</span>
                    <span className="shrink-0 text-[10px] text-aurora font-bold">Tier {item.source_tier}</span>
                  </div>
                  {item.locator && <div className="text-[10px] text-text-muted">{item.locator}</div>}
                  <div className="text-[10px] text-text-subtle break-all">
                    SHA-256: {item.content_hash}
                  </div>
                  <a
                    href={item.source_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-aurora hover:underline text-[10px] flex items-center gap-1 pt-0.5"
                  >
                    Open source record <ExternalLink className="h-2.5 w-2.5" />
                  </a>
                </div>
              )) : (
                <p className="rounded-xl border border-borderSubtle bg-surface p-3 text-[11px] text-text-muted">
                  No evidence record was bundled for this project. The project remains visible, but no source claim is substituted.
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
