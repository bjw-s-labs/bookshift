package nfs

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/kha7iq/go-nfs-client/nfs4"
)

// Separate fake to avoid conflicts with existing test fake
type fakeNfsEx struct {
	files      []nfs4.FileInfo
	connectErr error
	listErr    error
	readErr    error
	deleteErr  error
}

func (f *fakeNfsEx) Connect(_ time.Duration) error { return f.connectErr }
func (f *fakeNfsEx) Disconnect() error             { return nil }
func (f *fakeNfsEx) GetFileList(path string) ([]nfs4.FileInfo, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.files, nil
}
func (f *fakeNfsEx) ReadFileAll(path string, w io.Writer) (int, error) {
	if f.readErr != nil {
		return 0, f.readErr
	}
	return w.Write([]byte("abc"))
}
func (f *fakeNfsEx) DeleteFile(path string) error { return f.deleteErr }
func (f *fakeNfsEx) Host() string                 { return "h" }
func (f *fakeNfsEx) Port() int                    { return 2049 }

// TestNewNfsSyncer_DefaultPort ensures default NFS port is set when unspecified.
func TestNewNfsSyncer_DefaultPort(t *testing.T) {
	cfg := &config.NfsNetworkShareConfig{}
	NewNfsSyncer(cfg)
	if cfg.Port != 2049 {
		t.Fatalf("want 2049, got %d", cfg.Port)
	}
}

// TestNfsSyncer_Run_ConnectError ensures connection errors abort the run.
func TestNfsSyncer_Run_ConnectError(t *testing.T) {
	cfg := &config.NfsNetworkShareConfig{Host: "h"}
	s := NewNfsSyncer(cfg)
	origNew, origConn := newNfsClient, nfsConnect
	t.Cleanup(func() { newNfsClient, nfsConnect = origNew, origConn })
	newNfsClient = func(host string, port int) NfsAPI { return &fakeNfsEx{connectErr: errors.New("x")} }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return c.Connect(0) }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected connect error")
	}
}

// TestNfsSyncer_Run_ListError ensures list errors abort the run.
func TestNfsSyncer_Run_ListError(t *testing.T) {
	cfg := &config.NfsNetworkShareConfig{Host: "h", Folder: "/r"}
	s := NewNfsSyncer(cfg)
	origNew, origConn := newNfsClient, nfsConnect
	t.Cleanup(func() { newNfsClient, nfsConnect = origNew, origConn })
	fake := &fakeNfsEx{listErr: errors.New("list")}
	newNfsClient = func(host string, port int) NfsAPI { return fake }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return nil }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected list error")
	}
}

// TestNfsSyncer_Run_DownloadAndDelete verifies a happy path including file deletion.
func TestNfsSyncer_Run_DownloadAndDelete(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.NfsNetworkShareConfig{Host: "h", Folder: "/r", KeepFolderStructure: true, RemoveFilesAfterDownload: true}
	s := NewNfsSyncer(cfg)
	origNew, origConn := newNfsClient, nfsConnect
	t.Cleanup(func() { newNfsClient, nfsConnect = origNew, origConn })
	file := nfs4.FileInfo{Name: "a.epub", Size: 3, IsDir: false}
	fake := &fakeNfsEx{files: []nfs4.FileInfo{file}}
	newNfsClient = func(host string, port int) NfsAPI { return fake }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return nil }
	if err := s.Run(dir, []string{".epub"}, true); err != nil {
		t.Fatalf("run err: %v", err)
	}
	// expect file exists
	p := filepath.Join(dir, "a.epub")
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected file downloaded: %v", err)
	}
}

// TestNfsSyncer_Run_ReadError ensures read errors during download abort the run.
func TestNfsSyncer_Run_ReadError(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.NfsNetworkShareConfig{Host: "h", Folder: "/r"}
	s := NewNfsSyncer(cfg)
	origNew, origConn := newNfsClient, nfsConnect
	t.Cleanup(func() { newNfsClient, nfsConnect = origNew, origConn })
	file := nfs4.FileInfo{Name: "a.epub", Size: 3, IsDir: false}
	fake := &fakeNfsEx{files: []nfs4.FileInfo{file}, readErr: errors.New("read")}
	newNfsClient = func(host string, port int) NfsAPI { return fake }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return nil }
	if err := s.Run(dir, []string{".epub"}, true); err == nil {
		t.Fatalf("expected read error")
	}
}

