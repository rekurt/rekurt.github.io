# Chart article examples

Four independent Vite entries using official npm packages: depth 0.1.0, OpenKline core 0.2.0, React/Vue wrappers 0.1.0. React 19.3.1 and Vue 3.5.44 are pinned; the shared lockfile is ../../package-lock.json. No vendored/source replacement for any chart package.

From the `site` directory (Node >=24.20 <25):

```sh
npm ci
npm run check
npm test
npm run build
npm run dev
```

Astro serves the already built applications under `/article-demos/depth/` and `/article-demos/openkline/{core,react,vue}/`. `npm run build` always builds the Vite workspace before Astro. The generated `site/public/article-demos` output is ignored; do not edit it manually.

`model.ts` owns the reproducible scenarios; `live.ts` owns explicitly started, cancelable public feeds. Core/React/Vue use the same controller. Framework entry points demonstrate the imperative ref escape hatch for streaming updates and declarative props for full snapshots/options. Frameworks are not loaded by the vanilla entry point. Standalone demos are noindex; the English articles own canonical URLs and sitemap entries.

The two REST/WebSocket hosts are Binance's public market-data-only endpoints. No key, account or trade call is used. Live mode is opt-in; automated tests substitute controlled responses. It retains at most 300 candles and displays local receipt time, not exchange latency. A feed object makes at most six automatic reconnect attempts (1/2/4/8/16/30 seconds); a user can explicitly start a new one. Each reconnection reloads initial candle history. Synthetic data is always labeled synthetic.

Unit tests cover fixed book totals, prefix replay, timestamp conversion, merging, malformed input, bounded reconnects, stale status and cancellation. Browser coverage in `site/tests/e2e/chart-articles.spec.ts` includes desktop/mobile rendering, unchanged historical view during updates, indicators, repeated framework remounts, snapshot replacement and tab navigation. Real-provider checks are separate evidence and not a CI dependency.

Developed with AI assistance. These examples do not validate execution, predict prices, or establish production market-data completeness.
