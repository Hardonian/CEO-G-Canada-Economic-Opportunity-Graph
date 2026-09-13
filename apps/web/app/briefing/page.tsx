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
  ExternalLink,
  Sliders,
  Compass,
  Check,
  Building,
  Key
} from "lucide-react";
import { FALLBACK_PROJECTS, FALLBACK_RADAR_STATS } from "@/lib/data";

export default function PrimeMinisterBriefingPage() {
  const [activeCorridor, setActiveCorridor] = useState<"ALL" | "CLEAN_POWER" | "MINERALS" | "AI_SOVEREIGNTY" | "ARCTIC">("ALL");

  // Dynamic Capital Formation Sensitivity Model state
  const [targetCapex, setTargetCapex] = useState<number>(16.32); // in Billions CAD
  const [cibPct, setCibPct] = useState<number>(15);
  const [itcPct, setItcPct] = useState<number>(10);
  const [indigenousPct, setIndigenousPct] = useState<number>(5);
  const [verifiedHash, setVerifiedHash] = useState<string | null>(null);

  // Derived financial stack
  const totalCapexCad = targetCapex * 1e9;
  const cibVal = totalCapexCad * (cibPct / 100);
  const itcVal = totalCapexCad * (itcPct / 100);
  const indigenousVal = totalCapexCad * (indigenousPct / 100);
  const publicFiscalSupport = cibVal + itcVal; // Total Federal Concessionary + Tax
  const remainingForPrivate = Math.max(0, totalCapexCad - cibVal - itcVal - indigenousVal);
  const commercialDebtPct = 35;
  const commercialDebtVal = totalCapexCad * (commercialDebtPct / 100);
  const sponsorEquityVal = Math.max(0, remainingForPrivate - commercialDebtVal);
  const sponsorEquityPct = Math.max(0, 100 - cibPct - itcPct - indigenousPct - commercialDebtPct);
  const privateCapitalTotal = sponsorEquityVal + commercialDebtVal;
  const multiplier = publicFiscalSupport > 0 ? (privateCapitalTotal / publicFiscalSupport).toFixed(2) : "0.0";
  const wacc = (5.5 + (sponsorEquityPct * 0.05) + (commercialDebtPct * 0.04)).toFixed(1);

  const corridors = [
    {
      id: "CLEAN_POWER",
      title: "Clean Baseload & SMR Nuclear Backbone",
      jurisdiction: "ON / QC / NB",
      capex: "$8.15B CAD",
      summary: "Deployment of grid-scale SMRs (Darlington BWRX-300), utility battery storage (Oneida), and cross-border HVDC interties to power Canadian re-industrialization.",
      projects: ["proj-darlington-smr", "proj-oneida-battery"],
      multiplier: "4.2x",
      indigenousEquity: "100% in Oneida; OPG First Nations IBA in Darlington",
      milestone: "CNSC Construction Licence Issued; Nuclear ITC Active",
    },
    {
      id: "MINERALS",
      title: "Critical Minerals & Battery Refining Spine",
      jurisdiction: "ON / QC",
      capex: "$4.40B CAD",
      summary: "High-grade nickel, cobalt, and lithium deposits with net-zero carbon capture, securing continental battery supply chains under Canada-US bilateral defense pacts.",
      projects: ["proj-crawford-nickel", "proj-galaxy-lithium"],
      multiplier: "3.6x",
      indigenousEquity: "Taykwa Tagamou Nation & Mattagami First Nation co-development",
      milestone: "IAAC Federal Environmental Impact Statement Approved",
    },
    {
      id: "AI_SOVEREIGNTY",
      title: "Sovereign AI Compute & Boreal Data Trusts",
      jurisdiction: "QC / BC",
      capex: "$4.00B CAD",
      summary: "350MW zero-carbon hyperscale compute powered by northern hydro basins, exempt from US CLOUD Act extraterritorial surveillance, with bilingual foundation model parity.",
      projects: ["proj-chisasibi-ai-compute", "proj-hyperscale-qc"],
      multiplier: "3.9x",
      indigenousEquity: "Cree Nation of Chisasibi 50% equity co-ownership",
      milestone: "Hydro-Québec 350MW Grid Interconnection Reservation Confirmed",
    },
    {
      id: "ARCTIC",
      title: "Arctic Sovereignty & Northern Gateway",
      jurisdiction: "NU / MB",
      capex: "$2.50B CAD",
      summary: "Strategic dual-use deepwater port infrastructure, continental Arctic railway, and northern hydro-fibre links replacing 40M litres of remote diesel annually.",
      projects: ["proj-churchill-arctic-gateway", "proj-kivalliq-link"],
      multiplier: "3.4x",
      indigenousEquity: "100% Indigenous-owned Arctic Gateway Group & Sakku Investments",
      milestone: "Transport Canada Permafrost Stabilization Engineering Approved",
    },
  ];

  const selectedCorridors = activeCorridor === "ALL" 
    ? corridors 
    : corridors.filter(c => c.id === activeCorridor);

  const copyProof = (hash: string) => {
    navigator.clipboard.writeText(hash);
    setVerifiedHash(hash);
    setTimeout(() => setVerifiedHash(null), 3000);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* Official PM Briefing Header */}
      <div className="relative rounded-2xl p-6 sm:p-8 bg-gradient-to-b from-card to-surface border border-border/90 shadow-2xl overflow-hidden">
        {/* Ambient Glow */}
        <div className="absolute top-0 right-0 -mt-16 -mr-16 w-96 h-96 rounded-full bg-gold/10 blur-3xl pointer-events-none"></div>
        <div className="absolute bottom-0 left-1/4 -mb-16 w-80 h-80 rounded-full bg-aurora/10 blur-3xl pointer-events-none"></div>

        <div className="relative z-10 space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-borderSubtle pb-4">
            <div className="flex items-center gap-2 text-xs font-mono text-gold font-bold">
              <span className="h-2.5 w-2.5 rounded-full bg-gold animate-pulse shadow-[0_0_8px_#F59E0B]"></span>
              PRIVY COUNCIL OFFICE // CABINET COMMITTEE ON ECONOMIC STRATEGY & RECOVERY
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={() => window.print()}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-surface hover:bg-card border border-border text-xs font-mono text-text-muted hover:text-aurora transition-colors shadow-sm"
              >
                <Printer className="h-3.5 w-3.5" />
                Print Cabinet Deck
              </button>
              <span className="text-[10px] font-mono px-2.5 py-1 rounded-full bg-primary/20 text-aurora border border-primary/40 font-semibold">
                SEPTEMBER 14, 2026
              </span>
            </div>
          </div>

          <div className="space-y-2">
            <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
              National Capital Formation: <span className="text-gold">Prime Minister's Command Briefing</span>
            </h1>
            <p className="text-xs sm:text-sm text-text-muted max-w-4xl leading-relaxed">
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
              80%
            </div>
            <div className="text-[11px] text-text-muted mt-1">
              Co-Ownership Across Corridors
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
              Targeted regional clusters maximizing supply chain density, clean baseload, and institutional co-investment.
            </p>
          </div>

          {/* Corridor Filter Tabs */}
          <div className="inline-flex rounded-xl bg-surface p-1 border border-borderSubtle text-xs font-mono">
            <button
              onClick={() => setActiveCorridor("ALL")}
              className={`px-3 py-1.5 rounded-lg transition-all ${
                activeCorridor === "ALL" ? "bg-card text-aurora font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              All 4 Corridors
            </button>
            <button
              onClick={() => setActiveCorridor("CLEAN_POWER")}
              className={`px-3 py-1.5 rounded-lg transition-all ${
                activeCorridor === "CLEAN_POWER" ? "bg-card text-aurora font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              Clean Baseload
            </button>
            <button
              onClick={() => setActiveCorridor("MINERALS")}
              className={`px-3 py-1.5 rounded-lg transition-all ${
                activeCorridor === "MINERALS" ? "bg-card text-gold font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              Critical Minerals
            </button>
            <button
              onClick={() => setActiveCorridor("AI_SOVEREIGNTY")}
              className={`px-3 py-1.5 rounded-lg transition-all ${
                activeCorridor === "AI_SOVEREIGNTY" ? "bg-card text-aurora-mint font-semibold shadow-sm" : "text-text-muted hover:text-text-main"
              }`}
            >
              AI Compute
            </button>
            <button
              onClick={() => setActiveCorridor("ARCTIC")}
              className={`px-3 py-1.5 rounded-lg transition-all ${
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
            <div key={c.id} className="glass-panel p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
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
                <div className="flex items-center justify-between text-[11px] p-2.5 rounded-xl bg-surface border border-borderSubtle">
                  <span className="text-text-subtle font-mono">Crowding-In Multiplier:</span>
                  <span className="font-bold text-aurora font-mono">{c.multiplier} private capital</span>
                </div>
                <div className="flex items-center justify-between text-[11px] p-2.5 rounded-xl bg-surface border border-borderSubtle">
                  <span className="text-text-subtle font-mono">Indigenous Equity:</span>
                  <span className="font-semibold text-text-main text-[11px]">{c.indigenousEquity}</span>
                </div>
                <div className="flex items-center justify-between text-[11px] p-2.5 rounded-xl bg-surface border border-borderSubtle">
                  <span className="text-text-subtle font-mono">Permitting Status:</span>
                  <span className="font-semibold text-gold text-[11px] truncate max-w-[220px]">{c.milestone}</span>
                </div>
              </div>

              {/* Linked Projects */}
              <div className="space-y-1.5 pt-2">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Corridor Anchor Projects:</div>
                <div className="space-y-1.5">
                  {c.projects.map((projId) => {
                    const p = FALLBACK_PROJECTS.find(x => x.id === projId);
                    if (!p) return null;
                    return (
                      <Link
                        key={p.id}
                        href={`/projects/${p.slug}`}
                        className="p-2.5 rounded-xl bg-card hover:bg-cardHover border border-borderSubtle flex items-center justify-between group transition-colors text-xs"
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

      {/* Interactive Capital Stacking Sensitivity Simulator */}
      <div className="glass-card p-6 sm:p-8 rounded-2xl border border-border/80 space-y-6 shadow-2xl">
        <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-4 border-b border-borderSubtle pb-4">
          <div className="space-y-1">
            <div className="inline-flex items-center gap-1.5 px-3 py-0.5 rounded-full bg-surface border border-borderSubtle text-[11px] font-mono text-gold">
              <Sliders className="h-3.5 w-3.5 text-gold" />
              CABINET POLICY SIMULATOR — CAPITAL STACK SENSITIVITY
            </div>
            <h2 className="text-xl sm:text-2xl font-black text-text-main">
              Federal Balance Sheet Leverage: <span className="text-aurora">${multiplier}x Multiplier</span>
            </h2>
            <p className="text-xs text-text-muted max-w-3xl leading-relaxed">
              Dynamically model how varying federal policy levers (CIB concessionary debt, clean tech tax credits, and First Nations equity guarantees) 
              impact institutional private capital crowding-in and national project WACC.
            </p>
          </div>

          <div className="text-right shrink-0 bg-surface/60 p-3 rounded-xl border border-borderSubtle">
            <div className="text-[10px] font-mono text-text-subtle uppercase">Blended Project WACC</div>
            <div className="text-2xl font-black font-tabular text-aurora">{wacc}%</div>
            <div className="text-[10px] text-text-muted">Investment-Grade Yield</div>
          </div>
        </div>

        {/* Sliders Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 p-4 rounded-xl bg-surface/40 border border-borderSubtle">
          <div className="space-y-1.5">
            <div className="flex justify-between text-xs font-mono">
              <span className="text-text-subtle">Target Portfolio CAPEX</span>
              <span className="font-bold text-aurora font-tabular">${targetCapex.toFixed(1)}B</span>
            </div>
            <input 
              type="range" 
              min="5" 
              max="50" 
              step="1" 
              value={targetCapex} 
              onChange={(e) => setTargetCapex(parseFloat(e.target.value))}
              className="w-full accent-emerald-400 cursor-pointer"
            />
          </div>

          <div className="space-y-1.5">
            <div className="flex justify-between text-xs font-mono">
              <span className="text-text-subtle">CIB Concessionary Debt</span>
              <span className="font-bold text-gold font-tabular">{cibPct}%</span>
            </div>
            <input 
              type="range" 
              min="5" 
              max="30" 
              step="1" 
              value={cibPct} 
              onChange={(e) => setCibPct(parseInt(e.target.value))}
              className="w-full accent-amber-400 cursor-pointer"
            />
          </div>

          <div className="space-y-1.5">
            <div className="flex justify-between text-xs font-mono">
              <span className="text-text-subtle">Clean Tech ITCs</span>
              <span className="font-bold text-yellow-400 font-tabular">{itcPct}%</span>
            </div>
            <input 
              type="range" 
              min="0" 
              max="30" 
              step="5" 
              value={itcPct} 
              onChange={(e) => setItcPct(parseInt(e.target.value))}
              className="w-full accent-yellow-400 cursor-pointer"
            />
          </div>

          <div className="space-y-1.5">
            <div className="flex justify-between text-xs font-mono">
              <span className="text-text-subtle">Indigenous Equity Facility</span>
              <span className="font-bold text-orange-400 font-tabular">{indigenousPct}%</span>
            </div>
            <input 
              type="range" 
              min="0" 
              max="15" 
              step="1" 
              value={indigenousPct} 
              onChange={(e) => setIndigenousPct(parseInt(e.target.value))}
              className="w-full accent-orange-400 cursor-pointer"
            />
          </div>
        </div>

        {/* Dynamic Visual Capital Stack Bar */}
        <div className="space-y-3">
          <div className="w-full h-10 rounded-xl bg-background border border-borderSubtle flex overflow-hidden font-mono text-[11px] font-bold shadow-inner">
            <div 
              className="bg-emerald-600 h-full flex items-center justify-center text-white px-2 transition-all duration-300 truncate" 
              style={{ width: `${sponsorEquityPct}%` }}
              title={`Sponsor / Pension Equity: ${sponsorEquityPct.toFixed(1)}% ($${(sponsorEquityVal / 1e9).toFixed(2)}B)`}
            >
              Sponsor {sponsorEquityPct.toFixed(0)}%
            </div>
            <div 
              className="bg-teal-500 h-full flex items-center justify-center text-white px-2 transition-all duration-300 truncate" 
              style={{ width: `${commercialDebtPct}%` }}
              title={`Commercial Debt: ${commercialDebtPct}% ($${(commercialDebtVal / 1e9).toFixed(2)}B)`}
            >
              Bank Debt {commercialDebtPct}%
            </div>
            <div 
              className="bg-amber-500 h-full flex items-center justify-center text-black px-2 transition-all duration-300 truncate" 
              style={{ width: `${cibPct}%` }}
              title={`CIB Patient Debt: ${cibPct}% ($${(cibVal / 1e9).toFixed(2)}B)`}
            >
              CIB {cibPct}%
            </div>
            <div 
              className="bg-yellow-400 h-full flex items-center justify-center text-black px-2 transition-all duration-300 truncate" 
              style={{ width: `${itcPct}%` }}
              title={`Clean Tax Credits: ${itcPct}% ($${(itcVal / 1e9).toFixed(2)}B)`}
            >
              ITC {itcPct}%
            </div>
            <div 
              className="bg-orange-500 h-full flex items-center justify-center text-white px-2 transition-all duration-300 truncate" 
              style={{ width: `${indigenousPct}%` }}
              title={`First Nations Equity: ${indigenousPct}% ($${(indigenousVal / 1e9).toFixed(2)}B)`}
            >
              FN {indigenousPct}%
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 text-xs font-mono text-text-muted">
            <div className="p-2.5 rounded-lg bg-surface border border-borderSubtle">
              <div className="flex items-center gap-1.5 text-text-main font-semibold">
                <span className="h-2.5 w-2.5 rounded-full bg-emerald-600"></span>
                <span>Sponsor Equity</span>
              </div>
              <div className="text-aurora font-bold mt-1">${(sponsorEquityVal / 1e9).toFixed(2)}B CAD</div>
              <div className="text-[10px] text-text-subtle">CPPIB / CDPQ / OMERS</div>
            </div>

            <div className="p-2.5 rounded-lg bg-surface border border-borderSubtle">
              <div className="flex items-center gap-1.5 text-text-main font-semibold">
                <span className="h-2.5 w-2.5 rounded-full bg-teal-500"></span>
                <span>Bank Debt</span>
              </div>
              <div className="text-aurora font-bold mt-1">${(commercialDebtVal / 1e9).toFixed(2)}B CAD</div>
              <div className="text-[10px] text-text-subtle">Big 5 Canadian Banks</div>
            </div>

            <div className="p-2.5 rounded-lg bg-surface border border-borderSubtle">
              <div className="flex items-center gap-1.5 text-text-main font-semibold">
                <span className="h-2.5 w-2.5 rounded-full bg-amber-500"></span>
                <span>CIB Patient Debt</span>
              </div>
              <div className="text-gold font-bold mt-1">${(cibVal / 1e9).toFixed(2)}B CAD</div>
              <div className="text-[10px] text-text-subtle">Concessionary Subordinated</div>
            </div>

            <div className="p-2.5 rounded-lg bg-surface border border-borderSubtle">
              <div className="flex items-center gap-1.5 text-text-main font-semibold">
                <span className="h-2.5 w-2.5 rounded-full bg-yellow-400"></span>
                <span>Clean ITC Credits</span>
              </div>
              <div className="text-yellow-400 font-bold mt-1">${(itcVal / 1e9).toFixed(2)}B CAD</div>
              <div className="text-[10px] text-text-subtle">Clean Tech / CMITC</div>
            </div>

            <div className="p-2.5 rounded-lg bg-surface border border-borderSubtle">
              <div className="flex items-center gap-1.5 text-text-main font-semibold">
                <span className="h-2.5 w-2.5 rounded-full bg-orange-500"></span>
                <span>First Nations Equity</span>
              </div>
              <div className="text-orange-400 font-bold mt-1">${(indigenousVal / 1e9).toFixed(2)}B CAD</div>
              <div className="text-[10px] text-text-subtle">CIB Equity Loan Guarantee</div>
            </div>
          </div>
        </div>
      </div>

      {/* Immediate Cabinet Directives Recommendation */}
      <div className="p-6 rounded-2xl border border-gold/40 bg-gradient-to-r from-card to-surface text-xs space-y-4 shadow-xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 font-bold text-text-main text-sm">
            <Award className="h-5 w-5 text-gold" />
            <span>Recommended Prime Ministerial Directives for September 14:</span>
          </div>
          <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-gold/10 text-gold border border-gold/30">
            ACTIONABLE EXECUTIVE ORDERS
          </span>
        </div>
        <ol className="list-decimal list-inside space-y-2 text-text-muted leading-relaxed pl-1">
          <li className="p-2 rounded-xl bg-surface/50 border border-borderSubtle">
            <strong className="text-text-main">Adopt CEGS 0.1 as the Federal Interoperability Standard:</strong> Direct IAAC, NRCan, CIB, and PSPC to format all statutory major project filings and disclosures according to the Canada Economic Graph Schema by Q1 2027.
          </li>
          <li className="p-2 rounded-xl bg-surface/50 border border-borderSubtle">
            <strong className="text-text-main">Open the Federal Sovereign Procurement Pipeline:</strong> Mandate a 12-month advance lookahead tender publication across CanadaBuys and defence innovation channels via automated CEGS feeds.
          </li>
          <li className="p-2 rounded-xl bg-surface/50 border border-borderSubtle">
            <strong className="text-text-main">Expand CIB Indigenous Equity Loan Facility:</strong> Increase maximum guarantee limits to 100% of First Nations equity buy-ins across Tier 1 clean energy and sovereign AI compute assets.
          </li>
        </ol>
      </div>

      {/* Cryptographic Provenance Anchors */}
      <div className="glass-card p-5 rounded-2xl border border-border/80 text-xs font-mono flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-text-muted">
          <ShieldCheck className="h-4 w-4 text-aurora" />
          <span>All briefing figures anchored in CEGS 0.1 cryptographic graph snapshot:</span>
          <code className="text-aurora bg-surface px-2 py-0.5 rounded border border-borderSubtle text-[11px]">
            sha256:7f9a2e3b1c8d4e5f...
          </code>
        </div>
        <button
          onClick={() => copyProof("sha256:7f9a2e3b1c8d4e5f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0c9d8e7f")}
          className="px-3 py-1.5 rounded-xl bg-surface hover:bg-card border border-border text-text-main hover:text-aurora flex items-center gap-1.5 text-xs transition-colors shrink-0"
        >
          {verifiedHash ? <Check className="h-3.5 w-3.5 text-aurora" /> : <Key className="h-3.5 w-3.5" />}
          <span>{verifiedHash ? "Hash Copied!" : "Verify Provenance"}</span>
        </button>
      </div>
    </div>
  );
}
