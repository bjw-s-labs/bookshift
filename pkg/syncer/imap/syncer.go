package imap

import (
	"fmt"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type ImapSyncer struct {
	config *config.ImapConfig
}

func NewImapSyncer(shareConfig *config.ImapConfig) *ImapSyncer {
	// Set default port if nothing is specified
	if !(shareConfig.Port > 0) {
		shareConfig.Port = 143
	}

	return &ImapSyncer{
		config: shareConfig,
	}
}

func (s *ImapSyncer) Run(targetFolder string, validExtensions []string, overwriteExistingFiles bool) error {
	// Connect to the IMAP server
	imapConnection := newImapClient(s.config)
	if err := imapConnect(imapConnection, s.config.Mailbox); err != nil {
		return fmt.Errorf("could not connect to IMAP server %s: %w", s.config.Host, err)
	}
	defer imapDisconnect(imapConnection)

	// Collect messages from the IMAP server
	allMessages, err := imapCollect(imapConnection,
		!s.config.ProcessReadEmails,
		s.config.FilterField,
		s.config.FilterValue,
	)
	if err != nil {
		return err
	}

	// Download attachments for each message
	for _, m := range allMessages {
		if err := imapDownload(m,
			targetFolder,
			validExtensions,
			overwriteExistingFiles,
			s.config.RemoveEmailsAfterDownload,
		); err != nil {
			return err
		}
	}

	return nil
}
