package smb

import (
	"github.com/go-playground/sensitive"
	"github.com/jfjallid/go-smb/smb"
	"github.com/jfjallid/go-smb/spnego"
)

type SmbConnection struct {
	Host     string
	Port     int
	Username string
	Password sensitive.String
	Domain   string

	Connection *smb.Connection
}

func (s *SmbConnection) Connect() error {
	options := smb.Options{
		Host: s.Host,
		Port: s.Port,
		Initiator: &spnego.NTLMInitiator{
			User:     s.Username,
			Password: string(s.Password),
			Domain:   s.Domain,
		},
	}
	conn, err := smb.NewConnection(options)
	if err != nil {
		return err
	}

	s.Connection = conn
	return nil
}

func (s *SmbConnection) Disconnect() error {
	s.Connection.Close()
	return nil
}
