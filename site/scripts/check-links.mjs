import { access, readdir, readFile } from "node:fs/promises";
import { dirname, extname, join, relative, resolve, sep } from "node:path";
import { pathToFileURL } from "node:url";

import { parse } from "parse5";

async function htmlFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const nested = await Promise.all(entries.map(async (entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return htmlFiles(path);
    return entry.isFile() && extname(entry.name) === ".html" ? [path] : [];
  }));
  return nested.flat().sort();
}

function collectReferences(node, references = []) {
  if (node.attrs) {
    for (const attribute of node.attrs) {
      if (attribute.name === "href" || attribute.name === "src") references.push(attribute.value);
    }
  }
  for (const child of node.childNodes ?? []) collectReferences(child, references);
  return references;
}

async function exists(path) {
  try {
    await access(path);
    return true;
  } catch {
    return false;
  }
}

function targetCandidates(root, sourceFile, reference) {
  const path = reference.split("#", 1)[0].split("?", 1)[0];
  const decoded = decodeURIComponent(path);
  const sourceDirectory = dirname(sourceFile);
  const target = decoded.startsWith("/") ? resolve(root, `.${decoded}`) : resolve(sourceDirectory, decoded);
  const candidates = [target];
  if (decoded.endsWith("/") || extname(target) === "") candidates.push(join(target, "index.html"));
  if (extname(target) === "") candidates.push(`${target}.html`);
  return candidates;
}

function isInternal(reference) {
  if (!reference || reference.startsWith("#")) return false;
  return !/^(?:[a-z][a-z0-9+.-]*:|\/\/)/i.test(reference);
}

export async function checkInternalLinks(directory) {
  const root = resolve(directory);
  const files = await htmlFiles(root);
  let checkedLinks = 0;
  const errors = [];
  for (const file of files) {
    const document = parse(await readFile(file, "utf8"));
    for (const reference of collectReferences(document)) {
      if (!isInternal(reference)) continue;
      checkedLinks += 1;
      const candidates = targetCandidates(root, file, reference);
      const safeCandidates = candidates.filter((candidate) => relative(root, candidate).split(sep)[0] !== "..");
      if (!safeCandidates.length) {
        errors.push(`${relative(root, file)}: link escapes output root ${reference}`);
        continue;
      }
      if (!(await Promise.all(safeCandidates.map(exists))).some(Boolean)) {
        errors.push(`${relative(root, file)}: broken internal link ${reference}`);
      }
    }
  }
  if (errors.length) throw new Error(errors.join("\n"));
  return { checkedFiles: files.length, checkedLinks };
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  const result = await checkInternalLinks(process.argv[2] ?? "dist");
  process.stdout.write(`checked ${result.checkedLinks} internal references in ${result.checkedFiles} HTML files\n`);
}
