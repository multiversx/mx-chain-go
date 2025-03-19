package rating

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"golang.org/x/exp/slices"
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
	statusHandler               core.AppStatusHandler
	mutStatusHandler            sync.RWMutex
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

	// avoid any invalid configuration where ratings are not sorted by epoch
	slices.SortFunc(ratingsConfig.ShardChain.RatingStepsByEpoch, func(a, b config.RatingSteps) bool {
		return a.EnableEpoch < b.EnableEpoch
	})
	slices.SortFunc(ratingsConfig.MetaChain.RatingStepsByEpoch, func(a, b config.RatingSteps) bool {
		return a.EnableEpoch < b.EnableEpoch
	})

	if !checkForEpochZeroConfiguration(args) {
		return nil, process.ErrMissingConfigurationForEpochZero
	}

	currentChainParameters := args.ChainParametersHolder.CurrentChainParameters()
	shardChainRatingSteps, _ := getRatingStepsForEpoch(args.EpochNotifier.CurrentEpoch(), ratingsConfig.ShardChain.RatingStepsByEpoch)
	arg := computeRatingStepArg{
		shardSize:                       currentChainParameters.ShardMinNumNodes,
		consensusSize:                   currentChainParameters.ShardConsensusGroupSize,
		roundTimeMillis:                 args.RoundDurationMilliseconds,
		startRating:                     ratingsConfig.General.StartRating,
		maxRating:                       ratingsConfig.General.MaxRating,
		hoursToMaxRatingFromStartRating: shardChainRatingSteps.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          shardChainRatingSteps.ProposerDecreaseFactor,
		validatorDecreaseFactor:         shardChainRatingSteps.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  shardChainRatingSteps.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     shardChainRatingSteps.ProposerValidatorImportance,
	}
	shardRatingStep, err := computeRatingStep(arg)
	if err != nil {
		return nil, err
	}

	metaChainRatingSteps, _ := getRatingStepsForEpoch(args.EpochNotifier.CurrentEpoch(), ratingsConfig.MetaChain.RatingStepsByEpoch)
	arg = computeRatingStepArg{
		shardSize:                       currentChainParameters.MetachainMinNumNodes,
		consensusSize:                   currentChainParameters.MetachainConsensusGroupSize,
		roundTimeMillis:                 args.RoundDurationMilliseconds,
		startRating:                     ratingsConfig.General.StartRating,
		maxRating:                       ratingsConfig.General.MaxRating,
		hoursToMaxRatingFromStartRating: metaChainRatingSteps.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          metaChainRatingSteps.ProposerDecreaseFactor,
		validatorDecreaseFactor:         metaChainRatingSteps.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  metaChainRatingSteps.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     metaChainRatingSteps.ProposerValidatorImportance,
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
		statusHandler:               statusHandler.NewNilStatusHandler(),
	}

	err = ratingData.computeRatingStepsConfig(args.ChainParametersHolder.AllChainParameters())
	if err != nil {
		return nil, err
	}

	args.EpochNotifier.RegisterNotifyHandler(ratingData)

	return ratingData, nil
}

func checkForEpochZeroConfiguration(args RatingsDataArg) bool {
	_, foundShardChainRatingSteps := getRatingStepsForEpoch(0, args.Config.ShardChain.RatingStepsByEpoch)
	_, foundMetaChainRatingSteps := getRatingStepsForEpoch(0, args.Config.MetaChain.RatingStepsByEpoch)
	_, foundChainParams := getChainParamsForEpoch(0, args.ChainParametersHolder.AllChainParameters())

	return foundShardChainRatingSteps && foundMetaChainRatingSteps && foundChainParams
}

