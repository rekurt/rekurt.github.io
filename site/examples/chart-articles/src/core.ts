import { OHLCVChart } from "@rekurt/openkline-core";
import { candleApp } from "./candle-app";
candleApp("vanilla", (container, data) => {
  const chart = new OHLCVChart({
    container,
    symbol: "BTC/USDT",
    resolution: "1m",
    theme: "dark",
  });
  chart.setData(data);
  chart.setIndicatorConfigs([{ type: "sma", period: 20 }]);
  return {
    replace: (d) => chart.setData(d, { preserveView: true }),
    update: (c) => chart.updateLastCandle(c),
    type: (t) => chart.setChartType(t),
    sma: (enabled) =>
      chart.setIndicatorConfigs(enabled ? [{ type: "sma", period: 20 }] : []),
    chart: () => chart,
    destroy: () => chart.destroy(),
  };
});
