package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// RaterMock -
type RaterMock struct {
	GetRatingCalled                  func(string) uint32
	UpdateRatingFromTempRatingCalled func([]string) error
	GetStartRatingCalled             func() uint32
	ComputeIncreaseProposerCalled    func(shardId uint32, rating uint32) uint32
	ComputeDecreaseProposerCalled    func(shardId uint32, rating uint32) uint32
	ComputeIncreaseValidatorCalled   func(shardId uint32, rating uint32) uint32
	ComputeDecreaseValidatorCalled   func(shardId uint32, rating uint32) uint32
	GetChanceCalled                  func(rating uint32) uint32
	RatingReader                     sharding.RatingReader
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
func (rm *RaterMock) ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeIncreaseProposerCalled(shardId, currentRating)
}

// ComputeDecreaseProposer -
func (rm *RaterMock) ComputeDecreaseProposer(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeDecreaseProposerCalled(shardId, currentRating)
}

// ComputeIncreaseValidator -
func (rm *RaterMock) ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeIncreaseValidatorCalled(shardId, currentRating)
}

// ComputeDecreaseValidator -
func (rm *RaterMock) ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeDecreaseValidatorCalled(shardId, currentRating)
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
