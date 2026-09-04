import type { APIRoute } from "astro";

import { allSitePaths, escapeXml } from "../lib/routes";

export const GET: APIRoute = ({ site }) => {
  const paths = allSitePaths().filter((path) => path !== "/robots.txt" && path !== "/sitemap.xml");
  const urls = paths.map((path) => {
    const isRussian = path.startsWith("/ru/");
    const equivalent = isRussian ? path.replace(/^\/ru/, "") || "/" : path === "/" ? "/ru/" : `/ru${path}`;
    const enPath = isRussian ? equivalent : path;
    const ruPath = isRussian ? path : equivalent;
    return `<url><loc>${escapeXml(String(new URL(path, site)))}</loc><xhtml:link rel="alternate" hreflang="en" href="${escapeXml(String(new URL(enPath, site)))}"/><xhtml:link rel="alternate" hreflang="ru" href="${escapeXml(String(new URL(ruPath, site)))}"/></url>`;
  }).join("");
  return new Response(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">${urls}</urlset>`, { headers: { "Content-Type": "application/xml; charset=utf-8" } });
};
