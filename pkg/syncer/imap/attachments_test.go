package imap

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

// TestDetermineAttachmentParts_Filtering ensures parts are filtered by allowed extensions and fields set.
func TestDetermineAttachmentParts_Filtering(t *testing.T) {
	meta := buildMeta("s", "Alice", "alice", "example.com", "book.epub", 4)
	msg := &ImapMessage{}

	parts, err := msg.determineAttachmentParts(meta, []string{".epub"})
	if err != nil {
		t.Fatalf("determineAttachmentParts error: %v", err)
	}
	if len(parts) != 1 {
		t.Fatalf("want 1 part, got %d", len(parts))
	}
	if parts[0].filename != "book.epub" {
		t.Fatalf("filename mismatch: %q", parts[0].filename)
	}
	if parts[0].attachmentSize != 4 {
		t.Fatalf("size mismatch: %d", parts[0].attachmentSize)
	}

	parts, err = msg.determineAttachmentParts(meta, []string{".pdf"})
	if err != nil {
		t.Fatalf("determineAttachmentParts error: %v", err)
	}
	if len(parts) != 0 {
		t.Fatalf("want 0 parts when filtered out, got %d", len(parts))
	}
}

// TestDownloadAttachments_DownloadsAndDeletes verifies decoding, writing, and deletion flow.
func TestDownloadAttachments_DownloadsAndDeletes(t *testing.T) {
	uid := imap.UID(7)
	encoded := base64.StdEncoding.EncodeToString([]byte("DATA"))

	backend := &recordingBackend{
		meta:   map[imap.UID]*imapclient.FetchMessageBuffer{uid: buildMeta("Subject", "Bob", "bob", "example.com", "report.epub", 4)},
		bodies: map[imap.UID]*imapclient.FetchMessageBuffer{uid: buildBody(encoded)},
	}
	ic := &ImapClient{Host: "mail.example", Backend: backend}
	msg := NewImapMessage(uid, ic)

	dir := t.TempDir()
	if err := msg.DownloadAttachments(dir, []string{".epub"}, false, true); err != nil {
		t.Fatalf("DownloadAttachments error: %v", err)
	}

	dst := filepath.Join(dir, "report.epub")
	b, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(b) != "DATA" {
		t.Fatalf("content mismatch: %q", string(b))
	}

	if len(backend.stored) != 1 || backend.stored[0] != uid {
		t.Fatalf("expected StoreAddFlags for uid=%v, got %v", uid, backend.stored)
	}
	if backend.expunges != 1 {
		t.Fatalf("expected 1 expunge, got %d", backend.expunges)
	}
}

// TestDownloadAttachments_SkipExisting_NoOverwrite ensures existing files are skipped without overwrite.
func TestDownloadAttachments_SkipExisting_NoOverwrite(t *testing.T) {
	uid := imap.UID(9)
	encoded := base64.StdEncoding.EncodeToString([]byte("NEW"))
	backend := &recordingBackend{
		meta:   map[imap.UID]*imapclient.FetchMessageBuffer{uid: buildMeta("S", "Carol", "carol", "example.com", "book.epub", 3)},
		bodies: map[imap.UID]*imapclient.FetchMessageBuffer{uid: buildBody(encoded)},
	}
	ic := &ImapClient{Backend: backend}
	msg := NewImapMessage(uid, ic)

	dir := t.TempDir()
	dst := filepath.Join(dir, "book.epub")
	if err := os.WriteFile(dst, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("precreate: %v", err)
	}

	if err := msg.DownloadAttachments(dir, []string{".epub"}, false, false); err != nil {
		t.Fatalf("DownloadAttachments error: %v", err)
	}
	b, _ := os.ReadFile(dst)
	if string(b) != "OLD" {
		t.Fatalf("expected original content to remain, got %q", string(b))
	}
}

