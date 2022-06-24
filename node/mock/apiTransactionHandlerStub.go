package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
)

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
	GetTransactionCalled           func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPoolCalled      func() (*common.TransactionsPoolAPIResponse, error)
	UnmarshalTransactionCalled     func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	GetTransactionsForSenderCalled func(sender string) (*common.TransactionsForSenderApiResponse, error)
	UnmarshalReceiptCalled         func(receiptBytes []byte) (*transaction.ApiReceipt, error)
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

// GetTransactionsForSender -
func (tas *TransactionAPIHandlerStub) GetTransactionsForSender(sender string) (*common.TransactionsForSenderApiResponse, error) {
	if tas.GetTransactionsForSenderCalled != nil {
		return tas.GetTransactionsForSenderCalled(sender)
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

// IsInterfaceNil -
func (tas *TransactionAPIHandlerStub) IsInterfaceNil() bool {
	return tas == nil
}
