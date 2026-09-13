"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import {
  Activity,
  AlertTriangle,
  ArrowUpRight,
  BarChart3,
  Briefcase,
  CheckCircle2,
  ChevronRight,
  Gauge,
  Info,
  Landmark,
  LockKeyhole,
  Radio,
  Scale,
  ShieldCheck,
  SlidersHorizontal,
  Target,
  TrendingUp,
} from "lucide-react";
import type { LifecycleStage, Project } from "@/lib/types";

type Role = "government" | "investor";
type ScenarioId = "reference" | "delivery" | "stress" | "custom";
type Horizon = 3 | 5 | 10;

interface Scenario {
  id: ScenarioId;
  name: string;
  shortName: string;
  description: string;
  probabilityShift: number;
  costPressure: number;
  publicSupportShare: number;
  uncertainty: number;
  tone: "neutral" | "positive" | "warning" | "custom";
}

interface CustomAssumptions {
  probabilityShiftPp: number;
  costPressurePct: number;
  publicSupportPct: number;
  uncertaintyPp: number;
}

interface ModelledProject {
  project: Project;
  probability: number;
  low: number;
  high: number;
  confidence: number;
  adjustedCapex: number;
  priority: number;
}

interface ScenarioSummary {
  scenario: Scenario;
  modelled: ModelledProject[];
  adjustedCapex: number;
  advancingCapital: number;
  lowCapital: number;
  highCapital: number;
  expectedAssets: number;
  delayedCapital: number;
  publicCapitalAssumption: number;
  capitalPerPublicDollar: number;
  sovereigntyAlignment: number;
  evidenceConfidence: number;
}

interface ExposureRow {
  label: string;
  capex: number;
  share: number;
}

interface Recommendation {
  eyebrow: string;
  title: string;
  body: string;
  project?: Project;
  tone: "positive" | "warning" | "neutral";
}

const REFERENCE_SCENARIOS: Scenario[] = [
  {
    id: "reference",
    name: "Reference path",
    shortName: "Reference",
    description: "Current stage and score signals, with no directional delivery adjustment.",
    probabilityShift: 0,
    costPressure: 0,
    publicSupportShare: 0.12,
    uncertainty: 0.07,
    tone: "neutral",
  },
  {
    id: "delivery",
    name: "Coordinated delivery",
    shortName: "Delivery push",
    description: "Faster decisions and coordinated public support, with modest execution cost pressure.",
    probabilityShift: 0.1,
    costPressure: 0.05,
    publicSupportShare: 0.2,
    uncertainty: 0.06,
    tone: "positive",
  },
  {
    id: "stress",
    name: "Capital & schedule stress",
    shortName: "Downside",
    description: "Higher costs, slower milestones, and a wider planning range.",
    probabilityShift: -0.14,
    costPressure: 0.18,
    publicSupportShare: 0.12,
    uncertainty: 0.13,
    tone: "warning",
  },
];

const STAGE_PROBABILITY: Record<LifecycleStage, number> = {
  DISCOVERED: 0.12,
  ANNOUNCED: 0.2,
  REFERRED: 0.28,
  EARLY_DEVELOPMENT: 0.3,
  FEASIBILITY: 0.4,
  FINANCING: 0.54,
  ENVIRONMENTAL_REVIEW: 0.48,
  PERMITTING: 0.58,
  PROCUREMENT: 0.7,
  FID_LIKELY: 0.78,
  FID: 0.88,
  CONSTRUCTION: 0.93,
  COMMISSIONING: 0.96,
  OPERATING: 0.98,
  DELAYED: 0.22,
  PAUSED: 0.12,
  CANCELLED: 0.03,
};

const EVIDENCE_CONFIDENCE: Record<Project["confidence"], number> = {
  VERIFIED: 0.92,
  SUPPORTED: 0.78,
  REPORTED: 0.68,
  INFERRED: 0.6,
  CONFLICTED: 0.42,
  UNKNOWN: 0.34,
  STALE: 0.3,
  RETRACTED: 0.05,
};

const HORIZON_SHIFT: Record<Horizon, number> = {
  3: -0.1,
  5: 0,
  10: 0.09,
};

const clamp = (value: number, min = 0, max = 1) =>
  Math.min(max, Math.max(min, value));

const score = (project: Project, key: string, fallback = 50) =>
  project.scores?.[key] ?? fallback;

const formatCad = (value: number) => {
  const sign = value < 0 ? "-" : "";
  const absolute = Math.abs(value);
  if (absolute >= 1_000_000_000) {
    return `${sign}$${(absolute / 1_000_000_000).toFixed(absolute >= 10_000_000_000 ? 1 : 2)}B`;
  }
  if (absolute >= 1_000_000) return `${sign}$${(absolute / 1_000_000).toFixed(0)}M`;
  return `${sign}$${absolute.toLocaleString("en-CA", { maximumFractionDigits: 0 })}`;
};

const formatPercent = (value: number) => `${Math.round(value * 100)}%`;

const formatSignedPp = (value: number) => {
  const rounded = Math.round(value * 100);
  if (rounded === 0) return "0 pp";
  return `${rounded > 0 ? "+" : ""}${rounded} pp`;
};

