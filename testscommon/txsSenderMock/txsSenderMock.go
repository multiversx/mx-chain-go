package txsSenderMock

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// TxsSenderHandlerMock -
type TxsSenderHandlerMock struct {
	SendBulkTransactionsCalled func(txs []*transaction.Transaction) (uint64, error)
}

// SendBulkTransactions -
func (tsm *TxsSenderHandlerMock) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	if tsm.SendBulkTransactionsCalled != nil {
		return tsm.SendBulkTransactionsCalled(txs)
	}
	return 0, nil
}

// Close -
func (tsm *TxsSenderHandlerMock) Close() error {
	return nil
}

// IsInterfaceNil -
func (tsm *TxsSenderHandlerMock) IsInterfaceNil() bool {
	return tsm == nil
}
