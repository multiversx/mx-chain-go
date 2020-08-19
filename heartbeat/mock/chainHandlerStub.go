package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// ChainHandlerStub -
type ChainHandlerStub struct {
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
	CreateNewHeaderCalled           func() data.HeaderHandler
}

// GetGenesisHeader -
func (chs *ChainHandlerStub) GetGenesisHeader() data.HeaderHandler {
	if chs.GetGenesisHeaderCalled != nil {
		return chs.GetGenesisHeaderCalled()
	}
	return &block.Header{}
}

// SetGenesisHeader -
func (chs *ChainHandlerStub) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if chs.SetGenesisHeaderCalled != nil {
		return chs.SetGenesisHeaderCalled(genesisBlock)
	}
	return nil
}

// GetGenesisHeaderHash -
func (chs *ChainHandlerStub) GetGenesisHeaderHash() []byte {
	if chs.GetGenesisHeaderHashCalled != nil {
		return chs.GetGenesisHeaderHashCalled()
	}
	return nil
}

// SetGenesisHeaderHash -
func (chs *ChainHandlerStub) SetGenesisHeaderHash(hash []byte) {
	if chs.SetGenesisHeaderHashCalled != nil {
		chs.SetGenesisHeaderHashCalled(hash)
	}
}

// GetCurrentBlockHeader -
func (chs *ChainHandlerStub) GetCurrentBlockHeader() data.HeaderHandler {
	if chs.GetCurrentBlockHeaderCalled != nil {
		return chs.GetCurrentBlockHeaderCalled()
	}
	return nil
}

// SetCurrentBlockHeader -
func (chs *ChainHandlerStub) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if chs.SetCurrentBlockHeaderCalled != nil {
		return chs.SetCurrentBlockHeaderCalled(header)
	}
	return nil
}

// GetCurrentBlockHeaderHash -
func (chs *ChainHandlerStub) GetCurrentBlockHeaderHash() []byte {
	if chs.GetCurrentBlockHeaderHashCalled != nil {
		return chs.GetCurrentBlockHeaderHashCalled()
	}
	return nil
}

// SetCurrentBlockHeaderHash -
func (chs *ChainHandlerStub) SetCurrentBlockHeaderHash(hash []byte) {
	if chs.SetCurrentBlockHeaderHashCalled != nil {
		chs.SetCurrentBlockHeaderHashCalled(hash)
	}
}

// IsInterfaceNil -
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
