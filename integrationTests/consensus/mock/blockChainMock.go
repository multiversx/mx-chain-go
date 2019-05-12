package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

// BlockChainMock is a mock implementation of the blockchain interface
type BlockChainMock struct {
	BlockChain                      data.ChainHandler
	GetGenesisHeaderCalled          func() data.HeaderHandler
	SetGenesisHeaderCalled          func(handler data.HeaderHandler) error
	GetGenesisHeaderHashCalled      func() []byte
	SetGenesisHeaderHashCalled      func([]byte)
	GetCurrentBlockHeaderCalled     func() data.HeaderHandler
	SetCurrentBlockHeaderCalled     func(data.HeaderHandler) error
	GetCurrentBlockHeaderHashCalled func() []byte
	SetCurrentBlockHeaderHashCalled func([]byte)
	GetCurrentBlockBodyCalled       func() data.BodyHandler
	SetCurrentBlockBodyCalled       func(data.BodyHandler) error
	GetLocalHeightCalled            func() int64
	SetLocalHeightCalled            func(int64)
	GetNetworkHeightCalled          func() int64
	SetNetworkHeightCalled          func(int64)
	HasBadBlockCalled               func([]byte) bool
	PutBadBlockCalled               func([]byte)
}

// GetGenesisHeader returns the genesis block header pointer
func (bc *BlockChainMock) GetGenesisHeader() data.HeaderHandler {
	if bc.GetGenesisHeaderCalled != nil {
		return bc.GetGenesisHeaderCalled()
	}
	return bc.BlockChain.GetGenesisHeader()
}

// SetGenesisHeader sets the genesis block header pointer
func (bc *BlockChainMock) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if bc.SetGenesisHeaderCalled != nil {
		return bc.SetGenesisHeaderCalled(genesisBlock)
	}
	return bc.BlockChain.SetGenesisHeader(genesisBlock)
}

// GetGenesisHeaderHash returns the genesis block header hash
func (bc *BlockChainMock) GetGenesisHeaderHash() []byte {
	if bc.GetGenesisHeaderHashCalled != nil {
		return bc.GetGenesisHeaderHashCalled()
	}
	return bc.BlockChain.GetGenesisHeaderHash()
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *BlockChainMock) SetGenesisHeaderHash(hash []byte) {
	if bc.SetGenesisHeaderHashCalled != nil {
		bc.SetGenesisHeaderHashCalled(hash)
	}
	bc.BlockChain.SetGenesisHeaderHash(hash)
}

// GetCurrentBlockHeader returns current block header pointer
func (bc *BlockChainMock) GetCurrentBlockHeader() data.HeaderHandler {
	if bc.GetCurrentBlockHeaderCalled != nil {
		return bc.GetCurrentBlockHeaderCalled()
	}
	return bc.BlockChain.GetCurrentBlockHeader()
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *BlockChainMock) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if bc.SetCurrentBlockHeaderCalled != nil {
		return bc.SetCurrentBlockHeaderCalled(header)
	}
	return bc.BlockChain.SetCurrentBlockHeader(header)
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChainMock) GetCurrentBlockHeaderHash() []byte {
	if bc.GetCurrentBlockHeaderHashCalled != nil {
		return bc.GetCurrentBlockHeaderHashCalled()
	}
	return bc.BlockChain.GetCurrentBlockHeaderHash()
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChainMock) SetCurrentBlockHeaderHash(hash []byte) {
	if bc.SetCurrentBlockHeaderHashCalled != nil {
		bc.SetCurrentBlockHeaderHashCalled(hash)
	}
	bc.BlockChain.SetCurrentBlockHeaderHash(hash)
}

// GetCurrentBlockBody returns the tx block body pointer
func (bc *BlockChainMock) GetCurrentBlockBody() data.BodyHandler {
	if bc.GetCurrentBlockBodyCalled != nil {
		return bc.GetCurrentBlockBodyCalled()
	}
	return bc.BlockChain.GetCurrentBlockBody()
}

// SetCurrentBlockBody sets the tx block body pointer
func (bc *BlockChainMock) SetCurrentBlockBody(body data.BodyHandler) error {
	if bc.SetCurrentBlockBodyCalled != nil {
		return bc.SetCurrentBlockBodyCalled(body)
	}
	return bc.BlockChain.SetCurrentBlockBody(body)
}

// GetLocalHeight returns the height of the local chain
func (bc *BlockChainMock) GetLocalHeight() int64 {
	if bc.GetLocalHeightCalled != nil {
		return bc.GetLocalHeightCalled()
	}
	return bc.BlockChain.GetLocalHeight()
}

// SetLocalHeight sets the height of the local chain
func (bc *BlockChainMock) SetLocalHeight(height int64) {
	if bc.SetLocalHeightCalled != nil {
		bc.SetLocalHeightCalled(height)
	}
	bc.BlockChain.SetLocalHeight(height)
}

// GetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChainMock) GetNetworkHeight() int64 {
	if bc.GetNetworkHeightCalled != nil {
		return bc.GetNetworkHeightCalled()
	}
	return bc.BlockChain.GetNetworkHeight()
}

// SetNetworkHeight sets the percieved height of the network chain
func (bc *BlockChainMock) SetNetworkHeight(height int64) {
	if bc.SetNetworkHeightCalled != nil {
		bc.SetNetworkHeightCalled(height)
	}
	bc.BlockChain.SetNetworkHeight(height)
}

// HasBadBlock returns true if the provided hash is blacklisted as a bad block, or false otherwise
func (bc *BlockChainMock) HasBadBlock(blockHash []byte) bool {
	if bc.HasBadBlockCalled != nil {
		return bc.HasBadBlockCalled(blockHash)
	}
	return bc.BlockChain.HasBadBlock(blockHash)
}

// PutBadBlock adds the given serialized block to the bad block cache, blacklisting it
func (bc *BlockChainMock) PutBadBlock(blockHash []byte) {
	if bc.PutBadBlockCalled != nil {
		bc.PutBadBlockCalled(blockHash)
	}
	bc.BlockChain.PutBadBlock(blockHash)
}
