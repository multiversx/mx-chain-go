package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

func (eb *blockProcessor) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
	return eb.commitBlock(blockChain, header, block)
}

func (eb *blockProcessor) GetTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	return eb.getTransactionFromPool(destShardID, txHash)
}

func (eb *blockProcessor) RequestTransactionFromNetwork(body *block.TxBlockBody) {
	eb.requestBlockTransactions(body)
}

func (eb *blockProcessor) WaitForTxHashes() {
	eb.waitForTxHashes()
}

func (eb *blockProcessor) ReceivedTransaction(txHash []byte) {
	eb.receivedTransaction(txHash)
}

func (eb *blockProcessor) VerifyBlockSignature(header *block.Header) bool {
	return eb.verifyBlockSignature(header)
}
