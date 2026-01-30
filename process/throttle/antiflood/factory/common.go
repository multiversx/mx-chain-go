package factory

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
)

func floodPreventerConfigFetcher(confHandler common.AntifloodConfigsHandler, identifier string) config.FloodPreventerConfig {
	currentConfig := confHandler.GetCurrentConfig()
	switch identifier {
	case fastReactingIdentifier:
		return currentConfig.FastReacting
	case slowReactingIdentifier:
		return currentConfig.SlowReacting
	case outOfSpecsIdentifier:
		return currentConfig.OutOfSpecs
	default:
		// this case should not happen
		return currentConfig.FastReacting
	}
}
