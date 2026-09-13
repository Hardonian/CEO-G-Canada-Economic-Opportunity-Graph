import type { Metadata, Viewport } from "next";
import "./globals.css";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

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
    <html lang="en-CA" className="dark">
      <body className="flex min-h-screen flex-col bg-background text-text-main antialiased">
        <a href="#main-content" className="skip-link">
          Skip to main content
        </a>
        <Navbar />
        <main id="main-content" tabIndex={-1} className="flex-grow">
          {children}
        </main>
        <Footer />
      </body>
    </html>
  );
}
