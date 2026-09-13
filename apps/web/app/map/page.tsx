"use client";

import { useState } from "react";
import Link from "next/link";
import { MapPin, Filter, Layers, ChevronRight, Pickaxe, Zap, Anchor, Shield, Cpu } from "lucide-react";
import { FALLBACK_PROJECTS } from "@/lib/data";

export default function MapPage() {
  const [selectedSector, setSelectedSector] = useState("ALL");
  const [activeProject, setActiveProject] = useState(FALLBACK_PROJECTS[0]);

  const filtered = FALLBACK_PROJECTS.filter((p) => {
    if (selectedSector !== "ALL" && p.sector !== selectedSector) return false;
    return true;
  });

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-b border-border pb-6">
        <div>
          <div className="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-card border border-border text-[11px] font-mono text-accent-cyan mb-2">
            <MapPin className="h-3 w-3" />
            CANADIAN GEOSPATIAL INFRASTRUCTURE GIS
          </div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-main">
            National Economic Map
          </h1>
          <p className="text-xs sm:text-sm text-text-muted mt-1">
            Visualizing major capital investments, clean grid corridors, critical mineral camps, and Arctic strategic facilities across Canadian provinces and territories.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={selectedSector}
            onChange={(e) => setSelectedSector(e.target.value)}
            className="bg-card border border-border rounded px-3 py-1.5 text-text-main text-xs focus:outline-none focus:border-accent-cyan font-mono"
          >
            <option value="ALL">All Strategic Sectors</option>
            <option value="Nuclear & Clean Power">Nuclear & Clean Power</option>
            <option value="Critical Minerals">Critical Minerals</option>
            <option value="Clean Energy & Grid">Clean Energy & Grid</option>
            <option value="Transportation & Ports">Transportation & Ports</option>
            <option value="Defence & Arctic">Defence & Arctic</option>
          </select>
        </div>
      </div>

      {/* Main Map Visualizer & Sidebar */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-[600px]">
        {/* Interactive SVG Canadian Map Canvas */}
        <div className="lg:col-span-2 bg-[#050811] rounded border border-border relative overflow-hidden flex items-center justify-center p-4">
          <div className="absolute top-3 left-3 z-10 bg-card/90 backdrop-blur-sm border border-border p-2 rounded text-[11px] font-mono space-y-1">
            <div className="text-text-subtle uppercase text-[9px]">Map Layers</div>
            <div className="flex items-center gap-2 text-text-main">
              <span className="h-2 w-2 rounded-full bg-accent-cyan"></span> Active Facilities
              <span className="h-2 w-2 rounded-full bg-accent-green"></span> In Construction
              <span className="h-2 w-2 rounded-full bg-accent-gold"></span> Permitting / EA
            </div>
          </div>

          {/* Stylized High-Tech Canadian Map Topology */}
          <svg viewBox="0 0 1000 650" className="w-full h-full max-h-[550px] select-none">
            <defs>
              <radialGradient id="mapGlow" cx="50%" cy="50%" r="50%">
                <stop offset="0%" stopColor="#00F2FE" stopOpacity="0.08" />
                <stop offset="100%" stopColor="#00F2FE" stopOpacity="0" />
              </radialGradient>
            </defs>

            {/* Background Grid Lines */}
            <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
              <path d="M 40 0 L 0 0 0 40" fill="none" stroke="#10192E" strokeWidth="0.8" />
            </pattern>
            <rect width="1000" height="650" fill="url(#grid)" />
            <circle cx="500" cy="350" r="400" fill="url(#mapGlow)" />

            {/* Stylized Canada Landmass Polygon Outline */}
            <path
              d="M 120 180 L 180 150 L 260 120 L 380 90 L 520 80 L 680 70 L 780 110 L 850 160 L 920 220 L 880 320 L 820 400 L 750 480 L 680 500 L 550 520 L 420 540 L 300 550 L 180 520 L 100 420 L 70 300 Z"
              fill="#0A1224"
              stroke="#1C2B4B"
              strokeWidth="2"
              strokeDasharray="4 2"
            />

            {/* Provincial Regional Boundaries (Stylized) */}
            {/* Arctic / Nunavut / NWT */}
            <path d="M 380 90 L 450 200 L 600 220 L 700 150" fill="none" stroke="#16223B" strokeWidth="1" />
            {/* Quebec / Ontario */}
            <path d="M 600 320 L 650 480" fill="none" stroke="#16223B" strokeWidth="1" />
            {/* Prairies / BC */}
            <path d="M 280 340 L 320 540" fill="none" stroke="#16223B" strokeWidth="1" />

            {/* Project Nodes on Canvas */}
            {filtered.map((p) => {
              // Map real coords (Lat: 42-75, Long: -130 to -65) to Canvas (x: 100-900, y: 100-550)
              const x = 150 + ((p.longitude + 130) / 65) * 700;
              const y = 550 - ((p.latitude - 42) / 33) * 450;
              const isSelected = activeProject.id === p.id;

              let markerColor = "#00F2FE";
              if (p.current_stage === "CONSTRUCTION") markerColor = "#10B981";
              if (p.current_stage === "PERMITTING" || p.current_stage === "FEASIBILITY") markerColor = "#F59E0B";

              return (
                <g
                  key={p.id}
                  className="cursor-pointer transition-transform duration-200 hover:scale-125"
                  onClick={() => setActiveProject(p)}
                >
                  {/* Ping effect for selected */}
                  {isSelected && (
                    <circle cx={x} cy={y} r="18" fill="none" stroke={markerColor} strokeWidth="1.5" className="animate-ping" opacity="0.6" />
                  )}
                  <circle cx={x} cy={y} r={isSelected ? "8" : "6"} fill={markerColor} stroke="#050811" strokeWidth="2" />
                  <text
                    x={x + 10}
                    y={y + 4}
                    fill={isSelected ? "#FFFFFF" : "#94A3B8"}
                    fontSize="10"
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
        <div className="bg-card rounded border border-border p-5 flex flex-col justify-between">
          <div className="space-y-4">
            <div className="border-b border-borderSubtle pb-3">
              <div className="flex items-center gap-1.5 text-[10px] font-mono text-text-subtle uppercase">
                <span>Selected Map Asset</span>
                <span>•</span>
                <span className="text-accent-cyan">{activeProject.province}</span>
              </div>
              <h2 className="text-base font-bold text-text-main mt-1">
                {activeProject.name}
              </h2>
              <div className="text-xs text-text-muted mt-0.5">
                {activeProject.location_name}
              </div>
            </div>

            <p className="text-xs text-text-subtle leading-relaxed">
              {activeProject.summary}
            </p>

            <div className="grid grid-cols-2 gap-2 text-xs font-mono pt-2">
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Reported CAPEX</div>
                <div className="text-sm font-bold text-text-main mt-0.5">
                  ${(activeProject.capex_cad / 1e9).toFixed(2)}B CAD
                </div>
              </div>
              <div className="p-2.5 rounded bg-surface border border-borderSubtle">
                <div className="text-[9px] text-text-subtle uppercase">Lifecycle Stage</div>
                <div className="text-sm font-bold text-primary mt-0.5">
                  {activeProject.current_stage}
                </div>
              </div>
            </div>

            <div className="space-y-1.5 pt-2">
              <div className="text-[10px] font-mono text-text-subtle uppercase">Coordinates</div>
              <div className="text-xs font-mono text-text-muted bg-surface p-2 rounded border border-borderSubtle">
                Lat: {activeProject.latitude.toFixed(4)}° N, Long: {activeProject.longitude.toFixed(4)}° W
              </div>
            </div>
          </div>

          <div className="pt-4 border-t border-borderSubtle">
            <Link
              href={`/projects/${activeProject.slug}`}
              className="w-full py-2 rounded bg-primary text-white text-xs font-semibold hover:bg-primary-hover transition-colors flex items-center justify-center gap-1.5"
            >
              Open Full Investor Profile <ChevronRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
