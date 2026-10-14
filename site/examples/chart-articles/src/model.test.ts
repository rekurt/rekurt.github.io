import { describe, it, expect } from "vitest";
import {
  books,
  cumulative,
  replayAt,
  maxStep,
  parseCandle,
  parseBook,
  parseKline,
  mergeCandle,
} from "./model";
describe("deterministic article examples", () => {
  it("has the intended cumulative depth and spread changes", () => {
    expect(cumulative(books[0]!.book).bids.map((x) => x.cumulative)).toEqual([
      4, 10, 15, 18,
    ]);
    expect(cumulative(books[1]!.book).asks.map((x) => x.cumulative)).toEqual([
      4, 34, 39, 42,
    ]);
    expect(books[2]!.book.asks[0]!.price - books[2]!.book.bids[0]!.price).toBe(
      6,
    );
  });
  it("reveals one partial candle then updates it, never renders a future prefix", () => {
    const start = replayAt(0),
      partial = replayAt(1),
      final = replayAt(2);
    expect(start).toHaveLength(40);
    expect(partial).toHaveLength(41);
    expect(final).toHaveLength(41);
    expect(partial.at(-1)!.t).toBe(final.at(-1)!.t);
    expect(partial.at(-1)!.c).not.toBe(final.at(-1)!.c);
    for (let i = 0; i <= maxStep; i++)
      expect(replayAt(i).length).toBe(40 + Math.ceil(i / 2));
    expect(replayAt(maxStep)).toHaveLength(120);
    expect(replayAt(0)).toEqual(start);
  });
  it("converts milliseconds and updates only same or newer timestamps", () => {
    const c = parseCandle([1700000000000, "10", "12", "9", "11", "2"]);
    expect(c.t).toBe(1700000000);
    expect(
      parseKline({
        k: { t: 1700000000000, o: "10", h: "12", l: "9", c: "11", v: "2" },
      }),
    ).toEqual(c);
    expect(mergeCandle([c], { ...c, c: 12 })).toEqual([{ ...c, c: 12 }]);
    expect(mergeCandle([c], { ...c, t: c.t + 60 })).toHaveLength(2);
    const data = [c];
    expect(mergeCandle(data, { ...c, t: c.t - 60 })).toBe(data);
    expect(() => parseCandle([1, "10", "9", "12", "11", 2])).toThrow();
  });
  it("accepts full partial-book snapshots, rejects corrupt values", () => {
    expect(
      parseBook({ lastUpdateId: 1, bids: [["99", "2"]], asks: [["101", "3"]] })
        .bids,
    ).toEqual([{ price: 99, volume: 2 }]);
    expect(() =>
      parseBook({ lastUpdateId: 2, bids: [["bad", "2"]], asks: [] }),
    ).toThrow();
  });
});
