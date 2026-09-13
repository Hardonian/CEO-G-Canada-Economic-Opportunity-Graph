import Link from "next/link";
import { Terminal, Download, ArrowLeft, CheckCircle2, Building, BookOpen, Search, ShieldCheck } from "lucide-react";

export default function CEGSAdoptPage() {
  const personas = [
    {
      title: "Government & Economic Development Agencies",
      useCase: "Standardize provincial and municipal major project inventories into interoperable, machine-readable datasets without proprietary software lock-in.",
      command: "cog export project darlington-smr --format cegs > project.cegs.json",
    },
    {
      title: "Investigative & Financial Journalists",
      useCase: "Audit capital announcements and regulatory timelines with immutable event histories and cryptographic evidence hashes.",
      command: "cog cegs diff 2025_snapshot.json 2026_snapshot.json",
    },
    {
      title: "Infrastructure Funds & Institutional Investors",
      useCase: "Ingest structured project data into quantitative pipeline models, tracking stage velocity and capital gaps across Canada.",
      command: "curl -s http://localhost:8080/api/v1/cegs/export | jq .",
    },
    {
      title: "Suppliers & EPCM Contractors",
      useCase: "Automate discovery of downstream engineering, civil, electrical, and camp infrastructure requirements derived from project scale.",
      command: "cog search \"procurement nuclear\"",
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <div className="border-b border-border pb-6 space-y-2">
        <Link href="/cegs" className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-main transition-colors">
          <ArrowLeft className="h-3.5 w-3.5" /> Back to CEGS Specification
        </Link>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
          CEGS Standard Adoption Guide
        </h1>
        <p className="text-xs sm:text-sm text-text-muted max-w-4xl leading-relaxed">
          How Canadian organizations, researchers, investors, and public agencies can produce, validate, and exchange compliant economic datasets.
        </p>
      </div>

      {/* Personas Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {personas.map((p) => (
          <div key={p.title} className="bg-card p-5 rounded border border-border space-y-3">
            <h2 className="font-bold text-sm text-text-main">{p.title}</h2>
            <p className="text-xs text-text-subtle leading-relaxed">{p.useCase}</p>
            <div className="pt-2">
              <div className="text-[10px] font-mono text-text-subtle uppercase mb-1">Example Command</div>
              <div className="bg-[#050811] p-2.5 rounded border border-borderSubtle text-accent-cyan font-mono text-xs select-all">
                $ {p.command}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 3-Step Quickstart */}
      <div className="bg-card p-6 rounded border border-border space-y-4">
        <h2 className="text-base font-bold text-text-main">3-Step Adoption Workflow</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
          <div className="p-4 rounded bg-surface border border-borderSubtle space-y-2">
            <div className="font-bold font-mono text-accent-cyan">Step 1: Export or Format</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Wrap your project or organization records inside the canonical CEGS envelope with standard URI identifiers.
            </p>
          </div>
          <div className="p-4 rounded bg-surface border border-borderSubtle space-y-2">
            <div className="font-bold font-mono text-accent-green">Step 2: Validate Conformance</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Run <code>cog cegs validate &lt;file&gt;</code> to verify schema constraints, cryptographic SHA-256 hashes, and epistemic states.
            </p>
          </div>
          <div className="p-4 rounded bg-surface border border-borderSubtle space-y-2">
            <div className="font-bold font-mono text-accent-gold">Step 3: Ingest & Federate</div>
            <p className="text-text-subtle text-[11px] leading-relaxed">
              Publish compliant JSONL datasets or consume live CEGS feeds directly into CanadaOpportunityGraph or custom analytics pipelines.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
