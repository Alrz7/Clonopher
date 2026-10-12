package config

import (
	"clonopher/app/jevkit"
	"clonopher/app/services/composer"
	"clonopher/internal/loggy"
)

type Config struct {
	App            *App                       `koanf:"app"`
	Providers      map[string]jevkit.Provider `koanf:"providers"`
	ActiveProvider string                     `koanf:"active_provider"`
	ActiveModel    jevkit.JevModel                     `koanf:"active_model"`
	Api            *ApiOptions                `koanf:"api"`
	Cli            *CliOptions                `koanf:"cli"`
}

type App struct {
	RunMode string `koanf:"-"`
	Version string `koanf:"version"`
	Env     string `koanf:"env"`
}

type ApiOptions struct {
	Port int `koanf:"port"`
}

type CliOptions struct {
	Head string             `koanf:"head"`
	Src  string             `koanf:"src"`
	Info *composer.FileInfo `koanf:"info"`
}

func (c *Config) GetActiveProvider() *jevkit.Provider {
	provider, ok := c.Providers[c.ActiveProvider]
	if !ok {
		loggy.Fatalf("Active-Provider <%v> Is NOT initiated inside configs...", c.ActiveProvider)
	}
	return &provider
}