package nfs

import (
	"fmt"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type NfsSyncer struct {
	config *config.NfsNetworkShareConfig
}

func NewNfsSyncer(shareConfig config.NfsNetworkShareConfig) *NfsSyncer {
	// Set default port
	if !(shareConfig.Port > 0) {
		shareConfig.Port = 2049
	}

	return &NfsSyncer{
		config: &shareConfig,
	}
}

func (s *NfsSyncer) Run(targetFolder string, overwriteExistingFiles bool) error {
	// Connect to the NFS server
	nfsClient := NfsClient{
		Host: s.config.Host,
		Port: s.config.Port,
	}
	if err := nfsClient.Connect(); err != nil {
		return fmt.Errorf("could not connect to NFS server %s. %w", s.config.Host, err)
	}
	defer nfsClient.Disconnect()

	// Instantiate an NFS Folder
	nfsFolder := NewNfsFolder(s.config.Folder, &nfsClient)

	// Fetch all files in the folder
	allFiles, err := nfsFolder.FetchFiles(s.config.Folder, true)
	if err != nil {
		return err
	}

	// Download all files
	for _, file := range allFiles {
		if err := file.Download(
			targetFolder,
			file.CleanFileName(),
			overwriteExistingFiles,
			s.config.KeepFolderStructure,
			s.config.RemoveFilesAfterDownload,
		); err != nil {
			return err
		}
	}

	return nil
}
