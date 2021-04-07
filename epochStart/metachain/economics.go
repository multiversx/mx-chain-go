package metachain

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
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
	marshalizer           marshal.Marshalizer
	hasher                hashing.Hasher
	store                 dataRetriever.StorageService
	shardCoordinator      sharding.Coordinator
	rewardsHandler        process.RewardsHandler
	roundTime             process.RoundTimeDurationHandler
	genesisEpoch          uint32
	genesisNonce          uint64
	genesisTotalSupply    *big.Int
	economicsDataNotified epochStart.EpochEconomicsDataProvider
	stakingV2EnableEpoch  uint32
}

// ArgsNewEpochEconomics is the argument for the economics constructor
type ArgsNewEpochEconomics struct {
	Marshalizer           marshal.Marshalizer
	Hasher                hashing.Hasher
	Store                 dataRetriever.StorageService
	ShardCoordinator      sharding.Coordinator
	RewardsHandler        process.RewardsHandler
	RoundTime             process.RoundTimeDurationHandler
	GenesisEpoch          uint32
	GenesisNonce          uint64
	GenesisTotalSupply    *big.Int
	EconomicsDataNotified epochStart.EpochEconomicsDataProvider
	StakingV2EnableEpoch  uint32
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
	if check.IfNil(args.EconomicsDataNotified) {
		return nil, epochStart.ErrNilEconomicsDataProvider
	}
	if args.GenesisTotalSupply == nil {
		return nil, epochStart.ErrNilGenesisTotalSupply
	}

	e := &economics{
		marshalizer:           args.Marshalizer,
		hasher:                args.Hasher,
		store:                 args.Store,
		shardCoordinator:      args.ShardCoordinator,
		rewardsHandler:        args.RewardsHandler,
		roundTime:             args.RoundTime,
		genesisEpoch:          args.GenesisEpoch,
		genesisNonce:          args.GenesisNonce,
		genesisTotalSupply:    big.NewInt(0).Set(args.GenesisTotalSupply),
		economicsDataNotified: args.EconomicsDataNotified,
		stakingV2EnableEpoch:  args.StakingV2EnableEpoch,
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

	inflationRate := e.computeInflationRate(metaBlock.GetRound())
	rwdPerBlock := e.computeRewardsPerBlock(e.genesisTotalSupply, maxBlocksInEpoch, inflationRate, metaBlock.Epoch)
	totalRewardsToBeDistributed := big.NewInt(0).Mul(rwdPerBlock, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))

	newTokens := big.NewInt(0).Sub(totalRewardsToBeDistributed, metaBlock.AccumulatedFeesInEpoch)
	if newTokens.Cmp(big.NewInt(0)) < 0 {
		newTokens = big.NewInt(0)
		totalRewardsToBeDistributed = big.NewInt(0).Set(metaBlock.AccumulatedFeesInEpoch)
		rwdPerBlock.Div(totalRewardsToBeDistributed, big.NewInt(0).SetUint64(totalNumBlocksInEpoch))
	}

	remainingToBeDistributed := big.NewInt(0).Sub(totalRewardsToBeDistributed, metaBlock.DevFeesInEpoch)
	e.adjustRewardsPerBlockWithDeveloperFees(rwdPerBlock, metaBlock.DevFeesInEpoch, totalNumBlocksInEpoch)
	rewardsForLeaders := e.adjustRewardsPerBlockWithLeaderPercentage(rwdPerBlock, metaBlock.AccumulatedFeesInEpoch, metaBlock.DevFeesInEpoch, totalNumBlocksInEpoch, metaBlock.Epoch)
	remainingToBeDistributed = big.NewInt(0).Sub(remainingToBeDistributed, rewardsForLeaders)
	rewardsForProtocolSustainability := e.computeRewardsForProtocolSustainability(totalRewardsToBeDistributed, metaBlock.Epoch)
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
		TotalSupply:                      big.NewInt(0).Add(prevEpochEconomics.TotalSupply, newTokens),
		TotalToDistribute:                big.NewInt(0).Set(totalRewardsToBeDistributed),
		TotalNewlyMinted:                 big.NewInt(0).Set(newTokens),
		RewardsPerBlock:                  rwdPerBlock,
		RewardsForProtocolSustainability: rewardsForProtocolSustainability,
		NodePrice:                        big.NewInt(0).Set(prevEpochEconomics.NodePrice),
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

	maxPossibleNotarizedBlocks := e.maxPossibleNotarizedBlocks(metaBlock.Round, prevEpochStart)
	err = e.checkEconomicsInvariants(computedEconomics, inflationRate, maxBlocksInEpoch, totalNumBlocksInEpoch, metaBlock, metaBlock.Epoch, maxPossibleNotarizedBlocks)
	if err != nil {
		log.Warn("ComputeEndOfEpochEconomics", "error", err.Error())

		return nil, err
	}

	return &computedEconomics, nil
}

