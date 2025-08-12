package nfs

import (
	"errors"
	"io"
	"time"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

// fakeNfs implements NfsAPI for directory listings and simple reads.
type fakeNfs struct {
	files map[string][]nfs4.FileInfo
	reads map[string]string
}

func (f *fakeNfs) Connect(timeout time.Duration) error { return nil }
func (f *fakeNfs) Disconnect() error                   { return nil }
func (f *fakeNfs) GetFileList(path string) ([]nfs4.FileInfo, error) {
	if l, ok := f.files[path]; ok {
		return l, nil
	}
	return nil, errors.New("not found")
}
func (f *fakeNfs) ReadFileAll(path string, w io.Writer) (int, error) {
	s, ok := f.reads[path]
	if !ok {
		return 0, errors.New("not found")
	}
	return w.Write([]byte(s))
}
func (f *fakeNfs) DeleteFile(path string) error { return nil }
func (f *fakeNfs) Host() string                 { return "fake" }
func (f *fakeNfs) Port() int                    { return 2049 }

// recNfs focuses on file operations and records deletions.
type recNfs struct {
	reads   map[string]string
	readErr map[string]bool
	deleted []string
}

func (f *recNfs) Connect(timeout time.Duration) error              { return nil }
func (f *recNfs) Disconnect() error                                { return nil }
func (f *recNfs) GetFileList(path string) ([]nfs4.FileInfo, error) { return nil, errors.New("unused") }
func (f *recNfs) ReadFileAll(p string, w io.Writer) (int, error) {
	if f.readErr != nil && f.readErr[p] {
		return 0, errors.New("read error")
	}
	s, ok := f.reads[p]
	if !ok {
		return 0, errors.New("not found")
	}
	return w.Write([]byte(s))
}
func (f *recNfs) DeleteFile(p string) error { f.deleted = append(f.deleted, p); return nil }
func (f *recNfs) Host() string              { return "fake" }
func (f *recNfs) Port() int                 { return 2049 }
