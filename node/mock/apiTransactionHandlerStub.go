package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
)

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
	GetTransactionCalled         func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPoolCalled    func() (*common.TransactionsPoolAPIResponse, error)
	UnmarshalTransactionCalled   func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	UnmarshalReceiptCalled       func(receiptBytes []byte) (*transaction.ApiReceipt, error)
	PopulateComputedFieldsCalled func(tx *transaction.ApiTransactionResult)
}

// GetTransaction -
func (tas *TransactionAPIHandlerStub) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	if tas.GetTransactionCalled != nil {
		return tas.GetTransactionCalled(hash, withResults)
	}

	return nil, nil
}

// GetTransactionsPool -
func (tas *TransactionAPIHandlerStub) GetTransactionsPool() (*common.TransactionsPoolAPIResponse, error) {
	if tas.GetTransactionsPoolCalled != nil {
		return tas.GetTransactionsPoolCalled()
	}

	return nil, nil
}

// UnmarshalTransaction -
func (tas *TransactionAPIHandlerStub) UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	if tas.UnmarshalTransactionCalled != nil {
		return tas.UnmarshalTransactionCalled(txBytes, txType)
	}
	return nil, nil
}

// UnmarshalReceipt -
func (tas *TransactionAPIHandlerStub) UnmarshalReceipt(receiptBytes []byte) (*transaction.ApiReceipt, error) {
	if tas.UnmarshalReceiptCalled != nil {
		return tas.UnmarshalReceiptCalled(receiptBytes)
	}

	return nil, nil
}

// PopulateComputedFields -
func (tas *TransactionAPIHandlerStub) PopulateComputedFields(tx *transaction.ApiTransactionResult) {
	if tas.PopulateComputedFieldsCalled != nil {
		tas.PopulateComputedFieldsCalled(tx)
	}
}

// IsInterfaceNil -
func (tas *TransactionAPIHandlerStub) IsInterfaceNil() bool {
	return tas == nil
}
