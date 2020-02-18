package economics

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type endOfEpoch struct {
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	store             dataRetriever.StorageService
	dataPool          dataRetriever.PoolsHolder
	blockTracker      process.BlockTracker
	shardCoordinator  sharding.Coordinator
	epochStartTrigger process.EpochStartTriggerHandler
}

// ComputeRewardsPerBlock calculates the rewards per block value for the current epoch
func (e *endOfEpoch) ComputeRewardsPerBlock(metaBlock *block.MetaBlock) error {
	if !e.epochStartTrigger.IsEpochStart() || e.epochStartTrigger.Epoch() < 0 {
		return process.ErrNotEpochStartBlock
	}

	/*noncePerShardPrevEpoch, prevEpochStart, err := e.startNoncePerShardFromPreviousEpochStart(e.epochStartTrigger.Epoch() - 1)
	if err != nil {
		return err
	}

	noncePerShardCurrEpoch, err := e.startNoncePerShardFromLastCrossNotarized(metaBlock.GetNonce())
	if err != nil {
		return err
	}

	roundsPassedInEpoch := metaBlock.GetRound() - prevEpochStart.GetRound()
	maxRoundsInEpoch := roundsPassedInEpoch * uint64(e.shardCoordinator.NumberOfShards())*/

	return nil
}

func (e *endOfEpoch) startNoncePerShardFromPreviousEpochStart(epoch uint32) (map[uint32]uint64, *block.MetaBlock, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = 0
	}
	mapShardIdNonce[sharding.MetachainShardId] = 0

	if epoch == 0 {
		return mapShardIdNonce, nil, nil
	}

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	previousEpochStartMeta, err := process.GetMetaHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	mapShardIdNonce[sharding.MetachainShardId] = previousEpochStartMeta.GetNonce()
	for _, shardData := range previousEpochStartMeta.EpochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardId] = shardData.Nonce
	}

	return mapShardIdNonce, previousEpochStartMeta, nil
}

func (e *endOfEpoch) startNoncePerShardFromLastCrossNotarized(metaNonce uint64) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = 0
	}
	mapShardIdNonce[sharding.MetachainShardId] = metaNonce

	for shardID := uint32(0); shardID < e.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeaderForShard, _, err := e.blockTracker.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return nil, err
		}

		mapShardIdNonce[shardID] = lastCrossNotarizedHeaderForShard.GetNonce()
	}

	return mapShardIdNonce, nil
}

// GetRewardsPerBlock returns the calculated rewards per block value
func (e *endOfEpoch) GetRewardsPerBlock() *big.Int {
	return big.NewInt(0)
}

// VerifyRewardsPerBlock checks whether rewards per block value was correctly computed
func (e *endOfEpoch) VerifyRewardsPerBlock(metaBlock *block.MetaBlock) error {
	panic("implement me")
}
