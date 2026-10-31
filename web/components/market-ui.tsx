"use client";
import Link from "next/link";
import dynamic from "next/dynamic";
import { useState, useMemo, useSyncExternalStore, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  ArrowUpRight,
  ArrowDownRight,
  ArrowRight,
  Search,
  SlidersHorizontal,
  LayoutGrid,
  List,
  Clock,
  Database,
  GitCompareArrows,
  RefreshCw,
} from "lucide-react";
import {
  assetSchema,
  type Asset,
  type Snapshot,
  type Point,
} from "@/lib/contracts";
import {
  money,
  percent,
  stale,
  filterAssets,
  availableRanges,
  sliceRange,
  compareSeries,
  type Range,
} from "@/lib/market";
import { Favorite, useWatch } from "./shell";
const Chart = dynamic(() => import("./chart"), {
  ssr: false,
  loading: () => <div className="chart-loading">Loading chart…</div>,
});
const Distribution = dynamic(() => import("./distribution"), { ssr: false });
export function Heading({
  eyebrow,
  title,
  description,
  children,
}: {
  eyebrow: string;
  title: string;
  description: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="page-heading">
      <div>
        <div className="eyebrow">{eyebrow}</div>
        <h1>{title}</h1>
        <p>{description}</p>
      </div>
      {children}
    </div>
  );
}
export function State({ data }: { data: Snapshot }) {
  const router = useRouter();
  if (data.status === "ready") return null;
  return (
    <div className="empty-state">
      <Database size={32} />
      <h2>
        {data.status === "unavailable"
          ? "Market data is unavailable"
          : "Your market is ready to begin"}
      </h2>
      <p>
        {data.status === "unavailable"
          ? "We could not load the market dataset. Please try again shortly."
          : "No observations have been collected yet. Historical prices will appear after the first successful import or update."}
      </p>
      <button className="button" onClick={() => router.refresh()}>
        <RefreshCw size={16} />
        Try again
      </button>
    </div>
  );
}
export function Freshness({ data }: { data: Snapshot }) {
  const times = data.assets
    .flatMap((a) => (a.updated ? [a.updated] : []))
    .sort();
  const time = times.at(-1);
  const old = data.assets.filter((a) =>
    stale(a.updated, Date.parse(data.asOf)),
  ).length;
  return (
    <div className="freshness">
      <span>
        <i className={old ? "warning" : ""} />
        {old
          ? `${old} assets have stale observations`
          : "Latest available observations"}
      </span>
      <span>
        {data.run && data.run.status !== "success"
          ? `Ingestion ${data.run.status} · ${data.run.received}/${data.run.expected} assets · `
          : ""}
        {time
          ? `Updated ${new Date(time).toLocaleString("en-GB", { timeZone: "UTC", dateStyle: "medium", timeStyle: "short" })} UTC`
          : "Awaiting first update"}
      </span>
    </div>
  );
}
function Coin({ asset }: { asset: Asset }) {
  return (
    <Link className="coin" href={`/coins/${asset.id}`}>
      <span className={`coin-avatar tone-${asset.id.charCodeAt(0) % 5}`}>
        {asset.symbol.slice(0, 1).toUpperCase()}
      </span>
      <span>
        <strong>{asset.name}</strong>
        <small>{asset.symbol.toUpperCase()}</small>
      </span>
    </Link>
  );
}
export function Sparkline({ points }: { points: Point[] }) {
  if (points.length < 2)
    return <span className="muted">Collecting history</span>;
  const ps = points.slice(-30);
  const min = Math.min(...ps.map((p) => p.price)),
    max = Math.max(...ps.map((p) => p.price));
  const path = ps
    .map(
      (p, i) =>
        `${(i / (ps.length - 1)) * 100},${28 - ((p.price - min) / (max - min || 1)) * 24}`,
    )
    .join(" ");
  return (
    <svg
      width="100"
      height="32"
      viewBox="0 0 100 32"
      role="img"
      aria-label="Recent daily price trend"
    >
      <polyline
        points={path}
        stroke={ps.at(-1)!.price >= ps[0].price ? "#31c9ac" : "#ed8a8a"}
        fill="none"
        strokeWidth="1.8"
      />
    </svg>
  );
}
function Change({ value }: { value: number | null }) {
  return (
    <span
      className={
        value === null ? "muted" : value >= 0 ? "positive" : "negative"
      }
    >
      {value !== null &&
        (value >= 0 ? (
          <ArrowUpRight size={13} />
        ) : (
          <ArrowDownRight size={13} />
        ))}
      {percent(value)}
    </span>
  );
}
export function MarketTable({
  assets,
  compact = false,
}: {
  assets: Asset[];
  compact?: boolean;
}) {
  return (
    <div className="table-scroll">
      <table className={compact ? "compact" : ""}>
        <thead>
          <tr>
            <th aria-label="Watchlist" />
            <th>Asset</th>
            <th className="number">Price</th>
            <th className="number">24h change</th>
            <th className="number">Market cap</th>
            <th className="number">24h volume</th>
            <th>30-day samples</th>
            <th>Daily signal</th>
          </tr>
        </thead>
        <tbody>
          {assets.map((a) => (
            <tr key={a.id}>
              <td>
                <Favorite id={a.id} />
              </td>
              <td>
                <Coin asset={a} />
              </td>
              <td className="number strong">{money(a.price)}</td>
              <td className="number">
                <Change value={a.change} />
              </td>
              <td className="number muted">{money(a.cap, true)}</td>
              <td className="number muted">{money(a.volume, true)}</td>
              <td>
                <Sparkline points={a.history} />
              </td>
              <td>
                <span className="signal">
                  {a.analysis?.signal ?? "Unavailable"}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {!assets.length && (
        <p className="empty-inline">No assets match these filters.</p>
      )}
    </div>
  );
}
export function Dashboard({ data }: { data: Snapshot }) {
  const { ids } = useWatch();
  const assets = data.assets.filter((a) => a.active);
  const displayed = assets.length ? assets : data.assets;
  const priced = displayed.filter((a) => a.price !== null);
  const caps = displayed.filter((a) => a.cap !== null);
  const volume = displayed.filter((a) => a.volume !== null);
  const moves = displayed
    .filter((a) => a.change !== null)
    .sort((a, b) => b.change! - a.change!);
  const gain = moves[0],
    loss = moves.at(-1);
  const lead = displayed.find((a) => a.id === "bitcoin") ?? displayed[0];
  const distribution = ["Buy", "Hold", "Sell", "Unavailable"].map((name) => ({
    name,
    value: displayed.filter((a) =>
      name === "Unavailable"
        ? !a.analysis || a.analysis.coverage === 0
        : a.analysis?.signal.includes(name),
    ).length,
  }));
  const watch = displayed.filter((a) => ids.includes(a.id));
  return (
    <>
      <Heading
        eyebrow="THE BIG PICTURE"
        title="Market overview"
        description="A measured view of crypto. Price, momentum, and the context between."
      >
        <Link className="button primary" href="/market">
          Explore markets <ArrowUpRight size={16} />
        </Link>
      </Heading>
      <Freshness data={data} />
      <State data={data} />
      {data.status === "ready" && (
        <>
          <section className="metric-grid">
            <Metric
              label="TRACKED ASSETS"
              value={String(displayed.length)}
              detail={`${priced.length} with price observations`}
              icon={<Database size={18} />}
            />
            <Metric
              label="TRACKED MARKET CAP"
              value={
                caps.length
                  ? money(
                      caps.reduce((s, a) => s + a.cap!, 0),
                      true,
                    )
                  : "—"
              }
              detail={`Reported for ${caps.length}/${displayed.length} assets`}
              icon={<ArrowUpRight size={18} />}
            />
            <Metric
              label="TRACKED 24H VOLUME"
              value={
                volume.length
                  ? money(
                      volume.reduce((s, a) => s + a.volume!, 0),
                      true,
                    )
                  : "—"
              }
              detail={`Reported for ${volume.length}/${displayed.length} assets`}
              icon={<ActivityIcon />}
            />
            <Metric
              label="ADVANCING / DECLINING"
              value={`${moves.filter((a) => a.change! > 0).length} / ${moves.filter((a) => a.change! < 0).length}`}
              detail={`${moves.length} assets with 24h change`}
              icon={<GitCompareArrows size={18} />}
            />
          </section>
          <div className="overview-grid">
            <section className="panel feature-chart">
              <div className="panel-heading">
                <div>
                  <span className="eyebrow">PRICE PERSPECTIVE</span>
                  <h2>
                    {lead?.name ?? "Historical prices"}{" "}
                    <span className="muted small">/ USD</span>
                  </h2>
                </div>
                {lead && (
                  <Link href={`/coins/${lead.id}`} className="text-link">
                    Full analysis <ArrowUpRight size={15} />
                  </Link>
                )}
              </div>
              {lead && (
                <>
                  <div className="hero-price">
                    {money(lead.price)}{" "}
                    <span>
                      <Change value={lead.change} />
                    </span>
                  </div>
                  <PriceChart asset={lead} />
                </>
              )}
            </section>
            <section className="panel">
              <div className="panel-heading">
                <div>
                  <span className="eyebrow">TECHNICAL OUTLOOK</span>
                  <h2>Signal balance</h2>
                </div>
                <span className="tag">Daily</span>
              </div>
              <Distribution data={distribution} />
              <p className="panel-note">
                Weighted technical signals across tracked assets. Missing
                history is shown separately. Signal strength is heuristic, not a
                probability of profit.
              </p>
              <div className="movers">
                {[
                  { label: "STRONGEST 24H", a: gain },
                  { label: "WEAKEST 24H", a: loss },
                ].map(({ label, a }) => (
                  <div key={label}>
                    <span className="eyebrow">{label}</span>
                    {a ? (
                      <>
                        <Coin asset={a} />
                        <Change value={a.change} />
                      </>
                    ) : (
                      <p className="muted">Not available</p>
                    )}
                  </div>
                ))}
              </div>
            </section>
          </div>
          <section className="panel market-panel">
            <div className="panel-heading">
              <div>
                <span className="eyebrow">YOUR MARKET UNIVERSE</span>
                <h2>Market snapshot</h2>
              </div>
              <Link href="/market" className="text-link">
                View all assets <ArrowRight size={16} />
              </Link>
            </div>
            <MarketTable
              assets={[...displayed]
                .sort((a, b) => (b.cap ?? 0) - (a.cap ?? 0))
                .slice(0, 8)}
            />
          </section>
          <div className="two-columns">
            <section className="panel">
              <div className="panel-heading">
                <h2>Your watchlist</h2>
                <Link className="text-link" href="/watchlist">
                  Manage <ArrowRight size={15} />
                </Link>
              </div>
              {watch.length ? (
                watch.slice(0, 4).map((a) => (
                  <div key={a.id} className="asset-row">
                    <Coin asset={a} />
                    <Change value={a.change} />
                  </div>
                ))
              ) : (
                <p className="panel-note">
                  Star an asset to keep it close. Your watchlist stays in this
                  browser.
                </p>
              )}
            </section>
            <section className="panel">
              <div className="panel-heading">
                <h2>Latest observations</h2>
                <Clock size={17} />
              </div>
              {[...displayed]
                .sort((a, b) =>
                  (b.updated ?? "").localeCompare(a.updated ?? ""),
                )
                .slice(0, 4)
                .map((a) => (
                  <div key={a.id} className="asset-row">
                    <Coin asset={a} />
                    <span className="muted small">
                      {a.updated
                        ? new Date(a.updated)
                            .toISOString()
                            .slice(0, 16)
                            .replace("T", " ") + " UTC"
                        : "No observation"}
                    </span>
                  </div>
                ))}
            </section>
          </div>
        </>
      )}
    </>
  );
}
function ActivityIcon() {
  return <SlidersHorizontal size={18} />;
}
function Metric({
  label,
  value,
  detail,
  icon,
}: {
  label: string;
  value: string;
  detail: string;
  icon: React.ReactNode;
}) {
  return (
    <div className="metric">
      <div className="metric-label">
        {label}
        {icon}
      </div>
      <div className="metric-value">{value}</div>
      <p>{detail}</p>
    </div>
  );
}
export function PriceChart({
  asset,
  extra = [],
}: {
  asset: Asset;
  extra?: { name: string; points: Point[] }[];
}) {
  const [range, setRange] = useState<Range>("All");
  const points = useMemo(
    () => sliceRange(asset.history, range),
    [asset.history, range],
  );
  const series = useMemo(
    () => [
      { name: asset.symbol.toUpperCase(), points },
      ...extra.map((s) => ({
        ...s,
        points: s.points.filter(
          (p) => points.length && p.time >= points[0].time,
        ),
      })),
    ],
    [asset.symbol, points, extra],
  );
  return (
    <>
      <div className="chart-toolbar">
        <span className="muted small">Completed UTC-day samples</span>
        <div className="segmented" aria-label="Chart timeframe">
          {availableRanges(asset.history).map((r) => (
            <button
              key={r}
              aria-pressed={range === r}
              onClick={() => setRange(r)}
            >
              {r}
            </button>
          ))}
        </div>
      </div>
      {points.length > 1 ? (
        <Chart series={series} />
      ) : (
        <div className="chart-loading">
          At least two completed daily samples are needed for a chart.
        </div>
      )}
      <details className="data-details">
        <summary>View chart data ({points.length} samples)</summary>
        <div className="data-scroll">
          <table>
            <thead>
              <tr>
                <th>UTC date</th>
                <th>USD price</th>
              </tr>
            </thead>
            <tbody>
              {points.map((p) => (
                <tr key={p.time}>
                  <td>{p.time}</td>
                  <td>{money(p.price)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </details>
    </>
  );
}
export function Explorer({
  data,
  watchlist = false,
}: {
  data: Snapshot;
  watchlist?: boolean;
}) {
  const { ids } = useWatch();
  const params = useSearchParams();
  const router = useRouter();
  const q = params.get("q") ?? "";
  const sort = params.get("sort") ?? "cap";
  const direction = params.get("direction") ?? "desc";
  const [signal, setSignal] = useState("all");
  const [freshness, setFreshness] = useState("all");
  const [favorites, setFavorites] = useState(false);
  const compact = useSyncExternalStore(
    (fn) => {
      window.addEventListener("density", fn);
      return () => window.removeEventListener("density", fn);
    },
    () => {
      try {
        return localStorage.getItem("crypto-density-v1") === "compact";
      } catch {
        return false;
      }
    },
    () => false,
  );
  function setCompact(value: boolean) {
    try {
      localStorage.setItem(
        "crypto-density-v1",
        value ? "compact" : "comfortable",
      );
      window.dispatchEvent(new Event("density"));
    } catch {}
  }

  const [page, setPage] = useState(1);
  function update(key: string, v: string) {
    const p = new URLSearchParams(params);
    p.set(key, v);
    router.replace(`?${p.toString()}`, { scroll: false });
    setPage(1);
  }
  const assets = filterAssets(
    data.assets,
    q,
    sort,
    direction,
    watchlist || favorites ? ids : null,
    signal,
    freshness,
    Date.parse(data.asOf),
  );
  const pages = Math.max(1, Math.ceil(assets.length / 25));
  const current = Math.min(page, pages);
  return (
    <>
      <Heading
        eyebrow={watchlist ? "PERSONAL PERSPECTIVE" : "DISCOVER & RESEARCH"}
        title={watchlist ? "Your watchlist" : "Market explorer"}
        description={
          watchlist
            ? "The assets you follow, together in one place. Saved in this browser."
            : "Find your next research question. Explore the assets in our tracked universe."
        }
      />
      <Freshness data={data} />
      <State data={data} />
      {data.status === "ready" && (
        <section className="panel market-panel">
          <div className="filters">
            <label className="search">
              <Search size={17} />
              <input
                placeholder="Search assets…"
                aria-label="Search assets"
                value={q}
                onChange={(e) => update("q", e.target.value)}
              />
            </label>
            <select
              aria-label="Sort assets"
              value={sort}
              onChange={(e) => update("sort", e.target.value)}
            >
              <option value="cap">Market cap</option>
              <option value="price">Price</option>
              <option value="change">24h change</option>
              <option value="volume">Volume</option>
              <option value="name">Name</option>
            </select>
            <button
              className="button"
              onClick={() =>
                update("direction", direction === "asc" ? "desc" : "asc")
              }
            >
              {direction === "asc" ? "Ascending" : "Descending"}
            </button>
            <select
              aria-label="Signal filter"
              value={signal}
              onChange={(e) => {
                setSignal(e.target.value);
                setPage(1);
              }}
            >
              {["all", "Buy", "Hold", "Sell"].map((v) => (
                <option key={v} value={v}>
                  {v === "all" ? "All signals" : v}
                </option>
              ))}
            </select>
            <select
              aria-label="Freshness filter"
              value={freshness}
              onChange={(e) => {
                setFreshness(e.target.value);
                setPage(1);
              }}
            >
              <option value="all">Any freshness</option>
              <option value="fresh">Fresh</option>
              <option value="stale">Stale</option>
            </select>
            {!watchlist && (
              <button
                className="button"
                aria-pressed={favorites}
                onClick={() => {
                  setFavorites(!favorites);
                  setPage(1);
                }}
              >
                Favorites
              </button>
            )}
            <button
              className="icon-button"
              aria-label="Toggle compact rows"
              aria-pressed={compact}
              onClick={() => {
                setCompact(!compact);
                try {
                  localStorage.setItem(
                    "crypto-density-v1",
                    compact ? "comfortable" : "compact",
                  );
                } catch {}
              }}
            >
              {compact ? <List size={18} /> : <LayoutGrid size={18} />}
            </button>
          </div>
          <MarketTable
            assets={assets.slice((current - 1) * 25, current * 25)}
            compact={compact}
          />
          <div className="pagination">
            <span>
              {assets.length} assets · Page {current} of {pages}
            </span>
            <div>
              <button
                className="button"
                disabled={current === 1}
                onClick={() => setPage(current - 1)}
              >
                Previous
              </button>
              <button
                className="button"
                disabled={current === pages}
                onClick={() => setPage(current + 1)}
              >
                Next
              </button>
            </div>
          </div>
        </section>
      )}
    </>
  );
}
export function Comparison({ data }: { data: Snapshot }) {
  const params = useSearchParams();
  const initial = (params.get("ids") ?? "bitcoin,ethereum")
    .split(",")
    .filter((id) => data.assets.some((a) => a.id === id))
    .slice(0, 5);
  const [selected, setSelected] = useState(initial);
  const [range, setRange] = useState<Range>("All");
  const [loaded, setLoaded] = useState<Asset[]>([]);
  const [loadState, setLoadState] = useState("loading");
  const selectedKey = selected.join(",");
  useEffect(() => {
    const controller = new AbortController();
    fetch(`/api/compare?ids=${encodeURIComponent(selectedKey)}`, {
      signal: controller.signal,
    })
      .then(async (r) => {
        if (!r.ok) throw new Error();
        return r.json();
      })
      .then((d) => {
        setLoaded(d.assets.map((asset: unknown) => assetSchema.parse(asset)));
        setLoadState("ready");
      })
      .catch(() => {
        if (!controller.signal.aborted) {
          setLoaded([]);
          setLoadState("error");
        }
      });
    return () => controller.abort();
  }, [selectedKey]);
  const assets = loaded.filter((a) => selected.includes(a.id));
  const series = useMemo(() => compareSeries(assets, range), [assets, range]);
  return (
    <>
      <Heading
        eyebrow="RELATIVE PERFORMANCE"
        title="Compare assets"
        description="Different prices. A shared starting point. Compare up to five assets by percentage performance."
      />
      <State data={data} />
      <section className="panel">
        {loadState !== "ready" && (
          <p role="status">
            {loadState === "error"
              ? "Comparison data unavailable. Change the selection to retry."
              : "Loading comparison history…"}
          </p>
        )}
        <div className="comparison-select">
          {data.assets.map((a) => (
            <label
              key={a.id}
              className={`choice ${selected.includes(a.id) ? "chosen" : ""}`}
            >
              <input
                type="checkbox"
                checked={selected.includes(a.id)}
                disabled={selected.length === 5 && !selected.includes(a.id)}
                onChange={() =>
                  setSelected(
                    selected.includes(a.id)
                      ? selected.filter((id) => id !== a.id)
                      : [...selected, a.id],
                  )
                }
              />
              {a.symbol.toUpperCase()}
            </label>
          ))}
        </div>
        <div className="chart-toolbar">
          <span className="muted">
            Normalized from first shared completed UTC day
          </span>
          <div className="segmented">
            {(["7D", "30D", "90D", "1Y", "All"] as Range[])
              .filter(
                (r) =>
                  r === "All" ||
                  assets.every((a) => availableRanges(a.history).includes(r)),
              )
              .map((r) => (
                <button
                  key={r}
                  aria-pressed={range === r}
                  onClick={() => setRange(r)}
                >
                  {r}
                </button>
              ))}
          </div>
        </div>
        {series.length && series[0].points.length > 1 ? (
          <Chart series={series} percentage height={400} />
        ) : (
          <div className="empty-state">
            <GitCompareArrows />
            <h2>Choose assets with overlapping history</h2>
            <p>
              At least two assets and two shared-period samples are needed.
              Missing days are never interpolated.
            </p>
          </div>
        )}
        <div className="comparison-summary">
          {series.map((s) => (
            <div key={s.name}>
              <span>{s.name}</span>
              <strong>{percent(s.points.at(-1)?.price ?? null)}</strong>
              <small>Since {s.points[0]?.time}</small>
            </div>
          ))}
        </div>
      </section>
    </>
  );
}
