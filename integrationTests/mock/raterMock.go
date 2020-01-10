package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

type RaterMock struct {
	StartRating                      uint32
	MinRating                        uint32
	MaxRating                        uint32
	IncreaseProposer                 uint32
	DecreaseProposer                 uint32
	IncreaseValidator                uint32
	DecreaseValidator                uint32
	GetRatingCalled                  func(string) uint32
	UpdateRatingFromTempRatingCalled func(string)
	GetStartRatingCalled             func() uint32
	ComputeIncreaseProposerCalled    func(val uint32) uint32
	ComputeDecreaseProposerCalled    func(val uint32) uint32
	ComputeIncreaseValidatorCalled   func(val uint32) uint32
	ComputeDecreaseValidatorCalled   func(val uint32) uint32
	GetChanceCalled                  func(val uint32) uint32
	RatingReader                     sharding.RatingReader
}

func GetNewMockRater() *RaterMock {
	raterMock := &RaterMock{
		StartRating:       5,
		MinRating:         1,
		MaxRating:         10,
		IncreaseProposer:  1,
		DecreaseProposer:  1,
		IncreaseValidator: 1,
		DecreaseValidator: 1,
	}
	raterMock.GetRatingCalled = func(s string) uint32 {
		return raterMock.StartRating
	}
	raterMock.UpdateRatingFromTempRatingCalled = func(s string) {
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
	raterMock.GetChanceCalled = func(val uint32) uint32 {
		return raterMock.StartRating
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

func (rm *RaterMock) GetRating(pk string) uint32 {
	return rm.GetRatingCalled(pk)
}

func (rm *RaterMock) UpdateRatingFromTempRating(pk string) {
	rm.UpdateRatingFromTempRatingCalled(pk)
}

func (rm *RaterMock) GetStartRating() uint32 {
	return rm.GetStartRatingCalled()
}
func (rm *RaterMock) ComputeIncreaseProposer(rating uint32) uint32 {
	return rm.ComputeIncreaseProposerCalled(rating)
}
func (rm *RaterMock) ComputeDecreaseProposer(rating uint32) uint32 {
	return rm.ComputeDecreaseProposerCalled(rating)
}
func (rm *RaterMock) ComputeIncreaseValidator(rating uint32) uint32 {
	return rm.ComputeIncreaseValidatorCalled(rating)
}
func (rm *RaterMock) ComputeDecreaseValidator(rating uint32) uint32 {
	return rm.ComputeDecreaseValidatorCalled(rating)
}
func (rm *RaterMock) GetChance(rating uint32) uint32 {
	return rm.GetChanceCalled(rating)
}
func (rm *RaterMock) SetRatingReader(reader sharding.RatingReader) {
	rm.RatingReader = reader
}
func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}
