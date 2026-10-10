// Public identity verified using github_get_profile and the approved profile README.
// README source: rekurt/rekurt@599fe1898b5d2223369b18cd64e20c2f28f049b9.
export const author = {
  profileUrl: "https://github.com/rekurt",
  avatarUrl: "https://avatars.githubusercontent.com/u/13642481?v=4",
} as const;

export const selectedProjectSlugs = ["openkline", "depth", "dbdiff", "gitlab-downloader", "ymsdk", "prt"] as const;

export const projectDisplayNames: Readonly<Record<string,string>> = {
  openkline: "OpenKline", depth: "Depth", dbdiff: "dbdiff", "gitlab-downloader": "GitLab Dump", ymsdk: "ymsdk", prt: "prt",
};
