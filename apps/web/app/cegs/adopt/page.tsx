import Link from "next/link";
import { Terminal, Download, ArrowLeft, CheckCircle2, Building, BookOpen, Search, ShieldCheck, ArrowRight, Radio } from "lucide-react";

export default function CEGSAdoptPage() {
  const personas = [
    {
      title: "Government & Economic Development Agencies",
      useCase: "Standardize provincial, territorial, and municipal major project inventories into interoperable, machine-readable datasets without proprietary software lock-in.",
      command: "cog export project darlington-smr --format cegs > project.cegs.json",
    },
    {
      title: "Investigative & Financial Journalists",
      useCase: "Audit capital announcements and regulatory timelines with immutable event histories and cryptographic SHA-256 evidence hashes.",
      command: "cog cegs diff 2025_snapshot.json 2026_snapshot.json",
    },
    {
      title: "Infrastructure Funds & Institutional Investors",
      useCase: "Ingest structured project data into quantitative pipeline models, tracking stage velocity and capital gaps across Canada.",
      command: "curl -s \"$COG_BASE_URL/api/v1/cegs/export\" | jq .",
    },
    {
      title: "Suppliers & EPCM Contractors",
      useCase: "Automate discovery of downstream engineering, civil, electrical, and camp infrastructure requirements derived from project scale.",
      command: "cog search \"procurement nuclear\"",
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <div className="border-b border-border/80 pb-6 space-y-2">
        <Link href="/cegs" className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-aurora transition-colors font-mono">
          <ArrowLeft className="h-3.5 w-3.5" /> Back to CEGS Specification
        </Link>
        <div className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-1 shadow-sm">
          <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
          INTEROPERABILITY & FEDERATION PLAYBOOK
        </div>
        <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
          CEGS Standard <span className="text-aurora">Adoption Guide</span>
        </h1>
        <p className="text-xs sm:text-sm text-text-muted max-w-4xl leading-relaxed">
          How Canadian organizations, researchers, investors, and public agencies can produce, validate, and exchange compliant economic datasets.
        </p>
      </div>

      {/* Personas Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {personas.map((p) => (
          <div key={p.title} className="glass-card p-6 rounded-2xl border border-border/80 space-y-4 shadow-xl">
            <h2 className="font-bold text-base text-text-main">{p.title}</h2>
            <p className="text-xs text-text-muted leading-relaxed">{p.useCase}</p>
            <div className="pt-2">
              <div className="text-[10px] font-mono text-text-subtle uppercase mb-1.5">Example Open Command</div>
              <div className="bg-[#040806] p-3 rounded-xl border border-borderSubtle text-aurora font-mono text-xs select-all shadow-inner">
                $ {p.command}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* 3-Step Quickstart */}
      <div className="glass-card p-6 sm:p-8 rounded-2xl border border-border/80 space-y-4 shadow-xl">
        <h2 className="text-base font-bold text-text-main font-mono uppercase tracking-wider flex items-center gap-2">
          <ShieldCheck className="h-4 w-4 text-aurora" />
          <span>3-Step Adoption Workflow</span>
        </h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs font-mono">
          <div className="p-4 rounded-xl bg-surface border border-borderSubtle space-y-2">
            <div className="font-bold text-aurora text-sm">Step 1: Export or Format</div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Wrap your project or organization records inside the canonical CEGS envelope with standard URI identifiers.
            </p>
          </div>
          <div className="p-4 rounded-xl bg-surface border border-borderSubtle space-y-2">
            <div className="font-bold text-aurora-mint text-sm">Step 2: Validate Conformance</div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Run <code>cog cegs validate &lt;file&gt;</code> to verify schema constraints, cryptographic SHA-256 hashes, and epistemic states.
            </p>
          </div>
          <div className="p-4 rounded-xl bg-surface border border-borderSubtle space-y-2">
            <div className="font-bold text-gold text-sm">Step 3: Ingest & Federate</div>
            <p className="text-text-muted text-[11px] leading-relaxed">
              Publish compliant JSONL datasets or consume live CEGS feeds directly into CanadaOpportunityGraph or custom institutional pipelines.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
