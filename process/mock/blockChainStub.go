package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
)

// BlockChainStub is a mock implementation of the blockchain interface
type BlockChainStub struct {
	GetGenesisHeaderCalled          func() data.HeaderHandler
	SetGenesisHeaderCalled          func(handler data.HeaderHandler) error
	GetGenesisHeaderHashCalled      func() []byte
	SetGenesisHeaderHashCalled      func([]byte)
	GetCurrentBlockHeaderCalled     func() data.HeaderHandler
	SetCurrentBlockHeaderCalled     func(data.HeaderHandler) error
	GetCurrentBlockHeaderHashCalled func() []byte
	SetCurrentBlockHeaderHashCalled func([]byte)
	GetLocalHeightCalled            func() int64
	SetLocalHeightCalled            func(int64)
	GetNetworkHeightCalled          func() int64
	SetNetworkHeightCalled          func(int64)
	HasBadBlockCalled               func([]byte) bool
	PutBadBlockCalled               func([]byte)
}

// GetGenesisHeader returns the genesis block header pointer
func (bc *BlockChainStub) GetGenesisHeader() data.HeaderHandler {
	if bc.GetGenesisHeaderCalled != nil {
		return bc.GetGenesisHeaderCalled()
	}
	return &block.Header{}
}

// SetGenesisHeader sets the genesis block header pointer
func (bc *BlockChainStub) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if bc.SetGenesisHeaderCalled != nil {
		return bc.SetGenesisHeaderCalled(genesisBlock)
	}
	return nil
}

// GetGenesisHeaderHash returns the genesis block header hash
func (bc *BlockChainStub) GetGenesisHeaderHash() []byte {
	if bc.GetGenesisHeaderHashCalled != nil {
		return bc.GetGenesisHeaderHashCalled()
	}
	return nil
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bc *BlockChainStub) SetGenesisHeaderHash(hash []byte) {
	if bc.SetGenesisHeaderHashCalled != nil {
		bc.SetGenesisHeaderHashCalled(hash)
	}
}

// GetCurrentBlockHeader returns current block header pointer
func (bc *BlockChainStub) GetCurrentBlockHeader() data.HeaderHandler {
	if bc.GetCurrentBlockHeaderCalled != nil {
		return bc.GetCurrentBlockHeaderCalled()
	}
	return nil
}

// SetCurrentBlockHeader sets current block header pointer
func (bc *BlockChainStub) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if bc.SetCurrentBlockHeaderCalled != nil {
		return bc.SetCurrentBlockHeaderCalled(header)
	}
	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChainStub) GetCurrentBlockHeaderHash() []byte {
	if bc.GetCurrentBlockHeaderHashCalled != nil {
		return bc.GetCurrentBlockHeaderHashCalled()
	}
	return nil
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bc *BlockChainStub) SetCurrentBlockHeaderHash(hash []byte) {
	if bc.SetCurrentBlockHeaderHashCalled != nil {
		bc.SetCurrentBlockHeaderHashCalled(hash)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bc *BlockChainStub) IsInterfaceNil() bool {
	return bc == nil
}
