package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

// TxProcessorStub -
type TxProcessorStub struct {
	ProcessTransactionCalled              func(transaction *transaction.Transaction) (vmcommon.ReturnCode, error)
	RegisterUnExecutableTransactionCalled func(transaction *transaction.Transaction, txHash []byte) error
	VerifyTransactionCalled               func(tx *transaction.Transaction) error
	VerifyGuardianCalled                  func(tx *transaction.Transaction, account state.UserAccountHandler) error
}

// ProcessTransaction -
func (tps *TxProcessorStub) ProcessTransaction(transaction *transaction.Transaction) (vmcommon.ReturnCode, error) {
	if tps.ProcessTransactionCalled != nil {
		return tps.ProcessTransactionCalled(transaction)
	}

	return 0, nil
}

// RegisterUnExecutableTransaction -
func (tps *TxProcessorStub) RegisterUnExecutableTransaction(transaction *transaction.Transaction, txHash []byte) error {
	if tps.RegisterUnExecutableTransactionCalled != nil {
		return tps.RegisterUnExecutableTransactionCalled(transaction, txHash)
	}

	return nil
}

// VerifyTransaction -
func (tps *TxProcessorStub) VerifyTransaction(tx *transaction.Transaction) error {
	if tps.VerifyTransactionCalled != nil {
		return tps.VerifyTransactionCalled(tx)
	}

	return nil
}

// VerifyGuardian -
func (tps *TxProcessorStub) VerifyGuardian(tx *transaction.Transaction, account state.UserAccountHandler) error {
	if tps.VerifyGuardianCalled != nil {
		return tps.VerifyGuardianCalled(tx, account)
	}

	return nil
}

// IsInterfaceNil -
func (tps *TxProcessorStub) IsInterfaceNil() bool {
	return tps == nil
}
