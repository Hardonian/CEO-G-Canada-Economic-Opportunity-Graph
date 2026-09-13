"use client";

import React, { useState } from "react";
import Link from "next/link";
import { 
  Landmark, 
  TrendingUp, 
  ShieldCheck, 
  Layers, 
  ArrowUpRight, 
  Printer, 
  FileText, 
  Cpu, 
  Zap, 
  Anchor, 
  DollarSign, 
  Award,
  CheckCircle2,
  ExternalLink
} from "lucide-react";
import { FALLBACK_PROJECTS, FALLBACK_RADAR_STATS } from "@/lib/data";

export default function PrimeMinisterBriefingPage() {
  const [activeCorridor, setActiveCorridor] = useState<"ALL" | "CLEAN_POWER" | "MINERALS" | "AI_SOVEREIGNTY" | "ARCTIC">("ALL");

  const corridors = [
    {
      id: "CLEAN_POWER",
      title: "Clean Baseload & SMR Nuclear Backbone",
      jurisdiction: "ON / QC / NB",
      capex: "$8.15B CAD",
      summary: "Deployment of grid-scale SMRs (Darlington BWRX-300), utility battery storage (Oneida), and cross-border HVDC interties to power Canadian re-industrialization.",
      projects: ["proj-darlington-smr", "proj-oneida-battery", "proj-windsor-intertie"],
      multiplier: "4.2x",
      indigenousEquity: "100% in Oneida; OPG First Nations IBA in Darlington",
    },
    {
      id: "MINERALS",
      title: "Critical Minerals & Battery Refining Spine",
      jurisdiction: "ON / QC",
      capex: "$4.40B CAD",
      summary: "High-grade nickel, cobalt, and lithium deposits with net-zero carbon capture, securing continental battery supply chains under Canada-US bilateral defense pacts.",
      projects: ["proj-crawford-nickel", "proj-sudbury-nickel"],
      multiplier: "3.6x",
      indigenousEquity: "Taykwa Tagamou Nation & Mattagami First Nation co-development",
    },
    {
      id: "AI_SOVEREIGNTY",
      title: "Sovereign AI Compute & Boreal Data Trusts",
      jurisdiction: "QC / BC",
      capex: "$2.80B CAD",
      summary: "350MW zero-carbon hyperscale compute powered by northern hydro basins, exempt from US CLOUD Act extraterritorial surveillance, with bilingual foundation model parity.",
      projects: ["proj-chisasibi-ai"],
      multiplier: "3.9x",
      indigenousEquity: "Cree Nation of Chisasibi 50% equity co-ownership",
    },
    {
      id: "ARCTIC",
      title: "Arctic Sovereignty & Northern Gateway",
      jurisdiction: "NU / MB",
      capex: "$3.25B CAD",
      summary: "Strategic dual-use deepwater port infrastructure, continental Arctic railway, and northern hydro-fibre links replacing 40M litres of remote diesel annually.",
      projects: ["proj-churchill-arctic-gateway", "proj-kivalliq-hydro-fibre", "proj-arctic-ideas"],
      multiplier: "3.4x",
      indigenousEquity: "100% Indigenous-owned Arctic Gateway Group & Sakku Investments",
    },
  ];

  const selectedCorridors = activeCorridor === "ALL" 
    ? corridors 
    : corridors.filter(c => c.id === activeCorridor);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* Official PM Briefing Header */}
      <div className="relative rounded-2xl p-6 sm:p-8 bg-gradient-to-b from-card to-surface border border-border/90 shadow-2xl overflow-hidden">
        {/* Ambient Glow */}
        <div className="absolute top-0 right-0 -mt-16 -mr-16 w-96 h-96 rounded-full bg-gold/10 blur-3xl pointer-events-none"></div>
        <div className="absolute bottom-0 left-1/4 -mb-16 w-80 h-80 rounded-full bg-aurora/10 blur-3xl pointer-events-none"></div>

        <div className="relative z-10 space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-borderSubtle pb-4">
            <div className="flex items-center gap-2 text-xs font-mono text-gold">
              <span className="h-2 w-2 rounded-full bg-gold animate-pulse"></span>
              PRIVY COUNCIL OFFICE // CABINET COMMITTEE ON ECONOMIC RECOVERY & PRODUCTIVITY
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => window.print()}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-surface hover:bg-card border border-border text-xs font-mono text-text-muted hover:text-aurora transition-colors"
              >
                <Printer className="h-3.5 w-3.5" />
                Print Cabinet Deck
              </button>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-primary/20 text-aurora border border-primary/30">
                SEPTEMBER 14, 2026
              </span>
            </div>
          </div>

          <div className="space-y-2">
            <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
              National Capital Formation: <span className="text-gold">Prime Minister's Command Briefing</span>
            </h1>
            <p className="text-sm text-text-muted max-w-4xl leading-relaxed">
              Briefing memorandum prepared for <strong className="text-text-main">The Right Honourable Mark Carney</strong> on unlocking Canada's 
              <span className="text-aurora font-bold"> $150B+ Major Project Pipeline</span> through the open Canada Economic Graph Schema (CEGS 0.1) and quantitative capital crowding-in.
            </p>
          </div>
        </div>

        {/* Cabinet Macro KPIs Bar */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mt-8 pt-6 border-t border-borderSubtle relative z-10">
          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[10px] font-mono text-text-subtle uppercase">Indexed Megaprojects</div>
            <div className="text-3xl font-black font-tabular text-aurora mt-1">
              $16.32B <span className="text-xs text-text-subtle font-normal">CAD</span>
            </div>
            <div className="text-[11px] text-text-muted mt-1">
              10 Nation-Building Assets
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[10px] font-mono text-text-subtle uppercase">Private Crowding-In Multiplier</div>
            <div className="text-3xl font-black font-tabular text-gold mt-1">
              3.8x
            </div>
            <div className="text-[11px] text-text-muted mt-1">
              $3.80 Private / $1.00 Federal
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[10px] font-mono text-text-subtle uppercase">Indigenous Equity Rate</div>
            <div className="text-3xl font-black font-tabular text-aurora-mint mt-1">
              78%
            </div>
            <div className="text-[11px] text-text-muted mt-1">
              Co-Ownership & Board Seats
            </div>
          </div>

          <div className="glass-card p-4 rounded-xl border border-border/80">
            <div className="text-[10px] font-mono text-text-subtle uppercase">Accelerating Towards FID</div>
            <div className="text-3xl font-black font-tabular text-text-main mt-1">
              5 Assets
            </div>
            <div className="text-[11px] text-gold mt-1">
              $1.42B Flow This Week
            </div>
          </div>
        </div>
      </div>

      {/* Strategic Corridor Explorer */}
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div>
            <h2 className="text-lg font-bold text-text-main tracking-tight flex items-center gap-2">
              <Landmark className="h-5 w-5 text-gold" />
              Strategic Growth Corridors (2026–2035)
            </h2>
            <p className="text-xs text-text-muted mt-0.5">
              Targeted regional clusters maximizing supply chain density and institutional co-investment.
            </p>
          </div>

          {/* Corridor Filter Tabs */}
          <div className="inline-flex rounded-lg bg-surface p-1 border border-borderSubtle text-xs font-mono">
            <button
              onClick={() => setActiveCorridor("ALL")}
              className={`px-3 py-1 rounded-md transition-all ${
                activeCorridor === "ALL" ? "bg-card text-aurora font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              All 4 Corridors
            </button>
            <button
              onClick={() => setActiveCorridor("CLEAN_POWER")}
              className={`px-3 py-1 rounded-md transition-all ${
                activeCorridor === "CLEAN_POWER" ? "bg-card text-aurora font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              Clean Baseload
            </button>
            <button
              onClick={() => setActiveCorridor("MINERALS")}
              className={`px-3 py-1 rounded-md transition-all ${
                activeCorridor === "MINERALS" ? "bg-card text-gold font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              Critical Minerals
            </button>
            <button
              onClick={() => setActiveCorridor("AI_SOVEREIGNTY")}
              className={`px-3 py-1 rounded-md transition-all ${
                activeCorridor === "AI_SOVEREIGNTY" ? "bg-card text-aurora-mint font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              AI Compute
            </button>
            <button
              onClick={() => setActiveCorridor("ARCTIC")}
              className={`px-3 py-1 rounded-md transition-all ${
                activeCorridor === "ARCTIC" ? "bg-card text-gold font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              Arctic Gateway
            </button>
          </div>
        </div>

        {/* Corridor Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {selectedCorridors.map((c) => (
            <div key={c.id} className="glass-panel p-6 rounded-xl border border-border/80 space-y-4 shadow-xl">
              <div className="flex items-start justify-between gap-3 border-b border-borderSubtle pb-3">
                <div>
                  <div className="text-[10px] font-mono text-aurora uppercase tracking-wider">{c.jurisdiction}</div>
                  <h3 className="text-base font-bold text-text-main mt-0.5">{c.title}</h3>
                </div>
                <div className="text-right">
                  <div className="text-[10px] font-mono text-text-subtle uppercase">Target CAPEX</div>
                  <div className="text-lg font-black font-tabular text-gold">{c.capex}</div>
                </div>
              </div>

              <p className="text-xs text-text-muted leading-relaxed">
                {c.summary}
              </p>

              <div className="space-y-2 text-xs pt-1">
                <div className="flex items-center justify-between text-[11px] p-2 rounded bg-surface border border-borderSubtle">
                  <span className="text-text-subtle font-mono">Crowding-In Multiplier:</span>
                  <span className="font-bold text-aurora font-mono">{c.multiplier} private capital</span>
                </div>
                <div className="flex items-center justify-between text-[11px] p-2 rounded bg-surface border border-borderSubtle">
                  <span className="text-text-subtle font-mono">Indigenous Equity:</span>
                  <span className="font-semibold text-text-main text-[11px]">{c.indigenousEquity}</span>
                </div>
              </div>

              {/* Linked Projects */}
              <div className="space-y-1.5 pt-2">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Corridor Anchor Projects:</div>
                <div className="space-y-1">
                  {c.projects.map((projId) => {
                    const p = FALLBACK_PROJECTS.find(x => x.id === projId);
                    if (!p) return null;
                    return (
                      <Link
                        key={p.id}
                        href={`/projects/${p.slug}`}
                        className="p-2 rounded bg-card hover:bg-cardHover border border-borderSubtle flex items-center justify-between group transition-colors text-xs"
                      >
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-text-main group-hover:text-aurora transition-colors text-xs">
                            {p.name}
                          </span>
                          <span className="text-[10px] font-mono text-text-subtle">({p.province})</span>
                        </div>
                        <span className="font-bold font-tabular text-aurora text-xs">
                          ${(p.capex_cad / 1e9).toFixed(2)}B
                        </span>
                      </Link>
                    );
                  })}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Capital Stacking Architecture Section */}
      <div className="glass-card p-8 rounded-2xl border border-border/80 space-y-6 shadow-2xl">
        <div className="space-y-2">
          <div className="inline-flex items-center gap-1.5 px-3 py-0.5 rounded-full bg-surface border border-borderSubtle text-[11px] font-mono text-gold">
            <Layers className="h-3.5 w-3.5 text-gold" />
            FEDERAL BALANCE SHEET LEVERAGE MODEL
          </div>
          <h2 className="text-xl sm:text-2xl font-black text-text-main">
            How Canada Crowds-In <span className="text-aurora">$3.80 Private Capital</span> per $1.00 Federal Fiscal Support
          </h2>
          <p className="text-xs sm:text-sm text-text-muted max-w-4xl leading-relaxed">
            By layering non-dilutive Investment Tax Credits with patient CIB concessionary debt and Indigenous equity loans, 
            Canada lowers weighted average cost of capital (WACC) from 11.2% down to 6.8%, unlocking institutional investment grade debt.
          </p>
        </div>

        {/* Visual Capital Stack Bar */}
        <div className="space-y-3">
          <div className="w-full h-8 rounded-xl bg-background border border-borderSubtle flex overflow-hidden font-mono text-xs font-bold shadow-inner">
            <div className="bg-emerald-600 h-full flex items-center justify-center text-white px-2" style={{ width: "35%" }}>
              Sponsor Equity 35%
            </div>
            <div className="bg-teal-500 h-full flex items-center justify-center text-white px-2" style={{ width: "35%" }}>
              Commercial Debt 35%
            </div>
            <div className="bg-amber-500 h-full flex items-center justify-center text-black px-2" style={{ width: "15%" }}>
              CIB Concessionary 15%
            </div>
            <div className="bg-yellow-400 h-full flex items-center justify-center text-black px-2" style={{ width: "10%" }}>
              ITC 10%
            </div>
            <div className="bg-orange-500 h-full flex items-center justify-center text-white px-2" style={{ width: "5%" }}>
              5%
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 text-xs font-mono text-text-muted">
            <div className="flex items-center gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full bg-emerald-600"></span>
              <span>CPPIB / CDPQ / OMERS</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full bg-teal-500"></span>
              <span>Big 5 Canadian Banks</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full bg-amber-500"></span>
              <span>Canada Infra Bank</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full bg-yellow-400"></span>
              <span>Clean Tech / CMITC</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="h-2.5 w-2.5 rounded-full bg-orange-500"></span>
              <span>First Nations Equity</span>
            </div>
          </div>
        </div>
      </div>

      {/* Immediate Cabinet Directives Recommendation */}
      <div className="p-6 rounded-xl border border-gold/40 bg-gradient-to-r from-card to-surface text-xs space-y-3 shadow-xl">
        <div className="flex items-center gap-2 font-bold text-text-main text-sm">
          <Award className="h-5 w-5 text-gold" />
          <span>Recommended Prime Ministerial Directives for September 14:</span>
        </div>
        <ol className="list-decimal list-inside space-y-1.5 text-text-muted leading-relaxed pl-1">
          <li>
            <strong className="text-text-main">Adopt CEGS 0.1 as the Federal Standard:</strong> Direct IAAC, NRCan, CIB, and PSPC to format statutory major project data according to the Canada Economic Graph Schema.
          </li>
          <li>
            <strong className="text-text-main">Open the Federal Procurement Pipeline:</strong> Release a 12-month forward tender lookahead on CanadaBuys via automated CEGS feeds.
          </li>
          <li>
            <strong className="text-text-main">Expand CIB Indigenous Equity Loan Facility:</strong> Increase maximum guarantee limits to 100% of First Nations equity buy-ins across Tier 1 clean energy assets.
          </li>
        </ol>
      </div>
    </div>
  );
}
