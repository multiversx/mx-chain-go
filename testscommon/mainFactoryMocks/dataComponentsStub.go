package mainFactoryMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
)

// DataComponentsHolderStub -
type DataComponentsHolderStub struct {
	BlockchainCalled         func() data.ChainHandler
	SetBlockchainCalled      func(chain data.ChainHandler)
	StorageServiceCalled     func() dataRetriever.StorageService
	DatapoolCalled           func() dataRetriever.PoolsHolder
	MiniBlocksProviderCalled func() factory.MiniBlockProvider
	CloneCalled              func() interface{}
}

// Blockchain -
func (dchs *DataComponentsHolderStub) Blockchain() data.ChainHandler {
	if dchs.BlockchainCalled != nil {
		return dchs.BlockchainCalled()
	}
	return nil
}

// SetBlockchain -
func (dchs *DataComponentsHolderStub) SetBlockchain(chain data.ChainHandler) {
	if dchs.SetBlockchainCalled != nil {
		dchs.SetBlockchainCalled(chain)
	}
}

// StorageService -
func (dchs *DataComponentsHolderStub) StorageService() dataRetriever.StorageService {
	if dchs.StorageServiceCalled != nil {
		return dchs.StorageServiceCalled()
	}
	return nil
}

// Datapool -
func (dchs *DataComponentsHolderStub) Datapool() dataRetriever.PoolsHolder {
	if dchs.DatapoolCalled != nil {
		return dchs.DatapoolCalled()
	}
	return nil
}

// MiniBlocksProvider -
func (dchs *DataComponentsHolderStub) MiniBlocksProvider() factory.MiniBlockProvider {
	if dchs.MiniBlocksProviderCalled != nil {
		return dchs.MiniBlocksProviderCalled()
	}
	return nil
}

// Clone -
func (dchs *DataComponentsHolderStub) Clone() interface{} {
	if dchs.CloneCalled != nil {
		return dchs.CloneCalled()
	}
	return nil
}

// IsInterfaceNil -
func (dchs *DataComponentsHolderStub) IsInterfaceNil() bool {
	return dchs == nil
}
