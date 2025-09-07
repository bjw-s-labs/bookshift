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

type RunCommand struct {
	DryRun     bool `help:"Show what would be done without making changes"`
	NoProgress bool `help:"Disable progress bars during downloads"`
}

func (r *RunCommand) Run(cfg *config.Config, logger *slog.Logger) error {
	// Apply runtime toggles
	util.DryRun = r.DryRun
	util.ShowProgressBars = !r.NoProgress

	numberOfFilesAtStart, err := countFiles(cfg.TargetFolder, cfg.ValidExtensions, true)
	if err != nil {
		return err
	}

	for _, src := range cfg.Sources {
		switch src.Type {
		case "nfs":
			cfgNfs, ok := src.Config.(*config.NfsNetworkShareConfig)
			if !ok {
				logger.Error("invalid configuration type for NFS source")
				continue
			}
			if err := doNfs(cfgNfs, cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from NFS share", "error", err)
				continue
			}

		case "smb":
			cfgSmb, ok := src.Config.(*config.SmbNetworkShareConfig)
			if !ok {
				logger.Error("invalid configuration type for SMB source")
				continue
			}
			if err := doSmb(cfgSmb, cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				logger.Error("failed to sync from SMB share", "error", err)
				continue
			}

		case "imap":
			cfgImap, ok := src.Config.(*config.ImapConfig)
			if !ok {
				logger.Error("invalid configuration type for IMAP source")
				continue
			}
			if err := doImap(cfgImap, cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
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
