package metachain

import (
	"bytes"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.EndOfEpochEconomics = (*economics)(nil)

const numberOfDaysInYear = 365.0
const numberOfSecondsInDay = 86400

type economics struct {
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	store            dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	rewardsHandler   process.RewardsHandler
	roundTime        process.RoundTimeDurationHandler
	genesisEpoch     uint32
	genesisNonce     uint64
}

// ArgsNewEpochEconomics is the argument for the economics constructor
type ArgsNewEpochEconomics struct {
	Marshalizer      marshal.Marshalizer
	Hasher           hashing.Hasher
	Store            dataRetriever.StorageService
	ShardCoordinator sharding.Coordinator
	RewardsHandler   process.RewardsHandler
	RoundTime        process.RoundTimeDurationHandler
	GenesisEpoch     uint32
	GenesisNonce     uint64
}

// NewEndOfEpochEconomicsDataCreator creates a new end of epoch economics data creator object
func NewEndOfEpochEconomicsDataCreator(args ArgsNewEpochEconomics) (*economics, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, epochStart.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, epochStart.ErrNilShardCoordinator
	}
	if check.IfNil(args.RewardsHandler) {
		return nil, epochStart.ErrNilRewardsHandler
	}
	if check.IfNil(args.RoundTime) {
		return nil, process.ErrNilRounder
	}

	e := &economics{
		marshalizer:      args.Marshalizer,
		hasher:           args.Hasher,
		store:            args.Store,
		shardCoordinator: args.ShardCoordinator,
		rewardsHandler:   args.RewardsHandler,
		roundTime:        args.RoundTime,
		genesisEpoch:     args.GenesisEpoch,
		genesisNonce:     args.GenesisNonce,
	}
	return e, nil
}

// ComputeEndOfEpochEconomics calculates the rewards per block value for the current epoch
func (e *economics) ComputeEndOfEpochEconomics(
	metaBlock *block.MetaBlock,
) (*block.Economics, error) {
	if check.IfNil(metaBlock) {
		return nil, epochStart.ErrNilHeaderHandler
	}
	if metaBlock.AccumulatedFeesInEpoch == nil {
		return nil, epochStart.ErrNilTotalAccumulatedFeesInEpoch
	}
	if metaBlock.DevFeesInEpoch == nil {
		return nil, epochStart.ErrNilTotalDevFeesInEpoch
	}
	if !metaBlock.IsStartOfEpochBlock() || metaBlock.Epoch < e.genesisEpoch+1 {
		return nil, epochStart.ErrNotEpochStartBlock
	}

	noncesPerShardPrevEpoch, prevEpochStart, err := e.startNoncePerShardFromEpochStart(metaBlock.Epoch - 1)
	if err != nil {
		return nil, err
	}
	prevEpochEconomics := prevEpochStart.EpochStart.Economics

	noncesPerShardCurrEpoch, err := e.startNoncePerShardFromLastCrossNotarized(metaBlock.GetNonce(), metaBlock.EpochStart)
	if err != nil {
		return nil, err
	}

	roundsPassedInEpoch := metaBlock.GetRound() - prevEpochStart.GetRound()
	maxBlocksInEpoch := core.MaxUint64(1, roundsPassedInEpoch*uint64(e.shardCoordinator.NumberOfShards()+1))
	totalNumBlocksInEpoch := e.computeNumOfTotalCreatedBlocks(noncesPerShardPrevEpoch, noncesPerShardCurrEpoch)

	inflationRate, err := e.computeInflationRate(prevEpochEconomics.TotalSupply, prevEpochEconomics.NodePrice)
	if err != nil {
		return nil, err
	}

	rwdPerBlock := e.computeRewardsPerBlock(prevEpochEconomics.TotalSupply, maxBlocksInEpoch, inflationRate)
	totalRewardsToBeDistributed := big.NewInt(0).Mul(rwdPerBlock, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))

	newTokens := big.NewInt(0).Sub(totalRewardsToBeDistributed, metaBlock.AccumulatedFeesInEpoch)
	if newTokens.Cmp(big.NewInt(0)) < 0 {
		newTokens = big.NewInt(0)
		totalRewardsToBeDistributed = big.NewInt(0).Set(metaBlock.AccumulatedFeesInEpoch)
		rwdPerBlock.Div(totalRewardsToBeDistributed, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))
	}

	e.adjustRewardsPerBlockWithDeveloperFees(rwdPerBlock, metaBlock.DevFeesInEpoch, totalNumBlocksInEpoch)
	e.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, metaBlock.AccumulatedFeesInEpoch, totalNumBlocksInEpoch)
	rewardsForCommunity := e.computeRewardsForCommunity(totalRewardsToBeDistributed)
	// adjust rewards per block taking into consideration community rewards
	e.adjustRewardsPerBlockWithCommunityRewards(rwdPerBlock, rewardsForCommunity, totalNumBlocksInEpoch)

	prevEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, prevEpochStart)
	if err != nil {
		return nil, err
	}

	computedEconomics := block.Economics{
		TotalSupply:         big.NewInt(0).Add(prevEpochEconomics.TotalSupply, newTokens),
		TotalToDistribute:   big.NewInt(0).Set(totalRewardsToBeDistributed),
		TotalNewlyMinted:    big.NewInt(0).Set(newTokens),
		RewardsPerBlock:     rwdPerBlock,
		RewardsForCommunity: rewardsForCommunity,
		// TODO: get actual nodePrice from auction smart contract (currently on another feature branch, and not all features enabled)
		NodePrice:           big.NewInt(0).Set(prevEpochEconomics.NodePrice),
		PrevEpochStartRound: prevEpochStart.GetRound(),
		PrevEpochStartHash:  prevEpochStartHash,
	}

	return &computedEconomics, nil
}

