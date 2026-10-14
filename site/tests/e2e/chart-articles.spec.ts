import { test, expect } from "@playwright/test";

test("depth explains disappearing liquidity and cumulative volumes without network", async ({
  page,
}, info) => {
  const marketRequests: string[] = [];
  page.on("request", (r) => {
    if (r.url().includes("binance")) marketRequests.push(r.url());
  });
  await page.goto("/article-demos/depth/");
  await expect(page.locator("canvas").first()).toBeVisible();
  await expect(page.locator("#metrics")).toContainText("Spread: 2.00");
  await expect(page.locator("#metrics")).toContainText(
    "Bids: 18.00 · Asks: 18.00",
  );
  await page
    .getByRole("button", { name: "Large ask appears", exact: true })
    .click();
  await expect(page.locator("#metrics")).toContainText("Asks: 42.00");
  await expect(page.locator("#metrics")).toContainText("Imbalance: -0.400");
  await page
    .getByRole("button", {
      name: "Near-touch liquidity disappears",
      exact: true,
    })
    .click();
  await expect(page.locator("#metrics")).toContainText("Spread: 6.00");
  await page.screenshot({
    path: info.outputPath("depth-liquidity.png"),
    fullPage: true,
  });
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await expect(page.locator("#metrics")).toContainText("Spread: 2.00");
  expect(marketRequests).toEqual([]);
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});
for (const framework of ["core", "react", "vue"]) {
  test(`OpenKline ${framework}: partial/final replay, history, indicators and remount`, async ({
    page,
  }, info) => {
    const errors: string[] = [];
    page.on("pageerror", (e) => errors.push(e.message));
    await page.goto(`/article-demos/openkline/${framework}/`);
    const out = page.locator("#readout");
    await expect(page.locator("canvas").first()).toBeVisible();
    await expect(out).toHaveAttribute("data-actual-count", "40");
    await page.getByRole("button", { name: "Next step", exact: true }).click();
    await expect(out).toHaveAttribute("data-actual-count", "41");
    const partial = await out.getAttribute("data-close");
    const time = await out.getAttribute("data-time");
    await page.getByRole("button", { name: "Next step", exact: true }).click();
    await expect(out).toHaveAttribute("data-step", "2");
    await expect(out).toHaveAttribute("data-actual-count", "41");
    await expect(out).toHaveAttribute("data-time", time!);
    expect(await out.getAttribute("data-close")).not.toBe(partial);
    await page
      .getByRole("button", { name: "Inspect history", exact: true })
      .click();
    await expect(out).toHaveAttribute("data-follow", "false");
    await page.getByRole("button", { name: "Next step", exact: true }).click();
    await expect(out).toHaveAttribute("data-follow", "false");
    await expect(out).toHaveAttribute("data-actual-count", "42");
    await page.getByLabel("Chart view", { exact: true }).selectOption("line");
    await expect(out).toHaveAttribute("data-chart-type", "line");
    await page.getByLabel("SMA(20)", { exact: true }).uncheck();
    await expect(out).toHaveAttribute("data-indicators", "0");
    await page
      .getByRole("button", { name: "Go to latest", exact: true })
      .click();
    await expect(out).toHaveAttribute("data-follow", "true");
    await page.getByRole("button", { name: "Play", exact: true }).click();
    await expect(
      page.getByRole("button", { name: "Pause", exact: true }),
    ).toBeVisible();
    await page.getByRole("button", { name: "Pause", exact: true }).click();
    await page.getByLabel("Replay speed", { exact: true }).selectOption("4");
    await page.screenshot({
      path: info.outputPath(`openkline-${framework}.png`),
      fullPage: true,
    });
    for (let i = 0; i < 3; i++) {
      await page
        .getByRole("button", { name: "Remount chart", exact: true })
        .click();
      await expect(out).toHaveAttribute("data-actual-count", "40");
      await page
        .getByRole("button", { name: "Next step", exact: true })
        .click();
      await expect(out).toHaveAttribute("data-actual-count", "41");
    }
    expect(errors).toEqual([]);
    expect(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth,
      ),
    ).toBe(true);
  });
}
test("live depth is opt-in, replaces snapshots and returns to synthetic", async ({
  page,
}) => {
  await page.routeWebSocket(
    "wss://data-stream.binance.vision/ws/btcusdt@depth20",
    (ws) => {
      ws.send("bad JSON");
      ws.send(
        JSON.stringify({
          lastUpdateId: 1,
          bids: [["99", "2"]],
          asks: [["101", "3"]],
        }),
      );
      ws.send(
        JSON.stringify({
          lastUpdateId: 2,
          bids: [["98", "7"]],
          asks: [["102", "9"]],
        }),
      );
      ws.send(
        JSON.stringify({
          lastUpdateId: 1,
          bids: [["99", "2"]],
          asks: [["101", "3"]],
        }),
      );
    },
  );
  await page.goto("/article-demos/depth/");
  await page
    .getByRole("button", { name: "Connect live BTC/USDT", exact: true })
    .click();
  await expect(page.locator("#status")).toContainText("LIVE");
  await expect(page.locator("#metrics")).toContainText(
    "Bids: 7.00 · Asks: 9.00",
  );
  await page
    .getByRole("button", { name: "Synthetic mode", exact: true })
    .click();
  await expect(page.locator("#status")).toContainText("SYNTHETIC");
  await expect(page.locator("#metrics")).toContainText("Bids: 18.00");
});
for (const framework of ["core", "react", "vue"])
  test(`live ${framework} loads history, merges one minute and cleans up`, async ({
    page,
  }) => {
    await page.route("https://data-api.binance.vision/**", (route) =>
      route.fulfill({ json: [[1700000000000, "10", "12", "9", "11", "2"]] }),
    );
    await page.routeWebSocket(
      "wss://data-stream.binance.vision/ws/btcusdt@kline_1m",
      (ws) => {
        ws.send(
          JSON.stringify({
            k: { t: 1700000060000, o: "11", h: "14", l: "10", c: "13", v: "3" },
          }),
        );
        ws.send(
          JSON.stringify({
            k: { t: 1700000060000, o: "11", h: "15", l: "10", c: "14", v: "4" },
          }),
        );
        ws.send(
          JSON.stringify({
            k: { t: 1700000000000, o: "10", h: "12", l: "9", c: "11", v: "2" },
          }),
        );
      },
    );
    await page.goto(`/article-demos/openkline/${framework}/`);
    await page
      .getByRole("button", { name: "Connect live BTC/USDT", exact: true })
      .click();
    await expect(page.locator("#status")).toContainText("LIVE");
    await expect(page.locator("#readout")).toHaveAttribute(
      "data-actual-count",
      "2",
    );
    await expect(page.locator("#readout")).toHaveAttribute("data-close", "14");
    await page
      .getByRole("button", { name: "Synthetic replay", exact: true })
      .click();
    await expect(page.locator("#readout")).toHaveAttribute(
      "data-actual-count",
      "40",
    );
    await expect(page.locator("#readout")).toHaveAttribute(
      "data-mode",
      "synthetic",
    );
  });
