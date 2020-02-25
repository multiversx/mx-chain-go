package mock

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// BlockSizeThrottlerStub -
type BlockSizeThrottlerStub struct {
	MaxItemsToAddCalled   func() uint32
	AddCalled             func(round uint64, items uint32)
	SucceedCalled         func(round uint64)
	ComputeMaxItemsCalled func()
}

// MaxItemsToAdd -
func (bsts *BlockSizeThrottlerStub) MaxItemsToAdd() uint32 {
	if bsts.MaxItemsToAddCalled != nil {
		return bsts.MaxItemsToAddCalled()
	}

	return process.MaxItemsInBlock
}

// Add -
func (bsts *BlockSizeThrottlerStub) Add(round uint64, items uint32) {
	if bsts.AddCalled != nil {
		bsts.AddCalled(round, items)
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

// ComputeMaxItems -
func (bsts *BlockSizeThrottlerStub) ComputeMaxItems() {
	if bsts.ComputeMaxItemsCalled != nil {
		bsts.ComputeMaxItemsCalled()
		return
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bsts *BlockSizeThrottlerStub) IsInterfaceNil() bool {
	if bsts == nil {
		return true
	}
	return false
}
