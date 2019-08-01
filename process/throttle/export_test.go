package throttle

func (bst *blockSizeThrottle) SetMaxItems(maxItems uint32) {
	bst.maxItems = maxItems
}

func (bst *blockSizeThrottle) RoundInLastItemAdded() uint64 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].round
}

func (bst *blockSizeThrottle) RoundInItemAdded(index uint32) uint64 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[index].round
}

func (bst *blockSizeThrottle) ItemsInLastItemAdded() uint32 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].items
}

func (bst *blockSizeThrottle) ItemsInItemAdded(index uint32) uint32 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[index].items
}

func (bst *blockSizeThrottle) SucceedInLastItemAdded() bool {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].succeed
}

func (bst *blockSizeThrottle) SucceedInItemAdded(index uint32) bool {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[index].succeed
}

func (bst *blockSizeThrottle) MaxItemsInLastItemAdded() uint32 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].maxItems
}

func (bst *blockSizeThrottle) MaxItemsInItemAdded(index uint32) uint32 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[index].maxItems
}

func (bst *blockSizeThrottle) GetMaxItemsWhenSucceed(lastActionMaxItems uint32) uint32 {
	return bst.getMaxItemsWhenSucceed(lastActionMaxItems)
}

func (bst *blockSizeThrottle) GetCloserAboveMaxItemsUsedWithoutSucceed(currentMaxItems uint32) uint32 {
	return bst.getCloserAboveMaxItemsUsedWithoutSucceed(currentMaxItems)
}

func (bst *blockSizeThrottle) GetMaxItemsWhenNotSucceed(lastActionMaxItems uint32) uint32 {
	return bst.getMaxItemsWhenNotSucceed(lastActionMaxItems)
}

func (bst *blockSizeThrottle) GetCloserBelowMaxItemsUsedWithSucceed(currentMaxItems uint32) uint32 {
	return bst.getCloserBelowMaxItemsUsedWithSucceed(currentMaxItems)
}

func (bst *blockSizeThrottle) JumpAbovePercent() uint32 {
	return jumpAbovePercent
}

func (bst *blockSizeThrottle) JumpBelowPercent() uint32 {
	return jumpBelowPercent
}

func (bst *blockSizeThrottle) JumpAboveFactor() float32 {
	return jumpAboveFactor
}

func (bst *blockSizeThrottle) JumpBelowFactor() float32 {
	return jumpBelowFactor
}

func (bst *blockSizeThrottle) SetSucceed(round uint64, succeed bool) {
	bst.mutThrottler.Lock()
	for i := len(bst.statistics) - 1; i >= 0; i-- {
		if bst.statistics[i].round == round {
			bst.statistics[i].succeed = succeed
			break
		}
	}
	bst.mutThrottler.Unlock()
}
