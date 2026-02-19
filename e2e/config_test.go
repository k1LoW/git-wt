// config_test.go contains configuration and flag tests:
//   - TestE2E_CopyOptions: copy options tests (copyignored config/flag, copyuntracked, copymodified, multiple flags, flag overrides)
//   - TestE2E_Basedir: basedir tests (config, flag)
//   - TestE2E_Nocd: nocd tests (config, config_with_init, create_config)
//   - TestE2E_Hooks: hook tests (flag, config, multiple, not_run_on_existing, flag_overrides_config, failure, output_to_stderr)
//   - TestE2E_DeleteHooks: delete hook tests (flag, config, multiple, not_run_on_branch_only, flag_overrides_config, failure_prevents_deletion, hook_runs_in_worktree_directory, output_to_stderr)
//   - TestE2E_Remover: custom worktree remover tests (flag, config, flag_overrides_config, failure_prevents_deletion, prune_cleans_up)
//   - TestE2E_Complete: __complete command output tests
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

func TestE2E_CopyOptions(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("copyignored_config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile(".gitignore", ".env\n")
		repo.Commit("initial commit")

		// Create ignored file
		repo.CreateFile(".env", "SECRET=supersecret")

		// Enable copyignored
		repo.Git("config", "wt.copyignored", "true")

		// Create worktree
		out, err := runGitWt(t, binPath, repo.Root, "with-env")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Verify .env was copied
		envPath := filepath.Join(wtPath, ".env")
		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf(".env was not copied to worktree: %v", err)
		}

		if string(content) != "SECRET=supersecret" {
			t.Errorf(".env content = %q, want %q", string(content), "SECRET=supersecret")
		}
	})

	t.Run("copyignored_flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile(".gitignore", ".env\n")
		repo.Commit("initial commit")

		// Create ignored file
		repo.CreateFile(".env", "SECRET=flagtest")

		// Config is NOT set (default false)
		// Use --copyignored flag

		out, err := runGitWt(t, binPath, repo.Root, "--copyignored", "copyignored-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree with --copyignored flag: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify .env was copied
		envPath := filepath.Join(wtPath, ".env")
		content, err := os.ReadFile(envPath)
		if err != nil {
			t.Fatalf(".env was not copied to worktree despite --copyignored flag: %v", err)
		}

		if string(content) != "SECRET=flagtest" {
			t.Errorf(".env content = %q, want %q", string(content), "SECRET=flagtest")
		}
	})

	t.Run("copyuntracked_flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create untracked file
		repo.CreateFile("untracked.txt", "untracked-flag-content")

		// Use --copyuntracked flag
		out, err := runGitWt(t, binPath, repo.Root, "--copyuntracked", "copyuntracked-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree with --copyuntracked flag: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify untracked file was copied
		untrackedPath := filepath.Join(wtPath, "untracked.txt")
		content, err := os.ReadFile(untrackedPath)
		if err != nil {
			t.Fatalf("untracked.txt was not copied to worktree despite --copyuntracked flag: %v", err)
		}

		if string(content) != "untracked-flag-content" {
			t.Errorf("untracked.txt content = %q, want %q", string(content), "untracked-flag-content")
		}
	})

	t.Run("copymodified_flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile("tracked.txt", "original-content")
		repo.Commit("initial commit")

		// Modify tracked file
		repo.CreateFile("tracked.txt", "modified-flag-content")

		// Use --copymodified flag
		out, err := runGitWt(t, binPath, repo.Root, "--copymodified", "copymodified-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree with --copymodified flag: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify modified file was copied with modifications
		trackedPath := filepath.Join(wtPath, "tracked.txt")
		content, err := os.ReadFile(trackedPath)
		if err != nil {
			t.Fatalf("tracked.txt was not copied to worktree: %v", err)
		}

		if string(content) != "modified-flag-content" {
			t.Errorf("tracked.txt content = %q, want %q", string(content), "modified-flag-content")
		}
	})

	t.Run("multiple_flags", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile(".gitignore", ".env\n")
		repo.CreateFile("tracked.txt", "original")
		repo.Commit("initial commit")

		// Create various files
		repo.CreateFile(".env", "SECRET=multi")
		repo.CreateFile("untracked.txt", "untracked-multi")
		repo.CreateFile("tracked.txt", "modified-multi")

		// Use multiple flags
		out, err := runGitWt(t, binPath, repo.Root, "--copyignored", "--copyuntracked", "--copymodified", "multi-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree with multiple flags: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify all files were copied
		tests := []struct {
			file    string
			content string
		}{
			{".env", "SECRET=multi"},
			{"untracked.txt", "untracked-multi"},
			{"tracked.txt", "modified-multi"},
		}

		for _, tt := range tests {
			path := filepath.Join(wtPath, tt.file)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s was not copied to worktree: %v", tt.file, err)
				continue
			}
			if string(content) != tt.content {
				t.Errorf("%s content = %q, want %q", tt.file, string(content), tt.content)
			}
		}
	})

	t.Run("flag_overrides_config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile(".gitignore", ".env\n")
		repo.Commit("initial commit")

		// Set copyignored to false in config
		repo.Git("config", "wt.copyignored", "false")

		// Create ignored file
		repo.CreateFile(".env", "SECRET=override")

		// Without flag, .env should NOT be copied
		out1, err := runGitWt(t, binPath, repo.Root, "no-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out1)
		}
		wtPath1 := worktreePath(out1)

		envPath1 := filepath.Join(wtPath1, ".env")
		if _, err := os.Stat(envPath1); !os.IsNotExist(err) {
			t.Error(".env should NOT have been copied without --copyignored flag")
		}

		// With flag, .env SHOULD be copied
		out2, err := runGitWt(t, binPath, repo.Root, "--copyignored", "with-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree with flag: %v\noutput: %s", err, out2)
		}
		wtPath2 := worktreePath(out2)

		envPath2 := filepath.Join(wtPath2, ".env")
		if _, err := os.Stat(envPath2); os.IsNotExist(err) {
			t.Error(".env SHOULD have been copied with --copyignored flag")
		}
	})

	t.Run("copyignored_excludes_basedir", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile(".gitignore", ".env\n.worktrees/\n")
		repo.Commit("initial commit")

		// Set basedir inside the repo (which is gitignored)
		basedir := filepath.Join(repo.Root, ".worktrees")
		repo.Git("config", "wt.basedir", basedir)
		repo.Git("config", "wt.copyignored", "true")

		// Create ignored file
		repo.CreateFile(".env", "SECRET=basedir-test")

		// Create first worktree
		out1, err := runGitWt(t, binPath, repo.Root, "first-wt")
		if err != nil {
			t.Fatalf("failed to create first worktree: %v\noutput: %s", err, out1)
		}
		wtPath1 := worktreePath(out1)

		// Verify .env was copied to first worktree
		envPath1 := filepath.Join(wtPath1, ".env")
		if _, err := os.Stat(envPath1); os.IsNotExist(err) {
			t.Error(".env should have been copied to first worktree")
		}

		// Create second worktree - it should NOT copy files from first worktree
		out2, err := runGitWt(t, binPath, repo.Root, "second-wt")
		if err != nil {
			t.Fatalf("failed to create second worktree: %v\noutput: %s", err, out2)
		}
		wtPath2 := worktreePath(out2)

		// Verify .env was copied to second worktree
		envPath2 := filepath.Join(wtPath2, ".env")
		if _, err := os.Stat(envPath2); os.IsNotExist(err) {
			t.Error(".env should have been copied to second worktree")
		}

		// Verify files from first worktree were NOT copied to second worktree
		// (basedir should be excluded from copyignored)
		firstWtReadme := filepath.Join(wtPath2, "first-wt/README.md")
		if _, err := os.Stat(firstWtReadme); !os.IsNotExist(err) {
			t.Error("files from first worktree should NOT have been copied to second worktree (basedir should be excluded)")
		}

		// Also check that .worktrees/.gitignore was not copied
		basedirGitignore := filepath.Join(wtPath2, ".worktrees/.gitignore")
		if _, err := os.Stat(basedirGitignore); !os.IsNotExist(err) {
			t.Error(".worktrees/.gitignore should NOT have been copied (basedir should be excluded)")
		}
	})
}

