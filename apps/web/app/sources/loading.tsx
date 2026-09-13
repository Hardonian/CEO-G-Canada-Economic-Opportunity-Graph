export default function SourcesLoading() {
  return (
    <div className="mx-auto max-w-7xl space-y-6 px-4 py-8 sm:px-6 lg:px-8" aria-busy="true" aria-label="Loading public source directory">
      <div className="h-64 animate-pulse rounded-2xl border border-border bg-card" />
      <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
        {Array.from({ length: 5 }, (_, index) => <div key={index} className="h-24 animate-pulse rounded-xl border border-border bg-surface" />)}
      </div>
      <div className="grid gap-4 lg:grid-cols-2">
        {Array.from({ length: 4 }, (_, index) => <div key={index} className="h-72 animate-pulse rounded-2xl border border-border bg-card" />)}
      </div>
      <p className="sr-only">Loading public source directory…</p>
    </div>
  );
}
