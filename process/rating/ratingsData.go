package rating

import (
	"fmt"
	"math"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.RatingsInfoHandler = (*RatingsData)(nil)

const milisecondsInHour = 3600 * 1000

type computeRatingStepArg struct {
	shardSize                       uint32
	consensusSize                   uint32
	roundTimeMilis                  uint64
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
	startRating           uint32
	maxRating             uint32
	minRating             uint32
	signedBlocksThreshold float32
	metaRatingsStepData   process.RatingsStepHandler
	shardRatingsStepData  process.RatingsStepHandler
	selectionChances      []process.SelectionChance
}

// RatingsDataArg contains information for the creation of the new ratingsData
type RatingsDataArg struct {
	Config                   config.RatingsConfig
	ShardConsensusSize       uint32
	MetaConsensusSize        uint32
	ShardMinNodes            uint32
	MetaMinNodes             uint32
	RoundDurationMiliseconds uint64
}

// NewRatingsData creates a new RatingsData instance
func NewRatingsData(args RatingsDataArg) (*RatingsData, error) {
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
		shardSize:                       args.ShardMinNodes,
		consensusSize:                   args.ShardConsensusSize,
		roundTimeMilis:                  args.RoundDurationMiliseconds,
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
		shardSize:                       args.MetaMinNodes,
		consensusSize:                   args.MetaConsensusSize,
		roundTimeMilis:                  args.RoundDurationMiliseconds,
		startRating:                     ratingsConfig.General.StartRating,
		maxRating:                       ratingsConfig.General.MaxRating,
		hoursToMaxRatingFromStartRating: ratingsConfig.MetaChain.HoursToMaxRatingFromStartRating,
		proposerDecreaseFactor:          ratingsConfig.MetaChain.ProposerDecreaseFactor,
		validatorDecreaseFactor:         ratingsConfig.MetaChain.ValidatorDecreaseFactor,
		consecutiveMissedBlocksPenalty:  ratingsConfig.MetaChain.ConsecutiveMissedBlocksPenalty,
		proposerValidatorImportance:     ratingsConfig.MetaChain.ProposerValidatorImportance,
	}
	//metaRatingStep, err := computeRatingStep(arg)
	//if err != nil {
	//	return nil, err
	//}

	return &RatingsData{
		startRating:           ratingsConfig.General.StartRating,
		maxRating:             ratingsConfig.General.MaxRating,
		minRating:             ratingsConfig.General.MinRating,
		signedBlocksThreshold: ratingsConfig.General.SignedBlocksThreshold,
		metaRatingsStepData:   shardRatingStep,
		shardRatingsStepData:  shardRatingStep,
		selectionChances:      chances,
	}, nil
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
	blocksProducedInHours := uint64(arg.hoursToMaxRatingFromStartRating*milisecondsInHour) / arg.roundTimeMilis
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
	return rd.startRating
}

// MaxRating will return the max rating
func (rd *RatingsData) MaxRating() uint32 {
	return rd.maxRating
}

// MinRating will return the min rating
func (rd *RatingsData) MinRating() uint32 {
	return rd.minRating
}

// SignedBlocksThreshold will return the signed blocks threshold
func (rd *RatingsData) SignedBlocksThreshold() float32 {
	return rd.signedBlocksThreshold
}

// SelectionChances will return the array of selectionChances and thresholds
func (rd *RatingsData) SelectionChances() []process.SelectionChance {
	return rd.selectionChances
}

// MetaChainRatingsStepHandler returns the RatingsStepHandler used for the Metachain
func (rd *RatingsData) MetaChainRatingsStepHandler() process.RatingsStepHandler {
	return rd.metaRatingsStepData
}

// ShardChainRatingsStepHandler returns the RatingsStepHandler used for the ShardChains
func (rd *RatingsData) ShardChainRatingsStepHandler() process.RatingsStepHandler {
	return rd.shardRatingsStepData
}

// IsInterfaceNil returns true if underlying object is nil
func (rd *RatingsData) IsInterfaceNil() bool {
	return rd == nil
}
