package aerocast

import (
	"github.com/krigsherre/aerocast/pkg/engine"
)

type Config = engine.Config

func DefaultConfig() Config {
	return engine.DefaultConfig()
}