func TestE2E_Basedir(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Set custom basedir
		customBase := filepath.Join(repo.ParentDir(), "custom-wt-dir")
		repo.Git("config", "wt.basedir", customBase)

		// Create worktree
		out, err := runGitWt(t, binPath, repo.Root, "custom-branch")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify worktree was created in custom location
		expectedPath := filepath.Join(customBase, "custom-branch")
		if wtPath != expectedPath {
			t.Errorf("worktree path = %q, want %q", wtPath, expectedPath)
		}

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree was not created at custom path %s", wtPath)
		}
	})

	t.Run("flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Set config basedir to one location
		configBase := filepath.Join(repo.ParentDir(), "config-wt-dir")
		repo.Git("config", "wt.basedir", configBase)

		// Use flag to override to different location
		flagBase := filepath.Join(repo.ParentDir(), "flag-wt-dir")

		// Create worktree with --basedir flag
		out, err := runGitWt(t, binPath, repo.Root, "--basedir", flagBase, "flag-branch")
		if err != nil {
			t.Fatalf("failed to create worktree with --basedir flag: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify worktree was created in flag location, not config location
		expectedPath := filepath.Join(flagBase, "flag-branch")
		if wtPath != expectedPath {
			t.Errorf("worktree path = %q, want %q", wtPath, expectedPath)
		}

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree was not created at flag path %s", wtPath)
		}

		// Verify it was NOT created in config location
		configPath := filepath.Join(configBase, "flag-branch")
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Errorf("worktree should not have been created at config path %s", configPath)
		}
	})
}

