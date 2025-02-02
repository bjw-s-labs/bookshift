package imap

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/go-playground/sensitive"
)

type ImapClient struct {
	Host     string
	Port     int
	Username string
	Password sensitive.String

	Client *client.Client
}

func (ic *ImapClient) Connect(mailbox string) error {
	tlsn := ""
	connStr := fmt.Sprintf("%s:%v", ic.Host, ic.Port)

	tlsc := &tls.Config{}
	if tlsn != "" {
		tlsc.ServerName = tlsn
	}

	numRetries := 3

	// Attempt to establish an IMAP connection with TLS using the provided connection string and TLS configuration.
	imapClient, err := client.DialTLS(connStr, tlsc)
	// If there's an error establishing the connection, start retry logic.
	if err != nil {
		for numRetries > 0 {
			time.Sleep(1 * time.Second)

			imapClient, err = client.DialTLS(connStr, tlsc)

			if err != nil {
				numRetries--
			} else {
				break
			}
		}

		if err != nil {
			return err
		}
	}

	if err := imapClient.Login(ic.Username, string(ic.Password)); err != nil {
		return err
	}

	ic.Client = imapClient

	_, err = ic.Client.Select(mailbox, false)
	if err != nil {
		return fmt.Errorf("failed to select IMAP mailbox (%w)", err)
	}

	return nil
}

func (ic *ImapClient) Disconnect() error {
	if err := ic.Client.Logout(); err != nil {
		return err
	}
	return nil
}

func (ic *ImapClient) CollectMessages(filterReadMessages bool, filterHeader string, filterValue string) ([]*ImapMessage, error) {
	criteria := imap.NewSearchCriteria()

	if filterReadMessages {
		criteria.WithoutFlags = append(criteria.WithoutFlags, "\\Seen")
	}

	switch filterHeader {
	case "to":
		criteria.Header.Add("TO", filterValue)
	case "subject":
		criteria.Header.Add("SUBJECT", filterValue)
	}

	uids, err := ic.Client.Search(criteria)
	if err != nil {
		return nil, err
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	return nil, nil
}
