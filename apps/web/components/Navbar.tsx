"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import {
  Activity,
  Cpu,
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
} from "lucide-react";

export default function Navbar() {
  const pathname = usePathname();
  const [lang, setLang] = useState<"en" | "fr">("en");

  const navItems = [
    { href: "/briefing", label: lang === "en" ? "Executive Brief" : "Note exécutive", icon: Landmark, badge: lang === "en" ? "ANALYSIS" : "ANALYSE" },
    { href: "/", label: lang === "en" ? "Capital Radar" : "Radar du capital", icon: Radar },
    { href: "/planning", label: lang === "en" ? "Planning" : "Planification", icon: Target },
    { href: "/projects", label: lang === "en" ? "Projects" : "Projets", icon: FolderGit2 },
    { href: "/map", label: lang === "en" ? "Geospatial Map" : "Carte géospatiale", icon: MapPin },
    { href: "/capital", label: lang === "en" ? "Capital Stack" : "Structure du capital", icon: Layers },
    { href: "/procurement", label: lang === "en" ? "Procurement" : "Approvisionnement", icon: Activity },
    { href: "/ai-sovereignty", label: lang === "en" ? "AI Sovereignty" : "Souveraineté IA", icon: Cpu },
    { href: "/cegs", label: lang === "en" ? "CEGS Standard" : "Norme CEGS", icon: FileCode2, badge: "0.1" },
  ];

  const nextLanguage = lang === "en" ? "French" : "English";

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
            <button
              type="button"
              onClick={() => setLang(lang === "en" ? "fr" : "en")}
              className="inline-flex min-h-7 items-center gap-1.5 rounded px-1.5 text-text-muted transition-colors hover:bg-surface hover:text-text-main"
              aria-label={`Switch navigation to ${nextLanguage}`}
              title={`Switch navigation to ${nextLanguage}`}
            >
              <Languages aria-hidden="true" className="h-3.5 w-3.5" />
              <span className={lang === "en" ? "font-black text-text-main" : undefined}>EN</span>
              <span aria-hidden="true" className="text-text-subtle">/</span>
              <span className={lang === "fr" ? "font-black text-text-main" : undefined}>FR</span>
            </button>
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
              Independent economic intelligence
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
        </div>
      </div>

      <div data-print-hide="true" className="border-t border-borderSubtle bg-[#07110d]">
        <nav
          lang={lang === "fr" ? "fr-CA" : "en-CA"}
          aria-label={lang === "en" ? "Primary navigation" : "Navigation principale"}
          className="mx-auto max-w-7xl overflow-x-auto px-3 sm:px-5 lg:px-7"
        >
          <ul className="flex min-w-max list-none items-center gap-1 py-1.5" role="list">
            {navItems.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));

              return (
                <li key={item.href}>
                  <Link
                    href={item.href}
                    aria-current={isActive ? "page" : undefined}
                    className={`inline-flex min-h-9 items-center gap-1.5 whitespace-nowrap rounded-md border px-2.5 py-1.5 text-[11px] font-semibold transition-colors sm:px-3 sm:text-xs ${
                      isActive
                        ? "border-primary/70 bg-card text-aurora shadow-[inset_0_-2px_0_rgba(0,245,160,0.8)]"
                        : "border-transparent text-text-muted hover:border-borderSubtle hover:bg-surface hover:text-text-main"
                    }`}
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
