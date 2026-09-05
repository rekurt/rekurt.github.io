import type { APIRoute } from "astro";

import { alternatePath } from "../i18n/copy";
import { allSitePaths, escapeXml, locales } from "../lib/routes";

export const GET: APIRoute = ({ site }) => {
  const paths = allSitePaths().filter((path) => path !== "/robots.txt" && path !== "/sitemap.xml");
  const urls = paths.map((path) => {
    const alternates = locales.map((locale) => {
      const lang = locale === "zh-cn" ? "zh-CN" : locale;
      return `<xhtml:link rel="alternate" hreflang="${lang}" href="${escapeXml(String(new URL(alternatePath(path, locale), site)))}"/>`;
    }).join("");
    const english = escapeXml(String(new URL(alternatePath(path, "en"), site)));
    return `<url><loc>${escapeXml(String(new URL(path, site)))}</loc>${alternates}<xhtml:link rel="alternate" hreflang="x-default" href="${english}"/></url>`;
  }).join("");
  return new Response(`<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">${urls}</urlset>`, { headers: { "Content-Type": "application/xml; charset=utf-8" } });
};
