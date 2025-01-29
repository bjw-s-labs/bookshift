package nfs

import (
	"fmt"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type NfsSyncer struct {
	config *config.NfsNetworkShare
}

func NewNfsSyncer(shareConfig config.NfsNetworkShare) *NfsSyncer {
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
	err := nfsClient.Connect()
	if err != nil {
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
		err := file.Download(targetFolder, file.CleanFileName(), overwriteExistingFiles, s.config.KeepFolderStructure, s.config.RemoveFilesAfterDownload)
		if err != nil {
			return err
		}
	}

	// // Connect to the share
	// smbShareConnection := SmbShareConnection{
	// 	Share:         s.config.Share,
	// 	SmbConnection: &smbConnection,
	// }
	// err = smbShareConnection.Connect()
	// if err != nil {
	// 	return fmt.Errorf("could not connect to SMB share %s. %w", s.config.Share, err)
	// }
	// defer smbShareConnection.Disconnect()

	// // Fetch all files in the share
	// allFiles, err := smbShareConnection.FetchFiles(s.config.Folder, true)
	// if err != nil {
	// 	return err
	// }

	// // Download all files
	// for _, file := range allFiles {
	// 	err := file.Download(targetFolder, file.CleanFileName(), overwriteExistingFiles, s.config.KeepFolderStructure, s.config.RemoveFilesAfterDownload)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