// TestDownloadAttachments_OverwriteExisting ensures existing files are overwritten when requested.
func TestDownloadAttachments_OverwriteExisting(t *testing.T) {
	uid := imap.UID(10)
	encoded := base64.StdEncoding.EncodeToString([]byte("REPL"))
	backend := &recordingBackend{
		meta:   map[imap.UID]*imapclient.FetchMessageBuffer{uid: buildMeta("S", "Dave", "dave", "example.com", "paper.epub", 4)},
		bodies: map[imap.UID]*imapclient.FetchMessageBuffer{uid: buildBody(encoded)},
	}
	ic := &ImapClient{Backend: backend}
	msg := NewImapMessage(uid, ic)

	dir := t.TempDir()
	dst := filepath.Join(dir, "paper.epub")
	if err := os.WriteFile(dst, []byte("OLD"), 0o644); err != nil {
		t.Fatalf("precreate: %v", err)
	}

	if err := msg.DownloadAttachments(dir, []string{".epub"}, true, false); err != nil {
		t.Fatalf("DownloadAttachments error: %v", err)
	}
	b, _ := os.ReadFile(dst)
	if string(b) != "REPL" {
		t.Fatalf("expected overwritten content, got %q", string(b))
	}
}

// TestDownloadAttachments_RenameError ensures rename failures are surfaced.
func TestDownloadAttachments_RenameError(t *testing.T) {
	uid := imap.UID(21)
	meta := buildMeta("S", "Ann", "ann", "example.com", "x.epub", 4)
	body := buildBody(base64.StdEncoding.EncodeToString([]byte("DATA")))
	be := &recordingBackend{meta: map[imap.UID]*imapclient.FetchMessageBuffer{uid: meta}, bodies: map[imap.UID]*imapclient.FetchMessageBuffer{uid: body}}
	ic := &ImapClient{Backend: be}
	msg := NewImapMessage(uid, ic)

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "x.epub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := msg.DownloadAttachments(dir, []string{".epub"}, true, false); err == nil {
		t.Fatalf("expected rename error")
	}
}

// TestCollectMessages_BackendError ensures backend search errors are returned.
func TestCollectMessages_BackendError(t *testing.T) {
	ic := &ImapClient{Backend: &fakeBackend{searchErr: os.ErrNotExist}}
	if _, err := ic.CollectMessages(true, "to", "someone@example.com"); err == nil {
		t.Fatalf("expected search error")
	}
}

// TestDownloadAttachments_FetchBodyError ensures body fetch errors are returned.
func TestDownloadAttachments_FetchBodyError(t *testing.T) {
	uid := imap.UID(22)
	meta := buildMeta("S", "Zoe", "zoe", "example.com", "y.epub", 2)
	be := &errBodyBackend{meta: meta}
	ic := &ImapClient{Backend: be}
	msg := NewImapMessage(uid, ic)
	if err := msg.DownloadAttachments(t.TempDir(), []string{".epub"}, false, false); err == nil {
		t.Fatalf("expected fetch body error")
	}
}

// TestDownloadAttachments_CreateTempError ensures CreateTemp failures are handled.
func TestDownloadAttachments_CreateTempError(t *testing.T) {
	uid := imap.UID(23)
	meta := buildMeta("S", "Ian", "ian", "example.com", "z.epub", 2)
	body := buildBody(base64.StdEncoding.EncodeToString([]byte("D")))
	be := &recordingBackend{meta: map[imap.UID]*imapclient.FetchMessageBuffer{uid: meta}, bodies: map[imap.UID]*imapclient.FetchMessageBuffer{uid: body}}
	ic := &ImapClient{Backend: be}
	msg := NewImapMessage(uid, ic)

	file := filepath.Join(t.TempDir(), "notadir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("prep file: %v", err)
	}
	if err := msg.DownloadAttachments(file, []string{".epub"}, true, false); err == nil {
		t.Fatalf("expected create temp error")
	}
}

// TestDownloadAttachments_CreatesFolder ensures destination dir is created when absent.
func TestDownloadAttachments_CreatesFolder(t *testing.T) {
	uid := imap.UID(24)
	meta := buildMeta("S", "Nia", "nia", "example.com", "c.epub", 1)
	body := buildBody(base64.StdEncoding.EncodeToString([]byte("X")))
	be := &recordingBackend{meta: map[imap.UID]*imapclient.FetchMessageBuffer{uid: meta}, bodies: map[imap.UID]*imapclient.FetchMessageBuffer{uid: body}}
	ic := &ImapClient{Backend: be}
	msg := NewImapMessage(uid, ic)
	dst := filepath.Join(t.TempDir(), "nested", "dir")
	if err := msg.DownloadAttachments(dst, []string{".epub"}, false, false); err != nil {
		t.Fatalf("download: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "c.epub")); err != nil {
		t.Fatalf("expected file, err=%v", err)
	}
}
