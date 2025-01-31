package config

type Source struct {
	Type   string      `yaml:"type"`
	Config interface{} `yaml:"config" validate:"required"`
}

func (src *Source) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var tmpSource struct {
		Type   string
		Config interface{}
	}
	if err := unmarshal(&tmpSource); err != nil {
		return err
	}

	switch tmpSource.Type {
	case "nfs":
		var config struct {
			Type   string
			Config NfsNetworkShareConfig
		}
		if err := unmarshal(&config); err != nil {
			return err
		}
		src.Type = config.Type
		src.Config = config.Config

	case "smb":
		var config struct {
			Type   string
			Config SmbNetworkShareConfig
		}
		if err := unmarshal(&config); err != nil {
			return err
		}
		src.Type = config.Type
		src.Config = config.Config
	}

	return nil
}
