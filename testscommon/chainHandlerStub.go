package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ChainHandlerStub -
type ChainHandlerStub struct {
	GetGenesisHeaderCalled                 func() data.HeaderHandler
	SetGenesisHeaderCalled                 func(handler data.HeaderHandler) error
	GetGenesisHeaderHashCalled             func() []byte
	SetGenesisHeaderHashCalled             func([]byte)
	GetCurrentBlockHeaderCalled            func() data.HeaderHandler
	SetCurrentBlockHeaderAndRootHashCalled func(header data.HeaderHandler, rootHash []byte) error
	GetCurrentBlockHeaderHashCalled        func() []byte
	SetCurrentBlockHeaderHashCalled        func([]byte)
	GetCurrentBlockRootHashCalled          func() []byte
	SetFinalBlockInfoCalled                func(nonce uint64, headerHash []byte, rootHash []byte)
	GetFinalBlockInfoCalled                func() (nonce uint64, blockHash []byte, rootHash []byte)
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

// SetCurrentBlockHeaderAndRootHash -
func (stub *ChainHandlerStub) SetCurrentBlockHeaderAndRootHash(header data.HeaderHandler, rootHash []byte) error {
	if stub.SetCurrentBlockHeaderAndRootHashCalled != nil {
		return stub.SetCurrentBlockHeaderAndRootHashCalled(header, rootHash)
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

// GetCurrentBlockRootHash -
func (stub *ChainHandlerStub) GetCurrentBlockRootHash() []byte {
	if stub.GetCurrentBlockRootHashCalled != nil {
		return stub.GetCurrentBlockRootHashCalled()
	}
	return nil
}

// SetFinalBlockInfo -
func (stub *ChainHandlerStub) SetFinalBlockInfo(nonce uint64, headerHash []byte, rootHash []byte) {
	if stub.SetFinalBlockInfoCalled != nil {
		stub.SetFinalBlockInfoCalled(nonce, headerHash, rootHash)
	}
}

// GetFinalBlockInfo -
func (stub *ChainHandlerStub) GetFinalBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	if stub.GetFinalBlockInfoCalled != nil {
		return stub.GetFinalBlockInfoCalled()
	}

	return 0, nil, nil
}

// IsInterfaceNil -
func (stub *ChainHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
