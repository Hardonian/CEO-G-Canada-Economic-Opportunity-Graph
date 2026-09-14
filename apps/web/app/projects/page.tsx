"use client";

import { useState } from "react";
import Link from "next/link";
import { 
  Filter, 
  Search, 
  ArrowUpDown, 
  ChevronRight, 
  Layers, 
  MapPin, 
  Pickaxe, 
  Zap,
  LayoutGrid,
  List,
  ShieldCheck,
  TrendingUp,
  Radio
} from "lucide-react";
import { FALLBACK_PROJECTS } from "@/lib/data";

export default function ProjectsPage() {
  const [search, setSearch] = useState("");
  const [selectedSector, setSelectedSector] = useState("ALL");
  const [selectedProvince, setSelectedProvince] = useState("ALL");
  const [sortBy, setSortBy] = useState<"capex" | "buildability" | "investability" | "name">("capex");
  const [viewMode, setViewMode] = useState<"table" | "grid">("grid");

  const sectors = ["ALL", ...Array.from(new Set(FALLBACK_PROJECTS.map((project) => project.sector))).sort()];
  const provinces = ["ALL", ...Array.from(new Set(FALLBACK_PROJECTS.map((project) => project.province))).sort()];

  const filtered = FALLBACK_PROJECTS.filter((p) => {
    if (selectedSector !== "ALL" && p.sector !== selectedSector) return false;
    if (selectedProvince !== "ALL" && p.province !== selectedProvince) return false;
    if (search.trim() !== "") {
      const q = search.toLowerCase();
      const match = 
        p.name.toLowerCase().includes(q) || 
        p.summary.toLowerCase().includes(q) || 
        p.subsector.toLowerCase().includes(q) ||
        p.location_name.toLowerCase().includes(q);
      if (!match) return false;
    }
    return true;
  }).sort((a, b) => {
    if (sortBy === "capex") return b.capex_cad - a.capex_cad;
    if (sortBy === "buildability") return (b.scores?.buildability || 0) - (a.scores?.buildability || 0);
    if (sortBy === "investability") return (b.scores?.investability || 0) - (a.scores?.investability || 0);
    return a.name.localeCompare(b.name);
  });

  const totalCapex = filtered.reduce((acc, p) => acc + p.capex_cad, 0);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Title & Summary */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-border/80 pb-6">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
            <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
            MAJOR PROJECTS CYCLE & CAPITAL TRACKING
          </div>
          <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
            Canadian <span className="text-aurora">Major Projects Directory</span>
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-1.5 max-w-3xl leading-relaxed">
            Showing {filtered.length} source-linked project records representing ${(totalCapex / 1e9).toFixed(2)}B CAD in reported capital. 
            Scores are deterministic planning indicators, not forecasts or verification of project outcomes.
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* View Mode Toggle */}
          <div className="inline-flex rounded-xl bg-surface p-1 border border-borderSubtle">
            <button
              onClick={() => setViewMode("grid")}
              className={`p-1.5 rounded-lg transition-all ${
                viewMode === "grid" ? "bg-card text-aurora shadow-sm" : "text-text-subtle hover:text-text-main"
              }`}
              title="Card Grid View"
            >
              <LayoutGrid className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode("table")}
              className={`p-1.5 rounded-lg transition-all ${
                viewMode === "table" ? "bg-card text-aurora shadow-sm" : "text-text-subtle hover:text-text-main"
              }`}
              title="Table View"
            >
              <List className="h-4 w-4" />
            </button>
          </div>

          <div className="flex items-center gap-2 text-xs font-mono">
            <span className="text-text-subtle hidden sm:inline">Sort:</span>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="bg-card border border-border rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-aurora font-mono"
            >
              <option value="capex">Highest CAPEX ($ CAD)</option>
              <option value="buildability">Highest Buildability</option>
              <option value="investability">Highest Investability</option>
              <option value="name">Alphabetical Name</option>
            </select>
          </div>
        </div>
      </div>

      {/* Filters Bar */}
      <div className="glass-panel p-4 rounded-2xl border border-border/80 space-y-3 shadow-lg">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {/* Search Input */}
          <div className="relative">
            <Search className="h-4 w-4 absolute left-3 top-2.5 text-text-subtle" />
            <input
              type="text"
              placeholder="Search projects, proponents, minerals..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main placeholder:text-text-subtle focus:outline-none focus:border-aurora transition-colors font-mono"
            />
          </div>

          {/* Sector Selector */}
          <div>
            <select
              value={selectedSector}
              onChange={(e) => setSelectedSector(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main focus:outline-none focus:border-aurora transition-colors font-mono"
            >
              {sectors.map((s) => (
                <option key={s} value={s}>
                  Sector: {s === "ALL" ? "All Sectors" : s}
                </option>
              ))}
            </select>
          </div>

          {/* Province Selector */}
          <div>
            <select
              value={selectedProvince}
              onChange={(e) => setSelectedProvince(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main focus:outline-none focus:border-aurora transition-colors font-mono"
            >
              {provinces.map((pr) => (
                <option key={pr} value={pr}>
                  Province/Territory: {pr === "ALL" ? "All Regions" : pr}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Card Grid View */}
      {viewMode === "grid" ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filtered.map((p) => {
            const bScore = p.scores?.buildability || 0;
            const iScore = p.scores?.investability || 0;
            return (
              <div
                key={p.id}
                className="glass-card p-6 rounded-2xl border border-border/80 flex flex-col justify-between space-y-4 shadow-xl hover:border-primary/40 transition-all group"
              >
                <div className="space-y-3">
                  <div className="flex items-center justify-between text-xs font-mono">
                    <span className="px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-aurora font-bold">
                      {p.province}
                    </span>
                    <span className="px-2 py-0.5 rounded-md bg-gold/10 border border-gold/30 text-gold font-semibold">
                      {p.current_stage}
                    </span>
                  </div>

                  <Link href={`/projects/${p.slug}`} className="block">
                    <h2 className="font-bold text-text-main group-hover:text-aurora text-base leading-snug transition-colors">
                      {p.name}
                    </h2>
                  </Link>

                  <div className="text-xs text-text-subtle flex items-center gap-1.5">
                    <MapPin className="h-3.5 w-3.5 text-aurora shrink-0" />
                    <span className="truncate">{p.location_name}</span>
                  </div>

                  <p className="text-xs text-text-muted line-clamp-3 leading-relaxed">
                    {p.summary}
                  </p>
                </div>

                <div className="space-y-3 pt-3 border-t border-borderSubtle">
                  <div className="flex items-center justify-between text-xs font-mono">
                    <span className="text-text-subtle uppercase text-[10px]">Reported CAPEX:</span>
                    <span className="font-black text-base text-text-main font-tabular">
                      ${(p.capex_cad / 1e9).toFixed(2)}B CAD
                    </span>
                  </div>

                  {/* Buildability & Investability Bars */}
                  <div className="grid grid-cols-2 gap-2 text-[11px] font-mono">
                    <div className="p-2 rounded-xl bg-surface border border-borderSubtle">
                      <div className="text-[9px] text-text-subtle uppercase">Buildability</div>
                      <div className="font-bold text-aurora mt-0.5">{bScore.toFixed(0)}/100</div>
                    </div>
                    <div className="p-2 rounded-xl bg-surface border border-borderSubtle">
                      <div className="text-[9px] text-text-subtle uppercase">Investability</div>
                      <div className="font-bold text-gold mt-0.5">{iScore.toFixed(0)}/100</div>
                    </div>
                  </div>

                  <Link
                    href={`/projects/${p.slug}`}
                    className="w-full py-2.5 rounded-xl bg-surface hover:bg-card border border-border hover:border-aurora text-aurora text-xs font-bold transition-all flex items-center justify-center gap-1.5 shadow-sm"
                  >
                    Open Full Dossier <ChevronRight className="h-4 w-4" />
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      ) : (
        /* Results Table View */
        <div className="glass-card rounded-2xl border border-border/80 overflow-hidden shadow-2xl">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-[#08130E] text-text-muted font-mono uppercase text-[10px] border-b border-border/80">
                <tr>
                  <th className="px-5 py-3.5">Project & Location</th>
                  <th className="px-4 py-3.5">Sector</th>
                  <th className="px-4 py-3.5">Stage</th>
                  <th className="px-4 py-3.5 text-right">CAPEX (CAD)</th>
                  <th className="px-4 py-3.5 text-right">Buildability</th>
                  <th className="px-4 py-3.5 text-right">Investability</th>
                  <th className="px-5 py-3.5 text-right">Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-borderSubtle font-mono">
                {filtered.map((p) => {
                  const bScore = p.scores?.buildability || 0;
                  const iScore = p.scores?.investability || 0;
                  return (
                    <tr key={p.id} className="hover:bg-surface/50 transition-colors group">
                      <td className="px-5 py-4 font-sans">
                        <Link href={`/projects/${p.slug}`} className="font-semibold text-text-main group-hover:text-aurora text-sm block transition-colors">
                          {p.name}
                        </Link>
                        <div className="text-[11px] font-mono text-text-subtle flex items-center gap-1.5 mt-0.5">
                          <MapPin className="h-3 w-3 text-aurora" />
                          <span>{p.location_name} ({p.province})</span>
                          <span>•</span>
                          <span className="text-text-muted">{p.subsector}</span>
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <span className="inline-block px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-[11px] text-aurora font-mono">
                          {p.sector}
                        </span>
                      </td>
                      <td className="px-4 py-4">
                        <span className="inline-block px-2.5 py-0.5 rounded-md bg-gold/10 border border-gold/30 text-[10px] text-gold font-mono font-semibold">
                          {p.current_stage}
                        </span>
                      </td>
                      <td className="px-4 py-4 text-right font-bold font-tabular text-text-main">
                        ${(p.capex_cad / 1e9).toFixed(2)}B
                      </td>
                      <td className="px-4 py-4 text-right">
                        <span className="font-bold font-tabular text-aurora text-xs">
                          {bScore.toFixed(1)}
                        </span>
                        <span className="text-text-subtle text-[10px]">/100</span>
                      </td>
                      <td className="px-4 py-4 text-right">
                        <span className="font-bold font-tabular text-gold text-xs">
                          {iScore.toFixed(1)}
                        </span>
                        <span className="text-text-subtle text-[10px]">/100</span>
                      </td>
                      <td className="px-5 py-4 text-right">
                        <Link
                          href={`/projects/${p.slug}`}
                          className="inline-flex items-center gap-1 px-3 py-1.5 rounded-xl bg-surface border border-border hover:border-aurora text-aurora text-[11px] transition-all font-medium shadow-sm"
                        >
                          Profile <ChevronRight className="h-3.5 w-3.5" />
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
