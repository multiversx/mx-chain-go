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

func (bps *BlockProcessorStub) ProcessAndCommit(blockChain *blockchain.BlockChain, header *block.Header, body block.Body, haveTime func() time.Duration) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) CommitBlock(blockChain *blockchain.BlockChain, header *block.Header, block block.Body) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) RevertAccountState() {
	panic("implement me")
}

func (bps *BlockProcessorStub) ProcessBlock(blockChain *blockchain.BlockChain, header *block.Header, body block.Body, haveTime func() time.Duration) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) CreateGenesisBlock(balances map[string]*big.Int) (rootHash []byte, err error) {
	panic("implement me")
}

func (bps *BlockProcessorStub) CreateTxBlockBody(shardId uint32, maxTxInBlock int, round int32, haveTime func() bool) (block.Body, error) {
	panic("implement me")
}

func (bps *BlockProcessorStub) RemoveBlockTxsFromPool(body block.Body) error {
	panic("implement me")
}

func (bps *BlockProcessorStub) GetRootHash() []byte {
	panic("implement me")
}

func (bps BlockProcessorStub) CheckBlockValidity(blockChain *blockchain.BlockChain, header *block.Header) bool {
	panic("implement me")
}

func (bps BlockProcessorStub)CreateMiniBlockHeaders(body block.Body) ([]block.MiniBlockHeader, error) {
	panic("implement me")
}
