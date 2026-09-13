"use client";

import { useState } from "react";
import { 
  Cpu, 
  ShieldCheck, 
  ShieldAlert, 
  CheckCircle2, 
  XCircle, 
  Info, 
  Lock, 
  Globe, 
  Zap, 
  Scale, 
  Sliders, 
  Sparkles,
  Server,
  Layers,
  ArrowRight
} from "lucide-react";

export default function AISovereigntyPage() {
  // Interactive Simulator State
  const [dataResident, setDataResident] = useState(true);
  const [computeResident, setComputeResident] = useState(true);
  const [canadianOwned, setCanadianOwned] = useState(true);
  const [cloudActImmune, setCloudActImmune] = useState(true);
  const [cleanHydro, setCleanHydro] = useState(true);
  const [bilingualModels, setBilingualModels] = useState(true);
  const [law25Compliant, setLaw25Compliant] = useState(true);

  // Calculate live score
  const calculateScore = () => {
    let score = 0;
    if (dataResident) score += 15;
    if (computeResident) score += 15;
    if (canadianOwned) score += 20;
    if (cloudActImmune) score += 25; // Critical extraterritorial weighting
    if (cleanHydro) score += 10;
    if (bilingualModels) score += 7.5;
    if (law25Compliant) score += 7.5;
    return score;
  };

  const currentScore = calculateScore();

  const benchmarks = [
    {
      name: "Chisasibi Sovereign Clean Compute Hyperscale Cluster",
      location: "Chisasibi, Eeyou Istchee (Quebec)",
      type: "First Nations Co-Owned Sovereign AI Cluster",
      overallScore: 98.5,
      dataResidency: 10.0,
      computeResidency: 10.0,
      canadianOwnership: 10.0,
      foreignLegalProtection: 10.0, // Fully immune to US CLOUD Act
      deploymentAutonomy: 9.8,
      bilingualParity: 9.5,
      quebecLaw25: 10.0,
      cleanEnergy: 10.0,
      cleanPowerSource: "Hydro-Québec La Grande Complex (100% Zero-Carbon)",
      summary: "350MW sovereign AI cluster powered by northern hydro basins. 50% Cree Nation equity co-ownership, operating with zero US extraterritorial jurisdiction exposure.",
    },
    {
      name: "Mila Sovereign Supercompute Cluster",
      location: "Montreal, Quebec",
      type: "Academic & Sovereign GPU Cluster",
      overallScore: 97.5,
      dataResidency: 10.0,
      computeResidency: 10.0,
      canadianOwnership: 10.0,
      foreignLegalProtection: 10.0,
      deploymentAutonomy: 9.5,
      bilingualParity: 9.5,
      quebecLaw25: 10.0,
      cleanEnergy: 9.9,
      cleanPowerSource: "Hydro-Québec (99.8% Clean Hydro)",
      summary: "Dedicated non-profit Canadian supercompute infrastructure operating under Canadian statutory privacy laws with zero US CLOUD Act foreign extraterritorial jurisdiction.",
    },
    {
      name: "Decentralized Canadian Private Cloud & Co-location",
      location: "Calgary / Vancouver",
      type: "Canadian Private Infrastructure",
      overallScore: 84.0,
      dataResidency: 10.0,
      computeResidency: 10.0,
      canadianOwnership: 10.0,
      foreignLegalProtection: 9.5,
      deploymentAutonomy: 8.0,
      bilingualParity: 7.5,
      quebecLaw25: 7.5,
      cleanEnergy: 6.5,
      cleanPowerSource: "Alberta Grid (Gas/Wind) / BC Hydro",
      summary: "Canadian-owned and operated infrastructure with high legal sovereignty and customer key control, with variable energy provenance depending on provincial interconnection.",
    },
    {
      name: "Hyperscale Commercial Cloud (Central Canada)",
      location: "Toronto / Montreal",
      type: "US-Headquartered Multi-Tenant Cloud",
      overallScore: 61.5,
      dataResidency: 9.5,
      computeResidency: 9.0,
      canadianOwnership: 0.0,
      foreignLegalProtection: 2.0, // High US CLOUD Act exposure
      deploymentAutonomy: 3.5,
      bilingualParity: 8.0,
      quebecLaw25: 8.5,
      cleanEnergy: 9.0,
      cleanPowerSource: "Ontario Grid (Nuclear + Hydro) / Hydro-Québec",
      summary: "High-performance infrastructure physically deployed in Canadian data centres, but legally subject to foreign extraterritorial warrants (US CLOUD Act Section 103) through US parent corporation.",
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
          <Cpu className="h-3.5 w-3.5 text-aurora" />
          CANADIAN AI SOVEREIGNTY INDEX — METHODOLOGY V1.0
        </div>
        <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
          Canadian <span className="text-aurora">AI Sovereignty Index</span>
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-2 max-w-4xl leading-relaxed">
          A transparent, deterministic scoring framework evaluating artificial intelligence infrastructure and compute services 
          on data residency, compute residency, domestic ownership, foreign extraterritorial legal exposure (US CLOUD Act Section 103), 
          linguistic bilingual parity, and Quebec Law 25 privacy posture.
        </p>
      </div>

      {/* Interactive Sovereignty Calculator Simulator */}
      <div className="glass-card p-6 sm:p-8 rounded-2xl border border-border/80 space-y-6 shadow-2xl">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-borderSubtle pb-4">
          <div>
            <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-surface text-gold border border-borderSubtle text-[11px] font-mono">
              <Sliders className="h-3.5 w-3.5 text-gold" />
              INTERACTIVE COMPUTE AUDIT SIMULATOR
            </div>
            <h2 className="text-lg sm:text-xl font-bold text-text-main mt-1">
              Simulate Your Organization's AI Sovereignty Posture
            </h2>
            <p className="text-xs text-text-muted mt-0.5">
              Toggle statutory, infrastructure, and energy attributes to evaluate compliance with Canadian sovereignty requirements.
            </p>
          </div>

          <div className="text-right shrink-0 bg-surface p-3.5 rounded-xl border border-borderSubtle">
            <div className="text-[10px] font-mono text-text-subtle uppercase">Simulated Score</div>
            <div className="text-3xl font-black font-tabular text-aurora">
              {currentScore.toFixed(1)}<span className="text-xs font-normal text-text-subtle">/100</span>
            </div>
            <div className={`text-[10px] font-mono mt-0.5 font-bold ${
              currentScore >= 90 ? "text-aurora" : currentScore >= 70 ? "text-gold" : "text-red-400"
            }`}>
              {currentScore >= 90 ? "SOVEREIGN CLASS (TIER 1)" : currentScore >= 70 ? "PROTECTED (TIER 2)" : "EXTRATERRITORIAL EXPOSURE"}
            </div>
          </div>
        </div>

        {/* Interactive Toggle Buttons */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs font-mono">
          <button
            onClick={() => setDataResident(!dataResident)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              dataResident ? "bg-primary/10 border-primary/40 text-text-main" : "bg-surface border-borderSubtle text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">Data Residency</div>
              <div className="text-[10px] text-text-subtle mt-0.5">Stored within CA border</div>
            </div>
            {dataResident ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-text-subtle shrink-0" />}
          </button>

          <button
            onClick={() => setComputeResident(!computeResident)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              computeResident ? "bg-primary/10 border-primary/40 text-text-main" : "bg-surface border-borderSubtle text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">Compute Residency</div>
              <div className="text-[10px] text-text-subtle mt-0.5">GPUs physical in Canada</div>
            </div>
            {computeResident ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-text-subtle shrink-0" />}
          </button>

          <button
            onClick={() => setCanadianOwned(!canadianOwned)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              canadianOwned ? "bg-primary/10 border-primary/40 text-text-main" : "bg-surface border-borderSubtle text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">Canadian Ownership</div>
              <div className="text-[10px] text-text-subtle mt-0.5">&gt;50% Canadian voting equity</div>
            </div>
            {canadianOwned ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-text-subtle shrink-0" />}
          </button>

          <button
            onClick={() => setCloudActImmune(!cloudActImmune)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              cloudActImmune ? "bg-primary/10 border-primary/40 text-text-main" : "bg-red-950/20 border-red-800/40 text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">CLOUD Act Immunity</div>
              <div className="text-[10px] text-text-subtle mt-0.5">No foreign parent warrants</div>
            </div>
            {cloudActImmune ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-red-400 shrink-0" />}
          </button>

          <button
            onClick={() => setCleanHydro(!cleanHydro)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              cleanHydro ? "bg-primary/10 border-primary/40 text-text-main" : "bg-surface border-borderSubtle text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">Clean Energy Source</div>
              <div className="text-[10px] text-text-subtle mt-0.5">&gt;95% Hydro or Nuclear</div>
            </div>
            {cleanHydro ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-text-subtle shrink-0" />}
          </button>

          <button
            onClick={() => setBilingualModels(!bilingualModels)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              bilingualModels ? "bg-primary/10 border-primary/40 text-text-main" : "bg-surface border-borderSubtle text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">Bilingual Parity</div>
              <div className="text-[10px] text-text-subtle mt-0.5">French & English token balance</div>
            </div>
            {bilingualModels ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-text-subtle shrink-0" />}
          </button>

          <button
            onClick={() => setLaw25Compliant(!law25Compliant)}
            className={`p-3 rounded-xl border text-left flex items-start justify-between gap-2 transition-all ${
              law25Compliant ? "bg-primary/10 border-primary/40 text-text-main" : "bg-surface border-borderSubtle text-text-muted"
            }`}
          >
            <div>
              <div className="font-bold">Quebec Law 25</div>
              <div className="text-[10px] text-text-subtle mt-0.5">Statutory privacy certified</div>
            </div>
            {law25Compliant ? <CheckCircle2 className="h-4 w-4 text-aurora shrink-0" /> : <XCircle className="h-4 w-4 text-text-subtle shrink-0" />}
          </button>
        </div>
      </div>

      {/* Benchmark Scorecards */}
      <div className="space-y-6">
        <h2 className="text-lg font-bold text-text-main flex items-center gap-2">
          <Server className="h-5 w-5 text-aurora" />
          <span>Canadian Compute Facility Benchmarks</span>
        </h2>

        {benchmarks.map((b) => (
          <div key={b.name} className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 border-b border-borderSubtle pb-4">
              <div>
                <div className="flex items-center gap-2 flex-wrap">
                  <h3 className="text-base font-bold text-text-main">{b.name}</h3>
                  <span className="text-[10px] font-mono px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-aurora font-medium">
                    {b.type}
                  </span>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-muted">
                    {b.location}
                  </span>
                </div>
                <p className="text-xs text-text-subtle mt-1.5">{b.summary}</p>
              </div>

              <div className="text-right shrink-0">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Sovereignty Score</div>
                <div className="text-3xl font-black font-tabular text-aurora">
                  {b.overallScore.toFixed(1)}<span className="text-xs font-normal text-text-subtle">/100</span>
                </div>
              </div>
            </div>

            {/* Dimension Breakdown Grid */}
            <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3 text-xs font-mono">
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Data Residency</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.dataResidency.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Compute Residency</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.computeResidency.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">CA Ownership</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.canadianOwnership.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">CLOUD Act Immunity</div>
                <div className={`text-sm font-bold mt-0.5 ${b.foreignLegalProtection < 5 ? "text-red-400" : "text-aurora"}`}>
                  {b.foreignLegalProtection.toFixed(1)}/10
                </div>
              </div>
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Bilingual Parity</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.bilingualParity.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Quebec Law 25</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.quebecLaw25.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Clean Power</div>
                <div className="text-sm font-bold text-gold mt-0.5">{b.cleanEnergy.toFixed(1)}/10</div>
              </div>
            </div>

            <div className="text-[11px] text-text-subtle font-mono pt-1">
              Energy Provenance: <span className="text-text-muted">{b.cleanPowerSource}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Legal & Compliance Notice */}
      <div className="p-5 rounded-2xl border border-borderSubtle bg-surface text-xs space-y-2 text-text-subtle leading-relaxed shadow-lg">
        <div className="flex items-center gap-1.5 font-bold text-text-muted">
          <ShieldAlert className="h-4 w-4 text-gold" />
          <span>RESEARCH METHODOLOGY NOTIFICATION</span>
        </div>
        <p className="text-[11px]">
          The Canadian AI Sovereignty Index is an open research heuristic created to assess structural risks in data sovereignty and sovereign infrastructure resilience. 
          It <strong className="text-text-muted">does not constitute legal advice, statutory privacy certification, or compliance endorsement</strong> under PIPEDA or Quebec Law 25. 
          Evaluation code is open source and versioned under <code>cai-sovereignty-v1.0</code>.
        </p>
      </div>
    </div>
  );
}
