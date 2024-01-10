package mock

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/factory"
)

// DataComponentsStub -
type DataComponentsStub struct {
	BlockChain    data.ChainHandler
	Store         dataRetriever.StorageService
	DataPool      dataRetriever.PoolsHolder
	MbProvider    factory.MiniBlockProvider
	EconomicsData factory.EconomicsHandler
	mutDcm        sync.RWMutex
}

// Create -
func (dcs *DataComponentsStub) Create() error {
	return nil
}

// Close -
func (dcs *DataComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (dcs *DataComponentsStub) CheckSubcomponents() error {
	return nil
}

// Blockchain -
func (dcs *DataComponentsStub) Blockchain() data.ChainHandler {
	dcs.mutDcm.RLock()
	defer dcs.mutDcm.RUnlock()
	return dcs.BlockChain

}

// SetBlockchain -
func (dcs *DataComponentsStub) SetBlockchain(chain data.ChainHandler) error {
	dcs.mutDcm.Lock()
	dcs.BlockChain = chain
	dcs.mutDcm.Unlock()
	return nil
}

// StorageService -
func (dcs *DataComponentsStub) StorageService() dataRetriever.StorageService {
	return dcs.Store
}

// Datapool -
func (dcs *DataComponentsStub) Datapool() dataRetriever.PoolsHolder {
	return dcs.DataPool
}

// MiniBlocksProvider -
func (dcs *DataComponentsStub) MiniBlocksProvider() factory.MiniBlockProvider {
	return dcs.MbProvider
}

// EconomicsHandler -
func (dcs *DataComponentsStub) EconomicsHandler() factory.EconomicsHandler {
	return dcs.EconomicsData
}

// Clone -
func (dcs *DataComponentsStub) Clone() interface{} {
	return &DataComponentsStub{
		BlockChain: dcs.BlockChain,
		Store:      dcs.Store,
		DataPool:   dcs.DataPool,
		MbProvider: dcs.MbProvider,
	}
}

// String -
func (dcs *DataComponentsStub) String() string {
	return "DataComponentsStub"
}

// IsInterfaceNil -
func (dcs *DataComponentsStub) IsInterfaceNil() bool {
	return dcs == nil
}
