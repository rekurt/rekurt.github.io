## Try the liquidity experiment

The embedded lab compares an unchanged baseline with a second chart. Select the named scenarios or use **Next step**:

1. **Balanced book:** four bid levels and four ask levels, with 18 units on each side. The best bid is 99, the best ask is 101, and the spread is 2.
2. **Large ask appears:** the volume at 102 rises from 6 to 30. Total ask depth becomes 42 while bid depth remains 18. The displayed imbalance becomes −0.4; the spread is still 2.
3. **Near-touch liquidity disappears:** the nearest remaining bid is 97 and ask is 103, with one additional outer level on each side. Each side holds 8 units, and the spread widens to 6. Balanced totals do not imply a narrow spread.

The table below the charts shows cumulative volume from the best price outward. Bid levels accumulate toward lower prices; ask levels accumulate toward higher prices. The large ask changes the curve at its price and beyond it, rather than changing the best ask simply because it is large.

## Three practical uses

**An onboarding simulator.** Let a new user change one variable at a time: remove the nearest levels, then restore them. The baseline makes the change visible without asking the reader to remember the previous curve.

**A market-data integration check.** Put a numerical ladder beside the chart. When an incoming snapshot unexpectedly changes the curve, compare the best prices and cumulative totals before blaming the renderer. The article demo feeds complete snapshots through `setData`; it does not treat every WebSocket event as an incremental delta.

**An incident explanation.** Use a saved, appropriately sourced book sequence to explain what the interface displayed at a particular moment. This lab supplies synthetic states for that purpose; it is not a historical recording of a real liquidity event.

A displayed order is not a promise of execution. This chart alone cannot identify why an order disappeared, prove manipulation, account for hidden liquidity or predict the result of an order you might submit.

## What the code does

The package provides `DepthChart`, `setData`, `getMetrics` and `destroy`. The actual demo entry point is included below, followed by its deterministic data model. These are the files compiled for this page, not a separately maintained approximation of the example.

The baseline is intentionally fixed. In synthetic mode the second chart reports the change in imbalance relative to it. Live mode uses a different price scale and explicitly stops making that comparison.

## Optional live BTC/USDT view

**Connect live BTC/USDT** opens Binance's public market-data WebSocket only after a click. Each `btcusdt@depth20` message replaces the current visible book with up to 20 bid and 20 ask levels. It is a limited-depth snapshot, not the full exchange order book and not the diff-depth protocol.

The status shows the source and local receipt time. Receipt time is not a measurement of exchange latency. Old or duplicate update identifiers do not replace a newer book. A disconnection marks the data stale and starts a bounded reconnect sequence; returning to synthetic mode closes the connection.

Availability depends on the network and the provider. If the connection is unavailable, the deterministic lesson still works. The application never silently relabels synthetic data as live.

[Binance market-data endpoints](https://github.com/binance/binance-spot-api-docs/blob/master/faqs/market_data_only.md) · [Partial book stream specification](https://github.com/binance/binance-spot-api-docs/blob/master/web-socket-streams.md#partial-book-depth-streams)

## Reproduce and inspect

The repository linked below includes the four-entry chart-articles workspace, its lockfile and browser checks. With the repository's supported Node version, run:

```sh
git clone https://github.com/rekurt/rekurt.github.io.git
cd rekurt.github.io/site
npm ci
npm run build
npm run dev
```

Open `/articles/depth-liquidity-replay/` on the local Astro server. The demo consumes the published npm package; a local source checkout of depth is not substituted for it.

The checks exercise fixed cumulative totals, scenario transitions, reset, malformed input and a mocked live snapshot. Desktop and mobile browser checks inspect the rendered chart and controls. Those checks establish the behavior of this example, not production exchange connectivity, execution quality or an independent security review of the library.

[Depth source and documentation](https://github.com/rekurt/depth) · [Published package](https://www.npmjs.com/package/@rekurt/depth/v/0.1.0)

Disclosure: this article and its demo were developed with AI assistance. The example is a visualization experiment, with its validation boundaries stated above.
