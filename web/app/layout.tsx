import type { Metadata } from "next";
import { Providers, Shell } from "@/components/shell";
import "./globals.css";
const canonicalHost = process.env.VERCEL_PROJECT_PRODUCTION_URL;

export const metadata: Metadata = {
  ...(canonicalHost
    ? { metadataBase: new URL(`https://${canonicalHost}`) }
    : {}),
  applicationName: "Crypto Market Observatory",
  title: { default: "CRYPTO | Market Observatory", template: "%s | CRYPTO" },
  description:
    "Research cryptocurrency markets with sampled price history, Go-powered technical analysis, watchlists, and asset comparisons.",
  openGraph: {
    type: "website",
    siteName: "CRYPTO | Market Observatory",
    title: "CRYPTO | Market Observatory",
    description:
      "Independent cryptocurrency market research with sampled prices and Go-powered technical analysis.",
  },
  twitter: {
    card: "summary_large_image",
    title: "CRYPTO | Market Observatory",
    description:
      "Independent cryptocurrency market research with sampled prices and Go-powered technical analysis.",
  },
};
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <Providers>
          <Shell>{children}</Shell>
        </Providers>
      </body>
    </html>
  );
}
