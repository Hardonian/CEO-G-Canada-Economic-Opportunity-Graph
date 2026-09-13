"use client";

import { useState } from "react";
import Link from "next/link";
import { Filter, Search, ArrowUpDown, ChevronRight, Layers, MapPin, Pickaxe, Zap } from "lucide-react";
import { FALLBACK_PROJECTS } from "@/lib/data";

export default function ProjectsPage() {
  const [search, setSearch] = useState("");
  const [selectedSector, setSelectedSector] = useState("ALL");
  const [selectedProvince, setSelectedProvince] = useState("ALL");
  const [sortBy, setSortBy] = useState<"capex" | "buildability" | "investability" | "name">("capex");

  const sectors = [
    "ALL",
    "Nuclear & Clean Power",
    "Critical Minerals",
    "Clean Energy & Grid",
    "Transportation & Ports",
    "Defence & Arctic",
  ];

  const provinces = ["ALL", "ON", "QC", "NU"];

  const filtered = FALLBACK_PROJECTS.filter((p) => {
    if (selectedSector !== "ALL" && p.sector !== selectedSector) return false;
    if (selectedProvince !== "ALL" && p.province !== selectedProvince) return false;
    if (search.trim() !== "") {
      const q = search.toLowerCase();
      const match = p.name.toLowerCase().includes(q) || p.summary.toLowerCase().includes(q) || p.subsector.toLowerCase().includes(q);
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
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Title & Summary */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-border pb-6">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
            Canadian Major Projects Directory
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-1">
            Tracking {filtered.length} verified projects representing ${(totalCapex / 1e9).toFixed(2)}B CAD in capital investment.
          </p>
        </div>
        <div className="flex items-center gap-2 text-xs font-mono">
          <span className="text-text-subtle">Sort by:</span>
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="bg-card border border-border rounded px-2 py-1 text-text-main text-xs focus:outline-none focus:border-accent-cyan"
          >
            <option value="capex">Highest CAPEX ($ CAD)</option>
            <option value="buildability">Highest Buildability</option>
            <option value="investability">Highest Investability</option>
            <option value="name">Alphabetical Name</option>
          </select>
        </div>
      </div>

      {/* Filters Bar */}
      <div className="glass-panel p-4 rounded-xl border border-border/80 space-y-3 shadow-lg">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {/* Search Input */}
          <div className="relative">
            <Search className="h-4 w-4 absolute left-3 top-2.5 text-text-subtle" />
            <input
              type="text"
              placeholder="Search projects, proponents, minerals..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-3 py-1.5 rounded-lg bg-surface border border-border text-xs text-text-main placeholder:text-text-subtle focus:outline-none focus:border-aurora transition-colors"
            />
          </div>

          {/* Sector Selector */}
          <div>
            <select
              value={selectedSector}
              onChange={(e) => setSelectedSector(e.target.value)}
              className="w-full px-3 py-1.5 rounded-lg bg-surface border border-border text-xs text-text-main focus:outline-none focus:border-aurora transition-colors"
            >
              {sectors.map((s) => (
                <option key={s} value={s}>
                  Sector: {s}
                </option>
              ))}
            </select>
          </div>

          {/* Province Selector */}
          <div>
            <select
              value={selectedProvince}
              onChange={(e) => setSelectedProvince(e.target.value)}
              className="w-full px-3 py-1.5 rounded-lg bg-surface border border-border text-xs text-text-main focus:outline-none focus:border-aurora transition-colors"
            >
              {provinces.map((pr) => (
                <option key={pr} value={pr}>
                  Province/Territory: {pr}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Results Table */}
      <div className="glass-card rounded-xl border border-border/80 overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-[#08130E] text-text-muted font-mono uppercase text-[10px] border-b border-border/80">
              <tr>
                <th className="px-4 py-3">Project & Location</th>
                <th className="px-3 py-3">Sector</th>
                <th className="px-3 py-3">Stage</th>
                <th className="px-3 py-3 text-right">CAPEX (CAD)</th>
                <th className="px-3 py-3 text-right">Buildability</th>
                <th className="px-3 py-3 text-right">Investability</th>
                <th className="px-3 py-3 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-borderSubtle">
              {filtered.map((p) => {
                const bScore = p.scores?.buildability || 0;
                const iScore = p.scores?.investability || 0;
                return (
                  <tr key={p.id} className="hover:bg-cardHover/70 transition-colors group">
                    <td className="px-4 py-3.5">
                      <Link href={`/projects/${p.slug}`} className="font-semibold text-text-main group-hover:text-aurora text-sm block transition-colors">
                        {p.name}
                      </Link>
                      <div className="text-[11px] text-text-subtle flex items-center gap-1.5 mt-0.5">
                        <MapPin className="h-3 w-3 text-aurora" />
                        <span>{p.location_name} ({p.province})</span>
                        <span>•</span>
                        <span className="text-text-muted">{p.subsector}</span>
                      </div>
                    </td>
                    <td className="px-3 py-3.5">
                      <span className="inline-block px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-[11px] text-aurora font-mono">
                        {p.sector}
                      </span>
                    </td>
                    <td className="px-3 py-3.5">
                      <span className="inline-block px-2.5 py-0.5 rounded-full bg-gold/10 border border-gold/30 text-[10px] text-gold font-mono font-semibold">
                        {p.current_stage}
                      </span>
                    </td>
                    <td className="px-3 py-3.5 text-right font-bold font-tabular text-text-main">
                      ${(p.capex_cad / 1e9).toFixed(2)}B
                    </td>
                    <td className="px-3 py-3.5 text-right">
                      <span className="font-bold font-tabular text-aurora text-xs">
                        {bScore.toFixed(1)}
                      </span>
                      <span className="text-text-subtle text-[10px]">/100</span>
                    </td>
                    <td className="px-3 py-3.5 text-right">
                      <span className="font-bold font-tabular text-gold text-xs">
                        {iScore.toFixed(1)}
                      </span>
                      <span className="text-text-subtle text-[10px]">/100</span>
                    </td>
                    <td className="px-3 py-3.5 text-right">
                      <Link
                        href={`/projects/${p.slug}`}
                        className="inline-flex items-center gap-1 px-3 py-1 rounded-lg bg-surface border border-border hover:border-aurora text-aurora text-[11px] transition-colors font-medium shadow-sm"
                      >
                        Profile <ChevronRight className="h-3 w-3" />
                      </Link>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
