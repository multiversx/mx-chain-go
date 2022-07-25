package hooks

// GetMapActivationEpochs -
func (bh *BlockChainHookImpl) GetMapActivationEpochs() map[uint32]struct{} {
	return bh.mapActivationEpochs
}
