import { expect, it, vi } from "vitest";
const { query } = vi.hoisted(() => ({ query: vi.fn() }));
vi.mock("server-only", () => ({}));
vi.mock("pg", () => ({
  Pool: class {
    query = query;
  },
}));
// Cache successful values like unstable_cache; rejected reads remain retryable.
vi.mock("next/cache", () => ({
  unstable_cache: (read: () => Promise<unknown>) => {
    let stored: unknown;
    let cached = false;
    return async () => {
      if (cached) return stored;
      stored = await read();
      cached = true;
      return stored;
    };
  },
}));
import { snapshot } from "../lib/db";
it("does not cache unavailable fallbacks after a transient SQL failure", async () => {
  process.env.DATABASE_URL = "unused-in-mock";
  const log = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    query
      .mockRejectedValueOnce(
        Object.assign(new Error("sensitive server message"), { code: "08006" }),
      )
      .mockResolvedValueOnce({ rows: [] });
    expect((await snapshot()).status).toBe("unavailable");
    query
      .mockResolvedValueOnce({ rows: [] })
      .mockResolvedValueOnce({ rows: [] });
    expect((await snapshot()).status).toBe("empty");
    expect(query).toHaveBeenCalledTimes(4);
    expect((await snapshot()).status).toBe("empty");
    expect(query).toHaveBeenCalledTimes(4);
    expect(JSON.parse(log.mock.calls[0][0])).toMatchObject({
      category: "database",
      sqlstate: "08006",
    });
    expect(JSON.stringify(log.mock.calls)).not.toContain("sensitive");
  } finally {
    log.mockRestore();
  }
});