func (rd *RatingsData) computeRatingStepsConfig(chainParamsList []config.ChainParametersByEpochConfig) error {
	if len(chainParamsList) == 0 {
		return process.ErrEmptyChainParametersConfiguration
	}

	// there are multiple scenarios when ratingSteps can change:
	// 		1. chain parameters change in a specific epoch
	//		2. shard/meta rating steps change in a specific epoch
	// thus we extract first all configured epochs in a map, from all meta, chard and chain parameters
	// this way we make sure that for each activation epoch we got the proper config, taking all params into consideration
	configuredEpochsMap := make(map[uint32]struct{})
	for _, ratingStepsForEpoch := range rd.ratingsSetup.ShardChain.RatingStepsByEpoch {
		configuredEpochsMap[ratingStepsForEpoch.EnableEpoch] = struct{}{}
	}

	for _, ratingStepsForEpoch := range rd.ratingsSetup.MetaChain.RatingStepsByEpoch {
		configuredEpochsMap[ratingStepsForEpoch.EnableEpoch] = struct{}{}
	}

	for _, chainParams := range chainParamsList {
		configuredEpochsMap[chainParams.EnableEpoch] = struct{}{}
	}

	ratingsStepsConfig := make([]ratingsStepsData, 0)
	for epoch := range configuredEpochsMap {
		configForEpoch, err := rd.computeRatingStepsConfigForEpoch(epoch, chainParamsList)
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

func (rd *RatingsData) computeRatingStepsConfigForEpoch(
	epoch uint32,
	chainParamsList []config.ChainParametersByEpochConfig,
) (ratingsStepsData, error) {
	chainParams, _ := getChainParamsForEpoch(epoch, chainParamsList)

	shardChainRatingSteps, _ := getRatingStepsForEpoch(epoch, rd.ratingsSetup.ShardChain.RatingStepsByEpoch)
	shardRatingsStepsArgs := computeRatingStepArg{
		shardSize:                       chainParams.ShardMinNumNodes,
		consensusSize:                   chainParams.ShardConsensusGroupSize,
		roundTimeMillis:                 rd.roundDurationInMilliseconds,
		startRating:                     rd.ratingsSetup.General.StartRating,
		maxRating:                       rd.ratingsSetup.General.MaxRating,
		hoursToMaxRatingFromStartRating: shardChainRatingSteps.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          shardChainRatingSteps.ProposerDecreaseFactor,
		validatorDecreaseFactor:         shardChainRatingSteps.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  shardChainRatingSteps.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     shardChainRatingSteps.ProposerValidatorImportance,
	}
	shardRatingsStepData, err := computeRatingStep(shardRatingsStepsArgs)
	if err != nil {
		return ratingsStepsData{}, fmt.Errorf("%w while computing shard rating steps for epoch %d", err, chainParams.EnableEpoch)
	}

	metaChainRatingSteps, _ := getRatingStepsForEpoch(epoch, rd.ratingsSetup.MetaChain.RatingStepsByEpoch)
	metaRatingsStepsArgs := computeRatingStepArg{
		shardSize:                       chainParams.MetachainMinNumNodes,
		consensusSize:                   chainParams.MetachainConsensusGroupSize,
		roundTimeMillis:                 rd.roundDurationInMilliseconds,
		startRating:                     rd.ratingsSetup.General.StartRating,
		maxRating:                       rd.ratingsSetup.General.MaxRating,
		hoursToMaxRatingFromStartRating: metaChainRatingSteps.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          metaChainRatingSteps.ProposerDecreaseFactor,
		validatorDecreaseFactor:         metaChainRatingSteps.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  metaChainRatingSteps.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     metaChainRatingSteps.ProposerValidatorImportance,
	}
	metaRatingsStepData, err := computeRatingStep(metaRatingsStepsArgs)
	if err != nil {
		return ratingsStepsData{}, fmt.Errorf("%w while computing metachain rating steps for epoch %d", err, chainParams.EnableEpoch)
	}

	return ratingsStepsData{
		enableEpoch:          epoch,
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

	rd.updateRatingsMetrics(epoch)
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
	err := checkRatingStepsByEpochConfigForDest(settings.ShardChain.RatingStepsByEpoch, "shardChain")
	if err != nil {
		return err
	}

	return checkRatingStepsByEpochConfigForDest(settings.MetaChain.RatingStepsByEpoch, "metaChain")
}

func checkRatingStepsByEpochConfigForDest(ratingStepsByEpoch []config.RatingSteps, configDestination string) error {
	if len(ratingStepsByEpoch) == 0 {
		return fmt.Errorf("%w for %s",
			process.ErrInvalidRatingsConfig,
			configDestination)
	}

	for _, ratingStepsForEpoch := range ratingStepsByEpoch {
		if ratingStepsForEpoch.HoursToMaxRatingFromStartRating == 0 {
			return fmt.Errorf("%w hoursToMaxRatingFromStartRating: %s, epoch: %d",
				process.ErrHoursToMaxRatingFromStartRatingZero,
				configDestination,
				ratingStepsForEpoch.EnableEpoch)
		}
		if ratingStepsForEpoch.ConsecutiveMissedBlocksPenalty < 1 {
			return fmt.Errorf("%w: %s consecutiveMissedBlocksPenalty: %v, epoch: %d",
				process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne,
				configDestination,
				ratingStepsForEpoch.ConsecutiveMissedBlocksPenalty,
				ratingStepsForEpoch.EnableEpoch)
		}
		if ratingStepsForEpoch.ProposerDecreaseFactor > -1 || ratingStepsForEpoch.ValidatorDecreaseFactor > -1 {
			return fmt.Errorf("%w: %s decrease steps - proposer: %v, validator: %v, epoch: %d",
				process.ErrDecreaseRatingsStepMoreThanMinusOne,
				configDestination,
				ratingStepsForEpoch.ProposerDecreaseFactor,
				ratingStepsForEpoch.ValidatorDecreaseFactor,
				ratingStepsForEpoch.EnableEpoch)
		}
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

// SetStatusHandler sets the provided status handler if not nil
func (rd *RatingsData) SetStatusHandler(handler core.AppStatusHandler) error {
	if check.IfNil(handler) {
		return process.ErrNilAppStatusHandler
	}

	rd.mutStatusHandler.Lock()
	rd.statusHandler = handler
	rd.mutStatusHandler.Unlock()

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (rd *RatingsData) IsInterfaceNil() bool {
	return rd == nil
}

func (rd *RatingsData) updateRatingsMetrics(epoch uint32) {
	rd.mutStatusHandler.RLock()
	defer rd.mutStatusHandler.RUnlock()

	currentShardRatingsStep, _ := getRatingStepsForEpoch(epoch, rd.ratingsSetup.ShardChain.RatingStepsByEpoch)
	rd.statusHandler.SetUInt64Value(common.MetricRatingsShardChainHoursToMaxRatingFromStartRating, uint64(currentShardRatingsStep.HoursToMaxRatingFromStartRating))
	rd.statusHandler.SetStringValue(common.MetricRatingsShardChainProposerValidatorImportance, fmt.Sprintf("%f", currentShardRatingsStep.ProposerValidatorImportance))
	rd.statusHandler.SetStringValue(common.MetricRatingsShardChainProposerDecreaseFactor, fmt.Sprintf("%f", currentShardRatingsStep.ProposerDecreaseFactor))
	rd.statusHandler.SetStringValue(common.MetricRatingsShardChainValidatorDecreaseFactor, fmt.Sprintf("%f", currentShardRatingsStep.ValidatorDecreaseFactor))
	rd.statusHandler.SetStringValue(common.MetricRatingsShardChainConsecutiveMissedBlocksPenalty, fmt.Sprintf("%f", currentShardRatingsStep.ConsecutiveMissedBlocksPenalty))

	currentMetaRatingsStep, _ := getRatingStepsForEpoch(epoch, rd.ratingsSetup.MetaChain.RatingStepsByEpoch)
	rd.statusHandler.SetUInt64Value(common.MetricRatingsMetaChainHoursToMaxRatingFromStartRating, uint64(currentMetaRatingsStep.HoursToMaxRatingFromStartRating))
	rd.statusHandler.SetStringValue(common.MetricRatingsMetaChainProposerValidatorImportance, fmt.Sprintf("%f", currentMetaRatingsStep.ProposerValidatorImportance))
	rd.statusHandler.SetStringValue(common.MetricRatingsMetaChainProposerDecreaseFactor, fmt.Sprintf("%f", currentMetaRatingsStep.ProposerDecreaseFactor))
	rd.statusHandler.SetStringValue(common.MetricRatingsMetaChainValidatorDecreaseFactor, fmt.Sprintf("%f", currentMetaRatingsStep.ValidatorDecreaseFactor))
	rd.statusHandler.SetStringValue(common.MetricRatingsMetaChainConsecutiveMissedBlocksPenalty, fmt.Sprintf("%f", currentMetaRatingsStep.ConsecutiveMissedBlocksPenalty))
}

func getRatingStepsForEpoch(epoch uint32, ratingStepsPerEpoch []config.RatingSteps) (config.RatingSteps, bool) {
	var ratingSteps config.RatingSteps
	found := false
	for _, ratingStepsForEpoch := range ratingStepsPerEpoch {
		if ratingStepsForEpoch.EnableEpoch <= epoch {
			ratingSteps = ratingStepsForEpoch
			found = true
		}
	}

	return ratingSteps, found
}

func getChainParamsForEpoch(epoch uint32, chainParamsList []config.ChainParametersByEpochConfig) (config.ChainParametersByEpochConfig, bool) {
	slices.SortFunc(chainParamsList, func(a, b config.ChainParametersByEpochConfig) bool {
		return a.EnableEpoch < b.EnableEpoch
	})

	var chainParams config.ChainParametersByEpochConfig
	found := false
	for _, chainParamsForEpoch := range chainParamsList {
		if chainParamsForEpoch.EnableEpoch <= epoch {
			chainParams = chainParamsForEpoch
			found = true
		}
	}

	return chainParams, found
}
