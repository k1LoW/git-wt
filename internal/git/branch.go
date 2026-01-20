package git

import (
	"context"
	"os"
	"strings"
)

const gitDefaultBranch = "master"

// BranchExists checks if a branch exists (local or remote).
func BranchExists(ctx context.Context, name string) (bool, error) {
	// Check local branch
	cmd, err := gitCommand(ctx, "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err == nil {
		return true, nil
	}

	// Check remote branch (origin)
	cmd, err = gitCommand(ctx, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+name)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err == nil {
		return true, nil
	}

	return false, nil
}

// LocalBranchExists checks if a local branch exists.
func LocalBranchExists(ctx context.Context, name string) (bool, error) {
	cmd, err := gitCommand(ctx, "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err == nil {
		return true, nil
	}
	return false, nil
}

// CreateBranch creates a new branch at the current HEAD.
func CreateBranch(ctx context.Context, name string) error {
	cmd, err := gitCommand(ctx, "branch", name)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeleteBranch deletes a branch.
// If force is true, it uses -D (force delete), otherwise -d (safe delete).
func DeleteBranch(ctx context.Context, name string, force bool) error {
	return DeleteBranchInDir(ctx, name, force, "")
}

// DeleteBranchInDir deletes a branch from a specific directory.
// If dir is empty, uses current directory.
func DeleteBranchInDir(ctx context.Context, name string, force bool, dir string) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	var args []string
	if dir != "" {
		args = append(args, "-C", dir)
	}
	args = append(args, "branch", flag, name)
	cmd, err := gitCommand(ctx, args...)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsBranchMerged checks if a branch is merged into the current branch.
func IsBranchMerged(ctx context.Context, name string) (bool, error) {
	cmd, err := gitCommand(ctx, "branch", "--merged")
	if err != nil {
		return false, err
	}
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		// Remove leading * and spaces
		branch := strings.TrimSpace(strings.TrimPrefix(line, "*"))
		if branch == name {
			return true, nil
		}
	}
	return false, nil
}

// ListBranches returns a list of all local branch names.
func ListBranches(ctx context.Context) ([]string, error) {
	cmd, err := gitCommand(ctx, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// BranchCommitMessage returns the first line of the latest commit message for a branch.
func BranchCommitMessage(ctx context.Context, branch string) (string, error) {
	cmd, err := gitCommand(ctx, "log", "-1", "--format=%s", branch, "--")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ListRemoteBranches returns a list of all remote branch names (e.g., origin/main).
func ListRemoteBranches(ctx context.Context) ([]string, error) {
	cmd, err := gitCommand(ctx, "branch", "-r", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if line != "" && !strings.HasSuffix(line, "/HEAD") {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// DefaultBranch returns the default branch name (e.g., main, master).
func DefaultBranch(ctx context.Context) (string, error) {
	// Try to get from remote origin
	cmd, err := gitCommand(ctx, "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err == nil {
		// Output is like "origin/main", extract the branch name
		branch := strings.TrimSpace(string(out))
		branch = strings.TrimPrefix(branch, "origin/")
		return branch, nil
	}

	// Fallback: check git config init.defaultBranch
	cmd, err = gitCommand(ctx, "config", "--get", "init.defaultBranch")
	if err != nil {
		return "", err
	}
	out, _ = cmd.Output() //nostyle:handlerrors
	configBranch := strings.TrimSpace(string(out))
	if configBranch == "" {
		// If init.defaultBranch is not set, use Git's built-in default
		configBranch = gitDefaultBranch
	}
	return configBranch, nil
}

// IsDefaultBranch checks if the given branch name is the default branch.
func IsDefaultBranch(ctx context.Context, branch string) (bool, error) {
	defaultBranch, err := DefaultBranch(ctx)
	if err != nil {
		return false, err
	}
	if defaultBranch == "" {
		return false, nil
	}
	return branch == defaultBranch, nil
}
