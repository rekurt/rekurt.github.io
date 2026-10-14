import { afterEach, beforeEach, it, expect, vi } from "vitest";
import { LiveFeed } from "./live";
class Socket {
  static OPEN = 1;
  static all: Socket[] = [];
  readyState = 0;
  onopen?: () => void;
  onclose?: () => void;
  onerror?: () => void;
  onmessage?: (e: { data: string }) => void;
  constructor(public url: string) {
    Socket.all.push(this);
  }
  open() {
    this.readyState = 1;
    this.onopen?.();
  }
  message(data: unknown) {
    this.onmessage?.({ data: JSON.stringify(data) });
  }
  close() {
    if (this.readyState === 3) return;
    this.readyState = 3;
    this.onclose?.();
  }
}
beforeEach(() => {
  vi.useFakeTimers();
  Socket.all = [];
  vi.stubGlobal("WebSocket", Socket);
});
afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});
it("buffers while loading history, rejects malformed messages, and cleans up on stop", async () => {
  const receive = vi.fn((x) => {
      if (!x.good) throw Error();
    }),
    status = vi.fn();
  let done!: () => void;
  const feed = new LiveFeed({
    stream: "test",
    status,
    receive,
    initialize: () => new Promise<void>((r) => (done = r)),
  });
  feed.start();
  const ws = Socket.all[0]!;
  ws.open();
  ws.message({ good: true });
  expect(receive).not.toHaveBeenCalled();
  done();
  await Promise.resolve();
  expect(receive).toHaveBeenCalledOnce();
  ws.message({ bad: true });
  expect(status).toHaveBeenLastCalledWith("Ignored malformed market data");
  feed.stop();
  expect(vi.getTimerCount()).toBe(0);
});
it("uses six bounded reconnect delays and reinitializes on reconnection", async () => {
  const initialize = vi.fn(async () => {}),
    status = vi.fn(),
    feed = new LiveFeed({
      stream: "test",
      status,
      initialize,
      receive: () => {},
    });
  feed.start();
  const waits = [1000, 2000, 4000, 8000, 16000, 30000];
  for (let i = 0; i < 6; i++) {
    Socket.all.at(-1)!.open();
    await Promise.resolve();
    Socket.all.at(-1)!.close();
    await vi.advanceTimersByTimeAsync(waits[i]! - 1);
    expect(Socket.all).toHaveLength(i + 1);
    await vi.advanceTimersByTimeAsync(1);
    expect(Socket.all).toHaveLength(i + 2);
  }
  Socket.all.at(-1)!.close();
  expect(status).toHaveBeenLastCalledWith(
    expect.stringContaining("retry limit"),
  );
  expect(initialize).toHaveBeenCalledTimes(6);
  feed.stop();
  expect(vi.getTimerCount()).toBe(0);
});
it("ignores initialization completion after a stopped generation", async () => {
  const status = vi.fn();
  let reject!: (e: Error) => void;
  const feed = new LiveFeed({
    stream: "test",
    status,
    receive: () => {},
    initialize: () => new Promise((_, r) => (reject = r)),
  });
  feed.start();
  Socket.all[0]!.open();
  feed.stop();
  status.mockClear();
  reject(Error("aborted"));
  await Promise.resolve();
  expect(status).not.toHaveBeenCalled();
  expect(vi.getTimerCount()).toBe(0);
});
it("marks an open connection stale without fabricating new data", async () => {
  const status = vi.fn(),
    receive = vi.fn();
  const feed = new LiveFeed({ stream: "test", status, receive });
  feed.start();
  Socket.all[0]!.open();
  await Promise.resolve();
  Socket.all[0]!.message({ valid: true });
  await vi.advanceTimersByTimeAsync(12000);
  expect(status).toHaveBeenLastCalledWith(expect.stringContaining("stale"));
  expect(receive).toHaveBeenCalledOnce();
  feed.stop();
  expect(vi.getTimerCount()).toBe(0);
});
