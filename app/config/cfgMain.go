package config

import (
	"clonopher/internal/loggy"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/fileflow"
)

var K = koanf.New(".")
var Default = Config{
	App: &App{Roots: map[string]string{}},
	Api: &ApiOptions{},
	Cli: &CliOptions{},
}

func (c *Config) Setup(configPath string) (bool, error) {
	// configPath := filepath.Join(configDir, "clonopher.config.yaml")

	if !fileflow.Exists(configPath) {
		return false, loggy.Sayf("config file <%v> is not defined or accessible!", configPath)
	}

	err := K.Load(
		file.Provider(configPath),
		yaml.Parser(),
	)
	if err != nil {
		return false, err
	}

	err = K.Unmarshal("", c)
	if err != nil {
		return false, err
	}

	for title, prv := range c.JevProviders.All {
		if prv.Name == "" {
			prv.Name = title
		}
	}
	for title, prv := range c.LlmProviders.All {
		if prv.Name == "" {
			prv.Name = title
		}
	}
	return true, nil
}
