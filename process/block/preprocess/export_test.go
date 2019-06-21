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

func (scr *smartContractResults) AddScrHashToRequestedList(txHash []byte) {
	scr.mutScrsForBlock.Lock()
	defer scr.mutScrsForBlock.Unlock()

	if scr.scrForBlock == nil {
		scr.scrForBlock = make(map[string]*scrInfo)
	}
	scr.scrForBlock[string(txHash)] = &scrInfo{scrShardInfo: &scrShardInfo{}}
}

func (scr *smartContractResults) IsScrHashRequested(txHash []byte) bool {
	scr.mutScrsForBlock.Lock()
	defer scr.mutScrsForBlock.Unlock()

	return !scr.scrForBlock[string(txHash)].has
}

func (scr *smartContractResults) SetMissingScr(missingTxs int) {
	scr.mutScrsForBlock.Lock()
	scr.missingScrs = missingTxs
	scr.mutScrsForBlock.Unlock()
}
