package mock

// BlockSizeComputationStub -
type BlockSizeComputationStub struct {
	InitCalled                  func()
	AddNumMiniBlocksCalled      func(int)
	AddNumTxsCalled             func(int)
	IsMaxBlockSizeReachedCalled func(int, int) bool
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

// IsMaxBlockSizeReached -
func (bscs *BlockSizeComputationStub) IsMaxBlockSizeReached(numNewMiniBlocks int, numNewTxs int) bool {
	if bscs.IsMaxBlockSizeReachedCalled != nil {
		return bscs.IsMaxBlockSizeReachedCalled(numNewMiniBlocks, numNewTxs)
	}
	return false
}

// IsInterfaceNil -
func (bscs *BlockSizeComputationStub) IsInterfaceNil() bool {
	return bscs == nil
}
