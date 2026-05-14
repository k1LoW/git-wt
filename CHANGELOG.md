# Changelog

## [v0.28.0](https://github.com/k1LoW/git-wt/compare/v0.27.0...v0.28.0) - 2026-05-14
### Breaking Changes 🛠
- perf: batch branch commit-message lookups in shell completion by @tksm in https://github.com/k1LoW/git-wt/pull/187
### Other Changes
- chore(deps): bump the dependencies group across 1 directory with 4 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/181
- chore(deps): bump github.com/go-git/go-billy/v5 from 5.8.0 to 5.9.0 by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/188
- chore(deps): bump the dependencies group across 1 directory with 2 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/189

## [v0.27.0](https://github.com/k1LoW/git-wt/compare/v0.26.2...v0.27.0) - 2026-04-17
### Breaking Changes 🛠
- fix: error when start-point is specified for existing branch or worktree by @k1LoW in https://github.com/k1LoW/git-wt/pull/179
### New Features 🎉
- feat: add `-b`/`--branch` flag to specify branch name independently from worktree directory name by @glasser in https://github.com/k1LoW/git-wt/pull/176
### Other Changes
- chore(deps): bump golang.org/x/sys from 0.42.0 to 0.43.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/173
- chore(deps): bump Songmu/tagpr from 1.17.1 to 1.18.1 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/174
- chore(deps): bump github.com/go-git/go-git/v5 from 5.17.2 to 5.18.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/178
- chore(deps): bump Songmu/tagpr from 1.18.1 to 1.18.2 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/177

## [v0.26.2](https://github.com/k1LoW/git-wt/compare/v0.26.1...v0.26.2) - 2026-04-02
### Other Changes
- chore(deps): bump github.com/go-git/go-git/v5 from 5.17.1 to 5.17.2 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/167
- chore(deps): bump actions/setup-go from 6.3.0 to 6.4.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/168

## [v0.26.1](https://github.com/k1LoW/git-wt/compare/v0.26.0...v0.26.1) - 2026-03-31
### Other Changes
- chore(deps): bump github.com/go-git/go-git/v5 from 5.17.0 to 5.17.1 by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/165

## [v0.26.0](https://github.com/k1LoW/git-wt/compare/v0.25.0...v0.26.0) - 2026-03-28
### New Features 🎉
- feat: add `wt.symlink` config and `--symlink` flag by @ryoppippi in https://github.com/k1LoW/git-wt/pull/164
### Other Changes
- refactor: extract test helpers in e2e/bare_test.go by @usadamasa in https://github.com/k1LoW/git-wt/pull/153
- chore(deps): bump github.com/go-git/go-git/v5 from 5.16.5 to 5.17.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/156
- chore(deps): bump the dependencies group with 4 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/157
- chore(deps): bump aquasecurity/trivy-action from 0.34.1 to 0.34.2 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/158
- chore(deps): bump the dependencies group with 2 updates by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/159
- chore(deps): bump aquasecurity/trivy-action from 0.34.2 to 0.35.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/160
- chore(deps): bump reviewdog/action-golangci-lint from 2.8.0 to 2.10.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/161

## [v0.25.0](https://github.com/k1LoW/git-wt/compare/v0.24.0...v0.25.0) - 2026-02-19
### New Features 🎉
- feat: support delete operation in bare repositories by @usadamasa in https://github.com/k1LoW/git-wt/pull/150
### Other Changes
- chore(deps): bump Songmu/tagpr from 1.15.0 to 1.17.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/151

## [v0.24.0](https://github.com/k1LoW/git-wt/compare/v0.23.0...v0.24.0) - 2026-02-17
### New Features 🎉
- feat: support add/switch operations in bare repositories by @usadamasa in https://github.com/k1LoW/git-wt/pull/147
### Fix bug 🐛
- fix: use custom version template to output "git-wt version" instead of "git version" by @k1LoW in https://github.com/k1LoW/git-wt/pull/148

## [v0.23.0](https://github.com/k1LoW/git-wt/compare/v0.22.0...v0.23.0) - 2026-02-16
### New Features 🎉
- feat: add `wt.remover` config and `--remover` flag by @ryoppippi in https://github.com/k1LoW/git-wt/pull/144
### Other Changes
- fix: improve error handling in file copy, init output, and worktree lookup by @k1LoW in https://github.com/k1LoW/git-wt/pull/146

## [v0.22.0](https://github.com/k1LoW/git-wt/compare/v0.21.1...v0.22.0) - 2026-02-16
### New Features 🎉
- feat: support list command in bare repositories by @usadamasa in https://github.com/k1LoW/git-wt/pull/140
- feat: add `wt.deletehook` config and `--deletehook` flag by @ryoppippi in https://github.com/k1LoW/git-wt/pull/132
### Other Changes
- fix: print warning to stderr when file copy fails during worktree creation by @k1LoW in https://github.com/k1LoW/git-wt/pull/142

## [v0.21.1](https://github.com/k1LoW/git-wt/compare/v0.21.0...v0.21.1) - 2026-02-14
### Fix bug 🐛
- fix: run git from mainRoot only if we deleted the current worktree by @yoichi in https://github.com/k1LoW/git-wt/pull/137

## [v0.21.0](https://github.com/k1LoW/git-wt/compare/v0.20.0...v0.21.0) - 2026-02-14
### New Features 🎉
- feat: detect bare repositories and model 4-state RepoContext by @usadamasa in https://github.com/k1LoW/git-wt/pull/136
### Other Changes
- chore(deps): bump github.com/go-git/go-git/v5 from 5.16.4 to 5.16.5 by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/126
- Fix test isolation from user's global git config by @usadamasa in https://github.com/k1LoW/git-wt/pull/129
- chore(deps): bump golang.org/x/sys from 0.40.0 to 0.41.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/134
- chore(deps): bump aquasecurity/trivy-action from 0.33.1 to 0.34.0 in the dependencies group by @dependabot[bot] in https://github.com/k1LoW/git-wt/pull/135

## [v0.20.0](https://github.com/k1LoW/git-wt/compare/v0.19.1...v0.20.0) - 2026-02-09
### New Features 🎉
- feat: add `--json` flag to output worktree list in JSON format by @k1LoW in https://github.com/k1LoW/git-wt/pull/124

## [v0.19.1](https://github.com/k1LoW/git-wt/compare/v0.19.0...v0.19.1) - 2026-02-08
### Other Changes
- Fix fzf recipe to handle header row and current worktree marker by @dgrant in https://github.com/k1LoW/git-wt/pull/118

## [v0.19.0](https://github.com/k1LoW/git-wt/compare/v0.18.1...v0.19.0) - 2026-02-06
### New Features 🎉
- feat: add `wt.relative` config and `--relative` flag to append subdirectory to worktree path by @k1LoW in https://github.com/k1LoW/git-wt/pull/113
- feat: support filesystem paths for switching and deleting worktrees by @dgrant in https://github.com/k1LoW/git-wt/pull/105

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
