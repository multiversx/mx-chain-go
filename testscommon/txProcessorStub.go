package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TxProcessorStub -
type TxProcessorStub struct {
	ProcessTransactionCalled func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	VerifyTransactionCalled  func(tx *transaction.Transaction) error
}

// ProcessTransaction -
func (tps *TxProcessorStub) ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
	if tps.ProcessTransactionCalled != nil {
		return tps.ProcessTransactionCalled(transaction)
	}

	return 0, nil
}

// VerifyTransaction -
func (tps *TxProcessorStub) VerifyTransaction(tx *transaction.Transaction) error {
	if tps.VerifyTransactionCalled != nil {
		return tps.VerifyTransactionCalled(tx)
	}

	return nil
}

// IsInterfaceNil -
func (tps *TxProcessorStub) IsInterfaceNil() bool {
	return tps == nil
}
