import { it, expect, vi } from "vitest";
vi.mock("server-only", () => ({}));
vi.mock("next/cache", () => ({ unstable_cache: (fn: unknown) => fn }));
const url = process.env.MARKET_TEST_DATABASE_URL;
it.skipIf(!url)("reads the existing imported PostgreSQL dataset", async () => {
  process.env.DATABASE_URL = url;
  const { snapshot } = await import("../lib/db");
  const data = await snapshot();
  expect(data.status).toBe("ready");
  expect(data.assets.length).toBeGreaterThan(0);
});
