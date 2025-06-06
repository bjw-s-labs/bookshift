package nfs

import (
	"fmt"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type NfsSyncer struct {
	config *config.NfsNetworkShareConfig
}

func NewNfsSyncer(shareConfig *config.NfsNetworkShareConfig) *NfsSyncer {
	// Set default port
	if !(shareConfig.Port > 0) {
		shareConfig.Port = 2049
	}

	return &NfsSyncer{
		config: shareConfig,
	}
}

func (s *NfsSyncer) Run(targetFolder string, validExtensions []string, overwriteExistingFiles bool) error {
	// Connect to the NFS server
	nfsClient := NewNfsClient(s.config.Host, s.config.Port)
	if err := nfsClient.Connect(time.Duration(time.Second * 10)); err != nil {
		return fmt.Errorf("could not connect to NFS server %s: %w", s.config.Host, err)
	}
	defer nfsClient.Disconnect()

	// Instantiate an NFS Folder
	nfsFolder := NewNfsFolder(s.config.Folder, nfsClient)

	// Fetch all files in the folder
	allFiles, err := nfsFolder.FetchFiles(s.config.Folder, validExtensions, true)
	if err != nil {
		return fmt.Errorf("could not fetch files from folder %s on NFS server %s: %w", s.config.Folder, s.config.Host, err)
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
