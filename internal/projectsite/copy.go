package projectsite

type siteCopy struct {
	Family         string
	AllProjects    string
	ProjectHome    string
	Source         string
	Documentation  string
	Release        string
	Install        string
	Overview       string
	Version        string
	License        string
	Language       string
	Updated        string
	Current        string
	Unversioned    string
	NotDeclared    string
	Copy           string
	Copied         string
	RepositoryDocs string
	FamilyTitle    string
	FamilyIntro    string
	AuthorHub      string
	BackToProject  string
	SkipToContent  string
	GeneratedFrom  string
	PrimaryNav     string
	LanguageNav    string
	SiblingNav     string
	OpenProject    string
}

var copies = map[string]siteCopy{
	"en": {
		Family: "rekurt / systems", AllProjects: "Project family", ProjectHome: "Project home",
		Source: "Source", Documentation: "Documentation", Release: "Release", Install: "Install",
		Overview: "Repository documentation", Version: "Version", License: "License", Language: "Language",
		Updated: "Last update", Current: "Current", Unversioned: "Unversioned", NotDeclared: "Not declared",
		Copy: "Copy", Copied: "Copied", RepositoryDocs: "Authoritative documentation from the repository",
		FamilyTitle: "Built as one open-source system.", FamilyIntro: "Explore the other public projects by the same author, with verified source and release links.",
		AuthorHub: "Author portfolio", BackToProject: "Back to project", SkipToContent: "Skip to content",
		GeneratedFrom: "Generated from public GitHub metadata", PrimaryNav: "Primary navigation",
		LanguageNav: "Language", SiblingNav: "Related projects", OpenProject: "Open project",
	},
	"ru": {
		Family: "rekurt / systems", AllProjects: "Семейство проектов", ProjectHome: "Главная проекта",
		Source: "Исходники", Documentation: "Документация", Release: "Релиз", Install: "Установка",
		Overview: "Документация репозитория", Version: "Версия", License: "Лицензия", Language: "Язык",
		Updated: "Последнее обновление", Current: "Текущий", Unversioned: "Без версии", NotDeclared: "Не указано",
		Copy: "Копировать", Copied: "Скопировано", RepositoryDocs: "Авторитетная документация из репозитория",
		FamilyTitle: "Часть единой open-source системы.", FamilyIntro: "Другие публичные проекты автора с проверенными ссылками на исходники и релизы.",
		AuthorHub: "Портфолио автора", BackToProject: "Назад к проекту", SkipToContent: "Перейти к содержимому",
		GeneratedFrom: "Собрано из публичных данных GitHub", PrimaryNav: "Основная навигация",
		LanguageNav: "Язык", SiblingNav: "Связанные проекты", OpenProject: "Открыть проект",
	},
	"zh-cn": {
		Family: "rekurt / systems", AllProjects: "项目系列", ProjectHome: "项目主页",
		Source: "源代码", Documentation: "文档", Release: "版本", Install: "安装",
		Overview: "仓库文档", Version: "版本", License: "许可证", Language: "语言",
		Updated: "最近更新", Current: "当前", Unversioned: "未版本化", NotDeclared: "未声明",
		Copy: "复制", Copied: "已复制", RepositoryDocs: "来自仓库的权威文档",
		FamilyTitle: "同一开源体系中的项目。", FamilyIntro: "探索同一作者的其他公开项目，并访问已核实的源代码与版本链接。",
		AuthorHub: "作者作品集", BackToProject: "返回项目", SkipToContent: "跳到内容",
		GeneratedFrom: "根据公开 GitHub 元数据生成", PrimaryNav: "主导航",
		LanguageNav: "语言", SiblingNav: "相关项目", OpenProject: "打开项目",
	},
}

func copyFor(locale string) siteCopy {
	return copies[locale]
}
