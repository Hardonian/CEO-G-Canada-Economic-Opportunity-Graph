"use client";

import { Activity, ShieldCheck, AlertCircle, CheckCircle2, Clock, Database, RefreshCw } from "lucide-react";

export default function AdminPage() {
  const adapters = [
    {
      name: "iaac_registry",
      title: "Impact Assessment Agency of Canada (IAAC)",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 3,
      docsChanged: 3,
      failures: 0,
      latency: "42ms",
      lastCheck: "2 minutes ago",
    },
    {
      name: "canadabuys_procurement",
      title: "CanadaBuys Open Contracting Tenders",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 2,
      docsChanged: 2,
      failures: 0,
      latency: "58ms",
      lastCheck: "5 minutes ago",
    },
    {
      name: "nrcan_major_projects",
      title: "NRCan Major Projects Inventory",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 2,
      docsChanged: 2,
      failures: 0,
      latency: "35ms",
      lastCheck: "8 minutes ago",
    },
    {
      name: "ideas_defence_arctic",
      title: "DND Innovation for Defence Excellence and Security (IDEaS)",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 1,
      docsChanged: 1,
      failures: 0,
      latency: "28ms",
      lastCheck: "12 minutes ago",
    },
    {
      name: "cer_facilities",
      title: "Canada Energy Regulator (CER) Facility Filings",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 1,
      docsChanged: 1,
      failures: 0,
      latency: "49ms",
      lastCheck: "15 minutes ago",
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border pb-6 flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-card border border-border text-[11px] font-mono text-accent-cyan mb-2">
            <Activity className="h-3 w-3" />
            OPERATIONAL TELEMETRY & ADAPTER HEALTH
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
            Ingestion Pipeline Health Console
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-1">
            Real-time status of authoritative source adapters, document parse counters, change detection, and review queues.
          </p>
        </div>
        <button
          onClick={() => alert("Scheduled ingestion cycle triggered manually.")}
          className="px-3 py-1.5 rounded bg-card border border-border text-text-muted hover:text-text-main text-xs font-mono flex items-center gap-1.5 hover:bg-cardHover transition-colors"
        >
          <RefreshCw className="h-3.5 w-3.5 text-accent-cyan" /> Trigger Ingestion Cycle
        </button>
      </div>

      {/* Adapters Health Table */}
      <div className="bg-card rounded border border-border overflow-hidden">
        <div className="px-5 py-4 border-b border-border flex items-center justify-between">
          <h2 className="text-xs font-bold font-mono uppercase tracking-wider text-text-main">
            Authoritative Ingestion Adapters (5 Active)
          </h2>
          <span className="text-[10px] font-mono text-accent-green font-bold">ALL SYSTEMS HEALTHY</span>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-[#0B1224] text-text-subtle font-mono uppercase text-[10px] border-b border-border">
              <tr>
                <th className="px-5 py-3">Source Adapter</th>
                <th className="px-3 py-3">Tier</th>
                <th className="px-3 py-3">Status</th>
                <th className="px-3 py-3 text-right">Documents Seen</th>
                <th className="px-3 py-3 text-right">Failures</th>
                <th className="px-3 py-3 text-right">Latency</th>
                <th className="px-5 py-3 text-right">Last Verified</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-borderSubtle">
              {adapters.map((a) => (
                <tr key={a.name} className="hover:bg-cardHover transition-colors">
                  <td className="px-5 py-3.5">
                    <div className="font-semibold text-text-main text-sm">{a.title}</div>
                    <div className="text-[11px] font-mono text-text-subtle">{a.name}</div>
                  </td>
                  <td className="px-3 py-3.5">
                    <span className="px-2 py-0.5 rounded bg-surface border border-borderSubtle text-[10px] font-mono text-accent-green">
                      Tier {a.tier}
                    </span>
                  </td>
                  <td className="px-3 py-3.5">
                    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-accent-green/10 border border-accent-green/30 text-[11px] text-accent-green font-mono font-semibold">
                      <span className="h-1.5 w-1.5 rounded-full bg-accent-green"></span>
                      {a.status}
                    </span>
                  </td>
                  <td className="px-3 py-3.5 text-right font-mono font-tabular text-text-main">
                    {a.docsSeen}
                  </td>
                  <td className="px-3 py-3.5 text-right font-mono font-tabular text-text-subtle">
                    {a.failures}
                  </td>
                  <td className="px-3 py-3.5 text-right font-mono font-tabular text-accent-cyan">
                    {a.latency}
                  </td>
                  <td className="px-5 py-3.5 text-right font-mono text-text-muted text-[11px]">
                    {a.lastCheck}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
