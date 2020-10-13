package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/factory"
)

// DataComponentsMock -
type DataComponentsMock struct {
	Storage           dataRetriever.StorageService
	Blkc              data.ChainHandler
	DataPool          dataRetriever.PoolsHolder
	MiniBlockProvider factory.MiniBlockProvider
	EconomicsData     factory.EconomicsHandler
}

// StorageService -
func (dcm *DataComponentsMock) StorageService() dataRetriever.StorageService {
	return dcm.Storage
}

// Blockchain -
func (dcm *DataComponentsMock) Blockchain() data.ChainHandler {
	return dcm.Blkc
}

// Clone -
func (dcm *DataComponentsMock) Clone() interface{} {
	return &DataComponentsMock{
		Storage:  dcm.StorageService(),
		Blkc:     dcm.Blockchain(),
		DataPool: dcm.Datapool(),
	}
}

// Datapool -
func (dcm *DataComponentsMock) Datapool() dataRetriever.PoolsHolder {
	return dcm.DataPool
}

// MiniBlocksProvider -
func (dcm *DataComponentsMock) MiniBlocksProvider() factory.MiniBlockProvider {
	return dcm.MiniBlockProvider
}

// EconomicsHandler -
func (dcm *DataComponentsMock) EconomicsHandler() factory.EconomicsHandler {
	return dcm.EconomicsData
}

// SetBlockchain -
func (dcm *DataComponentsMock) SetBlockchain(chain data.ChainHandler) {
	dcm.Blkc = chain
}

// IsInterfaceNil -
func (dcm *DataComponentsMock) IsInterfaceNil() bool {
	return dcm == nil
}
