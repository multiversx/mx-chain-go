package config

import (
	"github.com/multiversx/mx-chain-core-go/core"
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
