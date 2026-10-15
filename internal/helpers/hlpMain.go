package helpers

import (
	"slices"

	"gopkg.in/yaml.v3"
)

func ToYAML(data map[string]any) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func IsNotOneOf[T comparable](obj T, src ...T) bool {
	return !slices.Contains(src, obj)
}
