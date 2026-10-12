package config

import (
	"clonopher/internal/loggy"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var K = koanf.New(".")

func (c *Config) Setup(configPath string) {
	// configPath := filepath.Join(configDir, "marble.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		loggy.Fatal(err)
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

}
