package git

import (
	"os/exec"
)

// gitCommand creates an exec.Cmd for git with the given arguments.
// It uses exec.LookPath to look up the git binary path.
func gitCommand(args ...string) (*exec.Cmd, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	return exec.Command(gitPath, args...), nil
}
