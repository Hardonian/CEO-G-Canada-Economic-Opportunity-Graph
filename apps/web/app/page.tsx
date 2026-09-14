import Link from "next/link";
import { 
  TrendingUp, 
  Activity, 
  FileCheck2, 
  ArrowUpRight, 
  Flame, 
  Zap, 
  Sparkles,
  ShieldCheck,
  Briefcase
} from "lucide-react";
import { getRadarData, getProjects, SNAPSHOT_MANIFEST } from "@/lib/data";
import DynamicRadarExplorer from "@/components/DynamicRadarExplorer";

export default async function HomePage() {
  const radarData = await getRadarData();
  const projects = await getProjects();
  const stats = radarData.stats;

  const accelerating = projects.filter((p) => (p.scores?.buildability || 0) >= 50);
  const sectorCount = new Set(projects.map((project) => project.sector)).size;
  const snapshotMode = radarData.source_mode !== "LIVE_UPSTREAM_API";

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* Flagship Hero Header: Canada Capital Radar */}
      <div className="relative rounded-2xl p-6 sm:p-8 bg-gradient-to-b from-card/80 to-surface/90 border border-border/80 shadow-2xl shadow-black/50 overflow-hidden">
        {/* Subtle Ambient Northern Lights / Aurora Backlight */}
        <div className="absolute top-0 right-0 -mt-12 -mr-12 w-96 h-96 rounded-full bg-aurora/10 blur-3xl pointer-events-none"></div>
        <div className="absolute bottom-0 left-1/3 -mb-16 w-80 h-80 rounded-full bg-gold/10 blur-3xl pointer-events-none"></div>

        <div className="relative z-10 flex flex-col md:flex-row md:items-end justify-between gap-6">
          <div className="space-y-3">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#050B08] border border-primary/40 text-xs font-mono text-aurora shadow-sm">
              <span className={`h-2 w-2 rounded-full bg-aurora shadow-[0_0_8px_#00F5A0] ${snapshotMode ? "" : "animate-pulse"}`}></span>
              FLAGSHIP MACRO RADAR — CANADIAN ECONOMIC INTELLIGENCE
            </div>
            <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-text-main">
              Canada <span className="text-aurora">Capital Radar</span>
            </h1>
            <p className="text-sm text-text-muted max-w-3xl leading-relaxed">
              Tracking <span className="text-aurora font-bold">${(stats.total_capex_cad / 1e9).toFixed(2)} Billion CAD</span> in 
              Canada&apos;s major projects cycle. Source-linked records capture reported capital, selected milestones, and deterministic planning scores.
            </p>
          </div>

          <div className="mt-5 inline-flex flex-wrap items-center gap-x-3 gap-y-1 rounded-lg border border-primary/30 bg-background/80 px-3 py-2 font-mono text-[10px] text-text-muted">
            <span className="font-bold text-aurora">{snapshotMode ? "REVIEWED SNAPSHOT" : "LIVE UPSTREAM API"}</span>
            <span>{projects.length} projects · {radarData.source_mode === "BUNDLED_REVIEWED_SNAPSHOT" ? `${SNAPSHOT_MANIFEST.record_counts.evidence} evidence records` : "upstream response"}</span>
            <span>Data through {new Intl.DateTimeFormat("en-CA", { dateStyle: "medium", timeZone: "UTC" }).format(new Date(radarData.generated_at ?? "2026-09-13T00:00:00Z"))} UTC</span>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            <Link
              href="/projects"
              className="px-4 py-2.5 rounded-xl bg-primary text-[#050B08] font-bold text-xs hover:bg-aurora-mint transition-all shadow-lg shadow-emerald-950/50 flex items-center gap-1.5"
            >
              Explore All Projects <ArrowUpRight className="h-4 w-4" />
            </Link>
            <Link
              href="/cegs"
              className="px-3.5 py-2.5 rounded-xl bg-card border border-border text-text-muted text-xs font-medium hover:text-aurora hover:border-primary/50 transition-all font-mono"
            >
              CEGS 0.1 Standard
            </Link>
          </div>
        </div>

        {/* Above-the-Fold Macro Metrics Bar */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 mt-8 relative z-10">
          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Tracked CAPEX</div>
            <div className="text-2xl font-black font-tabular text-aurora mt-1">
              ${(stats.total_capex_cad / 1e9).toFixed(2)}B
            </div>
            <div className="text-[10px] text-text-muted mt-1 flex items-center gap-1">
              <ShieldCheck className="h-3 w-3 text-aurora" />
              Source-reported CAD base
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Tracked Assets</div>
            <div className="text-2xl font-black font-tabular text-text-main mt-1">
              {stats.total_projects} <span className="text-xs text-text-subtle font-normal">Projects</span>
            </div>
            <div className="text-[10px] text-text-muted mt-1">
              Across {sectorCount} tracked sectors
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Moving This Week</div>
            <div className="text-2xl font-black font-tabular text-gold mt-1">
              ${(stats.capital_moving_week_cad / 1e6).toFixed(0)}M
            </div>
            <div className="text-[10px] text-gold mt-1">
              Active Milestone Flow
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Accelerating</div>
            <div className="text-2xl font-black font-tabular text-aurora mt-1 flex items-center gap-1.5">
              <TrendingUp className="h-5 w-5 text-aurora" />
              {accelerating.length}
            </div>
            <div className="text-[10px] text-text-muted mt-1">
              Buildability &gt; 50
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[11px] font-mono text-text-subtle uppercase">Tender Records</div>
            <div className="text-2xl font-black font-tabular text-gold mt-1">
              {stats.active_procurements_count}
            </div>
            <div className="text-[10px] text-text-muted mt-1">
              No live connector
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[11px] font-mono text-text-subtle uppercase">CEGS Standard</div>
            <div className="text-2xl font-black font-tabular text-aurora-mint mt-1 font-mono">
              v0.1
            </div>
            <div className="text-[10px] text-aurora mt-1 font-mono">
              Checksummed release
            </div>
          </div>
        </div>
      </div>

      {/* Main Grid: Interactive Dynamic Radar Explorer + Intelligence Sidebar */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Left Column (2 Cols): Dynamic Interactive Radar Explorer */}
        <div className="lg:col-span-2 space-y-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-aurora" />
              <h2 className="text-base font-bold text-text-main tracking-tight">
                Strategic Economic Opportunities & Progress Radar
              </h2>
            </div>
            <span className="text-xs text-text-subtle font-mono">
              Interactive snapshot filters
            </span>
          </div>

          {/* Mount the Dynamic Interactive Radar Component */}
          <DynamicRadarExplorer initialProjects={projects} />
        </div>

        {/* Right Column: Real-Time Economic Signals & Sector Breakdown */}
        <div className="space-y-6">
          {/* Recent Capital Signals */}
          <div className="glass-panel p-5 rounded-xl border border-border/80 space-y-4 shadow-lg">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-aurora" />
                <h3 className="font-bold text-xs uppercase tracking-wider text-text-main">
                  National Signals (Last 30 Days)
                </h3>
              </div>
              <span className="text-[10px] font-mono text-aurora font-semibold px-2 py-0.5 rounded bg-primary/20 border border-primary/30">
                {snapshotMode ? "NO LIVE FEED" : "LIVE"}
              </span>
            </div>

            <div className="space-y-3">
              {radarData.recent_signals?.map((sig: { id: string; type: string; timestamp: string; project_name: string; description: string }) => (
                <div key={sig.id} className="text-xs border-l-2 border-aurora pl-3 py-1 space-y-1 bg-surface/50 rounded-r-lg">
                  <div className="flex items-center justify-between text-[10px] font-mono text-text-subtle">
                    <span className="text-aurora font-semibold">{sig.type}</span>
                    <span>{new Date(sig.timestamp).toLocaleDateString("en-CA")}</span>
                  </div>
                  <div className="font-medium text-text-main text-xs">
                    {sig.project_name}
                  </div>
                  <p className="text-[11px] text-text-muted leading-relaxed">
                    {sig.description}
                  </p>
                </div>
              ))}
              {!radarData.recent_signals?.length && (
                <p className="rounded-lg border border-borderSubtle bg-surface/50 p-3 text-[11px] leading-relaxed text-text-muted">
                  No current signal feed is configured. The reviewed project and evidence snapshot remains available; no live events are implied.
                </p>
              )}
            </div>
          </div>

          {/* Sector Capital Allocation */}
          <div className="glass-panel p-5 rounded-xl border border-border/80 space-y-4 shadow-lg">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <h3 className="font-bold text-xs uppercase tracking-wider text-text-main">
                CAPEX by Strategic Sector
              </h3>
              <span className="text-[10px] font-mono text-gold font-semibold">$ CAD Total</span>
            </div>

            <div className="space-y-3 text-xs">
              {Object.entries(stats.sector_breakdown).map(([sector, rawCapex], idx) => {
                const capex = Number(rawCapex);
                const pct = stats.total_capex_cad > 0 ? (capex / stats.total_capex_cad) * 100 : 0;
                
                // Cycle distinctive Aurora, Gold, Mint, and Copper palettes
                const barColors = [
                  "bg-gradient-to-r from-emerald-500 to-teal-400",
                  "bg-gradient-to-r from-amber-500 to-yellow-400",
                  "bg-gradient-to-r from-teal-400 to-emerald-300",
                  "bg-gradient-to-r from-orange-500 to-amber-400",
                  "bg-gradient-to-r from-emerald-400 to-green-300",
                ];
                const activeColor = barColors[idx % barColors.length];

                return (
                  <div key={sector} className="space-y-1.5">
                    <div className="flex items-center justify-between text-[11px]">
                      <span className="text-text-muted font-medium">{sector}</span>
                      <span className="font-bold font-tabular text-text-main">
                        ${(capex / 1e9).toFixed(2)}B ({pct.toFixed(0)}%)
                      </span>
                    </div>
                    <div className="w-full h-2 rounded-full bg-background overflow-hidden border border-borderSubtle">
                      <div className={`h-full ${activeColor} rounded-full transition-all duration-500`} style={{ width: `${pct}%` }}></div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* CEGS Standard Reference Callout */}
          <div className="p-5 rounded-xl border border-primary/30 bg-gradient-to-br from-card to-surface text-xs space-y-3 shadow-lg">
            <div className="flex items-center gap-2 font-bold text-text-main text-xs">
              <FileCheck2 className="h-4 w-4 text-aurora" />
              <span>Reference Implementation of CEGS</span>
            </div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Every data point in this graph adheres to the <strong className="text-text-main">Canada Economic Graph Schema (CEGS 0.1)</strong>. Inspect the schemas, test conformance against validator rules, or download open snapshots.
            </p>
            <div className="pt-1">
              <Link 
                href="/cegs" 
                className="text-aurora hover:underline text-xs inline-flex items-center gap-1 font-semibold"
              >
                Inspect CEGS Schema & Examples →
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
