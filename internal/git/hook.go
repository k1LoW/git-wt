package git

import (
	"context"
	"fmt"
	"io"

	"github.com/k1LoW/exec"
)

// RunHooks executes the configured hooks in the given directory.
// Hook stdout/stderr are written to the provided writer.
// If a hook fails, it stops immediately and returns the error.
func RunHooks(ctx context.Context, hooks []string, dir string, w io.Writer) error {
	for _, hook := range hooks {
		cmd := exec.CommandContext(ctx, "sh", "-c", hook)
		cmd.Dir = dir
		cmd.Stdout = w
		cmd.Stderr = w
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("hook %q failed: %w", hook, err)
		}
	}
	return nil
}
