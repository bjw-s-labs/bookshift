package smb

import (
	"time"

	"github.com/go-playground/sensitive"
	"github.com/jfjallid/go-smb/smb"
	"github.com/jfjallid/go-smb/spnego"
)

type SmbConnection struct {
	Host     string
	Port     int
	Username string
	Password *sensitive.String
	Domain   string

	connection *smb.Connection
}

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
	conn, err := smb.NewConnection(options)
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
