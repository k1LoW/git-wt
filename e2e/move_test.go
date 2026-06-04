// move_test.go contains worktree rename tests:
//   - TestE2E_MoveWorktree: -m/-M renaming (basic, force, conflicts, default branch, slash branches)
//   - TestE2E_MoveCurrentWorktree: renaming the worktree we are currently inside
package e2e

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
	"github.com/k1LoW/git-wt/testutil"
)

func TestE2E_MoveWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("two_arg_rename", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "old-name")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		oldPath := worktreePath(out)

		out, err = runGitWt(t, binPath, repo.Root, "-m", "old-name", "new-name")
		if err != nil {
			t.Fatalf("git-wt -m failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("old worktree directory should have been moved: %s", oldPath)
		}
		newPath := filepath.Join(filepath.Dir(oldPath), "new-name")
		if _, err := os.Stat(newPath); err != nil {
			t.Errorf("new worktree directory should exist at %s: %v", newPath, err)
		}

		cmd := exec.Command("git", "branch", "--list", "new-name")
		cmd.Dir = repo.Root
		branchOut, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if !strings.Contains(string(branchOut), "new-name") {
			t.Errorf("new-name branch should exist, got: %s", branchOut)
		}
		cmd = exec.Command("git", "branch", "--list", "old-name")
		cmd.Dir = repo.Root
		branchOut, err = cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if strings.Contains(string(branchOut), "old-name") {
			t.Errorf("old-name branch should have been renamed, still found: %s", branchOut)
		}
	})

	t.Run("one_arg_renames_current", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "wt-cur")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		oldPath := worktreePath(out)

		out, err = runGitWt(t, binPath, oldPath, "-m", "wt-renamed")
		if err != nil {
			t.Fatalf("git-wt -m failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("old worktree directory should have been moved")
		}
		newPath := filepath.Join(filepath.Dir(oldPath), "wt-renamed")
		if _, err := os.Stat(newPath); err != nil {
			t.Errorf("renamed worktree directory should exist: %v", err)
		}
	})

	t.Run("one_arg_fails_outside_worktree", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "-m", "something")
		if err == nil {
			t.Fatalf("should fail when run from main repo, got: %s", out)
		}
		if !strings.Contains(out, "not a linked worktree") {
			t.Errorf("error should mention 'not a linked worktree', got: %s", out)
		}
	})

	t.Run("target_dir_exists_blocks_even_force", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		if _, err := runGitWt(t, binPath, repo.Root, "src"); err != nil {
			t.Fatalf("failed to create worktree src: %v", err)
		}
		if _, err := runGitWt(t, binPath, repo.Root, "dst"); err != nil {
			t.Fatalf("failed to create worktree dst: %v", err)
		}

		// Safe rename refuses target dir collisions.
		out, err := runGitWt(t, binPath, repo.Root, "-m", "src", "dst")
		if err == nil {
			t.Fatalf("rename should fail when target dir exists, got: %s", out)
		}
		if !strings.Contains(out, "already exists") {
			t.Errorf("error should mention 'already exists', got: %s", out)
		}

		// -M does NOT change target-dir behavior because `git worktree move
		// --force` does not overwrite an existing destination directory; force
		// only relaxes dirty/locked-worktree checks.
		out, err = runGitWt(t, binPath, repo.Root, "-M", "src", "dst")
		if err == nil {
			t.Fatalf("-M rename should also fail when target dir exists, got: %s", out)
		}
		if !strings.Contains(out, "already exists") {
			t.Errorf("error should mention 'already exists' under -M too, got: %s", out)
		}
	})

	t.Run("target_branch_exists_blocks_safe", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		if _, err := runGitWt(t, binPath, repo.Root, "src-only"); err != nil {
			t.Fatalf("failed to create worktree src-only: %v", err)
		}
		// Create a branch (without a worktree) that the rename would collide with.
		cmd := exec.Command("git", "branch", "taken")
		cmd.Dir = repo.Root
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create branch taken: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "-m", "src-only", "taken")
		if err == nil {
			t.Fatalf("rename should fail when target branch exists, got: %s", out)
		}
		if !strings.Contains(out, "already exists") {
			t.Errorf("error should mention 'already exists', got: %s", out)
		}
	})

	t.Run("force_overwrites_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		if _, err := runGitWt(t, binPath, repo.Root, "src-force"); err != nil {
			t.Fatalf("failed to create worktree src-force: %v", err)
		}
		// Create a branch that would collide; -M overrides via `git branch -M`.
		cmd := exec.Command("git", "branch", "taken-force")
		cmd.Dir = repo.Root
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create branch: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "-M", "src-force", "taken-force")
		if err != nil {
			t.Fatalf("git-wt -M should succeed: %v\noutput: %s", err, out)
		}
	})

	t.Run("renames_branch_with_slash_and_cleans_parents", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "feat/slashy")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		oldPath := worktreePath(out)
		oldParent := filepath.Dir(oldPath)

		out, err = runGitWt(t, binPath, repo.Root, "-m", "feat/slashy", "flat")
		if err != nil {
			t.Fatalf("git-wt -m failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(oldParent); !os.IsNotExist(err) {
			t.Errorf("empty parent directory should have been cleaned up: %s", oldParent)
		}
	})

	t.Run("blocks_renaming_default_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Switch off main so a worktree for main can be created.
		cmd := exec.Command("git", "checkout", "-b", "side")
		cmd.Dir = repo.Root
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to switch branch: %v", err)
		}
		if _, err := runGitWt(t, binPath, repo.Root, "main"); err != nil {
			t.Fatalf("failed to create worktree for main: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "-m", "main", "trunk")
		if err == nil {
			t.Fatalf("renaming the default branch should fail without --allow-delete-default, got: %s", out)
		}
		if !strings.Contains(out, "default branch") {
			t.Errorf("error should mention default branch, got: %s", out)
		}
	})

	t.Run("allows_default_rename_with_override", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		cmd := exec.Command("git", "checkout", "-b", "side")
		cmd.Dir = repo.Root
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to switch branch: %v", err)
		}
		if _, err := runGitWt(t, binPath, repo.Root, "main"); err != nil {
			t.Fatalf("failed to create worktree for main: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "-m", "--allow-delete-default", "main", "trunk")
		if err != nil {
			t.Fatalf("rename with --allow-delete-default should succeed: %v\noutput: %s", err, out)
		}

		cmd = exec.Command("git", "branch", "--list", "trunk")
		cmd.Dir = repo.Root
		branchOut, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if !strings.Contains(string(branchOut), "trunk") {
			t.Errorf("trunk branch should exist after rename, got: %s", branchOut)
		}
	})

	t.Run("detached_head_blocked", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Add a worktree, then detach its HEAD.
		out, err := runGitWt(t, binPath, repo.Root, "to-detach")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)
		cmd := exec.Command("git", "checkout", "--detach")
		cmd.Dir = wtPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to detach HEAD: %v", err)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-m", "to-detach", "anywhere")
		if err == nil {
			t.Fatalf("rename should fail on detached HEAD, got: %s", out)
		}
		if !strings.Contains(out, "detached HEAD") {
			t.Errorf("error should mention detached HEAD, got: %s", out)
		}
	})

	t.Run("rejects_b_flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		if _, err := runGitWt(t, binPath, repo.Root, "with-b"); err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "-m", "-b", "x", "with-b", "y")
		if err == nil {
			t.Fatalf("-m with -b should fail, got: %s", out)
		}
	})

	t.Run("respects_basedir_flag_override_with_dir_query", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		altBase := filepath.Join(repo.ParentDir(), "alt-base-dirq")

		// Create a worktree with -b so the directory name (dirX) differs
		// from the branch name (branchX), under an overridden basedir.
		if _, err := runGitWt(t, binPath, repo.Root, "--basedir", altBase, "-b", "branchX", "dirX"); err != nil {
			t.Fatalf("failed to create worktree with -b under alt basedir: %v", err)
		}

		// Rename by *directory* name (not branch name) while keeping the
		// basedir override. Without the override-aware fallback, this query
		// fails because FindWorktreeByBranchOrDir's basedir-relative match
		// path uses unoverridden config.
		out, err := runGitWt(t, binPath, repo.Root, "--basedir", altBase, "-m", "dirX", "dirY")
		if err != nil {
			t.Fatalf("rename by dir name with --basedir override failed: %v\noutput: %s", err, out)
		}

		newPath := filepath.Join(altBase, "dirY")
		if _, err := os.Stat(newPath); err != nil {
			t.Errorf("renamed worktree should exist at %q: %v", newPath, err)
		}
	})

	t.Run("respects_basedir_flag_override", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		altBase := filepath.Join(repo.ParentDir(), "alt-base")

		// Create the worktree under the overridden basedir.
		out, err := runGitWt(t, binPath, repo.Root, "--basedir", altBase, "alt-src")
		if err != nil {
			t.Fatalf("failed to create worktree under alt basedir: %v\noutput: %s", err, out)
		}
		oldPath := worktreePath(out)
		expectedOld := filepath.Join(altBase, "alt-src")
		if oldPath != expectedOld {
			// Tolerate symlink resolution differences on macOS.
			resolvedOld, errOld := filepath.EvalSymlinks(oldPath)
			resolvedExpected, errExp := filepath.EvalSymlinks(expectedOld)
			if errOld != nil || errExp != nil || resolvedOld != resolvedExpected {
				t.Fatalf("worktree should live under %q, got %q", expectedOld, oldPath)
			}
		}

		// Rename it using the same basedir override. Without the override
		// being honored by -m, the rename would compute paths against the
		// default .wt basedir and fail to find the worktree or clean up the
		// wrong tree.
		out, err = runGitWt(t, binPath, repo.Root, "--basedir", altBase, "-m", "alt-src", "alt-dst")
		if err != nil {
			t.Fatalf("rename with --basedir override failed: %v\noutput: %s", err, out)
		}

		newPath := filepath.Join(altBase, "alt-dst")
		if _, err := os.Stat(newPath); err != nil {
			t.Errorf("renamed worktree should exist at %q: %v", newPath, err)
		}
	})

	t.Run("branch_only_rename_when_dir_already_matches", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create a worktree where dir name (mydir) differs from branch name
		// (some-branch) via the -b flag.
		out, err := runGitWt(t, binPath, repo.Root, "-b", "some-branch", "mydir")
		if err != nil {
			t.Fatalf("failed to create worktree with -b: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Renaming to "mydir" should succeed: the directory already matches,
		// only the branch needs renaming. Before the fix this would fail with
		// "target worktree directory ... already exists".
		out, err = runGitWt(t, binPath, repo.Root, "-m", "some-branch", "mydir")
		if err != nil {
			t.Fatalf("branch-only rename should succeed when dir already matches: %v\noutput: %s", err, out)
		}

		// Directory unchanged.
		if _, err := os.Stat(wtPath); err != nil {
			t.Errorf("worktree directory should still exist at %q: %v", wtPath, err)
		}

		// Branch renamed.
		cmd := exec.Command("git", "branch", "--list", "mydir")
		cmd.Dir = repo.Root
		branchOut, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if !strings.Contains(string(branchOut), "mydir") {
			t.Errorf("mydir branch should exist after branch-only rename, got: %s", branchOut)
		}
		cmd = exec.Command("git", "branch", "--list", "some-branch")
		cmd.Dir = repo.Root
		branchOut, err = cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if strings.Contains(string(branchOut), "some-branch") {
			t.Errorf("some-branch should have been renamed away, still found: %s", branchOut)
		}
	})

	t.Run("rejects_invalid_branch_name_before_move", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "valid-src")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		oldPath := worktreePath(out)

		// "foo..bar" is rejected by git check-ref-format.
		out, err = runGitWt(t, binPath, repo.Root, "-m", "valid-src", "foo..bar")
		if err == nil {
			t.Fatalf("rename should reject invalid branch name, got: %s", out)
		}
		if !strings.Contains(out, "invalid branch name") {
			t.Errorf("error should mention 'invalid branch name', got: %s", out)
		}

		// Worktree directory must NOT have been moved.
		if _, err := os.Stat(oldPath); err != nil {
			t.Errorf("worktree directory should be untouched after rejected rename, got: %v", err)
		}
	})

	t.Run("rejects_three_args", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "-m", "a", "b", "c")
		if err == nil {
			t.Fatalf("three positional args should fail, got: %s", out)
		}
	})
}

