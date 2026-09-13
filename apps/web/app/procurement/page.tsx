"use client";

import Link from "next/link";
import { Activity, ExternalLink, Calendar, ShieldCheck, Building, Tag } from "lucide-react";
import { FALLBACK_PROCUREMENTS } from "@/lib/data";

export default function ProcurementPage() {
  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Header */}
      <div className="border-b border-border pb-6">
        <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-card border border-border text-[11px] font-mono text-accent-cyan mb-2">
          <Activity className="h-3 w-3" />
          CANADABUYS & SOVEREIGN PROCUREMENT PIPELINE
        </div>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
          Canadian Procurement Radar
        </h1>
        <p className="text-xs sm:text-sm text-text-muted mt-1 max-w-4xl leading-relaxed">
          Tracking active tenders, RFPs, advance notices, and defence innovation challenges across federal, Crown corporation, 
          and utility buyers. Every notice is linked to parent infrastructure projects when corroborated by public evidence.
        </p>
      </div>

      {/* Procurement Notices List */}
      <div className="space-y-4">
        {FALLBACK_PROCUREMENTS.map((proc) => (
          <div key={proc.id} className="bg-card p-5 rounded border border-border space-y-3 hover:border-text-subtle transition-all">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 border-b border-borderSubtle pb-3">
              <div className="space-y-1">
                <div className="flex items-center gap-2 flex-wrap text-xs font-mono">
                  <span className="px-2 py-0.5 rounded bg-surface border border-borderSubtle text-accent-cyan font-bold">
                    {proc.tender_id}
                  </span>
                  <span className="px-2 py-0.5 rounded bg-surface border border-borderSubtle text-text-muted">
                    {proc.buyer_type}
                  </span>
                  <span className="px-2 py-0.5 rounded bg-primary/10 border border-primary/30 text-primary font-semibold">
                    {proc.stage}
                  </span>
                  <span className="px-2 py-0.5 rounded bg-accent-green/10 border border-accent-green/30 text-accent-green font-semibold">
                    {proc.requirement_class}
                  </span>
                </div>
                <h2 className="font-bold text-text-main text-base pt-1">
                  {proc.title}
                </h2>
                <div className="text-xs text-text-subtle flex items-center gap-2">
                  <Building className="h-3 w-3" />
                  <span>Buyer: {proc.buyer}</span>
                </div>
              </div>

              <div className="text-right shrink-0">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Estimated CAD Value</div>
                <div className="text-lg font-bold font-tabular text-accent-green">
                  ${((proc.estimated_cad || 0) / 1e6).toFixed(1)}M CAD
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
