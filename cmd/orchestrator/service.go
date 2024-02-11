package main

import (
	"github.com/1tn-pw/orchestrator/internal/service"
	"github.com/bugfixes/go-bugfixes/logs"
	ConfigBuilder "github.com/keloran/go-config"
)

var (
	BuildVersion = "dev"
	BuildHash    = "unknown"
	ServiceName  = "service"
)

func main() {
	logs.Local().Infof("Starting %s version: %s, hash: %s", ServiceName, BuildVersion, BuildHash)

	cfg, err := ConfigBuilder.Build(ConfigBuilder.Local)
	if err != nil {
		_ = logs.Local().Errorf("Error building config: %s", err.Error())
		return
	}

	if err := service.New(cfg).Start(); err != nil {
		_ = logs.Local().Errorf("Error starting service: %s", err.Error())
		return
	}
}
