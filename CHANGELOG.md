# Changelog

## [v0.18.1](https://github.com/k1LoW/git-wt/compare/v0.18.0...v0.18.1) - 2026-02-05
### Fix bug 🐛
- fix: prevent file completions in fish shell for git wt by @dgrant in https://github.com/k1LoW/git-wt/pull/109
- fix: always error on legacy basedir migration instead of prompting by @k1LoW in https://github.com/k1LoW/git-wt/pull/112
### Other Changes
- chore(deps): bump Songmu/tagpr from 1.14.0 to 1.15.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/111

## [v0.18.0](https://github.com/k1LoW/git-wt/compare/v0.17.0...v0.18.0) - 2026-02-04
### New Features 🎉
- fix: preserve file timestamps when copying files by @dgrant in https://github.com/k1LoW/git-wt/pull/98
- fix: check for untracked/modified files before worktree deletion by @dgrant in https://github.com/k1LoW/git-wt/pull/102
- fix: --copy patterns now match untracked files by @dgrant in https://github.com/k1LoW/git-wt/pull/100

## [v0.17.0](https://github.com/k1LoW/git-wt/compare/v0.16.1...v0.17.0) - 2026-02-01
### Breaking Changes 🛠
- feat: change default wt.basedir from `../{gitroot}-wt` to `.wt` by @k1LoW in https://github.com/k1LoW/git-wt/pull/95
### Other Changes
- chore(deps): bump Songmu/tagpr from 1.12.1 to 1.14.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/94

## [v0.16.1](https://github.com/k1LoW/git-wt/compare/v0.16.0...v0.16.1) - 2026-01-27
### Breaking Changes 🛠
- fix: exclude `basedir` from file copying to prevent circular copying by @k1LoW in https://github.com/k1LoW/git-wt/pull/92

## [v0.16.0](https://github.com/k1LoW/git-wt/compare/v0.15.1...v0.16.0) - 2026-01-27
### New Features 🎉
- perf: optimize file copy with clonefile on macOS by @k1LoW in https://github.com/k1LoW/git-wt/pull/90

## [v0.15.1](https://github.com/k1LoW/git-wt/compare/v0.15.0...v0.15.1) - 2026-01-26
### Fix bug 🐛
- fix: disable Cobra's default completion subcommand to avoid confusion by @k1LoW in https://github.com/k1LoW/git-wt/pull/89
### Other Changes
- chore(deps): bump github.com/olekukonko/tablewriter from 1.1.2 to 1.1.3 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/84
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/85
- add fzf information for interactive selectn in README.md by @shishi in https://github.com/k1LoW/git-wt/pull/88

## [v0.15.0](https://github.com/k1LoW/git-wt/compare/v0.14.2...v0.15.0) - 2026-01-21
### New Features 🎉
- Add .gitignore and README.md to basedir on worktree creation by @Songmu in https://github.com/k1LoW/git-wt/pull/81

## [v0.14.2](https://github.com/k1LoW/git-wt/compare/v0.14.1...v0.14.2) - 2026-01-20
### Fix bug 🐛
- fix: move default branch check before worktree removal by @k1LoW in https://github.com/k1LoW/git-wt/pull/78

## [v0.14.2](https://github.com/k1LoW/git-wt/compare/v0.14.1...v0.14.2) - 2026-01-20
### Fix bug 🐛
- fix: move default branch check before worktree removal by @k1LoW in https://github.com/k1LoW/git-wt/pull/78

## [v0.14.1](https://github.com/k1LoW/git-wt/compare/v0.14.0...v0.14.1) - 2026-01-20
### Other Changes
- fix: allow worktree deletion for default branch while protecting the branch itself by @k1LoW in https://github.com/k1LoW/git-wt/pull/76

## [v0.14.0](https://github.com/k1LoW/git-wt/compare/v0.13.1...v0.14.0) - 2026-01-20
### Breaking Changes 🛠
- fix: improve default branch detection by @k1LoW in https://github.com/k1LoW/git-wt/pull/74
### New Features 🎉
- feat: protect default branch from accidental deletion by @k1LoW in https://github.com/k1LoW/git-wt/pull/75
### Other Changes
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/72

## [v0.13.1](https://github.com/k1LoW/git-wt/compare/v0.13.0...v0.13.1) - 2026-01-14
### Fix bug 🐛
- fix: MainRepoRoot to return proper repo root by @jackchuka in https://github.com/k1LoW/git-wt/pull/70

