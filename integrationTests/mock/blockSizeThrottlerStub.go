package mock

type BlockSizeThrottlerStub struct {
	MaxItemsToAddCalled func() uint32
	UpdateCalled        func(succeeded bool, round uint64, items uint32)
}

func (bsts *BlockSizeThrottlerStub) MaxItemsToAdd() uint32 {
	if bsts.MaxItemsToAddCalled != nil {
		return bsts.MaxItemsToAddCalled()
	}

	return 15000
}

func (bsts *BlockSizeThrottlerStub) Update(succeeded bool, round uint64, items uint32) {
	if bsts.UpdateCalled != nil {
		bsts.UpdateCalled(succeeded, round, items)
	}
}
