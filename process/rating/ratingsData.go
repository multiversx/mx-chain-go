package rating

import (
	"fmt"
	"math"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.RatingsInfoHandler = (*RatingsData)(nil)

const millisecondsInHour = 3600 * 1000

type computeRatingStepArg struct {
	shardSize                       uint32
	consensusSize                   uint32
	roundTimeMillis                 uint64
	startRating                     uint32
	maxRating                       uint32
	hoursToMaxRatingFromStartRating uint32
	proposerDecreaseFactor          float32
	validatorDecreaseFactor         float32
	consecutiveMissedBlocksPenalty  float32
	proposerValidatorImportance     float32
}

// RatingsData will store information about ratingsComputation
type RatingsData struct {
	startRating                 uint32
	maxRating                   uint32
	minRating                   uint32
	signedBlocksThreshold       float32
	metaRatingsStepData         process.RatingsStepHandler
	shardRatingsStepData        process.RatingsStepHandler
	selectionChances            []process.SelectionChance
	nodesSetup                  process.NodesSetupHandler
	ratingsSetup                config.RatingsConfig
	roundDurationInMilliseconds uint64
	mutConfiguration            sync.RWMutex
}

// RatingsDataArg contains information for the creation of the new ratingsData
type RatingsDataArg struct {
	EpochNotifier             process.EpochNotifier
	Config                    config.RatingsConfig
	NodesSetupHandler         process.NodesSetupHandler
	RoundDurationMilliseconds uint64
}

// NewRatingsData creates a new RatingsData instance
func NewRatingsData(args RatingsDataArg) (*RatingsData, error) {
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochHandler
	}
	if check.IfNil(args.NodesSetupHandler) {
		return nil, process.ErrNilNodesSetupHandler
	}
	ratingsConfig := args.Config
	err := verifyRatingsConfig(ratingsConfig)
	if err != nil {
		return nil, err
	}

	chances := make([]process.SelectionChance, 0)
	for _, chance := range ratingsConfig.General.SelectionChances {
		chances = append(chances, &SelectionChance{
			MaxThreshold:  chance.MaxThreshold,
			ChancePercent: chance.ChancePercent,
		})
	}

	arg := computeRatingStepArg{
		shardSize:                       args.NodesSetupHandler.MinNumberOfShardNodes(),
		consensusSize:                   args.NodesSetupHandler.GetShardConsensusGroupSize(),
		roundTimeMillis:                 args.RoundDurationMilliseconds,
		startRating:                     ratingsConfig.General.StartRating,
		maxRating:                       ratingsConfig.General.MaxRating,
		hoursToMaxRatingFromStartRating: ratingsConfig.ShardChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          ratingsConfig.ShardChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         ratingsConfig.ShardChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  ratingsConfig.ShardChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     ratingsConfig.ShardChain.ProposerValidatorImportance,
	}
	shardRatingStep, err := computeRatingStep(arg)
	if err != nil {
		return nil, err
	}

	arg = computeRatingStepArg{
		shardSize:                       args.NodesSetupHandler.MinNumberOfMetaNodes(),
		consensusSize:                   args.NodesSetupHandler.GetMetaConsensusGroupSize(),
		roundTimeMillis:                 args.RoundDurationMilliseconds,
		startRating:                     ratingsConfig.General.StartRating,
		maxRating:                       ratingsConfig.General.MaxRating,
		hoursToMaxRatingFromStartRating: ratingsConfig.MetaChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          ratingsConfig.MetaChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         ratingsConfig.MetaChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     ratingsConfig.MetaChain.ProposerValidatorImportance,
	}
	metaRatingStep, err := computeRatingStep(arg)
	if err != nil {
		return nil, err
	}

	return &RatingsData{
		startRating:                 ratingsConfig.General.StartRating,
		maxRating:                   ratingsConfig.General.MaxRating,
		minRating:                   ratingsConfig.General.MinRating,
		signedBlocksThreshold:       ratingsConfig.General.SignedBlocksThreshold,
		metaRatingsStepData:         metaRatingStep,
		shardRatingsStepData:        shardRatingStep,
		selectionChances:            chances,
		nodesSetup:                  args.NodesSetupHandler,
		ratingsSetup:                ratingsConfig,
		roundDurationInMilliseconds: args.RoundDurationMilliseconds,
	}, nil
}

