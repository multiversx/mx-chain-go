package disabled

// BlockSizeComputationHandler -
type BlockSizeComputationHandler struct {
}

// Init -
func (b *BlockSizeComputationHandler) Init() {
}

// AddNumMiniBlocks -
func (b *BlockSizeComputationHandler) AddNumMiniBlocks(_ int) {
}

// AddNumTxs -
func (b *BlockSizeComputationHandler) AddNumTxs(_ int) {
}

// IsMaxBlockSizeReached -
func (b *BlockSizeComputationHandler) IsMaxBlockSizeReached(_ int, _ int) bool {
	return false
}

// IsMaxBlockSizeWithoutThrottleReached -
func (b *BlockSizeComputationHandler) IsMaxBlockSizeWithoutThrottleReached(_ int, _ int) bool {
	return false
}

// IsInterfaceNil -
func (b *BlockSizeComputationHandler) IsInterfaceNil() bool {
	return b == nil
}
