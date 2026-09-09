// shell_test.go contains shell integration tests:
//   - TestE2E_InitScript: --init script generation (bash/zsh/fish/powershell, nocd, unsupported_shell)
//   - TestE2E_ShellIntegration_StdoutFormat: stdout format for shell integration compatibility
//   - TestE2E_ShellIntegration: shell integration cd tests (bash, zsh, fish, powershell, nocd, hook_failure)
//     Note: the PowerShell subtests only run on Windows; see the windows job in .github/workflows/ci.yml
package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
	"github.com/k1LoW/git-wt/testutil"
)

func TestE2E_InitScript(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("bash", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--init", "bash")
		if err != nil {
			t.Fatalf("git-wt --init bash failed: %v\noutput: %s", err, out)
		}

		contains := []string{"# git-wt shell hook for bash", "_git_wt()"}
		for _, s := range contains {
			if !strings.Contains(out, s) {
				t.Errorf("output should contain %q, got: %s", s, out)
			}
		}
	})

	t.Run("zsh", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--init", "zsh")
		if err != nil {
			t.Fatalf("git-wt --init zsh failed: %v\noutput: %s", err, out)
		}

		contains := []string{"# git-wt shell hook for zsh", "_git-wt()"}
		for _, s := range contains {
			if !strings.Contains(out, s) {
				t.Errorf("output should contain %q, got: %s", s, out)
			}
		}
	})

	t.Run("fish", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--init", "fish")
		if err != nil {
			t.Fatalf("git-wt --init fish failed: %v\noutput: %s", err, out)
		}

		contains := []string{"# git-wt shell hook for fish", "function git --wraps git"}
		for _, s := range contains {
			if !strings.Contains(out, s) {
				t.Errorf("output should contain %q, got: %s", s, out)
			}
		}
	})

	t.Run("powershell", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--init", "powershell")
		if err != nil {
			t.Fatalf("git-wt --init powershell failed: %v\noutput: %s", err, out)
		}

		contains := []string{"# git-wt shell hook for PowerShell", "Invoke-Git"}
		for _, s := range contains {
			if !strings.Contains(out, s) {
				t.Errorf("output should contain %q, got: %s", s, out)
			}
		}
	})

	t.Run("nocd", func(t *testing.T) {
		t.Parallel()
		out, err := runGitWt(t, binPath, t.TempDir(), "--init", "bash", "--nocd")
		if err != nil {
			t.Fatalf("git-wt --init bash --nocd failed: %v\noutput: %s", err, out)
		}

		// Should not contain the git wrapper function
		if strings.Contains(out, "git() {") {
			t.Error("output should not contain git wrapper when --nocd is used")
		}

		// Should still contain completion
		if !strings.Contains(out, "_git_wt()") {
			t.Error("output should contain completion function")
		}
	})

	t.Run("unsupported_shell", func(t *testing.T) {
		t.Parallel()
		_, err := runGitWt(t, binPath, t.TempDir(), "--init", "unsupported")
		if err == nil {
			t.Error("expected error for unsupported shell")
		}
	})
}

// TestE2E_ShellIntegration_StdoutFormat tests that git-wt output is compatible
// with shell integration (stdout contains only the path, suitable for cd).
func TestE2E_ShellIntegration_StdoutFormat(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("list_worktrees_stdout_is_not_directory", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		stdout, _, err := runGitWtStdout(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("git-wt failed: %v", err)
		}

		// List output should NOT be a valid directory path
		// (it's a table, so shell integration should not cd)
		info, err := os.Stat(stdout)
		if err == nil && info.IsDir() {
			t.Errorf("list output should not be a valid directory, got: %s", stdout)
		}
	})

	t.Run("create_worktree_stdout_is_directory", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "feature-shell")
		if err != nil {
			t.Fatalf("git-wt feature-shell failed: %v\nstderr: %s", err, stderr)
		}

		// stdout should be exactly one line (the path)
		lines := strings.Split(stdout, "\n")
		if len(lines) != 1 {
			t.Errorf("stdout should be exactly 1 line, got %d lines: %q", len(lines), stdout)
		}

		// stdout should be a valid directory
		info, err := os.Stat(stdout)
		if err != nil {
			t.Errorf("stdout path does not exist: %v", err)
		} else if !info.IsDir() {
			t.Errorf("stdout should be a directory, got: %s", stdout)
		}

		// stderr should contain git messages (not empty for new worktree)
		if stderr == "" {
			t.Log("warning: stderr is empty (git worktree add usually outputs to stderr)")
		}
	})

	t.Run("switch_worktree_stdout_is_directory", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree first
		_, _, err := runGitWtStdout(t, binPath, repo.Root, "existing-wt")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Switch to existing worktree
		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "existing-wt")
		if err != nil {
			t.Fatalf("git-wt existing-wt failed: %v\nstderr: %s", err, stderr)
		}

		// stdout should be exactly one line
		lines := strings.Split(stdout, "\n")
		if len(lines) != 1 {
			t.Errorf("stdout should be exactly 1 line, got %d lines: %q", len(lines), stdout)
		}

		// stdout should be a valid directory
		info, err := os.Stat(stdout)
		if err != nil {
			t.Errorf("stdout path does not exist: %v", err)
		} else if !info.IsDir() {
			t.Errorf("stdout should be a directory, got: %s", stdout)
		}

		// stderr should be empty for existing worktree (no git operation)
		if stderr != "" {
			t.Logf("note: stderr is not empty for existing worktree: %s", stderr)
		}
	})

	t.Run("delete_worktree_stdout_is_not_directory", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree first
		_, _, err := runGitWtStdout(t, binPath, repo.Root, "to-delete-shell")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Delete worktree
		stdout, _, err := runGitWtStdout(t, binPath, repo.Root, "-d", "to-delete-shell")
		if err != nil {
			t.Fatalf("git-wt -d failed: %v", err)
		}

		// Delete output should NOT be a valid directory
		// (it's a message, so shell integration should not cd)
		info, err := os.Stat(stdout)
		if err == nil && info.IsDir() {
			t.Errorf("delete output should not be a valid directory, got: %s", stdout)
		}
	})
}

