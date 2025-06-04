package config

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

type Source struct {
	Type   string       `yaml:"type" validate:"oneof=smb nfs imap"`
	Config SourceConfig `yaml:"config" validate:"required"`
}

// UnmarshalYAML implements custom YAML unmarshalling for Source,
// selecting the correct config struct based on the "type" field.
func (src *Source) UnmarshalYAML(input []byte) error {
	var raw map[string]interface{}
	if err := yaml.Unmarshal(input, &raw); err != nil {
		return err
	}

	typeVal, ok := raw["type"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid 'type' field in source config")
	}

	var configPtr interface{}
	switch typeVal {
	case "nfs":
		configPtr = &NfsNetworkShareConfig{}
	case "smb":
		configPtr = &SmbNetworkShareConfig{}
	case "imap":
		configPtr = &ImapConfig{}
	default:
		return fmt.Errorf("unsupported source type: %s", typeVal)
	}

	configRaw, err := yaml.Marshal(raw["config"])
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(configRaw, configPtr); err != nil {
		return err
	}

	src.Type = typeVal
	src.Config = configPtr
	return nil
}
