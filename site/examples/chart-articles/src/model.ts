import type { Candle } from "@rekurt/openkline-core";
import type { DepthSnapshot } from "@rekurt/depth";
export const books: { name: string; book: DepthSnapshot }[] = [
  {
    name: "Balanced book",
    book: {
      bids: [
        { price: 99, volume: 4 },
        { price: 98, volume: 6 },
        { price: 97, volume: 5 },
        { price: 96, volume: 3 },
      ],
      asks: [
        { price: 101, volume: 4 },
        { price: 102, volume: 6 },
        { price: 103, volume: 5 },
        { price: 104, volume: 3 },
      ],
    },
  },
  {
    name: "Large ask appears",
    book: {
      bids: [
        { price: 99, volume: 4 },
        { price: 98, volume: 6 },
        { price: 97, volume: 5 },
        { price: 96, volume: 3 },
      ],
      asks: [
        { price: 101, volume: 4 },
        { price: 102, volume: 30 },
        { price: 103, volume: 5 },
        { price: 104, volume: 3 },
      ],
    },
  },
  {
    name: "Near-touch liquidity disappears",
    book: {
      bids: [
        { price: 97, volume: 5 },
        { price: 96, volume: 3 },
      ],
      asks: [
        { price: 103, volume: 5 },
        { price: 104, volume: 3 },
      ],
    },
  },
];
export function cumulative(book: DepthSnapshot) {
  const side = (levels: DepthSnapshot["bids"], sign: number) => {
    let sum = 0;
    return [...levels]
      .sort((a, b) => sign * (a.price - b.price))
      .map((l) => ({
        price: l.price,
        volume: l.volume,
        cumulative: (sum += l.volume),
      }));
  };
  return { bids: side(book.bids, -1), asks: side(book.asks, 1) };
}
export const candles: Candle[] = Array.from({ length: 120 }, (_, i) => {
  const close = (j: number) => 42000 + Math.sin(j / 7) * 180 + j * 3;
  const o = close(i - 1),
    c = close(i);
  return {
    t: 1700000000 + i * 60,
    o,
    h: Math.max(o, c) + 30,
    l: Math.min(o, c) - 25,
    c,
    v: 100 + (i % 13) * 17,
  };
});
export const maxStep = (candles.length - 40) * 2;
export function replayAt(step: number): Candle[] {
  step = Math.max(0, Math.min(maxStep, step));
  const complete = 40 + Math.floor(step / 2);
  const visible = candles.slice(0, complete).map((c) => ({ ...c }));
  if (step % 2) {
    const final = candles[complete]!;
    const c = (final.o + final.c) / 2;
    visible.push({
      ...final,
      c,
      h: Math.max(final.o, c) + 10,
      l: Math.min(final.o, c) - 10,
      v: final.v / 2,
    });
  }
  return visible;
}
export function parseCandle(row: unknown): Candle {
  if (!Array.isArray(row) || row.length < 6) throw Error("Invalid candle");
  const [t, o, h, l, c, v] = row.map(Number);
  return validateCandle({ t: t! / 1000, o: o!, h: h!, l: l!, c: c!, v: v! });
}
export function validateCandle(c: Candle): Candle {
  if (
    !Object.values(c).every(Number.isFinite) ||
    c.t <= 0 ||
    c.v < 0 ||
    c.l <= 0 ||
    c.l > Math.min(c.o, c.c) ||
    c.h < Math.max(c.o, c.c)
  )
    throw Error("Invalid OHLCV values");
  return c;
}
export function parseKline(raw: unknown): Candle {
  const d = raw as {
    k?: { t: number; o: string; h: string; l: string; c: string; v: string };
  };
  if (!d?.k) throw Error("Invalid kline");
  const k = d.k;
  return parseCandle([k.t, k.o, k.h, k.l, k.c, k.v]);
}
export function mergeCandle(data: Candle[], c: Candle): Candle[] {
  const last = data.at(-1);
  if (!last || c.t > last.t) return [...data, c].slice(-300);
  if (c.t === last.t) return [...data.slice(0, -1), c];
  return data;
}
export function parseBook(raw: unknown): DepthSnapshot {
  const d = raw as { bids?: unknown; asks?: unknown; lastUpdateId?: unknown };
  const side = (x: unknown) => {
    if (!Array.isArray(x) || !x.length || x.length > 20)
      throw Error("Invalid depth");
    return x.map((row) => {
      if (!Array.isArray(row) || row.length < 2) throw Error("Invalid level");
      const [price, volume] = row.map(Number);
      if (
        !Number.isFinite(price) ||
        !Number.isFinite(volume) ||
        price! <= 0 ||
        volume! < 0
      )
        throw Error("Invalid level values");
      return { price: price!, volume: volume! };
    });
  };
  if (!Number.isSafeInteger(d?.lastUpdateId)) throw Error("Invalid update ID");
  return {
    bids: side(d.bids),
    asks: side(d.asks),
    sequence: d.lastUpdateId as number,
  };
}
