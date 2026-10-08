package config

import "clonopher/app/jevkit"

type Config struct {
	App       *App              `koanf:"app"`
	Providers []jevkit.Provider `koanf:"providers"`
}

type App struct {
	Version string `koanf:"version"`
	Env     string `koanf:"env"`
}
