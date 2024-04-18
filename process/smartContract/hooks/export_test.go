package hooks

// GetMapActivationEpochs -
func (bh *BlockChainHookImpl) GetMapActivationEpochs() map[uint32]struct{} {
	return bh.mapActivationEpochs
}

// IsCrossShardForPayableCheck -
func (sbh *sovereignBlockChainHook) IsCrossShardForPayableCheck(sndAddress []byte, rcvAddress []byte) bool {
	return sbh.isCrossShardForPayableCheck(sndAddress, rcvAddress)
}
