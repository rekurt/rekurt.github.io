import { getProducts, type Locale } from "./catalog";

export const locales: Locale[] = ["en", "ru", "zh-cn"];

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
    ...locales.flatMap((locale) => {
      const prefix = locale === "en" ? "" : `/${locale}`;
      return [
        ...staticPaths.map((path) => (path === "/" ? `${prefix}/` : `${prefix}${path}`)),
        ...productSlugs.map((slug) => `${prefix}/projects/${slug}/`),
      ];
    }),
    "/robots.txt",
    "/sitemap.xml",
  ];
}

export function escapeXml(value: string): string {
  return value.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&apos;");
}
