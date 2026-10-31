import { notFound } from "next/navigation";
import { detail } from "@/lib/db";
import { idSchema } from "@/lib/contracts";
import { CoinDetail } from "@/components/coin-detail";
import { State } from "@/components/market-ui";
export const dynamic = "force-dynamic";
export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return { title: idSchema.safeParse(id).success ? id : "Asset" };
}
export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  if (!idSchema.safeParse(id).success) notFound();
  const d = await detail(id);
  if (d.data.status === "unavailable") return <State data={d.data} />;
  if (!d.asset) notFound();
  return <CoinDetail asset={d.asset} series={d.series} />;
}
