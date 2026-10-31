import { z } from "zod";
import { idSchema } from "@/lib/contracts";
import { detail } from "@/lib/db";
export async function GET(req: Request) {
  const raw = new URL(req.url).searchParams.get("ids") ?? "";
  const p = z
    .array(idSchema)
    .max(5)
    .safeParse(raw ? raw.split(",") : []);
  if (!p.success)
    return Response.json(
      { error: { code: "INVALID_ASSETS" } },
      { status: 400 },
    );
  const results = await Promise.all([...new Set(p.data)].map(detail));
  if (results.some((r) => r.data.status === "unavailable"))
    return Response.json({ error: { code: "UNAVAILABLE" } }, { status: 503 });
  return Response.json(
    { assets: results.flatMap((r) => (r.asset ? [r.asset] : [])) },
    { headers: { "Cache-Control": "public, s-maxage=300" } },
  );
}
