import "server-only";
import { Pool } from "pg";
import { unstable_cache } from "next/cache";
import {
  assetSchema,
  analysisSchema,
  type Snapshot,
  type Point,
} from "./contracts";
let pool: Pool | undefined;
function database() {
  if (!process.env.DATABASE_URL) throw new Error("Database unavailable");
  return (pool ??= new Pool({
    connectionString: process.env.DATABASE_URL,
    max: 3,
    idleTimeoutMillis: 10000,
    connectionTimeoutMillis: 5000,
    statement_timeout: 5000,
    options: "-c default_transaction_read_only=on",
  }));
}
export const snapshot = unstable_cache(
  async (): Promise<Snapshot> => {
    const asOf = new Date().toISOString();
    try {
      const db = database();
      const [assets, runs] = await Promise.all([
        db.query(`SELECT a.*,o.price::float8,o.market_cap::float8 cap,o.volume::float8 volume,o.change_24h change,o.observed_at updated,n.result analysis,
 COALESCE(h.history,'[]'::json) history FROM assets a
 LEFT JOIN LATERAL(SELECT * FROM observations WHERE asset_id=a.id ORDER BY observed_at DESC LIMIT 1)o ON true
 LEFT JOIN LATERAL(SELECT result FROM analytics WHERE asset_id=a.id ORDER BY day DESC LIMIT 1)n ON true
 LEFT JOIN LATERAL(SELECT json_agg(json_build_object('time',day::text,'price',price::float8) ORDER BY day) history FROM (SELECT day,price FROM daily_samples WHERE asset_id=a.id ORDER BY day DESC LIMIT 31)d)h ON true
 ORDER BY a.active DESC,o.market_cap DESC NULLS LAST,a.id LIMIT 250`),
        db.query(
          `SELECT status,finished_at,received,expected FROM ingestion_runs ORDER BY id DESC LIMIT 1`,
        ),
      ]);
      const parsed = assets.rows.map((r) =>
        assetSchema.parse({
          ...r,
          updated: r.updated?.toISOString() ?? null,
          analysis: r.analysis ? analysisSchema.parse(r.analysis) : null,
        }),
      );
      const run = runs.rows[0];
      return {
        assets: parsed,
        status: parsed.length ? "ready" : "empty",
        asOf,
        run: run
          ? {
              status: run.status,
              finished: run.finished_at?.toISOString() ?? null,
              received: run.received,
              expected: run.expected,
            }
          : null,
      };
    } catch {
      console.error(JSON.stringify({ event: "market_read_failed" }));
      return { assets: [], status: "unavailable", asOf, run: null };
    }
  },
  ["market-snapshot-v1"],
  { revalidate: 300 },
);
export async function detail(id: string) {
  const data = await snapshot();
  const asset = data.assets.find((a) => a.id === id);
  if (!asset) return { data, asset: null, series: [] };
  try {
    const h = await database().query(
      `SELECT day::text time,price::float8 price FROM (SELECT day,price FROM daily_samples WHERE asset_id=$1 ORDER BY day DESC LIMIT 2000)s ORDER BY day`,
      [id],
    );
    const detailedAsset = { ...asset, history: h.rows as Point[] };
    const r = await database().query(
      `SELECT day::text time,result FROM (SELECT day,result FROM analytics WHERE asset_id=$1 ORDER BY day DESC LIMIT 2000)s ORDER BY day`,
      [id],
    );
    return {
      data,
      asset: detailedAsset,
      series: r.rows.map((r) => ({
        time: r.time,
        analysis: analysisSchema.parse(r.result),
      })),
    };
  } catch {
    return {
      data: { ...data, status: "unavailable" as const },
      asset,
      series: [],
    };
  }
}
export async function health() {
  try {
    await database().query("SELECT 1");
    return true;
  } catch {
    return false;
  }
}
export type IndicatorPoint = {
  time: string;
  analysis: ReturnType<typeof analysisSchema.parse>;
};
export type { Point };
