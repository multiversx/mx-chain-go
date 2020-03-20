package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// BlockSizeThrottlerStub -
type BlockSizeThrottlerStub struct {
	GetMaxSizeCalled     func() uint32
	AddCalled            func(round uint64, size uint32)
	SucceedCalled        func(round uint64)
	ComputeMaxSizeCalled func()
}

// GetMaxSize -
func (bsts *BlockSizeThrottlerStub) GetMaxSize() uint32 {
	if bsts.GetMaxSizeCalled != nil {
		return bsts.GetMaxSizeCalled()
	}

	return core.MaxSizeInBytes
}

// Add -
func (bsts *BlockSizeThrottlerStub) Add(round uint64, size uint32) {
	if bsts.AddCalled != nil {
		bsts.AddCalled(round, size)
		return
	}
}

// Succeed -
func (bsts *BlockSizeThrottlerStub) Succeed(round uint64) {
	if bsts.SucceedCalled != nil {
		bsts.SucceedCalled(round)
		return
	}
}

// ComputeMaxSize -
func (bsts *BlockSizeThrottlerStub) ComputeMaxSize() {
	if bsts.ComputeMaxSizeCalled != nil {
		bsts.ComputeMaxSizeCalled()
		return
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bsts *BlockSizeThrottlerStub) IsInterfaceNil() bool {
	return bsts == nil
}
