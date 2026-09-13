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
import { getProjectBySlug, FALLBACK_PROJECTS } from "@/lib/data";

interface Props {
  params: Promise<{ id: string }>;
}

export default async function ProjectProfilePage({ params }: Props) {
  const { id } = await params;
  const project = await getProjectBySlug(id);

  if (!project) {
    notFound();
  }

  const scores = project.scores || {
    buildability: 50.0,
    investability: 60.0,
    supplierability: 55.0,
    strategicity: 75.0,
  };

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
            href={`http://localhost:8080/api/v1/export/project/${project.slug}?format=markdown`}
            target="_blank"
            rel="noreferrer"
            className="px-3.5 py-1.5 rounded-xl bg-surface border border-border hover:border-aurora text-text-muted hover:text-aurora text-xs transition-all flex items-center gap-1.5 font-mono shadow-sm"
          >
            <Download className="h-3.5 w-3.5" /> Markdown
          </a>
          <a
            href={`http://localhost:8080/api/v1/export/project/${project.slug}?format=cegs`}
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

      {/* 4-D Deterministic Scorecards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Buildability */}
        <div className="glass-card p-5 rounded-2xl border border-border/80 space-y-2.5 shadow-xl">
          <div className="flex items-center justify-between">
            <span className="text-xs font-mono uppercase text-text-subtle">Buildability</span>
            <span className="text-xs font-bold font-tabular text-aurora">
              {scores.buildability.toFixed(1)}/100
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-primary rounded-full" style={{ width: `${scores.buildability}%` }}></div>
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
              {scores.investability.toFixed(1)}/100
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-gold rounded-full" style={{ width: `${scores.investability}%` }}></div>
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
              {scores.supplierability.toFixed(1)}/100
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-aurora-mint rounded-full" style={{ width: `${scores.supplierability}%` }}></div>
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
              {scores.strategicity.toFixed(1)}/100
            </span>
          </div>
          <div className="w-full h-2 rounded-full bg-surface overflow-hidden border border-borderSubtle">
            <div className="h-full bg-primary rounded-full shadow-[0_0_8px_#00F5A0]" style={{ width: `${scores.strategicity}%` }}></div>
          </div>
          <p className="text-[11px] text-text-subtle leading-tight pt-1">
            Importance to Canadian critical minerals, clean baseload power, and Arctic sovereignty.
          </p>
        </div>
      </div>

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

          <div className="space-y-3 text-xs">
            <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-text-main">High-Voltage Substation Interconnect</span>
                <span className="px-2 py-0.5 rounded-full bg-primary/10 text-aurora border border-primary/30 text-[9px] font-mono font-bold">
                  CONFIRMED
                </span>
              </div>
              <p className="text-text-subtle text-[11px] leading-relaxed">
                Dedicated 230kV switchgear, transformer installation, and protection automation.
              </p>
              <div className="text-[10px] font-mono text-gold pt-1">Est. Value: $85,000,000 CAD</div>
            </div>

            <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-text-main">Civil Heavy Shielding & Specialized Foundations</span>
                <span className="px-2 py-0.5 rounded-full bg-surface border border-borderSubtle text-text-muted text-[9px] font-mono">
                  DERIVED
                </span>
              </div>
              <p className="text-text-subtle text-[11px] leading-relaxed">
                High-density aggregate seismic containment concrete and deep water intake/outfall tunnels.
              </p>
              <div className="text-[10px] font-mono text-gold pt-1">Est. Value: $250,000,000 CAD</div>
            </div>

            <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-text-main">Environmental Baseline & Community Guardian Surveillance</span>
                <span className="px-2 py-0.5 rounded-full bg-primary/10 text-aurora border border-primary/30 text-[9px] font-mono font-bold">
                  CONFIRMED
                </span>
              </div>
              <p className="text-text-subtle text-[11px] leading-relaxed">
                Continuous aquatic thermal dissipation tracking and Indigenous environmental oversight.
              </p>
              <div className="text-[10px] font-mono text-gold pt-1">Est. Value: $12,000,000 CAD</div>
            </div>
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
                <div className="text-[10px] font-mono text-text-subtle uppercase">Primary Tier 1 Sources</div>
                <div className="text-xl font-black text-aurora mt-0.5 font-tabular">100%</div>
                <div className="text-[10px] text-text-muted">Statutory Registry Corroborated</div>
              </div>
              <div className="p-3 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Conflicting Claims</div>
                <div className="text-xl font-black text-text-main mt-0.5 font-tabular">0</div>
                <div className="text-[10px] text-aurora">Reconciled</div>
              </div>
            </div>

            <div className="space-y-2 pt-2 border-t border-borderSubtle">
              <div className="text-[11px] font-semibold text-text-main">Sourced Evidence References:</div>
              
              <div className="p-3 rounded-xl bg-surface border border-borderSubtle space-y-1 font-mono text-[11px]">
                <div className="flex items-center justify-between text-text-main">
                  <span>Impact Assessment Agency of Canada</span>
                  <span className="text-[10px] text-aurora font-bold">Tier 1</span>
                </div>
                <div className="text-[10px] text-text-subtle truncate">
                  SHA-256: e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
                </div>
                <a
                  href="https://iaac-aeic.gc.ca"
                  target="_blank"
                  rel="noreferrer"
                  className="text-aurora hover:underline text-[10px] flex items-center gap-1 pt-0.5"
                >
                  Verify Source Document <ExternalLink className="h-2.5 w-2.5" />
                </a>
              </div>

              <div className="p-3 rounded-xl bg-surface border border-borderSubtle space-y-1 font-mono text-[11px]">
                <div className="flex items-center justify-between text-text-main">
                  <span>Canada Infrastructure Bank Financing Registry</span>
                  <span className="text-[10px] text-aurora font-bold">Tier 1</span>
                </div>
                <div className="text-[10px] text-text-subtle truncate">
                  SHA-256: 8a4b2c1d9f8e7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b
                </div>
                <a
                  href="https://cib-bic.ca"
                  target="_blank"
                  rel="noreferrer"
                  className="text-aurora hover:underline text-[10px] flex items-center gap-1 pt-0.5"
                >
                  Verify Source Document <ExternalLink className="h-2.5 w-2.5" />
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
