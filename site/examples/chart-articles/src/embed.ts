// Same-origin height reporting keeps the embedded demo usable on narrow screens.
if (window.parent !== window) {
  const report = () =>
    window.parent.postMessage(
      {
        type: "rekurt-chart-height",
        height: Math.max(
          400,
          Math.ceil(document.body.getBoundingClientRect().height),
        ),
      },
      location.origin,
    );
  const observer = new ResizeObserver(report);
  observer.observe(document.body);
  window.addEventListener("message", (event) => {
    if (
      event.origin === location.origin &&
      event.source === window.parent &&
      event.data?.type === "rekurt-chart-height-request"
    )
      report();
  });
  window.addEventListener("pagehide", () => observer.disconnect(), {
    once: true,
  });
  report();
}
