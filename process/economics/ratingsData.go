package economics

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
)

// RatingsData will store information about ratingsComputation
type RatingsData struct {
	startRating                     uint64
	maxRating                       uint64
	minRating                       uint64
	increaseRatingStep              uint64
	decreaseRatingStep              uint64
	proposerExtraIncreaseRatingStep uint64
	proposerExtraDecreaseRatingStep uint64
}

func checkRatingsValues(ratingSettings config.RatingSettings) error {
	if ratingSettings.MinRating > ratingSettings.MaxRating {
		return process.ErrMaxRatingIsSmallerThanMinRating
	}
	if ratingSettings.MaxRating < ratingSettings.StartRating || ratingSettings.MinRating > ratingSettings.StartRating {
		return process.ErrStartRatingNotBetweenMinAndMax
	}

	return nil
}

// StartRating will return the start rating
func (rd *RatingsData) StartRating() uint64 {
	return rd.startRating
}

// MaxRating will return the max rating
func (rd *RatingsData) MaxRating() uint64 {
	return rd.maxRating
}

// MinRating will return the min rating
func (rd *RatingsData) MinRating() uint64 {
	return rd.minRating
}

// IncreaseRatingStep will return the step to increase the rating
func (rd *RatingsData) IncreaseRatingStep() uint64 {
	return rd.increaseRatingStep
}

// DecreaseRatingStep will return the step to decrease the rating
func (rd *RatingsData) DecreaseRatingStep() uint64 {
	return rd.decreaseRatingStep
}

// ProposerExtraIncreaseRatingStep will return the extra increase rating step
func (rd *RatingsData) ProposerExtraIncreaseRatingStep() uint64 {
	return rd.proposerExtraIncreaseRatingStep
}

// ProposerExtraDecreaseRatingStep will return the extra decrease rating step
func (rd *RatingsData) ProposerExtraDecreaseRatingStep() uint64 {
	return rd.proposerExtraDecreaseRatingStep
}
