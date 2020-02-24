package metachain

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const numberOfDaysInYear = 365.0
const numberOfSecondsInDay = 86400

type economics struct {
	marshalizer      marshal.Marshalizer
	store            dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	nodesCoordinator sharding.NodesCoordinator
	rewardsHandler   process.RewardsHandler
	roundTime        process.RoundTimeDurationHandler
}

// ArgsNewEpochEconomics
type ArgsNewEpochEconomics struct {
	Marshalizer      marshal.Marshalizer
	Store            dataRetriever.StorageService
	ShardCoordinator sharding.Coordinator
	NodesCoordinator sharding.NodesCoordinator
	RewardsHandler   process.RewardsHandler
	RoundTime        process.RoundTimeDurationHandler
}

// NewEndOfEpochEconomicsDataCreator creates a new end of epoch economics data creator object
func NewEndOfEpochEconomicsDataCreator(args ArgsNewEpochEconomics) (*economics, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilStore
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(args.RewardsHandler) {
		return nil, process.ErrNilRewardsHandler
	}
	if check.IfNil(args.RoundTime) {
		return nil, process.ErrNilRounder
	}

	e := &economics{
		marshalizer:      args.Marshalizer,
		store:            args.Store,
		shardCoordinator: args.ShardCoordinator,
		nodesCoordinator: args.NodesCoordinator,
		rewardsHandler:   args.RewardsHandler,
		roundTime:        args.RoundTime,
	}
	return e, nil
}

// ComputeRewardsPerBlock calculates the rewards per block value for the current epoch
func (e *economics) ComputeEndOfEpochEconomics(
	metaBlock *block.MetaBlock,
) (*block.Economics, error) {
	if check.IfNil(metaBlock) {
		return nil, process.ErrNilHeaderHandler
	}
	if metaBlock.AccumulatedFeesInEpoch == nil {
		return nil, process.ErrNilTotalAccumulatedFeesInEpoch
	}
	if !metaBlock.IsStartOfEpochBlock() || metaBlock.Epoch < 1 {
		return nil, process.ErrNotEpochStartBlock
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
	maxBlocksInEpoch := roundsPassedInEpoch * uint64(e.shardCoordinator.NumberOfShards()+1)
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
		rwdPerBlock = big.NewInt(0).Div(totalRewardsToBeDistributed, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))
	}

	computedEconomics := block.Economics{
		TotalSupply:            big.NewInt(0).Add(prevEpochEconomics.TotalSupply, newTokens),
		TotalToDistribute:      big.NewInt(0).Set(totalRewardsToBeDistributed),
		TotalNewlyMinted:       big.NewInt(0).Set(newTokens),
		RewardsPerBlockPerNode: e.computeRewardsPerValidatorPerBlock(rwdPerBlock),
		// TODO: get actual nodePrice from auction smart contract (currently on another feature branch, and not all features enabled)
		NodePrice: big.NewInt(0).Set(prevEpochEconomics.NodePrice),
	}

	return &computedEconomics, nil
}

// compute rewards per node per block
func (e *economics) computeRewardsPerValidatorPerBlock(rwdPerBlock *big.Int) *big.Int {
	numOfNodes := e.nodesCoordinator.GetNumTotalEligible()
	return big.NewInt(0).Div(rwdPerBlock, big.NewInt(0).SetUint64(numOfNodes))
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
	maxBlocksInADay := roundsPerDay * uint64(e.shardCoordinator.NumberOfShards()+1)

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
	totalNumBlocks += mapEndNonce[sharding.MetachainShardId] - mapStartNonce[sharding.MetachainShardId]

	return totalNumBlocks
}

func (e *economics) startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, *block.MetaBlock, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = 0
	}
	mapShardIdNonce[sharding.MetachainShardId] = 0

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	previousEpochStartMeta, err := process.GetMetaHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	if epoch == 0 {
		return mapShardIdNonce, previousEpochStartMeta, nil
	}

	mapShardIdNonce[sharding.MetachainShardId] = previousEpochStartMeta.GetNonce()
	for _, shardData := range previousEpochStartMeta.EpochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardId] = shardData.Nonce
	}

	return mapShardIdNonce, previousEpochStartMeta, nil
}

func (e *economics) startNoncePerShardFromLastCrossNotarized(metaNonce uint64, epochStart block.EpochStart) (map[uint32]uint64, error) {
	mapShardIdNonce := make(map[uint32]uint64, e.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < e.shardCoordinator.NumberOfShards(); i++ {
		mapShardIdNonce[i] = 0
	}
	mapShardIdNonce[sharding.MetachainShardId] = metaNonce

	for _, shardData := range epochStart.LastFinalizedHeaders {
		mapShardIdNonce[shardData.ShardId] = shardData.Nonce
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

	receivedEconomics := metaBlock.EpochStart.Economics
	if computedEconomics.TotalToDistribute.Cmp(receivedEconomics.TotalToDistribute) != 0 {
		return fmt.Errorf("%w total to distribute computed %d received %d",
			process.ErrEndOfEpochEconomicsDataDoesNotMatch, computedEconomics.TotalToDistribute, receivedEconomics.TotalToDistribute)
	}
	if computedEconomics.TotalNewlyMinted.Cmp(receivedEconomics.TotalNewlyMinted) != 0 {
		return fmt.Errorf("%w total newly minted computed %d received %d",
			process.ErrEndOfEpochEconomicsDataDoesNotMatch, computedEconomics.TotalNewlyMinted, receivedEconomics.TotalNewlyMinted)
	}
	if computedEconomics.TotalSupply.Cmp(receivedEconomics.TotalSupply) != 0 {
		return fmt.Errorf("%w total supply computed %d received %d",
			process.ErrEndOfEpochEconomicsDataDoesNotMatch, computedEconomics.TotalSupply, receivedEconomics.TotalSupply)
	}
	if computedEconomics.RewardsPerBlockPerNode.Cmp(receivedEconomics.RewardsPerBlockPerNode) != 0 {
		return fmt.Errorf("%wrewards per block per node computed %d received %d",
			process.ErrEndOfEpochEconomicsDataDoesNotMatch, computedEconomics.RewardsPerBlockPerNode, receivedEconomics.RewardsPerBlockPerNode)
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (e *economics) IsInterfaceNil() bool {
	return e == nil
}
