import assert from 'node:assert/strict';
import { readFile, readdir } from 'node:fs/promises';

const directory = new URL('../internal/projectsite/profiles/', import.meta.url);
const decode = (value) => value.replace(/&#(\d+);/g, (_, n) => String.fromCodePoint(Number(n))).replace(/&#x([0-9a-f]+);/gi, (_, n) => String.fromCodePoint(parseInt(n, 16))).replace(/&quot;/g, '"').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&');
const assets = new Set();
let pages = 0;

async function get(url) {
  const response = await fetch(url, { signal: AbortSignal.timeout(30000), redirect: 'error' });
  assert.equal(response.status, 200, `${url}: HTTP ${response.status}`);
  return response;
}

for (const filename of (await readdir(directory)).filter((name) => name.endsWith('.json')).sort()) {
  const profile = JSON.parse(await readFile(new URL(filename, directory), 'utf8'));
  const base = `https://rekurt.github.io/${profile.slug}/`;
  for (const [locale, copy] of Object.entries(profile.locales)) {
    const url = base + (locale === 'en' ? '' : `${locale}/`);
    const html = decode(await (await get(url)).text());
    const required = [
      `data-product-profile="${profile.slug}"`, `theme-${profile.theme}`,
      `<h1>${copy.headline}</h1>`, copy.intro, copy.cta, copy.closing,
      `href="${profile.primary}"`, `id="${profile.primary.slice(1)}"`,
      `rel="canonical" href="${url}"`, 'data-rekurt-family',
      ...copy.features.flatMap((item) => [item.title, item.body]),
      ...copy.steps.flatMap((item) => [item.title, item.body]),
    ];
    for (const value of required) assert.ok(html.includes(value), `${url}: missing ${value}`);
    const directoryURL = `${url}projects/`;
    const links = [...html.matchAll(/<a\b[^>]*href="([^"]+)"/g)].map((match) => new URL(match[1], url).href);
    assert.ok(links.includes(directoryURL), `${url}: missing directory link ${directoryURL}`);
    assert.ok(!html.includes(`${profile.slug}.system`), `${url}: generic hero remains`);
    if (profile.image) assert.ok(html.includes(`/${profile.image}"`), `${url}: product image missing`);
    for (const match of html.matchAll(/(?:src|href)="([^"]*(?:marketing\.css|raw\.githubusercontent\.com[^"\s]*\.(?:gif|png)))"/g)) {
      assets.add(new URL(match[1], url).href);
    }
    pages++;
  }
  console.log(`PASS ${profile.slug}: en, ru, zh-cn — content, CTA, theme, canonical, family`);
}

for (const url of assets) {
  const response = await get(url);
  const type = response.headers.get('content-type') || '';
  assert.ok(url.endsWith('.css') ? type.includes('text/css') : type.startsWith('image/'), `${url}: unexpected ${type}`);
  await response.arrayBuffer();
}
console.log(`Verified ${pages} published marketing pages and ${assets.size} assets.`);
