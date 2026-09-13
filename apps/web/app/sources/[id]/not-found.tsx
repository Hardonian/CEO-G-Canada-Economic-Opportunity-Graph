import Link from "next/link";
import { ArrowLeft, FileQuestion } from "lucide-react";

export default function SourceNotFound() {
  return (
    <div className="mx-auto max-w-3xl px-4 py-16 text-center sm:px-6 lg:px-8">
      <FileQuestion aria-hidden="true" className="mx-auto h-10 w-10 text-text-subtle" />
      <h1 className="mt-4 text-2xl font-black text-text-main">Source record not found</h1>
      <p className="mx-auto mt-2 max-w-xl text-sm text-text-muted">The requested identifier is not present in the public source census. Retired sources remain registered when historical metadata is available.</p>
      <Link href="/sources" className="mt-6 inline-flex min-h-10 items-center gap-2 rounded-lg border border-primary/60 bg-primary/10 px-4 text-xs font-bold text-aurora hover:bg-primary/20">
        <ArrowLeft aria-hidden="true" className="h-4 w-4" /> Browse public sources
      </Link>
    </div>
  );
}
