package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/factory"
)

// DataComponentsMock -
type DataComponentsMock struct {
	BlockChain    data.ChainHandler
	Store         dataRetriever.StorageService
	DataPool      dataRetriever.PoolsHolder
	MbProvider    factory.MiniBlockProvider
	EconomicsData factory.EconomicsHandler
	mutDcm        sync.RWMutex
}

// Create -
func (dcm *DataComponentsMock) Create() error {
	return nil
}

// Close -
func (dcm *DataComponentsMock) Close() error {
	return nil
}

// CheckSubcomponents -
func (dcm *DataComponentsMock) CheckSubcomponents() error {
	return nil
}

// Blockchain -
func (dcm *DataComponentsMock) Blockchain() data.ChainHandler {
	dcm.mutDcm.RLock()
	defer dcm.mutDcm.RUnlock()
	return dcm.BlockChain

}

// SetBlockchain -
func (dcm *DataComponentsMock) SetBlockchain(chain data.ChainHandler) {
	dcm.mutDcm.Lock()
	dcm.BlockChain = chain
	dcm.mutDcm.Unlock()
}

// StorageService -
func (dcm *DataComponentsMock) StorageService() dataRetriever.StorageService {
	return dcm.Store
}

// Datapool -
func (dcm *DataComponentsMock) Datapool() dataRetriever.PoolsHolder {
	return dcm.DataPool
}

// MiniBlocksProvider -
func (dcm *DataComponentsMock) MiniBlocksProvider() factory.MiniBlockProvider {
	return dcm.MbProvider
}

// EconomicsHandler -
func (dcm *DataComponentsMock) EconomicsHandler() factory.EconomicsHandler {
	return dcm.EconomicsData
}

// Clone -
func (dcm *DataComponentsMock) Clone() interface{} {
	return &DataComponentsMock{
		BlockChain: dcm.BlockChain,
		Store:      dcm.Store,
		DataPool:   dcm.DataPool,
		MbProvider: dcm.MbProvider,
	}
}

// IsInterfaceNil -
func (dcm *DataComponentsMock) IsInterfaceNil() bool {
	return dcm == nil
}
