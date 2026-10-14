# Depth: what happens when liquidity disappears

A market-depth curve looks stable until the orders near the best bid and ask disappear. The shape changes, the spread can widen, and a previously reassuring wall of liquidity may no longer be there. A small interactive experiment is a better way to explain that behavior than a static green-and-red screenshot.

This lab uses **@rekurt/depth 0.1.0**, a Canvas depth-chart package. It starts with invented prices around 100 and fixed volumes, so every reader can reproduce the same three states. Nothing in the synthetic replay is presented as an exchange feed.

The GitHub v0.1.0 release dates to June 12, 2026; this is a new walkthrough of that existing version.

