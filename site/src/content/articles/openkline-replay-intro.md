# OpenKline: replay candles in vanilla JavaScript, React and Vue

A useful chart replay reveals information in the order it becomes available. It does not draw tomorrow's candle and merely move a cursor over yesterday. It also needs to distinguish a change to the current candle from the arrival of a new one.

This example uses **@rekurt/openkline-core 0.2.0**, with **@rekurt/openkline-react 0.1.0** and **@rekurt/openkline-vue 0.1.0**. The three interfaces share one deterministic dataset and one replay controller. The framework components are integrations of the same chart core, so they belong in one walkthrough.

Core v0.2.0 was released on October 9, 2026. The core release date is separate from this article’s publication date.

