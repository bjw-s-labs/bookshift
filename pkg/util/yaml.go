package util

import (
	"strings"

	"github.com/go-playground/validator"
	"github.com/goccy/go-yaml"
)

func UnmarshalYamlIntoStruct(yamlData string, output interface{}) error {
	validate := validator.New()
	dec := yaml.NewDecoder(
		strings.NewReader(yamlData),
		yaml.Validator(validate),
		yaml.Strict(),
	)
	if err := dec.Decode(output); err != nil {
		return err
	}
	return nil
}
