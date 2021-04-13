package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// ChainHandlerStub -
type ChainHandlerStub struct {
	GetGenesisHeaderCalled      func() data.HeaderHandler
	GetGenesisHeaderHashCalled  func() []byte
	SetGenesisHeaderCalled      func(gb data.HeaderHandler) error
	SetGenesisHeaderHashCalled  func(hash []byte)
	SetCurrentBlockHeaderCalled func(bh data.HeaderHandler) error
	GetCurrentBlockHeaderCalled func() data.HeaderHandler
	CreateNewHeaderCalled       func() data.HeaderHandler
}

// GetGenesisHeader -
func (chs *ChainHandlerStub) GetGenesisHeader() data.HeaderHandler {
	return chs.GetGenesisHeaderCalled()
}

// SetGenesisHeader -
func (chs *ChainHandlerStub) SetGenesisHeader(gb data.HeaderHandler) error {
	return chs.SetGenesisHeaderCalled(gb)
}

// GetGenesisHeaderHash -
func (chs *ChainHandlerStub) GetGenesisHeaderHash() []byte {
	return chs.GetGenesisHeaderHashCalled()
}

// SetGenesisHeaderHash -
func (chs *ChainHandlerStub) SetGenesisHeaderHash(hash []byte) {
	chs.SetGenesisHeaderHashCalled(hash)
}

// GetCurrentBlockHeader -
func (chs *ChainHandlerStub) GetCurrentBlockHeader() data.HeaderHandler {
	if chs.GetCurrentBlockHeaderCalled != nil {
		return chs.GetCurrentBlockHeaderCalled()
	}

	return &block.Header{}
}

// SetCurrentBlockHeader -
func (chs *ChainHandlerStub) SetCurrentBlockHeader(bh data.HeaderHandler) error {
	if chs.SetCurrentBlockHeaderCalled != nil {
		return chs.SetCurrentBlockHeaderCalled(bh)
	}
	return nil
}

// GetCurrentBlockHeaderHash -
func (chs *ChainHandlerStub) GetCurrentBlockHeaderHash() []byte {
	panic("implement me")
}

// SetCurrentBlockHeaderHash -
func (chs *ChainHandlerStub) SetCurrentBlockHeaderHash(_ []byte) {

}

// IsInterfaceNil returns true if there is no value under the interface
func (chs *ChainHandlerStub) IsInterfaceNil() bool {
	return chs == nil
}

// CreateNewHeader -
func (chs *ChainHandlerStub) CreateNewHeader() data.HeaderHandler {
	if chs.CreateNewHeaderCalled != nil {
		return chs.CreateNewHeaderCalled()
	}

	return nil
}
