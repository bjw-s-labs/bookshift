package cmd

import (
	"os"

	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/bjw-s-labs/bookshift/pkg/kobo"
)

type KoboCommand struct {
	UpdateLibrary KoboUpdateLibraryCommand `cmd:"" help:"Update the library on your Kobo e-reader"`
	Uninstall     KoboUninstallCommand     `cmd:"" help:"Uninstall BookShift from your Kobo e-reader"`
}

type KoboInstallCommand struct{}
type KoboUninstallCommand struct{}
type KoboUpdateLibraryCommand struct{}

func (*KoboInstallCommand) Run(cfg *config.Config) error {
	return nil
}

func (*KoboUninstallCommand) Run(cfg *config.Config) error {
	const nickelMenuPath = "/mnt/onboard/.adds/nm"
	const nickelMenuConfigPath = nickelMenuPath + "/bookshift"
	const udevRulesFilePath = "/etc/udev/rules.d/97-bookshift.rules"

	os.Remove(nickelMenuConfigPath)
	os.Remove(udevRulesFilePath)
	os.RemoveAll("/usr/local/bookshift")

	return nil
}

func (*KoboUpdateLibraryCommand) Run(cfg *config.Config) error {
	return kobo.UpdateLibrary()
}