func (e *economics) printEconomicsData(
	metaBlock *block.MetaBlock,
	prevEpochEconomics block.Economics,
	inflationRate float64,
	newTokens *big.Int,
	computedEconomics block.Economics,
	totalRewardsToBeDistributed *big.Int,
	totalNumBlocksInEpoch uint64,
	rwdPerBlock *big.Int,
	rewardsForProtocolSustainability *big.Int,
) {
	header := []string{"identifier", "", "value"}

	var rewardsForLeaders *big.Int
	if metaBlock.Epoch > e.stakingV2EnableEpoch {
		rewardsForLeaders = core.GetIntTrimmedPercentageOfValue(metaBlock.AccumulatedFeesInEpoch, e.rewardsHandler.LeaderPercentage())
	} else {
		rewardsForLeaders = core.GetApproximatePercentageOfValue(metaBlock.AccumulatedFeesInEpoch, e.rewardsHandler.LeaderPercentage())
	}

	maxSupplyLength := len(prevEpochEconomics.TotalSupply.String())
	lines := []*display.LineData{
		e.newDisplayLine("epoch", "",
			e.alignRight(fmt.Sprintf("%d", metaBlock.Epoch), maxSupplyLength)),
		e.newDisplayLine("inflation rate", "",
			e.alignRight(fmt.Sprintf("%.6f", inflationRate), maxSupplyLength)),
		e.newDisplayLine("previous total supply", "(1)",
			e.alignRight(prevEpochEconomics.TotalSupply.String(), maxSupplyLength)),
		e.newDisplayLine("new tokens", "(2)",
			e.alignRight(newTokens.String(), maxSupplyLength)),
		e.newDisplayLine("current total supply", "(1+2)",
			e.alignRight(computedEconomics.TotalSupply.String(), maxSupplyLength)),
		e.newDisplayLine("accumulated fees in epoch", "(3)",
			e.alignRight(metaBlock.AccumulatedFeesInEpoch.String(), maxSupplyLength)),
		e.newDisplayLine("total rewards to be distributed", "(4)",
			e.alignRight(totalRewardsToBeDistributed.String(), maxSupplyLength)),
		e.newDisplayLine("total num blocks in epoch", "(5)",
			e.alignRight(fmt.Sprintf("%d", totalNumBlocksInEpoch), maxSupplyLength)),
		e.newDisplayLine("dev fees in epoch", "(6)",
			e.alignRight(metaBlock.DevFeesInEpoch.String(), maxSupplyLength)),
		e.newDisplayLine("leader fees in epoch", "(7)",
			e.alignRight(rewardsForLeaders.String(), maxSupplyLength)),
		e.newDisplayLine("reward per block", "(8)",
			e.alignRight(rwdPerBlock.String(), maxSupplyLength)),
		e.newDisplayLine("percent for protocol sustainability", "(9)",
			e.alignRight(fmt.Sprintf("%.6f", e.rewardsHandler.ProtocolSustainabilityPercentage()), maxSupplyLength)),
		e.newDisplayLine("reward for protocol sustainability", "(4 * 9)",
			e.alignRight(rewardsForProtocolSustainability.String(), maxSupplyLength)),
	}

	str, err := display.CreateTableString(header, lines)
	if err != nil {
		log.Error("economics.printEconomicsData", "error", err)
		return
	}

	log.Debug("computed economics data\n" + str)
}

func (e *economics) alignRight(val string, maxLen int) string {
	if len(val) >= maxLen {
		return val
	}

	return strings.Repeat(" ", maxLen-len(val)) + val
}

func (e *economics) newDisplayLine(values ...string) *display.LineData {
	return display.NewLineData(false, values)
}

