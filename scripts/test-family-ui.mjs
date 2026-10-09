import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { runInNewContext } from 'node:vm';

const script = readFileSync(new URL('../internal/projectsite/assets/family.js', import.meta.url), 'utf8');
function page(writeText, hash = '') {
  const listeners = {};
  const label = { textContent: 'Копировать' };
  const button = { disabled: false, dataset: { copy: 'npm install @rekurt/depth', copyLabel: 'Копировать' }, querySelector: () => label };
  const status = { textContent: '', dataset: { copied: 'Скопировано', copyFailed: 'Выделите команду вручную' } };
  const disclosure = { open: false };
  const timeouts = [];
  const document = {
    querySelector: selector => selector === '.copy-status' ? status : disclosure,
    addEventListener: (name, callback) => { listeners[name] = callback; },
  };
  const window = { location: { hash }, addEventListener: (name, callback) => { listeners[name] = callback; }, setTimeout: callback => { timeouts.push(callback); return timeouts.length; } };
  runInNewContext(script, { document, window, navigator: { clipboard: { writeText } }, clearTimeout() {} });
  const click = () => listeners.click({ target: { closest: selector => selector === '[data-copy]' ? button : null } });
  return { button, label, status, disclosure, timeouts, click, listeners, window };
}

test('copy writes the exact command and announces success without losing the original label', async () => {
  const written = [];
  const ui = page(async text => written.push(text));
  await ui.click();
  assert.deepEqual(written, ['npm install @rekurt/depth']);
  assert.equal(ui.status.textContent, 'Скопировано');
  assert.equal(ui.label.textContent, 'Скопировано');
  assert.equal(ui.button.disabled, false);
  ui.timeouts[0]();
  assert.equal(ui.label.textContent, 'Копировать');
});

test('clipboard refusal keeps the button usable and announces manual copying', async () => {
  const ui = page(async () => { throw new Error('Permission denied'); });
  await ui.click();
  assert.equal(ui.label.textContent, 'Копировать');
  assert.equal(ui.status.textContent, 'Выделите команду вручную');
  assert.equal(ui.button.disabled, false);
});

test('a second click while the clipboard request is pending does not send another request', async () => {
  let resolve;
  let requests = 0;
  const ui = page(() => { requests++; return new Promise(done => { resolve = done; }); });
  const pending = ui.click();
  await ui.click();
  assert.equal(requests, 1);
  resolve();
  await pending;
  assert.equal(ui.button.disabled, false);
});

test('documentation deep links and repeated navigation reveal the README', async () => {
  const ui = page(async () => {}, '#overview');
  assert.equal(ui.disclosure.open, true);
  ui.disclosure.open = false;
  await ui.listeners.click({ target: { closest: selector => selector === 'a[href="#overview"]' ? {} : null } });
  assert.equal(ui.disclosure.open, true);
  ui.disclosure.open = false;
  ui.listeners.hashchange();
  assert.equal(ui.disclosure.open, true);
});
