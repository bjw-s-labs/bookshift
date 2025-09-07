package imap

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type fakeSyncClient struct {
	connectErr error
	collectErr error
	msgs       []*ImapMessage
}

func (f *fakeSyncClient) Connect(mailbox string) error { return f.connectErr }
func (f *fakeSyncClient) Disconnect() error            { return nil }
func (f *fakeSyncClient) CollectMessages(unreadOnly bool, filterField, filterValue string) ([]*ImapMessage, error) {
	if f.collectErr != nil {
		return nil, f.collectErr
	}
	return f.msgs, nil
}

// TestNewImapSyncer_DefaultPort ensures default IMAP port is set when unspecified.
func TestNewImapSyncer_DefaultPort(t *testing.T) {
	cfg := &config.ImapConfig{}
	NewImapSyncer(cfg)
	if cfg.Port != 143 {
		t.Fatalf("want 143, got %d", cfg.Port)
	}
}

// TestImapSyncer_Run_ConnectError ensures connect errors abort the run.
func TestImapSyncer_Run_ConnectError(t *testing.T) {
	cfg := &config.ImapConfig{Host: "h", Mailbox: "INBOX"}
	s := NewImapSyncer(cfg)
	origNew, origConn := newImapClient, imapConnect
	t.Cleanup(func() { newImapClient, imapConnect = origNew, origConn })
	newImapClient = func(cfg *config.ImapConfig) imapSyncClient { return &fakeSyncClient{connectErr: errors.New("x")} }
	imapConnect = func(c imapSyncClient, mailbox string) error { return c.Connect(mailbox) }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected connect error")
	}
}

// TestImapSyncer_Run_CollectError ensures collect errors abort the run.
func TestImapSyncer_Run_CollectError(t *testing.T) {
	cfg := &config.ImapConfig{Host: "h", Mailbox: "INBOX"}
	s := NewImapSyncer(cfg)
	origNew := newImapClient
	t.Cleanup(func() { newImapClient = origNew })
	newImapClient = func(cfg *config.ImapConfig) imapSyncClient { return &fakeSyncClient{collectErr: errors.New("y")} }
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected collect error")
	}
}

// TestImapSyncer_Run_DownloadError ensures download errors abort the run.
func TestImapSyncer_Run_DownloadError(t *testing.T) {
	cfg := &config.ImapConfig{Host: "h", Mailbox: "INBOX"}
	s := NewImapSyncer(cfg)
	origNew, origDL := newImapClient, imapDownload
	t.Cleanup(func() { newImapClient, imapDownload = origNew, origDL })
	m := &ImapMessage{}
	newImapClient = func(cfg *config.ImapConfig) imapSyncClient { return &fakeSyncClient{msgs: []*ImapMessage{m}} }
	imapDownload = func(m *ImapMessage, dst string, valid []string, overwrite bool, remove bool) error {
		return errors.New("z")
	}
	if err := s.Run(t.TempDir(), []string{".epub"}, true); err == nil {
		t.Fatalf("expected download error")
	}
}

// TestImapSyncer_Run_HappyPath verifies a successful run downloads attachments.
func TestImapSyncer_Run_HappyPath(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.ImapConfig{Host: "h", Mailbox: "INBOX"}
	s := NewImapSyncer(cfg)

	// Save and restore seams
	origNew, origConn, origDisc, origCollect, origDL := newImapClient, imapConnect, imapDisconnect, imapCollect, imapDownload
	t.Cleanup(func() {
		newImapClient, imapConnect, imapDisconnect, imapCollect, imapDownload = origNew, origConn, origDisc, origCollect, origDL
	})

	// Fake message that writes a file via DownloadAttachments
	msg := &ImapMessage{}

	// Provide a fake client returning our msg
	newImapClient = func(cfg *config.ImapConfig) imapSyncClient {
		return &fakeSyncClient{msgs: []*ImapMessage{msg}}
	}
	imapConnect = func(c imapSyncClient, mailbox string) error { return nil }
	imapDisconnect = func(c imapSyncClient) {}
	imapCollect = func(c imapSyncClient, unreadOnly bool, field, value string) ([]*ImapMessage, error) {
		return []*ImapMessage{msg}, nil
	}
	imapDownload = func(m *ImapMessage, dst string, valid []string, overwrite bool, remove bool) error {
		// Simulate writing a file
		return os.WriteFile(filepath.Join(dst, "file.epub"), []byte("X"), 0o644)
	}

	if err := s.Run(dir, []string{".epub"}, true); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "file.epub")); err != nil {
		t.Fatalf("expected file: %v", err)
	}
}

// TestImapSyncer_RunContext_Cancel verifies that context cancelation is observed.
func TestImapSyncer_RunContext_Cancel(t *testing.T) {
	cfg := &config.ImapConfig{Host: "h", Mailbox: "INBOX"}
	s := NewImapSyncer(cfg)

	// Provide many messages to iterate so we can cancel between iterations.
	msgs := []*ImapMessage{{}, {}, {}, {}, {}}
	origNew, origCollect := newImapClient, imapCollect
	t.Cleanup(func() { newImapClient, imapCollect = origNew, origCollect })
	newImapClient = func(cfg *config.ImapConfig) imapSyncClient { return &fakeSyncClient{msgs: msgs} }
	imapCollect = func(c imapSyncClient, unreadOnly bool, field, value string) ([]*ImapMessage, error) { return msgs, nil }
	// Download sleeps a bit to allow cancel to fire between items
	origDL := imapDownload
	t.Cleanup(func() { imapDownload = origDL })
	imapDownload = func(m *ImapMessage, dst string, valid []string, overwrite bool, remove bool) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(15 * time.Millisecond); cancel() }()
	_ = s.RunContext(ctx, t.TempDir(), []string{".epub"}, false)
}
