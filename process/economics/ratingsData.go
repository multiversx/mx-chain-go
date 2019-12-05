package economics

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

// RatingsData will store information about ratingsComputation
type RatingsData struct {
	startRating                 uint32
	maxRating                   uint32
	minRating                   uint32
	proposerIncreaseRatingStep  uint32
	proposerDecreaseRatingStep  uint32
	validatorIncreaseRatingStep uint32
	validatorDecreaseRatingStep uint32
}

// NewRatingsData creates a new RatingsData instance
func NewRatingsData(
	settings config.RatingSettings,
) (*RatingsData, error) {
	if settings.MinRating > settings.MaxRating {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if settings.MaxRating < settings.StartRating || settings.MinRating > settings.StartRating {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}

	return &RatingsData{
		startRating:                 settings.StartRating,
		maxRating:                   settings.MaxRating,
		minRating:                   settings.MinRating,
		proposerIncreaseRatingStep:  settings.ProposerIncreaseRatingStep,
		proposerDecreaseRatingStep:  settings.ProposerDecreaseRatingStep,
		validatorIncreaseRatingStep: settings.ValidatorIncreaseRatingStep,
		validatorDecreaseRatingStep: settings.ValidatorDecreaseRatingStep,
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

// ProposerIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingsData) ProposerIncreaseRatingStep() uint32 {
	return rd.proposerIncreaseRatingStep
}

// ProposerDecreaseRatingStep will return the rating step decrease for proposer
func (rd *RatingsData) ProposerDecreaseRatingStep() uint32 {
	return rd.proposerDecreaseRatingStep
}

// ValidatorIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingsData) ValidatorIncreaseRatingStep() uint32 {
	return rd.validatorIncreaseRatingStep
}

// ValidatorDecreaseRatingStep will return the rating step decrease for validator
func (rd *RatingsData) ValidatorDecreaseRatingStep() uint32 {
	return rd.validatorDecreaseRatingStep
}
