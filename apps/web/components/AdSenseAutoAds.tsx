"use client";

import Script from "next/script";
import { useEffect, useState } from "react";

type Consent = "granted" | "denied" | null;

export default function AdSenseAutoAds({ publisherId }: { publisherId?: string }) {
  const validPublisherId = typeof publisherId === "string" && /^ca-pub-\d{16}$/.test(publisherId)
    ? publisherId
    : null;
  const [consent, setConsent] = useState<Consent>(null);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!validPublisherId) return;
    const stored = window.localStorage.getItem("cog-ad-consent-v1");
    setConsent(stored === "granted" || stored === "denied" ? stored : null);
    setReady(true);
  }, [validPublisherId]);

  if (!validPublisherId || !ready) return null;

  const choose = (value: Exclude<Consent, null>) => {
    window.localStorage.setItem("cog-ad-consent-v1", value);
    setConsent(value);
  };

  return (
    <>
      {consent === "granted" ? (
        <Script
          id="cog-adsense-auto-ads"
          async
          strategy="afterInteractive"
          crossOrigin="anonymous"
          data-privacy-treatments="disablePersonalization"
          src={`https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${validPublisherId}`}
        />
      ) : null}
      {consent === null ? (
        <aside
          aria-label="Advertising preference"
          className="fixed inset-x-3 bottom-3 z-[80] mx-auto max-w-3xl rounded-2xl border border-border bg-card/95 p-4 shadow-2xl backdrop-blur-xl"
          data-print-hide="true"
        >
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-xs leading-relaxed text-text-muted">
              Optional, non-personalized advertising helps fund this open public-interest research. Ads load only if you allow them; declining does not limit the site.
            </p>
            <div className="flex shrink-0 gap-2">
              <button type="button" onClick={() => choose("denied")} className="min-h-10 rounded-lg border border-border px-4 text-xs font-bold text-text-main hover:border-primary">
                Continue without ads
              </button>
              <button type="button" onClick={() => choose("granted")} className="min-h-10 rounded-lg bg-primary px-4 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
                Allow ads
              </button>
            </div>
          </div>
        </aside>
      ) : null}
    </>
  );
}