func TestE2E_Nocd(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("config", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name       string
			shell      string
			scriptFunc func(repoRoot, pathDir, branchName string) string
		}{
			{
				name:  "bash",
				shell: "bash",
				scriptFunc: func(repoRoot, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"
git wt %s
pwd
`, repoRoot, pathDir, branchName)
				},
			},
			{
				name:  "zsh",
				shell: "zsh",
				scriptFunc: func(repoRoot, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"
git wt %s
pwd
`, repoRoot, pathDir, branchName)
				},
			},
			{
				name:  "fish",
				shell: "fish",
				scriptFunc: func(repoRoot, pathDir, branchName string) string {
					return fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source
git wt %s
pwd
`, repoRoot, pathDir, branchName)
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

				repo.Git("config", "wt.nocd", "true")

				branchName := fmt.Sprintf("nocd-config-%s-test", tt.shell)
				script := tt.scriptFunc(repo.Root, filepath.Dir(binPath), branchName)
				cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s shell integration with wt.nocd config failed: %v\noutput: %s", tt.shell, err, out)
				}

				output := strings.TrimSpace(string(out))
				lines := strings.Split(output, "\n")
				pwd := lines[len(lines)-1]

				if strings.Contains(pwd, branchName) {
					t.Errorf("pwd should NOT contain worktree path when wt.nocd=true, got: %s", pwd)
				}
				if pwd != repo.Root {
					t.Errorf("pwd should be original repo root %q, got: %s", repo.Root, pwd)
				}
			})
		}
	})

	t.Run("config_with_init", func(t *testing.T) {
		t.Parallel()
		// Test that --init ignores wt.nocd config and always outputs git() wrapper.
		// The wt.nocd config only affects cd behavior at runtime, not the init output.
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Set wt.nocd in config
		repo.Git("config", "wt.nocd", "true")

		restore := repo.Chdir()
		defer restore()

		out, err := runGitWt(t, binPath, repo.Root, "--init", "bash")
		if err != nil {
			t.Fatalf("git-wt --init bash failed: %v\noutput: %s", err, out)
		}

		// With wt.nocd=true config, --init should still output git wrapper
		// (wt.nocd only affects runtime cd behavior, not init output)
		if !strings.Contains(out, "git() {") {
			t.Error("output should contain git wrapper even when wt.nocd=true config is set")
		}
	})

	t.Run("create_config", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name       string
			shell      string
			scriptFunc func(repoRoot, pathDir, branchName string) string
		}{
			{
				name:  "bash",
				shell: "bash",
				scriptFunc: func(repoRoot, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Create a worktree first (should NOT cd because wt.nocd=create)
git wt %s
NEW_PWD=$(pwd)

# Switch to existing worktree (should cd because wt.nocd=create allows existing)
git wt %s
EXISTING_PWD=$(pwd)

echo "NEW_PWD=$NEW_PWD"
echo "EXISTING_PWD=$EXISTING_PWD"
`, repoRoot, pathDir, branchName, branchName)
				},
			},
			{
				name:  "zsh",
				shell: "zsh",
				scriptFunc: func(repoRoot, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Create a worktree first (should NOT cd because wt.nocd=create)
git wt %s
NEW_PWD=$(pwd)

# Switch to existing worktree (should cd because wt.nocd=create allows existing)
git wt %s
EXISTING_PWD=$(pwd)

echo "NEW_PWD=$NEW_PWD"
echo "EXISTING_PWD=$EXISTING_PWD"
`, repoRoot, pathDir, branchName, branchName)
				},
			},
			{
				name:  "fish",
				shell: "fish",
				scriptFunc: func(repoRoot, pathDir, branchName string) string {
					return fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Create a worktree first (should NOT cd because wt.nocd=create)
git wt %s
set NEW_PWD (pwd)

# Switch to existing worktree (should cd because wt.nocd=create allows existing)
git wt %s
set EXISTING_PWD (pwd)

echo "NEW_PWD=$NEW_PWD"
echo "EXISTING_PWD=$EXISTING_PWD"
`, repoRoot, pathDir, branchName, branchName)
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

				repo.Git("config", "wt.nocd", "create")

				branchName := fmt.Sprintf("nocd-create-%s-new", tt.shell)
				script := tt.scriptFunc(repo.Root, filepath.Dir(binPath), branchName)
				cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s shell integration with wt.nocd=create config failed: %v\noutput: %s", tt.shell, err, out)
				}

				output := string(out)

				var newPwd, existingPwd string
				for _, line := range strings.Split(output, "\n") {
					if strings.HasPrefix(line, "NEW_PWD=") {
						newPwd = strings.TrimPrefix(line, "NEW_PWD=")
					}
					if strings.HasPrefix(line, "EXISTING_PWD=") {
						existingPwd = strings.TrimPrefix(line, "EXISTING_PWD=")
					}
				}

				if strings.Contains(newPwd, branchName) {
					t.Errorf("NEW_PWD should NOT contain worktree path when creating new worktree with wt.nocd=create, got: %s", newPwd)
				}

				if !strings.Contains(existingPwd, branchName) {
					t.Errorf("EXISTING_PWD should contain worktree path when switching to existing worktree with wt.nocd=create, got: %s", existingPwd)
				}
			})
		}
	})
}

