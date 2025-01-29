package cmd

import (
	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/nfs"
	"github.com/bjw-s-labs/bookshift/pkg/syncer/smb"
)

type RunCommand struct{}

func (*RunCommand) Run(cfg *config.Config) error {
	for _, nfsShare := range cfg.NfsShares {
		nfsSyncer := nfs.NewNfsSyncer(nfsShare)
		err := nfsSyncer.Run(cfg.TargetFolder, cfg.OverwriteExistingFiles)
		if err != nil {
			return err
		}
	}

	for _, smbShare := range cfg.SmbShares {
		smbSyncer := smb.NewSmbSyncer(smbShare)
		err := smbSyncer.Run(cfg.TargetFolder, cfg.OverwriteExistingFiles)
		if err != nil {
			return err
		}
	}
	return nil
}
