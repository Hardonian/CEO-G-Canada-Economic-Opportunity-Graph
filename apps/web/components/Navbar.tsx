"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { 
  Radar, 
  FolderGit2, 
  MapPin, 
  Layers, 
  Cpu, 
  ShieldCheck, 
  FileCode2, 
  BookOpen, 
  Search,
  Languages,
  Activity,
  Landmark,
  Target
} from "lucide-react";

export default function Navbar() {
  const pathname = usePathname();
  const [lang, setLang] = useState<"en" | "fr">("en");

  const navItems = [
    { href: "/briefing", label: lang === "en" ? "PM Briefing" : "Mémo PM", icon: Landmark, badge: "CABINET" },
    { href: "/", label: lang === "en" ? "Capital Radar" : "Radar du capital", icon: Radar },
    { href: "/planning", label: lang === "en" ? "Planning" : "Planification", icon: Target },
    { href: "/projects", label: lang === "en" ? "Projects" : "Projets", icon: FolderGit2 },
    { href: "/map", label: lang === "en" ? "Geospatial Map" : "Carte géospatiale", icon: MapPin },
    { href: "/capital", label: lang === "en" ? "Capital Stack" : "Plafond de capital", icon: Layers },
    { href: "/procurement", label: lang === "en" ? "Procurement" : "Approvisionnement", icon: Activity },
    { href: "/ai-sovereignty", label: lang === "en" ? "AI Sovereignty" : "Souveraineté IA", icon: Cpu },
    { href: "/cegs", label: "CEGS Standard", icon: FileCode2, badge: "0.1" },
  ];

  return (
    <header className="sticky top-0 z-50 border-b border-border/80 bg-[#050B08]/90 backdrop-blur-md">
      {/* Top Banner: Status & National Intelligence Ticker */}
      <div className="bg-[#08130E] px-4 py-1 text-xs border-b border-borderSubtle flex items-center justify-between text-text-muted">
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-1.5 text-aurora font-medium tracking-wide">
            <span className="h-2 w-2 rounded-full bg-aurora shadow-[0_0_8px_#00F5A0] animate-pulse"></span>
            NATIONAL CAPITAL COMMAND
          </span>
          <span className="hidden sm:inline text-border">|</span>
          <span className="hidden sm:inline text-text-muted">Tracking $16.32B CAD Across 10 Strategic Assets & Downstream Supply Chains</span>
        </div>
        <div className="flex items-center gap-3">
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-card text-aurora-mint border border-primary/30 text-[10px] font-mono shadow-sm">
            CEGS 0.1 SPEC COMPLIANT
          </span>
          <button 
            onClick={() => setLang(lang === "en" ? "fr" : "en")}
            className="flex items-center gap-1 text-text-muted hover:text-text-main hover:text-aurora transition-colors text-[11px] font-mono"
            title="Toggle Language / Basculer la langue"
          >
            <Languages className="h-3 w-3" />
            <span className="font-semibold">{lang.toUpperCase()}</span>
          </button>
        </div>
      </div>

      {/* Main Navigation Bar */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-14">
          {/* Brand Logo */}
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="h-9 w-9 rounded-lg bg-card border border-primary/40 flex items-center justify-center text-primary group-hover:border-aurora group-hover:shadow-[0_0_12px_rgba(0,245,160,0.3)] transition-all">
              <span className="font-bold text-base">🍁</span>
            </div>
            <div>
              <div className="font-bold text-sm tracking-tight text-text-main group-hover:text-aurora flex items-center gap-1.5 transition-colors">
                CanadaOpportunityGraph
                <span className="text-[10px] px-1.5 py-0.5 rounded bg-surface border border-border text-gold font-mono font-medium">CEGS v0.1</span>
              </div>
              <div className="text-[10px] text-text-muted tracking-wider uppercase font-mono">
                Economic Graph & Capital Radar
              </div>
            </div>
          </Link>

          {/* Nav Links */}
          <nav className="hidden md:flex items-center space-x-1 lg:space-x-1.5">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium transition-all ${
                    isActive
                      ? "bg-card text-aurora border border-primary/50 shadow-[0_0_12px_rgba(0,245,160,0.15)]"
                      : "text-text-muted hover:text-text-main hover:bg-surface/80"
                  }`}
                >
                  <Icon className={`h-3.5 w-3.5 ${isActive ? "text-aurora" : "text-text-subtle"}`} />
                  <span>{item.label}</span>
                  {item.badge && (
                    <span className="ml-1 text-[9px] font-mono px-1.5 py-0.5 rounded bg-gold/20 text-gold border border-gold/30">
                      {item.badge}
                    </span>
                  )}
                </Link>
              );
            })}
          </nav>

          {/* Right Action: Global Search & Terminal */}
          <div className="flex items-center gap-2">
            <Link
              href="/projects"
              className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-border bg-surface text-text-muted text-xs hover:border-aurora/50 hover:text-text-main transition-all shadow-inner"
            >
              <Search className="h-3.5 w-3.5 text-text-subtle" />
              <span className="hidden sm:inline">Search projects...</span>
              <kbd className="hidden lg:inline text-[10px] font-mono bg-card px-1.5 py-0.5 rounded border border-borderSubtle text-aurora">
                /
              </kbd>
            </Link>
          </div>
        </div>
      </div>
    </header>
  );
}
