package git

import "strings"

// NormalizeBranchName normalizes branch names for safe git refs and filesystem usage.
// It treats ":" as a display/input sugar and always replaces it with "/".
func NormalizeBranchName(name string) string {
	return strings.ReplaceAll(name, ":", "/")
}
