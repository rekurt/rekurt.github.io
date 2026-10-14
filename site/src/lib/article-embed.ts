export function attachChartHeight() {
  const frames = [
    ...document.querySelectorAll<HTMLIFrameElement>(".demo iframe"),
  ];
  const request = (frame: HTMLIFrameElement) =>
    frame.contentWindow?.postMessage(
      { type: "rekurt-chart-height-request" },
      location.origin,
    );
  for (const frame of frames) {
    frame.addEventListener("load", () => request(frame));
    request(frame);
  }
  window.addEventListener("message", (event) => {
    if (
      event.origin !== location.origin ||
      event.data?.type !== "rekurt-chart-height"
    )
      return;
    const frame = frames.find((frame) => frame.contentWindow === event.source),
      height = event.data.height;
    if (frame && Number.isFinite(height) && height >= 400 && height <= 4000)
      frame.style.height = `${height}px`;
  });
}
