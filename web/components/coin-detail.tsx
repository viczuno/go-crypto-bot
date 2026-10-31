"use client";
import Link from "next/link";
import dynamic from "next/dynamic";
import { useState, useMemo } from "react";
import { ArrowLeft, GitCompareArrows } from "lucide-react";
import type { Asset, Analysis } from "@/lib/contracts";
import { money, percent, stale } from "@/lib/market";
import { Favorite } from "./shell";
import { Heading, PriceChart } from "./market-ui";
const Chart = dynamic(() => import("./chart"), { ssr: false });
const explanations: Record<string, string> = {
  RSI: "14-day Wilder relative strength. Above 70 indicates overbought conditions; below 30 indicates oversold conditions.",
  MACD: "12/26-day exponential averages with a 9-period signal line. The histogram describes momentum, not a guaranteed reversal.",
  MovingAverage:
    "50- and 200-day simple moving averages. A crossover requires both the current and previous complete windows.",
  BollingerBands:
    "20-day mean with bands at two population standard deviations. Width describes recent sampled-price variation.",
};
export function CoinDetail({
  asset,
  series,
}: {
  asset: Asset;
  series: { time: string; analysis: Analysis }[];
}) {
  const [overlay, setOverlay] = useState("none");
  const [panel, setPanel] = useState("RSI");
  const extra = useMemo(() => {
    const keys =
      overlay === "MovingAverage"
        ? ["sma_50", "sma_200"]
        : overlay === "BollingerBands"
          ? ["upper_band", "middle_band", "lower_band"]
          : [];
    return keys.map((key) => ({
      name: key.replaceAll("_", " ").toUpperCase(),
      points: series.flatMap((p) => {
        const v = p.analysis.indicators.find((i) => i.name === overlay)?.values[
          key
        ];
        return v === undefined ? [] : [{ time: p.time, price: v }];
      }),
    }));
  }, [overlay, series]);
  const indicators = useMemo(() => {
    const keys =
      panel === "RSI" ? ["value"] : ["macd_line", "signal_line", "histogram"];
    return keys.map((key) => ({
      name: key === "value" ? "RSI" : key.replaceAll("_", " "),
      points: series.flatMap((p) => {
        const i = p.analysis.indicators.find((i) => i.name === panel);
        const v = key === "value" ? i?.value : i?.values[key];
        return v == null ? [] : [{ time: p.time, price: v }];
      }),
    }));
  }, [panel, series]);
  return (
    <>
      <Link className="back-link" href="/market">
        <ArrowLeft size={15} />
        All markets
      </Link>
      <Heading
        eyebrow={`${asset.symbol.toUpperCase()} / USD`}
        title={asset.name}
        description="Price history, daily momentum, and the evidence behind the signal."
      >
        <div className="actions">
          <Favorite id={asset.id} />
          <Link
            className="button"
            href={`/compare?ids=${asset.id},${asset.id === "bitcoin" ? "ethereum" : "bitcoin"}`}
          >
            <GitCompareArrows size={16} />
            Compare
          </Link>
        </div>
      </Heading>
      <div className="metric-grid">
        <div className="metric">
          <span className="eyebrow">LATEST OBSERVED PRICE</span>
          <div className="metric-value">{money(asset.price)}</div>
          <span
            className={
              asset.change !== null && asset.change < 0
                ? "negative"
                : "positive"
            }
          >
            {percent(asset.change)} <small>24h provider change</small>
          </span>
        </div>
        <div className="metric">
          <span className="eyebrow">MARKET CAPITALIZATION</span>
          <div className="metric-value">{money(asset.cap, true)}</div>
          <p>As reported by CoinGecko</p>
        </div>
        <div className="metric">
          <span className="eyebrow">24H TRADING VOLUME</span>
          <div className="metric-value">{money(asset.volume, true)}</div>
          <p>Provider aggregate</p>
        </div>
        <div className="metric">
          <span className="eyebrow">OBSERVATION STATUS</span>
          <div className="metric-value small-value">
            {stale(asset.updated) ? "Stale" : "Fresh"}
          </div>
          <p>
            {asset.updated
              ? new Date(asset.updated)
                  .toISOString()
                  .replace("T", " ")
                  .replace("Z", " UTC")
              : "Unavailable"}
          </p>
        </div>
      </div>
      <section className="panel">
        <div className="panel-heading">
          <h2>Price history</h2>
          <select
            aria-label="Price overlay"
            value={overlay}
            onChange={(e) => setOverlay(e.target.value)}
          >
            <option value="none">No overlay</option>
            <option value="MovingAverage">Moving averages</option>
            <option value="BollingerBands">Bollinger bands</option>
          </select>
        </div>
        <PriceChart asset={asset} extra={extra} />
      </section>
      <div className="section-heading">
        <div>
          <span className="eyebrow">READ THE MOMENTUM</span>
          <h2>Technical analysis</h2>
        </div>
        <span className="tag">Completed UTC days</span>
      </div>
      <div className="analysis-summary">
        <strong>{asset.analysis?.signal ?? "Unavailable"}</strong>
        <span>{asset.analysis?.coverage ?? 0}/4 indicators available</span>
        <span>{asset.analysis?.samples ?? 0} consecutive daily samples</span>
        <span>As of {asset.analysis?.day ?? "—"}</span>
      </div>
      <div className="indicator-grid">
        {(asset.analysis?.indicators ?? []).map((i) => (
          <article className="panel indicator" key={i.name}>
            <div className="panel-heading">
              <h3>
                {i.name === "MovingAverage"
                  ? "Moving averages"
                  : i.name === "BollingerBands"
                    ? "Bollinger bands"
                    : i.name}
              </h3>
              <span className="signal">{i.signal}</span>
            </div>
            <div className="indicator-value">
              {i.value === null
                ? "—"
                : i.value.toLocaleString("en-US", { maximumFractionDigits: 3 })}
            </div>
            <p>{explanations[i.name]}</p>
            {i.reason ? (
              <div className="notice">
                {i.reason}. Requires {i.required} daily samples.
              </div>
            ) : (
              <div className="strength">
                <span>Heuristic strength</span>
                <strong>{Math.round(i.strength * 100)} / 100</strong>
              </div>
            )}
          </article>
        ))}
      </div>
      <section className="panel">
        <div className="panel-heading">
          <h2>Indicator history</h2>
          <div className="segmented">
            {["RSI", "MACD"].map((p) => (
              <button
                key={p}
                aria-pressed={panel === p}
                onClick={() => setPanel(p)}
              >
                {p}
              </button>
            ))}
          </div>
        </div>
        {indicators.some((i) => i.points.length > 1) ? (
          <Chart series={indicators} unit={panel === "RSI" ? "RSI" : "USD"} />
        ) : (
          <div className="chart-loading">
            Insufficient consecutive daily history for this indicator.
          </div>
        )}
      </section>
      <p className="disclosure">
        Indicators summarize sampled historical prices. Gaps reset calculation
        windows; unavailable indicators are excluded from consensus. Signal
        strength is a heuristic score and is not a statistically validated
        probability of profit. No candles are inferred from sampled prices.
      </p>
    </>
  );
}