// compute the rewards for community - percentage from total rewards
func (e *economics) computeRewardsForCommunity(totalRewards *big.Int) *big.Int {
	rewardsForCommunity := core.GetPercentageOfValue(totalRewards, e.rewardsHandler.CommunityPercentage())
	return rewardsForCommunity
}

// adjustment for rewards given for each proposed block taking community rewards into consideration
func (e *economics) adjustRewardsPerBlockWithCommunityRewards(
	rwdPerBlock *big.Int,
	communityRewards *big.Int,
	blocksInEpoch uint64,
) {
	communityRewardsPerBlock := big.NewInt(0).Div(communityRewards, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, communityRewardsPerBlock)
}

// adjustment for rewards given for each proposed block taking developer fees into consideration
func (e *economics) adjustRewardsPerBlockWithDeveloperFees(
	rwdPerBlock *big.Int,
	developerFees *big.Int,
	blocksInEpoch uint64,
) {
	developerFeesPerBlock := big.NewInt(0).Div(developerFees, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, developerFeesPerBlock)
}

func (e *economics) adjustRewardsPerBlockWithLeaderPercentage(
	rwdPerBlock *big.Int,
	accumulatedFees *big.Int,
	blocksInEpoch uint64,
) {
	rewardsForLeaders := core.GetPercentageOfValue(accumulatedFees, e.rewardsHandler.LeaderPercentage())
	averageLeaderRewardPerBlock := big.NewInt(0).Div(rewardsForLeaders, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, averageLeaderRewardPerBlock)
}

// compute inflation rate from totalSupply and totalStaked
func (e *economics) computeInflationRate(_ *big.Int, _ *big.Int) (float64, error) {
	//TODO: use prevTotalSupply and nodePrice (number of eligible + number of waiting)
	// for epoch which ends now to compute inflation rate according to formula provided by L.
	return e.rewardsHandler.MaxInflationRate(), nil
}

