package smb

import (
	"context"
	"fmt"
	"time"

	"github.com/bjw-s-labs/bookshift/pkg/config"
)

type SmbSyncer struct {
	config *config.SmbNetworkShareConfig
}

func NewSmbSyncer(shareConfig *config.SmbNetworkShareConfig) *SmbSyncer {
	// Set default port
	if !(shareConfig.Port > 0) {
		shareConfig.Port = 445
	}

	return &SmbSyncer{
		config: shareConfig,
	}
}

func (s *SmbSyncer) Run(targetFolder string, validExtensions []string, overwriteExistingFiles bool) error {
	return s.RunContext(context.Background(), targetFolder, validExtensions, overwriteExistingFiles)
}

func (s *SmbSyncer) RunContext(ctx context.Context, targetFolder string, validExtensions []string, overwriteExistingFiles bool) error {
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

	// Connect to the SMB server
	smbConnection := SmbConnection{
		Host:     s.config.Host,
		Port:     s.config.Port,
		Username: s.config.Username,
		Password: s.config.Password,
		Domain:   s.config.Domain,
	}

	if err := smbConnect(&smbConnection); err != nil {
		return fmt.Errorf("could not connect to SMB server %s: %w", s.config.Host, err)
	}
	defer smbDisconnect(&smbConnection)

	// Connect to the share
	smbShareConnection := newSmbShare(s.config.Share, &smbConnection)
	if err := smbShareConnect(smbShareConnection); err != nil {
		return fmt.Errorf("could not connect to SMB share %s. %w", s.config.Share, err)
	}
	defer smbShareDisconnect(smbShareConnection)

	// Fetch all files in the share
	allFiles, err := smbShareConnection.FetchFiles(s.config.Folder, validExtensions, true)
	if err != nil {
		return err
	}

	// Download all files
	for _, file := range allFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := file.Download(
			targetFolder,
			"",
			overwriteExistingFiles,
			s.config.KeepFolderStructure,
			s.config.RemoveFilesAfterDownload,
		); err != nil {
			return err
		}
	}

	return nil
}
