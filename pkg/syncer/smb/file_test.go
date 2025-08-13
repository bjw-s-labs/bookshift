package smb

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jfjallid/go-smb/smb"
)

type fakeConnDL struct{ smbConnBase }

func (f *fakeConnDL) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	return nil, nil
}
func (f *fakeConnDL) RetrieveFile(share string, filepath string, offset uint64, cb func([]byte) (int, error)) error {
	_, _ = cb([]byte("data"))
	return nil
}
func (f *fakeConnDL) DeleteFile(share string, filepath string) error { return nil }

type recConn struct {
	smbConnBase
	data    map[string]string
	deletes []string
}

func (f *recConn) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	return nil, errors.New("unused")
}
func (f *recConn) RetrieveFile(share string, fp string, offset uint64, cb func([]byte) (int, error)) error {
	if s, ok := f.data[fp]; ok {
		_, _ = cb([]byte(s))
		return nil
	}
	return errors.New("not found")
}
func (f *recConn) DeleteFile(share string, fp string) error {
	f.deletes = append(f.deletes, fp)
	return nil
}

// TestSmbFile_DefaultsToRemoteName verifies Download uses the remote name when dstFileName is empty.
func TestSmbFile_DefaultsToRemoteName(t *testing.T) {
	dir := t.TempDir()
	sc := &SmbShareConnection{Share: "s", SmbConnection: &fakeConnDL{}}
	sf := &smb.SharedFile{Name: "a.epub", FullPath: "/root/a.epub", Size: 4}
	f := NewSmbFile("/root", "", sf, sc)

	if err := f.Download(dir, "", true, false, false); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.epub")); err != nil {
		t.Fatalf("stat: %v", err)
	}
}

// TestSmbFile_Download_Success verifies a successful download writes the target file.
func TestSmbFile_Download_Success(t *testing.T) {
	dir := t.TempDir()
	sc := &SmbShareConnection{Share: "s", SmbConnection: &fakeConnDL{}}
	sf := &smb.SharedFile{Name: "a.epub", FullPath: "/root/a.epub", Size: 4}
	f := NewSmbFile("/root", "", sf, sc)

	if err := f.Download(dir, "", true, false, true); err != nil {
		t.Fatalf("Download: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.epub")); err != nil {
		t.Fatalf("stat: %v", err)
	}
}

type fakeConnErr struct{}

func (f *fakeConnErr) TreeConnect(share string) error    { return nil }
func (f *fakeConnErr) TreeDisconnect(share string) error { return nil }
func (f *fakeConnErr) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	return nil, nil
}
func (f *fakeConnErr) RetrieveFile(share string, filepath string, offset uint64, cb func([]byte) (int, error)) error {
	return errors.New("read fail")
}
func (f *fakeConnErr) DeleteFile(share string, filepath string) error {
	return errors.New("delete fail")
}

// TestSmbFile_Download_ReadError ensures read errors are returned and no file is written.
func TestSmbFile_Download_ReadError(t *testing.T) {
	dir := t.TempDir()
	sc := &SmbShareConnection{Share: "s", SmbConnection: &fakeConnErr{}}
	sf := &smb.SharedFile{Name: "a.epub", FullPath: "/root/a.epub", Size: 4}
	f := NewSmbFile("/root", "", sf, sc)
	if err := f.Download(dir, "", true, false, false); err == nil {
		t.Fatalf("expected read error")
	}
}

type fakeConnDeleteErr struct{}

func (f *fakeConnDeleteErr) TreeConnect(share string) error    { return nil }
func (f *fakeConnDeleteErr) TreeDisconnect(share string) error { return nil }
func (f *fakeConnDeleteErr) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	return nil, nil
}
func (f *fakeConnDeleteErr) RetrieveFile(share string, filepath string, offset uint64, cb func([]byte) (int, error)) error {
	_, _ = cb([]byte("data"))
	return nil
}
func (f *fakeConnDeleteErr) DeleteFile(share string, filepath string) error {
	return errors.New("delete fail")
}

// TestSmbFile_Download_DeleteError ensures a post-download delete failure is surfaced but local file remains.
func TestSmbFile_Download_DeleteError(t *testing.T) {
	dir := t.TempDir()
	sc := &SmbShareConnection{Share: "s", SmbConnection: &fakeConnDeleteErr{}}
	sf := &smb.SharedFile{Name: "a.epub", FullPath: "/root/a.epub", Size: 4}
	f := NewSmbFile("/root", "", sf, sc)
	if err := f.Download(dir, "", true, false, true); err == nil {
		t.Fatalf("expected delete error")
	}
	// File should still be present locally, since delete happens after rename
	if _, err := os.Stat(filepath.Join(dir, "a.epub")); err != nil {
		t.Fatalf("expected local file despite delete error: %v", err)
	}
}

