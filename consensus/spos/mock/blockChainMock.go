package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
)

type BlockChainMock struct {
	blockchain.StorageService
	GenesisBlockCalled func() *block.Header
	SetGenesisBlockCalled func(*block.Header)
	GenesisHeaderHashCalled func() []byte
	SetGenesisHeaderHashCalled func([]byte)
	CurrentBlockHeaderCalled func() *block.Header
	SetCurrentBlockHeaderCalled func(*block.Header)
	CurrentBlockHeaderHashCalled func() []byte
	SetCurrentBlockHeaderHashCalled func([]byte)
	CurrentTxBlockBodyCalled func() block.Body
	SetCurrentTxBlockBodyCalled func(block.Body)
	LocalHeightCalled func() int64
	SetLocalHeightCalled func(int64)
	NetworkHeightCalled func() int64
	SetNetworkHeightCalled func(int64)
	IsBadBlockCalled func([]byte) bool
	PutBadBlockCalled func([]byte)
}

// GenesisBlock returns the genesis block header pointer
func (bc *BlockChainMock) GenesisBlock() *block.Header {
	if bc.GenesisBlockCalled != nil {
		return bc.GenesisBlockCalled()
	}
	return nil
}

// SetGenesisBlock sets the genesis block header pointer
func (bc *BlockChainMock) SetGenesisBlock(genesisBlock *block.Header) {
	if bc.SetGenesisBlockCalled != nil {
		bc.SetGenesisBlockCalled(genesisBlock)
	}
}

// GenesisHeaderHash returns the genesis block header hash
func (bc *BlockChainMock) GenesisHeaderHash() []byte {
	if bc.GenesisHeaderHashCalled != nil {
		return bc.GenesisHeaderHashCalled()
	}
	return nil
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *BlockChainMock) SetGenesisHeaderHash(hash []byte) {
	if bc.SetGenesisHeaderHashCalled != nil {
		bc.SetGenesisHeaderHashCalled(hash)
	}
}

// CurrentBlockHeader returns current block header pointer
func (bc *BlockChainMock) CurrentBlockHeader() *block.Header {
	if bc.CurrentBlockHeaderCalled != nil {
		return bc.CurrentBlockHeaderCalled()
	}
	return nil
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *BlockChainMock) SetCurrentBlockHeader(header *block.Header) {
	if bc.SetCurrentBlockHeaderCalled != nil {
		bc.SetCurrentBlockHeaderCalled(header)
	}
}

// CurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChainMock) CurrentBlockHeaderHash() []byte {
	if bc.CurrentBlockHeaderHashCalled != nil {
		return bc.CurrentBlockHeaderHashCalled()
	}
	return nil
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChainMock) SetCurrentBlockHeaderHash(hash []byte) {
	if bc.SetCurrentBlockHeaderHashCalled != nil {
		bc.SetCurrentBlockHeaderHashCalled(hash)
	}
}

// CurrentTxBlockBody returns the tx block body pointer
func (bc *BlockChainMock) CurrentTxBlockBody() block.Body {
	if bc.CurrentTxBlockBodyCalled != nil {
		return bc.CurrentTxBlockBodyCalled()
	}
	return nil
}

// SetCurrentTxBlockBody sets the tx block body pointer
func (bc *BlockChainMock) SetCurrentTxBlockBody(body block.Body) {
	if bc.SetCurrentTxBlockBodyCalled != nil {
		bc.SetCurrentTxBlockBodyCalled(body)
	}
}

// LocalHeight returns the height of the local chain
func (bc *BlockChainMock) LocalHeight() int64 {
	if bc.LocalHeightCalled != nil {
		return bc.LocalHeight()
	}
	return 0
}

// SetLocalHeight sets the height of the local chain
func (bc *BlockChainMock) SetLocalHeight(height int64) {
	if bc.SetLocalHeightCalled != nil {
		bc.SetLocalHeightCalled(height)
	}
}

// NetworkHeight sets the percieved height of the network chain
func (bc *BlockChainMock) NetworkHeight() int64 {
	if bc.NetworkHeightCalled != nil {
		return bc.NetworkHeightCalled()
	}
	return 0
}

// SetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChainMock) SetNetworkHeight(height int64) {
	if bc.SetNetworkHeightCalled != nil {
		bc.SetNetworkHeightCalled(height)
	}
}

// IsBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *BlockChainMock) IsBadBlock(blockHash []byte) bool {
	if bc.IsBadBlockCalled != nil {
		return bc.IsBadBlockCalled(blockHash)
	}
	return false
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *BlockChainMock) PutBadBlock(blockHash []byte) {
	if bc.PutBadBlockCalled != nil {
		bc.PutBadBlockCalled(blockHash)
	}
}