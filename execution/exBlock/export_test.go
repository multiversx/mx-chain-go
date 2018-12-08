package exBlock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"time"
)

func (eb *execBlock) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.Block) error {
	return eb.commitBlock(blockChain, header, block)
}

func (eb *execBlock) GetTransactionFromPool(destShardID uint32, txHash []byte) *transaction.Transaction {
	return eb.getTransactionFromPool(destShardID, txHash)
}

func (eb *execBlock) RequestTransactionFromNetwork(destShardID uint32, txHash []byte, waitTime time.Duration) *transaction.Transaction {
	return eb.requestTransactionFromNetwork(destShardID, txHash, waitTime)
}

func (eb *execBlock) ReceivedTransaction(txHash []byte) {
	eb.receivedTransaction(txHash)
}

func (eb *execBlock) VerifyBlockSignature(header *block.Header) bool {
	return eb.verifyBlockSignature(header)
}
