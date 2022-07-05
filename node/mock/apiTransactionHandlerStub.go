package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
)

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
	GetTransactionCalled                        func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	GetTransactionsPoolCalled                   func(fields string) (*common.TransactionsPoolAPIResponse, error)
	UnmarshalTransactionCalled                  func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	GetTransactionsPoolForSenderCalled          func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error)
	GetLastPoolNonceForSenderCalled             func(sender string) (uint64, error)
	GetTransactionsPoolNonceGapsForSenderCalled func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error)
	UnmarshalReceiptCalled                      func(receiptBytes []byte) (*transaction.ApiReceipt, error)
}

// GetTransaction -
func (tas *TransactionAPIHandlerStub) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	if tas.GetTransactionCalled != nil {
		return tas.GetTransactionCalled(hash, withResults)
	}

	return nil, nil
}

// GetTransactionsPool -
func (tas *TransactionAPIHandlerStub) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	if tas.GetTransactionsPoolCalled != nil {
		return tas.GetTransactionsPoolCalled(fields)
	}

	return nil, nil
}

// GetTransactionsPoolForSender -
func (tas *TransactionAPIHandlerStub) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	if tas.GetTransactionsPoolForSenderCalled != nil {
		return tas.GetTransactionsPoolForSenderCalled(sender, fields)
	}

	return nil, nil
}

// GetLastPoolNonceForSender -
func (tas *TransactionAPIHandlerStub) GetLastPoolNonceForSender(sender string) (uint64, error) {
	if tas.GetLastPoolNonceForSenderCalled != nil {
		return tas.GetLastPoolNonceForSenderCalled(sender)
	}

	return 0, nil
}

// GetTransactionsPoolNonceGapsForSender -
func (tas *TransactionAPIHandlerStub) GetTransactionsPoolNonceGapsForSender(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	if tas.GetTransactionsPoolNonceGapsForSenderCalled != nil {
		return tas.GetTransactionsPoolNonceGapsForSenderCalled(sender)
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