test("article tabs mount only the selected integration and preserve accessible navigation", async ({
  page,
}) => {
  const requests: string[] = [];
  page.on("request", (r) => requests.push(r.url()));
  await page.goto("/articles/openkline-replay-react-vue/");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    "OpenKline: replay candles in vanilla JavaScript, React and Vue",
  );
  const chart = page.frameLocator("#openkline-frame");
  await expect(chart.locator("#readout")).toHaveAttribute(
    "data-actual-count",
    "40",
  );
  await expect(page.locator("#openkline-frame")).toHaveAttribute(
    "style",
    /height:/,
  );
  expect(requests.some((url) => /\/assets\/(react|vue)-/.test(url))).toBe(
    false,
  );
  await page.getByRole("tab", { name: "React", exact: true }).click();
  await expect(chart.getByRole("heading", { level: 1 })).toContainText("React");
  await expect(chart.locator("#readout")).toHaveAttribute(
    "data-actual-count",
    "40",
  );
  await page
    .getByRole("tab", { name: "React", exact: true })
    .press("ArrowRight");
  await expect(
    page.getByRole("tab", { name: "Vue", exact: true }),
  ).toHaveAttribute("aria-selected", "true");
  await expect(chart.getByRole("heading", { level: 1 })).toContainText("Vue");
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth,
    ),
  ).toBe(true);
});
