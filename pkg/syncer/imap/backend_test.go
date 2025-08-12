package imap

import (
	"errors"
	"testing"

	"github.com/emersion/go-imap/v2"
)

// TestImapCollectMessages_Backend verifies collection using the Backend interface path.
func TestImapCollectMessages_Backend(t *testing.T) {
	ic := &ImapClient{Backend: &fakeBackend{uids: []imap.UID{1, 2, 3}}}
	msgs, err := ic.CollectMessages(true, "to", "someone@example.com")
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("want 3, got %d", len(msgs))
	}
}

// TestImapFetchByUID_NotConnected ensures fetchByUID requires a connected backend.
func TestImapFetchByUID_NotConnected(t *testing.T) {
	ic := &ImapClient{}
	if _, err := ic.fetchByUID(1, &imap.FetchOptions{}); err == nil {
		t.Fatalf("expected error")
	}
}

// TestImapDeleteFromServer_BackendErrors ensures store/expunge errors are propagated.
func TestImapDeleteFromServer_BackendErrors(t *testing.T) {
	ic := &ImapClient{Backend: &fakeBackend{storeErr: errors.New("x")}}
	m := NewImapMessage(1, ic)
	if err := m.DeleteFromServer(); err == nil {
		t.Fatalf("expected store error")
	}
	ic = &ImapClient{Backend: &fakeBackend{expungeErr: errors.New("x")}}
	m = NewImapMessage(1, ic)
	if err := m.DeleteFromServer(); err == nil {
		t.Fatalf("expected expunge error")
	}
}
