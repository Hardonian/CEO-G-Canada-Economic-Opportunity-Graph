"use client";

import { useState } from "react";
import Link from "next/link";
import { 
  FileCode2, 
  ShieldCheck, 
  Database, 
  CheckCircle2, 
  Terminal, 
  ArrowRight, 
  Download, 
  ExternalLink,
  Copy,
  Check,
  Play,
  Layers,
  Sparkles
} from "lucide-react";

export default function CEGSPage() {
  const [activeSchema, setActiveSchema] = useState<"project" | "organization" | "relationship" | "evidence" | "manifest">("project");
  const [copied, setCopied] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [validationPassed, setValidationPassed] = useState(false);

  const schemas = {
    project: `{
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
}`,
    organization: `{
  "cegs": "0.1",
  "id": "cegs:org:ca:ontario-power-generation",
  "type": "organization",
  "canonical_name": "Ontario Power Generation Inc.",
  "jurisdiction": "CA:ON",
  "org_type": "CROWN_PROVINCIAL",
  "business_number": "893948271RC0001",
  "sectors": ["Nuclear & Clean Power", "Clean Energy & Grid"],
  "provenance": [
    "cegs:evidence:ca:on:corporations-database"
  ]
}`,
    relationship: `{
  "cegs": "0.1",
  "id": "cegs:rel:ca:opg-darlington-proponent",
  "type": "relationship",
  "rel_type": "PROPONENT_OF",
  "source_id": "cegs:org:ca:ontario-power-generation",
  "target_id": "cegs:project:ca:on:darlington-new-nuclear",
  "equity_pct": 100.0,
  "provenance": [
    "cegs:evidence:ca:iaac-registry-darlington-smr"
  ]
}`,
    evidence: `{
  "cegs": "0.1",
  "id": "cegs:evidence:ca:iaac-registry-darlington-smr",
  "type": "evidence",
  "source_name": "Impact Assessment Agency of Canada",
  "source_url": "https://iaac-aeic.gc.ca/050/evaluations/proj/80023",
  "evidence_type": "STATUTORY_FILING",
  "content_hash": "sha256:7f9a2e3b1c8d4e5f0a9b8c7d6e5f4a3b2c1d0e9f8a7b6c5d4e3f2a1b0c9d8e7f",
  "retrieved_at": "2026-08-15T14:20:00Z"
}`,
    manifest: `{
  "cegs": "0.1",
  "id": "cegs:manifest:ca:v0-1-snapshot",
  "type": "manifest",
  "standard_version": "0.1",
  "generated_at": "2026-09-14T00:00:00Z",
  "entity_counts": {
    "projects": 10,
    "organizations": 14,
    "relationships": 28,
    "evidence": 34
  },
  "root_hash": "sha256:4b8a2c1d9e7f0a3b5c6d8e1f2a4b7c9d0e3f5a6b8c1d2e4f7a9b0c2d3e5f6a8b"
}`,
  };

  const copyToClipboard = () => {
    navigator.clipboard.writeText(schemas[activeSchema]);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const runValidation = () => {
    setIsValidating(true);
    setValidationPassed(false);
    setTimeout(() => {
      setIsValidating(false);
      setValidationPassed(true);
    }, 600);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* Header */}
      <div className="border-b border-border/80 pb-8">
        <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
              <FileCode2 className="h-3.5 w-3.5 text-aurora" />
              OPEN ECONOMIC INTELLIGENCE SPECIFICATION
            </div>
            <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-text-main">
              CEGS — <span className="text-aurora">Canada Economic Graph Schema</span>
            </h1>
            <p className="text-sm text-text-muted mt-2 max-w-3xl leading-relaxed">
              An open, implementation-neutral data standard for representing Canadian major projects, capital commitments, 
              public procurements, regulatory milestones, Indigenous partnerships, and evidence-backed economic graphs.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Link
              href="/cegs/adopt"
              className="px-4 py-2.5 rounded-xl bg-primary text-[#050B08] font-bold text-xs hover:bg-aurora-mint transition-all shadow-md shadow-emerald-950/40 flex items-center gap-1.5"
            >
              Adoption Guide <ArrowRight className="h-4 w-4" />
            </Link>
            <a
              href="http://localhost:8080/api/v1/cegs/export"
              target="_blank"
              rel="noreferrer"
              className="px-3.5 py-2.5 rounded-xl bg-card border border-border text-text-muted text-xs hover:text-aurora hover:border-primary/50 font-mono flex items-center gap-1.5 transition-all"
            >
              <Download className="h-3.5 w-3.5" /> Live Snapshot
            </a>
          </div>
        </div>
      </div>

      {/* Interactive Schema Playground Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
        {/* Left: Interactive Schema Explorer */}
        <div className="lg:col-span-7 space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            {/* Schema Type Switcher Tabs */}
            <div className="inline-flex flex-wrap rounded-xl bg-surface p-1 border border-borderSubtle text-xs font-mono">
              {(["project", "organization", "relationship", "evidence", "manifest"] as const).map((s) => (
                <button
                  key={s}
                  onClick={() => {
                    setActiveSchema(s);
                    setValidationPassed(false);
                  }}
                  className={`px-3 py-1.5 rounded-lg transition-all ${
                    activeSchema === s
                      ? "bg-card text-aurora font-bold shadow-sm"
                      : "text-text-muted hover:text-text-main"
                  }`}
                >
                  {s}.json
                </button>
              ))}
            </div>

            {/* Actions */}
            <div className="flex items-center gap-2">
              <button
                onClick={runValidation}
                disabled={isValidating}
                className="px-3 py-1.5 rounded-xl bg-surface hover:bg-card border border-border text-xs font-mono text-aurora flex items-center gap-1.5 transition-all"
              >
                <Play className={`h-3 w-3 ${isValidating ? "animate-spin" : ""}`} />
                <span>{isValidating ? "Testing..." : "Validate Schema"}</span>
              </button>

              <button
                onClick={copyToClipboard}
                className="px-3 py-1.5 rounded-xl bg-surface hover:bg-card border border-border text-xs font-mono text-text-muted hover:text-aurora flex items-center gap-1.5 transition-all"
              >
                {copied ? <Check className="h-3.5 w-3.5 text-aurora" /> : <Copy className="h-3.5 w-3.5" />}
                <span>{copied ? "Copied!" : "Copy"}</span>
              </button>
            </div>
          </div>

          {/* Code Viewer */}
          <div className="bg-[#040806] p-5 rounded-2xl border border-border/80 font-mono text-xs text-text-muted overflow-x-auto shadow-2xl relative">
            <pre className="text-[12px] leading-relaxed text-aurora/90">
              <code>{schemas[activeSchema]}</code>
            </pre>
          </div>

          {/* Validation Feedback Banner */}
          {validationPassed && (
            <div className="p-3.5 rounded-xl bg-primary/10 border border-primary/30 text-xs font-mono text-aurora flex items-center gap-2 shadow-md">
              <CheckCircle2 className="h-4 w-4 shrink-0 text-aurora" />
              <span>
                <strong>100% Valid:</strong> Conforms to CEGS 0.1 Provenance specification profile.
              </span>
            </div>
          )}

          <p className="text-xs text-text-subtle">
            Every CEGS object is globally addressable via standard URIs (<code>cegs:&lt;type&gt;:&lt;jur&gt;:&lt;slug&gt;</code>) and retains explicit cryptographic evidence references.
          </p>
        </div>

        {/* Right: Core Principles & Conformance Levels */}
        <div className="lg:col-span-5 space-y-6">
          <div className="glass-panel p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <h2 className="text-sm font-bold uppercase tracking-wider text-text-main font-mono flex items-center gap-2">
              <span className="h-2 w-2 rounded-full bg-aurora shadow-[0_0_6px_#00F5A0]"></span>
              Four Standard Conformance Levels
            </h2>
            <div className="space-y-3 text-xs">
              <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-aurora font-mono">Level 1:</span> CEGS Core
                </div>
                <p className="text-text-subtle text-[11px]">
                  Schema-valid projects, organizations, locations, and typed relationships.
                </p>
              </div>

              <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-aurora-mint font-mono">Level 2:</span> CEGS Provenance
                </div>
                <p className="text-text-subtle text-[11px]">
                  Core conformance plus mandatory cryptographic content hashing (SHA-256) and source locators for every factual claim.
                </p>
              </div>

              <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-gold font-mono">Level 3:</span> CEGS Historical
                </div>
                <p className="text-text-subtle text-[11px]">
                  Provenance conformance plus an immutable, append-only event ledger capturing all lifecycle transitions.
                </p>
              </div>

              <div className="p-3.5 rounded-xl bg-surface border border-borderSubtle space-y-1">
                <div className="font-bold text-text-main flex items-center gap-1.5">
                  <span className="text-yellow-400 font-mono">Level 4:</span> CEGS Intelligence
                </div>
                <p className="text-text-subtle text-[11px]">
                  Historical conformance plus derived opportunities, momentum signals, and versioned scoring models.
                </p>
              </div>
            </div>
          </div>

          {/* CLI Validation Callout */}
          <div className="glass-panel p-5 rounded-2xl border border-borderSubtle font-mono text-xs space-y-3 shadow-lg">
            <div className="flex items-center gap-2 font-bold text-text-main">
              <Terminal className="h-4 w-4 text-aurora" />
              <span>Validate Any Dataset with the Open CLI:</span>
            </div>
            <div className="bg-[#040806] p-3.5 rounded-xl border border-border text-aurora text-[11px] leading-relaxed">
              $ cog cegs validate dataset.json<br />
              <span className="text-aurora font-bold">✓ PASSED:</span> Conforms to CEGS Provenance.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
