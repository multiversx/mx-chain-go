package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// RaterMock -
type RaterMock struct {
	StartRating           uint32
	MinRating             uint32
	MaxRating             uint32
	Chance                uint32
	IncreaseProposer      int32
	DecreaseProposer      int32
	IncreaseValidator     int32
	DecreaseValidator     int32
	MetaIncreaseProposer  int32
	MetaDecreaseProposer  int32
	MetaIncreaseValidator int32
	MetaDecreaseValidator int32

	GetRatingCalled                func(string) uint32
	GetStartRatingCalled           func() uint32
	GetSignedBlocksThresholdCalled func() float32
	ComputeIncreaseProposerCalled  func(shardId uint32, rating uint32) uint32
	ComputeDecreaseProposerCalled  func(shardId uint32, rating uint32, consecutiveMissedBlocks uint32) uint32
	RevertIncreaseProposerCalled   func(shardId uint32, rating uint32, nrReverts uint32) uint32
	RevertIncreaseValidatorCalled  func(shardId uint32, rating uint32, nrReverts uint32) uint32
	ComputeIncreaseValidatorCalled func(shardId uint32, rating uint32) uint32
	ComputeDecreaseValidatorCalled func(shardId uint32, rating uint32) uint32
	GetChancesCalled               func(val uint32) uint32
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
	raterMock.ComputeIncreaseProposerCalled = func(shardId uint32, rating uint32) uint32 {
		var ratingStep int32
		if shardId == core.MetachainShardId {
			ratingStep = raterMock.MetaIncreaseProposer
		} else {
			ratingStep = raterMock.IncreaseProposer
		}
		return raterMock.computeRating(rating, ratingStep)
	}
	raterMock.RevertIncreaseProposerCalled = func(shardId uint32, rating uint32, nrReverts uint32) uint32 {
		var ratingStep int32
		if shardId == core.MetachainShardId {
			ratingStep = raterMock.MetaIncreaseValidator
		} else {
			ratingStep = raterMock.IncreaseValidator
		}
		computedStep := -ratingStep * int32(nrReverts)
		return raterMock.computeRating(rating, computedStep)
	}
	raterMock.ComputeDecreaseProposerCalled = func(shardId uint32, rating uint32, consecutiveMissedBlocks uint32) uint32 {
		var ratingStep int32
		if shardId == core.MetachainShardId {
			ratingStep = raterMock.MetaDecreaseProposer
		} else {
			ratingStep = raterMock.DecreaseProposer
		}
		return raterMock.computeRating(rating, ratingStep)
	}
	raterMock.ComputeIncreaseValidatorCalled = func(shardId uint32, rating uint32) uint32 {
		var ratingStep int32
		if shardId == core.MetachainShardId {
			ratingStep = raterMock.MetaIncreaseValidator
		} else {
			ratingStep = raterMock.IncreaseValidator
		}
		return raterMock.computeRating(rating, ratingStep)
	}
	raterMock.ComputeDecreaseValidatorCalled = func(shardId uint32, rating uint32) uint32 {
		var ratingStep int32
		if shardId == core.MetachainShardId {
			ratingStep = raterMock.MetaDecreaseValidator
		} else {
			ratingStep = raterMock.DecreaseValidator
		}
		return raterMock.computeRating(rating, ratingStep)
	}
	raterMock.GetChancesCalled = func(val uint32) uint32 {
		return raterMock.Chance
	}
	return raterMock
}

func (rm *RaterMock) computeRating(rating uint32, ratingStep int32) uint32 {
	newVal := int64(rating) + int64(ratingStep)
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
	if rm.GetRatingCalled != nil {
		return rm.GetRatingCalled(pk)
	}
	return 50
}

// GetStartRating -
func (rm *RaterMock) GetStartRating() uint32 {
	if rm.GetStartRatingCalled != nil {
		return rm.GetStartRatingCalled()
	}
	return 10
}

// GetSignedBlocksThreshold -
func (rm *RaterMock) GetSignedBlocksThreshold() float32 {
	if rm.GetSignedBlocksThresholdCalled != nil {
		return rm.GetSignedBlocksThresholdCalled()
	}
	return 1
}

// ComputeIncreaseProposer -
func (rm *RaterMock) ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32 {
	if rm.ComputeIncreaseProposerCalled != nil {
		return rm.ComputeIncreaseProposerCalled(shardId, currentRating)
	}
	return 1
}

// ComputeDecreaseProposer -
func (rm *RaterMock) ComputeDecreaseProposer(shardId uint32, currentRating uint32, consecutiveMisses uint32) uint32 {
	if rm.ComputeDecreaseProposerCalled != nil {
		return rm.ComputeDecreaseProposerCalled(shardId, currentRating, consecutiveMisses)
	}
	return 1
}

// RevertIncreaseValidator -
func (rm *RaterMock) RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32 {
	if rm.RevertIncreaseValidatorCalled != nil {
		return rm.RevertIncreaseProposerCalled(shardId, currentRating, nrReverts)
	}
	return 1
}

// ComputeIncreaseValidator -
func (rm *RaterMock) ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32 {
	if rm.ComputeIncreaseValidatorCalled != nil {
		return rm.ComputeIncreaseValidatorCalled(shardId, currentRating)
	}
	return 1
}

// ComputeDecreaseValidator -
func (rm *RaterMock) ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32 {
	if rm.ComputeDecreaseValidatorCalled != nil {
		return rm.ComputeDecreaseValidatorCalled(shardId, currentRating)
	}
	return 1
}

// GetChance -
func (rm *RaterMock) GetChance(rating uint32) uint32 {
	if rm.GetChancesCalled != nil {
		return rm.GetChancesCalled(rating)
	}
	return 1
}

// IsInterfaceNil -
func (rm *RaterMock) IsInterfaceNil() bool {
	return rm == nil
}