func TestE2E_Hooks(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree with --hook flag that creates a marker file
		out, err := runGitWt(t, binPath, repo.Root, "--hook", "touch hook-marker.txt", "hook-test")
		if err != nil {
			t.Fatalf("failed to create worktree with --hook flag: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify the hook created the marker file
		markerPath := filepath.Join(wtPath, "hook-marker.txt")
		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("hook-marker.txt was not created by hook")
		}
	})

	t.Run("config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Set hook in config
		repo.Git("config", "--add", "wt.hook", "touch config-hook-marker.txt")

		// Create worktree
		out, err := runGitWt(t, binPath, repo.Root, "hook-config-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify the hook created the marker file
		markerPath := filepath.Join(wtPath, "config-hook-marker.txt")
		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("config-hook-marker.txt was not created by hook from config")
		}
	})

	t.Run("multiple", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree with multiple hooks
		out, err := runGitWt(t, binPath, repo.Root, "--hook", "echo first > order.txt", "--hook", "echo second >> order.txt", "multi-hook-test")
		if err != nil {
			t.Fatalf("failed to create worktree with multiple hooks: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify both hooks ran in order
		orderPath := filepath.Join(wtPath, "order.txt")
		content, err := os.ReadFile(orderPath)
		if err != nil {
			t.Fatalf("order.txt was not created: %v", err)
		}

		expected := "first\nsecond\n"
		if string(content) != expected {
			t.Errorf("order.txt content = %q, want %q", string(content), expected)
		}
	})

	t.Run("not_run_on_existing", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree first (without hook)
		out, err := runGitWt(t, binPath, repo.Root, "existing-hook-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Switch to existing worktree with hook - hook should NOT run
		out2, err := runGitWt(t, binPath, repo.Root, "--hook", "touch should-not-exist.txt", "existing-hook-test")
		if err != nil {
			t.Fatalf("failed to switch to worktree: %v\noutput: %s", err, out2)
		}

		// Verify the hook did NOT create the file
		markerPath := filepath.Join(wtPath, "should-not-exist.txt")
		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("hook should NOT have run when switching to existing worktree")
		}
	})

	t.Run("flag_overrides_config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Set hook in config
		repo.Git("config", "--add", "wt.hook", "touch config-marker.txt")

		// Create worktree with --hook flag (should override config)
		out, err := runGitWt(t, binPath, repo.Root, "--hook", "touch flag-marker.txt", "hook-override-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		// Verify flag hook ran
		flagMarkerPath := filepath.Join(wtPath, "flag-marker.txt")
		if _, err := os.Stat(flagMarkerPath); os.IsNotExist(err) {
			t.Error("flag-marker.txt should have been created by --hook flag")
		}

		// Verify config hook did NOT run (flag overrides config)
		configMarkerPath := filepath.Join(wtPath, "config-marker.txt")
		if _, err := os.Stat(configMarkerPath); !os.IsNotExist(err) {
			t.Error("config-marker.txt should NOT have been created (--hook flag overrides config)")
		}
	})

	t.Run("failure_exits_with_error", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree with a failing hook followed by a successful hook
		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "--hook", "exit 1", "--hook", "touch after-failure.txt", "hook-failure-test")

		// Command should fail with exit code 1
		if err == nil {
			t.Fatal("command should fail when hook fails")
		}

		wtPath := strings.TrimSpace(stdout)

		// Verify stderr contains error about the failed hook
		if !strings.Contains(stderr, "hook") || !strings.Contains(stderr, "failed") {
			t.Errorf("stderr should contain error about failed hook, got: %s", stderr)
		}

		// Verify the second hook did NOT run (execution stops on first failure)
		markerPath := filepath.Join(wtPath, "after-failure.txt")
		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("second hook should NOT have run after first hook failed")
		}

		// Verify worktree was still created
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Error("worktree should have been created even though hook failed")
		}
	})

	t.Run("output_to_stderr", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree with a hook that outputs to stdout
		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "--hook", "echo hook-output-test", "hook-output-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\nstderr: %s", err, stderr)
		}

		// stdout should only contain the worktree path (for shell integration)
		lines := strings.Split(stdout, "\n")
		if len(lines) != 1 {
			t.Errorf("stdout should be exactly 1 line (worktree path), got %d lines: %q", len(lines), stdout)
		}

		// Hook output should be in stderr
		if !strings.Contains(stderr, "hook-output-test") {
			t.Errorf("hook output should be in stderr, got stderr: %s", stderr)
		}
	})
}

