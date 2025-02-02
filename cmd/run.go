package cmd

import (
	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/imap"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/nfs"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/smb"
)

type RunCommand struct{}

func (*RunCommand) Run(cfg *config.Config) error {
	for _, src := range cfg.Sources {
		switch src.Type {
		case "nfs":
			nfsSyncer := nfs.NewNfsSyncer(src.Config.(config.NfsNetworkShareConfig))
			if err := nfsSyncer.Run(cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				return err
			}

		case "smb":
			smbSyncer := smb.NewSmbSyncer(src.Config.(config.SmbNetworkShareConfig))
			if err := smbSyncer.Run(cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				return err
			}

		case "imap":
			smbSyncer := imap.NewImapSyncer(src.Config.(config.ImapConfig))
			if err := smbSyncer.Run(cfg.TargetFolder, cfg.ValidExtensions, cfg.OverwriteExistingFiles); err != nil {
				return err
			}
		}

	}
	return nil
}
