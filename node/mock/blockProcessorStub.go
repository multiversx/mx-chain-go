package mock

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type BlockProcessorStub struct {
}

func (bps *BlockProcessorStub) SetOnRequestTransaction(f func(destShardID uint32, txHash []byte)) {
	panic("implement me")
}

func (bps *BlockProcessorStub) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block *block.TxBlockBody) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) RevertAccountState() {
	panic("implement me")
}

func (bps *BlockProcessorStub) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body *block.TxBlockBody, haveTime func() time.Duration) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) CreateGenesisBlockBody(balances map[string]*big.Int, shardId uint32) (*block.StateBlockBody, error) {
	panic("implement me")
}

func (bps *BlockProcessorStub) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (*block.TxBlockBody, error) {
	panic("implement me")
}

func (bps *BlockProcessorStub) RemoveBlockTxsFromPool(body *block.TxBlockBody) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) GetRootHash() []byte {
	panic("implement me")
}

func (bps BlockProcessorStub) CheckBlockValidity(blockChain *blockchain.BlockChain, header *block.Header) bool {
	panic("implement me")
}
