package mock

import (
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/update"
	"github.com/multiversx/mx-chain-go/update/mock"

	"github.com/multiversx/mx-chain-core-go/data"
)

// GenesisBlockCreatorMock -
type GenesisBlockCreatorMock struct {
	ImportHandlerCalled       func() update.ImportHandler
	CreateGenesisBlocksCalled func() (map[uint32]data.HeaderHandler, error)
	GetIndexingDataCalled     func() map[uint32]*genesis.IndexingData
}

// ImportHandler -
func (gbc *GenesisBlockCreatorMock) ImportHandler() update.ImportHandler {
	return &mock.ImportHandlerStub{}
}

// CreateGenesisBlocks -
func (gbc *GenesisBlockCreatorMock) CreateGenesisBlocks() (map[uint32]data.HeaderHandler, error) {
	return nil, nil
}

// GetIndexingData -
func (gbc *GenesisBlockCreatorMock) GetIndexingData() map[uint32]*genesis.IndexingData {
	return nil
}
