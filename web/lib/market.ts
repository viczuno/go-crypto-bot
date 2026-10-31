import type { Asset, Point } from "./contracts";
export function money(v: number | null, compact = false) {
  if (v === null || !Number.isFinite(v)) return "—";
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    notation: compact ? "compact" : "standard",
    maximumFractionDigits: compact ? 2 : v < 1 ? 6 : 2,
  }).format(v);
}
export function percent(v: number | null) {
  return v === null || !Number.isFinite(v)
    ? "—"
    : `${v >= 0 ? "+" : ""}${v.toFixed(2)}%`;
}
export function stale(updated: string | null, now = Date.now()) {
  return !updated || now - Date.parse(updated) > 18 * 3600 * 1000;
}
export function filterAssets(
  assets: Asset[],
  q: string,
  sort: string,
  direction: string,
  favorites: string[] | null = null,
  signal = "all",
  freshness = "all",
  now = Date.now(),
) {
  return assets
    .filter(
      (a) =>
        (a.name + " " + a.symbol).toLowerCase().includes(q.toLowerCase()) &&
        (!favorites || favorites.includes(a.id)) &&
        (signal === "all" || a.analysis?.signal.includes(signal)) &&
        (freshness === "all" ||
          (freshness === "stale") === stale(a.updated, now)),
    )
    .sort((a, b) => {
      const key = sort as "cap" | "price" | "change" | "volume";
      if (sort === "name")
        return (direction === "asc" ? 1 : -1) * a.name.localeCompare(b.name);
      const x = a[key],
        y = b[key];
      if (x === null) return y === null ? a.id.localeCompare(b.id) : 1;
      if (y === null) return -1;
      return (
        (direction === "asc" ? 1 : -1) * (x - y) || a.id.localeCompare(b.id)
      );
    });
}
export const ranges = {
  "7D": 7,
  "30D": 30,
  "90D": 90,
  "1Y": 365,
  All: Infinity,
};
export type Range = keyof typeof ranges;
export function availableRanges(points: Point[]): Range[] {
  if (points.length < 2) return ["All"];
  const days =
    (Date.parse(points.at(-1)!.time) - Date.parse(points[0].time)) / 86400000;
  return (Object.keys(ranges) as Range[]).filter(
    (k) => ranges[k] <= days || k === "All",
  );
}
export function sliceRange(points: Point[], range: Range) {
  const end = Date.parse(points.at(-1)?.time ?? "");
  return points.filter(
    (p) => Date.parse(p.time) >= end - ranges[range] * 86400000,
  );
}
export function compareSeries(assets: Asset[], range: Range) {
  const series = assets.map((a) => ({
    id: a.id,
    name: a.symbol.toUpperCase(),
    points: sliceRange(a.history, range),
  }));
  if (series.length < 2) return [];
  const common = series[0].points.find((p) =>
    series.every((s) => s.points.some((x) => x.time === p.time)),
  );
  if (!common) return [];
  return series.map((s) => {
    const base = s.points.find((p) => p.time === common.time)!.price;
    return {
      name: s.name,
      points: s.points
        .filter((p) => p.time >= common.time)
        .map((p) => ({ time: p.time, price: (p.price / base - 1) * 100 })),
    };
  });
}