// TestE2E_ShellIntegration tests the actual shell integration with various shells.
func TestE2E_ShellIntegration(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("bash", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("bash"); err != nil {
			t.Skip("bash not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Test that shell integration works: eval the init script and run git wt
		script := fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Test: git wt <branch> should cd to the worktree
git wt shell-bash-test
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("bash", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("bash shell integration failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		// The last line should be the worktree path
		lines := strings.Split(output, "\n")
		pwd := lines[len(lines)-1]

		if !strings.Contains(pwd, "shell-bash-test") {
			t.Errorf("pwd should contain worktree path, got: %s", pwd)
		}
	})

	t.Run("zsh", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("zsh"); err != nil {
			t.Skip("zsh not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Test: git wt <branch> should cd to the worktree
git wt shell-zsh-test
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("zsh", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("zsh shell integration failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		pwd := lines[len(lines)-1]

		if !strings.Contains(pwd, "shell-zsh-test") {
			t.Errorf("pwd should contain worktree path, got: %s", pwd)
		}
	})

	t.Run("fish", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("fish"); err != nil {
			t.Skip("fish not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Test: git wt <branch> should cd to the worktree
git wt shell-fish-test
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("fish", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fish shell integration failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		pwd := lines[len(lines)-1]

		if !strings.Contains(pwd, "shell-fish-test") {
			t.Errorf("pwd should contain worktree path, got: %s", pwd)
		}
	})

	t.Run("powershell", func(t *testing.T) {
		t.Parallel()
		// PowerShell init script uses git.exe which is Windows-specific
		if runtime.GOOS != "windows" {
			t.Skip("PowerShell shell integration test is only supported on Windows")
		}

		// Try pwsh first (cross-platform), then powershell (Windows)
		var pwshPath string
		if p, err := exec.LookPath("pwsh"); err == nil {
			pwshPath = p
		} else if p, err := exec.LookPath("powershell"); err == nil {
			pwshPath = p
		} else {
			t.Skip("PowerShell not available")
		}

		binPathLocal := binPath
		// On Windows, binary needs .exe extension
		if runtime.GOOS == "windows" && !strings.HasSuffix(binPathLocal, ".exe") {
			binPathLocal += ".exe"
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
$ErrorActionPreference = "Stop"
Set-Location %q
$env:PATH = %q + [IO.Path]::PathSeparator + $env:PATH
Invoke-Expression (git wt --init powershell | Out-String)

# Test: git wt <branch> should cd to the worktree
git wt shell-pwsh-test
Get-Location | Select-Object -ExpandProperty Path
`, repo.Root, filepath.Dir(binPathLocal))

		cmd := exec.Command(pwshPath, "-NoProfile", "-Command", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("PowerShell shell integration failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		pwd := strings.TrimSpace(lines[len(lines)-1])

		if !strings.Contains(pwd, "shell-pwsh-test") {
			t.Errorf("pwd should contain worktree path, got: %s", pwd)
		}
	})

	t.Run("nocd_bash", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("bash"); err != nil {
			t.Skip("bash not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Test: git wt --nocd <branch> should NOT cd to the worktree
git wt --nocd nocd-bash-test
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("bash", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("bash shell integration with --nocd failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		pwd := lines[len(lines)-1]

		// pwd should be the original repo root, NOT the new worktree
		if strings.Contains(pwd, "nocd-bash-test") {
			t.Errorf("pwd should NOT contain worktree path when --nocd is used, got: %s", pwd)
		}
		if pwd != repo.Root {
			t.Errorf("pwd should be original repo root %q, got: %s", repo.Root, pwd)
		}
	})

	t.Run("nocd_zsh", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("zsh"); err != nil {
			t.Skip("zsh not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Test: git wt --nocd <branch> should NOT cd to the worktree
git wt --nocd nocd-zsh-test
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("zsh", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("zsh shell integration with --nocd failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		pwd := lines[len(lines)-1]

		// pwd should be the original repo root, NOT the new worktree
		if strings.Contains(pwd, "nocd-zsh-test") {
			t.Errorf("pwd should NOT contain worktree path when --nocd is used, got: %s", pwd)
		}
		if pwd != repo.Root {
			t.Errorf("pwd should be original repo root %q, got: %s", repo.Root, pwd)
		}
	})

	t.Run("nocd_fish", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("fish"); err != nil {
			t.Skip("fish not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Test: git wt --nocd <branch> should NOT cd to the worktree
git wt --nocd nocd-fish-test
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("fish", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fish shell integration with --nocd failed: %v\noutput: %s", err, out)
		}

		output := strings.TrimSpace(string(out))
		lines := strings.Split(output, "\n")
		pwd := lines[len(lines)-1]

		// pwd should be the original repo root, NOT the new worktree
		if strings.Contains(pwd, "nocd-fish-test") {
			t.Errorf("pwd should NOT contain worktree path when --nocd is used, got: %s", pwd)
		}
		if pwd != repo.Root {
			t.Errorf("pwd should be original repo root %q, got: %s", repo.Root, pwd)
		}
	})

	t.Run("hook_failure_bash", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("bash"); err != nil {
			t.Skip("bash not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Test: a failing hook should still cd to the created worktree, and should still
# report a non-zero exit code
git wt --hook "exit 3" hookfail-bash-test 2>/dev/null
echo "exit=$?"
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("bash", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("bash shell integration with failing hook failed: %v\noutput: %s", err, out)
		}

		assertHookFailureCd(t, string(out), "hookfail-bash-test")
	})

	t.Run("hook_failure_zsh", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("zsh"); err != nil {
			t.Skip("zsh not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Test: a failing hook should still cd to the created worktree, and should still
# report a non-zero exit code
git wt --hook "exit 3" hookfail-zsh-test 2>/dev/null
echo "exit=$?"
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("zsh", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("zsh shell integration with failing hook failed: %v\noutput: %s", err, out)
		}

		assertHookFailureCd(t, string(out), "hookfail-zsh-test")
	})

	t.Run("hook_failure_fish", func(t *testing.T) {
		t.Parallel()
		if _, err := exec.LookPath("fish"); err != nil {
			t.Skip("fish not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Test: a failing hook should still cd to the created worktree, and should still
# report a non-zero exit code
git wt --hook "exit 3" hookfail-fish-test 2>/dev/null
echo "exit=$status"
pwd
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command("fish", "-c", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("fish shell integration with failing hook failed: %v\noutput: %s", err, out)
		}

		assertHookFailureCd(t, string(out), "hookfail-fish-test")
	})

	t.Run("hook_failure_powershell", func(t *testing.T) {
		t.Parallel()
		// PowerShell init script uses git.exe which is Windows-specific
		if runtime.GOOS != "windows" {
			t.Skip("PowerShell shell integration test is only supported on Windows")
		}

		var pwshPath string
		if p, err := exec.LookPath("pwsh"); err == nil {
			pwshPath = p
		} else if p, err := exec.LookPath("powershell"); err == nil {
			pwshPath = p
		} else {
			t.Skip("PowerShell not available")
		}

		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		script := fmt.Sprintf(`
Set-Location %q
$env:PATH = %q + [IO.Path]::PathSeparator + $env:PATH
Invoke-Expression (git wt --init powershell | Out-String)

# Test: a failing hook should still cd to the created worktree, and should still
# report a non-zero exit code in $LASTEXITCODE.
# $? is not asserted. A PowerShell function call reports success to its caller
# regardless of what failed inside it, so the wrapper cannot drive $? and && does
# not stop after a failed hook. $LASTEXITCODE is the signal callers get.
git wt --hook "exit 3" hookfail-pwsh-test
$code = $LASTEXITCODE
Write-Output "exit=$code"
Get-Location | Select-Object -ExpandProperty Path
`, repo.Root, filepath.Dir(binPath))

		cmd := exec.Command(pwshPath, "-NoProfile", "-Command", script)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("PowerShell shell integration with failing hook failed: %v\noutput: %s", err, out)
		}

		assertHookFailureCd(t, string(out), "hookfail-pwsh-test")
	})
}

// assertHookFailureCd asserts that the shell wrapper changed directory into the
// newly created worktree while keeping the non-zero exit code of the failed hook.
// The script under test prints "exit=<code>" followed by pwd as its last two lines.
func assertHookFailureCd(t *testing.T, out, wtName string) {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("output should have at least 2 lines (exit code and pwd), got: %q", out)
	}
	exitLine := strings.TrimSpace(lines[len(lines)-2])
	pwd := strings.TrimSpace(lines[len(lines)-1])

	if !strings.HasPrefix(exitLine, "exit=") {
		t.Fatalf("second to last line should be the exit code, got: %q", exitLine)
	}
	if exitLine == "exit=0" {
		t.Error("exit code should stay non-zero when a hook fails")
	}
	if !strings.Contains(pwd, wtName) {
		t.Errorf("pwd should contain worktree path even though the hook failed, got: %s", pwd)
	}
}
