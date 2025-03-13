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
	Password sensitive.String

	Client *imapclient.Client
}

func (ic *ImapClient) Connect(mailbox string) error {
	connStr := fmt.Sprintf("%s:%v", ic.Host, ic.Port)

	c, err := imapclient.DialTLS(connStr, nil)
	// If there's an error establishing the connection, start retry logic.
	if err != nil {
		return err
	}

	if err := c.Login(ic.Username, string(ic.Password)).Wait(); err != nil {
		return err
	}

	_, err = c.Select(mailbox, nil).Wait()
	if err != nil {
		return fmt.Errorf("failed to select IMAP mailbox (%w)", err)
	}

	ic.Client = c
	return nil
}

func (ic *ImapClient) Disconnect() error {
	if err := ic.Client.Logout().Wait(); err != nil {
		return err
	}
	return nil
}

func (ic *ImapClient) CollectMessages(ignoreReadMessages bool, filterHeader string, filterValue string) ([]*ImapMessage, error) {
	criteria := &imap.SearchCriteria{}

	if ignoreReadMessages {
		criteria.NotFlag = append(criteria.NotFlag, "\\Seen")
	}

	criteria.NotFlag = append(criteria.NotFlag, "\\Deleted")

	switch filterHeader {
	case "to":
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "TO", Value: filterValue})
	case "subject":
		criteria.Header = append(criteria.Header, imap.SearchCriteriaHeaderField{Key: "SUBJECT", Value: filterValue})
	}

	uids, err := ic.Client.UIDSearch(criteria, nil).Wait()
	if err != nil {
		return nil, err
	}

	var filteredMessages []*ImapMessage
	for _, msgUid := range uids.AllUIDs() {
		filteredMessages = append(filteredMessages, &ImapMessage{uid: msgUid, imapClient: ic})
	}

	return filteredMessages, nil
}
