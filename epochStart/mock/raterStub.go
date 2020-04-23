package mock

// RaterStub -
type RaterStub struct {
	GetRatingCalled                func(string) uint32
	GetStartRatingCalled           func() uint32
	GetSignedBlocksThresholdCalled func() float32
	ComputeIncreaseProposerCalled  func(shardId uint32, rating uint32) uint32
	ComputeDecreaseProposerCalled  func(shardId uint32, rating uint32, consecutiveMissedBlocks uint32) uint32
	RevertIncreaseProposerCalled   func(shardId uint32, rating uint32, nrReverts uint32) uint32
	ComputeIncreaseValidatorCalled func(shardId uint32, rating uint32) uint32
	ComputeDecreaseValidatorCalled func(shardId uint32, rating uint32) uint32
	GetChanceCalled                func(rating uint32) uint32
}

// GetRating -
func (rm *RaterStub) GetRating(pk string) uint32 {
	return rm.GetRatingCalled(pk)
}

// GetStartRating -
func (rm *RaterStub) GetStartRating() uint32 {
	if rm.GetStartRatingCalled != nil {
		return rm.GetStartRatingCalled()
	}
	return 10
}

// GetSignedBlocksThreshold -
func (rm *RaterStub) GetSignedBlocksThreshold() float32 {
	return rm.GetSignedBlocksThresholdCalled()
}

// ComputeIncreaseProposer -
func (rm *RaterStub) ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeIncreaseProposerCalled(shardId, currentRating)
}

// ComputeDecreaseProposer -
func (rm *RaterStub) ComputeDecreaseProposer(shardId uint32, currentRating uint32, consecutiveMisses uint32) uint32 {
	return rm.ComputeDecreaseProposerCalled(shardId, currentRating, consecutiveMisses)
}

// RevertIncreaseValidator -
func (rm *RaterStub) RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32 {
	return rm.RevertIncreaseProposerCalled(shardId, currentRating, nrReverts)
}

// ComputeIncreaseValidator -
func (rm *RaterStub) ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeIncreaseValidatorCalled(shardId, currentRating)
}

// ComputeDecreaseValidator -
func (rm *RaterStub) ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32 {
	return rm.ComputeDecreaseValidatorCalled(shardId, currentRating)
}

// GetChance -
func (rm *RaterStub) GetChance(rating uint32) uint32 {
	if rm.GetChanceCalled != nil {
		return rm.GetChanceCalled(rating)
	}

	return 80
}

// IsInterfaceNil -
func (rm *RaterStub) IsInterfaceNil() bool {
	return rm == nil
}
