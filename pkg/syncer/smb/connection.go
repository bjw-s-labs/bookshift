package smb

import (
	"fmt"
	"time"

	"github.com/go-playground/sensitive"
	"github.com/jfjallid/go-smb/smb"
	"github.com/jfjallid/go-smb/spnego"
)

// Package-level errors
var (
	ErrSmbDisconnected = fmt.Errorf("not connected to the SMB server")
)

// SmbConnAPI captures the minimal operations used by share and file code, enabling mocks in tests.
type SmbConnAPI interface {
	// Tree/share operations
	TreeConnect(share string) error
	TreeDisconnect(share string) error
	ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error)

	// File operations
	RetrieveFile(share string, filepath string, offset uint64, callback func([]byte) (int, error)) error
	DeleteFile(share string, filepath string) error
}

// smbLowLevel captures the minimal calls used from the underlying go-smb Connection.
// This enables injecting a fake connection in tests.
type smbLowLevel interface {
	Close()
	TreeConnect(share string) error
	TreeDisconnect(share string) error
	ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error)
	RetrieveFile(share string, filepath string, offset uint64, callback func([]byte) (int, error)) error
	DeleteFile(share string, filepath string) error
}

type SmbConnection struct {
	Host     string
	Port     int
	Username string
	Password *sensitive.String
	Domain   string

	connection smbLowLevel
}

// dial hook for creating a new SMB connection (overridable in tests)
var smbDial = func(options smb.Options) (smbLowLevel, error) { return smb.NewConnection(options) }

func (s *SmbConnection) Connect() error {
	options := smb.Options{
		Host:        s.Host,
		Port:        s.Port,
		DialTimeout: time.Duration(10) * time.Second,
		Initiator: &spnego.NTLMInitiator{
			User:     s.Username,
			Password: string(*s.Password),
			Domain:   s.Domain,
		},
	}
	conn, err := smbDial(options)
	if err != nil {
		return err
	}

	if conn == nil {
		return ErrSmbDisconnected
	}

	s.connection = conn
	return nil
}

func (s *SmbConnection) Disconnect() error {
	if s.connection == nil {
		return ErrSmbDisconnected
	}

	s.connection.Close()
	s.connection = nil
	return nil
}

// TreeConnect connects to a share tree.
func (s *SmbConnection) TreeConnect(share string) error {
	if s.connection == nil {
		return ErrSmbDisconnected
	}
	return s.connection.TreeConnect(share)
}

// TreeDisconnect disconnects from a share tree.
func (s *SmbConnection) TreeDisconnect(share string) error {
	if s.connection == nil {
		return ErrSmbDisconnected
	}
	return s.connection.TreeDisconnect(share)
}

// ListDirectory lists directory entries on a share.
func (s *SmbConnection) ListDirectory(share string, subfolder string, pattern string) ([]smb.SharedFile, error) {
	if s.connection == nil {
		return nil, ErrSmbDisconnected
	}
	return s.connection.ListDirectory(share, subfolder, pattern)
}

func (s *SmbConnection) RetrieveFile(share string, filepath string, offset uint64, callback func([]byte) (int, error)) error {
	if s.connection == nil {
		return ErrSmbDisconnected
	}
	return s.connection.RetrieveFile(share, filepath, offset, callback)
}

func (s *SmbConnection) DeleteFile(share string, filepath string) error {
	if s.connection == nil {
		return ErrSmbDisconnected
	}
	return s.connection.DeleteFile(share, filepath)
}
