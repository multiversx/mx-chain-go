package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
)

type baseEconomics struct {
	marshalizer           marshal.Marshalizer
	store                 dataRetriever.StorageService
	shardCoordinator      ExtendedShardCoordinatorHandler
	economicsDataNotified epochStart.EpochEconomicsDataProvider
	genesisEpoch          uint32
	genesisNonce          uint64
}

func (e *baseEconomics) startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, data.MetaHeaderHandler, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = e.genesisNonce
	}
	mapShardIdNonce[core.MetachainShardId] = e.genesisNonce

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	previousEpochStartMeta, err := process.GetMetaHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	if epoch == e.genesisEpoch {
		return mapShardIdNonce, previousEpochStartMeta, nil
	}

	mapShardIdNonce[core.MetachainShardId] = previousEpochStartMeta.GetNonce()
	for _, shardData := range previousEpochStartMeta.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		mapShardIdNonce[shardData.GetShardID()] = shardData.GetNonce()
	}

	return mapShardIdNonce, previousEpochStartMeta, nil
}

func (e *baseEconomics) startNoncePerShardFromLastCrossNotarized(metaNonce uint64, epochStart data.EpochStartHandler) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = e.genesisNonce
	}
	mapShardIdNonce[core.MetachainShardId] = metaNonce

	for _, shardData := range epochStart.GetLastFinalizedHeaderHandlers() {
		mapShardIdNonce[shardData.GetShardID()] = shardData.GetNonce()
	}

	return mapShardIdNonce, nil
}

func (e *baseEconomics) computeNumOfTotalCreatedBlocks(
	mapStartNonce map[uint32]uint64,
	mapEndNonce map[uint32]uint64,
) uint64 {
	totalNumBlocks := uint64(0)
	var blocksInShard uint64
	blocksPerShard := make(map[uint32]uint64)
	shardMap := createShardsMap(e.shardCoordinator)
	for shardId := range shardMap {
		blocksInShard = mapEndNonce[shardId] - mapStartNonce[shardId]
		blocksPerShard[shardId] = blocksInShard
		totalNumBlocks += blocksInShard
		log.Debug("computeNumOfTotalCreatedBlocks",
			"shardID", shardId,
			"prevEpochLastNonce", mapEndNonce[shardId],
			"epochLastNonce", mapStartNonce[shardId],
			"nbBlocksEpoch", blocksPerShard[shardId],
		)
	}

	e.economicsDataNotified.SetNumberOfBlocks(totalNumBlocks)
	e.economicsDataNotified.SetNumberOfBlocksPerShard(blocksPerShard)

	return core.MaxUint64(1, totalNumBlocks)
}

func (e *baseEconomics) maxPossibleNotarizedBlocks(currentRound uint64, prev data.MetaHeaderHandler) uint64 {
	maxBlocks := uint64(0)
	for _, shardData := range prev.GetEpochStartHandler().GetLastFinalizedHeaderHandlers() {
		maxBlocks += currentRound - shardData.GetRound()
	}
	// For metaChain blocks
	maxBlocks += currentRound - prev.GetRound()

	return maxBlocks
}