func TestE2E_DeleteHooks(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		markerPath := filepath.Join(t.TempDir(), "delete-hook-marker.txt")

		out, err = runGitWt(t, binPath, repo.Root, "-D", "--deletehook", fmt.Sprintf("touch %s", markerPath), "delete-hook-flag-test")
		if err != nil {
			t.Fatalf("git-wt -D --deletehook failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("delete hook marker was not created")
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		markerPath := filepath.Join(t.TempDir(), "config-delete-hook-marker.txt")

		repo.Git("config", "--add", "wt.deletehook", fmt.Sprintf("touch %s", markerPath))

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-config-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-D", "delete-hook-config-test")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(markerPath); os.IsNotExist(err) {
			t.Error("delete hook marker from config was not created")
		}
	})

	t.Run("multiple", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-multi-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}

		orderPath := filepath.Join(t.TempDir(), "delete-hook-order.txt")

		out, err = runGitWt(t, binPath, repo.Root, "-D",
			"--deletehook", fmt.Sprintf("echo first > %s", orderPath),
			"--deletehook", fmt.Sprintf("echo second >> %s", orderPath),
			"delete-hook-multi-test")
		if err != nil {
			t.Fatalf("git-wt -D --deletehook (multiple) failed: %v\noutput: %s", err, out)
		}

		content, err := os.ReadFile(orderPath)
		if err != nil {
			t.Fatalf("order file was not created: %v", err)
		}

		expected := "first\nsecond\n"
		if string(content) != expected {
			t.Errorf("order file content = %q, want %q", string(content), expected)
		}
	})

	t.Run("not_run_on_branch_only", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		repo.Git("branch", "orphan-branch")

		markerPath := filepath.Join(t.TempDir(), "branch-only-marker.txt")

		out, err := runGitWt(t, binPath, repo.Root, "-D", "--deletehook", fmt.Sprintf("touch %s", markerPath), "orphan-branch")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
			t.Error("delete hook should NOT have run when deleting a branch without a worktree")
		}
	})

	t.Run("flag_overrides_config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		configMarker := filepath.Join(t.TempDir(), "config-marker.txt")
		flagMarker := filepath.Join(t.TempDir(), "flag-marker.txt")

		repo.Git("config", "--add", "wt.deletehook", fmt.Sprintf("touch %s", configMarker))

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-override-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-D", "--deletehook", fmt.Sprintf("touch %s", flagMarker), "delete-hook-override-test")
		if err != nil {
			t.Fatalf("git-wt -D --deletehook failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(flagMarker); os.IsNotExist(err) {
			t.Error("flag delete hook marker should have been created")
		}

		if _, err := os.Stat(configMarker); !os.IsNotExist(err) {
			t.Error("config delete hook should NOT have run (--deletehook flag overrides config)")
		}
	})

	t.Run("failure_prevents_deletion", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-fail-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		_, _, err = runGitWtStdout(t, binPath, repo.Root, "-D", "--deletehook", "exit 1", "delete-hook-fail-test")
		if err == nil {
			t.Fatal("command should fail when delete hook fails")
		}

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Error("worktree should NOT have been deleted when delete hook fails")
		}
	})

	t.Run("hook_runs_in_worktree_directory", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-cwd-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		pwdFile := filepath.Join(t.TempDir(), "hook-pwd.txt")

		out, err = runGitWt(t, binPath, repo.Root, "-D", "--deletehook", fmt.Sprintf("pwd > %s", pwdFile), "delete-hook-cwd-test")
		if err != nil {
			t.Fatalf("git-wt -D --deletehook failed: %v\noutput: %s", err, out)
		}

		content, err := os.ReadFile(pwdFile)
		if err != nil {
			t.Fatalf("pwd file was not created: %v", err)
		}

		hookPwd := strings.TrimSpace(string(content))
		if hookPwd != wtPath {
			t.Errorf("hook working directory = %q, want %q", hookPwd, wtPath)
		}
	})

	t.Run("output_to_stderr", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "delete-hook-stderr-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}

		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "-D", "--deletehook", "echo delete-hook-output-test", "delete-hook-stderr-test")
		if err != nil {
			t.Fatalf("git-wt -D --deletehook failed: %v\nstderr: %s", err, stderr)
		}

		if strings.Contains(stdout, "delete-hook-output-test") {
			t.Errorf("hook output should NOT be in stdout, got: %s", stdout)
		}

		if !strings.Contains(stderr, "delete-hook-output-test") {
			t.Errorf("hook output should be in stderr, got stderr: %s", stderr)
		}
	})
}

