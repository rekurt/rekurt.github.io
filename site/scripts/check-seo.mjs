import assert from 'node:assert/strict';
import { readFile, readdir } from 'node:fs/promises';
import { join, relative } from 'node:path';
import { parse } from 'parse5';

const root = new URL('../dist/', import.meta.url).pathname;
async function files(path) {
 const entries = await readdir(path, {withFileTypes:true});
 return (await Promise.all(entries.map(e => e.isDirectory() ? files(join(path,e.name)) : e.name.endsWith('.html') ? [join(path,e.name)] : []))).flat();
}
function nodes(node, result=[]) {
 result.push(node);
 for (const child of node.childNodes || []) nodes(child,result);
 return result;
}
const attr = (node,key) => node.attrs?.find(a => a.name === key)?.value;
const text = node => (node.value || '') + (node.childNodes || []).map(text).join('');
const documents = new Map();
const titles = new Set();
for (const file of await files(root)) {
 const html = parse(await readFile(file,'utf8'));
 const all = nodes(html);
 const path = '/' + relative(root,file).replace(/index\.html$/, '');
 const fail = message => `${path}: ${message}`;
 const tags = all.filter(n => n.tagName === 'meta');
 const meta = key => tags.filter(n => attr(n,'name') === key || attr(n,'property') === key);
 const is404 = path === '/404.html';
 assert.equal(meta('robots').length,1,fail('one robots directive'));
 assert.equal(meta('robots')[0] && attr(meta('robots')[0],'content').includes('noindex'),is404,fail('404 must be noindex; real pages must be indexable'));
 const title = all.filter(n => n.tagName === 'title');
 assert.equal(title.length,1,fail('one title'));
 const titleText = text(title[0]);
 assert.ok(titleText.trim().length > 8,fail('descriptive title'));
 assert.ok(!titles.has(titleText),fail('duplicate title'));
 titles.add(titleText);
 assert.equal(meta('description').length,1,fail('one description'));
 assert.ok(attr(meta('description')[0],'content').length >= 25,fail('descriptive snippet'));
 if (is404) continue;
 const canonicals = all.filter(n => n.tagName === 'link' && attr(n,'rel') === 'canonical');
 assert.equal(canonicals.length,1,fail('one canonical'));
 const canonical = attr(canonicals[0],'href');
 assert.equal(canonical, 'https://rekurt.github.io' + path,fail('self canonical matches output URL'));
 assert.equal(attr(meta('og:url')[0],'content'),canonical,fail('social URL equals canonical'));
 assert.equal(attr(meta('og:title')[0],'content'),titleText,fail('social title equals title'));
 assert.equal(attr(meta('twitter:title')[0],'content'),titleText,fail('Twitter title equals title'));
 assert.equal(all.filter(n => n.tagName === 'h1').length,1,fail('one main heading'));
 const schemas = all.filter(n => n.tagName === 'script' && attr(n,'type') === 'application/ld+json');
 assert.ok(schemas.length > 0,fail('structured data'));
 for (const schema of schemas) assert.ok(JSON.parse(text(schema))['@context'],fail('valid structured data'));
 const alternates = all.filter(n => n.tagName === 'link' && attr(n,'rel') === 'alternate' && attr(n,'hreflang'));
 assert.equal(alternates.length,4,fail('three locales plus default'));
 assert.equal(new Set(alternates.map(n => attr(n,'hreflang'))).size,4,fail('no duplicate hreflang'));
 documents.set(canonical,{alternates,fail});
}
const sitemap = await readFile(join(root,'sitemap.xml'),'utf8');
const robots = await readFile(join(root,'robots.txt'),'utf8');
const sitemapLines = robots.split("\n").filter(line => line.startsWith("Sitemap: "));
assert.equal(sitemapLines.length, 14, 'host root robots advertises the portfolio and 13 project sitemaps');
assert.equal(new Set(sitemapLines).size, 14, 'project sitemap references are unique');
assert.ok(robots.includes('https://rekurt.github.io/Mac-Coffee/family-sitemap.xml'), 'decorated project sitemap');
assert.ok(robots.includes('https://rekurt.github.io/depth/sitemap.xml'), 'generated project sitemap');
for (const [canonical,{alternates,fail}] of documents) {
 assert.ok(sitemap.includes(`<loc>${canonical}</loc>`),fail('indexable page in sitemap'));
 for (const link of alternates) {
  const target = documents.get(attr(link,'href'));
  assert.ok(target,fail('alternate target exists'));
  assert.ok(target.alternates.some(n => attr(n,'href') === canonical),fail('reciprocal language link'));
 }
}
assert.ok(!sitemap.includes('404.html'),'404 excluded from sitemap');
console.log(`SEO verified: ${documents.size} indexable pages — titles, canonicals, robots, social tags, JSON-LD, reciprocal hreflang and sitemap.`);
