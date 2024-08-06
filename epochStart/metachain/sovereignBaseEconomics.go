package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignBaseEconomics struct {
	*baseEconomics
}

func (e *sovereignBaseEconomics) startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, data.MetaHeaderHandler, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.TotalNumberOfShards())
	mapShardIdNonce[core.SovereignChainShardId] = e.genesisNonce

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	previousEpochStartSovHdr, err := process.GetSovereignChainHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	prevSovHdr, castOk := previousEpochStartSovHdr.(data.MetaHeaderHandler)
	if !castOk {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	if epoch == e.genesisEpoch {
		return mapShardIdNonce, prevSovHdr, nil
	}

	mapShardIdNonce[core.SovereignChainShardId] = prevSovHdr.GetNonce()
	return mapShardIdNonce, prevSovHdr, nil
}

func (e *sovereignBaseEconomics) startNoncePerShardFromLastCrossNotarized(metaNonce uint64, _ data.EpochStartHandler) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.TotalNumberOfShards())
	mapShardIdNonce[core.SovereignChainShardId] = metaNonce
	return mapShardIdNonce, nil
}

func (e *sovereignBaseEconomics) computeNumOfTotalCreatedBlocks(
	mapStartNonce map[uint32]uint64,
	mapEndNonce map[uint32]uint64,
) uint64 {
	totalNumBlocks := uint64(0)
	var blocksInShard uint64
	blocksPerShard := make(map[uint32]uint64)

	shardId := core.SovereignChainShardId

	blocksInShard = mapEndNonce[shardId] - mapStartNonce[shardId]
	blocksPerShard[shardId] = blocksInShard
	totalNumBlocks += blocksInShard
	log.Debug("computeNumOfTotalCreatedBlocks",
		"shardID", shardId,
		"prevEpochLastNonce", mapEndNonce[shardId],
		"epochLastNonce", mapStartNonce[shardId],
		"nbBlocksEpoch", blocksPerShard[shardId],
	)

	e.economicsDataNotified.SetNumberOfBlocks(totalNumBlocks)
	e.economicsDataNotified.SetNumberOfBlocksPerShard(blocksPerShard)

	return core.MaxUint64(1, totalNumBlocks)
}

func (e *sovereignBaseEconomics) maxPossibleNotarizedBlocks(currentRound uint64, prev data.MetaHeaderHandler) uint64 {
	return currentRound - prev.GetRound()
}