// compute the rewards for protocol sustainability - percentage from total rewards
func (e *economics) computeRewardsForProtocolSustainability(totalRewards *big.Int, epoch uint32) *big.Int {
	if epoch > e.stakingV2EnableEpoch {
		return core.GetIntTrimmedPercentageOfValue(totalRewards, e.rewardsHandler.ProtocolSustainabilityPercentage())
	}

	return core.GetApproximatePercentageOfValue(totalRewards, e.rewardsHandler.ProtocolSustainabilityPercentage())
}

// adjustment for rewards given for each proposed block taking protocol sustainability rewards into consideration
func (e *economics) adjustRewardsPerBlockWithProtocolSustainabilityRewards(
	rwdPerBlock *big.Int,
	protocolSustainabilityRewards *big.Int,
	blocksInEpoch uint64,
) {
	protocolSustainabilityRewardsPerBlock := big.NewInt(0).Div(protocolSustainabilityRewards, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, protocolSustainabilityRewardsPerBlock)
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
	developerFees *big.Int,
	blocksInEpoch uint64,
	epoch uint32,
) *big.Int {
	accumulatedFeesForValidators := big.NewInt(0).Set(accumulatedFees)
	var rewardsForLeaders *big.Int
	if epoch > e.stakingV2EnableEpoch {
		accumulatedFeesForValidators.Sub(accumulatedFeesForValidators, developerFees)
		rewardsForLeaders = core.GetIntTrimmedPercentageOfValue(accumulatedFeesForValidators, e.rewardsHandler.LeaderPercentage())
	} else {
		rewardsForLeaders = core.GetApproximatePercentageOfValue(accumulatedFeesForValidators, e.rewardsHandler.LeaderPercentage())
	}

	averageLeaderRewardPerBlock := big.NewInt(0).Div(rewardsForLeaders, big.NewInt(0).SetUint64(blocksInEpoch))
	rwdPerBlock.Sub(rwdPerBlock, averageLeaderRewardPerBlock)

	return rewardsForLeaders
}

// compute inflation rate from genesisTotalSupply and economics settings for that year
func (e *economics) computeInflationRate(currentRound uint64) float64 {
	roundsPerDay := numberOfSecondsInDay / uint64(e.roundTime.TimeDuration().Seconds())
	roundsPerYear := numberOfDaysInYear * roundsPerDay
	yearsIndex := uint32(currentRound/roundsPerYear) + 1

	return e.rewardsHandler.MaxInflationRate(yearsIndex)
}

// compute rewards per block from according to inflation rate and total supply from previous block and maxBlocksPerEpoch
func (e *economics) computeRewardsPerBlock(
	prevTotalSupply *big.Int,
	maxBlocksInEpoch uint64,
	inflationRate float64,
	epoch uint32,
) *big.Int {

	inflationRateForEpoch := e.computeInflationForEpoch(inflationRate, maxBlocksInEpoch)

	rewardsPerBlock := big.NewInt(0).Div(prevTotalSupply, big.NewInt(0).SetUint64(maxBlocksInEpoch))
	if epoch > e.stakingV2EnableEpoch {
		return core.GetIntTrimmedPercentageOfValue(rewardsPerBlock, inflationRateForEpoch)
	}

	return core.GetApproximatePercentageOfValue(rewardsPerBlock, inflationRateForEpoch)
}

func (e *economics) computeInflationForEpoch(inflationRate float64, maxBlocksInEpoch uint64) float64 {
	inflationRatePerDay := inflationRate / numberOfDaysInYear
	roundsPerDay := numberOfSecondsInDay / uint64(e.roundTime.TimeDuration().Seconds())
	maxBlocksInADay := core.MaxUint64(1, roundsPerDay*uint64(e.shardCoordinator.NumberOfShards()+1))

	inflationRateForEpoch := inflationRatePerDay * (float64(maxBlocksInEpoch) / float64(maxBlocksInADay))

	return inflationRateForEpoch
}

