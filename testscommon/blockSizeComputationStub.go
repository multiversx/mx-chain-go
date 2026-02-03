package testscommon

// BlockSizeComputationStub -
type BlockSizeComputationStub struct {
	InitCalled                                 func()
	AddNumMiniBlocksCalled                     func(int)
	AddNumTxsCalled                            func(int)
	AddNumExecResCalled                        func(int)
	IsMaxBlockSizeReachedCalled                func(int, int) bool
	IsMaxBlockSizeWithoutThrottleReachedCalled func(int, int) bool
	IsMaxExecResSizeReachedCalled              func(int) bool
	DecNumMiniBlocksCalled                     func(numMiniBlocks int)
	DecNumTxsCalled                            func(numTxs int)
	DecNumExecResCalled                        func(numTxs int)
}

// Init -
func (bscs *BlockSizeComputationStub) Init() {
	if bscs.InitCalled != nil {
		bscs.InitCalled()
	}
}

// AddNumMiniBlocks -
func (bscs *BlockSizeComputationStub) AddNumMiniBlocks(numMiniBlocks int) {
	if bscs.AddNumMiniBlocksCalled != nil {
		bscs.AddNumMiniBlocksCalled(numMiniBlocks)
	}
}

// AddNumTxs -
func (bscs *BlockSizeComputationStub) AddNumTxs(numTxs int) {
	if bscs.AddNumTxsCalled != nil {
		bscs.AddNumTxsCalled(numTxs)
	}
}

// AddNumExecRes -
func (bscs *BlockSizeComputationStub) AddNumExecRes(numExecRes int) {
	if bscs.AddNumExecResCalled != nil {
		bscs.AddNumExecResCalled(numExecRes)
	}
}

// DecNumMiniBlocks -
func (bscs *BlockSizeComputationStub) DecNumMiniBlocks(numMiniBlocks int) {
	if bscs.DecNumMiniBlocksCalled != nil {
		bscs.DecNumMiniBlocksCalled(numMiniBlocks)
	}
}

// DecNumTxs -
func (bscs *BlockSizeComputationStub) DecNumTxs(numTxs int) {
	if bscs.DecNumTxsCalled != nil {
		bscs.DecNumTxsCalled(numTxs)
	}
}

// DecNumExecRes -
func (bscs *BlockSizeComputationStub) DecNumExecRes(numExecRes int) {
	if bscs.DecNumExecResCalled != nil {
		bscs.DecNumExecResCalled(numExecRes)
	}
}

// IsMaxBlockSizeWithoutThrottleReached -
func (bscs *BlockSizeComputationStub) IsMaxBlockSizeWithoutThrottleReached(numNewMiniBlocks int, numNewTxs int) bool {
	if bscs.IsMaxBlockSizeWithoutThrottleReachedCalled != nil {
		return bscs.IsMaxBlockSizeWithoutThrottleReachedCalled(numNewMiniBlocks, numNewTxs)
	}
	return false
}

// IsMaxBlockSizeReached -
func (bscs *BlockSizeComputationStub) IsMaxBlockSizeReached(numNewMiniBlocks int, numNewTxs int) bool {
	if bscs.IsMaxBlockSizeReachedCalled != nil {
		return bscs.IsMaxBlockSizeReachedCalled(numNewMiniBlocks, numNewTxs)
	}
	return false
}

// IsMaxExecResSizeReached -
func (bscs *BlockSizeComputationStub) IsMaxExecResSizeReached(numNewExecRes int) bool {
	if bscs.IsMaxExecResSizeReachedCalled != nil {
		return bscs.IsMaxExecResSizeReachedCalled(numNewExecRes)
	}
	return false
}

// IsInterfaceNil -
func (bscs *BlockSizeComputationStub) IsInterfaceNil() bool {
	return bscs == nil
}
