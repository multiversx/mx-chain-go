package mock

import "github.com/ElrondNetwork/elrond-go-core/data/transaction"

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
	GetTransactionCalled       func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	UnmarshalTransactionCalled func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
	UnmarshalReceiptCalled     func(receiptBytes []byte) (*transaction.ApiReceipt, error)
}

// GetTransaction -
func (tas *TransactionAPIHandlerStub) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	if tas.GetTransactionCalled != nil {
		return tas.GetTransactionCalled(hash, withResults)
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
