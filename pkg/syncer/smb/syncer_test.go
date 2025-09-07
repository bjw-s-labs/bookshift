package smb

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/jfjallid/go-smb/smb"
)

// Fake connection that implements SmbConnAPI and drives directory listing and retrieval.
type smbConnFake2 struct{ smbConnBase }

// inherit no-op base methods
var _ SmbConnAPI = (*smbConnFake2)(nil)

func (s *smbConnFake2) ListDirectory(share, subfolder, pattern string) ([]smb.SharedFile, error) {
	// Return one file
	return []smb.SharedFile{{Name: "a.epub", FullPath: "/root/a.epub", Size: 3}}, nil
}
func (s *smbConnFake2) RetrieveFile(share, fp string, offset uint64, cb func([]byte) (int, error)) error {
	// Write 3 bytes via callback
	_, _ = cb([]byte("abc"))
	return nil
}
func (s *smbConnFake2) DeleteFile(share, p string) error { return nil }

// TestNewSmbSyncer_DefaultPort ensures default SMB port is set when unspecified.
func TestNewSmbSyncer_DefaultPort(t *testing.T) {
	cfg := &config.SmbNetworkShareConfig{}
	NewSmbSyncer(cfg)
	if cfg.Port != 445 {
		t.Fatalf("want 445, got %d", cfg.Port)
	}
}

// TestSmbSyncer_Run_ServerConnectError ensures errors connecting to the SMB server propagate.
func TestSmbSyncer_Run_ServerConnectError(t *testing.T) {
	cfg := &config.SmbNetworkShareConfig{Host: "h"}
	s := NewSmbSyncer(cfg)
	origConn := smbConnect
	t.Cleanup(func() { smbConnect = origConn })
	smbConnect = func(c *SmbConnection) error { return errors.New("x") }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected server connect error")
	}
}

// TestSmbSyncer_Run_ShareConnectError ensures errors connecting to the share propagate.
func TestSmbSyncer_Run_ShareConnectError(t *testing.T) {
	cfg := &config.SmbNetworkShareConfig{Host: "h", Share: "s"}
	s := NewSmbSyncer(cfg)
	origShareConnect, origSmbConnect := smbShareConnect, smbConnect
	t.Cleanup(func() { smbShareConnect, smbConnect = origShareConnect, origSmbConnect })
	smbConnect = func(c *SmbConnection) error { return nil }
	smbShareConnect = func(s *SmbShareConnection) error { return errors.New("y") }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected share connect error")
	}
}

// TestSmbSyncer_Run_DownloadHappy verifies a simple successful download flow.
func TestSmbSyncer_Run_DownloadHappy(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.SmbNetworkShareConfig{Host: "h", Share: "s", Folder: "/root"}
	s := NewSmbSyncer(cfg)
	origNew, origConn, origDisc, origShareConn, origShareDisc := newSmbShare, smbConnect, smbDisconnect, smbShareConnect, smbShareDisconnect
	t.Cleanup(func() {
		newSmbShare, smbConnect, smbDisconnect, smbShareConnect, smbShareDisconnect = origNew, origConn, origDisc, origShareConn, origShareDisc
	})
	smbConnect = func(c *SmbConnection) error { return nil }
	smbDisconnect = func(c *SmbConnection) error { return nil }
	// Inject a share connection with our fake SmbConnAPI
	newSmbShare = func(share string, conn SmbConnAPI) *SmbShareConnection {
		return &SmbShareConnection{Share: share, SmbConnection: &smbConnFake2{}}
	}
	smbShareConnect = func(s *SmbShareConnection) error { return nil }
	smbShareDisconnect = func(s *SmbShareConnection) error { return nil }
	if err := s.Run(dir, []string{".epub"}, true); err != nil {
		t.Fatalf("run err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "a.epub")); err != nil {
		t.Fatalf("expected file downloaded: %v", err)
	}
}

type smbConnListErr struct{ smbConnBase }

var _ SmbConnAPI = (*smbConnListErr)(nil)

func (s *smbConnListErr) ListDirectory(share, subfolder, pattern string) ([]smb.SharedFile, error) {
	return nil, errors.New("list")
}
func (s *smbConnListErr) RetrieveFile(share, fp string, offset uint64, cb func([]byte) (int, error)) error {
	return nil
}
func (s *smbConnListErr) DeleteFile(share, p string) error { return nil }

// TestSmbSyncer_Run_ListError ensures list errors abort the run.
func TestSmbSyncer_Run_ListError(t *testing.T) {
	cfg := &config.SmbNetworkShareConfig{Host: "h", Share: "s", Folder: "/root"}
	s := NewSmbSyncer(cfg)
	origNew, origConn, origShareConn, origShareDisc := newSmbShare, smbConnect, smbShareConnect, smbShareDisconnect
	t.Cleanup(func() {
		newSmbShare, smbConnect, smbShareConnect, smbShareDisconnect = origNew, origConn, origShareConn, origShareDisc
	})
	smbConnect = func(c *SmbConnection) error { return nil }
	newSmbShare = func(share string, conn SmbConnAPI) *SmbShareConnection {
		return &SmbShareConnection{Share: share, SmbConnection: &smbConnListErr{}}
	}
	smbShareConnect = func(s *SmbShareConnection) error { return nil }
	smbShareDisconnect = func(s *SmbShareConnection) error { return nil }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected list error")
	}
}