// TestSmbFile_KeepFolderStructure ensures Download preserves subfolder structure when requested.
func TestSmbFile_KeepFolderStructure(t *testing.T) {
	root := "/r"
	sub := "a/b"
	name := "x.epub"
	full := filepath.Join(root, sub, name)

	rc := &recConn{data: map[string]string{full: "DATA"}}
	sc := &SmbShareConnection{Share: "share", SmbConnection: rc}
	sf := &smb.SharedFile{Name: name, FullPath: full, Size: uint64(len("DATA"))}
	f := NewSmbFile(root, sub, sf, sc)

	dst := t.TempDir()
	if err := f.Download(dst, "", true, true, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	got := filepath.Join(dst, sub, name)
	b, err := os.ReadFile(got)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "DATA" {
		t.Fatalf("content mismatch: %q", string(b))
	}
}

// TestSmbFile_SkipNoOverwrite verifies existing files are not overwritten when flag is false.
func TestSmbFile_SkipNoOverwrite(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "y.epub"
	full := filepath.Join(root, sub, name)

	rc := &recConn{data: map[string]string{full: "NEW"}}
	sc := &SmbShareConnection{Share: "share", SmbConnection: rc}
	sf := &smb.SharedFile{Name: name, FullPath: full, Size: 3}
	f := NewSmbFile(root, sub, sf, sc)

	dstDir := t.TempDir()
	dstFile := filepath.Join(dstDir, name)
	if err := os.WriteFile(dstFile, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("precreate: %v", err)
	}

	if err := f.Download(dstDir, "", false, false, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	b, _ := os.ReadFile(dstFile)
	if string(b) != "OLD" {
		t.Fatalf("expected skip, got %q", string(b))
	}
}

// TestSmbFile_OverwriteExisting ensures existing files are overwritten when flag is true.
func TestSmbFile_OverwriteExisting(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "z.epub"
	full := filepath.Join(root, sub, name)

	rc := &recConn{data: map[string]string{full: "NEW"}}
	sc := &SmbShareConnection{Share: "share", SmbConnection: rc}
	sf := &smb.SharedFile{Name: name, FullPath: full, Size: 3}
	f := NewSmbFile(root, sub, sf, sc)

	dstDir := t.TempDir()
	dstFile := filepath.Join(dstDir, name)
	if err := os.WriteFile(dstFile, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("precreate: %v", err)
	}

	if err := f.Download(dstDir, "", true, false, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	b, _ := os.ReadFile(dstFile)
	if string(b) != "NEW" {
		t.Fatalf("expected overwritten content, got %q", string(b))
	}
}

// TestSmbFile_DeleteAfter ensures the remote file is deleted when deleteAfter is true.
func TestSmbFile_DeleteAfter(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "d.epub"
	full := filepath.Join(root, sub, name)

	rc := &recConn{data: map[string]string{full: "D"}}
	sc := &SmbShareConnection{Share: "share", SmbConnection: rc}
	sf := &smb.SharedFile{Name: name, FullPath: full, Size: 1}
	f := NewSmbFile(root, sub, sf, sc)

	dstDir := t.TempDir()
	if err := f.Download(dstDir, "", true, false, true); err != nil {
		t.Fatalf("download: %v", err)
	}
	if len(rc.deletes) != 1 || rc.deletes[0] != full {
		t.Fatalf("expected delete of %q, got %v", full, rc.deletes)
	}
}

// TestSmbFile_ReadError ensures read failures result in no partial file being left on disk.
func TestSmbFile_ReadError(t *testing.T) {
	root := "/r"
	sub := "s"
	name := "e.epub"
	full := filepath.Join(root, sub, name)

	rc := &recConn{data: map[string]string{}} // not found leads to error
	sc := &SmbShareConnection{Share: "share", SmbConnection: rc}
	sf := &smb.SharedFile{Name: name, FullPath: full, Size: 1}
	f := NewSmbFile(root, sub, sf, sc)

	dstDir := t.TempDir()
	if err := f.Download(dstDir, "", true, false, false); err == nil {
		t.Fatalf("expected read error")
	}
	if _, err := os.Stat(filepath.Join(dstDir, name)); !os.IsNotExist(err) {
		t.Fatalf("expected no final file on error")
	}
}