func (e *economics) computeNumOfTotalCreatedBlocks(
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

func (e *economics) maxPossibleNotarizedBlocks(currentRound uint64, prev *block.MetaBlock) uint64 {
	maxBlocks := uint64(0)
	for _, shardData := range prev.EpochStart.LastFinalizedHeaders {
		maxBlocks += currentRound - shardData.Round
	}

	return maxBlocks
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

func (e *economics) checkEconomicsInvariants(
	computedEconomics block.Economics,
	inflationRate float64,
	maxBlocksInEpoch uint64,
	totalNumBlocksInEpoch uint64,
	metaBlock *block.MetaBlock,
	epoch uint32,
	maxPossibleNotarizedBlocks uint64,
) error {
	if epoch <= e.stakingV2EnableEpoch {
		return nil
	}

	maxAllowedInflation := e.rewardsHandler.MaxInflationRate(1)
	if !core.IsInRangeInclusiveFloat64(inflationRate, 0, maxAllowedInflation) {
		return fmt.Errorf("%w, computed inflation %s, max allowed %s",
			epochStart.ErrInvalidInflationRate,
			strconv.FormatFloat(inflationRate, 'f', -1, 64),
			strconv.FormatFloat(maxAllowedInflation, 'f', -1, 64))

	}

	if !core.IsInRangeInclusive(metaBlock.AccumulatedFeesInEpoch, zero, e.genesisTotalSupply) {
		return fmt.Errorf("%w, computed accumulated fees %s, max allowed %s",
			epochStart.ErrInvalidAccumulatedFees,
			metaBlock.AccumulatedFeesInEpoch,
			e.genesisTotalSupply,
		)
	}

	actualMaxBlocks := maxBlocksInEpoch
	if maxPossibleNotarizedBlocks > actualMaxBlocks{
		actualMaxBlocks = maxPossibleNotarizedBlocks
	}

	inflationPerEpoch := e.computeInflationForEpoch(inflationRate, actualMaxBlocks)
	maxRewardsInEpoch := core.GetIntTrimmedPercentageOfValue(computedEconomics.TotalSupply, inflationPerEpoch)
	if maxRewardsInEpoch.Cmp(metaBlock.AccumulatedFeesInEpoch) < 0 {
		maxRewardsInEpoch = metaBlock.AccumulatedFeesInEpoch
	}

	if !core.IsInRangeInclusive(computedEconomics.RewardsForProtocolSustainability, zero, maxRewardsInEpoch) {
		return fmt.Errorf("%w, computed protocol sustainability rewards %s, max allowed %s",
			epochStart.ErrInvalidEstimatedProtocolSustainabilityRewards,
			computedEconomics.RewardsForProtocolSustainability,
			maxRewardsInEpoch,
		)
	}
	if !core.IsInRangeInclusive(computedEconomics.TotalNewlyMinted, zero, maxRewardsInEpoch) {
		return fmt.Errorf("%w, computed minted tokens %s, max allowed %s",
			epochStart.ErrInvalidAmountMintedTokens,
			computedEconomics.TotalNewlyMinted,
			maxRewardsInEpoch,
		)
	}

	if !core.IsInRangeInclusive(computedEconomics.TotalToDistribute, zero, maxRewardsInEpoch) {
		return fmt.Errorf("%w, computed total to distribute %s, max allowed %s",
			epochStart.ErrInvalidTotalToDistribute,
			computedEconomics.TotalToDistribute,
			maxRewardsInEpoch,
		)
	}

	rewardsSum := big.NewInt(0).Mul(big.NewInt(int64(totalNumBlocksInEpoch)), computedEconomics.RewardsPerBlock)
	if !core.IsInRangeInclusive(rewardsSum, zero, maxRewardsInEpoch) {
		return fmt.Errorf("%w, computed sum of rewards %s, max allowed %s",
			epochStart.ErrInvalidRewardsPerBlock,
			rewardsSum,
			maxRewardsInEpoch,
		)
	}

	return nil
}

// VerifyRewardsPerBlock checks whether rewards per block value was correctly computed
func (e *economics) VerifyRewardsPerBlock(
	metaBlock *block.MetaBlock,
	correctedProtocolSustainability *big.Int,
	computedEconomics *block.Economics,
) error {
	if computedEconomics == nil {
		return epochStart.ErrNilEconomicsData
	}
	if !metaBlock.IsStartOfEpochBlock() {
		return nil
	}

	computedEconomics.RewardsForProtocolSustainability.Set(correctedProtocolSustainability)
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
		"computed rewards for protocol sustainability", computed.RewardsForProtocolSustainability,
		"computed node price", computed.NodePrice,
		"\nreceived total to distribute", received.TotalToDistribute,
		"received total newly minted", received.TotalNewlyMinted,
		"received total supply", received.TotalSupply,
		"received rewards per block per node", received.RewardsPerBlock,
		"received rewards for protocol sustainability", received.RewardsForProtocolSustainability,
		"received node price", received.NodePrice,
	)
}
