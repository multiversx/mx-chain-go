package rating

import (
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.RatingsStepHandler = (*RatingStep)(nil)

// RatingStep will store information about ratingsComputation specific for a shard or metachain
type RatingStep struct {
	proposerIncreaseRatingStep     int32
	proposerDecreaseRatingStep     int32
	validatorIncreaseRatingStep    int32
	validatorDecreaseRatingStep    int32
	consecutiveMissedBlocksPenalty float32
}

// NewRatingStepData creates a new RatingStep instance
func NewRatingStepData(
	proposerIncreaseRatingStep int32,
	proposerDecreaseRatingStep int32,
	validatorIncreaseRatingStep int32,
	validatorDecreaseRatingStep int32,
	consecutiveMissedBlocksPenalty float32,
) process.RatingsStepHandler {
	return &RatingStep{
		proposerIncreaseRatingStep:     proposerIncreaseRatingStep,
		proposerDecreaseRatingStep:     proposerDecreaseRatingStep,
		validatorIncreaseRatingStep:    validatorIncreaseRatingStep,
		validatorDecreaseRatingStep:    validatorDecreaseRatingStep,
		consecutiveMissedBlocksPenalty: consecutiveMissedBlocksPenalty,
	}
}

// ProposerIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingStep) ProposerIncreaseRatingStep() int32 {
	return rd.proposerIncreaseRatingStep
}

// ProposerDecreaseRatingStep will return the rating step decrease for proposer
func (rd *RatingStep) ProposerDecreaseRatingStep() int32 {
	return rd.proposerDecreaseRatingStep
}

// ValidatorIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingStep) ValidatorIncreaseRatingStep() int32 {
	return rd.validatorIncreaseRatingStep
}

// ValidatorDecreaseRatingStep will return the rating step decrease for validator
func (rd *RatingStep) ValidatorDecreaseRatingStep() int32 {
	return rd.validatorDecreaseRatingStep
}

// ConsecutiveMissedBlocksPenalty will return the penalty increase for consecutive block misses
func (rd *RatingStep) ConsecutiveMissedBlocksPenalty() float32 {
	return rd.consecutiveMissedBlocksPenalty
}
