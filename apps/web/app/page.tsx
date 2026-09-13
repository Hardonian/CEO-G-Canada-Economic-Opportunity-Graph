import Link from "next/link";
import { 
  TrendingUp, 
  TrendingDown, 
  Activity, 
  Building2, 
  Coins, 
  FileCheck2, 
  ShieldAlert, 
  ArrowUpRight, 
  ChevronRight, 
  Flame, 
  Cpu, 
  Zap, 
  Pickaxe, 
  Anchor, 
  Compass, 
  Shield 
} from "lucide-react";
import { getRadarData, getProjects } from "@/lib/data";

export default async function HomePage() {
  const radarData = await getRadarData();
  const projects = await getProjects();
  const stats = radarData.stats;

  const accelerating = projects.filter((p) => (p.scores?.buildability || 0) >= 50);
  const earlyStage = projects.filter((p) => p.current_stage === "PERMITTING" || p.current_stage === "FEASIBILITY");

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* Flagship Hero Header: Canada Capital Radar */}
      <div className="border-b border-border pb-8">
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-2 px-2.5 py-1 rounded bg-card border border-border text-xs font-mono text-accent-cyan mb-3">
              <Activity className="h-3.5 w-3.5 text-accent-cyan animate-pulse" />
              FLAGSHIP MACRO RADAR — REAL-TIME NATIONAL INTELLIGENCE
            </div>
            <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-text-main">
              Canada Capital Radar
            </h1>
            <p className="text-sm text-text-muted mt-2 max-w-3xl leading-relaxed">
              Tracking <span className="text-text-main font-semibold">${(stats.total_capex_cad / 1e9).toFixed(2)} Billion CAD</span> across 
              Canada's major projects cycle. Follow capital commitments, regulatory approvals, procurement tenders, and downstream opportunities before construction begins.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/projects"
              className="px-4 py-2 rounded bg-primary text-white text-xs font-semibold hover:bg-primary-hover transition-colors shadow-sm flex items-center gap-1.5"
            >
              Explore All Projects <ArrowUpRight className="h-4 w-4" />
            </Link>
            <Link
              href="/cegs"
              className="px-3 py-2 rounded bg-surface border border-border text-text-muted text-xs font-medium hover:text-text-main hover:bg-card transition-colors font-mono"
            >
              CEGS 0.1 Spec
            </Link>
          </div>
        </div>

        {/* Above-the-Fold Macro Metrics Bar */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mt-8">
          <div className="bg-card p-3.5 rounded border border-border">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Tracked CAPEX</div>
            <div className="text-xl font-bold font-tabular text-text-main mt-1">
              ${(stats.total_capex_cad / 1e9).toFixed(2)}B
            </div>
            <div className="text-[10px] text-accent-green mt-1 flex items-center gap-0.5">
              <span>Verified CAD Base</span>
            </div>
          </div>

          <div className="bg-card p-3.5 rounded border border-border">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Tracked Projects</div>
            <div className="text-xl font-bold font-tabular text-text-main mt-1">
              {stats.total_projects} Major Assets
            </div>
            <div className="text-[10px] text-text-muted mt-1">
              Across 5 Key Sectors
            </div>
          </div>

          <div className="bg-card p-3.5 rounded border border-border">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Moving This Week</div>
            <div className="text-xl font-bold font-tabular text-accent-cyan mt-1">
              ${(stats.capital_moving_week_cad / 1e6).toFixed(0)}M CAD
            </div>
            <div className="text-[10px] text-accent-cyan mt-1">
              Active Milestone Flow
            </div>
          </div>

          <div className="bg-card p-3.5 rounded border border-border">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Accelerating</div>
            <div className="text-xl font-bold font-tabular text-accent-green mt-1 flex items-center gap-1.5">
              <TrendingUp className="h-4 w-4" />
              {accelerating.length} Projects
            </div>
            <div className="text-[10px] text-text-muted mt-1">
              Buildability &gt; 50
            </div>
          </div>

          <div className="bg-card p-3.5 rounded border border-border">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Open Tenders</div>
            <div className="text-xl font-bold font-tabular text-accent-gold mt-1">
              {stats.active_procurements_count} Notices
            </div>
            <div className="text-[10px] text-text-muted mt-1">
              CanadaBuys + DND
            </div>
          </div>

          <div className="bg-card p-3.5 rounded border border-border">
            <div className="text-[11px] font-mono text-text-subtle uppercase">CEGS Standard</div>
            <div className="text-xl font-bold font-tabular text-text-main mt-1 font-mono">
              v0.1 Active
            </div>
            <div className="text-[10px] text-accent-green mt-1">
              100% Cryptographic Sourced
            </div>
          </div>
        </div>
      </div>

      {/* Main Grid: Live Signals & Accelerating Projects */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column (2 Cols): Projects Accelerating Towards Construction */}
        <div className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Flame className="h-4 w-4 text-primary" />
              <h2 className="text-base font-bold text-text-main tracking-tight">
                Projects Accelerating Towards Execution
              </h2>
            </div>
            <Link href="/projects" className="text-xs text-accent-cyan hover:underline flex items-center gap-0.5 font-medium">
              View ranking <ChevronRight className="h-3.5 w-3.5" />
            </Link>
          </div>

          <div className="space-y-3">
            {projects.map((proj) => {
              const bScore = proj.scores?.buildability || 0;
              const iScore = proj.scores?.investability || 0;
              return (
                <Link
                  key={proj.id}
                  href={`/projects/${proj.slug}`}
                  className="block p-4 rounded bg-card border border-border hover:border-text-subtle hover:bg-cardHover transition-all group shadow-sm"
                >
                  <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-muted">
                          {proj.province}
                        </span>
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-surface border border-borderSubtle text-accent-cyan">
                          {proj.sector}
                        </span>
                        <span className="text-[11px] font-mono px-2 py-0.5 rounded bg-primary/10 border border-primary/30 text-primary font-semibold">
                          {proj.current_stage}
                        </span>
                      </div>
                      <h3 className="font-semibold text-text-main group-hover:text-accent-cyan transition-colors text-sm pt-1">
                        {proj.name}
                      </h3>
                      <p className="text-xs text-text-subtle line-clamp-2 leading-relaxed">
                        {proj.summary}
                      </p>
                    </div>

                    {/* Score Badges */}
                    <div className="flex sm:flex-col items-end justify-between sm:justify-start gap-2 shrink-0 pt-2 sm:pt-0">
                      <div className="text-right">
                        <div className="text-[10px] font-mono text-text-subtle uppercase">Reported CAPEX</div>
                        <div className="text-sm font-bold font-tabular text-text-main">
                          ${(proj.capex_cad / 1e9).toFixed(2)}B CAD
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="text-right">
                          <div className="text-[9px] font-mono text-text-subtle uppercase">Buildability</div>
                          <div className="text-xs font-bold font-tabular text-accent-green">
                            {bScore.toFixed(1)}/100
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="text-[9px] font-mono text-text-subtle uppercase">Investability</div>
                          <div className="text-xs font-bold font-tabular text-accent-cyan">
                            {iScore.toFixed(1)}/100
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </Link>
              );
            })}
          </div>
        </div>

        {/* Right Column: Real-Time Economic Signals & Sector Breakdown */}
        <div className="space-y-6">
          {/* Recent Capital Signals */}
          <div className="bg-card p-4 rounded border border-border space-y-4">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-2.5">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-accent-cyan" />
                <h3 className="font-bold text-xs uppercase tracking-wider text-text-main">
                  Signals Detected (Last 30 Days)
                </h3>
              </div>
              <span className="text-[10px] font-mono text-accent-green font-medium">LIVE</span>
            </div>

            <div className="space-y-3">
              {radarData.recent_signals?.map((sig: any) => (
                <div key={sig.id} className="text-xs border-l-2 border-accent-cyan pl-3 py-0.5 space-y-1">
                  <div className="flex items-center justify-between text-[10px] font-mono text-text-subtle">
                    <span className="text-accent-cyan font-semibold">{sig.type}</span>
                    <span>{new Date(sig.timestamp).toLocaleDateString("en-CA")}</span>
                  </div>
                  <div className="font-medium text-text-main text-xs">
                    {sig.project_name}
                  </div>
                  <p className="text-[11px] text-text-subtle leading-relaxed">
                    {sig.description}
                  </p>
                </div>
              ))}
            </div>
          </div>

          {/* Sector Capital Allocation */}
          <div className="bg-card p-4 rounded border border-border space-y-3">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-2.5">
              <h3 className="font-bold text-xs uppercase tracking-wider text-text-main">
                CAPEX by Strategic Sector
              </h3>
              <span className="text-[10px] font-mono text-text-subtle">$ CAD</span>
            </div>

            <div className="space-y-2.5 text-xs">
              {Object.entries(stats.sector_breakdown).map(([sector, rawCapex]) => {
                const capex = Number(rawCapex);
                const pct = stats.total_capex_cad > 0 ? (capex / stats.total_capex_cad) * 100 : 0;
                return (
                  <div key={sector} className="space-y-1">
                    <div className="flex items-center justify-between text-[11px]">
                      <span className="text-text-muted">{sector}</span>
                      <span className="font-bold font-tabular text-text-main">
                        ${(capex / 1e9).toFixed(2)}B ({pct.toFixed(0)}%)
                      </span>
                    </div>
                    <div className="w-full h-1.5 rounded-full bg-surface overflow-hidden">
                      <div className="h-full bg-accent-cyan rounded-full" style={{ width: `${pct}%` }}></div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* CEGS Standard Reference Callout */}
          <div className="p-4 rounded border border-borderSubtle bg-surface text-xs space-y-2">
            <div className="flex items-center gap-1.5 font-bold text-text-main text-xs">
              <FileCheck2 className="h-4 w-4 text-accent-green" />
              <span>Reference Implementation of CEGS</span>
            </div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Every data point above conforms to the <strong className="text-text-muted">Canada Economic Graph Schema (CEGS 0.1)</strong>. You can export the entire graph in open machine-readable format.
            </p>
            <div className="pt-1">
              <Link href="/cegs" className="text-accent-cyan hover:underline text-xs inline-flex items-center gap-1">
                Inspect CEGS Schema & Examples →
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
