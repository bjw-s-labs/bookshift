package nfs

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/kha7iq/go-nfs-client/nfs4"
)

type fakeNfsLow struct {
	listErr error
	readN   uint64
	readErr error
	delErr  error
	closed  bool
}

func (f *fakeNfsLow) Close()                                               { f.closed = true }
func (f *fakeNfsLow) GetFileList(path string) ([]nfs4.FileInfo, error)     { return nil, f.listErr }
func (f *fakeNfsLow) ReadFileAll(path string, w io.Writer) (uint64, error) { return f.readN, f.readErr }
func (f *fakeNfsLow) DeleteFile(path string) error                         { return f.delErr }

// TestNfsClient_NotConnectedErrors ensures client methods error when not connected and getters return fields.
func TestNfsClient_NotConnectedErrors(t *testing.T) {
	c := &NfsClient{host: "h", port: 1}
	if _, err := c.GetFileList("/"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := c.ReadFileAll("/x", nil); err == nil {
		t.Fatalf("expected error")
	}
	if err := c.DeleteFile("/x"); err == nil {
		t.Fatalf("expected error")
	}
	if c.Host() != "h" || c.Port() != 1 {
		t.Fatalf("host/port mismatch")
	}
}

// TestNfsClient_Connect_UsesDial verifies the dial seam is used and Disconnect closes.
func TestNfsClient_Connect_UsesDial(t *testing.T) {
	orig := nfsDialLow
	t.Cleanup(func() { nfsDialLow = orig })
	nfsDialLow = func(ctx context.Context, server string, auth nfs4.AuthParams) (nfsLowLevel, error) {
		return &fakeNfsLow{}, nil
	}
	c := NewNfsClient("h", 1)
	if err := c.Connect(0); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if c.client == nil {
		t.Fatalf("expected client set")
	}
	if err := c.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
}

// TestNfsClient_Connect_DialError ensures dial errors propagate during Connect.
func TestNfsClient_Connect_DialError(t *testing.T) {
	orig := nfsDialLow
	t.Cleanup(func() { nfsDialLow = orig })
	nfsDialLow = func(ctx context.Context, server string, auth nfs4.AuthParams) (nfsLowLevel, error) {
		return nil, errors.New("dial")
	}
	c := NewNfsClient("h", 1)
	if err := c.Connect(0); err == nil {
		t.Fatalf("expected dial error")
	}
}

// TestNfsClient_ReadFileAll_SizeClamp ensures large uint64 sizes are clamped to int range.
func TestNfsClient_ReadFileAll_SizeClamp(t *testing.T) {
	c := &NfsClient{host: "h", port: 1, client: &fakeNfsLow{readN: ^uint64(0)}}
	n, _ := c.ReadFileAll("/x", io.Discard)
	if n <= 0 {
		t.Fatalf("expected positive int from clamp, got %d", n)
	}
}
