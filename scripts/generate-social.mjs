import { readFile, mkdir, writeFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
const require = createRequire(new URL('../site/package.json',import.meta.url));
const sharp = require('sharp');
const titles = JSON.parse(await readFile(new URL('../internal/projectsite/seo-titles.json',import.meta.url),'utf8'));
const escape = s => s.replaceAll('&','&amp;').replaceAll('<','&lt;');
for (const [slug,localized] of Object.entries({...titles, rekurt:{en:'rekurt — engineering, open source and developer tools'}})) {
 const [name,summary] = localized.en.split(' — ');
 const words = summary.split(' '); const lines=[''];
 for (const word of words) { if ((lines.at(-1)+' '+word).trim().length > 44) lines.push(''); lines[lines.length-1] = (lines.at(-1)+' '+word).trim(); }
 const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="630"><rect width="1200" height="630" fill="#080b10"/><path d="M64 60H1136V570H64Z" fill="#0f1720" stroke="#263340"/><path d="M64 60H1136" stroke="#38d5ff" stroke-width="6"/><text x="105" y="132" fill="#38d5ff" font-family="sans-serif" font-size="25" letter-spacing="3">REKURT / OPEN SOURCE</text><text x="100" y="282" fill="#f2f5fa" font-family="sans-serif" font-weight="700" font-size="${name.length>16?65:90}">${escape(name)}</text>${lines.map((line,i)=>`<text x="105" y="${355+i*43}" fill="#a6b4c4" font-family="sans-serif" font-size="32">${escape(line)}</text>`).join('')}<text x="105" y="528" fill="#38d5ff" font-family="monospace" font-size="23">${slug === 'rekurt' ? 'rekurt.github.io' : 'github.com/rekurt'}</text><path d="M1000 495H1080M1060 475L1080 495L1060 515" fill="none" stroke="#38d5ff" stroke-width="4"/></svg>`;
 const png=await sharp(Buffer.from(svg)).png().toBuffer();
 for (const dir of ['site/public/social','internal/projectsite/assets/social']) {
  const path=new URL(`../${dir}/`,import.meta.url); await mkdir(path,{recursive:true}); await writeFile(new URL(slug+'.png',path),png);
 }
}
console.log('Generated 15 social cards (1200 × 630).');
