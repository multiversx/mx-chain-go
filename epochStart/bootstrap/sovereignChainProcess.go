package bootstrap

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
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

	scesb.bootStrapShardRequester = &sovereignBootStrapShardRequester{
		scesb,
	}

	scesb.getDataToSyncMethod = scesb.getDataToSync
	scesb.shardForLatestEpochComputer = scesb
	return scesb, nil
}

// todo : probably delete this
func (scesb *sovereignChainEpochStartBootstrap) getDataToSync(
	_ data.EpochStartShardDataHandler,
	shardNotarizedHeader data.ShardHeaderHandler,
) (*dataToSync, error) {
	return &dataToSync{
		ownShardHdr:       shardNotarizedHeader,
		rootHashToSync:    shardNotarizedHeader.GetRootHash(),
		withScheduled:     false,
		additionalHeaders: nil,
	}, nil
}

// GetShardIDForLatestEpoch returns the shard ID for the latest epoch
func (scesb *sovereignChainEpochStartBootstrap) GetShardIDForLatestEpoch() (uint32, bool, error) {
	return core.SovereignChainShardId, false, nil
}

func (e *sovereignChainEpochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