// EpochConfirmed will be called whenever a new epoch is called
func (rd *RatingsData) EpochConfirmed(epoch uint32, _ uint64) {
	log.Debug("RatingsData - epoch confirmed", "epoch", epoch)

	rd.mutConfiguration.Lock()
	defer rd.mutConfiguration.Unlock()

	shardRatingsStepData, err := computeRatingStep(computeRatingStepArg{
		shardSize:                       rd.nodesSetup.MinNumberOfShardNodes(),
		consensusSize:                   rd.nodesSetup.GetShardConsensusGroupSize(),
		roundTimeMillis:                 rd.roundDurationInMilliseconds,
		startRating:                     rd.ratingsSetup.General.StartRating,
		maxRating:                       rd.ratingsSetup.General.MaxRating,
		hoursToMaxRatingFromStartRating: rd.ratingsSetup.ShardChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          rd.ratingsSetup.ShardChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         rd.ratingsSetup.ShardChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  rd.ratingsSetup.ShardChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     rd.ratingsSetup.ShardChain.ProposerValidatorImportance,
	})
	if err != nil {
		log.Error("could not update shard ratings step", "error", err)
	}
	rd.shardRatingsStepData = shardRatingsStepData

	metaRatingsStepData, err := computeRatingStep(computeRatingStepArg{
		shardSize:                       rd.nodesSetup.MinNumberOfMetaNodes(),
		consensusSize:                   rd.nodesSetup.GetMetaConsensusGroupSize(),
		roundTimeMillis:                 rd.roundDurationInMilliseconds,
		startRating:                     rd.ratingsSetup.General.StartRating,
		maxRating:                       rd.ratingsSetup.General.MaxRating,
		hoursToMaxRatingFromStartRating: rd.ratingsSetup.MetaChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          rd.ratingsSetup.MetaChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         rd.ratingsSetup.MetaChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  rd.ratingsSetup.MetaChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     rd.ratingsSetup.MetaChain.ProposerValidatorImportance,
	})
	if err != nil {
		log.Error("could not update metachain ratings step", "error", err)
	}
	rd.metaRatingsStepData = metaRatingsStepData

	log.Debug("ratings data - epoch confirmed",
		"epoch", epoch,
		"shard min nodes", rd.nodesSetup.MinNumberOfShardNodes(),
		"shard consensus size", rd.nodesSetup.GetShardConsensusGroupSize(),
		"meta min nodes", rd.nodesSetup.MinNumberOfMetaNodes(),
		"metachain consensus size", rd.nodesSetup.GetMetaConsensusGroupSize(),
	)
}

func verifyRatingsConfig(settings config.RatingsConfig) error {
	if settings.General.MinRating < 1 {
		return process.ErrMinRatingSmallerThanOne
	}
	if settings.General.MinRating > settings.General.MaxRating {
		return fmt.Errorf("%w: minRating: %v, maxRating: %v",
			process.ErrMaxRatingIsSmallerThanMinRating,
			settings.General.MinRating,
			settings.General.MaxRating)
	}
	if settings.General.MaxRating < settings.General.StartRating || settings.General.MinRating > settings.General.StartRating {
		return fmt.Errorf("%w: minRating: %v, startRating: %v, maxRating: %v",
			process.ErrStartRatingNotBetweenMinAndMax,
			settings.General.MinRating,
			settings.General.StartRating,
			settings.General.MaxRating)
	}
	if settings.General.SignedBlocksThreshold > 1 || settings.General.SignedBlocksThreshold < 0 {
		return fmt.Errorf("%w signedBlocksThreshold: %v",
			process.ErrSignedBlocksThresholdNotBetweenZeroAndOne,
			settings.General.SignedBlocksThreshold)
	}
	if settings.ShardChain.HoursToMaxRatingFromStartRating == 0 {
		return fmt.Errorf("%w hoursToMaxRatingFromStartRating: shardChain",
			process.ErrHoursToMaxRatingFromStartRatingZero)
	}
	if settings.MetaChain.HoursToMaxRatingFromStartRating == 0 {
		return fmt.Errorf("%w hoursToMaxRatingFromStartRating: metachain",
			process.ErrHoursToMaxRatingFromStartRatingZero)
	}
	if settings.MetaChain.ConsecutiveMissedBlocksPenalty < 1 {
		return fmt.Errorf("%w: metaChain consecutiveMissedBlocksPenalty: %v",
			process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne,
			settings.MetaChain.ConsecutiveMissedBlocksPenalty)
	}
	if settings.ShardChain.ConsecutiveMissedBlocksPenalty < 1 {
		return fmt.Errorf("%w: shardChain consecutiveMissedBlocksPenalty: %v",
			process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne,
			settings.ShardChain.ConsecutiveMissedBlocksPenalty)
	}
	if settings.ShardChain.ProposerDecreaseFactor > -1 || settings.ShardChain.ValidatorDecreaseFactor > -1 {
		return fmt.Errorf("%w: shardChain decrease steps - proposer: %v, validator: %v",
			process.ErrDecreaseRatingsStepMoreThanMinusOne,
			settings.ShardChain.ProposerDecreaseFactor,
			settings.ShardChain.ValidatorDecreaseFactor)
	}
	if settings.MetaChain.ProposerDecreaseFactor > -1 || settings.MetaChain.ValidatorDecreaseFactor > -1 {
		return fmt.Errorf("%w: metachain decrease steps - proposer: %v, validator: %v",
			process.ErrDecreaseRatingsStepMoreThanMinusOne,
			settings.MetaChain.ProposerDecreaseFactor,
			settings.MetaChain.ValidatorDecreaseFactor)
	}
	return nil
}

