import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { runInNewContext } from 'node:vm';
const root = new URL('../', import.meta.url);
const asset = name => readFileSync(new URL('internal/projectsite/assets/' + name, root), 'utf8');
const vendor = asset('highlight-11.12.0.min.js');
const initialize = asset('syntax.js');
function engine() {
  const context = { document: { querySelectorAll: () => [] } };
  runInNewContext(vendor, context);
  runInNewContext(initialize, context);
  return context.hljs;
}
const decode = html => html.replace(/<[^>]+>/g, '').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&#x27;/g, "'").replace(/&amp;/g, '&');
const examples = {
  "typescript": "const count: number = 42;",
  "javascript": "const text = \"hello\";",
  "go": "func main() { fmt.Println(\"hello\") }",
  "rust": "fn main() { let count = 42; }",
  "sql": "SELECT id FROM users WHERE active = true;",
  "bash": "npm install \"@rekurt/depth\"",
  "json": "{\"count\":42,\"name\":\"hello\"}",
  "yaml": "name: hello\ncount: 42\n",
  "toml": "[project]\nname = \"hello\"",
  "dockerfile": "FROM node:24\nRUN npm ci",
  "diff": "- old\n+ new",
  "xml": "<button type=\"button\">Hello</button>",
  "cli": "npm install @rekurt/depth --save-dev\\ngit barber --list\\ngo get github.com/rekurt/ymsdk@v1.0.0"
};
for (const [language, source] of Object.entries(examples)) test(language + ' has tokens and keeps every source character', () => {
  const result = engine().highlight(source, { language, ignoreIllegals: true });
  assert.match(result.value, /class="hljs-/);
  assert.equal(decode(result.value), source);
});
test('HTML inside a source string stays escaped', () => {
  const source = 'const markup = "<img src=x onerror=alert(1)>";';
  const result = engine().highlight(source, {language:'typescript'});
  assert.ok(!result.value.includes('<img'));
  assert.equal(decode(result.value), source);
});
test('portfolio and generated project sites ship identical assets', () => {
  for (const name of ['highlight-11.12.0.min.js', 'syntax.js', 'syntax.css', 'highlight.LICENSE.txt']) {
    assert.equal(readFileSync(new URL('site/public/syntax/' + name, root),'utf8'),asset(name));
  }
  assert.match(asset('highlight.LICENSE.txt'), /BSD 3-Clause License/);
});
test('token palettes meet AA contrast on actual code surfaces', () => {
  function luminance(hex) {
    const rgb=hex.slice(1).match(/../g).map(v=>parseInt(v,16)/255).map(v=>v<=.04045?v/12.92:((v+.055)/1.055)**2.4);
    return .2126*rgb[0]+.7152*rgb[1]+.0722*rgb[2];
  }
  const palettes = [
    [['#d8b4fe','#a7f3d0','#fcd34d','#93c5fd','#a7b4c4','#67e8f9','#fda4af'],['#101b24','#192c22','#18261e','#1d2b3c','#1c222d','#101e21']],
    [['#6b21a8','#166534','#92400e','#1e40af','#4b5563','#155e75','#9f1239'],['#e1ddcf','#e9e7db','#ece4d9','#fffaf0','#fff0e3','#e9eadf']],
  ];
  for(const [colours, backgrounds] of palettes) for(const colour of colours) for(const background of backgrounds) {
    const c=luminance(colour), b=luminance(background);
    assert.ok((Math.max(c,b)+.05)/(Math.min(c,b)+.05)>=4.5, colour+' on '+background);
    assert.ok(asset('syntax.css').includes(colour));
  }
});
