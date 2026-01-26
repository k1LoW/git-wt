// basic_test.go contains basic functionality tests:
//   - TestE2E_ListWorktrees: listing worktrees and table formatting
//   - TestE2E_CreateWorktree: creating worktrees (basic, start-point, existing branch, from worktree)
//   - TestE2E_SwitchWorktree: switching to existing worktrees
//   - TestE2E_CLI: CLI behavior (version, help, argument validation)
package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
	"github.com/k1LoW/git-wt/testutil"
)

func TestE2E_ListWorktrees(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("git-wt failed: %v\noutput: %s", err, out)
		}

		// Should contain the main worktree
		if !strings.Contains(out, repo.Root) {
			t.Errorf("output should contain repo root %q, got: %s", repo.Root, out)
		}

		if !strings.Contains(out, "main") {
			t.Errorf("output should contain 'main' branch, got: %s", out)
		}
	})

	// Regression test for fish shell hook issue (PR #14)
	t.Run("table_format", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create multiple worktrees to ensure table has multiple rows
		_, err := runGitWt(t, binPath, repo.Root, "feature-a")
		if err != nil {
			t.Fatalf("failed to create worktree feature-a: %v", err)
		}
		_, err = runGitWt(t, binPath, repo.Root, "feature-b")
		if err != nil {
			t.Fatalf("failed to create worktree feature-b: %v", err)
		}

		// Run git wt (list worktrees)
		out, err := runGitWt(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("git-wt failed: %v\noutput: %s", err, out)
		}

		// Table output should have multiple lines (header + at least 3 worktrees)
		lines := strings.Split(out, "\n")
		if len(lines) < 4 {
			t.Errorf("table output should have at least 4 lines (header + 3 worktrees), got %d lines:\n%s", len(lines), out)
		}

		// Each worktree should be on its own line (not collapsed into one line)
		var mainLine, featureALine, featureBLine string
		for _, line := range lines {
			if strings.Contains(line, "main") {
				mainLine = line
			}
			if strings.Contains(line, "feature-a") {
				featureALine = line
			}
			if strings.Contains(line, "feature-b") {
				featureBLine = line
			}
		}

		if mainLine == "" {
			t.Error("main branch should be on its own line")
		}
		if featureALine == "" {
			t.Error("feature-a branch should be on its own line")
		}
		if featureBLine == "" {
			t.Error("feature-b branch should be on its own line")
		}

		// Ensure they are on different lines (not all on the same line)
		if mainLine == featureALine || mainLine == featureBLine || featureALine == featureBLine {
			t.Errorf("each worktree should be on its own line, but found duplicates:\nmain: %q\nfeature-a: %q\nfeature-b: %q", mainLine, featureALine, featureBLine)
		}
	})

	// Regression test for PR #14 which fixed fish hook output formatting
	t.Run("table_format_shell", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name       string
			shell      string
			scriptFunc func(repoRoot, pathDir, binPath string) string
		}{
			{
				name:  "bash",
				shell: "bash",
				scriptFunc: func(repoRoot, pathDir, binPath string) string {
					return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"
git wt
`, repoRoot, pathDir)
				},
			},
			{
				name:  "zsh",
				shell: "zsh",
				scriptFunc: func(repoRoot, pathDir, binPath string) string {
					return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"
git wt
`, repoRoot, pathDir)
				},
			},
			{
				name:  "fish",
				shell: "fish",
				scriptFunc: func(repoRoot, pathDir, binPath string) string {
					return fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source
git wt
`, repoRoot, pathDir)
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

				// Create worktrees first (without shell integration)
				_, err := runGitWt(t, binPath, repo.Root, "feature-a")
				if err != nil {
					t.Fatalf("failed to create worktree feature-a: %v", err)
				}
				_, err = runGitWt(t, binPath, repo.Root, "feature-b")
				if err != nil {
					t.Fatalf("failed to create worktree feature-b: %v", err)
				}

				script := tt.scriptFunc(repo.Root, filepath.Dir(binPath), binPath)
				cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s shell integration failed: %v\noutput: %s", tt.shell, err, out)
				}

				verifyTableFormat(t, string(out))
			})
		}
	})
}

// verifyTableFormat checks that table output has proper newline formatting.
func verifyTableFormat(t *testing.T, output string) {
	t.Helper()

	lines := strings.Split(output, "\n")

	if len(lines) < 4 {
		t.Errorf("table output should have at least 4 lines, got %d lines:\n%s", len(lines), output)
	}

	var mainLine, featureALine, featureBLine string
	for _, line := range lines {
		if strings.Contains(line, "main") {
			mainLine = line
		}
		if strings.Contains(line, "feature-a") {
			featureALine = line
		}
		if strings.Contains(line, "feature-b") {
			featureBLine = line
		}
	}

	if mainLine == "" {
		t.Error("main branch should be on its own line")
	}
	if featureALine == "" {
		t.Error("feature-a branch should be on its own line")
	}
	if featureBLine == "" {
		t.Error("feature-b branch should be on its own line")
	}

	if mainLine == featureALine || mainLine == featureBLine || featureALine == featureBLine {
		t.Errorf("each worktree should be on its own line, but found duplicates:\nmain: %q\nfeature-a: %q\nfeature-b: %q", mainLine, featureALine, featureBLine)
	}
}

func TestE2E_CreateWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "feature-branch")
		if err != nil {
			t.Fatalf("git-wt feature-branch failed: %v\noutput: %s", err, out)
		}

		if !strings.Contains(out, "feature-branch") {
			t.Errorf("output should contain worktree path with 'feature-branch', got: %s", out)
		}

		wtPath := worktreePath(out)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree directory was not created at %s", wtPath)
		}
	})

	t.Run("with_start_point", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		repo.CreateFile("main-file.txt", "main content")
		repo.Commit("main commit")

		repo.Git("branch", "old-base", "HEAD~1")

		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "feature-from-old", "old-base")
		if err != nil {
			t.Fatalf("git-wt with start-point failed: %v\nstderr: %s", err, stderr)
		}

		wtPath := strings.TrimSpace(stdout)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Fatalf("worktree was not created at %s", wtPath)
		}

		// Verify the worktree is based on old-base (should NOT have main-file.txt)
		mainFilePath := filepath.Join(wtPath, "main-file.txt")
		if _, err := os.Stat(mainFilePath); !os.IsNotExist(err) {
			t.Error("worktree should NOT have main-file.txt (should be based on old-base, not main)")
		}
	})

	t.Run("with_remote_start_point", func(t *testing.T) {
		t.Parallel()
		// Create a "remote" repo
		remoteRepo := testutil.NewTestRepo(t)
		remoteRepo.CreateFile("README.md", "# Remote")
		remoteRepo.Commit("remote initial commit")
		remoteRepo.CreateFile("remote-file.txt", "remote content")
		remoteRepo.Commit("remote second commit")

		// Clone the remote repo
		cloneDir := t.TempDir()
		clonePath := filepath.Join(cloneDir, "clone")
		cmd := exec.Command("git", "clone", remoteRepo.Root, clonePath)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git clone failed: %v\noutput: %s", err, out)
		}

		cmd = exec.Command("git", "checkout", "main")
		cmd.Dir = clonePath
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git checkout failed: %v\noutput: %s", err, out)
		}

		stdout, stderr, err := runGitWtStdout(t, binPath, clonePath, "feature-from-remote", "origin/main~1")
		if err != nil {
			t.Fatalf("git-wt with remote start-point failed: %v\nstderr: %s", err, stderr)
		}

		wtPath := strings.TrimSpace(stdout)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Fatalf("worktree was not created at %s", wtPath)
		}

		// Verify the worktree is based on the first commit (should NOT have remote-file.txt)
		remoteFilePath := filepath.Join(wtPath, "remote-file.txt")
		if _, err := os.Stat(remoteFilePath); !os.IsNotExist(err) {
			t.Error("worktree should NOT have remote-file.txt (should be based on origin/main~1)")
		}
	})

	t.Run("existing_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		repo.Git("branch", "existing-branch")

		out, err := runGitWt(t, binPath, repo.Root, "existing-branch")
		if err != nil {
			t.Fatalf("failed to create worktree for existing branch: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree was not created at %s", wtPath)
		}

		restore := repo.Chdir()
		defer restore()

		cmd := exec.Command("git", "branch", "--list", "existing-branch")
		branchOut, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}

		if !strings.Contains(string(branchOut), "existing-branch") {
			t.Error("existing-branch should still exist")
		}
	})

	t.Run("start_point_ignored_for_existing_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		repo.Git("branch", "existing-branch")

		repo.CreateFile("new-file.txt", "new content")
		repo.Commit("second commit")

		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "existing-branch", "main")
		if err != nil {
			t.Fatalf("git-wt failed: %v\nstderr: %s", err, stderr)
		}

		wtPath := strings.TrimSpace(stdout)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Fatalf("worktree was not created at %s", wtPath)
		}

		// Verify the worktree is based on existing-branch (should NOT have new-file.txt)
		newFilePath := filepath.Join(wtPath, "new-file.txt")
		if _, err := os.Stat(newFilePath); !os.IsNotExist(err) {
			t.Error("worktree should NOT have new-file.txt (start-point should be ignored for existing branch)")
		}
	})

	t.Run("from_worktree", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		repo.Git("config", "wt.basedir", "../{gitroot}-wt")

		out1, err := runGitWt(t, binPath, repo.Root, "feature1")
		if err != nil {
			t.Fatalf("failed to create first worktree: %v", err)
		}
		wt1Path := worktreePath(out1)

		out2, err := runGitWt(t, binPath, wt1Path, "feature2")
		if err != nil {
			t.Fatalf("failed to create second worktree from worktree: %v\noutput: %s", err, out2)
		}
		wt2Path := worktreePath(out2)

		if _, err := os.Stat(wt2Path); os.IsNotExist(err) {
			t.Errorf("second worktree was not created at %s", wt2Path)
		}

		expectedWt1Path := filepath.Clean(filepath.Join(repo.Root, "../repo-wt/feature1"))
		if wt1Path != expectedWt1Path {
			t.Errorf("first worktree path = %q, want %q", wt1Path, expectedWt1Path)
		}

		expectedWt2Path := filepath.Clean(filepath.Join(repo.Root, "../repo-wt/feature2"))
		if wt2Path != expectedWt2Path {
			t.Errorf("second worktree path = %q, want %q", wt2Path, expectedWt2Path)
		}
	})
}

func TestE2E_SwitchWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	out1, err := runGitWt(t, binPath, repo.Root, "feature")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out1)
	}
	wtPath := worktreePath(out1)

	// Running again should return the same path (switch behavior)
	out2, err := runGitWt(t, binPath, repo.Root, "feature")
	if err != nil {
		t.Fatalf("git-wt feature failed: %v\noutput: %s", err, out2)
	}

	if out2 != wtPath {
		t.Errorf("expected same path %q, got %q", wtPath, out2)
	}
}

func TestE2E_CLI(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("completion_subcommand_disabled", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// "git wt completion" should NOT show Cobra's default completion help.
		// Instead, it should treat "completion" as a branch name.
		out, err := runGitWt(t, binPath, repo.Root, "completion")

		// Since "completion" branch doesn't exist and is not a valid ref,
		// it should fail with an error (trying to create worktree for non-existent ref)
		// OR succeed by creating a new branch named "completion".
		// The key assertion is: it should NOT show Cobra's completion help message.
		if strings.Contains(out, "Generate the autocompletion script") {
			t.Errorf("should NOT show Cobra's default completion help, got: %s", out)
		}
		if strings.Contains(out, "completion [command]") {
			t.Errorf("should NOT show Cobra's completion subcommand help, got: %s", out)
		}

		// If the command succeeds, it means a worktree was created for branch "completion"
		if err == nil {
			wtPath := worktreePath(out)
			if !strings.Contains(wtPath, "completion") {
				t.Errorf("expected worktree path to contain 'completion', got: %s", wtPath)
			}
		}
		// If it fails, verify it's not because of Cobra's completion command
		// (should be a git error about invalid reference, not Cobra help)
	})

	t.Run("version", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--version")
		if err != nil {
			t.Fatalf("git-wt --version failed: %v\noutput: %s", err, out)
		}

		if !strings.Contains(out, "version") {
			t.Errorf("expected version output, got: %s", out)
		}
	})

	t.Run("help", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--help")
		if err != nil {
			t.Fatalf("git-wt --help failed: %v\noutput: %s", err, out)
		}

		expectedStrings := []string{
			"git wt [branch|worktree] [start-point]",
			"--delete",
			"--force-delete",
			"--init",
		}

		for _, s := range expectedStrings {
			if !strings.Contains(out, s) {
				t.Errorf("help output should contain %q", s)
			}
		}
	})

	t.Run("too_many_arguments", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		_, err := runGitWt(t, binPath, repo.Root, "branch-a", "branch-b", "branch-c")
		if err == nil {
			t.Fatal("command should fail when passing more than 2 arguments without -d/-D flag")
		}
	})
}
