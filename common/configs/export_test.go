package configs

import "github.com/multiversx/mx-chain-go/config"

// GetOrderedConfigsByEpoch -
func (pce *processConfigsByEpoch) GetOrderedConfigsByEpoch(epoch uint32) config.ProcessConfigByEpoch {
	if len(pce.orderedConfigByEpoch) == 0 {
		return config.ProcessConfigByEpoch{}
	}

	return pce.orderedConfigByEpoch[epoch]
}

// GetOrderedEpochStartConfigByEpoch -
func (cc *commonConfigs) GetOrderedEpochStartConfigByEpoch(epoch uint32) config.EpochStartConfigByEpoch {
	if len(cc.orderedEpochStartConfigByEpoch) == 0 {
		return config.EpochStartConfigByEpoch{}
	}

	return cc.orderedEpochStartConfigByEpoch[epoch]
}
