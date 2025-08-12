package nfs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

// TestNfsFile_DefaultsToRemoteName verifies Download uses the remote name when dstFileName is empty.
func TestNfsFile_DefaultsToRemoteName(t *testing.T) {
	// use the recNfs from file_behavior_test.go since it supports writes
	root := "/r"
	sub := "s"
	name := "a.kepub"
	remote := filepath.Join(root, sub, name)

	fake := &recNfs{reads: map[string]string{remote: "X"}}
	folder := &NfsFolder{Folder: root, nfsClient: fake}
	nf := NewNfsFile(root, sub, &nfs4.FileInfo{Name: name, Size: 1}, folder)

	dir := t.TempDir()
	if err := nf.Download(dir, "", true, false, false); err != nil {
		t.Fatalf("download: %v", err)
	}
}

// TestNfsFile_Download_Success verifies a successful download writes the expected file.
func TestNfsFile_Download_Success(t *testing.T) {
	fake := &fakeNfs{reads: map[string]string{"/r/s/a.epub": "content"}}
	folder := &NfsFolder{Folder: "/r/s", nfsClient: fake}
	nf := NewNfsFile("/r", "s", &nfs4.FileInfo{Name: "a.epub", Size: 7}, folder)

	dir := t.TempDir()
	if err := nf.Download(dir, "", true, false, true); err != nil {
		t.Fatalf("download: %v", err)
	}

	// ensure file exists
	if _, err := os.Stat(filepath.Join(dir, "a.epub")); err != nil {
		t.Fatalf("stat: %v", err)
	}
}

// TestNfsFile_Download_KeepFolderStructure ensures subfolder structure is preserved when requested.
func TestNfsFile_Download_KeepFolderStructure(t *testing.T) {
	root := "/nfs"
	sub := "a/b"
	name := "x.epub"
	remote := filepath.Join(root, sub, name)

	fake := &recNfs{reads: map[string]string{remote: "DATA"}}
	folder := &NfsFolder{Folder: root, nfsClient: fake}
	nf := NewNfsFile(root, sub, &nfs4.FileInfo{Name: name, Size: 4}, folder)

	dst := t.TempDir()
	if err := nf.Download(dst, "", true, true, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	// Must include subfolder in destination
	got := filepath.Join(dst, sub, name)
	b, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "DATA" {
		t.Fatalf("content mismatch: %q", string(b))
	}
}

// TestNfsFile_Download_SkipWhenExists_NoOverwrite verifies existing files are not overwritten when flag is false.
func TestNfsFile_Download_SkipWhenExists_NoOverwrite(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "y.epub"
	remote := filepath.Join(root, sub, name)

	fake := &recNfs{reads: map[string]string{remote: "NEW"}}
	folder := &NfsFolder{Folder: root, nfsClient: fake}
	nf := NewNfsFile(root, sub, &nfs4.FileInfo{Name: name, Size: 3}, folder)

	dstDir := t.TempDir()
	dstFile := filepath.Join(dstDir, name)
	if err := os.WriteFile(dstFile, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("precreate: %v", err)
	}

	if err := nf.Download(dstDir, name, false, false, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	b, _ := os.ReadFile(dstFile)
	if string(b) != "OLD" {
		t.Fatalf("expected skip without overwrite, got %q", string(b))
	}
}

// TestNfsFile_Download_OverwriteExisting ensures existing files are overwritten when flag is true.
func TestNfsFile_Download_OverwriteExisting(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "z.epub"
	remote := filepath.Join(root, sub, name)

	fake := &recNfs{reads: map[string]string{remote: "NEW"}}
	folder := &NfsFolder{Folder: root, nfsClient: fake}
	nf := NewNfsFile(root, sub, &nfs4.FileInfo{Name: name, Size: 3}, folder)

	dstDir := t.TempDir()
	dstFile := filepath.Join(dstDir, name)
	if err := os.WriteFile(dstFile, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("precreate: %v", err)
	}

	if err := nf.Download(dstDir, name, true, false, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	b, _ := os.ReadFile(dstFile)
	if string(b) != "NEW" {
		t.Fatalf("expected overwritten content, got %q", string(b))
	}
}

// TestNfsFile_Download_ReadError ensures read errors leave no partial file on disk.
func TestNfsFile_Download_ReadError(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "e.epub"
	remote := filepath.Join(root, sub, name)

	fake := &recNfs{reads: map[string]string{remote: "X"}, readErr: map[string]bool{remote: true}}
	folder := &NfsFolder{Folder: root, nfsClient: fake}
	nf := NewNfsFile(root, sub, &nfs4.FileInfo{Name: name, Size: 1}, folder)

	dstDir := t.TempDir()
	if err := nf.Download(dstDir, name, true, false, false); err == nil {
		t.Fatalf("expected read error")
	}
	// Destination should not exist after failure
	if _, err := os.Stat(filepath.Join(dstDir, name)); !os.IsNotExist(err) {
		t.Fatalf("expected no final file on error")
	}
}

// TestNfsFile_Download_DeleteAfter ensures remote file is deleted when deleteAfter is true.
func TestNfsFile_Download_DeleteAfter(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "d.epub"
	remote := filepath.Join(root, sub, name)

	fake := &recNfs{reads: map[string]string{remote: "D"}}
	folder := &NfsFolder{Folder: root, nfsClient: fake}
	nf := NewNfsFile(root, sub, &nfs4.FileInfo{Name: name, Size: 1}, folder)

	dstDir := t.TempDir()
	if err := nf.Download(dstDir, name, true, false, true); err != nil {
		t.Fatalf("download: %v", err)
	}
	if len(fake.deleted) != 1 || fake.deleted[0] != remote {
		t.Fatalf("expected delete of %q, got %v", remote, fake.deleted)
	}
}
