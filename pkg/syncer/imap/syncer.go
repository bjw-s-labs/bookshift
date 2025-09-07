package imap

import (
	"context"
	"fmt"
	"time"

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
	return s.RunContext(context.Background(), targetFolder, validExtensions, overwriteExistingFiles)
}

func (s *ImapSyncer) RunContext(ctx context.Context, targetFolder string, validExtensions []string, overwriteExistingFiles bool) error {
	if s.config.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(s.config.TimeoutSeconds)*time.Second)
		defer cancel()
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
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
