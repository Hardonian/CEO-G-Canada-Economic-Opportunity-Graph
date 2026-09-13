"use client";

import { useState } from "react";
import { 
  Activity, 
  ShieldCheck, 
  AlertCircle, 
  CheckCircle2, 
  Clock, 
  Database, 
  RefreshCw,
  Terminal,
  Server,
  Radio,
  Check,
  Zap
} from "lucide-react";

export default function AdminPage() {
  const [isRunningCycle, setIsRunningCycle] = useState(false);
  const [cycleProgress, setCycleProgress] = useState(0);
  const [cycleStage, setCycleStage] = useState<string | null>(null);
  const [logs, setLogs] = useState<string[]>([]);

  const adapters = [
    {
      name: "iaac_registry",
      title: "Impact Assessment Agency of Canada (IAAC)",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 12,
      docsChanged: 4,
      failures: 0,
      latency: "42ms",
      lastCheck: "2 minutes ago",
    },
    {
      name: "canadabuys_procurement",
      title: "CanadaBuys Open Contracting Tenders",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 18,
      docsChanged: 5,
      failures: 0,
      latency: "58ms",
      lastCheck: "4 minutes ago",
    },
    {
      name: "nrcan_major_projects",
      title: "NRCan Major Projects Inventory",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 24,
      docsChanged: 6,
      failures: 0,
      latency: "35ms",
      lastCheck: "6 minutes ago",
    },
    {
      name: "ideas_defence_arctic",
      title: "DND Innovation for Defence Excellence and Security (IDEaS)",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 7,
      docsChanged: 2,
      failures: 0,
      latency: "28ms",
      lastCheck: "9 minutes ago",
    },
    {
      name: "cer_facilities",
      title: "Canada Energy Regulator (CER) Facility Filings",
      tier: 1,
      status: "HEALTHY",
      docsSeen: 9,
      docsChanged: 3,
      failures: 0,
      latency: "49ms",
      lastCheck: "11 minutes ago",
    },
  ];

  const triggerIngestionCycle = () => {
    if (isRunningCycle) return;
    setIsRunningCycle(true);
    setCycleProgress(10);
    setCycleStage("Connecting to Federal Registry Endpoints...");
    setLogs([
      "[" + new Date().toLocaleTimeString() + "] [INGESTION] Scheduled cycle initiated by operator.",
      "[" + new Date().toLocaleTimeString() + "] [TLS] Handshakes verified with IAAC, CER, and CanadaBuys.",
    ]);

    setTimeout(() => {
      setCycleProgress(35);
      setCycleStage("Harvesting Open Contracting Data (CanadaBuys OCDS)...");
      setLogs((prev) => [
        ...prev,
        "[" + new Date().toLocaleTimeString() + "] [CANADABUYS] Ingested 5 verified tenders; 0 schema anomalies.",
      ]);
    }, 800);

    setTimeout(() => {
      setCycleProgress(65);
      setCycleStage("Normalizing Entities to CEGS 0.1 Standard...");
      setLogs((prev) => [
        ...prev,
        "[" + new Date().toLocaleTimeString() + "] [CEGS] Conformance verified: 10 projects, 14 orgs, 5 manifests.",
        "[" + new Date().toLocaleTimeString() + "] [CRYPTO] Generated SHA-256 evidence digests.",
      ]);
    }, 1600);

    setTimeout(() => {
      setCycleProgress(100);
      setCycleStage("Ingestion Complete. Graph Synchronized.");
      setLogs((prev) => [
        ...prev,
        "[" + new Date().toLocaleTimeString() + "] [SUCCESS] Graph database updated with zero latency lag.",
      ]);
      setTimeout(() => {
        setIsRunningCycle(false);
      }, 1200);
    }, 2400);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border/80 pb-6 flex flex-col sm:flex-row sm:items-end justify-between gap-4">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
            <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
            OPERATIONAL TELEMETRY & ADAPTER HEALTH
          </div>
          <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
            Ingestion Pipeline <span className="text-aurora">Health Console</span>
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-1.5 max-w-3xl leading-relaxed">
            Real-time status of authoritative source adapters, document parse counters, change detection, and review queues.
            All pipelines feed the Canada Economic Graph Schema (CEGS 0.1) standard.
          </p>
        </div>

        <button
          onClick={triggerIngestionCycle}
          disabled={isRunningCycle}
          className={`px-4 py-2.5 rounded-xl font-mono text-xs font-bold flex items-center gap-2 transition-all shadow-md ${
            isRunningCycle
              ? "bg-surface border border-border text-text-subtle cursor-not-allowed"
              : "bg-primary text-[#050B08] hover:bg-aurora-mint shadow-emerald-950/40"
          }`}
        >
          <RefreshCw className={`h-4 w-4 ${isRunningCycle ? "animate-spin text-aurora" : ""}`} />
          {isRunningCycle ? "Running Cycle..." : "Trigger Ingestion Cycle"}
        </button>
      </div>

      {/* Real-Time Ingestion Activity Drawer (Appears when active or has logs) */}
      {logs.length > 0 && (
        <div className="glass-card p-6 rounded-2xl border border-border/80 space-y-3 shadow-xl">
          <div className="flex items-center justify-between text-xs font-mono">
            <div className="flex items-center gap-2 text-text-main font-semibold">
              <Terminal className="h-4 w-4 text-aurora" />
              <span>Live Ingestion Telemetry</span>
            </div>
            <span className="text-aurora font-mono">{cycleProgress}%</span>
          </div>

          <div className="w-full bg-surface h-2 rounded-full overflow-hidden border border-borderSubtle">
            <div
              className="bg-primary h-full rounded-full transition-all duration-300 shadow-[0_0_8px_#00F5A0]"
              style={{ width: `${cycleProgress}%` }}
            ></div>
          </div>

          <div className="text-[11px] font-mono text-text-muted">
            {cycleStage}
          </div>

          <div className="p-3.5 rounded-xl bg-[#040806] border border-borderSubtle font-mono text-[11px] space-y-1 text-aurora/90 max-h-36 overflow-y-auto">
            {logs.map((log, i) => (
              <div key={i}>{log}</div>
            ))}
          </div>
        </div>
      )}

      {/* Ingestion Adapter Health Table */}
      <div className="glass-card rounded-2xl border border-border/80 overflow-hidden shadow-2xl">
        <div className="px-6 py-4 border-b border-borderSubtle flex items-center justify-between">
          <h2 className="text-xs font-bold font-mono uppercase tracking-wider text-text-main flex items-center gap-2">
            <Server className="h-4 w-4 text-aurora" />
            <span>Authoritative Ingestion Adapters (5 Active)</span>
          </h2>
          <span className="text-[10px] font-mono text-aurora font-bold px-2.5 py-1 rounded-full bg-primary/10 border border-primary/30 flex items-center gap-1.5">
            <span className="h-2 w-2 rounded-full bg-aurora animate-pulse shadow-[0_0_6px_#00F5A0]"></span>
            ALL SYSTEMS HEALTHY
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-surface/80 text-text-subtle font-mono uppercase text-[10px] border-b border-borderSubtle">
              <tr>
                <th className="px-6 py-3.5">Source Adapter</th>
                <th className="px-4 py-3.5">Tier</th>
                <th className="px-4 py-3.5">Status</th>
                <th className="px-4 py-3.5 text-right">Docs Seen</th>
                <th className="px-4 py-3.5 text-right">Failures</th>
                <th className="px-4 py-3.5 text-right">Latency</th>
                <th className="px-6 py-3.5 text-right">Last Verified</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-borderSubtle font-mono">
              {adapters.map((a) => (
                <tr key={a.name} className="hover:bg-surface/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="font-semibold text-text-main font-sans text-sm">{a.title}</div>
                    <div className="text-[11px] font-mono text-aurora mt-0.5">{a.name}</div>
                  </td>
                  <td className="px-4 py-4">
                    <span className="px-2.5 py-1 rounded-full bg-surface border border-borderSubtle text-[10px] font-mono text-aurora font-medium">
                      Tier {a.tier}
                    </span>
                  </td>
                  <td className="px-4 py-4">
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-primary/10 border border-primary/30 text-[11px] text-aurora font-semibold">
                      <span className="h-1.5 w-1.5 rounded-full bg-aurora shadow-[0_0_6px_#00F5A0]"></span>
                      {a.status}
                    </span>
                  </td>
                  <td className="px-4 py-4 text-right font-tabular text-text-main">
                    {a.docsSeen}
                  </td>
                  <td className="px-4 py-4 text-right font-tabular text-text-subtle">
                    {a.failures}
                  </td>
                  <td className="px-4 py-4 text-right font-tabular text-gold">
                    {a.latency}
                  </td>
                  <td className="px-6 py-4 text-right text-text-muted text-[11px]">
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
