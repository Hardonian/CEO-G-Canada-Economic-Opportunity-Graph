import Link from "next/link";
import { Shield, FileText, Database, GitFork, ExternalLink, Terminal } from "lucide-react";

export default function Footer() {
  return (
    <footer className="border-t border-border/80 bg-[#040806] text-text-muted text-xs">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
          {/* Col 1: Mission */}
          <div className="md:col-span-1 space-y-3">
            <div className="flex items-center gap-2">
              <span className="text-base">🍁</span>
              <span className="font-bold text-text-main text-sm">CanadaOpportunityGraph</span>
            </div>
            <p className="text-text-subtle leading-relaxed">
              The canonical machine-readable intelligence layer for major Canadian economic development, infrastructure capital, and sovereign procurement.
            </p>
            <div className="pt-2">
              <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-card border border-primary/30 text-[11px] font-mono text-aurora shadow-sm">
                <Database className="h-3 w-3" />
                CEGS 0.1 Reference Implementation
              </div>
            </div>
          </div>

          {/* Col 2: Platform Lenses */}
          <div>
            <h4 className="font-semibold text-text-main text-xs uppercase tracking-wider mb-3">Intelligence Lenses</h4>
            <ul className="space-y-2">
              <li><Link href="/" className="hover:text-text-main transition-colors">Capital Radar</Link></li>
              <li><Link href="/projects" className="hover:text-text-main transition-colors">Major Projects Directory</Link></li>
              <li><Link href="/map" className="hover:text-text-main transition-colors">Geospatial Infrastructure Map</Link></li>
              <li><Link href="/capital" className="hover:text-text-main transition-colors">Canadian Capital Stack</Link></li>
              <li><Link href="/procurement" className="hover:text-text-main transition-colors">Procurement Pipeline</Link></li>
              <li><Link href="/ai-sovereignty" className="hover:text-text-main transition-colors">AI Sovereignty Index</Link></li>
            </ul>
          </div>

          {/* Col 3: Standards & Data */}
          <div>
            <h4 className="font-semibold text-text-main text-xs uppercase tracking-wider mb-3">CEGS Standard</h4>
            <ul className="space-y-2">
              <li><Link href="/cegs" className="hover:text-text-main transition-colors">Specification v0.1</Link></li>
              <li><Link href="/cegs/adopt" className="hover:text-text-main transition-colors">Adoption Guide</Link></li>
              <li><Link href="/methodology" className="hover:text-text-main transition-colors">Scoring Methodology</Link></li>
              <li><a href="http://localhost:8080/api/v1/cegs/export" target="_blank" rel="noreferrer" className="hover:text-text-main transition-colors inline-flex items-center gap-1">Live CEGS JSON Export <ExternalLink className="h-2.5 w-2.5" /></a></li>
              <li><Link href="/admin" className="hover:text-text-main transition-colors">Source Health Telemetry</Link></li>
            </ul>
          </div>

          {/* Col 4: Developers & Open Source */}
          <div>
            <h4 className="font-semibold text-text-main text-xs uppercase tracking-wider mb-3">Developers & CLI</h4>
            <div className="space-y-2 font-mono text-[11px]">
              <div className="bg-surface p-2 rounded border border-borderSubtle text-text-subtle select-all">
                $ cog search "nuclear ontario"<br />
                $ cog cegs validate file.json
              </div>
              <p className="text-text-subtle text-[11px]">
                Open-source CLI and dataset released under Apache-2.0 and CC-BY-4.0.
              </p>
            </div>
          </div>
        </div>

        {/* Legal Disclaimer Box */}
        <div className="pt-6 border-t border-borderSubtle text-[11px] text-text-subtle space-y-2 leading-relaxed">
          <p>
            <strong className="text-text-muted">STATUTORY & REGULATORY DISCLAIMER:</strong> CanadaOpportunityGraph is an independent, open-source economic research platform and is not an agency of the Government of Canada or any provincial jurisdiction. Information is compiled from public authoritative filings (Impact Assessment Agency of Canada, Canadian Energy Regulator, CanadaBuys, NRCan).
          </p>
          <p>
            Deterministic scores (Buildability, Investability, Supplierability, Strategicity, AI Sovereignty Index) are mathematical research indicators and <strong className="text-text-muted">do not constitute investment advice, legal advice, or tax advice</strong>. Sourced facts retain cryptographic SHA-256 evidence provenance.
          </p>
        </div>

        <div className="mt-6 flex flex-col sm:flex-row items-center justify-between text-[11px] text-text-subtle pt-4 border-t border-borderSubtle">
          <div>© 2026 CanadaOpportunityGraph Open Consortium. All rights reserved.</div>
          <div className="flex items-center gap-4 mt-2 sm:mt-0">
            <span>en-CA / fr-CA Bilingual Ready</span>
            <span>•</span>
            <span>WCAG 2.2 AA Conformance Target</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
