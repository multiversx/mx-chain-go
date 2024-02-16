package notifier

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

type nodesConfigProviderAPI struct {
	*nodesConfigProvider
	stakingV4Step2Epoch          uint32
	stakingV4Step3MaxNodesConfig config.MaxNodesChangeConfig
}

// NewNodesConfigProviderAPI returns a new instance of nodes config provider for API calls only, which provides the current
// max nodes change config based on the current epoch
func NewNodesConfigProviderAPI(
	epochNotifier process.EpochNotifier,
	cfg config.EnableEpochs,
) (*nodesConfigProviderAPI, error) {
	nodesCfgProvider, err := NewNodesConfigProvider(epochNotifier, cfg.MaxNodesChangeEnableEpoch)
	if err != nil {
		return nil, err
	}

	stakingV4Step3MaxNodesConfig, err := getStakingV4Step3MaxNodesConfig(nodesCfgProvider.allNodesConfigs, cfg.StakingV4Step3EnableEpoch)
	if err != nil {
		return nil, err
	}

	return &nodesConfigProviderAPI{
		nodesConfigProvider:          nodesCfgProvider,
		stakingV4Step2Epoch:          cfg.StakingV4Step2EnableEpoch,
		stakingV4Step3MaxNodesConfig: stakingV4Step3MaxNodesConfig,
	}, nil
}

func getStakingV4Step3MaxNodesConfig(
	allNodesConfigs []config.MaxNodesChangeConfig,
	stakingV4Step3EnableEpoch uint32,
) (config.MaxNodesChangeConfig, error) {
	for _, cfg := range allNodesConfigs {
		if cfg.EpochEnable == stakingV4Step3EnableEpoch {
			return cfg, nil
		}
	}

	return config.MaxNodesChangeConfig{}, errNoMaxNodesConfigChangeForStakingV4
}

// GetCurrentNodesConfig retrieves the current configuration of nodes. However, when invoked during epoch stakingV4 step 2
// through API calls, it will provide the nodes configuration as it will appear in epoch stakingV4 step 3. This adjustment
// is made because, with the transition to step 3 at the epoch change, the maximum number of nodes will be reduced.
// Therefore, calling this API during step 2 aims to offer a preview of the upcoming epoch, accurately reflecting the
// adjusted number of nodes that will qualify from the auction.
func (ncp *nodesConfigProviderAPI) GetCurrentNodesConfig() config.MaxNodesChangeConfig {
	ncp.mutex.RLock()
	defer ncp.mutex.RUnlock()

	if ncp.currentNodesConfig.EpochEnable == ncp.stakingV4Step2Epoch {
		return ncp.stakingV4Step3MaxNodesConfig
	}

	return ncp.currentNodesConfig
}

// IsInterfaceNil checks if the underlying pointer is nil
func (ncp *nodesConfigProviderAPI) IsInterfaceNil() bool {
	return ncp == nil
}
