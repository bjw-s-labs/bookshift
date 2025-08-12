package imap

import (
	// "errors"
	"testing"

	"github.com/emersion/go-imap/v2"
)

// type fakeImap struct{}

// func (f *fakeImap) Connect(mailbox string) error { return nil }
// func (f *fakeImap) Disconnect() error            { return nil }
// func (f *fakeImap) CollectMessages(ignoreReadMessages bool, filterHeader string, filterValue string) ([]*ImapMessage, error) {
// 	return nil, errors.New("not connected to real server")
// }

// TestImapClientFetchByUID_NotConnected ensures fetchByUID errors when Client is nil.
func TestImapClientFetchByUID_NotConnected(t *testing.T) {
	ic := &ImapClient{}
	if _, err := ic.fetchByUID(imap.UID(1), &imap.FetchOptions{}); err == nil {
		t.Fatalf("expected error when Client==nil")
	}
}

// TestDetermineAttachmentParts_Filter ensures filtering by allowed extensions using helpers.
func TestDetermineAttachmentParts_Filter(t *testing.T) {
	meta := buildMeta("s", "Alice", "alice", "example.com", "a.epub", 4)
	msg := &ImapMessage{}
	parts, err := msg.determineAttachmentParts(meta, []string{".epub"})
	if err != nil {
		t.Fatalf("determineAttachmentParts: %v", err)
	}
	if len(parts) != 1 || parts[0].filename != "a.epub" || parts[0].attachmentSize != 4 {
		t.Fatalf("unexpected parts: %#v", parts)
	}

	parts, err = msg.determineAttachmentParts(meta, []string{".pdf"})
	if err != nil {
		t.Fatalf("determineAttachmentParts: %v", err)
	}
	if len(parts) != 0 {
		t.Fatalf("expected 0 parts after filtering, got %d", len(parts))
	}
}

// TestNewImapMessage_UID sanity checks NewImapMessage stores UID.
func TestNewImapMessage_UID(t *testing.T) {
	uid := imap.UID(42)
	msg := NewImapMessage(uid, &ImapClient{})
	if msg.uid != uid {
		t.Fatalf("want uid %v, got %v", uid, msg.uid)
	}
}
