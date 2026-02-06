package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/exec"
)

const (
	configKeyBaseDir       = "wt.basedir"
	configKeyCopyIgnored   = "wt.copyignored"
	configKeyCopyUntracked = "wt.copyuntracked"
	configKeyCopyModified  = "wt.copymodified"
	configKeyNoCopy        = "wt.nocopy"
	configKeyCopy          = "wt.copy"
	configKeyHook          = "wt.hook"
	configKeyNoCd          = "wt.nocd"
	configKeyRelative      = "wt.relative"
)

// Config holds all wt configuration values.
type Config struct {
	BaseDir       string
	CopyIgnored   bool
	CopyUntracked bool
	CopyModified  bool
	NoCopy        []string
	Copy          []string
	Hooks         []string
	NoCd          bool
	Relative      bool
}

// GitConfig retrieves all git config values for a key.
func GitConfig(ctx context.Context, key string) ([]string, error) { //nolint:revive //nostyle:repetition
	cmd, err := gitCommand(ctx, "config", "--get-all", key)
	if err != nil {
		return nil, err
	}
	out, err := cmd.Output()
	if err != nil {
		// git config returns exit code 1 if key is not found
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

// ShowPrefix returns the path prefix of the current directory relative to the repository root.
// It runs "git rev-parse --show-prefix" and strips the trailing slash.
// Returns an empty string when at the repository root.
func ShowPrefix(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--show-prefix")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSpace(string(out)), "/"), nil
}

// RepoRoot returns the root directory of the current git repository (or worktree).
func RepoRoot(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// MainRepoRoot returns the root directory of the main git repository.
// Unlike RepoRoot, this returns the main repository root even when called from a worktree.
func MainRepoRoot(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	gitCommonDir := strings.TrimSpace(string(out))

	// The main repo root is the parent of the .git directory
	return filepath.Dir(gitCommonDir), nil
}

// RepoName returns the name of the current git repository (directory name).
func RepoName(ctx context.Context) (string, error) {
	root, err := MainRepoRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}

// LoadConfig loads configuration from git config with default values.
func LoadConfig(ctx context.Context) (Config, error) {
	cfg := Config{}

	// BaseDir
	baseDir, err := GitConfig(ctx, configKeyBaseDir)
	if err != nil {
		return cfg, err
	}
	if len(baseDir) == 0 {
		cfg.BaseDir = ".wt"
	} else {
		cfg.BaseDir = baseDir[len(baseDir)-1]
	}

	// CopyIgnored
	val, err := GitConfig(ctx, configKeyCopyIgnored)
	if err != nil {
		return cfg, err
	}
	cfg.CopyIgnored = len(val) > 0 && val[len(val)-1] == "true"

	// CopyUntracked
	val, err = GitConfig(ctx, configKeyCopyUntracked)
	if err != nil {
		return cfg, err
	}
	cfg.CopyUntracked = len(val) > 0 && val[len(val)-1] == "true"

	// CopyModified
	val, err = GitConfig(ctx, configKeyCopyModified)
	if err != nil {
		return cfg, err
	}
	cfg.CopyModified = len(val) > 0 && val[len(val)-1] == "true"

	// NoCopy
	noCopy, err := GitConfig(ctx, configKeyNoCopy)
	if err != nil {
		return cfg, err
	}
	cfg.NoCopy = noCopy

	// Copy
	copyPatterns, err := GitConfig(ctx, configKeyCopy)
	if err != nil {
		return cfg, err
	}
	cfg.Copy = copyPatterns

	// Hooks
	hooks, err := GitConfig(ctx, configKeyHook)
	if err != nil {
		return cfg, err
	}
	cfg.Hooks = hooks

	// NoCd
	val, err = GitConfig(ctx, configKeyNoCd)
	if err != nil {
		return cfg, err
	}
	cfg.NoCd = len(val) > 0 && val[len(val)-1] == "true"

	// Relative
	val, err = GitConfig(ctx, configKeyRelative)
	if err != nil {
		return cfg, err
	}
	cfg.Relative = len(val) > 0 && val[len(val)-1] == "true"

	return cfg, nil
}

// expandTemplate expands template variables in a string.
// Supported variables:
//   - {gitroot}: repository root directory name
func expandTemplate(ctx context.Context, s string) (string, error) {
	// Expand {gitroot}
	if strings.Contains(s, "{gitroot}") {
		repoName, err := RepoName(ctx)
		if err != nil {
			return "", err
		}
		s = strings.ReplaceAll(s, "{gitroot}", repoName)
	}

	return s, nil
}

// ExpandPath expands ~ to home directory and resolves relative paths.
// Relative paths are resolved from the main repository root, not the current worktree.
func ExpandPath(ctx context.Context, path string) (string, error) {
	// Expand ~
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	} else if path == "~" {
		return os.UserHomeDir()
	}

	// If already absolute, return as is
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	// Resolve relative path from main repo root (not current worktree)
	repoRoot, err := MainRepoRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(repoRoot, path)), nil
}

// ExpandBaseDir expands template variables and path for the given base directory pattern.
func ExpandBaseDir(ctx context.Context, baseDir string) (string, error) {
	// Expand template variables
	expanded, err := expandTemplate(ctx, baseDir)
	if err != nil {
		return "", err
	}

	// Expand path (~ and relative paths)
	expanded, err = ExpandPath(ctx, expanded)
	if err != nil {
		return "", err
	}

	return expanded, nil
}

// IsBaseDirConfigured checks if wt.basedir is explicitly configured in git config.
func IsBaseDirConfigured(ctx context.Context) (bool, error) {
	baseDir, err := GitConfig(ctx, configKeyBaseDir)
	if err != nil {
		return false, err
	}
	return len(baseDir) > 0, nil
}

// SetConfig sets a git config value.
func SetConfig(ctx context.Context, key, value string) error {
	cmd, err := gitCommand(ctx, "config", key, value)
	if err != nil {
		return err
	}
	return cmd.Run()
}

// WorktreePathFor returns the full path for a worktree given a base directory pattern and branch name.
func WorktreePathFor(ctx context.Context, baseDir, branch string) (string, error) {
	expandedBaseDir, err := ExpandBaseDir(ctx, baseDir)
	if err != nil {
		return "", err
	}

	return filepath.Join(expandedBaseDir, branch), nil
}
