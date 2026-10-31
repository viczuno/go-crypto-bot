import { describe, it, expect } from "vitest";
import {
  money,
  percent,
  stale,
  availableRanges,
  compareSeries,
  filterAssets,
} from "../lib/market";
import { assetSchema, querySchema, type Asset } from "../lib/contracts";
const asset = (
  id: string,
  points: { time: string; price: number }[],
): Asset => ({
  id,
  name: id,
  symbol: id,
  active: true,
  price: 1,
  cap: null,
  volume: null,
  change: null,
  updated: null,
  history: points,
  analysis: null,
});
describe("market semantics", () => {
  it("distinguishes missing and zero", () => {
    expect(money(null)).toBe("—");
    expect(percent(0)).toBe("+0.00%");
    expect(money(0.000123)).toBe("$0.000123");
  });
  it("marks absent or old observations stale", () => {
    expect(stale(null)).toBe(true);
    expect(
      stale("2025-01-01T00:00:00Z", Date.parse("2025-01-01T19:00:00Z")),
    ).toBe(true);
  });
  it("offers only supported ranges", () => {
    expect(
      availableRanges([
        { time: "2025-01-01", price: 1 },
        { time: "2025-01-10", price: 2 },
      ]),
    ).toEqual(["7D", "All"]);
  });
  it("normalizes at common date and preserves unequal history", () => {
    const a = asset("a", [
      { time: "2025-01-01", price: 10 },
      { time: "2025-01-02", price: 20 },
      { time: "2025-01-04", price: 30 },
    ]);
    const b = asset("b", [
      { time: "2025-01-02", price: 100 },
      { time: "2025-01-03", price: 90 },
      { time: "2025-01-04", price: 110 },
    ]);
    const c = compareSeries([a, b], "All");
    expect(c[0].points).toEqual([
      { time: "2025-01-02", price: 0 },
      { time: "2025-01-04", price: 50 },
    ]);
    expect(c[1].points[2].price).toBeCloseTo(10);
  });
  it("filters and keeps nulls last", () => {
    const a = asset("bitcoin", []),
      b = { ...asset("ethereum", []), cap: 10 };
    expect(filterAssets([a, b], "", "cap", "asc").map((a) => a.id)).toEqual([
      "ethereum",
      "bitcoin",
    ]);
    expect(filterAssets([a, b], "bit", "cap", "desc")).toEqual([a]);
  });
  it("rejects unbounded queries and malformed contracts", () => {
    expect(querySchema.safeParse({ limit: 100000 }).success).toBe(false);
    expect(assetSchema.safeParse({ id: "a" }).success).toBe(false);
  });
});
