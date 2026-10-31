import Link from "next/link";
export default function NotFound() {
  return (
    <div className="empty-state">
      <h1>Asset not found</h1>
      <p>This asset is not in the tracked market universe.</p>
      <Link className="button" href="/market">
        Explore markets
      </Link>
    </div>
  );
}
