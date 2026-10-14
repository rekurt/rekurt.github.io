# Branch cleanup after merge, squash and rebase with git-barber

A merged pull request does not always leave a feature branch that Git considers merged. Squash merges and rebases can preserve the changes while replacing the commits. That is where branch cleanup becomes a review task rather than a simple ancestry query.

[git-barber](https://github.com/rekurt/git-barber) provides a CLI and terminal interface for reviewing cleanup candidates. This note examines version 0.3.0, released on September 3, 2026. The original scenario check ran on September 18; it was repeated on October 15 before publication.

## Why ancestry alone misses some finished branches

An ordinary merge keeps the feature commits reachable from the target branch. A squash merge instead records their combined changes as a new commit. After a rebase, the integrated commits can have different identifiers from the original feature branch.

That distinction matters when using a familiar command:

```sh
git branch --merged main
```

Its ancestry-based answer is useful, but it does not answer every question about whether a branch's work has landed. Nor does a disappeared remote branch answer that question: an upstream can be deleted while local work remains unmerged.

## Read the candidates first

Install the version discussed here, then inspect the listing inside a repository:

```sh
cargo install git-barber --version 0.3.0 --locked
git-barber --list
git-barber --json
```

Use the repository's documentation to choose the base branch and protection patterns for your own workflow. These listing commands are the starting point; an unattended deletion command should not be the first experiment in a working repository.

## What the disposable repository showed

The test created a new repository and a local bare remote. It kept the original feature refs after three different integrations:

| Scenario | Integration performed | Reported kind |
| --- | --- | --- |
| Ordinary merge | A merge commit on main | `merged` |
| Squash | Two feature commits combined into one main commit | `squash` |
| Rebase | An integration branch rebased onto a diverged main, then fast-forwarded | `rebase` |

The rebase case used an actual rebase rather than manually copying a commit. Keeping the original feature ref let the check exercise the distinction between the old and integrated histories.

The fixture also contained an unmerged feature branch, a protected `release/*` branch, a branch checked out in a separate worktree, and an unmerged branch whose upstream had disappeared.

Inside this disposable fixture, `--yes` removed the three merge, squash and rebase candidates. Main, the unmerged branch, the protected branch, the worktree branch and the gone-upstream branch remained.

## The important exception: gone does not mean integrated

A separate step enabled `--include-gone` together with `--yes`. That step deleted the fixture's unmerged gone-upstream branch. The result is a useful boundary: missing upstream is a cleanup signal, not proof that the changes reached main.

The interactive interface offers a confirmation step, but the CLI can explicitly skip it. A description such as “deletion always asks for confirmation” would therefore be wrong.

For real work, review the base, candidates and protections before choosing a deletion action. If you need to retain work, keep a suitable ref or backup independently of the cleanup tool.

## What this establishes

The scenarios demonstrate classification and selection in a small, controlled Git history on macOS arm64. They do not prove correctness for every conflict resolution, patch transformation, repository configuration or remote provider. No branch on an external hosting service was deleted in this test.

The practical benefit is a richer review of branches whose work may have landed despite changed commit identities. The final decision still depends on the repository and the work that must be retained.

[Source and documentation](https://github.com/rekurt/git-barber) · [v0.3.0 release](https://github.com/rekurt/git-barber/releases/tag/v0.3.0)

Disclosure: this article was drafted with AI assistance. The disposable-repository scenarios were executed locally; the scope above describes those checks.
