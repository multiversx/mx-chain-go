package hooks

// SetFlagOptimizeNFTStore -
func (bh *BlockChainHookImpl) SetFlagOptimizeNFTStore(value bool) {
	bh.flagOptimizeNFTStore.SetValue(value)
}
