package helpers

import (
	"gopkg.in/yaml.v3"
)

func ToYAML(data map[string]any) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
