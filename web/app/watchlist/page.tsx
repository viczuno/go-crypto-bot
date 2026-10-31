import { Suspense } from "react";
import { snapshot } from "@/lib/db";
import { Explorer } from "@/components/market-ui";
export const dynamic = "force-dynamic";
export const metadata = { title: "Watchlist" };
export default async function Page() {
  return (
    <Suspense>
      <Explorer data={await snapshot()} watchlist />
    </Suspense>
  );
}
