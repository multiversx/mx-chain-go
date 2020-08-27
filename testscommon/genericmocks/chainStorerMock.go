package genericmocks

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ChainStorerMock -
type ChainStorerMock struct {
	Transactions *StorerMock
	Rewards      *StorerMock
	Unsigned     *StorerMock
}

// NewChainStorerMock -
func NewChainStorerMock() *ChainStorerMock {
	return &ChainStorerMock{
		Transactions: NewStorerMock("Transactions", 0),
		Rewards:      NewStorerMock("Rewards", 0),
		Unsigned:     NewStorerMock("Unsigned", 0),
	}
}

// CloseAll -
func (sm *ChainStorerMock) CloseAll() error {
	return nil
}

// AddStorer -
func (sm *ChainStorerMock) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	panic("not supported")
}

// GetStorer -
func (sm *ChainStorerMock) GetStorer(unitType dataRetriever.UnitType) storage.Storer {
	if unitType == dataRetriever.TransactionUnit {
		return sm.Transactions
	}
	if unitType == dataRetriever.RewardTransactionUnit {
		return sm.Rewards
	}
	if unitType == dataRetriever.UnsignedTransactionUnit {
		return sm.Unsigned
	}

	panic("storer missing, add it")
}

// Has -
func (sm *ChainStorerMock) Has(unitType dataRetriever.UnitType, key []byte) error {
	return nil
}

// Get -
func (sm *ChainStorerMock) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	return sm.GetStorer(unitType).Get(key)
}

// Put -
func (sm *ChainStorerMock) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	return sm.GetStorer(unitType).Put(key, value)
}

// GetAll -
func (sm *ChainStorerMock) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	panic("not supported")
}

// SetEpochForPutOperation -
func (sm *ChainStorerMock) SetEpochForPutOperation(epoch uint32) {
	panic("not supported")
}

// Destroy -
func (sm *ChainStorerMock) Destroy() error {
	return nil
}

// IsInterfaceNil -
func (sm *ChainStorerMock) IsInterfaceNil() bool {
	return sm == nil
}
