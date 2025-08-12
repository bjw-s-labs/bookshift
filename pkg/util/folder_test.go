package util

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCountFilesInFolder ensures CountFilesInFolder traverses the provided
// directory recursively and counts only files with allowed extensions.
func TestCountFilesInFolder(t *testing.T) {
	dir := t.TempDir()
	// Create files
	mustWrite := func(name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}
	mustWrite("a.epub")
	mustWrite("b.txt")
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "c.kepub"), []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cnt, err := CountFilesInFolder(dir, []string{".epub", ".kepub"}, true)
	if err != nil {
		t.Fatalf("CountFilesInFolder error: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("CountFilesInFolder=%d, want 2", cnt)
	}
}
