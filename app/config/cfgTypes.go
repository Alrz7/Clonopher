package config

import (
	"clonopher/app/classifiers"
	"clonopher/app/providers"
	"clonopher/internal/loggy"
)

type Config struct {
	App          *App             `koanf:"app"`
	Api          *ApiOptions      `koanf:"api"`
	Cli          *CliOptions      `koanf:"cli"`
	JevProviders *ProviderOptions `koanf:"jev_providers"`
	LlmProviders *ProviderOptions `koanf:"llm_providers"`
}

type App struct {
	RunMode      string `koanf:"-"`
	Version      string `koanf:"version"`
	Env          string `koanf:"env"`
	ActionPolicy string `koanf:"-"`
}

type ProviderOptions struct {
	All            map[string]*providers.Provider `koanf:"all"`
	ActiveProvider string                         `koanf:"active_provider"`
	ActiveModel    string                         `koanf:"active_model"`
}
type ApiOptions struct {
	Port int `koanf:"port"`
}

type CliOptions struct {
	Head string                  `koanf:"head"`
	Src  string                  `koanf:"src"`
	Info *classifiers.ObjectInfo `koanf:"info"`
}

func (c *Config) GetActiveJevProvider() *providers.Provider {
	provider, ok := c.JevProviders.All[c.JevProviders.ActiveProvider]
	if !ok {
		loggy.Fatalf("Active-JevProvider <%v> Is NOT initiated inside configs...", c.JevProviders.ActiveProvider)
	}
	return provider
}
func (c *Config) GetActiveLlmProvider() *providers.Provider {
	provider, ok := c.LlmProviders.All[c.LlmProviders.ActiveProvider]
	if !ok {
		loggy.Fatalf("Active-LlmProvider <%v> Is NOT initiated inside configs...", c.JevProviders.ActiveProvider)
	}
	return provider
}
