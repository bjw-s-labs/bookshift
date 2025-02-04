package imap

import (
	"fmt"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type ImapSyncer struct {
	config *config.ImapConfig
}

func NewImapSyncer(shareConfig config.ImapConfig) *ImapSyncer {
	// Set default port
	if !(shareConfig.Port > 0) {
		shareConfig.Port = 143
	}

	return &ImapSyncer{
		config: &shareConfig,
	}
}

func (s *ImapSyncer) Run(targetFolder string, validExtensions []string, overwriteExistingFiles bool) error {
	// Connect to the IMAP server
	imapConnection := ImapClient{
		Host:     s.config.Host,
		Port:     s.config.Port,
		Username: s.config.Username,
		Password: s.config.Password,
	}

	if err := imapConnection.Connect(s.config.Mailbox); err != nil {
		return fmt.Errorf("could not connect to IMAP server %s. %w", s.config.Host, err)
	}
	defer imapConnection.Disconnect()

	allMessages, err := imapConnection.CollectMessages(
		!s.config.ProcessReadEmails,
		s.config.FilterField,
		s.config.FilterValue,
	)
	if err != nil {
		return err
	}

	for _, message := range allMessages {
		if err := message.DownloadAttachments(
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
