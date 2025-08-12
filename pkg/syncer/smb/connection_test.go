package smb

import (
	"errors"
	"testing"

	"github.com/go-playground/sensitive"
	"github.com/jfjallid/go-smb/smb"
)

// TestSmbConnection_Disconnect_NotConnected verifies Disconnect errors when not connected.
func TestSmbConnection_Disconnect_NotConnected(t *testing.T) {
	var s SmbConnection
	if err := s.Disconnect(); err == nil {
		t.Fatalf("expected ErrSmbDisconnected")
	}
}

// TestSmbConnection_RetrieveFile_NotConnected ensures RetrieveFile fails if not connected.
func TestSmbConnection_RetrieveFile_NotConnected(t *testing.T) {
	var s SmbConnection
	if err := s.RetrieveFile("share", "/path", 0, func([]byte) (int, error) { return 0, nil }); err == nil {
		t.Fatalf("expected ErrSmbDisconnected")
	}
}

// TestSmbConnection_DeleteFile_NotConnected ensures DeleteFile fails if not connected.
func TestSmbConnection_DeleteFile_NotConnected(t *testing.T) {
	var s SmbConnection
	if err := s.DeleteFile("share", "/path"); err == nil {
		t.Fatalf("expected ErrSmbDisconnected")
	}
}

type fakeLow struct {
	treeConnErr error
	treeDiscErr error
	listErr     error
	retrErr     error
	delErr      error
	files       []smb.SharedFile
	closed      bool
}

func (f *fakeLow) Close()                            { f.closed = true }
func (f *fakeLow) TreeConnect(share string) error    { return f.treeConnErr }
func (f *fakeLow) TreeDisconnect(share string) error { return f.treeDiscErr }
func (f *fakeLow) ListDirectory(share, sub, patt string) ([]smb.SharedFile, error) {
	return f.files, f.listErr
}
func (f *fakeLow) RetrieveFile(share, fp string, off uint64, cb func([]byte) (int, error)) error {
	return f.retrErr
}
func (f *fakeLow) DeleteFile(share, fp string) error { return f.delErr }

// TestSmbConnection_Connect_UsesDialHook verifies the dial seam is used to establish a connection.
func TestSmbConnection_Connect_UsesDialHook(t *testing.T) {
	orig := smbDial
	t.Cleanup(func() { smbDial = orig })
	smbDial = func(opts smb.Options) (smbLowLevel, error) { return &fakeLow{}, nil }

	pw := sensitive.String("pw")
	s := &SmbConnection{Password: &pw}
	if err := s.Connect(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if s.connection == nil {
		t.Fatalf("expected connection to be set")
	}
}

// TestSmbConnection_Connect_DialError ensures dial errors propagate.
func TestSmbConnection_Connect_DialError(t *testing.T) {
	orig := smbDial
	t.Cleanup(func() { smbDial = orig })
	smbDial = func(opts smb.Options) (smbLowLevel, error) { return nil, errors.New("dial") }
	pw := sensitive.String("pw")
	s := &SmbConnection{Password: &pw}
	if err := s.Connect(); err == nil {
		t.Fatalf("expected dial error")
	}
}

// TestSmbConnection_TreeAndList_PassThrough ensures methods pass through to low level implementation.
func TestSmbConnection_TreeAndList_PassThrough(t *testing.T) {
	fl := &fakeLow{files: []smb.SharedFile{{Name: "a"}}}
	s := &SmbConnection{connection: fl}
	if err := s.TreeConnect("share"); err != nil {
		t.Fatalf("tree connect: %v", err)
	}
	if _, err := s.ListDirectory("share", "sub", "*"); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := s.TreeDisconnect("share"); err != nil {
		t.Fatalf("tree disconnect: %v", err)
	}
}

// TestSmbConnection_Disconnect_Closes ensures Disconnect closes the underlying connection.
func TestSmbConnection_Disconnect_Closes(t *testing.T) {
	fl := &fakeLow{}
	s := &SmbConnection{connection: fl}
	if err := s.Disconnect(); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if !fl.closed {
		t.Fatalf("expected closed true")
	}
}
