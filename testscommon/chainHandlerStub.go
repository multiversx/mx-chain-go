package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// ChainHandlerStub -
type ChainHandlerStub struct {
	GetGenesisHeaderCalled                 func() data.HeaderHandler
	SetGenesisHeaderCalled                 func(handler data.HeaderHandler) error
	GetGenesisHeaderHashCalled             func() []byte
	SetGenesisHeaderHashCalled             func([]byte)
	GetCurrentBlockHeaderCalled            func() data.HeaderHandler
	SetCurrentBlockHeaderCalled            func(data.HeaderHandler) error
	GetCurrentBlockHeaderHashCalled        func() []byte
	SetCurrentBlockHeaderHashCalled        func([]byte)
	GetCurrentBlockCommittedRootHashCalled func() []byte
}

// GetGenesisHeader -
func (stub *ChainHandlerStub) GetGenesisHeader() data.HeaderHandler {
	if stub.GetGenesisHeaderCalled != nil {
		return stub.GetGenesisHeaderCalled()
	}
	return nil
}

// SetGenesisHeader -
func (stub *ChainHandlerStub) SetGenesisHeader(genesisBlock data.HeaderHandler) error {
	if stub.SetGenesisHeaderCalled != nil {
		return stub.SetGenesisHeaderCalled(genesisBlock)
	}
	return nil
}

// GetGenesisHeaderHash -
func (stub *ChainHandlerStub) GetGenesisHeaderHash() []byte {
	if stub.GetGenesisHeaderHashCalled != nil {
		return stub.GetGenesisHeaderHashCalled()
	}
	return nil
}

// SetGenesisHeaderHash -
func (stub *ChainHandlerStub) SetGenesisHeaderHash(hash []byte) {
	if stub.SetGenesisHeaderHashCalled != nil {
		stub.SetGenesisHeaderHashCalled(hash)
	}
}

// GetCurrentBlockHeader -
func (stub *ChainHandlerStub) GetCurrentBlockHeader() data.HeaderHandler {
	if stub.GetCurrentBlockHeaderCalled != nil {
		return stub.GetCurrentBlockHeaderCalled()
	}
	return nil
}

// SetCurrentBlockHeader -
func (stub *ChainHandlerStub) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if stub.SetCurrentBlockHeaderCalled != nil {
		return stub.SetCurrentBlockHeaderCalled(header)
	}
	return nil
}

// GetCurrentBlockHeaderHash -
func (stub *ChainHandlerStub) GetCurrentBlockHeaderHash() []byte {
	if stub.GetCurrentBlockHeaderHashCalled != nil {
		return stub.GetCurrentBlockHeaderHashCalled()
	}
	return nil
}

// SetCurrentBlockHeaderHash -
func (stub *ChainHandlerStub) SetCurrentBlockHeaderHash(hash []byte) {
	if stub.SetCurrentBlockHeaderHashCalled != nil {
		stub.SetCurrentBlockHeaderHashCalled(hash)
	}
}

// GetCurrentBlockCommittedRootHash -
func (stub *ChainHandlerStub) GetCurrentBlockCommittedRootHash() []byte {
	if stub.GetCurrentBlockCommittedRootHashCalled != nil {
		return stub.GetCurrentBlockCommittedRootHashCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *ChainHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
