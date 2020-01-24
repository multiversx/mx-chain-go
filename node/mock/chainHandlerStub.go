package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type ChainHandlerStub struct {
	GetGenesisHeaderCalled      func() data.HeaderHandler
	GetGenesisHeaderHashCalled  func() []byte
	SetGenesisHeaderCalled      func(gb data.HeaderHandler) error
	SetGenesisHeaderHashCalled  func(hash []byte)
	SetCurrentBlockHeaderCalled func(bh data.HeaderHandler) error
	SetCurrentBlockBodyCalled   func(body data.BodyHandler) error
}

func (chs *ChainHandlerStub) GetGenesisHeader() data.HeaderHandler {
	return chs.GetGenesisHeaderCalled()
}

func (chs *ChainHandlerStub) SetGenesisHeader(gb data.HeaderHandler) error {
	return chs.SetGenesisHeaderCalled(gb)
}

func (chs *ChainHandlerStub) GetGenesisHeaderHash() []byte {
	return chs.GetGenesisHeaderHashCalled()
}

func (chs *ChainHandlerStub) SetGenesisHeaderHash(hash []byte) {
	chs.SetGenesisHeaderHashCalled(hash)
}

func (chs *ChainHandlerStub) GetCurrentBlockHeader() data.HeaderHandler {
	return &block.Header{}
}

func (chs *ChainHandlerStub) SetCurrentBlockHeader(bh data.HeaderHandler) error {
	if chs.SetCurrentBlockHeaderCalled != nil {
		return chs.SetCurrentBlockHeaderCalled(bh)
	}
	return nil
}

func (chs *ChainHandlerStub) GetCurrentBlockHeaderHash() []byte {
	panic("implement me")
}

func (chs *ChainHandlerStub) SetCurrentBlockHeaderHash(hash []byte) {

}

func (chs *ChainHandlerStub) GetCurrentBlockBody() data.BodyHandler {
	panic("implement me")
}

func (chs *ChainHandlerStub) SetCurrentBlockBody(body data.BodyHandler) error {
	if chs.SetCurrentBlockBodyCalled != nil {
		return chs.SetCurrentBlockBodyCalled(body)
	}
	return nil
}

func (chs *ChainHandlerStub) GetLocalHeight() int64 {
	panic("implement me")
}

func (chs *ChainHandlerStub) SetLocalHeight(height int64) {
	panic("implement me")
}

func (chs *ChainHandlerStub) GetNetworkHeight() int64 {
	panic("implement me")
}

func (chs *ChainHandlerStub) SetNetworkHeight(height int64) {
	panic("implement me")
}

func (chs *ChainHandlerStub) HasBadBlock(blockHash []byte) bool {
	panic("implement me")
}

func (chs *ChainHandlerStub) PutBadBlock(blockHash []byte) {
	panic("implement me")
}

// IsInterfaceNil returns true if there is no value under the interface
func (chs *ChainHandlerStub) IsInterfaceNil() bool {
	return chs == nil
}
