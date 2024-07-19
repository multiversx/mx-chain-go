package metachain

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignEconomics struct {
	*economics
}

func NewSovereignEconomics(ec *economics) (*sovereignEconomics, error) {
	if check.IfNil(ec) {
		return nil, process.ErrNilEconomicsData
	}

	ec.baseEconomicsHandler = &sovereignBaseEconomics{
		marshalizer:           ec.marshalizer,
		store:                 ec.store,
		shardCoordinator:      ec.shardCoordinator,
		economicsDataNotified: ec.economicsDataNotified,
		genesisEpoch:          ec.genesisEpoch,
		genesisNonce:          ec.genesisNonce,
	}

	return &sovereignEconomics{
		ec,
	}, nil
}

// todo: interfaceIsNil

/*
// ComputeEndOfEpochEconomics calculates the rewards per block value for the current epoch
func (e *sovereignEconomics) ComputeEndOfEpochEconomics(
	metaBlock data.MetaHeaderHandler,
) (*block.Economics, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	if metaBlock.GetAccumulatedFeesInEpoch() == nil {
		return nil, epochStart.ErrNilTotalAccumulatedFeesInEpoch
	}
	if metaBlock.GetDevFeesInEpoch() == nil {
		return nil, epochStart.ErrNilTotalDevFeesInEpoch
	}
	if !metaBlock.IsStartOfEpochBlock() || metaBlock.GetEpoch() < e.genesisEpoch+1 {
		return nil, epochStart.ErrNotEpochStartBlock
	}

	noncesPerShardPrevEpoch, prevEpochStart, err := e.startNoncePerShardFromEpochStart(metaBlock.GetEpoch() - 1)
	if err != nil {
		return nil, err
	}
	prevEpochEconomics := prevEpochStart.GetEpochStartHandler().GetEconomicsHandler()

	noncesPerShardCurrEpoch, err := e.startNoncePerShardFromLastCrossNotarized(metaBlock.GetNonce(), metaBlock.GetEpochStartHandler())
	if err != nil {
		return nil, err
	}

	roundsPassedInEpoch := metaBlock.GetRound() - prevEpochStart.GetRound()
	maxBlocksInEpoch := core.MaxUint64(1, roundsPassedInEpoch*uint64(e.shardCoordinator.NumberOfShards()))
	totalNumBlocksInEpoch := e.computeNumOfTotalCreatedBlocks(noncesPerShardPrevEpoch, noncesPerShardCurrEpoch)

	inflationRate := e.computeInflationRate(metaBlock.GetRound())
	rwdPerBlock := e.computeRewardsPerBlock(e.genesisTotalSupply, maxBlocksInEpoch, inflationRate, metaBlock.GetEpoch())
	totalRewardsToBeDistributed := big.NewInt(0).Mul(rwdPerBlock, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))

	newTokens := big.NewInt(0).Sub(totalRewardsToBeDistributed, metaBlock.GetAccumulatedFeesInEpoch())
	if newTokens.Cmp(big.NewInt(0)) < 0 {
		newTokens = big.NewInt(0)
		totalRewardsToBeDistributed = big.NewInt(0).Set(metaBlock.GetAccumulatedFeesInEpoch())
		rwdPerBlock.Div(totalRewardsToBeDistributed, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))
	}

	remainingToBeDistributed := big.NewInt(0).Sub(totalRewardsToBeDistributed, metaBlock.GetDevFeesInEpoch())
	e.adjustRewardsPerBlockWithDeveloperFees(rwdPerBlock, metaBlock.GetDevFeesInEpoch(), totalNumBlocksInEpoch)
	rewardsForLeaders := e.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, metaBlock.GetAccumulatedFeesInEpoch(), metaBlock.GetDevFeesInEpoch(), totalNumBlocksInEpoch, metaBlock.GetEpoch())
	remainingToBeDistributed = big.NewInt(0).Sub(remainingToBeDistributed, rewardsForLeaders)
	rewardsForProtocolSustainability := e.computeRewardsForProtocolSustainability(totalRewardsToBeDistributed, metaBlock.GetEpoch())
	remainingToBeDistributed = big.NewInt(0).Sub(remainingToBeDistributed, rewardsForProtocolSustainability)
	// adjust rewards per block taking into consideration protocol sustainability rewards
	e.adjustRewardsPerBlockWithProtocolSustainabilityRewards(rwdPerBlock, rewardsForProtocolSustainability, totalNumBlocksInEpoch)

	if big.NewInt(0).Cmp(totalRewardsToBeDistributed) > 0 {
		totalRewardsToBeDistributed = big.NewInt(0)
		remainingToBeDistributed = big.NewInt(0)
	}

	e.economicsDataNotified.SetLeadersFees(rewardsForLeaders)
	e.economicsDataNotified.SetRewardsToBeDistributed(totalRewardsToBeDistributed)
	e.economicsDataNotified.SetRewardsToBeDistributedForBlocks(remainingToBeDistributed)

	prevEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, prevEpochStart)
	if err != nil {
		return nil, err
	}

	computedEconomics := block.Economics{
		TotalSupply:                      big.NewInt(0).Add(prevEpochEconomics.GetTotalSupply(), newTokens),
		TotalToDistribute:                big.NewInt(0).Set(totalRewardsToBeDistributed),
		TotalNewlyMinted:                 big.NewInt(0).Set(newTokens),
		RewardsPerBlock:                  rwdPerBlock,
		RewardsForProtocolSustainability: rewardsForProtocolSustainability,
		NodePrice:                        big.NewInt(0).Set(prevEpochEconomics.GetNodePrice()),
		PrevEpochStartRound:              prevEpochStart.GetRound(),
		PrevEpochStartHash:               prevEpochStartHash,
	}

	e.printEconomicsData(
		metaBlock,
		prevEpochEconomics,
		inflationRate,
		newTokens,
		computedEconomics,
		totalRewardsToBeDistributed,
		totalNumBlocksInEpoch,
		rwdPerBlock,
		rewardsForProtocolSustainability,
	)

	maxPossibleNotarizedBlocks := e.maxPossibleNotarizedBlocks(metaBlock.GetRound(), prevEpochStart)
	err = e.checkEconomicsInvariants(computedEconomics, inflationRate, maxBlocksInEpoch, totalNumBlocksInEpoch, metaBlock, metaBlock.GetEpoch(), maxPossibleNotarizedBlocks)
	if err != nil {
		log.Warn("ComputeEndOfEpochEconomics", "error", err.Error())

		return nil, err
	}

	return &computedEconomics, nil
}

func (e *sovereignEconomics) startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, data.MetaHeaderHandler, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.TotalNumberOfShards())
	mapShardIdNonce[core.SovereignChainShardId] = e.genesisNonce

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	previousEpochStartMeta, err := process.GetSovereignChainHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	ret, castOk := previousEpochStartMeta.(data.MetaHeaderHandler)
	if !castOk {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	if epoch == e.genesisEpoch {
		return mapShardIdNonce, ret, nil
	}

	mapShardIdNonce[core.SovereignChainShardId] = ret.GetNonce()
	return mapShardIdNonce, ret, nil
}

func (e *sovereignEconomics) startNoncePerShardFromLastCrossNotarized(metaNonce uint64, _ data.EpochStartHandler) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards())
	mapShardIdNonce[core.SovereignChainShardId] = metaNonce
	return mapShardIdNonce, nil
}

func (e *sovereignEconomics) computeNumOfTotalCreatedBlocks(
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

func (e *sovereignEconomics) maxPossibleNotarizedBlocks(currentRound uint64, prev data.MetaHeaderHandler) uint64 {
	return currentRound - prev.GetRound()
}
*/
