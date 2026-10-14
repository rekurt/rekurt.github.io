import { createRef } from "react";
import { createRoot } from "react-dom/client";
import { flushSync } from "react-dom";
import { OHLCVChart, type OHLCVChartRef } from "@rekurt/openkline-react";
import type {
  Candle,
  ChartType,
  IndicatorConfig,
} from "@rekurt/openkline-core";
import { candleApp } from "./candle-app";
candleApp("React", (container, initial) => {
  const root = createRoot(container),
    ref = createRef<OHLCVChartRef>();
  let data: Candle[] = initial,
    type: ChartType = "candles",
    indicators: IndicatorConfig[] = [{ type: "sma", period: 20 }];
  const render = () =>
    flushSync(() =>
      root.render(
        <OHLCVChart
          ref={ref}
          symbol="BTC/USDT"
          resolution="1m"
          theme="dark"
          data={data}
          chartType={type}
          indicators={indicators}
          style={{ height: "100%" }}
        />,
      ),
    );
  render();
  return {
    replace: (d) => {
      data = d;
      render();
    },
    update: (c) => ref.current?.updateLastCandle(c),
    type: (t) => {
      type = t;
      render();
    },
    sma: (enabled) => {
      indicators = enabled ? [{ type: "sma", period: 20 }] : [];
      render();
    },
    chart: () => ref.current?.chart ?? null,
    destroy: () => flushSync(() => root.unmount()),
  };
});
