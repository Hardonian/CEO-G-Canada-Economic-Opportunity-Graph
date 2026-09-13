"use client";

import { Cpu, ShieldCheck, ShieldAlert, CheckCircle2, XCircle, Info, Lock, Globe } from "lucide-react";

export default function AISovereigntyPage() {
  const benchmarks = [
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
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border pb-6">
        <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-card border border-border text-[11px] font-mono text-accent-cyan mb-2">
          <Cpu className="h-3 w-3" />
          CANADIAN AI SOVEREIGNTY INDEX — METHODOLOGY V1.0
        </div>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
          Canadian AI Sovereignty Index
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-1 max-w-4xl leading-relaxed">
          A transparent, deterministic scoring framework evaluating artificial intelligence infrastructure and compute services 
          on data residency, compute residency, domestic ownership, foreign extraterritorial legal exposure (US CLOUD Act), 
          linguistic bilingual parity, and Quebec Law 25 privacy posture.
        </p>
      </div>

      {/* Benchmark Scorecards */}
      <div className="space-y-6">
        {benchmarks.map((b) => (
          <div key={b.name} className="bg-card p-6 rounded border border-border space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 border-b border-borderSubtle pb-4">
              <div>
                <div className="flex items-center gap-2 flex-wrap">
                  <h2 className="text-base font-bold text-text-main">{b.name}</h2>
                  <span className="text-[10px] font-mono px-2 py-0.2 rounded bg-surface border border-borderSubtle text-accent-cyan">
                    {b.type}
                  </span>
                  <span className="text-[10px] font-mono px-2 py-0.2 rounded bg-surface border border-borderSubtle text-text-muted">
                    {b.location}
                  </span>
                </div>
                <p className="text-xs text-text-subtle mt-1">{b.summary}</p>
              </div>

              <div className="text-right shrink-0">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Sovereignty Score</div>
                <div className="text-2xl font-extrabold font-tabular text-accent-cyan">
                  {b.overallScore.toFixed(1)}<span className="text-xs font-normal text-text-subtle">/100</span>
                </div>
              </div>
            </div>

            {/* Dimension Breakdown Grid */}
            <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3 text-xs font-mono">
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Data Residency</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.dataResidency.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Compute Residency</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.computeResidency.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">CA Ownership</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.canadianOwnership.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">CLOUD Act Immunity</div>
                <div className={`text-sm font-bold mt-0.5 ${b.foreignLegalProtection < 5 ? "text-primary" : "text-accent-green"}`}>
                  {b.foreignLegalProtection.toFixed(1)}/10
                </div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Bilingual Parity</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.bilingualParity.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Quebec Law 25</div>
                <div className="text-sm font-bold text-text-main mt-0.5">{b.quebecLaw25.toFixed(1)}/10</div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Clean Power</div>
                <div className="text-sm font-bold text-accent-green mt-0.5">{b.cleanEnergy.toFixed(1)}/10</div>
              </div>
            </div>

            <div className="text-[11px] text-text-subtle font-mono pt-1">
              Energy Provenance: <span className="text-text-muted">{b.cleanPowerSource}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Legal & Compliance Notice */}
      <div className="p-4 rounded border border-borderSubtle bg-surface text-xs space-y-2 text-text-subtle leading-relaxed">
        <div className="flex items-center gap-1.5 font-bold text-text-muted">
          <ShieldAlert className="h-4 w-4 text-accent-gold" />
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
