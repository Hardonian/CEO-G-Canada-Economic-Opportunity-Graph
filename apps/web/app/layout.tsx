import type { Metadata } from "next";
import { Analytics } from "@vercel/analytics/next";
import "./globals.css";
import Navbar from "@/components/Navbar";
import Footer from "@/components/Footer";

export const metadata: Metadata = {
  title: "CanadaOpportunityGraph | The National Infrastructure & Economic Graph",
  description: "Know what Canada is building before everyone else does. Follow Canadian capital before it becomes Canadian construction. Official CEGS reference implementation.",
  keywords: ["Canada infrastructure", "critical minerals", "nuclear power", "clean energy", "procurement", "CEGS", "economic graph"],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en-CA" className="dark">
      <body className="min-h-screen flex flex-col bg-background text-text-main antialiased selection:bg-primary/20 selection:text-white">
        <Navbar />
        <main className="flex-grow">{children}</main>
        <Footer />
        <Analytics />
      </body>
    </html>
  );
}
