package config

import (
	"clonopher/internal/loggy"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var K = koanf.New(".")
var Default = Config{
	App: &App{Heads: map[string]string{}},
	Api: &ApiOptions{},
}

func (c *Config) Setup(configPath string) {
	// configPath := filepath.Join(configDir, "marble.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// loggy.Fatal(err)
		return
	}

	err := K.Load(
		file.Provider(configPath),
		yaml.Parser(),
	)
	if err != nil {
		loggy.Fatal(err)
	}

	err = K.Unmarshal("", c)
	if err != nil {
		loggy.Fatal(err)
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
}
