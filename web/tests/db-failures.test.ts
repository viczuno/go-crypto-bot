import { beforeEach, it, expect, vi } from "vitest";
const { query } = vi.hoisted(() => ({ query: vi.fn() }));
vi.mock("server-only", () => ({}));
vi.mock("next/cache", () => ({ unstable_cache: (fn: unknown) => fn }));
vi.mock("pg", () => ({
  Pool: class {
    query = query;
  },
}));
import { snapshot } from "../lib/db";
beforeEach(() => {
  query.mockReset();
  process.env.DATABASE_URL = "unused-in-mock";
});
const archived = {
  id: "bitcoin",
  name: "Bitcoin",
  symbol: "BTC",
  active: false,
  price: 77146,
  cap: null,
  volume: null,
  change: null,
  updated: new Date("2026-10-30T17:40:47Z"),
  history: [{ time: "2026-10-29", price: 78000 }],
  analysis: {
    day: "2026-10-29",
    samples: 1,
    signal: "Unavailable",
    strength: 0,
    coverage: 0,
    indicators: [
      {
        name: "MovingAverage",
        value: null,
        signal: "Unavailable",
        strength: 0,
        values: {},
        required: 200,
        reason: "Insufficient consecutive completed UTC days",
      },
    ],
  },
};
it("accepts imported history with null market metadata and unavailable indicators", async () => {
  query
    .mockResolvedValueOnce({ rows: [archived] })
    .mockResolvedValueOnce({ rows: [] });
  const result = await snapshot();
  expect(result.status).toBe("ready");
  expect(result.assets[0]).toMatchObject({
    cap: null,
    change: null,
    updated: "2026-10-30T17:40:47.000Z",
  });
});
it("classifies permission failures while health-style connectivity can succeed", async () => {
  const log = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    query
      .mockRejectedValueOnce(
        Object.assign(new Error("permission denied secret"), { code: "42501" }),
      )
      .mockResolvedValueOnce({ rows: [] });
    expect((await snapshot()).status).toBe("unavailable");
    expect(JSON.parse(log.mock.calls[0][0])).toEqual({
      event: "market_read_failed",
      stage: "assets_query",
      category: "permission",
      sqlstate: "42501",
    });
  } finally {
    log.mockRestore();
  }
});
it("identifies malformed analytics through safe validation paths", async () => {
  const log = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    query
      .mockResolvedValueOnce({
        rows: [
          {
            ...archived,
            analysis: { ...archived.analysis, coverage: "secret" },
          },
        ],
      })
      .mockResolvedValueOnce({ rows: [] });
    expect((await snapshot()).status).toBe("unavailable");
    const event = JSON.parse(log.mock.calls[0][0]);
    expect(event.category).toBe("validation");
    expect(event.issues[0].path).toEqual(["coverage"]);
    expect(JSON.stringify(event)).not.toContain("secret");
  } finally {
    log.mockRestore();
  }
});
