"use client";

import { useState } from "react";
import Link from "next/link";
import { 
  MapPin, 
  Layers, 
  ChevronRight, 
  Pickaxe, 
  Zap, 
  Anchor, 
  Shield, 
  Cpu, 
  ArrowUpRight, 
  Building2, 
  ShieldCheck, 
  Activity,
  Compass,
  Radio
} from "lucide-react";
import { FALLBACK_PROJECTS } from "@/lib/data";

export default function MapPage() {
  const [selectedSector, setSelectedSector] = useState("ALL");
  const [selectedStage, setSelectedStage] = useState("ALL");
  const [activeProject, setActiveProject] = useState(FALLBACK_PROJECTS[0]);
  const [hoveredProject, setHoveredProject] = useState<typeof FALLBACK_PROJECTS[0] | null>(null);

  const sectors = [
    "ALL",
    "Nuclear & Clean Power",
    "Critical Minerals",
    "Clean Energy & Grid",
    "AI Compute & Data Centres",
    "Transportation & Ports",
    "Defence & Arctic",
  ];

  const stages = ["ALL", "CONSTRUCTION", "PERMITTING", "FEASIBILITY", "OPERATING"];

  const filtered = FALLBACK_PROJECTS.filter((p) => {
    if (selectedSector !== "ALL" && p.sector !== selectedSector) return false;
    if (selectedStage !== "ALL" && p.current_stage !== selectedStage) return false;
    return true;
  });

  const totalFilteredCapex = filtered.reduce((acc, p) => acc + p.capex_cad, 0);

  // Strategic Corridor Transmission Vectors (Connecting Hubs)
  const corridors = [
    // Clean Hydro from Chisasibi (QC) to Montreal
    { from: "proj-chisasibi-ai-compute", to: "proj-hyperscale-qc", name: "Baie-James Clean Hydro Corridor", color: "#00F5A0" },
    // Churchill Arctic Rail to Gillam/Prairies
    { from: "proj-churchill-arctic-gateway", to: "proj-kivalliq-link", name: "Hudson Bay Strategic Northern Vector", color: "#F59E0B" },
    // Bruce to Darlington Ontario Clean Nuclear Belt
    { from: "proj-bruce-nuclear", to: "proj-darlington-smr", name: "Ontario 10GW Clean Nuclear Baseload Spine", color: "#00F5A0" },
    // Timmins Nickel to Southern Ontario EV Battery Supply Chain
    { from: "proj-crawford-nickel", to: "proj-oneida-battery", name: "Critical Mineral - EV Battery Highway", color: "#F59E0B" },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Top Section Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-border/80 pb-6">
        <div>
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-card border border-primary/40 text-[11px] font-mono text-aurora mb-3 shadow-sm">
            <Radio className="h-3.5 w-3.5 text-aurora animate-pulse" />
            GEOSPATIAL SOVEREIGN CAPITAL RADAR — CEGS 0.1
          </div>
          <h1 className="text-2xl sm:text-4xl font-black tracking-tight text-text-main">
            National Economic <span className="text-aurora">Infrastructure Map</span>
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-1.5 max-w-3xl leading-relaxed">
            Visualizing <span className="text-aurora font-semibold">${(totalFilteredCapex / 1e9).toFixed(2)}B CAD</span> across {filtered.length} verified nation-building assets: 
            clean baseload nuclear reactors, high-voltage Arctic transmission, critical mineral extraction camps, and sovereign AI compute clusters.
          </p>
        </div>

        {/* Filter Controls */}
        <div className="flex flex-wrap items-center gap-2">
          <select
            value={selectedSector}
            onChange={(e) => setSelectedSector(e.target.value)}
            className="bg-card border border-border rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-primary font-mono transition-colors"
          >
            {sectors.map((s) => (
              <option key={s} value={s}>
                {s === "ALL" ? "All Strategic Sectors" : s}
              </option>
            ))}
          </select>

          <select
            value={selectedStage}
            onChange={(e) => setSelectedStage(e.target.value)}
            className="bg-card border border-border rounded-xl px-3 py-2 text-text-main text-xs focus:outline-none focus:border-primary font-mono transition-colors"
          >
            {stages.map((st) => (
              <option key={st} value={st}>
                {st === "ALL" ? "All Lifecycle Stages" : st}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Main Map Visualizer & Sidebar */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 min-h-[620px]">
        {/* Interactive SVG Canadian Map Canvas */}
        <div className="lg:col-span-8 bg-[#040806] rounded-2xl border border-border/80 relative overflow-hidden flex items-center justify-center p-4 shadow-2xl">
          {/* Subtle Ambient Radial Grid Glow */}
          <div className="absolute inset-0 bg-[radial-gradient(#00F5A0_1px,transparent_1px)] [background-size:24px_24px] opacity-10 pointer-events-none"></div>

          {/* Map Status Floating Legend */}
          <div className="absolute top-4 left-4 z-10 bg-card/90 backdrop-blur-md border border-border/80 p-3 rounded-xl text-[11px] font-mono space-y-2 shadow-xl">
            <div className="text-text-subtle uppercase text-[9px] font-bold tracking-wider flex items-center gap-1.5">
              <Compass className="h-3 w-3 text-aurora" />
              Sovereign Grid Layers
            </div>
            <div className="flex flex-wrap items-center gap-3 text-text-main">
              <div className="flex items-center gap-1.5">
                <span className="h-2 w-2 rounded-full bg-aurora shadow-[0_0_8px_#00F5A0]"></span>
                <span className="text-[10px]">Construction</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="h-2 w-2 rounded-full bg-gold shadow-[0_0_8px_#F59E0B]"></span>
                <span className="text-[10px]">Permitting / EA</span>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="h-2 w-2 rounded-full bg-aurora-mint"></span>
                <span className="text-[10px]">Feasibility</span>
              </div>
            </div>
          </div>

          {/* High-Tech Canadian Map Topology */}
          <svg viewBox="0 0 1000 650" className="w-full h-full max-h-[560px] select-none">
            <defs>
              <linearGradient id="auroraVector" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#00F5A0" stopOpacity="0.8" />
                <stop offset="100%" stopColor="#F59E0B" stopOpacity="0.4" />
              </linearGradient>
              <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
                <feGaussianBlur stdDeviation="3" result="blur" />
                <feComposite in="SourceGraphic" in2="blur" operator="over" />
              </filter>
            </defs>

            {/* Canadian Sovereign Territory Landmass (Stylized Topology) */}
            <path
              d="M 120 180 L 180 150 L 260 120 L 380 90 L 520 80 L 680 70 L 780 110 L 850 160 L 920 220 L 880 320 L 820 400 L 750 480 L 680 500 L 550 520 L 420 540 L 300 550 L 180 520 L 100 420 L 70 300 Z"
              fill="#0A140F"
              stroke="#0D2E1E"
              strokeWidth="2"
            />

            {/* Maritime & Great Lakes Water Insets (Stylized) */}
            <path
              d="M 600 470 Q 640 450 670 480 Q 650 510 610 500 Z"
              fill="#040806"
              stroke="#0D2E1E"
              strokeWidth="1"
            />
            <path
              d="M 520 180 Q 560 160 580 210 Q 530 250 500 210 Z"
              fill="#040806"
              stroke="#0D2E1E"
              strokeWidth="1"
            />

            {/* Provincial & Territorial Regional Boundaries */}
            {/* Arctic / Nunavut & Northwest Territories */}
            <path d="M 380 90 L 450 200 L 600 220 L 700 150" fill="none" stroke="#123B27" strokeWidth="1.2" strokeDasharray="3 3" />
            {/* Quebec & Ontario */}
            <path d="M 580 320 L 640 470" fill="none" stroke="#123B27" strokeWidth="1.2" strokeDasharray="3 3" />
            {/* Prairies & British Columbia */}
            <path d="M 280 340 L 320 540" fill="none" stroke="#123B27" strokeWidth="1.2" strokeDasharray="3 3" />

            {/* Regional Labels */}
            <text x="520" y="140" fill="#2D5A43" fontSize="11" fontFamily="monospace" letterSpacing="2">ARCTIC WATERS / NUNAVUT</text>
            <text x="350" y="440" fill="#2D5A43" fontSize="10" fontFamily="monospace" letterSpacing="1.5">WESTERN PRAIRIES</text>
            <text x="610" y="420" fill="#2D5A43" fontSize="10" fontFamily="monospace" letterSpacing="1.5">ONTARIO INDUSTRIAL BELT</text>
            <text x="710" y="360" fill="#2D5A43" fontSize="10" fontFamily="monospace" letterSpacing="1.5">QUEBEC CLEAN HYDRO</text>

            {/* Strategic Infrastructure Corridors (Transmission & Rail Vectors) */}
            {corridors.map((c, idx) => {
              const fromP = FALLBACK_PROJECTS.find(p => p.id === c.from);
              const toP = FALLBACK_PROJECTS.find(p => p.id === c.to);
              if (!fromP || !toP) return null;

              const x1 = 150 + ((fromP.longitude + 130) / 65) * 700;
              const y1 = 550 - ((fromP.latitude - 42) / 33) * 450;
              const x2 = 150 + ((toP.longitude + 130) / 65) * 700;
              const y2 = 550 - ((toP.latitude - 42) / 33) * 450;

              return (
                <g key={idx}>
                  <line
                    x1={x1}
                    y1={y1}
                    x2={x2}
                    y2={y2}
                    stroke={c.color}
                    strokeWidth="1.5"
                    strokeDasharray="4 3"
                    opacity="0.6"
                  />
                </g>
              );
            })}

            {/* Project Nodes on Canvas */}
            {filtered.map((p) => {
              // Map coordinates: Lat (42 to 75), Long (-130 to -65) -> Canvas (1000 x 650)
              const x = 150 + ((p.longitude + 130) / 65) * 700;
              const y = 550 - ((p.latitude - 42) / 33) * 450;
              const isSelected = activeProject.id === p.id;
              const isHovered = hoveredProject?.id === p.id;

              let markerColor = "#00F5A0"; // Construction
              if (p.current_stage === "PERMITTING" || p.current_stage === "ENVIRONMENTAL_REVIEW") markerColor = "#F59E0B";
              if (p.current_stage === "FEASIBILITY" || p.current_stage === "ANNOUNCED") markerColor = "#4ECCA3";

              return (
                <g
                  key={p.id}
                  className="cursor-pointer transition-all duration-300"
                  onClick={() => setActiveProject(p)}
                  onMouseEnter={() => setHoveredProject(p)}
                  onMouseLeave={() => setHoveredProject(null)}
                >
                  {/* Outer Pulsing Glow */}
                  {isSelected && (
                    <circle
                      cx={x}
                      cy={y}
                      r="20"
                      fill="none"
                      stroke={markerColor}
                      strokeWidth="1.5"
                      className="animate-ping"
                      opacity="0.4"
                    />
                  )}

                  {/* Active Border Ring */}
                  <circle
                    cx={x}
                    cy={y}
                    r={isSelected || isHovered ? "11" : "7"}
                    fill="#050B08"
                    stroke={markerColor}
                    strokeWidth={isSelected ? "2.5" : "1.5"}
                    filter="url(#glow)"
                  />

                  {/* Inner Solid Pin Core */}
                  <circle
                    cx={x}
                    cy={y}
                    r={isSelected || isHovered ? "6" : "4"}
                    fill={markerColor}
                  />

                  {/* Node Label Text */}
                  <text
                    x={x + 14}
                    y={y + 4}
                    fill={isSelected ? "#00F5A0" : isHovered ? "#FFFFFF" : "#94A3B8"}
                    fontSize="11"
                    fontFamily="monospace"
                    fontWeight={isSelected ? "bold" : "normal"}
                  >
                    {p.name.split(" ")[0]}
                  </text>
                </g>
              );
            })}
          </svg>
        </div>

        {/* Selected Project Dossier Card */}
        <div className="lg:col-span-4 glass-card rounded-2xl border border-border/80 p-6 flex flex-col justify-between shadow-2xl">
          <div className="space-y-5">
            {/* Header Badge & Province */}
            <div className="border-b border-borderSubtle pb-4">
              <div className="flex items-center justify-between text-[11px] font-mono">
                <span className="px-2.5 py-0.5 rounded-full bg-primary/10 border border-primary/30 text-aurora font-bold">
                  {activeProject.province} // {activeProject.current_stage}
                </span>
                <span className="text-text-subtle font-mono">{activeProject.id}</span>
              </div>
              <h2 className="text-lg font-black text-text-main mt-2 tracking-tight leading-snug">
                {activeProject.name}
              </h2>
              <div className="text-xs text-text-muted mt-1 flex items-center gap-1.5">
                <MapPin className="h-3.5 w-3.5 text-aurora" />
                <span>{activeProject.location_name}</span>
              </div>
            </div>

            {/* Summary */}
            <p className="text-xs text-text-muted leading-relaxed">
              {activeProject.summary}
            </p>

            {/* Key Metrics Grid */}
            <div className="grid grid-cols-2 gap-3 text-xs font-mono">
              <div className="p-3 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[10px] text-text-subtle uppercase">Reported CAPEX</div>
                <div className="text-base font-black text-aurora mt-0.5 font-tabular">
                  ${(activeProject.capex_cad / 1e9).toFixed(2)}B CAD
                </div>
              </div>
              <div className="p-3 rounded-xl bg-surface border border-borderSubtle">
                <div className="text-[10px] text-text-subtle uppercase">Subsector</div>
                <div className="text-xs font-bold text-text-main mt-0.5 truncate">
                  {activeProject.subsector}
                </div>
              </div>
            </div>

            {/* CEGS Quantitative Scores */}
            <div className="space-y-2 pt-1">
              <div className="text-[10px] font-mono text-text-subtle uppercase flex justify-between">
                <span>Buildability & Readiness</span>
                <span className="text-aurora font-bold">{activeProject.scores?.buildability.toFixed(0)}/100</span>
              </div>
              <div className="w-full bg-surface h-2 rounded-full overflow-hidden border border-borderSubtle">
                <div 
                  className="bg-primary h-full rounded-full transition-all duration-500" 
                  style={{ width: `${activeProject.scores?.buildability || 50}%` }}
                ></div>
              </div>

              <div className="text-[10px] font-mono text-text-subtle uppercase flex justify-between pt-1">
                <span>National Strategicity</span>
                <span className="text-gold font-bold">{activeProject.scores?.strategicity.toFixed(0)}/100</span>
              </div>
              <div className="w-full bg-surface h-2 rounded-full overflow-hidden border border-borderSubtle">
                <div 
                  className="bg-gold h-full rounded-full transition-all duration-500" 
                  style={{ width: `${activeProject.scores?.strategicity || 75}%` }}
                ></div>
              </div>
            </div>

            {/* Geographic Coordinates & Evidence */}
            <div className="p-3 rounded-xl bg-surface/50 border border-borderSubtle text-[11px] font-mono space-y-1">
              <div className="text-text-subtle uppercase text-[9px]">Geospatial Footprint</div>
              <div className="text-text-muted">
                Lat: {activeProject.latitude.toFixed(4)}° N, Long: {activeProject.longitude.toFixed(4)}° W
              </div>
              <div className="text-[10px] text-aurora-mint flex items-center gap-1 pt-1">
                <ShieldCheck className="h-3 w-3" /> Anchored in CEGS 0.1 Cryptographic Graph
              </div>
            </div>
          </div>

          {/* Action Link to Full Project Dossier */}
          <div className="pt-4 border-t border-borderSubtle mt-4">
            <Link
              href={`/projects/${activeProject.slug}`}
              className="w-full py-2.5 rounded-xl bg-primary text-[#050B08] text-xs font-bold hover:bg-aurora-mint transition-all shadow-md shadow-emerald-950/40 flex items-center justify-center gap-1.5"
            >
              Open Full Cryptographic Dossier <ChevronRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
