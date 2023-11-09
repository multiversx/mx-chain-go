package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// BlockSizeThrottlerStub -
type BlockSizeThrottlerStub struct {
	GetCurrentMaxSizeCalled     func() uint32
	AddCalled                   func(round uint64, size uint32)
	SucceedCalled               func(round uint64)
	ComputeCurrentMaxSizeCalled func()
}

// GetCurrentMaxSize -
func (bsts *BlockSizeThrottlerStub) GetCurrentMaxSize() uint32 {
	if bsts.GetCurrentMaxSizeCalled != nil {
		return bsts.GetCurrentMaxSizeCalled()
	}

	return uint32(core.MegabyteSize * 90 / 100)
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

// ComputeCurrentMaxSize -
func (bsts *BlockSizeThrottlerStub) ComputeCurrentMaxSize() {
	if bsts.ComputeCurrentMaxSizeCalled != nil {
		bsts.ComputeCurrentMaxSizeCalled()
		return
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bsts *BlockSizeThrottlerStub) IsInterfaceNil() bool {
	return bsts == nil
}
