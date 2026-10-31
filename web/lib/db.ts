import "server-only";
import { logReadError, type ReadStage } from "./read-error";
import { Pool } from "pg";
import { marketSnapshotSQL } from "./market-query";
import { unstable_cache } from "next/cache";
import {
  assetSchema,
  analysisSchema,
  type Snapshot,
  type Point,
} from "./contracts";
class LoggedReadError extends Error {}
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
const cachedSnapshot = unstable_cache(
  async (): Promise<Snapshot> => {
    const asOf = new Date().toISOString();
    let stage: ReadStage = "connect";
    try {
      const db = database();
      const query = async (queryStage: ReadStage, sql: string) => {
        try {
          return await db.query(sql);
        } catch (error) {
          logReadError("market_read_failed", queryStage, error);
          throw new LoggedReadError();
        }
      };
      const [assets, runs] = await Promise.all([
        query("assets_query", marketSnapshotSQL),
        query(
          "runs_query",
          `SELECT status,finished_at,received,expected FROM ingestion_runs ORDER BY id DESC LIMIT 1`,
        ),
      ]);
      stage = "asset_validation";
      const parsed = assets.rows.map((r) =>
        assetSchema.parse({
          ...r,
          updated: r.updated?.toISOString() ?? null,
          analysis: r.analysis ? analysisSchema.parse(r.analysis) : null,
        }),
      );
      stage = "run_mapping";
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
    } catch (error) {
      if (!(error instanceof LoggedReadError))
        logReadError("market_read_failed", stage, error);
      // Throw only a sanitized marker: failed reads must never become cache entries.
      throw new LoggedReadError();
    }
  },
  ["market-snapshot-v3"],
  { revalidate: 300 },
);
export async function snapshot(): Promise<Snapshot> {
  try {
    return await cachedSnapshot();
  } catch {
    // The read logs its original error safely before throwing. The fallback is
    // deliberately outside unstable_cache so the next request can retry.
    return {
      assets: [],
      status: "unavailable",
      asOf: new Date().toISOString(),
      run: null,
    };
  }
}
export async function detail(id: string) {
  const data = await snapshot();
  const asset = data.assets.find((a) => a.id === id);
  if (!asset) return { data, asset: null, series: [] };
  let stage: ReadStage = "history_query";
  try {
    const h = await database().query(
      `SELECT day::text time,price::float8 price FROM (SELECT day,price FROM daily_samples WHERE asset_id=$1 ORDER BY day DESC LIMIT 2000)s ORDER BY day`,
      [id],
    );
    const detailedAsset = { ...asset, history: h.rows as Point[] };
    stage = "analytics_query";
    const r = await database().query(
      `SELECT day::text time,result FROM (SELECT day,result FROM analytics WHERE asset_id=$1 ORDER BY day DESC LIMIT 2000)s ORDER BY day`,
      [id],
    );
    stage = "detail_validation";
    return {
      data,
      asset: detailedAsset,
      series: r.rows.map((r) => ({
        time: r.time,
        analysis: analysisSchema.parse(r.result),
      })),
    };
  } catch (error) {
    logReadError("market_detail_read_failed", stage, error);
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
