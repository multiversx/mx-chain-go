package mock

type BlockSizeThrottlerStub struct {
	MaxItemsToAddCalled   func() uint32
	AddCalled             func(round uint64, items uint32)
	SucceedCalled         func(round uint64)
	ComputeMaxItemsCalled func()
}

func (bsts *BlockSizeThrottlerStub) MaxItemsToAdd() uint32 {
	if bsts.MaxItemsToAddCalled != nil {
		return bsts.MaxItemsToAddCalled()
	}

	return 15000
}

func (bsts *BlockSizeThrottlerStub) Add(round uint64, items uint32) {
	if bsts.AddCalled != nil {
		bsts.AddCalled(round, items)
		return
	}
}

func (bsts *BlockSizeThrottlerStub) Succeed(round uint64) {
	if bsts.SucceedCalled != nil {
		bsts.SucceedCalled(round)
		return
	}
}

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
