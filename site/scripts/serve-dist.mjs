import { createReadStream } from "node:fs";
import { stat } from "node:fs/promises";
import { createServer } from "node:http";
import { extname, join, normalize, resolve } from "node:path";

const root = resolve("dist");
const port = Number(process.env.PORT ?? 4323);
const contentTypes = { ".css": "text/css", ".html": "text/html; charset=utf-8", ".js": "text/javascript", ".json": "application/json", ".svg": "image/svg+xml", ".txt": "text/plain; charset=utf-8", ".xml": "application/xml; charset=utf-8" };

async function resolveFile(pathname) {
  const decoded = decodeURIComponent(pathname);
  const candidate = normalize(join(root, decoded));
  if (!candidate.startsWith(root)) return null;
  const paths = decoded.endsWith("/") ? [join(candidate, "index.html")] : [candidate, join(candidate, "index.html")];
  for (const path of paths) {
    try {
      if ((await stat(path)).isFile()) return path;
    } catch {}
  }
  return null;
}

createServer(async (request, response) => {
  const pathname = new URL(request.url ?? "/", `http://${request.headers.host}`).pathname;
  const path = await resolveFile(pathname) ?? join(root, "404.html");
  response.statusCode = path.endsWith("404.html") ? 404 : 200;
  response.setHeader("Content-Type", contentTypes[extname(path)] ?? "application/octet-stream");
  createReadStream(path).pipe(response);
}).listen(port, "127.0.0.1", () => process.stdout.write(`static preview: http://127.0.0.1:${port}\n`));
