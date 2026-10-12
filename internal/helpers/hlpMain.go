package helpers

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

func DoesSupportModel(provider string, supportedModels []string, targetModel string) error {
	if !slices.Contains(supportedModels, targetModel) {
		return fmt.Errorf("provider [%v] does Not support model [%v]", provider, targetModel)
	}
	return nil
}

func ToYAML(data map[string]any) (string, error) {
	b, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
