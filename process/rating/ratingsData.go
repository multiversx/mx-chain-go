package rating

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

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

// NewRatingsData creates a new RatingsData instance
func NewRatingsData(settings *config.RatingsConfig) (*RatingsData, error) {
	if settings.General.MinRating < 1 {
		return nil, process.ErrMinRatingSmallerThanOne
	}
	if settings.General.MinRating > settings.General.MaxRating {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if settings.General.MaxRating < settings.General.StartRating || settings.General.MinRating > settings.General.StartRating {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}
	if settings.General.SignedBlocksThreshold > 1 || settings.General.SignedBlocksThreshold < 0 {
		return nil, process.ErrSignedBlocksThresholdNotBetweenZeroAndOne
	}
	if settings.MetaChain.ConsecutiveMissedBlocksPenalty < 1 ||
		settings.ShardChain.ConsecutiveMissedBlocksPenalty < 1 {
		return nil, process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne
	}

	chances := make([]process.SelectionChance, 0)
	for _, chance := range settings.General.SelectionChances {
		chances = append(chances, &SelectionChance{
			MaxThreshold:  chance.MaxThreshold,
			ChancePercent: chance.ChancePercent,
		})
	}

	return &RatingsData{
		startRating:           settings.General.StartRating,
		maxRating:             settings.General.MaxRating,
		minRating:             settings.General.MinRating,
		signedBlocksThreshold: settings.General.SignedBlocksThreshold,
		metaRatingsStepData:   NewRatingStepData(settings.MetaChain.RatingSteps),
		shardRatingsStepData:  NewRatingStepData(settings.ShardChain.RatingSteps),
		selectionChances:      chances,
	}, nil
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
