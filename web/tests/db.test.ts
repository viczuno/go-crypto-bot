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

it.skipIf(!url)(
  "casts nullable market numerics and validates actual analytics rows",
  async () => {
    const { Client } = await import("pg");
    const { marketSnapshotSQL } = await import("../lib/market-query");
    const { assetSchema, analysisSchema } = await import("../lib/contracts");
    const client = new Client({
      connectionString: url,
      options: "-c default_transaction_read_only=on",
    });
    try {
      await client.connect();
      // Read-only CTEs exercise the actual snapshot SELECT with both NUMERIC and
      // double-precision provider fields, without changing any persisted records.
      for (const type of ["numeric", "double precision"]) {
        const result = await client.query(`WITH
        assets AS (SELECT 'test'::text id,'Test'::text name,'TST'::text symbol,false active),
        observations AS (SELECT 'test'::text asset_id,1.25::numeric price,NULL::numeric market_cap,NULL::numeric volume,2.5::${type} change_24h,'2026-10-30T00:00:00Z'::timestamptz observed_at),
        analytics AS (SELECT 'test'::text asset_id,now()::date AS day,NULL::jsonb result),
        daily_samples AS (SELECT 'test'::text asset_id,now()::date AS day,1.25::numeric price)
        ${marketSnapshotSQL}`);
        expect(result.fields.find((f) => f.name === "change")?.dataTypeID).toBe(
          701,
        );
        const row = result.rows[0];
        expect(row.change).toBe(2.5);
        expect(
          assetSchema.safeParse({ ...row, updated: row.updated.toISOString() })
            .success,
        ).toBe(true);
      }
      const actual = await client.query(marketSnapshotSQL);
      for (const row of actual.rows)
        expect(
          assetSchema.safeParse({
            ...row,
            updated: row.updated?.toISOString() ?? null,
          }).success,
        ).toBe(true);
      const analytics = await client.query(
        "SELECT result FROM analytics ORDER BY asset_id,day LIMIT 2000",
      );
      expect(analytics.rows.length).toBeGreaterThan(0);
      for (const row of analytics.rows)
        expect(analysisSchema.safeParse(row.result).success).toBe(true);
    } finally {
      await client.end();
    }
  },
);
