package mock

import "github.com/ElrondNetwork/elrond-go/data/transaction"

// TxProcessorStub -
type TxProcessorStub struct {
	ProcessTransactionCalled func(transaction *transaction.Transaction) error
}

// ProcessTransaction -
func (tps *TxProcessorStub) ProcessTransaction(transaction *transaction.Transaction) error {
	if tps.ProcessTransactionCalled != nil {
		return tps.ProcessTransactionCalled(transaction)
	}

	return nil
}

// IsInterfaceNil -
func (tps *TxProcessorStub) IsInterfaceNil() bool {
	return tps == nil
}
