import { detail } from "@/lib/db";
import { idSchema } from "@/lib/contracts";
export async function GET(
  _req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const { id } = await params;
  if (!idSchema.safeParse(id).success)
    return Response.json({ error: { code: "INVALID_ID" } }, { status: 400 });
  const d = await detail(id);
  if (d.data.status === "unavailable")
    return Response.json({ error: { code: "UNAVAILABLE" } }, { status: 503 });
  if (!d.asset)
    return Response.json({ error: { code: "NOT_FOUND" } }, { status: 404 });
  return Response.json(
    { history: d.asset.history, indicators: d.series },
    { headers: { "Cache-Control": "public, s-maxage=300" } },
  );
}