// compute rewards per block from according to inflation rate and total supply from previous block and maxBlocksPerEpoch
func (e *economics) computeRewardsPerBlock(
	prevTotalSupply *big.Int,
	maxBlocksInEpoch uint64,
	inflationRate float64,
) *big.Int {

	inflationRatePerDay := inflationRate / numberOfDaysInYear
	roundsPerDay := numberOfSecondsInDay / uint64(e.roundTime.TimeDuration().Seconds())
	maxBlocksInADay := core.MaxUint64(1, roundsPerDay*uint64(e.shardCoordinator.NumberOfShards()+1))

	inflationRateForEpoch := inflationRatePerDay * (float64(maxBlocksInEpoch) / float64(maxBlocksInADay))

	rewardsPerBlock := big.NewInt(0).Div(prevTotalSupply, big.NewInt(0).SetUint64(maxBlocksInEpoch))
	rewardsPerBlock = core.GetPercentageOfValue(rewardsPerBlock, inflationRateForEpoch)

	return rewardsPerBlock
}

func (e *economics) computeNumOfTotalCreatedBlocks(
	mapStartNonce map[uint32]uint64,
	mapEndNonce map[uint32]uint64,
) uint64 {
	totalNumBlocks := uint64(0)
	for shardId := uint32(0); shardId < e.shardCoordinator.NumberOfShards(); shardId++ {
		totalNumBlocks += mapEndNonce[shardId] - mapStartNonce[shardId]
	}
	totalNumBlocks += mapEndNonce[core.MetachainShardId] - mapStartNonce[core.MetachainShardId]

	return core.MaxUint64(1, totalNumBlocks)
}

func (e *economics) startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, *block.MetaBlock, error) {
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
	for _, shardData := range previousEpochStartMeta.EpochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardID] = shardData.Nonce
	}

	return mapShardIdNonce, previousEpochStartMeta, nil
}

func (e *economics) startNoncePerShardFromLastCrossNotarized(metaNonce uint64, epochStart block.EpochStart) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = e.genesisNonce
	}
	mapShardIdNonce[core.MetachainShardId] = metaNonce

	for _, shardData := range epochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardID] = shardData.Nonce
	}

	return mapShardIdNonce, nil
}

// VerifyRewardsPerBlock checks whether rewards per block value was correctly computed
func (e *economics) VerifyRewardsPerBlock(
	metaBlock *block.MetaBlock,
) error {
	if !metaBlock.IsStartOfEpochBlock() {
		return nil
	}
	computedEconomics, err := e.ComputeEndOfEpochEconomics(metaBlock)
	if err != nil {
		return err
	}
	computedEconomicsHash, err := core.CalculateHash(e.marshalizer, e.hasher, computedEconomics)
	if err != nil {
		return err
	}

	receivedEconomics := metaBlock.EpochStart.Economics
	receivedEconomicsHash, err := core.CalculateHash(e.marshalizer, e.hasher, &receivedEconomics)
	if err != nil {
		return err
	}

	if !bytes.Equal(receivedEconomicsHash, computedEconomicsHash) {
		logEconomicsDifferences(computedEconomics, &receivedEconomics)
		return epochStart.ErrEndOfEpochEconomicsDataDoesNotMatch
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *economics) IsInterfaceNil() bool {
	return e == nil
}

func logEconomicsDifferences(computed *block.Economics, received *block.Economics) {
	log.Warn("VerifyRewardsPerBlock error",
		"\ncomputed total to distribute", computed.TotalToDistribute,
		"computed total newly minted", computed.TotalNewlyMinted,
		"computed total supply", computed.TotalSupply,
		"computed rewards per block per node", computed.RewardsPerBlock,
		"computed rewards for community", computed.RewardsForCommunity,
		"computed node price", computed.NodePrice,
		"\nreceived total to distribute", received.TotalToDistribute,
		"received total newly minted", received.TotalNewlyMinted,
		"received total supply", received.TotalSupply,
		"received rewards per block per node", received.RewardsPerBlock,
		"received rewards for community", received.RewardsForCommunity,
		"received node price", received.NodePrice,
	)
}
