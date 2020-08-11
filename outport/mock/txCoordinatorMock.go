package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type mapOfTxs = map[string]data.TransactionHandler

// TxCoordinatorMock -
type TxCoordinatorMock struct {
	txsByBlockType map[block.Type]mapOfTxs
}

// NewTxCoordinatorMock -
func NewTxCoordinatorMock() *TxCoordinatorMock {
	return &TxCoordinatorMock{
		txsByBlockType: map[block.Type]mapOfTxs{
			block.TxBlock:                  make(mapOfTxs),
			block.StateBlock:               make(mapOfTxs),
			block.SmartContractResultBlock: make(mapOfTxs),
			block.InvalidBlock:             make(mapOfTxs),
			block.ReceiptBlock:             make(mapOfTxs),
			block.RewardsBlock:             make(mapOfTxs),
		},
	}
}

// GetAllCurrentUsedTxs -
func (m *TxCoordinatorMock) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	txs, _ := m.txsByBlockType[blockType]
	return txs
}

// AddTx -
func (m *TxCoordinatorMock) AddTx(blockType block.Type, hash string, tx data.TransactionHandler) {
	m.txsByBlockType[blockType][hash] = tx
}

// IsInterfaceNil -
func (m *TxCoordinatorMock) IsInterfaceNil() bool {
	return m == nil
}
