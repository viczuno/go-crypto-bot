import { Suspense } from "react";
import { snapshot } from "@/lib/db";
import { Comparison } from "@/components/market-ui";
export const dynamic = "force-dynamic";
export const metadata = { title: "Compare assets" };
export default async function Page() {
  return (
    <Suspense>
      <Comparison data={await snapshot()} />
    </Suspense>
  );
}
