## Follow the orders

Start with **Balanced book**. There are four levels on each side: 18 units of bids and 18 of asks. The best bid is 99, the best ask is 101.

Select **Large ask appears**. The sell order at 102 grows from 6 to 30 units. This leaves the spread at 2: the best ask is still 101. The extra volume appears farther along the ask curve.

Now select **Near-touch liquidity disappears**. The orders at 99, 98, 101 and 102 are gone. The best remaining prices are 97 and 103. Each side has 8 units, but the gap between them is three times wider.

| State | Best bid | Best ask | Spread | Bid volume | Ask volume |
| --- | ---: | ---: | ---: | ---: | ---: |
| Balanced | 99 | 101 | 2 | 18 | 18 |
| Large ask | 99 | 101 | 2 | 18 | 42 |
| Nearest levels removed | 97 | 103 | 6 | 8 | 8 |

The curves show cumulative volume. Bids accumulate from the highest price down; asks accumulate from the lowest price up. The table under the demo exposes those running totals, which is useful when checking a data adapter: compare the numbers at each price with the shape being drawn.

The displayed imbalance is `(bid volume − ask volume) / total volume`. In the second state it is `(18 − 42) / 60 = −0.4`. That single number loses the location of the orders. Keep the price levels visible if the question is where the liquidity sits.

## Feeding the chart

Create a `DepthChart` with a container, then pass complete books to `setData`. Read the spread, midpoint and totals through `getMetrics`. Call `destroy` when removing the chart.

The code at the end of this page is the code used by the demo. The sample data lives alongside the entry point, so you can change one level and rerun the example. For an onboarding screen, keep the fixed chart as a reference. For a feed-debugging screen, replace that reference with the previous snapshot to make a jump easier to inspect.

## Switch to BTC/USDT

**Connect live BTC/USDT** opens Binance's `btcusdt@depth20` stream. Each message replaces the visible book with the top 20 levels per side. This is a partial-book snapshot stream; an incremental diff feed needs different handling.

The live chart shows the source and local receipt time. It ignores older update IDs, marks disconnected data stale and attempts to reconnect. **Synthetic mode** closes the stream and restores the sample book. The fixed chart remains synthetic, and its price scale is unrelated to BTC/USDT.

Visible orders can be cancelled. Their disappearance does not tell you why they vanished, and displayed volume does not guarantee an execution price. This example only displays the supplied book.

[Stream format](https://github.com/binance/binance-spot-api-docs/blob/master/web-socket-streams.md#partial-book-depth-streams) · [Public market-data endpoints](https://github.com/binance/binance-spot-api-docs/blob/master/faqs/market_data_only.md)

## Run it locally

Use Node 24.20 or newer within the 24.x series:

```sh
git clone https://github.com/rekurt/rekurt.github.io.git
cd rekurt.github.io/site
npm ci
npm run build
npm run dev
```

Open `/articles/depth-liquidity-replay/`. The example uses the published depth package pinned in the lockfile. Tests cover the fixed totals, state changes, snapshot replacement and malformed input; live connectivity was checked separately in a browser.

Version 0.1.0 was released on June 12, 2026.

[Source](https://github.com/rekurt/depth) · [npm package](https://www.npmjs.com/package/@rekurt/depth/v/0.1.0)

*Written with AI assistance.*