## [v0.13.0](https://github.com/k1LoW/git-wt/compare/v0.12.1...v0.13.0) - 2026-01-12
### Breaking Changes 🛠
- fix: use an invalid branch name as the detached HEAD marker by @yoichi in https://github.com/k1LoW/git-wt/pull/68
### Other Changes
- chore(deps): bump Songmu/tagpr from 1.10.0 to 1.11.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/66

## [v0.12.1](https://github.com/k1LoW/git-wt/compare/v0.12.0...v0.12.1) - 2026-01-08
### Fix bug 🐛
- fix: invoke cd regardless of branch deletion result by @yoichi in https://github.com/k1LoW/git-wt/pull/64

## [v0.12.0](https://github.com/k1LoW/git-wt/compare/v0.11.0...v0.12.0) - 2026-01-07
### New Features 🎉
- feat: add `wt.nocd=create` option to allow cd for existing worktrees by @k1LoW in https://github.com/k1LoW/git-wt/pull/62

## [v0.11.0](https://github.com/k1LoW/git-wt/compare/v0.10.0...v0.11.0) - 2026-01-06
### New Features 🎉
- feat: safely delete current worktree and return to repository root by @k1LoW in https://github.com/k1LoW/git-wt/pull/59

## [v0.11.0](https://github.com/k1LoW/git-wt/compare/v0.10.0...v0.11.0) - 2026-01-06
### New Features 🎉
- feat: safely delete current worktree and return to repository root by @k1LoW in https://github.com/k1LoW/git-wt/pull/59

## [v0.10.0](https://github.com/k1LoW/git-wt/compare/v0.9.1...v0.10.0) - 2026-01-06
### Breaking Changes 🛠
- feat: support start-point argument for worktree creation by @k1LoW in https://github.com/k1LoW/git-wt/pull/55
### New Features 🎉
- feat: add wt.copy to always copy specific gitignored files by @k1LoW in https://github.com/k1LoW/git-wt/pull/53
- feat: allow deleting branches without associated worktrees by @k1LoW in https://github.com/k1LoW/git-wt/pull/54
### Fix bug 🐛
- fix: enable start-point completion by passing all args to __complete by @k1LoW in https://github.com/k1LoW/git-wt/pull/57

## [v0.9.1](https://github.com/k1LoW/git-wt/compare/v0.9.0...v0.9.1) - 2026-01-06
### Fix bug 🐛
- fix: use string match instead of test in fish shell hook to avoid errors with short options by @k1LoW in https://github.com/k1LoW/git-wt/pull/50

## [v0.9.0](https://github.com/k1LoW/git-wt/compare/v0.8.2...v0.9.0) - 2026-01-05
### New Features 🎉
- feat: support creating and deleting multiple worktrees at once by @k1LoW in https://github.com/k1LoW/git-wt/pull/47

## [v0.8.2](https://github.com/k1LoW/git-wt/compare/v0.8.1...v0.8.2) - 2026-01-05
### Fix bug 🐛
- fix: shell wrappers now respect `wt.nocd` config setting by @k1LoW in https://github.com/k1LoW/git-wt/pull/43
### Other Changes
- docs: add recipes for peco and tmux in README by @k1LoW in https://github.com/k1LoW/git-wt/pull/45

## [v0.8.1](https://github.com/k1LoW/git-wt/compare/v0.8.0...v0.8.1) - 2026-01-05
### Other Changes
- fix: use `[worktree: branch=X]` format when worktree dir matches branch name by @k1LoW in https://github.com/k1LoW/git-wt/pull/41

## [v0.8.0](https://github.com/k1LoW/git-wt/compare/v0.7.0...v0.8.0) - 2026-01-05
### Breaking Changes 🛠
- fix: rename `--no-switch-directory` flag to `--nocd` by @k1LoW in https://github.com/k1LoW/git-wt/pull/37
### New Features 🎉
- feat: add `wt.nocd` config option to disable directory switching by default by @k1LoW in https://github.com/k1LoW/git-wt/pull/39
### Other Changes
- chore: improve delete message when worktree dir differs from branch name by @k1LoW in https://github.com/k1LoW/git-wt/pull/40

## [v0.7.0](https://github.com/k1LoW/git-wt/compare/v0.6.0...v0.7.0) - 2026-01-05
### Breaking Changes 🛠
- fix: improve list layout by @k1LoW in https://github.com/k1LoW/git-wt/pull/33
### New Features 🎉
- fix: enable flag completion in shell integration by @k1LoW in https://github.com/k1LoW/git-wt/pull/35
- feat: add descriptions to branch/worktree completion by @k1LoW in https://github.com/k1LoW/git-wt/pull/36

