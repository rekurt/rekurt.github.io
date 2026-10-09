document.addEventListener("click", async (event) => {
  const button = event.target.closest("[data-copy]");
  if (!button) return;
  const label = button.querySelector("span");
  const original = label.textContent;
  try {
    await navigator.clipboard.writeText(button.dataset.copy);
    label.textContent = document.documentElement.lang === "ru" ? "Скопировано" : document.documentElement.lang === "zh-CN" ? "已复制" : "Copied";
    window.setTimeout(() => { label.textContent = original; }, 1600);
  } catch {
    label.textContent = button.dataset.copy;
  }
});
