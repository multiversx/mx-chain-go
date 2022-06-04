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
	SetCurrentBlockHeaderAndRootHashCalled func(header data.HeaderHandler, rootHash []byte) error
	GetCurrentBlockHeaderHashCalled        func() []byte
	SetCurrentBlockHeaderHashCalled        func([]byte)
	GetCurrentBlockRootHashCalled          func() []byte
	SetFinalBlockInfoCalled                func(nonce uint64, headerHash []byte, rootHash []byte)
	GetFinalBlockInfoCalled                func() (nonce uint64, blockHash []byte, rootHash []byte)

	currentBlockHeader   data.HeaderHandler
	currentBlockHash     []byte
	currentBlockRootHash []byte

	finalBlockNonce    uint64
	finalBlockHash     []byte
	finalBlockRootHash []byte
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
	return stub.currentBlockHeader
}

// SetCurrentBlockHeaderAndRootHash -
func (stub *ChainHandlerStub) SetCurrentBlockHeaderAndRootHash(header data.HeaderHandler, rootHash []byte) error {
	if stub.SetCurrentBlockHeaderAndRootHashCalled != nil {
		return stub.SetCurrentBlockHeaderAndRootHashCalled(header, rootHash)
	}

	stub.currentBlockHeader = header
	stub.currentBlockRootHash = rootHash
	return nil
}

// GetCurrentBlockHeaderHash -
func (stub *ChainHandlerStub) GetCurrentBlockHeaderHash() []byte {
	if stub.GetCurrentBlockHeaderHashCalled != nil {
		return stub.GetCurrentBlockHeaderHashCalled()
	}
	return stub.currentBlockHash
}

// SetCurrentBlockHeaderHash -
func (stub *ChainHandlerStub) SetCurrentBlockHeaderHash(hash []byte) {
	if stub.SetCurrentBlockHeaderHashCalled != nil {
		stub.SetCurrentBlockHeaderHashCalled(hash)
	}

	stub.currentBlockHash = hash
}

// GetCurrentBlockRootHash -
func (stub *ChainHandlerStub) GetCurrentBlockRootHash() []byte {
	if stub.GetCurrentBlockRootHashCalled != nil {
		return stub.GetCurrentBlockRootHashCalled()
	}
	return stub.currentBlockRootHash
}

// SetFinalBlockInfoCalled -
func (stub *ChainHandlerStub) SetFinalBlockInfo(nonce uint64, headerHash []byte, rootHash []byte) {
	if stub.SetFinalBlockInfoCalled != nil {
		stub.SetFinalBlockInfoCalled(nonce, headerHash, rootHash)
	}

	stub.finalBlockNonce = nonce
	stub.finalBlockHash = headerHash
	stub.finalBlockRootHash = rootHash
}

// GetFinalBlockInfo -
func (stub *ChainHandlerStub) GetFinalBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	if stub.GetFinalBlockInfoCalled != nil {
		return stub.GetFinalBlockInfoCalled()
	}

	return stub.finalBlockNonce, stub.finalBlockHash, stub.finalBlockRootHash
}

// IsInterfaceNil -
func (stub *ChainHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
