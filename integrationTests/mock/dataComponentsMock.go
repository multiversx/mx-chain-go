package mock

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// MiniBlockProvider defines what a miniblock data provider should do
type MiniBlockProvider interface {
	GetMiniBlocks(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	GetMiniBlocksFromPool(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte)
	IsInterfaceNil() bool
}

type DataComponentsMock struct {
	Storage           dataRetriever.StorageService
	DataPool          dataRetriever.PoolsHolder
	BlockChain        data.ChainHandler
	MiniBlockProvider MiniBlockProvider
	mutBlockchain     sync.RWMutex
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

// MiniBlocksProvider -
func (dcm *DataComponentsMock) MiniBlocksProvider() MiniBlockProvider {
	dcm.mutBlockchain.RLock()
	defer dcm.mutBlockchain.RUnlock()

	return dcm.MiniBlockProvider
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
