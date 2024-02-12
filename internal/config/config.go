package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v8"
	ConfigBuilder "github.com/keloran/go-config"
)

type Config struct {
	Services
	ConfigBuilder.Config
}

type Services struct {
	ShortService string `env:"SHORT_SERVICE" envDefault:"short-service.1tn-pw:3000"`
}

func BuildServices(cfg *Config) error {
	services := &Services{}
	if err := env.Parse(services); err != nil {
		return logs.Errorf("Error parsing services: %v", err)
	}
	cfg.Services = *services
	return nil
}

func Build() (*Config, error) {
	cfg := &Config{}
	gcc, err := ConfigBuilder.Build(ConfigBuilder.Local)
	if err != nil {
		return nil, logs.Errorf("Error building config: %v", err)
	}

	cfg.Config = *gcc
	if err := BuildServices(cfg); err != nil {
		return nil, logs.Errorf("Error building services: %v", err)
	}

	return cfg, nil
}