func TestE2E_Relative(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("config_new_worktree", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile("some/path/file.txt", "content")
		repo.Commit("initial commit")

		// Enable wt.relative
		repo.Git("config", "wt.relative", "true")

		// Run from subdirectory
		subdir := filepath.Join(repo.Root, "some", "path")
		stdout, _, err := runGitWtStdout(t, binPath, subdir, "relative-new-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Output path should end with some/path
		if !strings.HasSuffix(stdout, filepath.Join("some", "path")) {
			t.Errorf("stdout should end with some/path, got: %s", stdout)
		}

		// The path should be a valid directory
		if _, err := os.Stat(stdout); os.IsNotExist(err) {
			t.Errorf("resolved path should exist: %s", stdout)
		}
	})

	t.Run("config_existing_worktree", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile("some/path/file.txt", "content")
		repo.Commit("initial commit")

		// Create worktree first (from root, no relative)
		_, err := runGitWt(t, binPath, repo.Root, "relative-existing-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Enable wt.relative
		repo.Git("config", "wt.relative", "true")

		// Switch to existing worktree from subdirectory
		subdir := filepath.Join(repo.Root, "some", "path")
		stdout, _, err := runGitWtStdout(t, binPath, subdir, "relative-existing-test")
		if err != nil {
			t.Fatalf("failed to switch worktree: %v", err)
		}

		// Output path should end with some/path
		if !strings.HasSuffix(stdout, filepath.Join("some", "path")) {
			t.Errorf("stdout should end with some/path, got: %s", stdout)
		}
	})

	t.Run("flag_override", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile("sub/file.txt", "content")
		repo.Commit("initial commit")

		// Config is NOT set (default false)
		// Use --relative flag from subdirectory
		subdir := filepath.Join(repo.Root, "sub")
		stdout, _, err := runGitWtStdout(t, binPath, subdir, "--relative", "relative-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Output path should end with sub
		if !strings.HasSuffix(stdout, "sub") {
			t.Errorf("stdout should end with 'sub', got: %s", stdout)
		}
	})

	t.Run("fallback_when_subdir_not_exists", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create an untracked directory (won't exist in the new worktree)
		untrackedDir := filepath.Join(repo.Root, "untracked-dir")
		if err := os.MkdirAll(untrackedDir, 0o755); err != nil {
			t.Fatalf("failed to create untracked dir: %v", err)
		}

		// Enable wt.relative
		repo.Git("config", "wt.relative", "true")

		// Run from untracked directory
		stdout, _, err := runGitWtStdout(t, binPath, untrackedDir, "relative-fallback-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Output path should NOT contain untracked-dir (fallback to worktree root)
		if strings.Contains(stdout, "untracked-dir") {
			t.Errorf("stdout should not contain 'untracked-dir' (should fallback), got: %s", stdout)
		}
	})

	t.Run("at_repo_root", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Enable wt.relative
		repo.Git("config", "wt.relative", "true")

		// Run from repo root
		stdout, _, err := runGitWtStdout(t, binPath, repo.Root, "relative-root-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Output should be normal worktree path (no subdirectory appended)
		if !strings.HasSuffix(stdout, "relative-root-test") {
			t.Errorf("stdout should end with branch name at repo root, got: %s", stdout)
		}
	})

	t.Run("disabled_by_default", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.CreateFile("some/path/file.txt", "content")
		repo.Commit("initial commit")

		// Config NOT set (default false)

		// Run from subdirectory
		subdir := filepath.Join(repo.Root, "some", "path")
		stdout, _, err := runGitWtStdout(t, binPath, subdir, "relative-disabled-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Output should NOT contain some/path (relative is disabled)
		if strings.HasSuffix(stdout, filepath.Join("some", "path")) {
			t.Errorf("stdout should NOT end with some/path when relative is disabled, got: %s", stdout)
		}

		// Output should end with branch name
		if !strings.HasSuffix(stdout, "relative-disabled-test") {
			t.Errorf("stdout should end with branch name, got: %s", stdout)
		}
	})
}

func TestE2E_Remover(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("flag", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "remover-flag-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		out, err = runGitWt(t, binPath, repo.Root, "-D", "--remover", "rm -rf", "remover-flag-test")
		if err != nil {
			t.Fatalf("git-wt -D --remover failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree directory should have been removed by custom remover")
		}
	})

	t.Run("config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		repo.Git("config", "wt.remover", "rm -rf")

		out, err := runGitWt(t, binPath, repo.Root, "remover-config-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		out, err = runGitWt(t, binPath, repo.Root, "-D", "remover-config-test")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree directory should have been removed by custom remover from config")
		}
	})

	t.Run("flag_overrides_config", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		configMarker := filepath.Join(t.TempDir(), "config-remover-ran.txt")
		flagMarker := filepath.Join(t.TempDir(), "flag-remover-ran.txt")

		repo.Git("config", "wt.remover", fmt.Sprintf("sh -c 'touch %s && rm -rf \"$1\"' --", configMarker))

		out, err := runGitWt(t, binPath, repo.Root, "remover-override-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-D",
			"--remover", fmt.Sprintf("sh -c 'touch %s && rm -rf \"$1\"' --", flagMarker),
			"remover-override-test")
		if err != nil {
			t.Fatalf("git-wt -D --remover failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(flagMarker); os.IsNotExist(err) {
			t.Error("flag remover marker should have been created")
		}

		if _, err := os.Stat(configMarker); !os.IsNotExist(err) {
			t.Error("config remover should NOT have run (--remover flag overrides config)")
		}
	})

	t.Run("failure_prevents_deletion", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "remover-fail-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		_, _, err = runGitWtStdout(t, binPath, repo.Root, "-D", "--remover", "exit 1", "remover-fail-test")
		if err == nil {
			t.Fatal("command should fail when remover fails")
		}

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Error("worktree should NOT have been deleted when remover fails")
		}
	})

	t.Run("prune_cleans_up", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "remover-prune-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-D", "--remover", "rm -rf", "remover-prune-test")
		if err != nil {
			t.Fatalf("git-wt -D --remover failed: %v\noutput: %s", err, out)
		}

		wtList, err := runGitWt(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("failed to list worktrees: %v\noutput: %s", err, wtList)
		}

		if strings.Contains(wtList, "remover-prune-test") {
			t.Errorf("worktree list should not contain removed worktree, got: %s", wtList)
		}
	})

	t.Run("path_with_spaces", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		spaceBase := filepath.Join(t.TempDir(), "dir with spaces")
		if err := os.MkdirAll(spaceBase, 0o755); err != nil {
			t.Fatalf("failed to create dir with spaces: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "--basedir", spaceBase, "space-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
		}
		wtPath := worktreePath(out)

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Fatalf("worktree should exist at %q", wtPath)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-D", "--basedir", spaceBase, "--remover", "rm -rf", "space-test")
		if err != nil {
			t.Fatalf("git-wt -D --remover with spaces failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree directory should have been removed despite spaces in path")
		}
	})
}

