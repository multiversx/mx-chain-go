package disabled

// BlockSizeComputationHandler implements the BlockSizeComputationHandler interface but does nothing as it is disabled
type BlockSizeComputationHandler struct {
}

// Init does nothing as it is a disabled component
func (b *BlockSizeComputationHandler) Init() {
}

// AddNumMiniBlocks does nothing as it is a disabled component
func (b *BlockSizeComputationHandler) AddNumMiniBlocks(_ int) {
}

// AddNumTxs does nothing as it is a disabled component
func (b *BlockSizeComputationHandler) AddNumTxs(_ int) {
}

// AddNumExecRes does nothing as it is a disabled component
func (b *BlockSizeComputationHandler) AddNumExecRes(_ int) {
}

// DecNumMiniBlocks does nothing
func (b *BlockSizeComputationHandler) DecNumMiniBlocks(numMiniBlocks int) {
}

// DecNumTxs does nothing
func (b *BlockSizeComputationHandler) DecNumTxs(numTxs int) {
}

// DecNumExecRes does nothing
func (b *BlockSizeComputationHandler) DecNumExecRes(numTxs int) {
}

// IsMaxBlockSizeReached returns false as it is a disabled components
func (b *BlockSizeComputationHandler) IsMaxBlockSizeReached(_ int, _ int) bool {
	return false
}

// IsMaxBlockSizeWithoutThrottleReached returns false as it is a disabled component
func (b *BlockSizeComputationHandler) IsMaxBlockSizeWithoutThrottleReached(_ int, _ int) bool {
	return false
}

// IsMaxExecResSizeReached returns false as it is a disabled components
func (b *BlockSizeComputationHandler) IsMaxExecResSizeReached(_ int) bool {
	return false
}

// IsInterfaceNil returns true if underlying object is nil
func (b *BlockSizeComputationHandler) IsInterfaceNil() bool {
	return b == nil
}
