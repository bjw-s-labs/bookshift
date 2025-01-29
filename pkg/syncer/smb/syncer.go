package smb

import (
	"fmt"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type SmbSyncer struct {
	config *config.SmbNetworkShare
}

func NewSmbSyncer(shareConfig config.SmbNetworkShare) *SmbSyncer {
	// Set default port
	if !(shareConfig.Port > 0) {
		shareConfig.Port = 445
	}

	return &SmbSyncer{
		config: &shareConfig,
	}
}

func (s *SmbSyncer) Run(targetFolder string, overwriteExistingFiles bool) error {
	// Connect to the SMB server
	smbConnection := SmbConnection{
		Host:     s.config.Host,
		Port:     s.config.Port,
		Username: s.config.Username,
		Password: s.config.Password,
		Domain:   s.config.Domain,
	}

	err := smbConnection.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to SMB server %s. %w", s.config.Host, err)
	}
	defer smbConnection.Disconnect()

	// Connect to the share
	smbShareConnection := NewSmbShareConnection(s.config.Share, &smbConnection)
	err = smbShareConnection.Connect()
	if err != nil {
		return fmt.Errorf("could not connect to SMB share %s. %w", s.config.Share, err)
	}
	defer smbShareConnection.Disconnect()

	// Fetch all files in the share
	allFiles, err := smbShareConnection.FetchFiles(s.config.Folder, true)
	if err != nil {
		return err
	}

	// Download all files
	for _, file := range allFiles {
		err := file.Download(targetFolder, file.CleanFileName(), overwriteExistingFiles, s.config.KeepFolderStructure, s.config.RemoveFilesAfterDownload)
		if err != nil {
			return err
		}
	}

	return nil
}