// TestE2E_Complete tests the __complete command output.
func TestE2E_Complete(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create some branches for completion
	repo.Git("branch", "feature-one")
	repo.Git("branch", "feature-two")

	t.Run("empty_input_returns_branches", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain branch names
		if !strings.Contains(out, "main") {
			t.Errorf("completion should contain 'main', got: %s", out)
		}
		if !strings.Contains(out, "feature-one") {
			t.Errorf("completion should contain 'feature-one', got: %s", out)
		}
		if !strings.Contains(out, "feature-two") {
			t.Errorf("completion should contain 'feature-two', got: %s", out)
		}
	})

	t.Run("partial_input_filters_branches", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "feat")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain matching branches
		if !strings.Contains(out, "feature-one") {
			t.Errorf("completion should contain 'feature-one', got: %s", out)
		}
		if !strings.Contains(out, "feature-two") {
			t.Errorf("completion should contain 'feature-two', got: %s", out)
		}
	})

	t.Run("dash_input_returns_flags", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "-")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain flags with descriptions (tab-separated)
		expectedFlags := []string{
			"--delete",
			"--force-delete",
			"--init",
			"--basedir",
			"--copyignored",
			"--copyuntracked",
			"--copymodified",
			"--hook",
			"--deletehook",
			"--remover",
			"--nocopy",
			"--nocd",
			"--relative",
			"-d",
			"-D",
		}

		for _, flag := range expectedFlags {
			if !strings.Contains(out, flag) {
				t.Errorf("completion should contain %q, got: %s", flag, out)
			}
		}

		// Verify tab-separated format (flag\tdescription)
		lines := strings.Split(out, "\n")
		hasTabSeparated := false
		for _, line := range lines {
			if strings.Contains(line, "\t") && strings.HasPrefix(line, "-") {
				hasTabSeparated = true
				break
			}
		}
		if !hasTabSeparated {
			t.Errorf("completion should have tab-separated descriptions, got: %s", out)
		}
	})

	t.Run("double_dash_input_returns_long_flags", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "--")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain long flags
		if !strings.Contains(out, "--delete") {
			t.Errorf("completion should contain '--delete', got: %s", out)
		}
		if !strings.Contains(out, "--basedir") {
			t.Errorf("completion should contain '--basedir', got: %s", out)
		}
	})

	t.Run("branch_completion_has_description", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Branches without worktree should have [branch] description
		lines := strings.Split(out, "\n")
		hasBranchDesc := false
		for _, line := range lines {
			if strings.Contains(line, "[branch]") {
				hasBranchDesc = true
				break
			}
		}
		if !hasBranchDesc {
			t.Errorf("branch completion should have [branch] description, got: %s", out)
		}

		// main branch (worktree) should have [branch: worktree=...] description
		hasWorktreeDesc := false
		for _, line := range lines {
			if strings.HasPrefix(line, "main\t") && strings.Contains(line, "[branch: worktree=") {
				hasWorktreeDesc = true
				break
			}
		}
		if !hasWorktreeDesc {
			t.Errorf("main branch should have [branch: worktree=...] description, got: %s", out)
		}
	})

	t.Run("branch_completion_has_commit_message", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain "initial commit" from the commit message
		if !strings.Contains(out, "initial commit") {
			t.Errorf("branch completion should include commit message, got: %s", out)
		}
	})

	t.Run("worktree_completion_has_description", func(t *testing.T) {
		// Create a worktree - the directory name will be the same as branch name
		_, err := runGitWt(t, binPath, repo.Root, "wt-test-branch")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should have worktree directory with [worktree: branch=...] description
		// Note: when branch name equals directory name, both entries exist
		lines := strings.Split(out, "\n")
		hasWorktreeDirDesc := false
		for _, line := range lines {
			// Check for [worktree: branch=...] (directory entry) or [branch: worktree=...] (branch entry)
			if strings.Contains(line, "[worktree: branch=") || strings.Contains(line, "[branch: worktree=") {
				hasWorktreeDirDesc = true
				break
			}
		}
		if !hasWorktreeDirDesc {
			t.Errorf("worktree completion should have description with worktree info, got: %s", out)
		}
	})
}
