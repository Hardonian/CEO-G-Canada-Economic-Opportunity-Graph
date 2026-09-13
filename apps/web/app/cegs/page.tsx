import Link from "next/link";
import { FileCode2, ShieldCheck, Database, CheckCircle2, Terminal, ArrowRight, Download, ExternalLink } from "lucide-react";

export default function CEGSPage() {
  const sampleJSON = `{
  "cegs": "0.1",
  "id": "cegs:project:ca:on:darlington-new-nuclear",
  "type": "project",
  "canonical_name": "Darlington Small Modular Reactor (SMR) Project",
  "jurisdiction": "CA:ON",
  "sector": "Nuclear & Clean Power",
  "stage": "CONSTRUCTION",
  "capex": {
    "amount": 3400000000,
    "currency": "CAD",
    "amount_type": "reported"
  },
  "proponents": [
    "cegs:org:ca:ontario-power-generation"
  ],
  "provenance": [
    "cegs:evidence:ca:iaac-registry-darlington-smr"
  ]
}`;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* Header */}
      <div className="border-b border-border pb-8">
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-card border border-border text-[11px] font-mono text-accent-cyan mb-2">
              <FileCode2 className="h-3.5 w-3.5" />
              OPEN ECONOMIC INTELLIGENCE SPECIFICATION
            </div>
            <h1 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-text-main">
              CEGS — Canada Economic Graph Schema
            </h1>
            <p className="text-sm text-text-muted mt-2 max-w-3xl leading-relaxed">
              An open, implementation-neutral data standard for representing Canadian major projects, capital commitments, 
              public procurements, regulatory milestones, Indigenous partnerships, and evidence-backed economic graphs.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/cegs/adopt"
              className="px-4 py-2 rounded bg-primary text-white text-xs font-semibold hover:bg-primary-hover transition-colors shadow-sm flex items-center gap-1.5"
            >
              Adoption Guide <ArrowRight className="h-4 w-4" />
            </Link>
            <a
              href="http://localhost:8080/api/v1/cegs/export"
              target="_blank"
              rel="noreferrer"
              className="px-3 py-2 rounded bg-card border border-border text-text-muted text-xs hover:text-text-main hover:bg-cardHover font-mono flex items-center gap-1.5"
            >
              <Download className="h-3.5 w-3.5" /> Live Snapshot
            </a>
          </div>
        </div>
      </div>

      {/* Grid: Actual JSON Sample & Standard Axioms */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 items-start">
        {/* Left: Raw Canonical JSON Example */}
        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs font-mono text-text-subtle">
            <span>Canonical CEGS 0.1 Resource Payload</span>
            <span className="text-accent-cyan">project.schema.json</span>
          </div>
          <div className="bg-[#050811] p-4 rounded border border-border font-mono text-xs text-text-muted overflow-x-auto selection:bg-accent-cyan/20">
            <pre className="text-[12px] leading-relaxed text-accent-cyan/90">
              <code>{sampleJSON}</code>
            </pre>
          </div>
          <p className="text-xs text-text-subtle">
            Every CEGS object is globally addressable via standard URIs (<code>cegs:&lt;type&gt;:&lt;jur&gt;:&lt;slug&gt;</code>) and retains explicit evidence references.
          </p>
        </div>

        {/* Right: Core Principles & Conformance Levels */}
        <div className="space-y-6">
          <div className="bg-card p-5 rounded border border-border space-y-4">
            <h2 className="text-sm font-bold uppercase tracking-wider text-text-main font-mono">
              Four Standard Conformance Levels
            </h2>
            <div className="space-y-3 text-xs">
              <div className="p-3 rounded bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-accent-cyan font-mono">Level 1:</span> CEGS Core
                </div>
                <p className="text-text-subtle text-[11px]">
                  Schema-valid projects, organizations, locations, and typed relationships.
                </p>
              </div>

              <div className="p-3 rounded bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-accent-green font-mono">Level 2:</span> CEGS Provenance
                </div>
                <p className="text-text-subtle text-[11px]">
                  Core conformance plus mandatory cryptographic content hashing (SHA-256) and source locators for every factual claim.
                </p>
              </div>

              <div className="p-3 rounded bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-accent-gold font-mono">Level 3:</span> CEGS Historical
                </div>
                <p className="text-text-subtle text-[11px]">
                  Provenance conformance plus an immutable, append-only event ledger capturing all lifecycle transitions.
                </p>
              </div>

              <div className="p-3 rounded bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-primary font-mono">Level 4:</span> CEGS Intelligence
                </div>
                <p className="text-text-subtle text-[11px]">
                  Historical conformance plus derived opportunities, momentum signals, and versioned scoring models.
                </p>
              </div>
            </div>
          </div>

          {/* CLI Validation Callout */}
          <div className="bg-surface p-4 rounded border border-borderSubtle font-mono text-xs space-y-2">
            <div className="flex items-center gap-1.5 font-bold text-text-main">
              <Terminal className="h-4 w-4 text-accent-cyan" />
              <span>Validate Any Dataset with the Open CLI:</span>
            </div>
            <div className="bg-[#050811] p-2.5 rounded border border-border text-accent-cyan text-[11px]">
              $ cog cegs validate dataset.json<br />
              ✓ PASSED: Conforms to CEGS Provenance.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
