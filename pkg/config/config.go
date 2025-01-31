package config

import (
	"os"
	"strings"

	"github.com/go-playground/validator"
	"github.com/goccy/go-yaml"
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

	validate := validator.New()
	dec := yaml.NewDecoder(
		strings.NewReader(string(ymlFile)),
		yaml.Validator(validate),
		yaml.Strict(),
	)
	if err := dec.Decode(cfg); err != nil {
		return err
	}

	return nil
}
