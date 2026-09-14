"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useRef, useEffect } from "react";
import {
  Activity,
  Cpu,
  Database,
  FileCode2,
  FolderGit2,
  Landmark,
  Languages,
  Layers,
  MapPin,
  Radar,
  Search,
  ShieldCheck,
  Target,
  Menu,
  X,
} from "lucide-react";
import { useLanguage } from "@/components/LanguageProvider";

export default function Navbar() {
  const pathname = usePathname();
  const { language: lang, setLanguage } = useLanguage();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const mobileMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && mobileMenuOpen) {
        setMobileMenuOpen(false);
      }
    };
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [mobileMenuOpen]);

  const navItems = [
    { href: "/briefing", label: lang === "en" ? "Executive Brief" : "Note exécutive", icon: Landmark, badge: lang === "en" ? "ANALYSIS" : "ANALYSE" },
    { href: "/", label: lang === "en" ? "Capital Radar" : "Radar du capital", icon: Radar },
    { href: "/planning", label: lang === "en" ? "Planning" : "Planification", icon: Target },
    { href: "/projects", label: lang === "en" ? "Projects" : "Projets", icon: FolderGit2 },
    { href: "/map", label: lang === "en" ? "Geospatial Map" : "Carte géospatiale", icon: MapPin },
    { href: "/capital", label: lang === "en" ? "Capital Stack" : "Structure du capital", icon: Layers },
    { href: "/procurement", label: lang === "en" ? "Procurement" : "Approvisionnement", icon: Activity },
    { href: "/sources", label: lang === "en" ? "Public Data" : "Données publiques", icon: Database },
    { href: "/ai-sovereignty", label: lang === "en" ? "AI Sovereignty" : "Souveraineté IA", icon: Cpu },
    { href: "/cegs", label: lang === "en" ? "CEGS Standard" : "Norme CEGS", icon: FileCode2, badge: "0.1" },
  ];

  return (
    <header
      data-site-header="true"
      className="sticky top-0 z-50 border-b border-border bg-background/95 shadow-[0_10px_32px_-24px_rgba(0,0,0,0.95)] backdrop-blur-xl"
    >
      <div data-print-hide="true" className="border-b border-borderSubtle bg-[#08130e]">
        <div className="mx-auto flex min-h-9 max-w-7xl items-center justify-between gap-3 px-4 text-[10px] font-mono uppercase tracking-[0.12em] text-text-muted sm:px-6 lg:px-8">
          <div className="flex min-w-0 items-center gap-2.5">
            <span className="inline-flex shrink-0 items-center gap-1.5 rounded border border-crimson/60 bg-crimson/10 px-2 py-1 font-bold text-crimson-light">
              <span aria-hidden="true" className="h-1.5 w-1.5 rounded-full bg-crimson" />
              Independent
            </span>
            <span className="hidden truncate sm:inline">Public-interest Canadian capital intelligence</span>
          </div>

          <div className="flex shrink-0 items-center gap-2 sm:gap-3">
            <span className="hidden text-text-subtle xl:inline">Open-source research · Not a government service</span>
            <span className="inline-flex items-center gap-1 rounded border border-primary/40 bg-primary/10 px-2 py-1 font-semibold text-aurora">
              <ShieldCheck aria-hidden="true" className="h-3 w-3" />
              CEGS 0.1
            </span>
            <div className="inline-flex min-h-8 items-center rounded-md border border-borderSubtle bg-background p-0.5" role="group" aria-label={lang === "en" ? "Interface language" : "Langue de l’interface"}>
              <Languages aria-hidden="true" className="h-3.5 w-3.5" />
              <button
                type="button"
                onClick={() => setLanguage("en")}
                aria-pressed={lang === "en"}
                aria-label="English"
                title="English"
                className={`ml-1 min-h-7 rounded px-2 font-bold transition-colors ${lang === "en" ? "bg-primary text-[#050b08]" : "text-text-muted hover:bg-surface hover:text-text-main"}`}
              >
                EN
              </button>
              <button
                type="button"
                onClick={() => setLanguage("fr")}
                aria-pressed={lang === "fr"}
                aria-label="Français"
                title="Français"
                className={`min-h-7 rounded px-2 font-bold transition-colors ${lang === "fr" ? "bg-primary text-[#050b08]" : "text-text-muted hover:bg-surface hover:text-text-main"}`}
              >
                FR
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
        <Link
          href="/"
          className="group flex min-w-0 items-center gap-3 rounded-lg"
          aria-label="CanadaOpportunityGraph home"
        >
          <span className="relative flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border bg-card font-mono text-sm font-black tracking-tight text-text-main shadow-inner transition-colors group-hover:border-primary">
            <span aria-hidden="true" className="absolute inset-x-0 top-0 h-1 bg-crimson" />
            CA
          </span>
          <span className="min-w-0">
            <span className="flex items-center gap-2">
              <span className="truncate text-sm font-black tracking-tight text-text-main transition-colors group-hover:text-aurora sm:text-base">
                CanadaOpportunityGraph
              </span>
              <span className="hidden rounded border border-gold/50 bg-gold/10 px-1.5 py-0.5 font-mono text-[9px] font-bold text-gold sm:inline">
                CEGS v0.1
              </span>
            </span>
            <span className="block truncate font-mono text-[9px] uppercase tracking-[0.14em] text-text-subtle sm:text-[10px]">
              {lang === "en" ? "Independent economic intelligence" : "Veille économique indépendante"}
            </span>
          </span>
        </Link>

        <div data-print-hide="true" className="flex shrink-0 items-center gap-2">
          <Link
            href="/projects"
            className="inline-flex min-h-10 items-center gap-2 rounded-lg border border-border bg-surface px-3 text-xs font-semibold text-text-muted shadow-inner transition-colors hover:border-primary hover:bg-card hover:text-text-main"
            aria-label={lang === "en" ? "Search projects" : "Rechercher des projets"}
          >
            <Search aria-hidden="true" className="h-4 w-4 text-aurora" />
            <span className="hidden sm:inline">{lang === "en" ? "Search projects" : "Rechercher"}</span>
            <kbd className="hidden rounded border border-borderSubtle bg-background px-1.5 py-0.5 font-mono text-[10px] text-text-muted lg:inline">
              /
            </kbd>
          </Link>

          <button
            type="button"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            aria-expanded={mobileMenuOpen}
            aria-controls="mobile-navigation"
            aria-label={mobileMenuOpen ? (lang === "en" ? "Close main navigation" : "Fermer la navigation principale") : (lang === "en" ? "Open main navigation" : "Ouvrir la navigation principale")}
            className="lg:hidden inline-flex min-h-10 items-center justify-center rounded-lg border border-border bg-surface px-3 text-text-main transition-colors hover:border-primary hover:bg-card"
          >
            {mobileMenuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>
      </div>

      <div data-print-hide="true" className={`border-t border-borderSubtle bg-[#07110d] lg:block ${mobileMenuOpen ? "block" : "lg:block hidden"}`} id="mobile-navigation">
        <nav
          lang={lang === "fr" ? "fr-CA" : "en-CA"}
          aria-label={lang === "en" ? "Primary navigation" : "Navigation principale"}
          className="mx-auto max-w-7xl overflow-x-auto px-3 sm:px-5 lg:px-7"
        >
          <ul className="flex min-w-max list-none items-center gap-1 py-1.5 lg:py-2" role="list">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));

              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    aria-current={isActive ? "page" : undefined}
                    className={`inline-flex min-h-9 lg:min-h-10 items-center gap-1.5 whitespace-nowrap rounded-md border px-2.5 py-1.5 text-[11px] font-semibold transition-colors sm:px-3 sm:text-xs ${
                      isActive
                        ? "border-primary/70 bg-card text-aurora shadow-[inset_0_-2px_0_rgba(0,245,160,0.8)]"
                        : "border-transparent text-text-muted hover:border-borderSubtle hover:bg-surface hover:text-text-main"
                    }`}
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    <Icon aria-hidden="true" className={`h-3.5 w-3.5 ${isActive ? "text-aurora" : "text-text-subtle"}`} />
                    <span>{item.label}</span>
                    {item.badge && (
                      <span className="ml-0.5 rounded border border-gold/40 bg-gold/10 px-1.5 py-0.5 font-mono text-[8px] font-bold text-gold">
                        {item.badge}
                      </span>
                    )}
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </div>
    </header>
  );
}
