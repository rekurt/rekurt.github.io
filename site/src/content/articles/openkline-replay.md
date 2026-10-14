## Replay a forming candle

The initial view contains 40 synthetic one-minute candles. Every pair of steps then reveals one more candle:

- The first step adds a partially formed candle with a new timestamp.
- The second step updates that same timestamp with its completed values.
- The next step begins the following minute.

Use **Play**, **Next step**, **Speed**, and **Reset** to inspect that sequence. The readout reports the visible count and latest revealed time. SMA(20) receives only the revealed candles; future candles are not passed into the chart or its indicator configuration.

These values are deliberately synthetic and reproducible. The demonstration is not a backtest, a trading signal or evidence of investment performance.

## Inspect history while updates continue

Select **Inspect history**, then advance the replay. The chart should remain on the earlier view rather than pulling you back to the right edge. **Go to latest** explicitly restores following.

This is useful in an incident-review interface: a person may be examining an earlier sequence while fresh information arrives. Automatic updates should not silently erase that navigation choice. The example also offers a **Candles/Line** selector, an SMA toggle and a **Remount chart** button for checking component lifecycle behavior.

## Three useful applications

**A teaching interface.** Pause on a forming candle, discuss the information visible at that point, and complete it with the next step. The same timestamp is updated instead of creating an extra bar for every tick.

**A replay-based UI regression.** Feed exactly the same sequence to vanilla, React and Vue. Compare counts, latest values and navigation behavior. This isolates integration differences from changes in a live feed; a successful example does not establish universal feature parity between packages.

**An operational timeline.** Adapt the input model to a genuine OHLCV source, keep history navigation stable, and distinguish source time from local receipt time. Non-market time series need a meaningful OHLCV interpretation; a candle renderer is not automatically the right representation for every metric.

## Core and wrappers share the data contract

A candle has `o`, `h`, `l`, `c`, `v` and `t` fields. This example uses UTC Unix seconds for `t`. Binance supplies milliseconds, so the live adapter divides its timestamp by 1,000 before constructing a candle.

The vanilla adapter calls `setData` for a full snapshot and `updateLastCandle` for a forming or newly arriving candle. Full replacements preserve the view. React uses the wrapper component and its ref; Vue uses the wrapper component and its exposed methods. Their framework lifecycles dispose of the chart on unmount. The shared controller also stops playback and live connections when it is torn down.

The actual, type-checked entry points for all three frameworks appear below. They use React 19 and Vue 3.5 versions pinned in the example lockfile, with the published core and wrapper packages. A vendored core archive is not involved.

## Optional live mode and its limits

Live mode first loads recent BTC/USDT one-minute candles from the public REST endpoint, buffering WebSocket messages while history loads. It then replaces the last candle when timestamps match, appends a newer timestamp, and ignores older updates. At most 300 bars are retained.

When the stream disconnects, the interface marks the data stale. It reconnects after 1, 2, 4, 8, 16 and 30 seconds, then stops automatic attempts. A successful reconnection reloads the history snapshot before applying queued events. Returning to synthetic replay stops the feed and restores the deterministic starting state.

This is a bounded browser demo, not an archival feed or a complete market-data platform. It does not promise recovery of every tick. The public stream may be unavailable on a particular network; the synthetic replay remains usable without it.

[Official public market-data endpoints](https://github.com/binance/binance-spot-api-docs/blob/master/faqs/market_data_only.md) · [Kline stream specification](https://github.com/binance/binance-spot-api-docs/blob/master/web-socket-streams.md#klinecandlestick-streams-for-utc)

## Run the same examples locally

```sh
git clone https://github.com/rekurt/rekurt.github.io.git
cd rekurt.github.io/site
npm ci
npm run build
npm run dev
```

Open `/articles/openkline-replay-react-vue/`. Each framework has its own built entry point, so choosing the vanilla view does not load React or Vue.

Checks cover the shared sequence, partial/final candle transitions, time conversion, input validation, bounded reconnection and cancellation. Browser scenarios cover all three integrations, history navigation, indicator changes, repeated remounting and mocked live input on desktop and mobile. Real provider availability is checked separately from deterministic tests.

[Core](https://github.com/rekurt/openkline) · [React wrapper](https://github.com/rekurt/openkline-react) · [Vue wrapper](https://github.com/rekurt/openkline-vue)

Disclosure: this article and its demo were developed with AI assistance. The checks validate these examples and their stated scenarios, not every library feature or a trading strategy.
