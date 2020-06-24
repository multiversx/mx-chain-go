package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type DataComponentsMock struct {
	Storage       dataRetriever.StorageService
	DataPool      dataRetriever.PoolsHolder
	BlockChain    data.ChainHandler
	mutBlockchain sync.RWMutex
}

// StorageService -
func (dcm *DataComponentsMock) StorageService() dataRetriever.StorageService {
	return dcm.Storage
}

// Datapool -
func (dcm *DataComponentsMock) Datapool() dataRetriever.PoolsHolder {
	return dcm.DataPool
}

// Blockchain -
func (dcm *DataComponentsMock) Blockchain() data.ChainHandler {
	dcm.mutBlockchain.RLock()
	defer dcm.mutBlockchain.RUnlock()

	return dcm.BlockChain
}

// Clone -
func (dcm *DataComponentsMock) Clone() interface{} {
	return &DataComponentsMock{
		Storage:       dcm.Storage,
		DataPool:      dcm.DataPool,
		BlockChain:    dcm.BlockChain,
		mutBlockchain: sync.RWMutex{},
	}
}

// SetBlockchain -
func (dcm *DataComponentsMock) SetBlockchain(chain data.ChainHandler) {
	dcm.mutBlockchain.Lock()
	dcm.BlockChain = chain
	dcm.mutBlockchain.Unlock()
}

// IsInterfaceNil -
func (dcm *DataComponentsMock) IsInterfaceNil() bool {
	return dcm == nil
}
