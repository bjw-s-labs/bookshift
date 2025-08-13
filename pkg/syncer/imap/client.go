package imap

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/go-playground/sensitive"
)

type ImapClient struct {
	Host     string
	Port     int
	Username string
	Password *sensitive.String

	Client  imapConn
	Backend ImapOps
}

// Connect establishes a TLS IMAP connection, logs in, selects the mailbox, and wires the backend.
func (ic *ImapClient) Connect(mailbox string) error {
	connStr := fmt.Sprintf("%s:%v", ic.Host, ic.Port)

	// Connect to the IMAP server (via seam)
	client, real, err := imapDial(connStr)
	if err != nil {
		return err
	}

	if err := client.Login(ic.Username, string(*ic.Password)).Wait(); err != nil {
		return err
	}

	if _, err := client.Select(mailbox, nil).Wait(); err != nil {
		return fmt.Errorf("failed to select IMAP mailbox (%w)", err)
	}

	ic.Client = client
	if real != nil {
		ic.Backend = &realImapOps{c: real}
	}
	return nil
}

// Disconnect logs out and closes the IMAP connection if present.
func (ic *ImapClient) Disconnect() error {
	// If the client is nil, there's nothing to disconnect from.
	if ic.Client == nil {
		return nil
	}

	// Log out of the IMAP server
	if err := ic.Client.Logout().Wait(); err != nil {
		return err
	}

	// Terminate the connection to the IMAP server
	if err := ic.Client.Close(); err != nil {
		return err
	}
	return nil
}

func (ic *ImapClient) CollectMessages(ignoreReadMessages bool, filterHeader string, filterValue string) ([]*ImapMessage, error) {
	// If the backend is nil return an error indicating that the client is not connected.
	if ic.Backend == nil {
		return nil, fmt.Errorf("failed to collect messages, IMAP client not connected")
	}

	criteria := &imap.SearchCriteria{}

	// Ignore seen messages if requested
	if ignoreReadMessages {
		criteria.NotFlag = append(criteria.NotFlag, "\\Seen")
	}

	// Always ignore deleted messages
	criteria.NotFlag = append(criteria.NotFlag, "\\Deleted")

	switch filterHeader {
	case "to":
		// Search for messages where the recipient's email address matches the filter value
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "TO", Value: filterValue})
	case "subject":
		// Search for messages where the subject line matches the filter value
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "SUBJECT", Value: filterValue})
	}

	uids, err := ic.Backend.UIDSearch(criteria)
	if err != nil {
		return nil, err
	}

	// Iterate over all matching message UIDs and create new ImapMessage instances.
	var filteredMessages []*ImapMessage
	for _, msgUid := range uids {
		filteredMessages = append(filteredMessages, NewImapMessage(msgUid, ic))
	}

	return filteredMessages, nil
}

// fetchByUID retrieves a single message by UID using the backend.
func (ic *ImapClient) fetchByUID(uid imap.UID, options *imap.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	// If the backend is nil return an error indicating that the client is not connected.
	if ic.Backend == nil {
		return nil, fmt.Errorf("failed to fetch message, IMAP client not connected")
	}
	return ic.Backend.FetchOneByUID(uid, options)
}
