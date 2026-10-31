import type { Metadata } from "next";
import { Analytics } from "@vercel/analytics/next";
import { Providers, Shell } from "@/components/shell";
import "./globals.css";
export const metadata: Metadata = {
  title: { default: "Crypto · Market Observatory", template: "%s · Crypto" },
  description:
    "Cryptocurrency market research with sampled price history, Go-powered technical analysis, watchlists and comparisons.",
};
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <Providers>
          <Shell>{children}</Shell>
        </Providers>
        <Analytics />
      </body>
    </html>
  );
}
