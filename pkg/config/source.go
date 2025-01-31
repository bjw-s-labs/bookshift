package config

import (
	"github.com/gookit/validate"
	"gopkg.in/yaml.v3"
)

type Source struct {
	Type   string      `yaml:"type"`
	Config interface{} `yaml:"config" validate:"required"`
}

// Interface compliance
var _ yaml.Unmarshaler = &Source{}

func (src *Source) UnmarshalYAML(node *yaml.Node) error {
	var t struct {
		Type string `yaml:"type"`
	}
	if err := node.Decode(&t); err != nil {
		return err
	}

	switch t.Type {
	case "nfs":
		var c struct {
			Config NfsNetworkShareConfig `yaml:"config"`
		}
		if err := node.Decode(&c); err != nil {
			return err
		}
		src.Type = t.Type
		src.Config = c.Config

	case "smb":
		var c struct {
			Config SmbNetworkShareConfig `yaml:"config"`
		}
		if err := node.Decode(&c); err != nil {
			return err
		}
		src.Type = t.Type
		src.Config = c.Config
	}
	return nil
}

func (src *Source) Validate() validate.Errors {
	v := validate.Struct(src.Config)
	v.StopOnError = false
	return v.ValidateE()
}
