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
	SetLastExecutedBlockInfoCalled         func(nonce uint64, headerHash []byte, rootHash []byte)
	GetLastExecutedBlockInfoCalled         func() (uint64, []byte, []byte)
	SetCurrentBlockHeaderCalled            func(header data.HeaderHandler) error
	GetLastExecutedBlockHeaderCalled       func() data.HeaderHandler
	SetLastExecutedBlockHeaderCalled       func(header data.HeaderHandler) error
}

// GetGenesisHeader -
func (stub *ChainHandlerStub) GetGenesisHeader() data.HeaderHandler {
	if stub.GetGenesisHeaderCalled != nil {
		return stub.GetGenesisHeaderCalled()
	}

	return &HeaderHandlerStub{}
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

// SetLastExecutedBlockInfo -
func (stub *ChainHandlerStub) SetLastExecutedBlockInfo(nonce uint64, headerHash []byte, rootHash []byte) {
	if stub.SetLastExecutedBlockInfoCalled != nil {
		stub.SetLastExecutedBlockInfoCalled(nonce, headerHash, rootHash)
	}
}

// GetFinalBlockInfo -
func (stub *ChainHandlerStub) GetFinalBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	if stub.GetFinalBlockInfoCalled != nil {
		return stub.GetFinalBlockInfoCalled()
	}

	return 0, nil, nil
}

// GetLastExecutedBlockInfo -
func (stub *ChainHandlerStub) GetLastExecutedBlockInfo() (nonce uint64, blockHash []byte, rootHash []byte) {
	if stub.GetLastExecutedBlockInfoCalled != nil {
		return stub.GetLastExecutedBlockInfoCalled()
	}

	return 0, nil, nil
}

// SetCurrentBlockHeader -
func (stub *ChainHandlerStub) SetCurrentBlockHeader(header data.HeaderHandler) error {
	if stub.SetCurrentBlockHeaderCalled != nil {
		return stub.SetCurrentBlockHeaderCalled(header)
	}

	return nil
}

// GetLastExecutedBlockHeader -
func (stub *ChainHandlerStub) GetLastExecutedBlockHeader() data.HeaderHandler {
	if stub.GetLastExecutedBlockHeaderCalled != nil {
		return stub.GetLastExecutedBlockHeaderCalled()
	}

	return nil
}

// SetLastExecutedBlockHeader -
func (stub *ChainHandlerStub) SetLastExecutedBlockHeader(header data.HeaderHandler) error {
	if stub.SetLastExecutedBlockHeaderCalled != nil {
		return stub.SetLastExecutedBlockHeaderCalled(header)
	}

	return nil
}

// IsInterfaceNil -
func (stub *ChainHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
