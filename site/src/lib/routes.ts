import { getProducts, type Locale } from "./catalog";

export function productPaths(locale: Locale) {
  return getProducts(locale).map((product) => ({
    params: { slug: product.slug },
    props: { locale, product },
  }));
}

export function allSitePaths(): string[] {
  const staticPaths = ["/", "/projects/", "/registry/", "/about/"];
  const productSlugs = getProducts("en").map((product) => product.slug);
  return [
    ...staticPaths,
    ...productSlugs.map((slug) => `/projects/${slug}/`),
    ...staticPaths.map((path) => (path === "/" ? "/ru/" : `/ru${path}`)),
    ...productSlugs.map((slug) => `/ru/projects/${slug}/`),
    "/robots.txt",
    "/sitemap.xml",
  ];
}

export function escapeXml(value: string): string {
  return value.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&apos;");
}