const formatSnapshotDate = (value: string) =>
  new Intl.DateTimeFormat("en-CA", {
    day: "numeric",
    month: "short",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(value));

function rolePriority(project: Project, probability: number, role: Role) {
  const buildability = score(project, "buildability");
  const investability = score(project, "investability");
  const supplierability = score(project, "supplierability");
  const strategicity = score(project, "strategicity");

  if (role === "government") {
    return (
      strategicity * 0.38 +
      buildability * 0.24 +
      supplierability * 0.2 +
      investability * 0.08 +
      probability * 100 * 0.1
    );
  }

  return (
    investability * 0.4 +
    buildability * 0.24 +
    supplierability * 0.14 +
    strategicity * 0.1 +
    probability * 100 * 0.12
  );
}

function modelProject(
  project: Project,
  scenario: Scenario,
  horizon: Horizon,
  role: Role,
): ModelledProject {
  const buildability = score(project, "buildability");
  const investability = score(project, "investability");
  const strategicity = score(project, "strategicity");
  const confidence = EVIDENCE_CONFIDENCE[project.confidence];
  const scoreAdjustment =
    (buildability - 50) * 0.0024 +
    (investability - 50) * 0.0013;
  const strategicSensitivity = 0.7 + strategicity / 300;
  const probability = clamp(
    STAGE_PROBABILITY[project.current_stage] +
      scoreAdjustment +
      HORIZON_SHIFT[horizon] +
      scenario.probabilityShift * strategicSensitivity,
    0.03,
    0.98,
  );
  const evidenceBand = (1 - confidence) * 0.22;
  const readinessBand = (100 - buildability) * 0.00045;
  const totalBand = scenario.uncertainty + evidenceBand + readinessBand;

  return {
    project,
    probability,
    low: clamp(probability - totalBand, 0.01, 0.98),
    high: clamp(probability + totalBand, 0.03, 0.99),
    confidence,
    adjustedCapex: project.capex_cad * (1 + scenario.costPressure),
    priority: rolePriority(project, probability, role),
  };
}

function summarizeScenario(
  projects: Project[],
  scenario: Scenario,
  horizon: Horizon,
  role: Role,
): ScenarioSummary {
  const modelled = projects.map((project) =>
    modelProject(project, scenario, horizon, role),
  );
  const adjustedCapex = modelled.reduce((total, item) => total + item.adjustedCapex, 0);
  const advancingCapital = modelled.reduce(
    (total, item) => total + item.adjustedCapex * item.probability,
    0,
  );
  const lowCapital = modelled.reduce(
    (total, item) => total + item.adjustedCapex * item.low,
    0,
  );
  const highCapital = modelled.reduce(
    (total, item) => total + item.adjustedCapex * item.high,
    0,
  );
  const publicCapitalAssumption = adjustedCapex * scenario.publicSupportShare;
  const weighted = (key: "confidence" | "sovereignty") => {
    if (adjustedCapex === 0) return 0;
    return (
      modelled.reduce((total, item) => {
        const value =
          key === "confidence"
            ? item.confidence * 100
            : score(item.project, "strategicity") * 0.65 +
              score(item.project, "supplierability") * 0.35;
        return total + value * item.adjustedCapex;
      }, 0) / adjustedCapex
    );
  };

  return {
    scenario,
    modelled,
    adjustedCapex,
    advancingCapital,
    lowCapital,
    highCapital,
    expectedAssets: modelled.reduce((total, item) => total + item.probability, 0),
    delayedCapital: Math.max(0, adjustedCapex - advancingCapital),
    publicCapitalAssumption,
    capitalPerPublicDollar:
      publicCapitalAssumption > 0 ? advancingCapital / publicCapitalAssumption : 0,
    sovereigntyAlignment: weighted("sovereignty"),
    evidenceConfidence: weighted("confidence"),
  };
}

function buildExposure(
  projects: Project[],
  getLabel: (project: Project) => string,
): ExposureRow[] {
  const total = projects.reduce((sum, project) => sum + project.capex_cad, 0);
  const grouped = new Map<string, number>();
  projects.forEach((project) => {
    const label = getLabel(project);
    grouped.set(label, (grouped.get(label) ?? 0) + project.capex_cad);
  });

  return Array.from(grouped.entries())
    .map(([label, capex]) => ({
      label,
      capex,
      share: total > 0 ? capex / total : 0,
    }))
    .sort((a, b) => b.capex - a.capex);
}

function pipelineBucket(stage: LifecycleStage) {
  if (["DISCOVERED", "ANNOUNCED", "REFERRED", "EARLY_DEVELOPMENT", "FEASIBILITY"].includes(stage)) {
    return "Formation";
  }
  if (["FINANCING", "ENVIRONMENTAL_REVIEW", "PERMITTING", "PROCUREMENT", "FID_LIKELY", "FID"].includes(stage)) {
    return "Decision window";
  }
  if (["CONSTRUCTION", "COMMISSIONING", "OPERATING"].includes(stage)) return "Delivery";
  return "At risk";
}

function getAction(project: Project, role: Role) {
  if (role === "government") {
    if (score(project, "strategicity") - score(project, "buildability") >= 25) {
      return "Resolve delivery blockers";
    }
    if (project.current_stage === "PROCUREMENT") return "Coordinate demand signal";
    if (project.current_stage === "CONSTRUCTION") return "Protect critical path";
    return "Align next milestone";
  }

  if (score(project, "investability") >= 75) return "Advance diligence";
  if (["PROCUREMENT", "CONSTRUCTION"].includes(project.current_stage)) return "Map supplier entry";
  return "Watch milestone evidence";
}

function scenarioToneClasses(tone: Scenario["tone"], active: boolean) {
  if (!active) return "border-border/80 bg-card/40 hover:border-primary/40";
  if (tone === "warning") return "border-gold/60 bg-gold/10 shadow-[0_0_28px_rgba(245,158,11,0.08)]";
  if (tone === "custom") return "border-accent-purple/60 bg-accent-purple/10 shadow-[0_0_28px_rgba(168,85,247,0.08)]";
  return "border-primary/60 bg-primary/10 shadow-[0_0_28px_rgba(0,245,160,0.08)]";
}

function ScenarioCard({
  summary,
  active,
  onSelect,
}: {
  summary: ScenarioSummary;
  active: boolean;
  onSelect: () => void;
}) {
  const { scenario } = summary;
  return (
    <button
      type="button"
      onClick={onSelect}
      aria-pressed={active}
      className={`rounded-2xl border p-4 text-left transition-all ${scenarioToneClasses(scenario.tone, active)}`}
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-[10px] font-mono uppercase tracking-[0.16em] text-text-subtle">
            {scenario.shortName}
          </div>
          <h3 className="mt-1 text-sm font-bold text-text-main">{scenario.name}</h3>
        </div>
        <span
          className={`mt-0.5 h-2.5 w-2.5 rounded-full ${
            active ? "bg-aurora shadow-[0_0_10px_#00F5A0]" : "bg-border"
          }`}
        />
      </div>
      <p className="mt-2 min-h-[48px] text-[11px] leading-relaxed text-text-muted">
        {scenario.description}
      </p>
      <div className="mt-4 border-t border-borderSubtle pt-3">
        <div className="font-tabular text-xl font-black text-text-main">
          {formatCad(summary.advancingCapital)}
        </div>
        <div className="mt-0.5 text-[10px] font-mono uppercase text-text-subtle">
          capital associated with advancement
        </div>
        <div className="mt-3 flex items-center justify-between text-[10px] font-mono">
          <span className="text-text-subtle">Planning range</span>
          <span className="text-text-muted">
            {formatCad(summary.lowCapital)}–{formatCad(summary.highCapital)}
          </span>
        </div>
      </div>
    </button>
  );
}

function RangeBar({ low, base, high }: { low: number; base: number; high: number }) {
  const lowPct = clamp(low) * 100;
  const basePct = clamp(base) * 100;
  const highPct = clamp(high) * 100;
  return (
    <div className="relative h-2 w-full rounded-full bg-surface" aria-hidden="true">
      <div
        className="absolute top-0 h-2 rounded-full bg-primary/30"
        style={{ left: `${lowPct}%`, width: `${Math.max(1, highPct - lowPct)}%` }}
      />
      <div
        className="absolute -top-0.5 h-3 w-1 rounded-full bg-aurora shadow-[0_0_7px_rgba(0,245,160,0.8)]"
        style={{ left: `calc(${basePct}% - 2px)` }}
      />
    </div>
  );
}

export default function DecisionPlanningWorkbench({ projects }: { projects: Project[] }) {
  const [role, setRole] = useState<Role>("government");
  const [horizon, setHorizon] = useState<Horizon>(5);
  const [activeScenarioId, setActiveScenarioId] = useState<ScenarioId>("delivery");
  const [selectedIds, setSelectedIds] = useState<Set<string>>(
    () => new Set(projects.map((project) => project.id)),
  );
  const [custom, setCustom] = useState<CustomAssumptions>({
    probabilityShiftPp: 4,
    costPressurePct: 8,
    publicSupportPct: 16,
    uncertaintyPp: 9,
  });

  const customScenario = useMemo<Scenario>(
    () => ({
      id: "custom",
      name: "Custom planning case",
      shortName: "Your case",
      description: "User-defined assumptions for a mandate-specific sensitivity test.",
      probabilityShift: custom.probabilityShiftPp / 100,
      costPressure: custom.costPressurePct / 100,
      publicSupportShare: custom.publicSupportPct / 100,
      uncertainty: custom.uncertaintyPp / 100,
      tone: "custom",
    }),
    [custom],
  );

  const scenarios = useMemo(
    () => [...REFERENCE_SCENARIOS, customScenario],
    [customScenario],
  );
  const selectedProjects = useMemo(
    () => projects.filter((project) => selectedIds.has(project.id)),
    [projects, selectedIds],
  );
  const summaries = useMemo(
    () =>
      scenarios.map((scenario) =>
        summarizeScenario(selectedProjects, scenario, horizon, role),
      ),
    [horizon, role, scenarios, selectedProjects],
  );
  const activeSummary =
    summaries.find((summary) => summary.scenario.id === activeScenarioId) ?? summaries[0];
  const referenceSummary = summaries[0];

  const sectorExposure = useMemo(
    () => buildExposure(selectedProjects, (project) => project.sector),
    [selectedProjects],
  );
  const provinceExposure = useMemo(
    () => buildExposure(selectedProjects, (project) => project.province),
    [selectedProjects],
  );
  const pipelineExposure = useMemo(
    () => buildExposure(selectedProjects, (project) => pipelineBucket(project.current_stage)),
    [selectedProjects],
  );
  const sectorHhi = sectorExposure.reduce((total, row) => total + row.share ** 2, 0);
  const newestUpdate = projects.reduce(
    (latest, project) =>
      project.last_meaningful_update > latest ? project.last_meaningful_update : latest,
    projects[0]?.last_meaningful_update ?? "2026-01-01T00:00:00Z",
  );
  const scoredRecordCount = selectedProjects.filter(
    (project) =>
      project.capex_cad > 0 &&
      project.scores &&
      ["buildability", "investability", "supplierability", "strategicity"].every(
        (key) => typeof project.scores?.[key] === "number",
      ),
  ).length;
  const dataCoverage = selectedProjects.length > 0 ? scoredRecordCount / selectedProjects.length : 0;

  const allProjectModels = useMemo(() => {
    const activeScenario =
      scenarios.find((scenario) => scenario.id === activeScenarioId) ?? scenarios[0];
    return projects
      .map((project) => modelProject(project, activeScenario, horizon, role))
      .sort((a, b) => b.priority - a.priority);
  }, [activeScenarioId, horizon, projects, role, scenarios]);

  const recommendations = useMemo<Recommendation[]>(() => {
    if (selectedProjects.length === 0) return [];
    const ranked = [...activeSummary.modelled].sort((a, b) => b.priority - a.priority);
    const readinessGap = [...selectedProjects].sort(
      (a, b) =>
        score(b, "strategicity") -
        score(b, "buildability") -
        (score(a, "strategicity") - score(a, "buildability")),
    )[0];
    const topSector = sectorExposure[0];
    const topProvince = provinceExposure[0];
    const top = ranked[0];

    if (role === "government") {
      return [
        {
          eyebrow: "Delivery intervention",
          title: `Close the readiness gap at ${readinessGap.name}`,
          body: `Strategicity is ${score(readinessGap, "strategicity").toFixed(0)}/100 versus buildability at ${score(readinessGap, "buildability").toFixed(0)}/100. Convene owners around the next permission, financing, or procurement milestone before adding new capital assumptions.`,
          project: readinessGap,
          tone: "warning",
        },
        {
          eyebrow: "Portfolio sequence",
          title: `Put ${top.project.name} in the first decision tranche`,
          body: `It ranks highest under the public-value lens, combining strategicity, supplier potential, readiness, and a ${formatPercent(top.probability)} modelled milestone likelihood in the selected case.`,
          project: top.project,
          tone: "positive",
        },
        {
          eyebrow: "Concentration guardrail",
          title: `Stress-test ${topProvince?.label ?? "regional"} and ${topSector?.label ?? "sector"} dependency together`,
          body: `${topProvince?.label ?? "The leading region"} represents ${formatPercent(topProvince?.share ?? 0)} of selected CAPEX and ${topSector?.label ?? "the leading sector"} represents ${formatPercent(topSector?.share ?? 0)}. Test shared labour, grid, permitting, and supply-chain constraints before portfolio approval.`,
          tone: "neutral",
        },
      ];
    }

    const stress = summarizeScenario(selectedProjects, REFERENCE_SCENARIOS[2], horizon, role);
    const downsideGap = referenceSummary.advancingCapital - stress.advancingCapital;
    return [
      {
        eyebrow: "Diligence priority",
        title: `Lead with ${top.project.name}`,
        body: `It ranks highest under the institutional lens, balancing investability, buildability, evidence confidence, and near-term milestone posture. Validate commercial terms and counterparties before treating the score as an allocation signal.`,
        project: top.project,
        tone: "positive",
      },
      {
        eyebrow: "Downside budget",
        title: `Reserve for a ${formatCad(Math.max(0, downsideGap))} reference-to-stress gap`,
        body: "The gap is a deterministic sensitivity result, not value-at-risk. Use it to size diligence, contingencies, and staged commitments across the selected portfolio.",
        tone: "warning",
      },
      {
        eyebrow: "Exposure limit",
        title: `Set a deliberate cap for ${topSector?.label ?? "the leading sector"}`,
        body: `${formatPercent(topSector?.share ?? 0)} of selected CAPEX sits in this sector. Pair any concentration limit with project-level milestone gates so nominal diversification does not hide correlated execution risk.`,
        tone: "neutral",
      },
    ];
  }, [activeSummary.modelled, horizon, provinceExposure, referenceSummary.advancingCapital, role, sectorExposure, selectedProjects]);

  const toggleProject = (id: string) => {
    setSelectedIds((current) => {
      const next = new Set(current);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const selectRoleLeaders = () => {
    const leaders = allProjectModels.slice(0, 5).map((item) => item.project.id);
    setSelectedIds(new Set(leaders));
  };

  const updateCustom = (key: keyof CustomAssumptions, value: number) => {
    setCustom((current) => ({ ...current, [key]: value }));
    setActiveScenarioId("custom");
  };

  const activeDelta = activeSummary.advancingCapital - referenceSummary.advancingCapital;

  return (
    <div className="mx-auto max-w-7xl space-y-8 px-4 py-8 sm:px-6 lg:px-8">
      <section className="relative overflow-hidden rounded-3xl border border-border/80 bg-[#07110c] px-5 py-7 shadow-2xl sm:px-8 sm:py-9">
        <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_78%_10%,rgba(0,245,160,0.13),transparent_32%),radial-gradient(circle_at_10%_100%,rgba(245,158,11,0.08),transparent_34%)]" />
        <div className="relative grid gap-7 lg:grid-cols-[1fr_360px] lg:items-end">
          <div>
            <div className="mb-4 inline-flex items-center gap-2 rounded-full border border-primary/40 bg-primary/10 px-3 py-1 text-[10px] font-mono uppercase tracking-[0.16em] text-aurora">
              <Radio className="h-3.5 w-3.5" />
              Decision intelligence / planning model v1.0
            </div>
            <h1 className="max-w-4xl text-3xl font-black tracking-tight text-text-main sm:text-5xl">
              National decision <span className="text-aurora">planning workbench</span>
            </h1>
            <p className="mt-4 max-w-3xl text-sm leading-6 text-text-muted">
              Turn the project graph into transparent, scenario-tested capital decisions. Compare delivery paths,
              quantify portfolio concentration and uncertainty, and build an auditable action queue for public or
              institutional mandates.
            </p>
            <div className="mt-5 flex flex-wrap items-center gap-2 text-[10px] font-mono">
              <span className="rounded-full border border-border bg-surface/80 px-2.5 py-1 text-text-muted">
                {projects.length} bundled CEGS-format records
              </span>
              <span className="rounded-full border border-border bg-surface/80 px-2.5 py-1 text-text-muted">
                Data through {formatSnapshotDate(newestUpdate)}
              </span>
              <span className="rounded-full border border-gold/30 bg-gold/10 px-2.5 py-1 text-gold">
                Indicative model — not a live forecast
              </span>
            </div>
          </div>

          <div className="rounded-2xl border border-primary/30 bg-card/70 p-5 backdrop-blur">
            <div className="flex items-start gap-3">
              <div className="rounded-xl border border-primary/30 bg-primary/10 p-2 text-aurora">
                <LockKeyhole className="h-5 w-5" />
              </div>
              <div>
                <div className="text-xs font-bold text-text-main">Sovereign calculation boundary</div>
                <p className="mt-1 text-[11px] leading-relaxed text-text-muted">
                  Scenario math and portfolio selections run in this page. This workbench sends no portfolio inputs
                  to a third-party AI or inference API.
                </p>
              </div>
            </div>
            <div className="mt-4 grid grid-cols-3 gap-2 border-t border-borderSubtle pt-4 text-center font-mono text-[9px] uppercase text-text-subtle">
              <span>Deterministic</span>
              <span>Explainable</span>
              <span>Export-safe</span>
            </div>
          </div>
        </div>
      </section>

      <section className="glass-panel rounded-2xl border border-border/80 p-4 shadow-xl">
        <div className="grid gap-4 lg:grid-cols-[1fr_auto_auto] lg:items-center">
          <div className="flex items-center gap-3">
            <div className="rounded-xl border border-border bg-surface p-2 text-aurora">
              {role === "government" ? <Landmark className="h-4 w-4" /> : <Briefcase className="h-4 w-4" />}
            </div>
            <div>
              <div className="text-xs font-bold text-text-main">Decision lens</div>
              <div className="text-[10px] text-text-subtle">
                Changes prioritization—not the underlying project data.
              </div>
            </div>
          </div>

          <div className="inline-flex rounded-xl border border-borderSubtle bg-surface p-1">
            <button
              type="button"
              aria-pressed={role === "government"}
              onClick={() => setRole("government")}
              className={`rounded-lg px-3 py-2 text-[11px] font-semibold transition-all ${
                role === "government" ? "bg-card text-aurora shadow" : "text-text-muted hover:text-text-main"
              }`}
            >
              Government mandate
            </button>
            <button
              type="button"
              aria-pressed={role === "investor"}
              onClick={() => setRole("investor")}
              className={`rounded-lg px-3 py-2 text-[11px] font-semibold transition-all ${
                role === "investor" ? "bg-card text-aurora shadow" : "text-text-muted hover:text-text-main"
              }`}
            >
              Institutional investor
            </button>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-[10px] font-mono uppercase text-text-subtle">Horizon</span>
            <div className="inline-flex rounded-xl border border-borderSubtle bg-surface p-1">
              {([3, 5, 10] as Horizon[]).map((year) => (
                <button
                  type="button"
                  key={year}
                  aria-pressed={horizon === year}
                  onClick={() => setHorizon(year)}
                  className={`rounded-lg px-3 py-1.5 text-[11px] font-mono transition-all ${
                    horizon === year ? "bg-card text-aurora" : "text-text-muted hover:text-text-main"
                  }`}
                >
                  {year}Y
                </button>
              ))}
            </div>
          </div>
        </div>
      </section>

      {selectedProjects.length === 0 ? (
        <section className="rounded-2xl border border-gold/30 bg-gold/10 p-6 text-center">
          <AlertTriangle className="mx-auto h-6 w-6 text-gold" />
          <h2 className="mt-2 text-sm font-bold text-text-main">No assets in the planning portfolio</h2>
          <p className="mt-1 text-xs text-text-muted">Select assets in the decision queue or restore the full portfolio.</p>
          <button
            type="button"
            onClick={() => setSelectedIds(new Set(projects.map((project) => project.id)))}
            className="mt-4 rounded-xl bg-gold px-4 py-2 text-xs font-bold text-background"
          >
            Restore all assets
          </button>
        </section>
      ) : (
        <>
          <section aria-labelledby="active-outlook-heading" className="space-y-4">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <div className="text-[10px] font-mono uppercase tracking-[0.15em] text-aurora">Selected outlook</div>
                <h2 id="active-outlook-heading" className="mt-1 text-xl font-black text-text-main">
                  {activeSummary.scenario.name} · {horizon}-year horizon
                </h2>
              </div>
              <div className="inline-flex items-center gap-2 text-[10px] font-mono text-text-muted">
                <ShieldCheck className="h-4 w-4 text-aurora" />
                Evidence confidence {activeSummary.evidenceConfidence.toFixed(0)}/100 · score coverage {formatPercent(dataCoverage)}
              </div>
            </div>

            <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
              <div className="glass-card rounded-2xl border border-border/80 p-5">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-mono uppercase text-text-subtle">Capital advancing</span>
                  <TrendingUp className="h-4 w-4 text-aurora" />
                </div>
                <div className="mt-2 font-tabular text-2xl font-black text-text-main">
                  {formatCad(activeSummary.advancingCapital)}
                </div>
                <div className={`mt-1 text-[10px] font-mono ${activeDelta >= 0 ? "text-aurora" : "text-gold"}`}>
                  {activeDelta >= 0 ? "+" : ""}{formatCad(activeDelta)} vs reference
                </div>
                <div className="mt-4">
                  <RangeBar
                    low={activeSummary.lowCapital / activeSummary.adjustedCapex}
                    base={activeSummary.advancingCapital / activeSummary.adjustedCapex}
                    high={activeSummary.highCapital / activeSummary.adjustedCapex}
                  />
                </div>
                <div className="mt-2 text-[9px] font-mono text-text-subtle">
                  Range {formatCad(activeSummary.lowCapital)}–{formatCad(activeSummary.highCapital)}
                </div>
              </div>

              <div className="glass-card rounded-2xl border border-border/80 p-5">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-mono uppercase text-text-subtle">Assets advancing</span>
                  <Activity className="h-4 w-4 text-aurora" />
                </div>
                <div className="mt-2 font-tabular text-2xl font-black text-text-main">
                  {activeSummary.expectedAssets.toFixed(1)}
                  <span className="text-sm font-medium text-text-subtle"> / {selectedProjects.length}</span>
                </div>
                <p className="mt-2 text-[10px] leading-relaxed text-text-muted">
                  Probability-weighted assets reaching their next material capital milestone.
                </p>
              </div>

              <div className="glass-card rounded-2xl border border-border/80 p-5">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-mono uppercase text-text-subtle">Capital at risk / deferred</span>
                  <Gauge className="h-4 w-4 text-gold" />
                </div>
                <div className="mt-2 font-tabular text-2xl font-black text-text-main">
                  {formatCad(activeSummary.delayedCapital)}
                </div>
                <p className="mt-2 text-[10px] leading-relaxed text-text-muted">
                  Cost-adjusted portfolio less probability-weighted advancing capital; not expected loss.
                </p>
              </div>

              <div className="glass-card rounded-2xl border border-border/80 p-5">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-mono uppercase text-text-subtle">Sovereignty alignment</span>
                  <ShieldCheck className="h-4 w-4 text-aurora" />
                </div>
                <div className="mt-2 font-tabular text-2xl font-black text-text-main">
                  {activeSummary.sovereigntyAlignment.toFixed(0)}<span className="text-sm text-text-subtle">/100</span>
                </div>
                <p className="mt-2 text-[10px] leading-relaxed text-text-muted">
                  CAPEX-weighted strategicity (65%) and supplierability (35%); a planning index, not certification.
                </p>
              </div>
            </div>
          </section>

          <section aria-labelledby="scenario-heading" className="space-y-4">
            <div className="flex items-center justify-between gap-3">
              <div>
                <div className="text-[10px] font-mono uppercase tracking-[0.15em] text-text-subtle">Sensitivity analysis</div>
                <h2 id="scenario-heading" className="mt-1 text-lg font-black text-text-main">Compare decision paths</h2>
              </div>
              <div className="hidden items-center gap-1.5 text-[10px] text-text-subtle sm:flex">
                <Info className="h-3.5 w-3.5" /> Select a case to drive the rest of the workbench
              </div>
            </div>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
              {summaries.map((summary) => (
                <ScenarioCard
                  key={summary.scenario.id}
                  summary={summary}
                  active={activeScenarioId === summary.scenario.id}
                  onSelect={() => setActiveScenarioId(summary.scenario.id)}
                />
              ))}
            </div>

            <div className="grid gap-5 lg:grid-cols-12">
              <div className="overflow-hidden rounded-2xl border border-border/80 bg-card/40 lg:col-span-8">
                <div className="border-b border-borderSubtle px-5 py-4">
                  <h3 className="flex items-center gap-2 text-xs font-bold text-text-main">
                    <Scale className="h-4 w-4 text-aurora" /> Scenario comparison matrix
                  </h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full min-w-[680px] text-left text-xs">
                    <thead className="bg-[#08130E] font-mono text-[9px] uppercase tracking-wider text-text-subtle">
                      <tr>
                        <th className="px-5 py-3">Case</th>
                        <th className="px-4 py-3 text-right">Likelihood shift</th>
                        <th className="px-4 py-3 text-right">Cost pressure</th>
                        <th className="px-4 py-3 text-right">Public support*</th>
                        <th className="px-4 py-3 text-right">Capital / public $</th>
                        <th className="px-5 py-3 text-right">At risk / deferred</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-borderSubtle">
                      {summaries.map((summary) => (
                        <tr
                          key={summary.scenario.id}
                          className={activeScenarioId === summary.scenario.id ? "bg-primary/5" : "hover:bg-surface/50"}
                        >
                          <td className="px-5 py-3.5 font-semibold text-text-main">{summary.scenario.shortName}</td>
                          <td className="px-4 py-3.5 text-right font-mono text-text-muted">
                            {formatSignedPp(summary.scenario.probabilityShift)}
                          </td>
                          <td className="px-4 py-3.5 text-right font-mono text-text-muted">
                            {formatPercent(summary.scenario.costPressure)}
                          </td>
                          <td className="px-4 py-3.5 text-right font-mono text-text-muted">
                            {formatCad(summary.publicCapitalAssumption)}
                          </td>
                          <td className="px-4 py-3.5 text-right font-mono font-bold text-aurora">
                            {summary.capitalPerPublicDollar.toFixed(1)}×
                          </td>
                          <td className="px-5 py-3.5 text-right font-mono text-text-main">
                            {formatCad(summary.delayedCapital)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="border-t border-borderSubtle px-5 py-3 text-[9px] leading-relaxed text-text-subtle">
                  * Public support is an editable portfolio-level assumption, not identified or committed government funding.
                  “Capital / public $” is advancing capital divided by that assumption; it is not an economic multiplier.
                </div>
              </div>

              <div className="rounded-2xl border border-accent-purple/30 bg-accent-purple/5 p-5 lg:col-span-4">
                <div className="flex items-center justify-between">
                  <h3 className="flex items-center gap-2 text-xs font-bold text-text-main">
                    <SlidersHorizontal className="h-4 w-4 text-accent-purple" /> Build your case
                  </h3>
                  <span className="rounded-full border border-accent-purple/30 px-2 py-0.5 text-[9px] font-mono text-accent-purple">
                    LOCAL
                  </span>
                </div>
                <div className="mt-5 space-y-4">
                  {[
                    {
                      key: "probabilityShiftPp" as const,
                      label: "Milestone likelihood shift",
                      min: -20,
                      max: 20,
                      suffix: " pp",
                    },
                    {
                      key: "costPressurePct" as const,
                      label: "CAPEX pressure",
                      min: -10,
                      max: 30,
                      suffix: "%",
                    },
                    {
                      key: "publicSupportPct" as const,
                      label: "Public support assumption",
                      min: 0,
                      max: 30,
                      suffix: "%",
                    },
                    {
                      key: "uncertaintyPp" as const,
                      label: "Base uncertainty buffer",
                      min: 3,
                      max: 20,
                      suffix: " pp",
                    },
                  ].map((control) => (
                    <label key={control.key} className="block">
                      <span className="flex items-center justify-between text-[10px] font-mono">
                        <span className="text-text-muted">{control.label}</span>
                        <output className="font-bold text-text-main">
                          {custom[control.key] > 0 && control.key === "probabilityShiftPp" ? "+" : ""}
                          {custom[control.key]}{control.suffix}
                        </output>
                      </span>
                      <input
                        type="range"
                        min={control.min}
                        max={control.max}
                        step={1}
                        value={custom[control.key]}
                        onChange={(event) => updateCustom(control.key, Number(event.target.value))}
                        className="mt-2 h-1.5 w-full cursor-pointer accent-purple-500"
                      />
                    </label>
                  ))}
                </div>
                <p className="mt-5 border-t border-accent-purple/20 pt-4 text-[9px] leading-relaxed text-text-subtle">
                  Levers apply uniformly across selected assets. They do not represent approved policy, committed financing,
                  or a calibrated probability forecast.
                </p>
              </div>
            </div>
          </section>

          <section aria-labelledby="exposure-heading" className="space-y-4">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
              <div>
                <div className="text-[10px] font-mono uppercase tracking-[0.15em] text-text-subtle">Portfolio construction</div>
                <h2 id="exposure-heading" className="mt-1 text-lg font-black text-text-main">Exposure & concentration map</h2>
              </div>
              <div className="flex flex-wrap items-center gap-2 text-[10px] font-mono">
                <span className="rounded-full border border-border bg-surface px-2.5 py-1 text-text-muted">
                  {selectedProjects.length} assets · {formatCad(selectedProjects.reduce((sum, project) => sum + project.capex_cad, 0))} reported CAPEX
                </span>
                <span className={`rounded-full border px-2.5 py-1 ${sectorHhi < 0.25 ? "border-primary/30 bg-primary/10 text-aurora" : "border-gold/30 bg-gold/10 text-gold"}`}>
                  Sector HHI {sectorHhi.toFixed(2)} · {sectorHhi < 0.25 ? "distributed" : sectorHhi < 0.4 ? "moderate" : "concentrated"}
                </span>
              </div>
            </div>

            <div className="grid gap-4 lg:grid-cols-3">
              {[
                { title: "Sector exposure", rows: sectorExposure, icon: BarChart3 },
                { title: "Regional exposure", rows: provinceExposure, icon: Landmark },
                { title: "Pipeline posture", rows: pipelineExposure, icon: Activity },
              ].map((group) => {
                const Icon = group.icon;
                return (
                  <div key={group.title} className="rounded-2xl border border-border/80 bg-card/40 p-5">
                    <h3 className="flex items-center gap-2 text-xs font-bold text-text-main">
                      <Icon className="h-4 w-4 text-aurora" /> {group.title}
                    </h3>
                    <div className="mt-4 space-y-3">
                      {group.rows.map((row) => (
                        <div key={row.label}>
                          <div className="mb-1.5 flex items-center justify-between gap-3 text-[10px]">
                            <span className="truncate text-text-muted">{row.label}</span>
                            <span className="shrink-0 font-mono text-text-main">
                              {formatCad(row.capex)} · {formatPercent(row.share)}
                            </span>
                          </div>
                          <div className="h-1.5 overflow-hidden rounded-full bg-surface">
                            <div
                              className="h-full rounded-full bg-gradient-to-r from-primary to-aurora-teal"
                              style={{ width: `${Math.max(2, row.share * 100)}%` }}
                            />
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          </section>

          <section aria-labelledby="recommendations-heading" className="space-y-4">
            <div>
              <div className="text-[10px] font-mono uppercase tracking-[0.15em] text-text-subtle">Action layer</div>
              <h2 id="recommendations-heading" className="mt-1 text-lg font-black text-text-main">
                Recommended next decisions
              </h2>
              <p className="mt-1 text-[11px] text-text-muted">
                Rules-based suggestions from the selected lens and scenario; each recommendation exposes its rationale.
              </p>
            </div>
            <div className="grid gap-4 lg:grid-cols-3">
              {recommendations.map((recommendation) => (
                <article key={recommendation.title} className="glass-card flex flex-col rounded-2xl border border-border/80 p-5">
                  <div className="flex items-center justify-between">
                    <span className={`text-[9px] font-mono uppercase tracking-wider ${recommendation.tone === "warning" ? "text-gold" : recommendation.tone === "positive" ? "text-aurora" : "text-text-subtle"}`}>
                      {recommendation.eyebrow}
                    </span>
                    {recommendation.tone === "warning" ? (
                      <AlertTriangle className="h-4 w-4 text-gold" />
                    ) : recommendation.tone === "positive" ? (
                      <Target className="h-4 w-4 text-aurora" />
                    ) : (
                      <Scale className="h-4 w-4 text-text-subtle" />
                    )}
                  </div>
                  <h3 className="mt-3 text-sm font-bold leading-snug text-text-main">{recommendation.title}</h3>
                  <p className="mt-2 flex-1 text-[11px] leading-relaxed text-text-muted">{recommendation.body}</p>
                  {recommendation.project && (
                    <Link
                      href={`/projects/${recommendation.project.slug}`}
                      className="mt-4 inline-flex items-center gap-1 border-t border-borderSubtle pt-3 text-[10px] font-semibold text-aurora"
                    >
                      Open evidence dossier <ChevronRight className="h-3.5 w-3.5" />
                    </Link>
                  )}
                </article>
              ))}
            </div>
          </section>
        </>
      )}

      <section aria-labelledby="queue-heading" className="overflow-hidden rounded-2xl border border-border/80 bg-card/40 shadow-xl">
        <div className="flex flex-col gap-3 border-b border-borderSubtle px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 id="queue-heading" className="flex items-center gap-2 text-sm font-bold text-text-main">
              <Target className="h-4 w-4 text-aurora" /> Decision queue
            </h2>
            <p className="mt-1 text-[10px] text-text-subtle">Ranked for the active lens. Checkboxes define the modelled portfolio.</p>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-[10px] font-mono">
            <button
              type="button"
              onClick={() => setSelectedIds(new Set(projects.map((project) => project.id)))}
              className="rounded-lg border border-border bg-surface px-2.5 py-1.5 text-text-muted hover:border-primary/40 hover:text-text-main"
            >
              All assets
            </button>
            <button
              type="button"
              onClick={selectRoleLeaders}
              className="rounded-lg border border-primary/30 bg-primary/10 px-2.5 py-1.5 text-aurora"
            >
              Top 5 by lens
            </button>
            <button
              type="button"
              onClick={() => setSelectedIds(new Set())}
              className="rounded-lg border border-border bg-surface px-2.5 py-1.5 text-text-muted hover:text-text-main"
            >
              Clear
            </button>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full min-w-[880px] text-left text-xs">
            <thead className="bg-[#08130E] font-mono text-[9px] uppercase tracking-wider text-text-subtle">
              <tr>
                <th className="w-12 px-5 py-3"><span className="sr-only">Include</span></th>
                <th className="px-3 py-3">Asset</th>
                <th className="px-3 py-3 text-right">Exposure</th>
                <th className="px-3 py-3 text-right">Lens priority</th>
                <th className="px-3 py-3">Milestone likelihood & range</th>
                <th className="px-3 py-3 text-right">vs reference</th>
                <th className="px-5 py-3">Suggested next move</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-borderSubtle">
              {allProjectModels.map((item) => {
                const included = selectedIds.has(item.project.id);
                const reference = modelProject(item.project, REFERENCE_SCENARIOS[0], horizon, role);
                const delta = item.probability - reference.probability;
                return (
                  <tr key={item.project.id} className={`transition-colors ${included ? "hover:bg-surface/60" : "opacity-45 hover:opacity-70"}`}>
                    <td className="px-5 py-4">
                      <input
                        type="checkbox"
                        checked={included}
                        onChange={() => toggleProject(item.project.id)}
                        aria-label={`${included ? "Remove" : "Add"} ${item.project.name} ${included ? "from" : "to"} portfolio`}
                        className="h-4 w-4 rounded border-border bg-surface accent-emerald-500"
                      />
                    </td>
                    <td className="px-3 py-4">
                      <Link href={`/projects/${item.project.slug}`} className="font-semibold text-text-main hover:text-aurora">
                        {item.project.name}
                      </Link>
                      <div className="mt-1 text-[9px] font-mono text-text-subtle">
                        {item.project.province} · {item.project.current_stage} · {item.project.confidence}
                      </div>
                    </td>
                    <td className="px-3 py-4 text-right font-mono font-semibold text-text-main">
                      {formatCad(item.project.capex_cad)}
                    </td>
                    <td className="px-3 py-4 text-right">
                      <span className="font-mono font-bold text-aurora">{item.priority.toFixed(0)}</span>
                      <span className="text-[9px] text-text-subtle">/100</span>
                    </td>
                    <td className="w-48 px-3 py-4">
                      <div className="mb-2 flex items-center justify-between font-mono text-[9px]">
                        <span className="text-text-subtle">{formatPercent(item.low)}</span>
                        <span className="font-bold text-text-main">{formatPercent(item.probability)}</span>
                        <span className="text-text-subtle">{formatPercent(item.high)}</span>
                      </div>
                      <RangeBar low={item.low} base={item.probability} high={item.high} />
                    </td>
                    <td className={`px-3 py-4 text-right font-mono ${delta >= 0 ? "text-aurora" : "text-gold"}`}>
                      {formatSignedPp(delta)}
                    </td>
                    <td className="px-5 py-4 text-[10px] font-medium text-text-muted">{getAction(item.project, role)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </section>

      <section className="grid gap-4 rounded-2xl border border-borderSubtle bg-surface/60 p-5 md:grid-cols-[1fr_auto] md:items-center">
        <div className="flex items-start gap-3">
          <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-aurora" />
          <div>
            <h2 className="text-xs font-bold text-text-main">Model governance & appropriate use</h2>
            <p className="mt-1 max-w-4xl text-[10px] leading-relaxed text-text-subtle">
              Outputs are deterministic planning estimates derived from lifecycle stage, four existing CEGS scores,
              evidence class, reported CAPEX, horizon, and visible scenario assumptions. Ranges are heuristic sensitivity
              bands, not statistically calibrated confidence intervals. Results are not investment, procurement, legal,
              credit, or cabinet advice; verify source evidence and approvals before acting.
            </p>
          </div>
        </div>
        <Link
          href="/methodology"
          className="inline-flex items-center justify-center gap-1.5 rounded-xl border border-border bg-card px-3 py-2 text-[10px] font-semibold text-aurora hover:border-primary/50"
        >
          Review methodology <ArrowUpRight className="h-3.5 w-3.5" />
        </Link>
      </section>
    </div>
  );
}
