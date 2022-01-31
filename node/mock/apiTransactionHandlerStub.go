package mock

import "github.com/ElrondNetwork/elrond-go-core/data/transaction"

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
	UnmarshalTransactionCalled func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
}

// GetTransaction -
func (tas *TransactionAPIHandlerStub) GetTransaction(_ string, _ bool) (*transaction.ApiTransactionResult, error) {
	return nil, nil
}

// UnmarshalTransaction -
func (tas *TransactionAPIHandlerStub) UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	if tas.UnmarshalTransactionCalled != nil {
		return tas.UnmarshalTransactionCalled(txBytes, txType)
	}
	return nil, nil
}

// IsInterfaceNil -
func (tas *TransactionAPIHandlerStub) IsInterfaceNil() bool {
	return tas == nil
}
