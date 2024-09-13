package bootstrap

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
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

func (e *sovereignChainEpochStartBootstrap) syncHeadersFrom(meta data.MetaHeaderHandler) (map[string]data.HeaderHandler, error) {
	hashesToRequest := make([][]byte, 0, 1)
	shardIds := make([]uint32, 0, 1)

	if meta.GetEpoch() > e.startEpoch+1 { // no need to request genesis block
		hashesToRequest = append(hashesToRequest, meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())
		shardIds = append(shardIds, core.SovereignChainShardId)
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeToWaitForRequestedData)
	err := e.headersSyncer.SyncMissingHeadersByHash(shardIds, hashesToRequest, ctx)
	cancel()
	if err != nil {
		return nil, err
	}

	syncedHeaders, err := e.headersSyncer.GetHeaders()
	if err != nil {
		return nil, err
	}

	if meta.GetEpoch() == e.startEpoch+1 {
		syncedHeaders[string(meta.GetEpochStartHandler().GetEconomicsHandler().GetPrevEpochStartHash())] = &block.SovereignChainHeader{}
	}

	return syncedHeaders, nil
}

func (e *sovereignChainEpochStartBootstrap) IsInterfaceNil() bool {
	return e == nil
}
