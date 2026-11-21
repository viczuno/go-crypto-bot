import { expect, it, vi } from "vitest";
const { query } = vi.hoisted(() => ({ query: vi.fn() }));
vi.mock("server-only", () => ({}));
vi.mock("pg", () => ({
  Pool: class {
    query = query;
  },
}));
import { snapshot } from "../lib/db";
it("reads fresh rows on every call and recovers after a transient SQL failure", async () => {
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
    // No snapshot cache: a later call must query again so data is never stale.
    query
      .mockResolvedValueOnce({ rows: [] })
      .mockResolvedValueOnce({ rows: [] });
    const again = await snapshot();
    expect(query).toHaveBeenCalledTimes(6);
    expect(again.status).toBe("empty");
    expect(Date.parse(again.asOf)).toBeLessThanOrEqual(Date.now());
    expect(JSON.parse(log.mock.calls[0][0])).toMatchObject({
      category: "database",
      sqlstate: "08006",
    });
    expect(JSON.stringify(log.mock.calls)).not.toContain("sensitive");
  } finally {
    log.mockRestore();
  }
});