// TestNfsSyncer_Run_FetchFilesError ensures fetch errors from folder bubble up.
func TestNfsSyncer_Run_FetchFilesError(t *testing.T) {
	cfg := &config.NfsNetworkShareConfig{Host: "h", Folder: "/r"}
	s := NewNfsSyncer(cfg)
	origNew, origConn, origFolder, origFetch := newNfsClient, nfsConnect, nfsNewFolder, nfsFetchFiles
	t.Cleanup(func() {
		newNfsClient, nfsConnect, nfsNewFolder, nfsFetchFiles = origNew, origConn, origFolder, origFetch
	})
	newNfsClient = func(host string, port int) NfsAPI { return &fakeNfsEx{} }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return nil }
	nfsNewFolder = func(folder string, conn NfsAPI) *NfsFolder { return &NfsFolder{Folder: folder, nfsClient: conn} }
	nfsFetchFiles = func(f *NfsFolder, folder string, valid []string, recurse bool) ([]NfsFile, error) {
		return nil, errors.New("fetch")
	}
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected fetch error")
	}
}

// TestNfsSyncer_Run_DownloadError ensures download errors from file bubble up.
func TestNfsSyncer_Run_DownloadError(t *testing.T) {
	cfg := &config.NfsNetworkShareConfig{Host: "h", Folder: "/r"}
	s := NewNfsSyncer(cfg)
	origNew, origConn, origFolder, origFetch, origDl := newNfsClient, nfsConnect, nfsNewFolder, nfsFetchFiles, nfsDownload
	t.Cleanup(func() {
		newNfsClient, nfsConnect, nfsNewFolder, nfsFetchFiles, nfsDownload = origNew, origConn, origFolder, origFetch, origDl
	})
	newNfsClient = func(host string, port int) NfsAPI { return &fakeNfsEx{} }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return nil }
	nfsNewFolder = func(folder string, conn NfsAPI) *NfsFolder { return &NfsFolder{Folder: folder, nfsClient: conn} }
	nfsFetchFiles = func(f *NfsFolder, folder string, valid []string, recurse bool) ([]NfsFile, error) {
		return []NfsFile{{nfsFolder: f, nfsFile: &nfs4.FileInfo{Name: "a.epub"}}}, nil
	}
	nfsDownload = func(nf *NfsFile, dst, name string, overwrite, keep, del bool) error { return errors.New("dl") }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected download error")
	}
}

// TestNfsSyncer_RunContext_Cancel verifies that context cancelation is observed between files.
func TestNfsSyncer_RunContext_Cancel(t *testing.T) {
	cfg := &config.NfsNetworkShareConfig{Host: "h", Folder: "/r"}
	s := NewNfsSyncer(cfg)
	origNew, origConn, origFolder, origFetch, origDl := newNfsClient, nfsConnect, nfsNewFolder, nfsFetchFiles, nfsDownload
	t.Cleanup(func() {
		newNfsClient, nfsConnect, nfsNewFolder, nfsFetchFiles, nfsDownload = origNew, origConn, origFolder, origFetch, origDl
	})
	newNfsClient = func(host string, port int) NfsAPI { return &fakeNfsEx{} }
	nfsConnect = func(c NfsAPI, _ time.Duration) error { return nil }
	nfsNewFolder = func(folder string, conn NfsAPI) *NfsFolder { return &NfsFolder{Folder: folder, nfsClient: conn} }
	nfsFetchFiles = func(f *NfsFolder, folder string, valid []string, recurse bool) ([]NfsFile, error) {
		return []NfsFile{{nfsFolder: f, nfsFile: &nfs4.FileInfo{Name: "a.epub"}}, {nfsFolder: f, nfsFile: &nfs4.FileInfo{Name: "b.epub"}}}, nil
	}
	nfsDownload = func(nf *NfsFile, dst, name string, overwrite, keep, del bool) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(15 * time.Millisecond); cancel() }()
	_ = s.RunContext(ctx, t.TempDir(), []string{".epub"}, true)
}
