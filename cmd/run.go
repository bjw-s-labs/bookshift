package cmd

import (
	"log/slog"

	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/bjw-s-labs/bookshift/pkg/kobo"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/imap"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/nfs"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/smb"
	"github.com/bjw-s-labs/bookshift/pkg/util"
)

type RunCommand struct{}

func (*RunCommand) Run(cfg *config.Config, logger *slog.Logger) error {
	numberOfFilesAtStart, err := util.CountFilesInFolder(cfg.TargetFolder, cfg.ValidExtensions, true)
	if err != nil {
		return err
	}

	for _, src := range cfg.Sources {
		switch src.Type {
		case "nfs":
			nfsSyncer := nfs.NewNfsSyncer(src.Config.(*config.NfsNetworkShareConfig))
			if err := nfsSyncer.Run(cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from NFS share", "error", err)
				continue
			}

		case "smb":
			smbSyncer := smb.NewSmbSyncer(src.Config.(*config.SmbNetworkShareConfig))
			if err := smbSyncer.Run(cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from SMB share", "error", err)
				continue
			}

		case "imap":
			imapSyncer := imap.NewImapSyncer(src.Config.(*config.ImapConfig))
			if err := imapSyncer.Run(cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from IMAP server", "error", err)
				continue
			}
		}
	}

	numberOfFilesAtEnd, err := util.CountFilesInFolder(cfg.TargetFolder, cfg.ValidExtensions, true)
	if err != nil {
		return err
	}

	slog.Info("Processed all configured sources", "books_downloaded", numberOfFilesAtEnd-numberOfFilesAtStart)

	if numberOfFilesAtEnd > numberOfFilesAtStart {
		if kobo.IsKoboDevice() {
			slog.Info("Kobo device detected, updating library")
			if err := kobo.UpdateLibrary(); err != nil {
				return err
			}
		}
	}

	return nil
}
