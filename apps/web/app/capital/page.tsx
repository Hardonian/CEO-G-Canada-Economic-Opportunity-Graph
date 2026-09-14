"use client";

import { useMemo, useState } from "react";
import { Calculator, CircleDollarSign, ShieldAlert } from "lucide-react";

function clamp(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function formatCad(value: number) {
  return new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    maximumFractionDigits: 0,
  }).format(value);
}

export default function CapitalPage() {
  const [capexMillions, setCapexMillions] = useState(500);
  const [seniorDebt, setSeniorDebt] = useState(45);
  const [sponsorEquity, setSponsorEquity] = useState(35);
  const [otherCapital, setOtherCapital] = useState(20);

  const model = useMemo(() => {
    const capex = clamp(capexMillions, 0, 1_000_000) * 1_000_000;
    const percentages = [seniorDebt, sponsorEquity, otherCapital].map((value) => clamp(value, 0, 100));
    const totalPercent = percentages.reduce((total, value) => total + value, 0);
    return {
      capex,
      totalPercent,
      allocations: [
        { label: "Senior debt assumption", percent: percentages[0], value: capex * percentages[0] / 100 },
        { label: "Sponsor equity assumption", percent: percentages[1], value: capex * percentages[1] / 100 },
        { label: "Other capital assumption", percent: percentages[2], value: capex * percentages[2] / 100 },
      ],
    };
  }, [capexMillions, seniorDebt, sponsorEquity, otherCapital]);

  const controls = [
    { label: "Senior debt", value: seniorDebt, setValue: setSeniorDebt },
    { label: "Sponsor equity", value: sponsorEquity, setValue: setSponsorEquity },
    { label: "Other capital", value: otherCapital, setValue: setOtherCapital },
  ];

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <header className="border-b border-border/80 pb-6">
        <div className="mb-3 inline-flex items-center gap-1.5 rounded-full border border-primary/40 bg-card px-3 py-1 font-mono text-[11px] text-aurora">
          <Calculator className="h-3.5 w-3.5" /> CAPITAL STACK SCENARIO
        </div>
        <h1 className="text-2xl font-black tracking-tight text-text-main sm:text-4xl">
          Transparent <span className="text-aurora">Capital Arithmetic</span>
        </h1>
        <p className="mt-2 max-w-3xl text-sm leading-relaxed text-text-muted">
          Build a simple capital-stack scenario from your own assumptions. This calculator does not determine program eligibility, financing availability, tax treatment, cost of capital, or expected returns.
        </p>
      </header>

      <div className="grid gap-6 lg:grid-cols-[0.9fr_1.1fr]">
        <section className="glass-card space-y-6 rounded-2xl border border-border/80 p-6">
          <label className="block text-xs font-mono text-text-muted">
            Total project CAPEX (CAD millions)
            <input
              type="number"
              min="0"
              max="1000000"
              step="10"
              value={capexMillions}
              onChange={(event) => setCapexMillions(Number(event.target.value) || 0)}
              className="mt-2 w-full rounded-xl border border-border bg-surface px-3 py-2 text-base font-bold text-text-main outline-none focus:border-aurora"
            />
          </label>

          {controls.map((control) => (
            <label key={control.label} className="block text-xs font-mono text-text-muted">
              <span className="flex justify-between"><span>{control.label}</span><strong className="text-text-main">{control.value}%</strong></span>
              <input
                type="range"
                min="0"
                max="100"
                step="1"
                value={control.value}
                onChange={(event) => control.setValue(Number(event.target.value))}
                className="mt-2 w-full accent-emerald-400"
              />
            </label>
          ))}
        </section>

        <section className="glass-card space-y-5 rounded-2xl border border-border/80 p-6">
          <div className="flex items-start justify-between gap-4 border-b border-borderSubtle pb-4">
            <div>
              <div className="text-[10px] font-mono uppercase text-text-subtle">Scenario CAPEX</div>
              <div className="mt-1 text-2xl font-black text-gold">{formatCad(model.capex)}</div>
            </div>
            <CircleDollarSign className="h-7 w-7 text-aurora" />
          </div>

          {model.allocations.map((allocation) => (
            <div key={allocation.label} className="rounded-xl border border-borderSubtle bg-surface p-4">
              <div className="flex items-center justify-between gap-3 text-xs">
                <span className="text-text-muted">{allocation.label}</span>
                <span className="font-mono font-bold text-text-main">{allocation.percent}%</span>
              </div>
              <div className="mt-1 text-lg font-black text-aurora">{formatCad(allocation.value)}</div>
            </div>
          ))}

          <div aria-live="polite" className={`rounded-xl border p-4 text-xs ${model.totalPercent === 100 ? "border-primary/40 bg-primary/10 text-aurora" : "border-gold/40 bg-gold/10 text-gold"}`}>
            Allocation total: <strong>{model.totalPercent}%</strong>. {model.totalPercent === 100 ? "The arithmetic balances to total CAPEX." : "Adjust assumptions until the stack totals 100%."}
          </div>
        </section>
      </div>

      <div className="flex gap-3 rounded-2xl border border-borderSubtle bg-surface p-5 text-xs leading-relaxed text-text-subtle">
        <ShieldAlert className="h-5 w-5 shrink-0 text-gold" />
        <p>All outputs are arithmetic scenarios only. Validate financing terms, incentives, guarantees, eligibility, tax consequences, and current law with the responsible institution and qualified advisers.</p>
      </div>
    </div>
  );
}
