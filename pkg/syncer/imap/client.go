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

	Client *imapclient.Client
}

func (ic *ImapClient) Connect(mailbox string) error {
	connStr := fmt.Sprintf("%s:%v", ic.Host, ic.Port)

	// Connect to the IMAP server
	c, err := imapclient.DialTLS(connStr, nil)
	if err != nil {
		return err
	}

	// Log in to the IMAP server
	if err := c.Login(ic.Username, string(*ic.Password)).Wait(); err != nil {
		return err
	}

	// Select the desired mailbox
	_, err = c.Select(mailbox, nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to select IMAP mailbox (%w)", err)
	}

	// Store the client
	ic.Client = c
	return nil
}

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
	// If the client is nil return an error indicating that the client is not connected.
	if ic.Client == nil {
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

	// Execute the search command
	searchData, err := ic.Client.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, err
	}

	// Iterate over all matching message UIDs and create new ImapMessage instances.
	var filteredMessages []*ImapMessage
	searchUIDs := searchData.AllUIDs()
	for _, msgUid := range searchUIDs {
		filteredMessages = append(filteredMessages, NewImapMessage(msgUid, ic))
	}

	return filteredMessages, nil
}

func (ic *ImapClient) fetchByUID(uid imap.UID, options *imap.FetchOptions) (*imapclient.FetchMessageBuffer, error) {
	// If the client is nil return an error indicating that the client is not connected.
	if ic.Client == nil {
		return nil, fmt.Errorf("failed to fetch message, IMAP client not connected")
	}

	// Fetch the message with the specified UID using the provided options.
	messages, err := ic.Client.Fetch(imap.UIDSetNum(uid), options).Collect()
	if err != nil {
		return nil, err
	}

	// Check if exactly one message was returned. If not, return an error.
	if len(messages) != 1 {
		return nil, fmt.Errorf("len(messages) = %v, want 1", len(messages))
	}

	// Return the fetched message.
	return messages[0], nil
}
