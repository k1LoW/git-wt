package git

import "testing"

func TestNormalizeBranchName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"preserve slash", "someone/main", "someone/main"},
		{"replace colon", "someone:main", "someone/main"},
		{"multiple colons", "a:b:c", "a/b/c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeBranchName(tt.in)
			if got != tt.want {
				t.Errorf("NormalizeBranchName(%q) = %q, want %q", tt.in, got, tt.want) //nostyle:errorstrings
			}
		})
	}
}
