import type { APIRoute } from "astro";
import { getProducts } from "../lib/catalog";

// Crawlers read robots.txt only at the host root. Project robots files under
// /project/ cannot advertise their own sitemaps to the crawler.
export const GET: APIRoute = ({ site }) => {
  const decorated = new Set(["openkline", "mac-coffee", "chislo"]);
  const projectSitemaps = getProducts("en").flatMap(product => product.links
    .filter(link => link.kind === "website" || link.kind === "documentation")
    .map(link => new URL(link.url))
    .filter(url => url.origin === "https://rekurt.github.io" && /^\/[^/]+\/$/.test(url.pathname))
    .map(url => String(new URL(decorated.has(product.slug) ? "family-sitemap.xml" : "sitemap.xml", url))));
  const sitemaps = [...new Set([String(new URL("/sitemap.xml", site)), ...projectSitemaps])];
  return new Response(`User-agent: *\nAllow: /\n${sitemaps.map(url => `Sitemap: ${url}`).join("\n")}\n`, { headers: { "Content-Type": "text/plain; charset=utf-8" } });
};
