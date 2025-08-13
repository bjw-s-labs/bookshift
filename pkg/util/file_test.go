package util

import "testing"

// TestSafeFileName ensures that SafeFileName normalizes paths and filenames by removing
// unsafe characters, collapsing whitespace/dashes, and lowercasing extensions.
func TestSafeFileName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Hello World.epub", "hello-world.epub"},
		{"../../weird\\path?.txt", "weirdpath.txt"},
		{"multi--dash  name .kepub", "multi-dash-name-.kepub"},
		{"UPPER_case+Name.PDF", "upper-case-name.pdf"},
	}
	for _, tt := range tests {
		if got := SafeFileName(tt.in); got != tt.want {
			t.Fatalf("SafeFileName(%q)=%q, want %q", tt.in, got, tt.want)
		}
	}
}
