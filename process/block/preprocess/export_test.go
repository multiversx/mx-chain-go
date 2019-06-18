package preprocess

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

func (txs *transactions) WaitForTxHashes(waitTime time.Duration) {
	txs.waitForTxHashes(waitTime)
}

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

func (txs *transactions) CreateTxInfo(tx *transaction.Transaction, senderShardID uint32, receiverShardID uint32) *txInfo {
	txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	return &txInfo{tx: tx, txShardInfo: txShardInfo}
}

func (txs *transactions) SetTxsForBlock(hash string, txInfo *txInfo) {
	txs.mutTxsForBlock.Lock()
	txs.txsForBlock[hash] = txInfo
	txs.mutTxsForBlock.Unlock()
}
