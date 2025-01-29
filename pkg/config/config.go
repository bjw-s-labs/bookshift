package config

import (
	"github.com/go-playground/sensitive"
	"github.com/gookit/validate"
	yaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	LogLevel               string            `koanf:"log_level"`
	TargetFolder           string            `koanf:"target_folder" validate:"required"`
	OverwriteExistingFiles bool              `koanf:"overwrite_existing_files"`
	NfsShares              []NfsNetworkShare `koanf:"nfs_shares"`
	SmbShares              []SmbNetworkShare `koanf:"smb_shares"`
}

type NfsNetworkShare struct {
	Host                     string `koanf:"host" validate:"required"`
	Port                     int    `koanf:"port"`
	Folder                   string `koanf:"folder" validate:"required"`
	KeepFolderStructure      bool   `koanf:"keep_folderstructure"`
	RemoveFilesAfterDownload bool   `koanf:"remove_files_after_download"`
}

type SmbNetworkShare struct {
	Host                     string           `koanf:"host" validate:"required"`
	Port                     int              `koanf:"port"`
	Username                 string           `koanf:"username"`
	Password                 sensitive.String `koanf:"password"`
	Domain                   string           `koanf:"domain"`
	Share                    string           `koanf:"share" validate:"required"`
	Folder                   string           `koanf:"folder" validate:"required"`
	KeepFolderStructure      bool             `koanf:"keep_folderstructure"`
	RemoveFilesAfterDownload bool             `koanf:"remove_files_after_download"`
}

func (cfg *Config) Load(path string) error {
	var k = koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return err
	}

	if err := k.Unmarshal("", cfg); err != nil {
		return err
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) Validate() validate.Errors {
	v := validate.Struct(cfg)
	v.StopOnError = false
	return v.ValidateE()
}
