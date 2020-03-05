package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// RaterMock -
type RaterMock struct {
	StartRating       uint32
	MinRating         uint32
	MaxRating         uint32
	IncreaseProposer  uint32
	DecreaseProposer  uint32
	IncreaseValidator uint32
	DecreaseValidator uint32

	GetRatingCalled                func(string) uint32
	GetStartRatingCalled           func() uint32
	ComputeIncreaseProposerCalled  func(val uint32) uint32
	ComputeDecreaseProposerCalled  func(val uint32) uint32
	ComputeIncreaseValidatorCalled func(val uint32) uint32
	ComputeDecreaseValidatorCalled func(val uint32) uint32
	UpdateListAndIndexCalled       func(pubKey string, list string, index int) error
	RatingReader                   sharding.RatingReader
}

// GetNewMockRater -
func GetNewMockRater() *RaterMock {
	raterMock := &RaterMock{}
	raterMock.GetRatingCalled = func(s string) uint32 {
		return raterMock.StartRating
	}

	raterMock.GetStartRatingCalled = func() uint32 {
		return raterMock.StartRating
	}
	raterMock.ComputeIncreaseProposerCalled = func(val uint32) uint32 {
		return raterMock.computeRating(val, int32(raterMock.IncreaseProposer))
	}
	raterMock.ComputeDecreaseProposerCalled = func(val uint32) uint32 {
		return raterMock.computeRating(val, int32(0-raterMock.DecreaseProposer))
	}
	raterMock.ComputeIncreaseValidatorCalled = func(val uint32) uint32 {
		return raterMock.computeRating(val, int32(raterMock.IncreaseValidator))
	}
	raterMock.ComputeDecreaseValidatorCalled = func(val uint32) uint32 {
		return raterMock.computeRating(val, int32(0-raterMock.DecreaseValidator))
	}

	return raterMock
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

// SetRatingReader -
func (rm *RaterMock) SetRatingReader(reader sharding.RatingReader) {
	rm.RatingReader = reader
}

// SetListIndexUpdater -
func (rm *RaterMock) SetListIndexUpdater(updater sharding.ListIndexUpdaterHandler) {
}

// UpdateListAndIndex -
func (rm *RaterMock) UpdateListAndIndex(pubKey string, shardID uint32, list string, index int) error {
	if rm.UpdateListAndIndexCalled != nil {
		return rm.UpdateListAndIndexCalled(pubKey, list, index)
	}

	return nil
}

// IsInterfaceNil -
func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}
