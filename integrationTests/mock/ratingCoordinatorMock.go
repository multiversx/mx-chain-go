package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// RatingCoordinatorMock defines the behaviour of a struct able to do ratings for validators
type RatingCoordinatorMock struct {
}

func (rc *RatingCoordinatorMock) GetValidatorListAccordingToRating(
	v []sharding.Validator,
	round uint64,
) []sharding.Validator {
	return v
}

func (rc *RatingCoordinatorMock) IsInterfaceNil() bool {
	return false
}

func (rc *RatingCoordinatorMock) IncreaseRating([]string) {

}

func (rc *RatingCoordinatorMock) DecreaseRating([]string) {

}
