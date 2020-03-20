package mock

// RatingsData will store information about ratingsComputation specific for a shard or metachain
type RatingStepMock struct {
	ProposerIncreaseRatingStepProperty  int32
	ProposerDecreaseRatingStepProperty  int32
	ValidatorIncreaseRatingStepProperty int32
	ValidatorDecreaseRatingStepProperty int32
}

// ProposerIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingStepMock) ProposerIncreaseRatingStep() int32 {
	return rd.ProposerIncreaseRatingStepProperty
}

// ProposerDecreaseRatingStep will return the rating step decrease for proposer
func (rd *RatingStepMock) ProposerDecreaseRatingStep() int32 {
	return rd.ProposerDecreaseRatingStepProperty
}

// ValidatorIncreaseRatingStep will return the rating step increase for validator
func (rd *RatingStepMock) ValidatorIncreaseRatingStep() int32 {
	return rd.ValidatorIncreaseRatingStepProperty
}

// ValidatorDecreaseRatingStep will return the rating step decrease for validator
func (rd *RatingStepMock) ValidatorDecreaseRatingStep() int32 {
	return rd.ValidatorDecreaseRatingStepProperty
}
