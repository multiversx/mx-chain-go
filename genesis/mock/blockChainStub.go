package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
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
	CreateNewHeaderCalled           func() data.HeaderHandler
}

// GetGenesisHeader returns the genesis block header pointer
func (bcs *BlockChainStub) GetGenesisHeader() data.HeaderHandler {
	if bcs.GetGenesisHeaderCalled != nil {
		return bcs.GetGenesisHeaderCalled()
	}
	return nil
}

// SetGenesisHeader sets the genesis block header pointer
func (bcs *BlockChainStub) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if bcs.SetGenesisHeaderCalled != nil {
		return bcs.SetGenesisHeaderCalled(genesisBlock)
	}
	return nil
}

// GetGenesisHeaderHash returns the genesis block header hash
func (bcs *BlockChainStub) GetGenesisHeaderHash() []byte {
	if bcs.GetGenesisHeaderHashCalled != nil {
		return bcs.GetGenesisHeaderHashCalled()
	}
	return nil
}

// SetGenesisHeaderHash sets the genesis block header hash
func (bcs *BlockChainStub) SetGenesisHeaderHash(hash []byte) {
	if bcs.SetGenesisHeaderHashCalled != nil {
		bcs.SetGenesisHeaderHashCalled(hash)
	}
}

// GetCurrentBlockHeader returns current block header pointer
func (bcs *BlockChainStub) GetCurrentBlockHeader() data.HeaderHandler {
	if bcs.GetCurrentBlockHeaderCalled != nil {
		return bcs.GetCurrentBlockHeaderCalled()
	}
	return nil
}

// SetCurrentBlockHeader sets current block header pointer
func (bcs *BlockChainStub) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if bcs.SetCurrentBlockHeaderCalled != nil {
		return bcs.SetCurrentBlockHeaderCalled(header)
	}
	return nil
}

// GetCurrentBlockHeaderHash returns the current block header hash
func (bcs *BlockChainStub) GetCurrentBlockHeaderHash() []byte {
	if bcs.GetCurrentBlockHeaderHashCalled != nil {
		return bcs.GetCurrentBlockHeaderHashCalled()
	}
	return nil
}

// SetCurrentBlockHeaderHash returns the current block header hash
func (bcs *BlockChainStub) SetCurrentBlockHeaderHash(hash []byte) {
	if bcs.SetCurrentBlockHeaderHashCalled != nil {
		bcs.SetCurrentBlockHeaderHashCalled(hash)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bcs *BlockChainStub) IsInterfaceNil() bool {
	return bcs == nil
}

// CreateNewHeader -
func (bcs *BlockChainStub) CreateNewHeader() data.HeaderHandler {
	if bcs.CreateNewHeaderCalled != nil {
		return bcs.CreateNewHeaderCalled()
	}

	return nil
}
