// syncer_seams.go: Test seams for the IMAP syncer.
//
// This file defines small interfaces and function variables that production code
// uses by default, but tests can override to substitute fakes. Keeping these
// seams in a separate file keeps the main implementation clean while preserving
// simple, package-scoped overrides in tests.
package imap

import "github.com/bjw-s-labs/bookshift/pkg/config"

// test hooks (seams) for dependency injection in tests

type imapSyncClient interface {
	Connect(mailbox string) error
	Disconnect() error
	CollectMessages(unreadOnly bool, filterField, filterValue string) ([]*ImapMessage, error)
}

var (
	newImapClient = func(cfg *config.ImapConfig) imapSyncClient {
		return &ImapClient{Host: cfg.Host, Port: cfg.Port, Username: cfg.Username, Password: cfg.Password}
	}
	imapConnect    = func(c imapSyncClient, mailbox string) error { return c.Connect(mailbox) }
	imapDisconnect = func(c imapSyncClient) { _ = c.Disconnect() }
	imapCollect    = func(c imapSyncClient, unreadOnly bool, field, value string) ([]*ImapMessage, error) {
		return c.CollectMessages(unreadOnly, field, value)
	}
	imapDownload = func(m *ImapMessage, dst string, valid []string, overwrite bool, remove bool) error {
		return m.DownloadAttachments(dst, valid, overwrite, remove)
	}
)
