package genericMocks

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
)

// ChainStorerMock -
type ChainStorerMock struct {
	BlockHeaders  *StorerMock
	Metablocks    *StorerMock
	Miniblocks    *StorerMock
	Transactions  *StorerMock
	Rewards       *StorerMock
	Unsigned      *StorerMock
	Logs          *StorerMock
	MetaHdrNonce  *StorerMock
	ShardHdrNonce *StorerMock
	Receipts      *StorerMock
	ScheduledSCRs *StorerMock
	Others        *StorerMock
}

// NewChainStorerMock -
func NewChainStorerMock(epoch uint32) *ChainStorerMock {
	return &ChainStorerMock{
		BlockHeaders:  NewStorerMockWithEpoch(epoch),
		Metablocks:    NewStorerMockWithEpoch(epoch),
		Miniblocks:    NewStorerMockWithEpoch(epoch),
		Transactions:  NewStorerMockWithEpoch(epoch),
		Rewards:       NewStorerMockWithEpoch(epoch),
		Unsigned:      NewStorerMockWithEpoch(epoch),
		Logs:          NewStorerMockWithEpoch(epoch),
		MetaHdrNonce:  NewStorerMockWithEpoch(epoch),
		ShardHdrNonce: NewStorerMockWithEpoch(epoch),
		Receipts:      NewStorerMockWithEpoch(epoch),
		ScheduledSCRs: NewStorerMockWithEpoch(epoch),
		Others:        NewStorerMockWithEpoch(epoch),
	}
}

// CloseAll -
func (sm *ChainStorerMock) CloseAll() error {
	return nil
}

// AddStorer -
func (sm *ChainStorerMock) AddStorer(_ dataRetriever.UnitType, _ storage.Storer) {
	panic("not supported")
}

// GetStorer -
func (sm *ChainStorerMock) GetStorer(unitType dataRetriever.UnitType) (storage.Storer, error) {
	switch unitType {
	case dataRetriever.BlockHeaderUnit:
		return sm.BlockHeaders, nil
	case dataRetriever.MetaBlockUnit:
		return sm.Metablocks, nil
	case dataRetriever.MiniBlockUnit:
		return sm.Miniblocks, nil
	case dataRetriever.TransactionUnit:
		return sm.Transactions, nil
	case dataRetriever.RewardTransactionUnit:
		return sm.Rewards, nil
	case dataRetriever.UnsignedTransactionUnit:
		return sm.Unsigned, nil
	case dataRetriever.TxLogsUnit:
		return sm.Logs, nil
	case dataRetriever.MetaHdrNonceHashDataUnit:
		return sm.MetaHdrNonce, nil
	case dataRetriever.ShardHdrNonceHashDataUnit:
		return sm.ShardHdrNonce, nil
	case dataRetriever.ReceiptsUnit:
		return sm.Receipts, nil
	case dataRetriever.ScheduledSCRsUnit:
		return sm.ScheduledSCRs, nil
	}

	// According to: dataRetriever/interface.go
	if unitType > dataRetriever.ShardHdrNonceHashDataUnit {
		return sm.MetaHdrNonce, nil
	}

	return sm.Others, nil
}

// Has -
func (sm *ChainStorerMock) Has(_ dataRetriever.UnitType, _ []byte) error {
	return nil
}

// Get -
func (sm *ChainStorerMock) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	storer, err := sm.GetStorer(unitType)
	if err != nil {
		return nil, err
	}

	return storer.Get(key)
}

// Put -
func (sm *ChainStorerMock) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	storer, err := sm.GetStorer(unitType)
	if err != nil {
		return err
	}

	return storer.Put(key, value)
}

// GetAll -
func (sm *ChainStorerMock) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	storer, err := sm.GetStorer(unitType)
	if err != nil {
		return nil, err
	}

	data := make(map[string][]byte)
	for _, key := range keys {
		buff, errGet := storer.Get(key)
		if errGet != nil {
			return nil, errGet
		}

		data[string(key)] = buff
	}

	return data, nil
}

// SetEpochForPutOperation -
func (sm *ChainStorerMock) SetEpochForPutOperation(_ uint32) {
}

// GetAllStorers -
func (sm *ChainStorerMock) GetAllStorers() map[dataRetriever.UnitType]storage.Storer {
	return map[dataRetriever.UnitType]storage.Storer{
		dataRetriever.BlockHeaderUnit:           sm.BlockHeaders,
		dataRetriever.MetaBlockUnit:             sm.Metablocks,
		dataRetriever.MiniBlockUnit:             sm.Miniblocks,
		dataRetriever.TransactionUnit:           sm.Transactions,
		dataRetriever.RewardTransactionUnit:     sm.Rewards,
		dataRetriever.UnsignedTransactionUnit:   sm.Unsigned,
		dataRetriever.TxLogsUnit:                sm.Logs,
		dataRetriever.MetaHdrNonceHashDataUnit:  sm.MetaHdrNonce,
		dataRetriever.ShardHdrNonceHashDataUnit: sm.ShardHdrNonce,
		dataRetriever.ReceiptsUnit:              sm.Receipts,
		dataRetriever.ScheduledSCRsUnit:         sm.ScheduledSCRs,
	}
}

// Destroy -
func (sm *ChainStorerMock) Destroy() error {
	return nil
}

// IsInterfaceNil -
func (sm *ChainStorerMock) IsInterfaceNil() bool {
	return sm == nil
}
