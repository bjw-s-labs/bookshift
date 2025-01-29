package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/bjw-s-labs/bookshift/pkg/config"
)

const appName = "bookshift"
const appDescription = "A tool to download e-books from a share to your e-reader"
const defaultConfigFile = "config.yaml"

var version string

type VersionFlag string

func (v VersionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                       { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

type CLI struct {
	ConfigFile string `short:"c" help:"Location of the configuration file." default:"${config_file}" type:"existingfile" env:"BOOKSHIFT_CONFIG_FILE"`

	Run     RunCommand  `cmd:"" help:"Transfer books to your e-reader"`
	Version VersionFlag `       help:"Print version information and quit" short:"v" name:"version"`
}

func Execute() error {
	if version == "" {
		version = "development"
	}

	cli := CLI{
		Version: VersionFlag(version),
	}

	// Display help if no args are provided instead of an error message
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "--help")
	}

	ctx := kong.Parse(&cli,
		kong.Name(appName),
		kong.Description(appDescription),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version":     string(cli.Version),
			"config_file": defaultConfigFile,
		},
	)

	var appConfig config.Config

	err := appConfig.Load(cli.ConfigFile)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = ctx.Run(&appConfig)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return nil
}