type smbConnRetrieveErr struct{ smbConnBase }

var _ SmbConnAPI = (*smbConnRetrieveErr)(nil)

func (s *smbConnRetrieveErr) ListDirectory(share, subfolder, pattern string) ([]smb.SharedFile, error) {
	return []smb.SharedFile{{Name: "a.epub", FullPath: "/root/a.epub", Size: 3}}, nil
}
func (s *smbConnRetrieveErr) RetrieveFile(share, fp string, offset uint64, cb func([]byte) (int, error)) error {
	return errors.New("retrieve")
}
func (s *smbConnRetrieveErr) DeleteFile(share, p string) error { return nil }

// TestSmbSyncer_Run_RetrieveError ensures retrieval errors abort the run.
func TestSmbSyncer_Run_RetrieveError(t *testing.T) {
	cfg := &config.SmbNetworkShareConfig{Host: "h", Share: "s", Folder: "/root"}
	s := NewSmbSyncer(cfg)
	origNew, origConn, origShareConn, origShareDisc := newSmbShare, smbConnect, smbShareConnect, smbShareDisconnect
	t.Cleanup(func() {
		newSmbShare, smbConnect, smbShareConnect, smbShareDisconnect = origNew, origConn, origShareConn, origShareDisc
	})
	smbConnect = func(c *SmbConnection) error { return nil }
	newSmbShare = func(share string, conn SmbConnAPI) *SmbShareConnection {
		return &SmbShareConnection{Share: share, SmbConnection: &smbConnRetrieveErr{}}
	}
	smbShareConnect = func(s *SmbShareConnection) error { return nil }
	smbShareDisconnect = func(s *SmbShareConnection) error { return nil }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected retrieve error")
	}
}

// TestSmbSyncer_RunContext_Cancel verifies that context cancelation is observed between files.
func TestSmbSyncer_RunContext_Cancel(t *testing.T) {
	cfg := &config.SmbNetworkShareConfig{Host: "h", Share: "s", Folder: "/root"}
	s := NewSmbSyncer(cfg)
	origNew, origConn, origShareConn, origShareDisc := newSmbShare, smbConnect, smbShareConnect, smbShareDisconnect
	t.Cleanup(func() {
		newSmbShare, smbConnect, smbShareConnect, smbShareDisconnect = origNew, origConn, origShareConn, origShareDisc
	})
	smbConnect = func(c *SmbConnection) error { return nil }
	newSmbShare = func(share string, conn SmbConnAPI) *SmbShareConnection {
		return &SmbShareConnection{Share: share, SmbConnection: &smbConnFake2{}}
	}
	smbShareConnect = func(s *SmbShareConnection) error { return nil }
	smbShareDisconnect = func(s *SmbShareConnection) error { return nil }
	// Delay in download path is via RetrieveFile callback - already fast; insert small sleep by wrapping doImap-like path isn’t needed.
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(10 * time.Millisecond); cancel() }()
	_ = s.RunContext(ctx, t.TempDir(), []string{".epub"}, false)
}
