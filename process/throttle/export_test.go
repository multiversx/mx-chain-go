package throttle

func (bst *blockSizeThrottle) SetCurrentMaxSize(currentMaxSize uint32) {
	bst.currentMaxSize = currentMaxSize
}

func (bst *blockSizeThrottle) RoundInLastSizeAdded() uint64 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].round
}

func (bst *blockSizeThrottle) SizeInLastSizeAdded() uint32 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].size
}

func (bst *blockSizeThrottle) SucceedInLastSizeAdded() bool {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].succeed
}

func (bst *blockSizeThrottle) SucceedInSizeAdded(index uint32) bool {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[index].succeed
}

func (bst *blockSizeThrottle) CurrentMaxSizeInLastSizeAdded() uint32 {
	bst.mutThrottler.RLock()
	defer bst.mutThrottler.RUnlock()

	return bst.statistics[len(bst.statistics)-1].currentMaxSize
}

func (bst *blockSizeThrottle) GetMaxSizeWhenSucceed(lastActionMaxSize uint32) uint32 {
	return bst.getMaxSizeWhenSucceed(lastActionMaxSize)
}

func (bst *blockSizeThrottle) GetCloserAboveCurrentMaxSizeUsedWithoutSucceed(currentMaxSize uint32) uint32 {
	return bst.getCloserAboveCurrentMaxSizeUsedWithoutSucceed(currentMaxSize)
}

func (bst *blockSizeThrottle) GetMaxSizeWhenNotSucceed(lastActionMaxSize uint32) uint32 {
	return bst.getMaxSizeWhenNotSucceed(lastActionMaxSize)
}

func (bst *blockSizeThrottle) GetCloserBelowCurrentMaxSizeUsedWithSucceed(currentMaxSize uint32) uint32 {
	return bst.getCloserBelowCurrentMaxSizeUsedWithSucceed(currentMaxSize)
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
