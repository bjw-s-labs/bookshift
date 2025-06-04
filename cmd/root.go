package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/bjw-s-labs/bookshift/pkg/config"
	"github.com/lmittmann/tint"
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
	Kobo    KoboCommand `cmd:"" help:"Manage BookShift on your Kobo e-reader"`
	Version VersionFlag `       help:"Print version information and quit" short:"v" name:"version"`
}

func Execute() error {
	if version == "" {
		version = "development"
	}

	// Configure the default logger
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	logger := slog.New(
		tint.NewHandler(
			os.Stderr,
			&tint.Options{
				Level: logLevel,
			},
		),
	)
	slog.SetDefault(logger)

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

	// Create Configuration object and set defaults
	appConfig := config.Config{
		LogLevel: "info",
	}

	// Load the configuration from the config file
	if err := appConfig.Load(cli.ConfigFile); err != nil {
		errors := strings.Split(err.Error(), "\n")
		for _, e := range errors {
			slog.Error(e)
		}
		os.Exit(1)
	}

	// Set the log level based on the configuration
	logLevel.UnmarshalText([]byte(appConfig.LogLevel))

	slog.Debug("Running")

	// Run the application
	if err := ctx.Run(&appConfig, logger); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return nil
}
