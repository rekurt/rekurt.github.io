import type { Locale } from "../lib/catalog";

const en = {
  siteName: "rekurt / systems",
  navProjects: "Projects",
  navRegistry: "Repository registry",
  navAbout: "About",
  switchLanguage: "Русский",
  openMenu: "Open navigation",
  closeMenu: "Close navigation",
  linkWebsite: "Website",
  linkDocumentation: "Documentation",
  linkSource: "Source",
  linkRelease: "Release",
  version: "Version",
  unversioned: "Unversioned",
  featured: "Featured",
  maintainedFork: "Maintained fork",
  updated: "Updated",
  repositories: "Repositories",
  viewProject: "View project",
  footerSource: "Portfolio source",
  footerStatus: "Generated from public GitHub data",
  skipToContent: "Skip to content",
  copied: "Copied",
  copyCommand: "Copy install command",
} as const;

export type CopyKey = keyof typeof en;

const ru = {
  siteName: "rekurt / systems",
  navProjects: "Проекты",
  navRegistry: "Реестр репозиториев",
  navAbout: "Об авторе",
  switchLanguage: "English",
  openMenu: "Открыть навигацию",
  closeMenu: "Закрыть навигацию",
  linkWebsite: "Сайт",
  linkDocumentation: "Документация",
  linkSource: "Исходники",
  linkRelease: "Релиз",
  version: "Версия",
  unversioned: "Без версии",
  featured: "Избранное",
  maintainedFork: "Поддерживаемый форк",
  updated: "Обновлено",
  repositories: "Репозитории",
  viewProject: "Открыть проект",
  footerSource: "Исходники портфолио",
  footerStatus: "Собрано из публичных данных GitHub",
  skipToContent: "Перейти к содержимому",
  copied: "Скопировано",
  copyCommand: "Скопировать команду установки",
} satisfies Record<CopyKey, string>;

export const copy = { en, ru } as const;

export type Copy = (typeof copy)[Locale];

export function alternatePath(path: string, locale: Locale): string {
  const withoutRussianPrefix = path.replace(/^\/ru(?=\/|$)/, "") || "/";
  return locale === "ru" ? `/ru${withoutRussianPrefix}`.replace(/\/+/g, "/") : withoutRussianPrefix;
}
