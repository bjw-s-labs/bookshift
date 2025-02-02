package imap

import (
	"fmt"
	"log/slog"

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

	allMessages, err := imapConnection.CollectMessages()
	if err != nil {
		return err
	}
	slog.Debug("Collected", "messages", allMessages)

	// TODO: Implement the rest of the code here

	// // Fetch all files in the share
	// allFiles, err := smbShareConnection.FetchFiles(s.config.Folder, validExtensions, true)
	// if err != nil {
	// 	return err
	// }

	// // Download all files
	// for _, file := range allFiles {
	// 	if err := file.Download(
	// 		targetFolder,
	// 		file.CleanFileName(),
	// 		overwriteExistingFiles,
	// 		s.config.KeepFolderStructure,
	// 		s.config.RemoveFilesAfterDownload,
	// 	); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
