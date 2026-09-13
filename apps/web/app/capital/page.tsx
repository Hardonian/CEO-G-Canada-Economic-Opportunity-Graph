"use client";

import { useState } from "react";
import { 
  Layers, 
  ShieldAlert, 
  CheckCircle2, 
  AlertTriangle, 
  FileText, 
  ChevronRight, 
  Calculator, 
  Landmark,
  Sparkles,
  PieChart,
  Radio
} from "lucide-react";

export default function CapitalStackPage() {
  const [capexInput, setCapexInput] = useState("1000000000"); // $1B CAD default
  const [selectedSector, setSelectedSector] = useState("Nuclear & Clean Power");
  const [isIndigenousPartner, setIsIndigenousPartner] = useState(true);

  const programs = [
    {
      id: "clean_tech_itc",
      name: "Clean Technology Investment Tax Credit (Clean Tech ITC)",
      admin: "Canada Revenue Agency (CRA) / Finance Canada",
      type: "Refundable Tax Credit",
      rate: 30,
      laborReq: true,
      eligibleSectors: ["Clean Energy & Grid", "Nuclear & Clean Power", "AI Compute & Data Centres"],
      statutoryRef: "Income Tax Act, s. 127.45",
      summary: "Refundable 30% tax credit on capital costs of eligible clean energy generation, SMRs, storage, and zero-emission compute power equipment.",
      exclusivity: ["clean_electricity_itc", "clean_hydrogen_itc"],
    },
    {
      id: "clean_electricity_itc",
      name: "Clean Electricity Investment Tax Credit",
      admin: "CRA / Natural Resources Canada",
      type: "Refundable Tax Credit",
      rate: 15,
      laborReq: true,
      eligibleSectors: ["Nuclear & Clean Power", "Clean Energy & Grid"],
      statutoryRef: "Budget Implementation Act / ITA",
      summary: "15% refundable credit applicable to Crown corporations, publicly owned utilities, and private developers investing in clean generation and interprovincial transmission.",
      exclusivity: ["clean_tech_itc"],
    },
    {
      id: "critical_minerals_itc",
      name: "Critical Mineral Exploration & Processing Tax Credit (CMITC)",
      admin: "Canada Revenue Agency (CRA)",
      type: "Refundable Tax Credit",
      rate: 30,
      laborReq: false,
      eligibleSectors: ["Critical Minerals"],
      statutoryRef: "Income Tax Act, s. 127(9)",
      summary: "30% investment tax credit for eligible critical mineral exploration and processing targeting Canada's 34 prioritized minerals.",
      exclusivity: [],
    },
    {
      id: "cib_financing",
      name: "Canada Infrastructure Bank (CIB) Clean Power Concessionary Debt",
      admin: "Canada Infrastructure Bank",
      type: "Concessionary Debt & Loan Guarantees",
      rate: 40,
      laborReq: true,
      eligibleSectors: ["Clean Energy & Grid", "Nuclear & Clean Power", "Transportation & Ports", "AI Compute & Data Centres", "Defence & Arctic"],
      statutoryRef: "Canada Infrastructure Bank Act",
      summary: "Long-term patient capital at below-commercial rates, equity loan guarantees, and Indigenous equity participation loans.",
      exclusivity: [],
    },
    {
      id: "sif_net_zero",
      name: "Strategic Innovation Fund (SIF) - Net Zero Accelerator",
      admin: "Innovation, Science and Economic Development (ISED)",
      type: "Federal Contribution",
      rate: 20,
      laborReq: false,
      eligibleSectors: ["Industrial & Manufacturing", "Critical Minerals", "Clean Energy & Grid", "AI Compute & Data Centres"],
      statutoryRef: "ISED SIF Policy Guidelines",
      summary: "Large-scale federal capital grants directly de-risking anchor manufacturing, gigafactory, and sovereign AI compute investments.",
      exclusivity: [],
    },
  ];

  const capex = parseFloat(capexInput) || 0;

  // Filter matching programs
  const matchingPrograms = programs.filter((p) => p.eligibleSectors.includes(selectedSector));

  // Compute stacking
  const results = matchingPrograms.map((prog) => {
    const grossVal = capex * (prog.rate / 100);
    let matchClass = "LIKELY_MATCH";
    let warning = "";

    if (prog.exclusivity.length > 0) {
      const hasConflict = matchingPrograms.some((m) => prog.exclusivity.includes(m.id));
      if (hasConflict) {
        matchClass = "REQUIRES_REVIEW";
        warning = "Stacking conflict: Cannot claim simultaneously with mutually exclusive credit on identical assets.";
      }
    }

    return {
      ...prog,
      grossVal,
      matchClass,
      warning,
    };
  });

  const totalTheoreticalPotential = results.reduce((acc, r) => acc + r.grossVal, 0);

  // Capital Stack Proportions for Visual Bar
  const itcShare = Math.min(30, (capex * 0.25) / (capex || 1) * 100);
  const cibShare = Math.min(25, (capex * 0.20) / (capex || 1) * 100);
  const indigenousShare = isIndigenousPartner ? 10 : 0;
  const bankDebtShare = 35;
  const sponsorEquityShare = Math.max(0, 100 - itcShare - cibShare - indigenousShare - bankDebtShare);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
          <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
          CANADIAN CAPITAL STACK & PROGRAM INTELLIGENCE
        </div>
        <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
          Capital Stack & <span className="text-gold">Incentive Simulator</span>
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-2 max-w-4xl leading-relaxed">
          Model interactions between federal clean economy Investment Tax Credits (Clean Tech ITC, Clean Electricity, CMITC), 
          the Canada Infrastructure Bank (CIB), the Strategic Innovation Fund (SIF), and Indigenous loan guarantees to calculate net equity required.
        </p>
      </div>

      {/* Interactive Simulator Bar */}
      <div className="glass-panel p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
        <div className="flex items-center gap-2 text-xs font-bold text-text-main font-mono">
          <Calculator className="h-4 w-4 text-aurora" />
          <span>Simulate Project Capital Stack Parameters</span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div>
            <label className="block text-[11px] font-mono text-text-subtle uppercase mb-1">Total Project CAPEX (CAD)</label>
            <div className="relative">
              <span className="absolute left-3 top-2 text-text-subtle text-xs font-mono">$</span>
              <input
                type="number"
                value={capexInput}
                onChange={(e) => setCapexInput(e.target.value)}
                className="w-full pl-7 pr-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main font-mono focus:outline-none focus:border-aurora transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="block text-[11px] font-mono text-text-subtle uppercase mb-1">Project Strategic Sector</label>
            <select
              value={selectedSector}
              onChange={(e) => setSelectedSector(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main font-mono focus:outline-none focus:border-aurora transition-colors"
            >
              <option value="Nuclear & Clean Power">Nuclear & Clean Power</option>
              <option value="Critical Minerals">Critical Minerals</option>
              <option value="Clean Energy & Grid">Clean Energy & Grid</option>
              <option value="AI Compute & Data Centres">AI Compute & Data Centres</option>
              <option value="Transportation & Ports">Transportation & Ports</option>
              <option value="Industrial & Manufacturing">Industrial & Manufacturing</option>
              <option value="Defence & Arctic">Defence & Arctic</option>
            </select>
          </div>

          <div>
            <label className="block text-[11px] font-mono text-text-subtle uppercase mb-1">Indigenous Equity Partnership</label>
            <button
              onClick={() => setIsIndigenousPartner(!isIndigenousPartner)}
              className={`w-full py-2 px-3 rounded-xl border text-xs font-mono text-left transition-all flex items-center justify-between ${
                isIndigenousPartner
                  ? "bg-primary/10 border-primary/40 text-aurora font-semibold shadow-sm"
                  : "bg-surface border-border text-text-muted"
              }`}
            >
              <span>{isIndigenousPartner ? "✓ Eligible for CIB Indigenous Guarantee" : "Standard Corporate Proponent"}</span>
              <span className="text-[10px] font-bold text-aurora font-mono">{isIndigenousPartner ? "+10% FN" : "0%"}</span>
            </button>
          </div>
        </div>

        {/* Visual Capital Stack Composition Bar */}
        <div className="space-y-2 pt-2">
          <div className="text-[10px] font-mono text-text-subtle uppercase flex justify-between">
            <span>Resulting Blended Capital Stack ($CAD {(capex / 1e6).toFixed(0)}M Base)</span>
            <span className="text-aurora font-bold font-tabular">WACC: ~6.4%</span>
          </div>
          <div className="w-full h-8 rounded-xl bg-background border border-borderSubtle flex overflow-hidden font-mono text-[11px] font-bold shadow-inner">
            <div 
              className="bg-emerald-600 h-full flex items-center justify-center text-white px-2 transition-all duration-300 truncate"
              style={{ width: `${sponsorEquityShare}%` }}
              title={`Sponsor Equity: ${sponsorEquityShare.toFixed(0)}%`}
            >
              Equity {sponsorEquityShare.toFixed(0)}%
            </div>
            <div 
              className="bg-teal-500 h-full flex items-center justify-center text-white px-2 transition-all duration-300 truncate"
              style={{ width: `${bankDebtShare}%` }}
              title={`Bank Debt: ${bankDebtShare}%`}
            >
              Debt {bankDebtShare}%
            </div>
            <div 
              className="bg-amber-500 h-full flex items-center justify-center text-black px-2 transition-all duration-300 truncate"
              style={{ width: `${cibShare}%` }}
              title={`CIB Concessionary: ${cibShare.toFixed(0)}%`}
            >
              CIB {cibShare.toFixed(0)}%
            </div>
            <div 
              className="bg-yellow-400 h-full flex items-center justify-center text-black px-2 transition-all duration-300 truncate"
              style={{ width: `${itcShare}%` }}
              title={`Clean Tax Credits: ${itcShare.toFixed(0)}%`}
            >
              ITC {itcShare.toFixed(0)}%
            </div>
            {isIndigenousPartner && (
              <div 
                className="bg-orange-500 h-full flex items-center justify-center text-white px-2 transition-all duration-300 truncate"
                style={{ width: `${indigenousShare}%` }}
                title={`Indigenous Equity: ${indigenousShare}%`}
              >
                FN {indigenousShare}%
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Stacking Evaluation Results */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-bold uppercase tracking-wider text-text-main font-mono flex items-center gap-2">
            <Layers className="h-4 w-4 text-aurora" />
            <span>Identified Funding & Incentive Programs ({results.length} Eligible)</span>
          </h2>
          <div className="text-xs font-mono text-text-subtle">
            Theoretical Combined Support: <span className="font-bold text-aurora font-tabular">${(totalTheoreticalPotential / 1e6).toFixed(0)}M CAD</span>
          </div>
        </div>

        <div className="grid grid-cols-1 gap-4">
          {results.map((res) => (
            <div key={res.id} className="glass-card p-6 rounded-2xl border border-border/80 space-y-3 shadow-xl">
              <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-2 border-b border-borderSubtle pb-3">
                <div>
                  <div className="flex items-center gap-2 flex-wrap">
                    <span className="font-bold text-text-main text-base">{res.name}</span>
                    <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-surface border border-borderSubtle text-aurora">
                      {res.type}
                    </span>
                    <span className={`text-[10px] font-mono px-2.5 py-0.5 rounded-full border font-semibold ${
                      res.matchClass === "LIKELY_MATCH"
                        ? "bg-primary/10 border-primary/30 text-aurora"
                        : "bg-gold/10 border-gold/30 text-gold"
                    }`}>
                      {res.matchClass}
                    </span>
                  </div>
                  <div className="text-[11px] text-text-subtle font-mono mt-1">
                    Administrator: {res.admin} | Statutory Ref: {res.statutoryRef}
                  </div>
                </div>

                <div className="text-right shrink-0">
                  <div className="text-[10px] font-mono text-text-subtle uppercase">Potential Public Support</div>
                  <div className="text-lg font-black font-tabular text-aurora">
                    ${(res.grossVal / 1e6).toFixed(1)}M CAD <span className="text-xs text-text-subtle font-normal">({res.rate}%)</span>
                  </div>
                </div>
              </div>

              <p className="text-xs text-text-muted leading-relaxed">
                {res.summary}
              </p>

              {res.warning && (
                <div className="flex items-center gap-2 p-3 rounded-xl bg-gold/10 border border-gold/30 text-gold text-[11px] font-mono">
                  <AlertTriangle className="h-4 w-4 shrink-0" />
                  <span>{res.warning}</span>
                </div>
              )}

              {res.laborReq && (
                <div className="text-[10px] font-mono text-text-subtle flex items-center gap-1.5 pt-1">
                  <span className="h-1.5 w-1.5 rounded-full bg-aurora"></span>
                  <span>Mandatory: Prevailing wage and minimum 10% apprentice labor hours required to secure top-tier incentive rate.</span>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Statutory Disclaimer */}
      <div className="p-5 rounded-2xl border border-borderSubtle bg-surface text-xs space-y-2 text-text-subtle leading-relaxed shadow-lg">
        <div className="flex items-center gap-1.5 font-bold text-text-muted">
          <ShieldAlert className="h-4 w-4 text-gold" />
          <span>STATUTORY DISCLAIMER & TAX NOTIFICATION</span>
        </div>
        <p className="text-[11px]">
          The figures, stacking models, and qualification statuses displayed above are computed for preliminary planning and economic research purposes only. 
          They do not constitute formal legal, accounting, financial, or tax advice. Actual Investment Tax Credit claims and government assistance ceilings 
          must be formally adjudicated under the <em>Income Tax Act</em> and certified by the Canada Revenue Agency or the respective Crown corporation administrator.
        </p>
      </div>
    </div>
  );
}
