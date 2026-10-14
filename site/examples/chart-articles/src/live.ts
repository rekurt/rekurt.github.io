const waits = [1000, 2000, 4000, 8000, 16000, 30000];
export interface LiveOptions {
  stream: string;
  initialize?: (signal: AbortSignal) => Promise<void>;
  receive: (data: unknown) => void;
  status: (text: string) => void;
}
/** One explicitly started feed. No connections are opened at module load. */
export class LiveFeed {
  private deadline?: ReturnType<typeof setTimeout>;
  private socket?: WebSocket;
  private timer?: ReturnType<typeof setTimeout>;
  private monitor?: ReturnType<typeof setInterval>;
  private abort?: AbortController;
  private stopped = false;
  private retries = 0;
  private epoch = 0;
  private last = 0;
  constructor(private options: LiveOptions) {}
  start() {
    this.stopped = false;
    this.connect();
    this.monitor = setInterval(() => {
      if (
        this.socket?.readyState === WebSocket.OPEN &&
        this.last &&
        Date.now() - this.last > 10000
      )
        this.options.status(
          "Live data stale · last received " +
            new Date(this.last).toLocaleTimeString(),
        );
    }, 2000);
  }
  private connect() {
    if (this.stopped) return;
    const epoch = ++this.epoch;
    this.options.status("Connecting to Binance public market data…");
    const ws = (this.socket = new WebSocket(
      "wss://data-stream.binance.vision/ws/" + this.options.stream,
    ));
    let ready = false;
    let queue: unknown[] = [];
    this.abort?.abort();
    const abort = (this.abort = new AbortController());
    const deadline = (this.deadline = setTimeout(() => ws.close(), 10000));
    ws.onopen = async () => {
      try {
        await this.options.initialize?.(abort.signal);
        if (
          this.stopped ||
          epoch !== this.epoch ||
          ws.readyState !== WebSocket.OPEN
        )
          return;
        ready = true;
        this.options.status("Connected · waiting for fresh market data…");
        for (const data of queue) this.deliver(data);
        queue = [];
      } catch {
        if (!this.stopped && epoch === this.epoch) {
          this.options.status("Initial market data unavailable");
          ws.close();
        }
      }
    };
    ws.onmessage = (e) => {
      if (this.stopped || epoch !== this.epoch) return;
      try {
        const data = JSON.parse(e.data);
        if ((data as { e?: string }).e === "serverShutdown") {
          ws.close();
          return;
        }
        if (ready) this.deliver(data);
        else {
          queue.push(data);
          if (queue.length > 100) ws.close();
        }
      } catch {
        this.options.status("Ignored malformed market data");
      }
    };
    ws.onerror = () => {
      if (this.stopped || epoch !== this.epoch) return;
      this.options.status("Market stream unavailable");
      ws.close();
    };
    ws.onclose = () => {
      clearTimeout(deadline);
      abort.abort();
      if (this.stopped || epoch !== this.epoch) return;
      this.epoch++;
      if (this.retries >= waits.length) {
        clearInterval(this.monitor);
        this.options.status(
          "Live unavailable · retry limit reached. Return to synthetic or reconnect.",
        );
        return;
      }
      const delay = waits[this.retries++]!;
      this.options.status(
        "Disconnected · data stale · reconnecting in " + delay / 1000 + "s",
      );
      this.timer = setTimeout(() => this.connect(), delay);
    };
  }
  private deliver(data: unknown) {
    try {
      this.options.receive(data);
      clearTimeout(this.deadline);
      this.last = Date.now();
      this.options.status(
        "LIVE · Binance · received " + new Date(this.last).toLocaleTimeString(),
      );
    } catch {
      this.options.status("Ignored malformed market data");
    }
  }
  stop() {
    this.stopped = true;
    this.epoch++;
    this.abort?.abort();
    clearTimeout(this.timer);
    clearTimeout(this.deadline);
    clearInterval(this.monitor);
    this.socket?.close();
    this.socket = undefined;
  }
}
