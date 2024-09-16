package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/errors"
)

type sovereignChainEpochStartBootstrap struct {
	*epochStartBootstrap
}

// NewSovereignChainEpochStartBootstrap creates a new instance of sovereignChainEpochStartBootstrap
func NewSovereignChainEpochStartBootstrap(epochStartBootstrap *epochStartBootstrap) (*sovereignChainEpochStartBootstrap, error) {
	if epochStartBootstrap == nil {
		return nil, errors.ErrNilEpochStartBootstrapper
	}

	scesb := &sovereignChainEpochStartBootstrap{
		epochStartBootstrap,
	}

	scesb.bootStrapShardProcessor = &sovereignBootStrapShardRequester{
		scesb,
	}

	scesb.shardForLatestEpochComputer = scesb
	return scesb, nil
}

// GetShardIDForLatestEpoch returns the shard ID for the latest epoch
func (scesb *sovereignChainEpochStartBootstrap) GetShardIDForLatestEpoch() (uint32, bool, error) {
	return core.SovereignChainShardId, false, nil
}

func (e *sovereignChainEpochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
