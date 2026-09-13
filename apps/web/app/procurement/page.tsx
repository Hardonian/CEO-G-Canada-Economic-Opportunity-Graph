"use client";

import { useState } from "react";
import Link from "next/link";
import { 
  Activity, 
  ExternalLink, 
  Calendar, 
  ShieldCheck, 
  Building, 
  Tag, 
  Search, 
  Filter, 
  Clock, 
  ArrowUpRight,
  FileCheck,
  Radio
} from "lucide-react";
import { FALLBACK_PROCUREMENTS, FALLBACK_PROJECTS } from "@/lib/data";

export default function ProcurementPage() {
  const [search, setSearch] = useState("");
  const [selectedBuyerType, setSelectedBuyerType] = useState("ALL");
  const [selectedStage, setSelectedStage] = useState("ALL");

  const buyerTypes = ["ALL", "Federal", "Crown", "Utility / Indigenous", "Indigenous Co-Ownership", "Federal / Indigenous"];
  const stages = ["ALL", "Active RFP", "RFI", "Advance Notice"];

  const filtered = FALLBACK_PROCUREMENTS.filter((proc) => {
    if (selectedBuyerType !== "ALL" && proc.buyer_type !== selectedBuyerType) return false;
    if (selectedStage !== "ALL" && proc.stage !== selectedStage) return false;
    if (search.trim() !== "") {
      const q = search.toLowerCase();
      const match = 
        proc.title.toLowerCase().includes(q) ||
        proc.tender_id.toLowerCase().includes(q) ||
        proc.buyer.toLowerCase().includes(q) ||
        proc.categories.some(c => c.toLowerCase().includes(q));
      if (!match) return false;
    }
    return true;
  });

  const totalProcurementCad = filtered.reduce((acc, p) => acc + (p.estimated_cad || 0), 0);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      {/* Header */}
      <div className="border-b border-border/80 pb-6 flex flex-col md:flex-row md:items-end justify-between gap-4">
        <div>
          <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
            <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
            CANADABUYS & SOVEREIGN PROCUREMENT PIPELINE
          </div>
          <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
            Canadian <span className="text-gold">Procurement Radar</span>
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-2 max-w-3xl leading-relaxed">
            Tracking active tenders, RFPs, advance notices, and sovereign innovation contracts across federal, Crown corporation, 
            and Indigenous utility buyers. All notices are cryptographically corroborated against CanadaBuys and linked to major infrastructure assets.
          </p>
        </div>

        <div className="text-right shrink-0 bg-surface/70 p-3.5 rounded-2xl border border-borderSubtle">
          <div className="text-[10px] font-mono text-text-subtle uppercase">Tracked Procurement Flow</div>
          <div className="text-2xl font-black font-tabular text-gold">
            ${(totalProcurementCad / 1e6).toFixed(1)}M <span className="text-xs text-text-subtle font-normal">CAD</span>
          </div>
          <div className="text-[10px] text-aurora font-mono">{filtered.length} Active Solicitations</div>
        </div>
      </div>

      {/* Interactive Search & Filters Bar */}
      <div className="glass-panel p-4 rounded-2xl border border-border/80 space-y-3 shadow-lg">
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {/* Search Bar */}
          <div className="relative">
            <Search className="h-4 w-4 absolute left-3 top-2.5 text-text-subtle" />
            <input
              type="text"
              placeholder="Search tender ID, buyer, keyword..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main placeholder:text-text-subtle focus:outline-none focus:border-aurora transition-colors font-mono"
            />
          </div>

          {/* Buyer Type Filter */}
          <div>
            <select
              value={selectedBuyerType}
              onChange={(e) => setSelectedBuyerType(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main focus:outline-none focus:border-aurora transition-colors font-mono"
            >
              {buyerTypes.map((bt) => (
                <option key={bt} value={bt}>
                  Buyer: {bt === "ALL" ? "All Buyers" : bt}
                </option>
              ))}
            </select>
          </div>

          {/* Stage Filter */}
          <div>
            <select
              value={selectedStage}
              onChange={(e) => setSelectedStage(e.target.value)}
              className="w-full px-3 py-2 rounded-xl bg-surface border border-border text-xs text-text-main focus:outline-none focus:border-aurora transition-colors font-mono"
            >
              {stages.map((st) => (
                <option key={st} value={st}>
                  Stage: {st === "ALL" ? "All Stages" : st}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Procurement Notices List */}
      <div className="space-y-4">
        {filtered.map((proc) => (
          <div key={proc.id} className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl hover:border-primary/40 transition-all">
            <div className="flex flex-col sm:flex-row sm:items-start justify-between gap-3 border-b border-borderSubtle pb-4">
              <div className="space-y-1.5">
                <div className="flex items-center gap-2 flex-wrap text-xs font-mono">
                  <span className="px-2.5 py-0.5 rounded-full bg-surface border border-borderSubtle text-aurora font-bold">
                    {proc.tender_id}
                  </span>
                  <span className="px-2 py-0.5 rounded-md bg-surface border border-borderSubtle text-text-muted">
                    {proc.buyer_type}
                  </span>
                  <span className="px-2.5 py-0.5 rounded-md bg-gold/10 border border-gold/30 text-gold font-semibold">
                    {proc.stage}
                  </span>
                  <span className="px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-aurora font-semibold flex items-center gap-1">
                    <ShieldCheck className="h-3 w-3" />
                    {proc.requirement_class}
                  </span>
                </div>
                <h2 className="font-bold text-text-main text-base pt-1">
                  {proc.title}
                </h2>
                <div className="text-xs text-text-subtle flex items-center gap-2">
                  <Building className="h-3.5 w-3.5 text-aurora" />
                  <span>Buyer: <strong className="text-text-main font-medium">{proc.buyer}</strong></span>
                </div>
              </div>

              <div className="text-right shrink-0">
                <div className="text-[10px] font-mono text-text-subtle uppercase">Estimated CAD Value</div>
                <div className="text-2xl font-black font-tabular text-gold">
                  ${((proc.estimated_cad || 0) / 1e6).toFixed(1)}M <span className="text-xs text-text-subtle font-normal">CAD</span>
                </div>
              </div>
            </div>

            {/* Categories Tag Pills */}
            <div className="flex flex-wrap items-center gap-2">
              {proc.categories.map((cat) => (
                <span key={cat} className="text-[10px] font-mono px-2.5 py-1 rounded-lg bg-surface border border-borderSubtle text-text-muted flex items-center gap-1">
                  <Tag className="h-2.5 w-2.5 text-aurora" />
                  {cat}
                </span>
              ))}
            </div>

            {/* Closing Date & Official Portal Action */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-3 border-t border-borderSubtle text-xs font-mono">
              <div className="text-text-subtle flex items-center gap-2">
                <Calendar className="h-3.5 w-3.5 text-gold" />
                <span>Closing Date: <strong className="text-text-main">{proc.closing_date ? new Date(proc.closing_date).toLocaleDateString("en-CA") : "N/A"}</strong></span>
                <span className="text-[10px] px-2 py-0.5 rounded bg-gold/10 text-gold border border-gold/30">
                  Active Notice
                </span>
              </div>
              <a
                href={proc.source_url}
                target="_blank"
                rel="noreferrer"
                className="px-3.5 py-1.5 rounded-xl bg-card hover:bg-cardHover border border-border text-aurora hover:border-primary/50 inline-flex items-center gap-1.5 font-semibold transition-all shadow-sm"
              >
                Official Tender Portal <ExternalLink className="h-3.5 w-3.5" />
              </a>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
