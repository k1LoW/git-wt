package git

import (
	"os"
	"strings"
)

// BranchExists checks if a branch exists (local or remote).
func BranchExists(name string) (bool, error) {
	// Check local branch
	cmd, err := gitCommand("show-ref", "--verify", "--quiet", "refs/heads/"+name)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err == nil {
		return true, nil
	}

	// Check remote branch (origin)
	cmd, err = gitCommand("show-ref", "--verify", "--quiet", "refs/remotes/origin/"+name)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err == nil {
		return true, nil
	}

	return false, nil
}

// LocalBranchExists checks if a local branch exists.
func LocalBranchExists(name string) (bool, error) {
	cmd, err := gitCommand("show-ref", "--verify", "--quiet", "refs/heads/"+name)
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err == nil {
		return true, nil
	}
	return false, nil
}

// CreateBranch creates a new branch at the current HEAD.
func CreateBranch(name string) error {
	cmd, err := gitCommand("branch", name)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DeleteBranch deletes a branch.
// If force is true, it uses -D (force delete), otherwise -d (safe delete).
func DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	cmd, err := gitCommand("branch", flag, name)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// IsBranchMerged checks if a branch is merged into the current branch.
func IsBranchMerged(name string) (bool, error) {
	cmd, err := gitCommand("branch", "--merged")
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
func ListBranches() ([]string, error) {
	cmd, err := gitCommand("branch", "--format=%(refname:short)")
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

// GetDefaultBranch returns the default branch name (e.g., main, master).
func GetDefaultBranch() (string, error) {
	// Try to get from remote origin
	cmd, err := gitCommand("symbolic-ref", "refs/remotes/origin/HEAD", "--short")
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

	// Fallback: check common default branch names
	for _, name := range []string{"main", "master"} {
		exists, err := LocalBranchExists(name)
		if err != nil {
			continue
		}
		if exists {
			return name, nil
		}
	}

	return "", nil
}
