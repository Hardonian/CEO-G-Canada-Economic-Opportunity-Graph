"use client";

import { useEffect, useState } from "react";

export default function GlobalError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  const [language, setLanguage] = useState<"en" | "fr">("en");

  useEffect(() => {
    setLanguage(document.documentElement.dataset.language === "fr" ? "fr" : "en");
    console.error("[ui] Root render failed", { digest: error.digest, message: error.message });
  }, [error]);

  const french = language === "fr";
  return (
    <html lang={french ? "fr-CA" : "en-CA"} className="dark">
      <body className="min-h-screen bg-[#050b08] p-6 font-sans text-[#f7faf8]">
        <div className="mx-auto flex max-w-2xl justify-end gap-1">
          <button onClick={() => setLanguage("en")} aria-pressed={!french} className="rounded border border-[#314b40] px-3 py-2">EN</button>
          <button onClick={() => setLanguage("fr")} aria-pressed={french} className="rounded border border-[#314b40] px-3 py-2">FR</button>
        </div>
        <main className="mx-auto mt-16 max-w-2xl rounded-2xl border border-[#ff5d73] bg-[#12231b] p-8">
          <h1 className="text-2xl font-black">{french ? "L’application n’a pas pu démarrer" : "The application could not start"}</h1>
          <p className="mt-3 text-sm text-[#d0ddd6]">
            {french ? "Aucune donnée de remplacement n’a été inventée. Réessayez ou communiquez l’identifiant de diagnostic." : "No replacement data was invented. Try again or report the diagnostic identifier."}
          </p>
          {error.digest && <p className="mt-3 font-mono text-xs text-[#a8c4b8]">{french ? "Identifiant" : "Digest"}: {error.digest}</p>}
          <button onClick={reset} className="mt-6 rounded-lg bg-[#00d488] px-4 py-2 text-sm font-black text-[#050b08]">
            {french ? "Réessayer" : "Try again"}
          </button>
        </main>
      </body>
    </html>
  );
}
