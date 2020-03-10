package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// RaterMock -
type RaterMock struct {
	StartRating                    uint32
	MinRating                      uint32
	MaxRating                      uint32
	IncreaseProposer               uint32
	DecreaseProposer               uint32
	IncreaseValidator              uint32
	DecreaseValidator              uint32
	GetRatingCalled                func(string) uint32
	GetStartRatingCalled           func() uint32
	ComputeIncreaseProposerCalled  func(val uint32) uint32
	ComputeDecreaseProposerCalled  func(val uint32) uint32
	ComputeIncreaseValidatorCalled func(val uint32) uint32
	ComputeDecreaseValidatorCalled func(val uint32) uint32
	GetChancesCalled               func(string) uint32
	RatingReader                   sharding.RatingReader
}

// GetRating -
func (rm *RaterMock) GetRating(pk string) uint32 {
	return rm.GetRatingCalled(pk)
}

// GetRatings -
func (rm *RaterMock) GetRatings([]string) map[string]uint32 {
	return make(map[string]uint32)
}

// GetStartRating -
func (rm *RaterMock) GetStartRating() uint32 {
	return rm.GetStartRatingCalled()
}

// ComputeIncreaseProposer -
func (rm *RaterMock) ComputeIncreaseProposer(val uint32) uint32 {
	return rm.ComputeIncreaseProposerCalled(val)
}

// ComputeDecreaseProposer -
func (rm *RaterMock) ComputeDecreaseProposer(val uint32) uint32 {
	return rm.ComputeDecreaseProposerCalled(val)
}

// ComputeIncreaseValidator -
func (rm *RaterMock) ComputeIncreaseValidator(val uint32) uint32 {
	return rm.ComputeIncreaseValidatorCalled(val)
}

// ComputeDecreaseValidator -
func (rm *RaterMock) ComputeDecreaseValidator(val uint32) uint32 {
	return rm.ComputeDecreaseValidatorCalled(val)
}

// GetChances -
func (rm *RaterMock) GetChances(pk string) uint32 {
	return rm.GetChancesCalled(pk)
}

// SetRatingReader -
func (rm *RaterMock) SetRatingReader(reader sharding.RatingReader) {
	rm.RatingReader = reader
}

// IsInterfaceNil -
func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}
