import Link from "next/link";

export default function NotFound() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-16 text-center sm:px-6 lg:px-8">
      <p className="font-mono text-xs font-bold text-aurora">404</p>
      <h1 className="mt-3 text-3xl font-black text-text-main">Page not found</h1>
      <p className="mx-auto mt-3 max-w-xl text-sm text-text-muted">The requested page or record is not present in the reviewed public snapshot.</p>
      <Link href="/" className="mt-6 inline-flex min-h-10 items-center rounded-lg bg-primary px-4 text-xs font-black text-[#050b08] hover:bg-aurora-mint">
        Return to Capital Radar
      </Link>
    </div>
  );
}
