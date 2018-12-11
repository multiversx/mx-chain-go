package exBlock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

func (eb *execBlock) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
	return eb.commitBlock(blockChain, header, block)
}

func (eb *execBlock) GetTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	return eb.getTransactionFromPool(destShardID, txHash)
}

func (eb *execBlock) RequestTransactionFromNetwork(body *block.TxBlockBody) {
	eb.requestBlockTransactions(body)
}

func (eb *execBlock) WaitForTxHashes() {
	eb.waitForTxHashes()
}

func (eb *execBlock) ReceivedTransaction(txHash []byte) {
	eb.receivedTransaction(txHash)
}

func (eb *execBlock) VerifyBlockSignature(header *block.Header) bool {
	return eb.verifyBlockSignature(header)
}
