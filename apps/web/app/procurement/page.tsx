"use client";

import Link from "next/link";
import { Activity, ExternalLink, Calendar, ShieldCheck, Building, Tag } from "lucide-react";
import { FALLBACK_PROCUREMENTS } from "@/lib/data";

export default function ProcurementPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
          <Activity className="h-3.5 w-3.5 text-aurora" />
          CANADABUYS & SOVEREIGN PROCUREMENT PIPELINE
        </div>
        <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
          Canadian <span className="text-gold">Procurement Radar</span>
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-2 max-w-4xl leading-relaxed">
          Tracking active tenders, RFPs, advance notices, and defence innovation challenges across federal, Crown corporation, 
          and utility buyers. Every notice is linked to parent infrastructure projects when corroborated by public evidence.
        </p>
      </div>

      {/* Procurement Notices List */}
      <div className="space-y-4">
        {FALLBACK_PROCUREMENTS.map((proc) => (
          <div key={proc.id} className="glass-card p-6 rounded-xl border border-border/80 space-y-3 shadow-lg">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 border-b border-borderSubtle pb-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2 flex-wrap text-xs font-mono">
                  <span className="px-2.5 py-0.5 rounded-full bg-surface border border-borderSubtle text-aurora font-bold">
                    {proc.tender_id}
                  </span>
                  <span className="px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-muted">
                    {proc.buyer_type}
                  </span>
                  <span className="px-2 py-0.5 rounded bg-gold/10 border border-gold/30 text-gold font-semibold">
                    {proc.stage}
                  </span>
                  <span className="px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-aurora font-semibold">
                    {proc.requirement_class}
                  </span>
                </div>
                <h2 className="font-bold text-text-main text-base pt-1">
                  {proc.title}
                </h2>
                <div className="text-xs text-text-subtle flex items-center gap-2">
                  <Building className="h-3 w-3 text-aurora" />
                  <span>Buyer: {proc.buyer}</span>
                </div>
              </div>

              <div className="text-right shrink-0">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Estimated CAD Value</div>
                <div className="text-xl font-extrabold font-tabular text-gold">
                  ${((proc.estimated_cad || 0) / 1e6).toFixed(1)}M <span className="text-xs text-text-subtle font-normal">CAD</span>
                </div>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-2 pt-1">
              {proc.categories.map((cat) => (
                <span key={cat} className="text-[10px] font-mono px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-muted flex items-center gap-1">
                  <Tag className="h-2.5 w-2.5" />
                  {cat}
                </span>
              ))}
            </div>

            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pt-3 border-t border-borderSubtle text-xs font-mono">
              <div className="text-text-subtle flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5" />
                <span>Closing Date: {proc.closing_date ? new Date(proc.closing_date).toLocaleDateString("en-CA") : "N/A"}</span>
              </div>
              <a
                href={proc.source_url}
                target="_blank"
                rel="noreferrer"
                className="text-accent-cyan hover:underline inline-flex items-center gap-1 font-semibold"
              >
                Official Tender Portal <ExternalLink className="h-3 w-3" />
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
