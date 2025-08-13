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
	numberOfFilesAtStart, err := countFiles(cfg.TargetFolder, cfg.ValidExtensions, true)
	if err != nil {
		return err
	}

	for _, src := range cfg.Sources {
		switch src.Type {
		case "nfs":
			if err := doNfs(src.Config.(*config.NfsNetworkShareConfig), cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from NFS share", "error", err)
				continue
			}

		case "smb":
			if err := doSmb(src.Config.(*config.SmbNetworkShareConfig), cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from SMB share", "error", err)
				continue
			}

		case "imap":
			if err := doImap(src.Config.(*config.ImapConfig), cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from IMAP server", "error", err)
				continue
			}
		}
	}

	numberOfFilesAtEnd, err := countFiles(cfg.TargetFolder, cfg.ValidExtensions, true)
	if err != nil {
		return err
	}

	slog.Info("Processed all configured sources", "books_downloaded", numberOfFilesAtEnd-numberOfFilesAtStart)

	if numberOfFilesAtEnd > numberOfFilesAtStart {
		if isKoboDevice() {
			slog.Info("Kobo device detected, updating library")
			if err := updateKoboLibrary(); err != nil {
				return err
			}
		}
	}

	return nil
}

// test seams (overridable in tests)
var (
	countFiles = util.CountFilesInFolder
	doNfs      = func(cfg *config.NfsNetworkShareConfig, target string, valid []string, overwrite bool) error {
		return nfs.NewNfsSyncer(cfg).Run(target, valid, overwrite)
	}
	doSmb = func(cfg *config.SmbNetworkShareConfig, target string, valid []string, overwrite bool) error {
		return smb.NewSmbSyncer(cfg).Run(target, valid, overwrite)
	}
	doImap = func(cfg *config.ImapConfig, target string, valid []string, overwrite bool) error {
		return imap.NewImapSyncer(cfg).Run(target, valid, overwrite)
	}
	isKoboDevice      = kobo.IsKoboDevice
	updateKoboLibrary = kobo.UpdateLibrary
)
