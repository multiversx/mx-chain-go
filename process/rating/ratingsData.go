package rating

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.RatingsInfoHandler = (*RatingsData)(nil)

const millisecondsInHour = 3600 * 1000

type ratingsStepsData struct {
	enableEpoch          uint32
	shardRatingsStepData process.RatingsStepHandler
	metaRatingsStepData  process.RatingsStepHandler
}

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
	currentRatingsStepData      ratingsStepsData
	ratingsStepsConfig          []ratingsStepsData
	selectionChances            []process.SelectionChance
	chainParametersHandler      process.ChainParametersHandler
	ratingsSetup                config.RatingsConfig
	roundDurationInMilliseconds uint64
	mutConfiguration            sync.RWMutex
}

// RatingsDataArg contains information for the creation of the new ratingsData
type RatingsDataArg struct {
	EpochNotifier             process.EpochNotifier
	Config                    config.RatingsConfig
	ChainParametersHolder     process.ChainParametersHandler
	RoundDurationMilliseconds uint64
}

// NewRatingsData creates a new RatingsData instance
func NewRatingsData(args RatingsDataArg) (*RatingsData, error) {
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(args.ChainParametersHolder) {
		return nil, process.ErrNilChainParametersHandler
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

	currentChainParameters := args.ChainParametersHolder.CurrentChainParameters()
	arg := computeRatingStepArg{
		shardSize:                       currentChainParameters.ShardMinNumNodes,
		consensusSize:                   currentChainParameters.ShardConsensusGroupSize,
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
		shardSize:                       currentChainParameters.MetachainMinNumNodes,
		consensusSize:                   currentChainParameters.MetachainConsensusGroupSize,
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

	ratingsConfigValue := ratingsStepsData{
		enableEpoch:          args.EpochNotifier.CurrentEpoch(),
		shardRatingsStepData: shardRatingStep,
		metaRatingsStepData:  metaRatingStep,
	}

	ratingData := &RatingsData{
		startRating:                 ratingsConfig.General.StartRating,
		maxRating:                   ratingsConfig.General.MaxRating,
		minRating:                   ratingsConfig.General.MinRating,
		signedBlocksThreshold:       ratingsConfig.General.SignedBlocksThreshold,
		currentRatingsStepData:      ratingsConfigValue,
		selectionChances:            chances,
		chainParametersHandler:      args.ChainParametersHolder,
		ratingsSetup:                ratingsConfig,
		roundDurationInMilliseconds: args.RoundDurationMilliseconds,
	}

	err = ratingData.computeRatingStepsConfig(args.ChainParametersHolder.AllChainParameters())
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(ratingData)

	return ratingData, nil
}

func (rd *RatingsData) computeRatingStepsConfig(chainParamsList []config.ChainParametersByEpochConfig) error {
	if len(chainParamsList) == 0 {
		return process.ErrEmptyChainParametersConfiguration
	}

	ratingsStepsConfig := make([]ratingsStepsData, 0)
	for _, chainParams := range chainParamsList {
		configForEpoch, err := rd.computeRatingStepsConfigForParams(chainParams)
		if err != nil {
			return err
		}

		ratingsStepsConfig = append(ratingsStepsConfig, configForEpoch)
	}

	// sort the config values descending
	sort.SliceStable(ratingsStepsConfig, func(i, j int) bool {
		return ratingsStepsConfig[i].enableEpoch > ratingsStepsConfig[j].enableEpoch
	})

	earliestConfig := ratingsStepsConfig[len(ratingsStepsConfig)-1]
	if earliestConfig.enableEpoch != 0 {
		return process.ErrMissingConfigurationForEpochZero
	}

	rd.ratingsStepsConfig = ratingsStepsConfig

	return nil
}

func (rd *RatingsData) computeRatingStepsConfigForParams(chainParams config.ChainParametersByEpochConfig) (ratingsStepsData, error) {
	shardRatingsStepsArgs := computeRatingStepArg{
		shardSize:                       chainParams.ShardMinNumNodes,
		consensusSize:                   chainParams.ShardConsensusGroupSize,
		roundTimeMillis:                 rd.roundDurationInMilliseconds,
		startRating:                     rd.ratingsSetup.General.StartRating,
		maxRating:                       rd.ratingsSetup.General.MaxRating,
		hoursToMaxRatingFromStartRating: rd.ratingsSetup.ShardChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          rd.ratingsSetup.ShardChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         rd.ratingsSetup.ShardChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  rd.ratingsSetup.ShardChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     rd.ratingsSetup.ShardChain.ProposerValidatorImportance,
	}
	shardRatingsStepData, err := computeRatingStep(shardRatingsStepsArgs)
	if err != nil {
		return ratingsStepsData{}, fmt.Errorf("%w while computing shard rating steps for epoch %d", err, chainParams.EnableEpoch)
	}

	metaRatingsStepsArgs := computeRatingStepArg{
		shardSize:                       chainParams.MetachainMinNumNodes,
		consensusSize:                   chainParams.MetachainConsensusGroupSize,
		roundTimeMillis:                 rd.roundDurationInMilliseconds,
		startRating:                     rd.ratingsSetup.General.StartRating,
		maxRating:                       rd.ratingsSetup.General.MaxRating,
		hoursToMaxRatingFromStartRating: rd.ratingsSetup.MetaChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          rd.ratingsSetup.MetaChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         rd.ratingsSetup.MetaChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  rd.ratingsSetup.MetaChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     rd.ratingsSetup.MetaChain.ProposerValidatorImportance,
	}
	metaRatingsStepData, err := computeRatingStep(metaRatingsStepsArgs)
	if err != nil {
		return ratingsStepsData{}, fmt.Errorf("%w while computing metachain rating steps for epoch %d", err, chainParams.EnableEpoch)
	}

	return ratingsStepsData{
		enableEpoch:          chainParams.EnableEpoch,
		shardRatingsStepData: shardRatingsStepData,
		metaRatingsStepData:  metaRatingsStepData,
	}, nil
}

// EpochConfirmed will be called whenever a new epoch is confirmed
func (rd *RatingsData) EpochConfirmed(epoch uint32, _ uint64) {
	log.Debug("RatingsData - epoch confirmed", "epoch", epoch)

	rd.mutConfiguration.Lock()
	defer rd.mutConfiguration.Unlock()

	newVersion, err := rd.getMatchingVersion(epoch)
	if err != nil {
		log.Error("RatingsData.EpochConfirmed - cannot get matching version", "epoch", epoch, "error", err)
		return
	}

	if rd.currentRatingsStepData.enableEpoch == newVersion.enableEpoch {
		return
	}

	rd.currentRatingsStepData = newVersion

	log.Debug("updated shard ratings step data",
		"epoch", epoch,
		"proposer increase rating step", newVersion.shardRatingsStepData.ProposerIncreaseRatingStep(),
		"proposer decrease rating step", newVersion.shardRatingsStepData.ProposerDecreaseRatingStep(),
		"validator increase rating step", newVersion.shardRatingsStepData.ValidatorIncreaseRatingStep(),
		"validator decrease rating step", newVersion.shardRatingsStepData.ValidatorDecreaseRatingStep(),
	)

	log.Debug("updated metachain ratings step data",
		"epoch", epoch,
		"proposer increase rating step", newVersion.metaRatingsStepData.ProposerIncreaseRatingStep(),
		"proposer decrease rating step", newVersion.metaRatingsStepData.ProposerDecreaseRatingStep(),
		"validator increase rating step", newVersion.metaRatingsStepData.ValidatorIncreaseRatingStep(),
		"validator decrease rating step", newVersion.metaRatingsStepData.ValidatorDecreaseRatingStep(),
	)
}

func (rd *RatingsData) getMatchingVersion(epoch uint32) (ratingsStepsData, error) {
	// the config values are sorted in descending order, so the matching version is the first one whose enable epoch is less or equal than the provided epoch
	for _, ratingsStepConfig := range rd.ratingsStepsConfig {
		if ratingsStepConfig.enableEpoch <= epoch {
			return ratingsStepConfig, nil
		}
	}

	// the code should never reach this point, since the config values are checked on the constructor
	return ratingsStepsData{}, process.ErrNoMatchingConfigForProvidedEpoch
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

	return rd.currentRatingsStepData.metaRatingsStepData
}

// ShardChainRatingsStepHandler returns the RatingsStepHandler used for the ShardChains
func (rd *RatingsData) ShardChainRatingsStepHandler() process.RatingsStepHandler {
	rd.mutConfiguration.RLock()
	defer rd.mutConfiguration.RUnlock()

	return rd.currentRatingsStepData.shardRatingsStepData
}

// IsInterfaceNil returns true if underlying object is nil
func (rd *RatingsData) IsInterfaceNil() bool {
	return rd == nil
}
