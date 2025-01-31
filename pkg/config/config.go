package config

import (
	"fmt"
	"os"

	"github.com/gookit/validate"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel               string   `yaml:"log_level"`
	TargetFolder           string   `yaml:"target_folder" validate:"required"`
	OverwriteExistingFiles bool     `yaml:"overwrite_existing_files"`
	Sources                []Source `yaml:"sources"`
}

func (cfg *Config) Load(path string) error {
	ymlFile, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(ymlFile, &cfg); err != nil {
		return err
	}

	// Set defaults
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Validate sources separately because validator somehow doesn't validate the list items themselves
	for i, source := range cfg.Sources {
		if err := source.Validate(); err != nil {
			return fmt.Errorf("Sources.%d.Config %w", i, err)
		}
	}

	return nil
}

func (cfg *Config) Validate() validate.Errors {
	v := validate.Struct(cfg)
	v.StopOnError = false
	return v.ValidateE()
}
