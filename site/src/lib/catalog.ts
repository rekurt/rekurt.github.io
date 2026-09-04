import snapshotData from "../data/generated/catalog.json";

export type Locale = "en" | "ru";
export type LinkKind = "website" | "documentation" | "source" | "release";

export interface CatalogLink {
  kind: LinkKind;
  url: string;
}

export interface Version {
  value: string;
  source: "release" | "tag" | "manifest";
  url: string;
  publishedAt?: string;
}

export interface Readme {
  html: string;
  sourceUrl: string;
  sha: string;
}

export interface Product {
  slug: string;
  primaryRepo: string;
  repositories: string[];
  kind: string;
  domain: string;
  featured: boolean;
  maintainedFork: boolean;
  upstream?: string;
  summary: { en: string; ru: string };
  install: string[];
  links: CatalogLink[];
  version?: Version;
  readme?: Readme;
}

export interface Repository {
  nameWithOwner: string;
  name: string;
  description?: string;
  url: string;
  visibility: "public";
  fork: boolean;
  parent?: string;
  hasPages: boolean;
  homepage?: string;
  language?: string;
  license?: string;
  topics?: string[];
  defaultBranch: string;
  headSha?: string;
  updatedAt: string;
  pushedAt: string;
  archived: boolean;
  stars: number;
  role: string;
  links: CatalogLink[];
  version?: Version;
  readme?: Readme;
}

export interface Catalog {
  schemaVersion: number;
  owner: string;
  syncedAt: string;
  products: Product[];
  repositories: Repository[];
}

export interface LocalizedProduct extends Omit<Product, "summary"> {
  name: string;
  summary: string;
  summaries: Product["summary"];
  languages: string[];
}

const snapshot = snapshotData as unknown as Catalog;

if (snapshot.schemaVersion !== 1) {
  throw new Error(`Unsupported catalog schema: ${snapshot.schemaVersion}`);
}

const repositoryByName = new Map(snapshot.repositories.map((repository) => [repository.nameWithOwner, repository]));

function clone<T>(value: T): T {
  return structuredClone(value);
}

function localize(product: Product, locale: Locale): LocalizedProduct {
  const languages = product.repositories
    .map((name) => repositoryByName.get(name)?.language)
    .filter((language): language is string => Boolean(language));

  return {
    ...clone(product),
    name: product.primaryRepo.split("/").at(-1) ?? product.slug,
    summary: product.summary[locale],
    summaries: clone(product.summary),
    languages: [...new Set(languages)],
  };
}

export function getCatalog(): Catalog {
  return clone(snapshot);
}

export function getProducts(locale: Locale): LocalizedProduct[] {
  return snapshot.products.map((product) => localize(product, locale));
}

export function getProduct(slug: string, locale: Locale): LocalizedProduct {
  const product = snapshot.products.find((candidate) => candidate.slug === slug);
  if (!product) {
    throw new Error(`Unknown product: ${slug}`);
  }
  return localize(product, locale);
}

export function getRepository(nameWithOwner: string): Repository {
  const repository = repositoryByName.get(nameWithOwner);
  if (!repository) {
    throw new Error(`Unknown repository: ${nameWithOwner}`);
  }
  return clone(repository);
}