func TestE2E_MoveCurrentWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("outputs_new_path_with_shell_integration", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "renaming-me")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		cmd := exec.Command(binPath, "-m", "renamed-out")
		cmd.Dir = wtPath
		cmd.Env = append(os.Environ(), "GIT_WT_SHELL_INTEGRATION=1")
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -m failed: %v\nstderr: %s", err, stderrBuf.String())
		}
		stdout := strings.TrimSpace(stdoutBuf.String())
		if stdout == "" {
			t.Fatal("expected stdout but got none")
		}

		expectedNew := filepath.Join(filepath.Dir(wtPath), "renamed-out")
		assertLastLine(t, stdout, expectedNew)

		if _, err := os.Stat(expectedNew); err != nil {
			t.Errorf("renamed worktree should exist at %s: %v", expectedNew, err)
		}
	})

	t.Run("no_stdout_without_shell_integration", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "quiet-rename")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		stdout, stderr, err := runGitWtStdout(t, binPath, wtPath, "-m", "quiet-renamed")
		if err != nil {
			t.Fatalf("git-wt -m failed: %v\nstderr: %s", err, stderr)
		}
		if strings.TrimSpace(stdout) != "" {
			t.Errorf("stdout should be empty without GIT_WT_SHELL_INTEGRATION, got: %s", stdout)
		}
		if !strings.Contains(stderr, "Renamed") && !strings.Contains(stderr, "Moved") {
			t.Errorf("stderr should mention rename/move, got: %s", stderr)
		}
	})

	t.Run("shell_integration_cd_to_new_path", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name       string
			shell      string
			scriptFunc func(repoRoot, wtPath, pathDir, newName string) string
		}{
			{
				name:  "bash",
				shell: "bash",
				scriptFunc: func(repoRoot, wtPath, pathDir, newName string) string {
					return fmt.Sprintf(`
set -e
export PATH="%s:$PATH"
eval "$(git wt --init bash)"
cd %q
git wt -m %s
pwd
`, pathDir, wtPath, newName)
				},
			},
			{
				name:  "zsh",
				shell: "zsh",
				scriptFunc: func(repoRoot, wtPath, pathDir, newName string) string {
					return fmt.Sprintf(`
set -e
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"
cd %q
git wt -m %s
pwd
`, pathDir, wtPath, newName)
				},
			},
			{
				name:  "fish",
				shell: "fish",
				scriptFunc: func(repoRoot, wtPath, pathDir, newName string) string {
					return fmt.Sprintf(`
set -x PATH %s $PATH
git wt --init fish | source
cd %q
git wt -m %s
pwd
`, pathDir, wtPath, newName)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				if _, err := exec.LookPath(tt.shell); err != nil {
					t.Skipf("%s not available", tt.shell)
				}

				repo := testutil.NewTestRepo(t)
				repo.CreateFile("README.md", "# Test")
				repo.Commit("initial commit")

				oldName := fmt.Sprintf("shell-%s-old", tt.shell)
				newName := fmt.Sprintf("shell-%s-new", tt.shell)

				out, err := runGitWt(t, binPath, repo.Root, oldName)
				if err != nil {
					t.Fatalf("failed to create worktree: %v", err)
				}
				oldPath := worktreePath(out)
				expectedNew := filepath.Join(filepath.Dir(oldPath), newName)

				script := tt.scriptFunc(repo.Root, oldPath, filepath.Dir(binPath), newName)
				cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
				cmdOut, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s shell integration failed: %v\noutput: %s", tt.shell, err, cmdOut)
				}

				output := strings.TrimSpace(string(cmdOut))
				lines := strings.Split(output, "\n")
				pwd := lines[len(lines)-1]

				// Resolve symlinks for comparison (macOS /private/var vs /var).
				resolvedExpected, errExp := filepath.EvalSymlinks(expectedNew)
				resolvedPwd, errPwd := filepath.EvalSymlinks(pwd)
				if pwd != expectedNew && (errExp != nil || errPwd != nil || resolvedExpected != resolvedPwd) {
					t.Errorf("pwd should be %q after rename, got %q\nfull output: %s", expectedNew, pwd, output)
				}
			})
		}
	})
}
