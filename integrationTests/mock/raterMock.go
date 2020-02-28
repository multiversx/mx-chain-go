package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// RaterMock -
type RaterMock struct {
	StartRating                      uint32
	MinRating                        uint32
	MaxRating                        uint32
	Chance                           uint32
	IncreaseProposer                 uint32
	DecreaseProposer                 uint32
	IncreaseValidator                uint32
	DecreaseValidator                uint32
	GetRatingCalled                  func(string) uint32
	UpdateRatingFromTempRatingCalled func([]string) error
	GetStartRatingCalled             func() uint32
	ComputeIncreaseProposerCalled    func(val uint32) uint32
	ComputeDecreaseProposerCalled    func(val uint32) uint32
	ComputeIncreaseValidatorCalled   func(val uint32) uint32
	ComputeDecreaseValidatorCalled   func(val uint32) uint32
	GetChanceCalled                  func(val uint32) uint32
	RatingReader                     sharding.RatingReader
}

func (rm *RaterMock) computeRating(val uint32, ratingStep int32) uint32 {
	newVal := int64(val) + int64(ratingStep)
	if newVal < int64(rm.MinRating) {
		return rm.MinRating
	}
	if newVal > int64(rm.MaxRating) {
		return rm.MaxRating
	}
	return uint32(newVal)
}

// GetRating -
func (rm *RaterMock) GetRating(pk string) uint32 {
	return rm.GetRatingCalled(pk)
}

// UpdateRatingFromTempRating -
func (rm *RaterMock) UpdateRatingFromTempRating(pks []string) error {
	return rm.UpdateRatingFromTempRatingCalled(pks)
}

// GetStartRating -
func (rm *RaterMock) GetStartRating() uint32 {
	return rm.GetStartRatingCalled()
}

// ComputeIncreaseProposer -
func (rm *RaterMock) ComputeIncreaseProposer(rating uint32) uint32 {
	return rm.ComputeIncreaseProposerCalled(rating)
}

// ComputeDecreaseProposer -
func (rm *RaterMock) ComputeDecreaseProposer(rating uint32) uint32 {
	return rm.ComputeDecreaseProposerCalled(rating)
}

// ComputeIncreaseValidator -
func (rm *RaterMock) ComputeIncreaseValidator(rating uint32) uint32 {
	return rm.ComputeIncreaseValidatorCalled(rating)
}

// ComputeDecreaseValidator -
func (rm *RaterMock) ComputeDecreaseValidator(rating uint32) uint32 {
	return rm.ComputeDecreaseValidatorCalled(rating)
}

// GetChance -
func (rm *RaterMock) GetChance(rating uint32) uint32 {
	return rm.GetChanceCalled(rating)
}

// SetRatingReader -
func (rm *RaterMock) SetRatingReader(reader sharding.RatingReader) {
	rm.RatingReader = reader
}

// IsInterfaceNil -
func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}
