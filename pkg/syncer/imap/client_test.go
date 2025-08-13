package imap

import (
	"errors"
	"os"
	"testing"

	imapv2 "github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-playground/sensitive"
)

// TestImapCollectMessages_NotConnected ensures CollectMessages returns an error when not connected.
func TestImapCollectMessages_NotConnected(t *testing.T) {
	ic := &ImapClient{}
	if _, err := ic.CollectMessages(true, "to", "x"); err == nil {
		t.Fatalf("expected not connected error")
	}
}

// TestImapCollectMessages_Subject_NoUnread collects messages without unread filter using subject.
func TestImapCollectMessages_Subject_NoUnread(t *testing.T) {
	ic := &ImapClient{Backend: &fakeBackend{uids: []imapv2.UID{10, 20}}}
	msgs, err := ic.CollectMessages(false, "subject", "Weekly")
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("want 2, got %d", len(msgs))
	}
}

// TestImapDisconnect_NoClient ensures Disconnect is a no-op when Client is nil.
func TestImapDisconnect_NoClient(t *testing.T) {
	ic := &ImapClient{}
	if err := ic.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
}

// TestImapDeleteFromServer_Success verifies delete flows succeed on the backend.
func TestImapDeleteFromServer_Success(t *testing.T) {
	ic := &ImapClient{Backend: &fakeBackend{}}
	m := NewImapMessage(1, ic)
	if err := m.DeleteFromServer(); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

// TestImapFetchByUID_BackendOK ensures fetchByUID returns data when backend provides it.
func TestImapFetchByUID_BackendOK(t *testing.T) {
	fb := &fakeBackend{fetch: map[imapv2.UID]*imapclient.FetchMessageBuffer{1: {}}}
	ic := &ImapClient{Backend: fb}
	if _, err := ic.fetchByUID(1, &imapv2.FetchOptions{}); err != nil {
		t.Fatalf("fetch: %v", err)
	}
}

// TestImapFetchByUID_BackendError ensures backend fetch errors are returned.
func TestImapFetchByUID_BackendError(t *testing.T) {
	fb := &fakeBackend{fetchErr: os.ErrPermission}
	ic := &ImapClient{Backend: fb}
	if _, err := ic.fetchByUID(2, &imapv2.FetchOptions{}); err == nil {
		t.Fatalf("expected error")
	}
}

// TestImapDeleteFromServer_NotConnected ensures DeleteFromServer requires a connected backend.
func TestImapDeleteFromServer_NotConnected(t *testing.T) {
	m := NewImapMessage(1, &ImapClient{})
	if err := m.DeleteFromServer(); err == nil {
		t.Fatalf("expected not connected error")
	}
}

// TestImapClient_Connect_And_Disconnect verifies the dial seam and proper disconnect/cleanup.
func TestImapClient_Connect_And_Disconnect(t *testing.T) {
	orig := imapDial
	t.Cleanup(func() { imapDial = orig })
	imapDial = func(addr string) (imapConn, *imapclient.Client, error) {
		return &fakeConn{}, nil, nil
	}
	pw := sensitive.String("pw")
	ic := &ImapClient{Host: "h", Port: 993, Password: &pw}
	if err := ic.Connect("INBOX"); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if ic.Client == nil {
		t.Fatalf("expected client set")
	}
	// Backend may be nil when real client is nil in seam
	if err := ic.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
}

// TestImapClient_Connect_LoginError ensures login errors are returned.
func TestImapClient_Connect_LoginError(t *testing.T) {
	orig := imapDial
	t.Cleanup(func() { imapDial = orig })
	imapDial = func(addr string) (imapConn, *imapclient.Client, error) {
		return &fakeConn{loginErr: errors.New("x")}, nil, nil
	}
	pw := sensitive.String("pw")
	ic := &ImapClient{Password: &pw}
	if err := ic.Connect("INBOX"); err == nil {
		t.Fatalf("expected login error")
	}
}

// TestImapClient_Connect_SelectError ensures mailbox select errors are returned.
func TestImapClient_Connect_SelectError(t *testing.T) {
	orig := imapDial
	t.Cleanup(func() { imapDial = orig })
	imapDial = func(addr string) (imapConn, *imapclient.Client, error) {
		return &fakeConn{selErr: errors.New("x")}, nil, nil
	}
	pw := sensitive.String("pw")
	ic := &ImapClient{Password: &pw}
	if err := ic.Connect("INBOX"); err == nil {
		t.Fatalf("expected select error")
	}
}

// TestImapClient_Connect_DialError ensures dialing errors are returned.
func TestImapClient_Connect_DialError(t *testing.T) {
	orig := imapDial
	t.Cleanup(func() { imapDial = orig })
	imapDial = func(addr string) (imapConn, *imapclient.Client, error) { return nil, nil, errors.New("dial") }
	pw := sensitive.String("pw")
	ic := &ImapClient{Password: &pw}
	if err := ic.Connect("INBOX"); err == nil {
		t.Fatalf("expected dial error")
	}
}

// TestImapClient_Connect_BackendSet ensures Backend is initialized when real client is present.
func TestImapClient_Connect_BackendSet(t *testing.T) {
	orig := imapDial
	t.Cleanup(func() { imapDial = orig })
	imapDial = func(addr string) (imapConn, *imapclient.Client, error) {
		return &fakeConn{}, &imapclient.Client{}, nil
	}
	pw := sensitive.String("pw")
	ic := &ImapClient{Password: &pw}
	if err := ic.Connect("INBOX"); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if ic.Backend == nil {
		t.Fatalf("expected backend set when real client present")
	}
}

// TestImapDisconnect_LogoutError ensures logout errors are returned by Disconnect.
func TestImapDisconnect_LogoutError(t *testing.T) {
	ic := &ImapClient{Client: &fakeConn2{logoutErr: errors.New("x")}}
	if err := ic.Disconnect(); err == nil {
		t.Fatalf("expected logout error")
	}
}

// TestImapDisconnect_CloseError ensures close errors are returned by Disconnect.
func TestImapDisconnect_CloseError(t *testing.T) {
	ic := &ImapClient{Client: &fakeConn2{closeErr: errors.New("x")}}
	if err := ic.Disconnect(); err == nil {
		t.Fatalf("expected close error")
	}
}

// TestDetermineAttachmentParts_NilMessage ensures an error is returned for nil metadata.
func TestDetermineAttachmentParts_NilMessage(t *testing.T) {
	msg := &ImapMessage{}
	if _, err := msg.determineAttachmentParts(nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}

// TestDownloadAttachments_InvalidBase64 ensures invalid base64 data triggers a decode error.
func TestDownloadAttachments_InvalidBase64(t *testing.T) {
	uid := imapv2.UID(11)
	backend := &recordingBackend{
		meta:   map[imapv2.UID]*imapclient.FetchMessageBuffer{uid: buildMeta("S", "Eve", "eve", "example.com", "file.epub", 4)},
		bodies: map[imapv2.UID]*imapclient.FetchMessageBuffer{uid: {BodySection: []imapclient.FetchBodySectionBuffer{{Section: &imapv2.FetchItemBodySection{Part: []int{1}}, Bytes: []byte("%%%invalid%%%")}}}},
	}
	ic := &ImapClient{Backend: backend}
	msg := NewImapMessage(uid, ic)
	if err := msg.DownloadAttachments(t.TempDir(), []string{".epub"}, false, false); err == nil {
		t.Fatalf("expected decode error")
	}
}
