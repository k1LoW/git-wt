package git

import (
	"context"

	"github.com/k1LoW/exec"
)

// gitCommand creates an exec.Cmd for git with the given context and arguments.
// It uses exec.LookPath to look up the git binary path.
func gitCommand(ctx context.Context, args ...string) (*exec.Cmd, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, gitPath, args...), nil
}
