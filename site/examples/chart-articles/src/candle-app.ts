import "./embed";
import type { Candle, OHLCVChart, ChartType } from "@rekurt/openkline-core";
import {
  replayAt,
  maxStep,
  parseCandle,
  parseKline,
  mergeCandle,
} from "./model";
import { LiveFeed } from "./live";
import "./style.css";
export interface ChartAdapter {
  replace: (data: Candle[]) => void;
  update: (c: Candle) => void;
  type: (t: ChartType) => void;
  sma: (enabled: boolean) => void;
  chart: () => OHLCVChart | null;
  destroy: () => void;
}
export type Mount = (container: HTMLElement, data: Candle[]) => ChartAdapter;
export function candleApp(name: string, mount: Mount) {
  const app = document.querySelector<HTMLElement>("#app")!;
  app.innerHTML = `<h1>OpenKline · ${name} replay</h1><p>Reveal only the candles available at this step. Pan into history while new bars arrive, then return to the latest data.</p><div class="controls"><button id="play">Play</button><button id="step">Next step</button><button id="reset">Reset</button><label>Speed <select id="speed" aria-label="Replay speed"><option value="1">1×</option><option value="4">4×</option></select></label><label>View <select id="type" aria-label="Chart view"><option value="candles">Candles</option><option value="line">Line</option></select></label><label><input id="sma" type="checkbox" checked> SMA(20)</label><button id="history">Inspect history</button><button id="latest">Go to latest</button><button id="live">Connect live BTC/USDT</button><button id="synthetic">Synthetic replay</button><button id="remount">Remount chart</button></div><p id="status" role="status">SYNTHETIC · replay paused</p><div class="panel"><div id="chart" class="chart"></div></div><p id="readout" class="metrics" aria-live="polite"></p><p><small>Synthetic by default. Optional live mode: Binance BTC/USDT 1m candles, at most 300 retained bars. Updates describe visualization, not trading results.</small></p>`;
  let step = 0,
    data = replayAt(0),
    playing = false,
    speed = 1,
    isLive = false,
    timer: ReturnType<typeof setTimeout> | undefined,
    feed: LiveFeed | undefined;
  let adapter = mount(document.querySelector("#chart")!, data);
  const status = (s: string) => {
    document.querySelector("#status")!.textContent = s;
  };
  const readout = () => {
    const last = data.at(-1)!;
    const state = adapter.chart()?.saveLayoutState();
    const el = document.querySelector<HTMLElement>("#readout")!;
    el.dataset.step = String(step);
    el.dataset.count = String(data.length);
    el.dataset.actualCount = String(adapter.chart()?.getBuffer().length ?? 0);
    el.dataset.time = String(last.t);
    el.dataset.close = String(last.c);
    el.dataset.mode = isLive ? "live" : "synthetic";
    el.dataset.chartType = String(state?.chartType);
    el.dataset.indicators = String(state?.indicators.length);
    el.dataset.follow = String(state?.viewport.autoFollow ?? true);
    el.textContent = `${isLive ? "Live" : "Replay step " + step + "/" + maxStep} · visible candles: ${data.length} · latest UTC: ${new Date(last.t * 1000).toISOString()}\nLast close: ${last.c.toFixed(3)} · following latest: ${state?.viewport.autoFollow ?? true}`;
  };
  const pause = () => {
    playing = false;
    clearTimeout(timer);
    document.querySelector("#play")!.textContent = "Play";
  };
  const advance = () => {
    if (isLive) return;
    if (step >= maxStep) {
      pause();
      return;
    }
    data = replayAt(++step);
    adapter.update(data.at(-1)!);
    status(
      "SYNTHETIC · " +
        (step % 2 ? "new candle forming" : "current candle completed"),
    );
    requestAnimationFrame(readout);
  };
  const tick = () => {
    if (!playing) return;
    advance();
    if (playing) timer = setTimeout(tick, 1000 / speed);
  };
  function synthetic() {
    pause();
    feed?.stop();
    feed = undefined;
    isLive = false;
    step = 0;
    data = replayAt(0);
    adapter.replace(data);
    adapter.chart()?.goToLive();
    status("SYNTHETIC · replay paused");
    requestAnimationFrame(readout);
  }
  app.querySelector<HTMLButtonElement>("#play")!.onclick = () => {
    if (isLive) return;
    if (playing) {
      pause();
      return;
    }
    playing = true;
    document.querySelector("#play")!.textContent = "Pause";
    tick();
  };
  app.querySelector<HTMLButtonElement>("#step")!.onclick = () => {
    pause();
    advance();
  };
  app.querySelector<HTMLButtonElement>("#reset")!.onclick = synthetic;
  app.querySelector<HTMLButtonElement>("#synthetic")!.onclick = synthetic;
  app.querySelector<HTMLSelectElement>("#speed")!.onchange = (e) => {
    speed = Number((e.target as HTMLSelectElement).value);
  };
  app.querySelector<HTMLSelectElement>("#type")!.onchange = (e) => {
    adapter.type((e.target as HTMLSelectElement).value as ChartType);
    requestAnimationFrame(readout);
  };
  app.querySelector<HTMLInputElement>("#sma")!.onchange = (e) => {
    adapter.sma((e.target as HTMLInputElement).checked);
    requestAnimationFrame(readout);
  };
  app.querySelector<HTMLButtonElement>("#history")!.onclick = () => {
    const chart = adapter.chart();
    if (chart) {
      const state = chart.saveLayoutState();
      state.viewport = {
        ...state.viewport,
        startIndex: 0,
        candleWidth: 18,
        autoFollow: false,
      };
      chart.loadState(state);
    }
    readout();
  };
  app.querySelector<HTMLButtonElement>("#latest")!.onclick = () => {
    adapter.chart()?.goToLive();
    readout();
  };
  app.querySelector<HTMLButtonElement>("#remount")!.onclick = () => {
    pause();
    feed?.stop();
    feed = undefined;
    isLive = false;
    step = 0;
    data = replayAt(0);
    adapter.destroy();
    adapter = mount(document.querySelector("#chart")!, data);
    app.querySelector<HTMLSelectElement>("#type")!.value = "candles";
    app.querySelector<HTMLInputElement>("#sma")!.checked = true;
    status("SYNTHETIC · chart remounted");
    requestAnimationFrame(readout);
  };
  app.querySelector<HTMLButtonElement>("#live")!.onclick = () => {
    pause();
    feed?.stop();
    isLive = true;
    adapter.replace([]);
    document.querySelector("#readout")!.textContent =
      "Waiting for live candles";
    let firstSnapshot = true;
    feed = new LiveFeed({
      stream: "btcusdt@kline_1m",
      status,
      initialize: async (signal) => {
        const r = await fetch(
          "https://data-api.binance.vision/api/v3/klines?symbol=BTCUSDT&interval=1m&limit=120",
          { signal },
        );
        if (!r.ok) throw Error("Market data unavailable");
        const rows: unknown = await r.json();
        if (!Array.isArray(rows) || !rows.length)
          throw Error("Invalid candle response");
        const next = rows.map(parseCandle);
        if (next.some((c, i) => i > 0 && c.t <= next[i - 1]!.t))
          throw Error("Unordered candles");
        signal.throwIfAborted();
        data = next;
        adapter.replace(data);
        if (firstSnapshot) adapter.chart()?.goToLive();
        firstSnapshot = false;
        await new Promise<void>((resolve) =>
          requestAnimationFrame(() => resolve()),
        );
        signal.throwIfAborted();
        readout();
      },
      receive: (raw) => {
        const candle = parseKline(raw);
        const next = mergeCandle(data, candle);
        if (next === data) return;
        data = next;
        if (data.length === 300) adapter.replace(data);
        else adapter.update(candle);
        requestAnimationFrame(readout);
      },
    });
    feed.start();
  };
  const cleanup = () => {
    pause();
    feed?.stop();
    adapter.destroy();
  };
  window.addEventListener("pagehide", cleanup, { once: true });
  window.addEventListener("pageshow", (e) => {
    if (e.persisted) location.reload();
  });
  requestAnimationFrame(readout);
}