func computeRatingStep(
	arg computeRatingStepArg,
) (process.RatingsStepHandler, error) {
	blocksProducedInHours := uint64(arg.hoursToMaxRatingFromStartRating*millisecondsInHour) / arg.roundTimeMillis
	ratingDifference := arg.maxRating - arg.startRating

	proposerProbability := float32(blocksProducedInHours) / float32(arg.shardSize)
	validatorProbability := proposerProbability * float32(arg.consensusSize)

	totalImportance := arg.proposerValidatorImportance + 1

	ratingFromProposer := float32(ratingDifference) * (arg.proposerValidatorImportance / totalImportance)
	ratingFromValidator := float32(ratingDifference) * (1 / totalImportance)

	proposerIncrease := ratingFromProposer / proposerProbability
	validatorIncrease := ratingFromValidator / validatorProbability
	proposerDecrease := proposerIncrease * arg.proposerDecreaseFactor
	validatorDecrease := validatorIncrease * arg.validatorDecreaseFactor

	if proposerIncrease > math.MaxInt32 {
		return nil, fmt.Errorf("%w proposerIncrease overflowed %v", process.ErrOverflow, proposerIncrease)
	}
	if validatorIncrease > math.MaxInt32 {
		return nil, fmt.Errorf("%w validatorIncrease overflowed %v", process.ErrOverflow, validatorIncrease)
	}
	if proposerDecrease < math.MinInt32 {
		return nil, fmt.Errorf("%w proposerDecrease overflowed %v", process.ErrOverflow, proposerDecrease)
	}
	if validatorDecrease < math.MinInt32 {
		return nil, fmt.Errorf("%w validatorDecrease overflowed %v", process.ErrOverflow, validatorDecrease)
	}
	if int32(proposerIncrease) < 1 {
		return nil, fmt.Errorf("%w proposerIncrease zero: %v", process.ErrIncreaseStepLowerThanOne, proposerIncrease)
	}
	if int32(validatorIncrease) < 1 {
		return nil, fmt.Errorf("%w validatorIncrease zero: %v", process.ErrIncreaseStepLowerThanOne, validatorIncrease)
	}

	return &RatingStep{
		proposerIncreaseRatingStep:     int32(proposerIncrease),
		proposerDecreaseRatingStep:     int32(proposerDecrease),
		validatorIncreaseRatingStep:    int32(validatorIncrease),
		validatorDecreaseRatingStep:    int32(validatorDecrease),
		consecutiveMissedBlocksPenalty: arg.consecutiveMissedBlocksPenalty}, nil
}

// StartRating will return the start rating
func (rd *RatingsData) StartRating() uint32 {
	// no need for mutex protection since this value is only set on constructor
	return rd.startRating
}

// MaxRating will return the max rating
func (rd *RatingsData) MaxRating() uint32 {
	// no need for mutex protection since this value is only set on constructor
	return rd.maxRating
}

// MinRating will return the min rating
func (rd *RatingsData) MinRating() uint32 {
	// no need for mutex protection since this value is only set on constructor
	return rd.minRating
}

// SignedBlocksThreshold will return the signed blocks threshold
func (rd *RatingsData) SignedBlocksThreshold() float32 {
	// no need for mutex protection since this value is only set on constructor
	return rd.signedBlocksThreshold
}

// SelectionChances will return the array of selectionChances and thresholds
func (rd *RatingsData) SelectionChances() []process.SelectionChance {
	// no need for mutex protection since this value is only set on constructor
	return rd.selectionChances
}

// MetaChainRatingsStepHandler returns the RatingsStepHandler used for the Metachain
func (rd *RatingsData) MetaChainRatingsStepHandler() process.RatingsStepHandler {
	rd.mutConfiguration.RLock()
	defer rd.mutConfiguration.RUnlock()

	return rd.metaRatingsStepData
}

// ShardChainRatingsStepHandler returns the RatingsStepHandler used for the ShardChains
func (rd *RatingsData) ShardChainRatingsStepHandler() process.RatingsStepHandler {
	rd.mutConfiguration.RLock()
	defer rd.mutConfiguration.RUnlock()

	return rd.shardRatingsStepData
}

// IsInterfaceNil returns true if underlying object is nil
func (rd *RatingsData) IsInterfaceNil() bool {
	return rd == nil
}
