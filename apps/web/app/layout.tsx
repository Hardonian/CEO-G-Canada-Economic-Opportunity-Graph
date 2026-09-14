import type { Metadata, Viewport } from "next";
import "./globals.css";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";
import { LanguageProvider, LocalizedContent } from "@/components/LanguageProvider";
import AdSenseAutoAds from "@/components/AdSenseAutoAds";

export const metadata: Metadata = {
  title: {
    default: "CanadaOpportunityGraph | Independent Canadian Capital Intelligence",
    template: "%s | CanadaOpportunityGraph",
  },
  description: "Independent, open-source intelligence for planning around Canadian infrastructure, capital, procurement, and economic development. CEGS reference implementation.",
  keywords: ["Canada infrastructure", "critical minerals", "nuclear power", "clean energy", "procurement", "CEGS", "economic graph"],
};

export const viewport: Viewport = {
  colorScheme: "dark",
  themeColor: "#050B08",
  width: "device-width",
  initialScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en-CA" data-language="en" className="dark" suppressHydrationWarning>
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: `(function(){try{var l=localStorage.getItem('cog-language');if(l!=='fr'&&l!=='en'){l=(document.cookie.match(/(?:^|; )cog-language=(en|fr)/)||[])[1]}if(l==='fr'){document.documentElement.lang='fr-CA';document.documentElement.dataset.language='fr'}}catch(e){}})();`,
          }}
        />
      </head>
      <body className="flex min-h-screen flex-col bg-background text-text-main antialiased">
        <LanguageProvider>
          <LocalizedContent>
            <a href="#main-content" className="skip-link">
              Skip to main content
            </a>
            <Navbar />
            <main id="main-content" tabIndex={-1} className="flex-grow" {...{"google-side-rail-overlap": "false"}}>
              {children}
            </main>
            <Footer />
            <AdSenseAutoAds publisherId={process.env.NEXT_PUBLIC_ADSENSE_PUBLISHER_ID} />
          </LocalizedContent>
        </LanguageProvider>
      </body>
    </html>
  );
}
