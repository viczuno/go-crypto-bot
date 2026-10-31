"use client";
import { useEffect, useRef } from "react";
import { useTheme } from "next-themes";
import {
  createChart,
  LineSeries,
  ColorType,
  type Time,
} from "lightweight-charts";
import type { Point } from "@/lib/contracts";
const palette = ["#31c9ac", "#8799ff", "#f1b85b", "#e879b0", "#55b6ed"];
export default function Chart({
  series,
  height = 300,
  percentage = false,
  unit = "USD",
}: {
  series: { name: string; points: Point[] }[];
  height?: number;
  percentage?: boolean;
  unit?: string;
}) {
  const host = useRef<HTMLDivElement>(null);
  const tooltip = useRef<HTMLDivElement>(null);
  const { resolvedTheme } = useTheme();
  useEffect(() => {
    if (!host.current) return;
    const dark = resolvedTheme !== "light";
    const chart = createChart(host.current, {
      height,
      autoSize: true,
      layout: {
        background: { type: ColorType.Solid, color: "transparent" },
        textColor: dark ? "#8996aa" : "#536279",
        fontFamily: "Arial",
        attributionLogo: true,
      },
      grid: {
        vertLines: { visible: false },
        horzLines: { color: dark ? "#233044" : "#e9edf2" },
      },
      rightPriceScale: { borderVisible: false },
      timeScale: { borderVisible: false, timeVisible: false },
      crosshair: {
        vertLine: { color: "#758496", labelBackgroundColor: "#26394b" },
        horzLine: { color: "#758496", labelBackgroundColor: "#26394b" },
      },
    });
    const observedTimes = [
      ...new Set(series.flatMap((s) => s.points.map((p) => p.time))),
    ].sort();
    const allTimes: string[] = [];
    if (observedTimes.length) {
      const start = Date.parse(observedTimes[0]),
        end = Date.parse(observedTimes.at(-1)!);
      for (let t = start; t <= end; t += 86400000)
        allTimes.push(new Date(t).toISOString().slice(0, 10));
    }
    const handles = series.map((s, i) => {
      const line = chart.addSeries(LineSeries, {
        color: palette[i % palette.length],
        lineWidth: 2,
        title: s.name,
        priceFormat: percentage
          ? { type: "percent", precision: 2, minMove: 0.01 }
          : { type: "price", precision: 6, minMove: 0.000001 },
      });
      const values = new Map(s.points.map((p) => [p.time, p.price]));
      line.setData(
        allTimes.map((t) =>
          values.has(t)
            ? { time: t as Time, value: values.get(t)! }
            : { time: t as Time },
        ),
      );
      return line;
    });
    chart.timeScale().fitContent();
    chart.subscribeCrosshairMove((p) => {
      if (!tooltip.current) return;
      if (!p.time) {
        tooltip.current.textContent =
          "Move across the chart to inspect daily samples";
        return;
      }
      const values = handles
        .map((h, i) => {
          const v = p.seriesData.get(h);
          return v && "value" in v
            ? `${series[i].name}: ${percentage || unit !== "USD" ? "" : "$"}${v.value.toLocaleString("en-US", { maximumFractionDigits: 6 })}${percentage ? "%" : ""}`
            : "";
        })
        .filter(Boolean);
      tooltip.current.textContent = `${typeof p.time === "string" ? p.time : JSON.stringify(p.time)} · ${values.join(" · ")}`;
    });
    const reset = () => chart.timeScale().fitContent();
    host.current.addEventListener("reset-chart", reset);
    const element = host.current;
    return () => {
      element.removeEventListener("reset-chart", reset);
      chart.remove();
    };
  }, [series, height, percentage, resolvedTheme, unit]);
  return (
    <div className="chart-frame">
      <button
        className="button"
        onClick={() => host.current?.dispatchEvent(new Event("reset-chart"))}
      >
        Reset zoom
      </button>
      <div
        ref={host}
        style={{ height }}
        aria-label="Interactive historical chart"
      />
      <div className="chart-tooltip" ref={tooltip}>
        Move across the chart to inspect daily samples
      </div>
      <p className="chart-credit">
        Charts by{" "}
        <a href="https://www.tradingview.com/" target="_blank" rel="noreferrer">
          TradingView
        </a>{" "}
        · Scroll to zoom, drag to pan
      </p>
    </div>
  );
}
