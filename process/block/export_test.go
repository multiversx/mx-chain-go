package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

func (bp *blockProcessor) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
	return bp.commitBlock(blockChain, header, block)
}

func (bp *blockProcessor) GetTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	return bp.getTransactionFromPool(destShardID, txHash)
}

func (bp *blockProcessor) RequestTransactionFromNetwork(body *block.TxBlockBody) {
	bp.requestBlockTransactions(body)
}

func (bp *blockProcessor) WaitForTxHashes() {
	bp.waitForTxHashes()
}

func (bp *blockProcessor) ReceivedTransaction(txHash []byte) {
	bp.receivedTransaction(txHash)
}
