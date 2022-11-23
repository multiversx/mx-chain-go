package bootstrap

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/errors"
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

	scesb.getDataToSyncMethod = scesb.getDataToSync

	return scesb, nil
}

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
