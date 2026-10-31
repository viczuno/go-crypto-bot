import { snapshot } from "@/lib/db";
import { Dashboard } from "@/components/market-ui";
export const dynamic = "force-dynamic";
export default async function Page() {
  return <Dashboard data={await snapshot()} />;
}
