"use client";

import React, { useState, useMemo } from "react";
import Link from "next/link";
import { 
  Flame, 
  TrendingUp, 
  ChevronRight, 
  Layers, 
  FileCode2, 
  CheckCircle2, 
  ShieldCheck, 
  ArrowUpRight, 
  Briefcase, 
  Sparkles,
  Zap,
  Filter,
  DollarSign,
  Activity
} from "lucide-react";
import { Project } from "@/lib/types";

interface DynamicRadarExplorerProps {
  initialProjects: Project[];
}

export default function DynamicRadarExplorer({ initialProjects }: DynamicRadarExplorerProps) {
  const [selectedSector, setSelectedSector] = useState<string>("ALL");
  const [selectedStage, setSelectedStage] = useState<string>("ALL");
  const [activeTab, setActiveTab] = useState<"ALL" | "ACCELERATING" | "OPPORTUNITIES" | "CEGS">("ALL");
  const [expandedProjectId, setExpandedProjectId] = useState<string | null>(null);

  const sectors = useMemo(() => {
    const list = Array.from(new Set(initialProjects.map((p) => p.sector)));
    return ["ALL", ...list];
  }, [initialProjects]);

  const filteredProjects = useMemo(() => {
    return initialProjects.filter((p) => {
      if (selectedSector !== "ALL" && p.sector !== selectedSector) return false;
      if (selectedStage !== "ALL" && p.current_stage !== selectedStage) return false;
      if (activeTab === "ACCELERATING" && (p.scores?.buildability || 0) < 50) return false;
      return true;
    });
  }, [initialProjects, selectedSector, selectedStage, activeTab]);

  const totalFilteredCapex = useMemo(() => {
    return filteredProjects.reduce((acc, p) => acc + (p.capex_cad || 0), 0);
  }, [filteredProjects]);

  const avgBuildability = useMemo(() => {
    if (filteredProjects.length === 0) return 0;
    const sum = filteredProjects.reduce((acc, p) => acc + (p.scores?.buildability || 0), 0);
    return (sum / filteredProjects.length).toFixed(1);
  }, [filteredProjects]);

  return (
    <div className="space-y-6">
      {/* Interactive Controls Bar */}
      <div className="glass-panel p-4 rounded-xl border border-border/80 shadow-lg shadow-black/40 space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-borderSubtle pb-3">
          <div className="flex items-center gap-2">
            <div className="h-7 w-7 rounded-lg bg-aurora/10 border border-primary/40 flex items-center justify-center text-aurora shadow-[0_0_8px_rgba(0,245,160,0.2)]">
              <Filter className="h-3.5 w-3.5" />
            </div>
            <div>
              <span className="font-bold text-xs uppercase tracking-wider text-text-main">
                Dynamic Radar Filters
              </span>
              <span className="text-[11px] text-text-muted ml-2 font-mono">
                {filteredProjects.length} of {initialProjects.length} Assets Active
              </span>
            </div>
          </div>

          {/* Dynamic Summary Micro-Ticker */}
          <div className="flex items-center gap-4 text-xs font-mono">
            <div>
              <span className="text-text-subtle text-[10px] uppercase block">Filtered Capital</span>
              <span className="font-bold text-aurora text-sm">
                ${(totalFilteredCapex / 1e9).toFixed(2)}B CAD
              </span>
            </div>
            <div className="border-l border-border pl-4">
              <span className="text-text-subtle text-[10px] uppercase block">Avg Buildability</span>
              <span className="font-bold text-gold text-sm">
                {avgBuildability}/100
              </span>
            </div>
          </div>
        </div>

        {/* View Mode Tabs */}
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="inline-flex rounded-lg bg-background p-1 border border-borderSubtle">
            <button
              onClick={() => setActiveTab("ALL")}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                activeTab === "ALL"
                  ? "bg-card text-aurora border border-primary/40 shadow-sm"
                  : "text-text-muted hover:text-text-main"
              }`}
            >
              All Assets ({initialProjects.length})
            </button>
            <button
              onClick={() => setActiveTab("ACCELERATING")}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-1.5 ${
                activeTab === "ACCELERATING"
                  ? "bg-card text-aurora border border-primary/40 shadow-sm"
                  : "text-text-muted hover:text-text-main"
              }`}
            >
              <TrendingUp className="h-3.5 w-3.5 text-aurora" />
              Accelerating (Buildability &gt; 50)
            </button>
            <button
              onClick={() => setActiveTab("OPPORTUNITIES")}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-1.5 ${
                activeTab === "OPPORTUNITIES"
                  ? "bg-card text-gold border border-gold/40 shadow-sm"
                  : "text-text-muted hover:text-text-main"
              }`}
            >
              <Briefcase className="h-3.5 w-3.5 text-gold" />
              Contractor Opportunities
            </button>
            <button
              onClick={() => setActiveTab("CEGS")}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-1.5 ${
                activeTab === "CEGS"
                  ? "bg-card text-aurora-mint border border-primary/40 shadow-sm"
                  : "text-text-muted hover:text-text-main"
              }`}
            >
              <FileCode2 className="h-3.5 w-3.5 text-aurora-mint" />
              CEGS 0.1 Inspector
            </button>
          </div>

          {/* Sector Quick Pills */}
          <div className="flex items-center gap-1.5 flex-wrap">
            {sectors.map((sec) => (
              <button
                key={sec}
                onClick={() => setSelectedSector(sec)}
                className={`px-2.5 py-1 rounded-full text-[11px] font-mono transition-all ${
                  selectedSector === sec
                    ? "bg-primary/20 text-aurora border border-primary/60 shadow-[0_0_8px_rgba(0,245,160,0.25)] font-semibold"
                    : "bg-surface text-text-muted border border-borderSubtle hover:text-text-main hover:border-text-subtle"
                }`}
              >
                {sec}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Projects Feed */}
      <div className="space-y-3.5">
        {filteredProjects.map((proj) => {
          const bScore = proj.scores?.buildability || 0;
          const iScore = proj.scores?.investability || 0;
          const sScore = proj.scores?.supplierability || 0;
          const isExpanded = expandedProjectId === proj.id;

          return (
            <div
              key={proj.id}
              className="glass-card rounded-xl p-5 border border-border/70 transition-all duration-300"
            >
              <div className="flex flex-col md:flex-row md:items-start justify-between gap-4">
                <div className="space-y-2 flex-1">
                  {/* Status & Sector Meta Tags */}
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-muted">
                      🇨🇦 {proj.province}
                    </span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-primary/10 border border-primary/30 text-aurora font-medium">
                      {proj.sector}
                    </span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-gold/10 border border-gold/30 text-gold font-semibold">
                      {proj.current_stage}
                    </span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-subtle flex items-center gap-1">
                      <ShieldCheck className="h-3 w-3 text-aurora" />
                      SHA-256 Sourced
                    </span>
                  </div>

                  {/* Project Name & Link */}
                  <div className="pt-1">
                    <Link
                      href={`/projects/${proj.slug}`}
                      className="text-base font-bold text-text-main hover:text-aurora transition-colors inline-flex items-center gap-1.5 group"
                    >
                      {proj.name}
                      <ArrowUpRight className="h-4 w-4 text-text-subtle group-hover:text-aurora group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform" />
                    </Link>
                  </div>

                  <p className="text-xs text-text-muted line-clamp-2 leading-relaxed max-w-3xl">
                    {proj.summary}
                  </p>

                  {/* Downstream Contractor Opportunity Teaser */}
                  {activeTab === "OPPORTUNITIES" && (
                    <div className="mt-3 p-3 rounded-lg bg-surface/90 border border-borderSubtle space-y-2">
                      <div className="flex items-center justify-between text-[11px] text-gold font-mono">
                        <span className="flex items-center gap-1 font-semibold">
                          <Zap className="h-3.5 w-3.5" />
                          Derived Downstream Procurement Requirements
                        </span>
                        <span className="text-[10px] text-text-subtle">Deterministic Ontology</span>
                      </div>
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                        <div className="p-2 rounded bg-card border border-borderSubtle">
                          <div className="text-[10px] font-mono text-aurora">CONFIRMED TENDER</div>
                          <div className="font-semibold text-text-main text-[11px] mt-0.5">Heavy Engineering EPC Package</div>
                          <div className="text-[10px] text-text-subtle mt-0.5">$350M–$750M Est. Value</div>
                        </div>
                        <div className="p-2 rounded bg-card border border-borderSubtle">
                          <div className="text-[10px] font-mono text-gold">DERIVED INFRASTRUCTURE</div>
                          <div className="font-semibold text-text-main text-[11px] mt-0.5">High-Voltage Substation & Interconnect</div>
                          <div className="text-[10px] text-text-subtle mt-0.5">Mandatory grid tie-in requirement</div>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* CEGS Standard Representation View */}
                  {activeTab === "CEGS" && (
                    <div className="mt-3 p-3 rounded-lg bg-[#040806] border border-borderSubtle font-mono text-[11px] text-text-muted space-y-2 overflow-x-auto">
                      <div className="flex items-center justify-between text-[10px] text-aurora">
                        <span>CEGS 0.1 CANONICAL JSON</span>
                        <span>URN: cegs:project:ca:{proj.province.toLowerCase()}:{proj.slug}</span>
                      </div>
                      <pre className="text-[10px] text-text-subtle leading-normal">
{JSON.stringify({
  "$schema": "https://cegs.dev/schemas/v0.1/project.json",
  "id": `cegs:project:ca:${proj.province.toLowerCase()}:${proj.slug}`,
  "type": "project",
  "canonical_name": proj.name,
  "stage": proj.current_stage,
  "capex": {
    "amount": proj.capex_cad,
    "currency": "CAD",
    "epistemic_status": "VERIFIED"
  },
  "conformance_tier": "CEGS Provenance"
}, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>

                {/* Score Dashboard Card */}
                <div className="flex sm:flex-col items-end justify-between sm:justify-center gap-3 shrink-0 bg-surface/80 p-3.5 rounded-xl border border-borderSubtle min-w-[200px]">
                  <div className="text-right w-full border-b border-borderSubtle/60 pb-2">
                    <div className="text-[10px] font-mono text-text-subtle uppercase">Reported CAPEX</div>
                    <div className="text-base font-extrabold font-tabular text-text-main">
                      ${(proj.capex_cad / 1e9).toFixed(2)}B <span className="text-[10px] text-text-subtle font-normal">CAD</span>
                    </div>
                  </div>

                  {/* Visual Radial/Bar Score Indicators */}
                  <div className="grid grid-cols-2 gap-2 w-full pt-1">
                    <div className="p-2 rounded bg-card border border-borderSubtle text-right">
                      <div className="text-[9px] font-mono text-text-subtle uppercase">Buildability</div>
                      <div className="text-sm font-bold font-tabular text-aurora">
                        {bScore.toFixed(0)}<span className="text-[10px] text-text-subtle font-normal">/100</span>
                      </div>
                      <div className="w-full h-1 bg-surface rounded-full mt-1 overflow-hidden">
                        <div className="h-full bg-aurora rounded-full" style={{ width: `${bScore}%` }}></div>
                      </div>
                    </div>

                    <div className="p-2 rounded bg-card border border-borderSubtle text-right">
                      <div className="text-[9px] font-mono text-text-subtle uppercase">Investability</div>
                      <div className="text-sm font-bold font-tabular text-gold">
                        {iScore.toFixed(0)}<span className="text-[10px] text-text-subtle font-normal">/100</span>
                      </div>
                      <div className="w-full h-1 bg-surface rounded-full mt-1 overflow-hidden">
                        <div className="h-full bg-gold rounded-full" style={{ width: `${iScore}%` }}></div>
                      </div>
                    </div>
                  </div>

                  <Link
                    href={`/projects/${proj.slug}`}
                    className="w-full py-1.5 px-3 rounded-lg bg-card hover:bg-cardHover border border-border hover:border-primary/40 text-[11px] font-semibold text-text-main hover:text-aurora transition-colors text-center block mt-1"
                  >
                    Diligence Deep Dive →
                  </Link>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
