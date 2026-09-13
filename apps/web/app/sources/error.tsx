"use client";

import Link from "next/link";
import { AlertTriangle, RefreshCw } from "lucide-react";

export default function SourcesError({ reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <div className="mx-auto max-w-3xl px-4 py-16 sm:px-6 lg:px-8">
      <section role="alert" className="rounded-2xl border border-crimson/50 bg-crimson/10 p-6">
        <AlertTriangle aria-hidden="true" className="h-7 w-7 text-crimson-light" />
        <h1 className="mt-3 text-xl font-black text-text-main">The public data view could not be rendered</h1>
        <p className="mt-2 text-sm text-text-muted">Try the request again. If the source API is unavailable, the directory will identify that state without substituting records.</p>
        <div className="mt-5 flex flex-wrap gap-3">
          <button type="button" onClick={reset} className="inline-flex min-h-10 items-center gap-2 rounded-lg bg-primary px-4 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
            <RefreshCw aria-hidden="true" className="h-4 w-4" /> Try again
          </button>
          <Link href="/" className="inline-flex min-h-10 items-center rounded-lg border border-border bg-surface px-4 text-xs font-bold text-text-main hover:border-primary hover:text-aurora">Return to Capital Radar</Link>
        </div>
      </section>
    </div>
  );
}
