import "./embed";
import { DepthChart, type DepthSnapshot } from "@rekurt/depth";
import { books, cumulative, parseBook } from "./model";
import { LiveFeed } from "./live";
import "./style.css";
const app = document.querySelector<HTMLElement>("#app")!;
app.innerHTML = `<h1>Depth · liquidity replay</h1><p>Compare a fixed baseline with changing visible liquidity. Synthetic units, not an execution forecast.</p><div class="controls"><button id="next">Next step</button><button id="reset">Reset</button>${books.map((b, i) => `<button data-scene="${i}">${b.name}</button>`).join("")}<button id="live">Connect live BTC/USDT</button><button id="synthetic">Synthetic mode</button></div><p id="status" role="status">SYNTHETIC · deterministic snapshots</p><div class="comparison"><section class="panel"><h2>Baseline · balanced</h2><div id="baseline" class="chart"></div><p id="baseline-metrics" class="metrics"></p></section><section class="panel"><h2 id="scene">Balanced book</h2><div id="current" class="chart"></div><p id="metrics" class="metrics"></p></section></div><p><small>Live: Binance BTC/USDT, top 20 levels per side, partial snapshots. Local receipt time is not exchange latency. This is not a full order book.</small></p><div class="table-wrap"><table><caption>Current cumulative levels (base-asset volume)</caption><thead><tr><th>Side</th><th>Price</th><th>Level volume</th><th>Cumulative</th></tr></thead><tbody id="levels"></tbody></table></div>`;
const baseline = new DepthChart({
  container: document.querySelector("#baseline")!,
  theme: "dark",
  data: books[0]!.book,
  shape: "step",
});
const current = new DepthChart({
  container: document.querySelector("#current")!,
  theme: "dark",
  data: books[0]!.book,
  shape: "step",
});
const base = baseline.getMetrics();
const metrics = (m: ReturnType<DepthChart["getMetrics"]>) =>
  `Spread: ${m.spread?.toFixed(2)} · Midpoint: ${m.midPrice?.toFixed(2)}\nBids: ${m.bidTotal.toFixed(2)} · Asks: ${m.askTotal.toFixed(2)} · Imbalance: ${m.imbalance.toFixed(3)}`;
document.querySelector("#baseline-metrics")!.textContent = metrics(base);
let index = 0;
let feed: LiveFeed | undefined;
let sequence = -1;
let live = false;
const status = (text: string) => {
  document.querySelector("#status")!.textContent = text;
};
function show(book: DepthSnapshot, name: string) {
  current.setData(book);
  const m = current.getMetrics();
  document.querySelector("#scene")!.textContent = name;
  document.querySelector("#metrics")!.textContent =
    metrics(m) +
    `\nΔ imbalance vs synthetic baseline: ${(m.imbalance - base.imbalance).toFixed(3)}`;
  const rows = cumulative(book);
  document.querySelector("#levels")!.innerHTML = Object.entries(rows)
    .flatMap(([side, levels]) =>
      levels.map(
        (l) =>
          `<tr><td>${side}</td><td>${l.price.toFixed(2)}</td><td>${l.volume.toFixed(3)}</td><td>${l.cumulative.toFixed(3)}</td></tr>`,
      ),
    )
    .join("");
}
function synthetic(i: number) {
  feed?.stop();
  feed = undefined;
  live = false;
  index = i;
  sequence = -1;
  status("SYNTHETIC · deterministic snapshots");
  show(books[index]!.book, books[index]!.name);
}
app
  .querySelectorAll<HTMLButtonElement>("[data-scene]")
  .forEach((b) => (b.onclick = () => synthetic(Number(b.dataset.scene))));
app.querySelector<HTMLButtonElement>("#next")!.onclick = () =>
  synthetic((index + 1) % books.length);
app.querySelector<HTMLButtonElement>("#reset")!.onclick = () => synthetic(0);
app.querySelector<HTMLButtonElement>("#synthetic")!.onclick = () =>
  synthetic(0);
app.querySelector<HTMLButtonElement>("#live")!.onclick = () => {
  feed?.stop();
  live = true;
  sequence = -1;
  show({ bids: [], asks: [] }, "Waiting for live BTC/USDT");
  document.querySelector("#metrics")!.textContent =
    "No live snapshot received yet";
  feed = new LiveFeed({
    stream: "btcusdt@depth20",
    status,
    receive: (raw) => {
      const book = parseBook(raw);
      const next = Number(book.sequence);
      if (next <= sequence) return;
      sequence = next;
      show(book, "LIVE · BTC/USDT top 20");
      document.querySelector("#metrics")!.textContent =
        metrics(current.getMetrics()) +
        "\nLive market · not comparable to the synthetic price scale";
    },
  });
  feed.start();
};
function cleanup() {
  feed?.stop();
  baseline.destroy();
  current.destroy();
}
window.addEventListener("pagehide", cleanup, { once: true });
window.addEventListener("pageshow", (e) => {
  if (e.persisted) location.reload();
});
synthetic(0);
