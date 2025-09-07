package config

import (
	"os"

	"github.com/bjw-s-labs/bookshift/pkg/util"
)

type Config struct {
	LogLevel               string   `yaml:"log_level"`
	TargetFolder           string   `yaml:"target_folder" validate:"required"`
	OverwriteExistingFiles bool     `yaml:"overwrite_existing_files"`
	ValidExtensions        []string `yaml:"valid_extensions" validate:"required"`
	Sources                []Source `yaml:"sources"`
	Concurrency            int      `yaml:"concurrency"`
}

func (cfg *Config) Load(path string) error {
	ymlFile, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := util.UnmarshalYamlIntoStruct(string(ymlFile), cfg); err != nil {
		return err
	}

	return nil
}
