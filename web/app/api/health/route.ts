import { health } from "@/lib/db";
export async function GET() {
  const ok = await health();
  return Response.json(
    { status: ok ? "ok" : "unavailable" },
    { status: ok ? 200 : 503, headers: { "Cache-Control": "no-store" } },
  );
}
