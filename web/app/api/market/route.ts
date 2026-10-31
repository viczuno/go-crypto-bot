import { snapshot } from "@/lib/db";
import { querySchema } from "@/lib/contracts";
import { filterAssets } from "@/lib/market";
export async function GET(req: Request) {
  const p = querySchema.safeParse(
    Object.fromEntries(new URL(req.url).searchParams),
  );
  if (!p.success)
    return Response.json(
      {
        error: { code: "INVALID_QUERY", message: "Invalid market parameters" },
      },
      { status: 400 },
    );
  const data = await snapshot();
  if (data.status === "unavailable")
    return Response.json(
      {
        error: {
          code: "UNAVAILABLE",
          message: "Market data temporarily unavailable",
        },
      },
      { status: 503 },
    );
  const { q, sort, direction, page, limit } = p.data;
  const assets = filterAssets(data.assets, q, sort, direction);
  return Response.json(
    {
      ...data,
      assets: assets.slice((page - 1) * limit, page * limit),
      total: assets.length,
      page,
      limit,
    },
    {
      headers: {
        "Cache-Control": "public, s-maxage=300, stale-while-revalidate=60",
      },
    },
  );
}
