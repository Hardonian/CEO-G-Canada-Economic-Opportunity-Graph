"use client";

import { AlertTriangle, RotateCcw } from "lucide-react";
import { useEffect } from "react";
import { useLanguage } from "@/components/LanguageProvider";

export default function ApplicationError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  const { language } = useLanguage();
  const french = language === "fr";

  useEffect(() => {
    console.error("[ui] Route render failed", { digest: error.digest, message: error.message });
  }, [error]);

  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 lg:px-8">
      <section role="alert" className="rounded-2xl border border-crimson/50 bg-crimson/10 p-7">
        <AlertTriangle aria-hidden="true" className="h-8 w-8 text-crimson-light" />
        <h1 className="mt-4 text-2xl font-black text-text-main">
          {french ? "Cette vue n’a pas pu être chargée" : "This view could not be loaded"}
        </h1>
        <p className="mt-2 text-sm leading-relaxed text-text-muted">
          {french
            ? "Les données vérifiées n’ont pas été remplacées. Réessayez; si le problème persiste, utilisez l’identifiant ci-dessous pour le suivi."
            : "No vetted data was substituted. Try again; if the issue persists, use the identifier below for support."}
        </p>
        {error.digest && <p className="mt-3 font-mono text-xs text-text-subtle">{french ? "Identifiant" : "Digest"}: {error.digest}</p>}
        <button type="button" onClick={reset} className="mt-6 inline-flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
          <RotateCcw aria-hidden="true" className="h-4 w-4" />
          {french ? "Réessayer" : "Try again"}
        </button>
      </section>
    </div>
  );
}
