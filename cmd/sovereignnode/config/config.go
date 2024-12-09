package config

import "github.com/multiversx/mx-chain-go/config"

// SovereignConfig holds sovereign node config
type SovereignConfig struct {
	*config.Configs
	SovereignExtraConfig *config.SovereignConfig
}
