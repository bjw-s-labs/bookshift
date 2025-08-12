package smb

import (
	"errors"
	"testing"

	"github.com/jfjallid/go-smb/smb"
)

type fakeConn struct {
	smbConnBase
	files map[string][]smb.SharedFile
}

func (f *fakeConn) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	if l, ok := f.files[subfolder]; ok {
		return l, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeConn) RetrieveFile(share string, filepath string, offset uint64, cb func([]byte) (int, error)) error {
	return nil
}
func (f *fakeConn) DeleteFile(share string, filepath string) error { return nil }

// TestSmbShareConnectionFetchFiles_Filter ensures FetchFiles filters by extension.
func TestSmbShareConnectionFetchFiles_Filter(t *testing.T) {
	fc := &fakeConn{
		files: map[string][]smb.SharedFile{
			"/root": {{FullPath: "/root/a.txt", Name: "a.txt"}, {FullPath: "/root/b.epub", Name: "b.epub"}},
		},
	}
	sc := NewSmbShareConnection("share", fc)
	files, err := sc.FetchFiles("/root", []string{".epub"}, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1, got %d", len(files))
	}
}

type fakeConnConn struct {
	smbConnBase
	treeErr error
}

func (f *fakeConnConn) TreeConnect(share string) error { return f.treeErr }
func (f *fakeConnConn) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	return nil, nil
}
func (f *fakeConnConn) RetrieveFile(share string, filepath string, offset uint64, cb func([]byte) (int, error)) error {
	return nil
}
func (f *fakeConnConn) DeleteFile(share string, filepath string) error { return nil }

// TestSmbShareConnection_Connect_ShareNotFound ensures a missing share returns the proper error.
func TestSmbShareConnection_Connect_ShareNotFound(t *testing.T) {
	sc := NewSmbShareConnection("missing", &fakeConnConn{treeErr: smb.StatusMap[smb.StatusBadNetworkName]})
	if err := sc.Connect(); err == nil {
		t.Fatalf("expected share not found error")
	}
}

type fakeConnList struct {
	smbConnBase
	files map[string][]smb.SharedFile
	errs  map[string]error
}

func (f *fakeConnList) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	if e, ok := f.errs[subfolder]; ok {
		return nil, e
	}
	if l, ok := f.files[subfolder]; ok {
		return l, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeConnList) RetrieveFile(share string, filepath string, offset uint64, cb func([]byte) (int, error)) error {
	return nil
}
func (f *fakeConnList) DeleteFile(share string, filepath string) error { return nil }

// TestSmbShareConnection_FetchFiles_AccessDenied verifies access denied is surfaced.
func TestSmbShareConnection_FetchFiles_AccessDenied(t *testing.T) {
	fc := &fakeConnList{errs: map[string]error{"/root": smb.StatusMap[smb.StatusAccessDenied]}}
	sc := NewSmbShareConnection("share", fc)
	if _, err := sc.FetchFiles("/root", nil, true); err == nil {
		t.Fatalf("expected access denied error")
	}
}

// TestSmbShareConnection_Recurse_ChildErrorSkips ensures errors in subfolders don't stop recursion.
func TestSmbShareConnection_Recurse_ChildErrorSkips(t *testing.T) {
	fc := &fakeConnList{
		files: map[string][]smb.SharedFile{
			"/root": {{FullPath: "/root/dir", Name: "dir", IsDir: true}, {FullPath: "/root/f.epub", Name: "f.epub"}},
		},
		errs: map[string]error{"/root/dir": errors.New("broken")},
	}
	sc := NewSmbShareConnection("share", fc)
	files, err := sc.FetchFiles("/root", []string{".epub"}, true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1 file, got %d", len(files))
	}
}

// TestSmbShareConnection_SkipJunctions ensures junction directories are skipped when recursing.
func TestSmbShareConnection_SkipJunctions(t *testing.T) {
	fc := &fakeConnList{
		files: map[string][]smb.SharedFile{
			"/root": {
				{FullPath: "/root/junc", Name: "junc", IsDir: true, IsJunction: true},
				{FullPath: "/root/ok.epub", Name: "ok.epub"},
			},
		},
	}
	sc := NewSmbShareConnection("share", fc)
	files, err := sc.FetchFiles("/root", []string{".epub"}, true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("want 1 file, got %d", len(files))
	}
}
