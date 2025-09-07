package cmd

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	// Root cancellation context (Ctrl+C / SIGTERM)
	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	numberOfFilesAtStart, err := countFiles(cfg.TargetFolder, cfg.ValidExtensions, true)
	if err != nil {
		return err
	}

	// Process sources concurrently with a bound to avoid overwhelming endpoints
	conc := cfg.Concurrency
	if conc <= 0 {
		conc = 3
	}
	if len(cfg.Sources) < conc {
		conc = len(cfg.Sources)
	}
	if conc < 1 {
		conc = 1
	}
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup

	for _, s := range cfg.Sources {
		src := s // capture
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Per-source context (with timeout if configured)
			ctx := rootCtx
			switch src.Type {
			case "nfs":
				cfgNfs, ok := src.Config.(*config.NfsNetworkShareConfig)
				if !ok {
					logger.Error("invalid configuration type for NFS source")
					return
				}
				if cfgNfs.TimeoutSeconds > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, time.Duration(cfgNfs.TimeoutSeconds)*time.Second)
					defer cancel()
				}
				if err := doNfs(ctx, cfgNfs, cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
					logger.Error("failed to sync from NFS share", "error", err)
				}

			case "smb":
				cfgSmb, ok := src.Config.(*config.SmbNetworkShareConfig)
				if !ok {
					logger.Error("invalid configuration type for SMB source")
					return
				}
				if cfgSmb.TimeoutSeconds > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, time.Duration(cfgSmb.TimeoutSeconds)*time.Second)
					defer cancel()
				}
				if err := doSmb(ctx, cfgSmb, cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
					logger.Error("failed to sync from SMB share", "error", err)
				}

			case "imap":
				cfgImap, ok := src.Config.(*config.ImapConfig)
				if !ok {
					logger.Error("invalid configuration type for IMAP source")
					return
				}
				if cfgImap.TimeoutSeconds > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, time.Duration(cfgImap.TimeoutSeconds)*time.Second)
					defer cancel()
				}
				if err := doImap(ctx, cfgImap, cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
					logger.Error("failed to sync from IMAP server", "error", err)
				}
			}
		}()
	}
	wg.Wait()

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
	doNfs      = func(ctx context.Context, cfg *config.NfsNetworkShareConfig, target string, valid []string, overwrite bool) error {
		return nfs.NewNfsSyncer(cfg).RunContext(ctx, target, valid, overwrite)
	}
	doSmb = func(ctx context.Context, cfg *config.SmbNetworkShareConfig, target string, valid []string, overwrite bool) error {
		return smb.NewSmbSyncer(cfg).RunContext(ctx, target, valid, overwrite)
	}
	doImap = func(ctx context.Context, cfg *config.ImapConfig, target string, valid []string, overwrite bool) error {
		return imap.NewImapSyncer(cfg).RunContext(ctx, target, valid, overwrite)
	}
	isKoboDevice      = kobo.IsKoboDevice
	updateKoboLibrary = kobo.UpdateLibrary
)
