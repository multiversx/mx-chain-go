package preprocess

func (txs *transactions) ReceivedTransaction(txHash []byte) {
	txs.receivedTransaction(txHash)
}

func (txs *transactions) AddTxHashToRequestedList(txHash []byte) {
	txs.mutTxsForBlock.Lock()
	defer txs.mutTxsForBlock.Unlock()

	if txs.txsForBlock == nil {
		txs.txsForBlock = make(map[string]*txInfo)
	}
	txs.txsForBlock[string(txHash)] = &txInfo{txShardInfo: &txShardInfo{}}
}

func (txs *transactions) IsTxHashRequested(txHash []byte) bool {
	txs.mutTxsForBlock.Lock()
	defer txs.mutTxsForBlock.Unlock()

	return !txs.txsForBlock[string(txHash)].has
}

func (txs *transactions) SetMissingTxs(missingTxs int) {
	txs.mutTxsForBlock.Lock()
	txs.missingTxs = missingTxs
	txs.mutTxsForBlock.Unlock()
}
