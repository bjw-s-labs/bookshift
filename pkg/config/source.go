package config

import (
	"github.com/bjw-s-labs/bookshift/pkg/util"
	"github.com/goccy/go-yaml"
)

type Source struct {
	Type   string      `yaml:"type" validate:"oneof=smb nfs"`
	Config interface{} `yaml:"config" validate:"required"`
}

func (src *Source) UnmarshalYAML(input []byte) error {
	var tmpSource struct {
		Type   string
		Config interface{}
	}
	if err := yaml.Unmarshal(input, &tmpSource); err != nil {
		return err
	}

	switch tmpSource.Type {
	case "nfs":
		var config struct {
			Type   string
			Config NfsNetworkShareConfig
		}

		if err := util.UnmarshalYamlIntoStruct(string(input), &config); err != nil {
			return err
		}
		src.Type = config.Type
		src.Config = config.Config

	case "smb":
		var config struct {
			Type   string
			Config SmbNetworkShareConfig
		}
		if err := util.UnmarshalYamlIntoStruct(string(input), &config); err != nil {
			return err
		}
		src.Type = config.Type
		src.Config = config.Config

	case "imap":
		var config struct {
			Type   string
			Config ImapConfig
		}
		if err := util.UnmarshalYamlIntoStruct(string(input), &config); err != nil {
			return err
		}
		src.Type = config.Type
		src.Config = config.Config
	}

	return nil
}
