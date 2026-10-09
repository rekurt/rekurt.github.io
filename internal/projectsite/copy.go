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
		Family: "rekurt / systems", AllProjects: "All projects", ProjectHome: "Project home",
		Source: "Source", Documentation: "Documentation", Release: "Release", Install: "Install",
		Overview: "Documentation", Version: "Version", License: "License", Language: "Language",
		Updated: "Last update", Current: "Current", Unversioned: "Unversioned", NotDeclared: "Not declared",
		Copy: "Copy", Copied: "Copied", RepositoryDocs: "Setup, examples and API reference.",
		FamilyTitle: "More projects by rekurt", FamilyIntro: "Libraries, command-line tools and applications.",
		AuthorHub: "Author portfolio", BackToProject: "Back to project", SkipToContent: "Skip to content",
		GeneratedFrom: "Generated from public GitHub metadata", PrimaryNav: "Primary navigation",
		LanguageNav: "Language", SiblingNav: "Related projects", OpenProject: "Open project",
	},
	"ru": {
		Family: "rekurt / systems", AllProjects: "Все проекты", ProjectHome: "Главная проекта",
		Source: "Исходники", Documentation: "Документация", Release: "Релиз", Install: "Установка",
		Overview: "Документация", Version: "Версия", License: "Лицензия", Language: "Язык",
		Updated: "Последнее обновление", Current: "Текущий", Unversioned: "Без версии", NotDeclared: "Не указано",
		Copy: "Копировать", Copied: "Скопировано", RepositoryDocs: "Настройка, примеры использования и справочник API.",
		FamilyTitle: "Другие проекты rekurt", FamilyIntro: "Библиотеки, консольные инструменты и приложения.",
		AuthorHub: "Портфолио автора", BackToProject: "Назад к проекту", SkipToContent: "Перейти к содержимому",
		GeneratedFrom: "Собрано из публичных данных GitHub", PrimaryNav: "Основная навигация",
		LanguageNav: "Язык", SiblingNav: "Связанные проекты", OpenProject: "Открыть проект",
	},
	"zh-cn": {
		Family: "rekurt / systems", AllProjects: "全部项目", ProjectHome: "项目主页",
		Source: "源代码", Documentation: "文档", Release: "版本", Install: "安装",
		Overview: "仓库文档", Version: "版本", License: "许可证", Language: "语言",
		Updated: "最近更新", Current: "当前", Unversioned: "未版本化", NotDeclared: "未声明",
		Copy: "复制", Copied: "已复制", RepositoryDocs: "配置说明、使用示例和 API 参考。",
		FamilyTitle: "rekurt 的其他项目", FamilyIntro: "程序库、命令行工具和应用程序。",
		AuthorHub: "作者作品集", BackToProject: "返回项目", SkipToContent: "跳到内容",
		GeneratedFrom: "根据公开 GitHub 元数据生成", PrimaryNav: "主导航",
		LanguageNav: "语言", SiblingNav: "相关项目", OpenProject: "打开项目",
	},
}

func copyFor(locale string) siteCopy {
	return copies[locale]
}
