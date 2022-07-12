package hooks

// SetFlagOptimizeNFTStore -
func (bh *BlockChainHookImpl) SetFlagOptimizeNFTStore(value bool) {
	bh.flagOptimizeNFTStore.SetValue(value)
}

// GetMapActivationEpochs -
func (bh *BlockChainHookImpl) GetMapActivationEpochs() map[uint32]struct{} {
	return bh.mapActivationEpochs
}
