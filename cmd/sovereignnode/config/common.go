package config

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
)

// LoadSovereignNotifierConfig returns NotifierConfig by reading it from the provided file
func LoadSovereignNotifierConfig(filepath string) (*NotifierConfig, error) {
	cfg := &NotifierConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadSovereignGeneralConfig returns the extra config necessary by sovereign by reading it from the provided file
func LoadSovereignGeneralConfig(filepath string) (*config.SovereignConfig, error) {
	cfg := &config.SovereignConfig{}
	err := core.LoadTomlFile(cfg, filepath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