## [v0.6.0](https://github.com/k1LoW/git-wt/compare/v0.5.2...v0.6.0) - 2026-01-04
### New Features 🎉
- feat: add `wt.hooks` config and `--hook` flag to run commands after creating new worktree by @k1LoW in https://github.com/k1LoW/git-wt/pull/29
- fix: rename config key `wt.hooks` to `wt.hook` by @k1LoW in https://github.com/k1LoW/git-wt/pull/32

## [v0.5.2](https://github.com/k1LoW/git-wt/compare/v0.5.1...v0.5.2) - 2026-01-03
### New Features 🎉
- feat: allow `--no-switch-directory` with `git wt <branch>` by @k1LoW in https://github.com/k1LoW/git-wt/pull/27
### Other Changes
- chore(deps): bump the dependencies group with 4 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/25
- chore: disable gpg signing for tests by @k1LoW in https://github.com/k1LoW/git-wt/pull/28

## [v0.5.1](https://github.com/k1LoW/git-wt/compare/v0.5.0...v0.5.1) - 2025-12-28
### Other Changes
- refactor: use k1LoW/exec instead by @k1LoW in https://github.com/k1LoW/git-wt/pull/24

## [v0.5.0](https://github.com/k1LoW/git-wt/compare/v0.4.0...v0.5.0) - 2025-12-28
### New Features 🎉
- feat: add `wt.nocopy` config and `--nocopy` flag to exclude files from copying by @k1LoW in https://github.com/k1LoW/git-wt/pull/20

## [v0.4.0](https://github.com/k1LoW/git-wt/compare/v0.3.0...v0.4.0) - 2025-12-28
### New Features 🎉
- feat: add config override flags (`--basedir`, `--copyignored`, `--copyuntracked`, `--copymodified`) by @k1LoW in https://github.com/k1LoW/git-wt/pull/19
### Fix bug 🐛
- fix: ensure worktrees use correct base directory when created from another worktree by @tnagatomi in https://github.com/k1LoW/git-wt/pull/18

## [v0.3.0](https://github.com/k1LoW/git-wt/compare/v0.2.3...v0.3.0) - 2025-12-27
### Breaking Changes 🛠
- fix: set gostyle by @k1LoW in https://github.com/k1LoW/git-wt/pull/16
### Fix bug 🐛
- Fix git-wt fish hook output formatting by @osamu2001 in https://github.com/k1LoW/git-wt/pull/14
### Other Changes
- feat: add context.Context support to all git operations by @k1LoW in https://github.com/k1LoW/git-wt/pull/15

## [v0.2.3](https://github.com/k1LoW/git-wt/compare/v0.2.2...v0.2.3) - 2025-12-26
### Fix bug 🐛
- fix: enhance worktree handling with directory name support by @k1LoW in https://github.com/k1LoW/git-wt/pull/11

## [v0.2.2](https://github.com/k1LoW/git-wt/compare/v0.2.1...v0.2.2) - 2025-12-26
### Fix bug 🐛
- fix: support multiple arguments for `git wt` wrappers by @k1LoW in https://github.com/k1LoW/git-wt/pull/8

## [v0.2.2](https://github.com/k1LoW/git-wt/compare/v0.2.1...v0.2.2) - 2025-12-26
### Fix bug 🐛
- fix: support multiple arguments for `git wt` wrappers by @k1LoW in https://github.com/k1LoW/git-wt/pull/8

## [v0.2.1](https://github.com/k1LoW/git-wt/compare/v0.2.0...v0.2.1) - 2025-12-26
### Breaking Changes 🛠
- fix: rename `ignore-switch-directory` flag to `no-switch-directory` by @k1LoW in https://github.com/k1LoW/git-wt/pull/7

## [v0.2.0](https://github.com/k1LoW/git-wt/compare/v0.1.0...v0.2.0) - 2025-12-26
### New Features 🎉
- feat: add support for copying files to new worktrees by @k1LoW in https://github.com/k1LoW/git-wt/pull/3
### Other Changes
- fix: add `--ignore-switch-directory` option for shell initialization by @k1LoW in https://github.com/k1LoW/git-wt/pull/5

## [v0.1.0](https://github.com/k1LoW/git-wt/commits/v0.1.0) - 2025-12-26
