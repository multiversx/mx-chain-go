package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// TxVersionCheckerStub -
type TxVersionCheckerStub struct {
	IsSignedWithHashCalled     func(tx *transaction.Transaction) bool
	IsGuardedTransactionCalled func(tx *transaction.Transaction) bool
	CheckTxVersionCalled       func(tx *transaction.Transaction) error
}

// IsSignedWithHash will return true if transaction is signed with hash
func (tvcs *TxVersionCheckerStub) IsSignedWithHash(tx *transaction.Transaction) bool {
	if tvcs.IsSignedWithHashCalled != nil {
		return tvcs.IsSignedWithHashCalled(tx)
	}
	return false
}

// IsGuardedTransaction will return true if transaction also holds a guardian signature
func (tvcs *TxVersionCheckerStub) IsGuardedTransaction(tx *transaction.Transaction) bool {
	if tvcs.IsGuardedTransactionCalled != nil {
		return tvcs.IsGuardedTransactionCalled(tx)
	}
	return false
}

// CheckTxVersion will check transaction version
func (tvcs *TxVersionCheckerStub) CheckTxVersion(tx *transaction.Transaction) error {
	if tvcs.CheckTxVersionCalled != nil {
		return tvcs.CheckTxVersionCalled(tx)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tvcs *TxVersionCheckerStub) IsInterfaceNil() bool {
	return tvcs == nil
}
