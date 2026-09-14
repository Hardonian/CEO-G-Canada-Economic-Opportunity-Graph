"use client";

import { useMemo, useState } from "react";
import { CheckCircle2, Cpu, Info, ShieldAlert, Sliders, XCircle } from "lucide-react";

type Criterion = {
  id: string;
  label: string;
  description: string;
  weight: number;
};

const criteria: Criterion[] = [
  { id: "data", label: "Canadian data residency", description: "Workload data is stored in Canada.", weight: 15 },
  { id: "compute", label: "Canadian compute residency", description: "The workload executes on infrastructure in Canada.", weight: 15 },
  { id: "control", label: "Operational control", description: "Documented Canadian control exists for privileged operations and keys.", weight: 20 },
  { id: "legal", label: "Legal exposure reviewed", description: "Qualified counsel has assessed applicable domestic and foreign disclosure obligations.", weight: 20 },
  { id: "continuity", label: "Continuity and portability", description: "Exit, recovery, and provider-portability controls have been tested.", weight: 15 },
  { id: "privacy", label: "Privacy controls assessed", description: "A current privacy impact and jurisdictional review exists.", weight: 10 },
  { id: "language", label: "Official-language needs tested", description: "Required English and French service quality has been evaluated.", weight: 5 },
];

export default function AISovereigntyPage() {
  const [answers, setAnswers] = useState<Record<string, boolean>>(() =>
    Object.fromEntries(criteria.map((criterion) => [criterion.id, false])),
  );

  const score = useMemo(
    () => criteria.reduce((total, criterion) => total + (answers[criterion.id] ? criterion.weight : 0), 0),
    [answers],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <header className="border-b border-border/80 pb-6">
        <div className="mb-3 inline-flex items-center gap-1.5 rounded-full border border-primary/40 bg-card px-3 py-1 font-mono text-[11px] text-aurora">
          <Cpu className="h-3.5 w-3.5" /> AI INFRASTRUCTURE SELF-ASSESSMENT
        </div>
        <h1 className="text-2xl font-black tracking-tight text-text-main sm:text-4xl">
          AI Sovereignty <span className="text-aurora">Control Checklist</span>
        </h1>
        <p className="mt-2 max-w-4xl text-sm leading-relaxed text-text-muted">
          A transparent weighted checklist for early planning conversations. It does not certify a provider, facility, legal position, privacy compliance, security posture, or procurement eligibility.
        </p>
      </header>

      <section className="glass-card rounded-2xl border border-border/80 p-6 shadow-xl sm:p-8">
        <div className="flex flex-col justify-between gap-4 border-b border-borderSubtle pb-5 sm:flex-row sm:items-center">
          <div>
            <div className="inline-flex items-center gap-1.5 font-mono text-[11px] text-gold">
              <Sliders className="h-3.5 w-3.5" /> INTERACTIVE HEURISTIC
            </div>
            <h2 className="mt-1 text-lg font-bold text-text-main">Documented controls present</h2>
            <p className="mt-1 text-xs text-text-muted">Only select a control when you have evidence that supports it.</p>
          </div>
          <div className="rounded-xl border border-borderSubtle bg-surface px-5 py-3 text-right">
            <div className="font-mono text-[10px] uppercase text-text-subtle">Checklist coverage</div>
            <div className="text-3xl font-black text-aurora">{score}<span className="text-xs font-normal text-text-subtle">/100</span></div>
          </div>
        </div>

        <div className="mt-6 grid gap-3 md:grid-cols-2">
          {criteria.map((criterion) => {
            const selected = answers[criterion.id];
            return (
              <button
                key={criterion.id}
                type="button"
                aria-pressed={selected}
                onClick={() => setAnswers((current) => ({ ...current, [criterion.id]: !current[criterion.id] }))}
                className={`flex items-start justify-between gap-4 rounded-xl border p-4 text-left transition-colors ${selected ? "border-primary/40 bg-primary/10" : "border-borderSubtle bg-surface hover:border-border"}`}
              >
                <div>
                  <div className="flex items-center gap-2 text-sm font-bold text-text-main">
                    {criterion.label}
                    <span className="font-mono text-[10px] font-normal text-aurora">{criterion.weight} pts</span>
                  </div>
                  <p className="mt-1 text-xs leading-relaxed text-text-muted">{criterion.description}</p>
                </div>
                {selected ? <CheckCircle2 className="h-5 w-5 shrink-0 text-aurora" /> : <XCircle className="h-5 w-5 shrink-0 text-text-subtle" />}
              </button>
            );
          })}
        </div>
      </section>

      <section className="grid gap-4 md:grid-cols-2">
        <div className="rounded-2xl border border-borderSubtle bg-surface p-5">
          <div className="flex items-center gap-2 font-bold text-text-main"><Info className="h-4 w-4 text-aurora" /> How to use the result</div>
          <p className="mt-2 text-xs leading-relaxed text-text-muted">Treat unchecked items as diligence questions. Attach evidence, name the accountable owner, record the review date, and reassess when architecture or providers change.</p>
        </div>
        <div className="rounded-2xl border border-gold/30 bg-gold/5 p-5">
          <div className="flex items-center gap-2 font-bold text-text-main"><ShieldAlert className="h-4 w-4 text-gold" /> Limits</div>
          <p className="mt-2 text-xs leading-relaxed text-text-muted">The score is a user-entered coverage total, not a risk probability or independent verification. Legal exposure can depend on facts beyond ownership or physical location.</p>
        </div>
      </section>
    </div>
  );
}
