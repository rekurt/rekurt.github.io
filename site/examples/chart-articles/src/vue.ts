import { createApp, defineComponent, h, shallowRef } from "vue";
import { OHLCVChart } from "@rekurt/openkline-vue";
import type {
  Candle,
  ChartType,
  IndicatorConfig,
  OHLCVChart as CoreChart,
} from "@rekurt/openkline-core";
import { candleApp } from "./candle-app";
candleApp("Vue", (container, initial) => {
  const data = shallowRef<Candle[]>(initial),
    type = shallowRef<ChartType>("candles"),
    indicators = shallowRef<IndicatorConfig[]>([{ type: "sma", period: 20 }]);
  let exposed: {
    chart: CoreChart | null;
    updateLastCandle: (c: Candle) => void;
  } | null = null;
  const app = createApp(
    defineComponent({
      setup: () => () =>
        h(OHLCVChart, {
          ref: (value: unknown) => {
            exposed = value as typeof exposed;
          },
          symbol: "BTC/USDT",
          resolution: "1m",
          theme: "dark",
          data: data.value,
          chartType: type.value,
          indicators: indicators.value,
        }),
    }),
  );
  app.mount(container);
  return {
    replace: (d) => {
      data.value = d;
    },
    update: (c) => exposed?.updateLastCandle(c),
    type: (t) => {
      type.value = t;
    },
    sma: (enabled) => {
      indicators.value = enabled ? [{ type: "sma", period: 20 }] : [];
    },
    chart: () => exposed?.chart ?? null,
    destroy: () => app.unmount(),
  };
});
