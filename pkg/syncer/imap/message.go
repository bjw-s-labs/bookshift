package imap

import (
	"fmt"

	"github.com/emersion/go-imap/v2"
)

type ImapMessage struct {
	uid        imap.UID
	imapClient *ImapClient
}

// ImapMessage represents a single message in a mailbox, addressed by UID, and
// provides helpers to download attachments and delete the message.
func NewImapMessage(uid imap.UID, ic *ImapClient) *ImapMessage {
	return &ImapMessage{
		uid:        uid,
		imapClient: ic,
	}
}

// DeleteFromServer marks the message as \Deleted and expunges it from the mailbox.
func (im *ImapMessage) DeleteFromServer() error {
	if im.imapClient.Backend == nil {
		return fmt.Errorf("failed to delete message, IMAP client not connected")
	}
	if err := im.imapClient.Backend.StoreAddFlags(imap.UID(im.uid), []imap.Flag{imap.FlagDeleted}); err != nil {
		return err
	}
	if err := im.imapClient.Backend.Expunge(); err != nil {
		return err
	}
	return nil
}
