package mock

import (
	"github.com/ElrondNetwork/elrond-go/process/economics"
)

type RatingsInfoMock struct {
	StartRatingVal                 uint32
	MaxRatingVal                   uint32
	MinRatingVal                   uint32
	ProposerIncreaseRatingStepVal  uint32
	ProposerDecreaseRatingStepVal  uint32
	ValidatorIncreaseRatingStepVal uint32
	ValidatorDecreaseRatingStepVal uint32
	SelectionChancesVal            []economics.SelectionChance
}

// StartRating will return the start rating
func (rd *RatingsInfoMock) StartRating() uint32 {
	return rd.StartRatingVal
}

// MaxRating will return the max rating
func (rd *RatingsInfoMock) MaxRating() uint32 {
	return rd.MaxRatingVal
}

// MinRating will return the min rating
func (rd *RatingsInfoMock) MinRating() uint32 {
	return rd.MinRatingVal
}

// ProposerIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingsInfoMock) ProposerIncreaseRatingStep() uint32 {
	return rd.ProposerIncreaseRatingStepVal
}

// ProposerDecreaseRatingStep will return the rating step decrease for proposer
func (rd *RatingsInfoMock) ProposerDecreaseRatingStep() uint32 {
	return rd.ProposerDecreaseRatingStepVal
}

// ValidatorIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingsInfoMock) ValidatorIncreaseRatingStep() uint32 {
	return rd.ValidatorIncreaseRatingStepVal
}

// ValidatorDecreaseRatingStep will return the rating step decrease for validator
func (rd *RatingsInfoMock) ValidatorDecreaseRatingStep() uint32 {
	return rd.ValidatorDecreaseRatingStepVal
}

// SelectionChances will return the array of selectionChances and thresholds
func (rd *RatingsInfoMock) SelectionChances() []economics.SelectionChance {
	return rd.SelectionChancesVal
}
