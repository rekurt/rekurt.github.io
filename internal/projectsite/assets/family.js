const copyTimers = new WeakMap();

function openDocumentation() {
  const disclosure = document.querySelector('.documentation-disclosure');
  if (disclosure) disclosure.open = true;
}

if (window.location.hash === '#overview') openDocumentation();
window.addEventListener('hashchange', () => {
  if (window.location.hash === '#overview') openDocumentation();
});

document.addEventListener('click', async (event) => {
  if (event.target.closest('a[href="#overview"]')) openDocumentation();
  const button = event.target.closest('[data-copy]');
  if (!button || button.disabled) return;
  const label = button.querySelector('span');
  const original = button.dataset.copyLabel || label.textContent;
  const status = document.querySelector('.copy-status');
  clearTimeout(copyTimers.get(button));
  button.disabled = true;
  if (status) status.textContent = '';
  try {
    await navigator.clipboard.writeText(button.dataset.copy);
    const copied = status?.dataset.copied || ({ ru: 'Скопировано', 'zh-CN': '已复制' }[document.documentElement.lang] || 'Copied');
    label.textContent = copied;
    if (status) status.textContent = copied;
    copyTimers.set(button, window.setTimeout(() => { label.textContent = original; }, 1600));
  } catch {
    label.textContent = original;
    if (status) status.textContent = status.dataset.copyFailed;
  } finally {
    button.disabled = false;
  }
});
