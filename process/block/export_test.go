package block

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func (bp *blockProcessor) GetTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	return bp.getTransactionFromPool(destShardID, txHash)
}

func (bp *blockProcessor) RequestTransactionFromNetwork(body *block.TxBlockBody) int {
	return bp.requestBlockTransactions(body)
}

func (bp *blockProcessor) WaitForTxHashes(waitTime time.Duration) {
	bp.waitForTxHashes(waitTime)
}

func (bp *blockProcessor) ReceivedTransaction(txHash []byte) {
	bp.receivedTransaction(txHash)
}

func (bp *blockProcessor) ComputeHeaderHash(hdr *block.Header) ([]byte, error) {
	return bp.computeHeaderHash(hdr)
}

func (bp *blockProcessor) DisplayLogInfo(header *block.Header, txBlock *block.TxBlockBody, headerHash []byte) {
	bp.displayLogInfo(header, txBlock, headerHash)
}

func SortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	return sortTxByNonce(txShardStore)
}
