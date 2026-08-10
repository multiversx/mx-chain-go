package mbSelection

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

// MiniBlockSelectionSessionStub -
type MiniBlockSelectionSessionStub struct {
	ResetSelectionSessionCalled                 func()
	GetMiniBlockHeaderHandlersCalled            func() []data.MiniBlockHeaderHandler
	GetMiniBlocksCalled                         func() block.MiniBlockSlice
	GetMiniBlockHashesCalled                    func() [][]byte
	AddReferencedHeaderCalled                   func(metaBlock data.HeaderHandler, metaBlockHash []byte)
	GetReferencedHeaderHashesCalled             func() [][]byte
	GetReferencedHeadersCalled                  func() []data.HeaderHandler
	GetLastHeaderCalled                         func() data.HeaderHandler
	GetGasProvidedCalled                        func() uint64
	GetNumTxsAddedCalled                        func() uint32
	AddMiniBlocksAndHashesCalled                func(miniBlocksAndHashes []block.MiniblockAndHash) error
	CreateAndAddMiniBlockFromTransactionsCalled func(txHashes [][]byte) error
}

// ResetSelectionSession -
func (mbss *MiniBlockSelectionSessionStub) ResetSelectionSession() {
	if mbss.ResetSelectionSessionCalled != nil {
		mbss.ResetSelectionSessionCalled()
	}
}

// GetMiniBlockHeaderHandlers -
func (mbss *MiniBlockSelectionSessionStub) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if mbss.GetMiniBlockHeaderHandlersCalled != nil {
		return mbss.GetMiniBlockHeaderHandlersCalled()
	}
	return nil
}

// GetMiniBlocks -
func (mbss *MiniBlockSelectionSessionStub) GetMiniBlocks() block.MiniBlockSlice {
	if mbss.GetMiniBlocksCalled != nil {
		return mbss.GetMiniBlocksCalled()
	}
	return nil
}

// GetMiniBlockHashes -
func (mbss *MiniBlockSelectionSessionStub) GetMiniBlockHashes() [][]byte {
	if mbss.GetMiniBlockHashesCalled != nil {
		return mbss.GetMiniBlockHashesCalled()
	}
	return nil
}

// AddReferencedHeader -
func (mbss *MiniBlockSelectionSessionStub) AddReferencedHeader(metaBlock data.HeaderHandler, metaBlockHash []byte) {
	if mbss.AddReferencedHeaderCalled != nil {
		mbss.AddReferencedHeaderCalled(metaBlock, metaBlockHash)
	}
}

// GetReferencedHeaderHashes -
func (mbss *MiniBlockSelectionSessionStub) GetReferencedHeaderHashes() [][]byte {
	if mbss.GetReferencedHeaderHashesCalled != nil {
		return mbss.GetReferencedHeaderHashesCalled()
	}
	return nil
}

// GetReferencedHeaders -
func (mbss *MiniBlockSelectionSessionStub) GetReferencedHeaders() []data.HeaderHandler {
	if mbss.GetReferencedHeadersCalled != nil {
		return mbss.GetReferencedHeadersCalled()
	}
	return nil
}

// GetLastHeader -
func (mbss *MiniBlockSelectionSessionStub) GetLastHeader() data.HeaderHandler {
	if mbss.GetLastHeaderCalled != nil {
		return mbss.GetLastHeaderCalled()
	}
	return nil
}

// GetGasProvided -
func (mbss *MiniBlockSelectionSessionStub) GetGasProvided() uint64 {
	if mbss.GetGasProvidedCalled != nil {
		return mbss.GetGasProvidedCalled()
	}
	return 0
}

// GetNumTxsAdded -
func (mbss *MiniBlockSelectionSessionStub) GetNumTxsAdded() uint32 {
	if mbss.GetNumTxsAddedCalled != nil {
		return mbss.GetNumTxsAddedCalled()
	}
	return 0
}

// AddMiniBlocksAndHashes -
func (mbss *MiniBlockSelectionSessionStub) AddMiniBlocksAndHashes(miniBlocksAndHashes []block.MiniblockAndHash) error {
	if mbss.AddMiniBlocksAndHashesCalled != nil {
		return mbss.AddMiniBlocksAndHashesCalled(miniBlocksAndHashes)
	}
	return nil
}

// CreateAndAddMiniBlockFromTransactions -
func (mbss *MiniBlockSelectionSessionStub) CreateAndAddMiniBlockFromTransactions(txHashes [][]byte) error {
	if mbss.CreateAndAddMiniBlockFromTransactionsCalled != nil {
		return mbss.CreateAndAddMiniBlockFromTransactionsCalled(txHashes)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbss *MiniBlockSelectionSessionStub) IsInterfaceNil() bool {
	return mbss == nil
}
