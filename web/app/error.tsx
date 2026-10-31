"use client";
export default function ErrorPage({ reset }: { reset: () => void }) {
  return (
    <div className="empty-state">
      <h1>Unable to load this view</h1>
      <p>Please retry. Your saved watchlist is still in this browser.</p>
      <button className="button" onClick={reset}>
        Try again
      </button>
    </div>
  );
}
