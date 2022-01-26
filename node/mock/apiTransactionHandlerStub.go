package mock

import "github.com/ElrondNetwork/elrond-go-core/data/transaction"

// TransactionAPIHandlerStub -
type TransactionAPIHandlerStub struct {
}

// GetTransaction -
func (tas *TransactionAPIHandlerStub) GetTransaction(_ string, _ bool) (*transaction.ApiTransactionResult, error) {
	return nil, nil
}

// UnmarshalTransaction -
func (tas *TransactionAPIHandlerStub) UnmarshalTransaction(_ []byte, _ transaction.TxType) (*transaction.ApiTransactionResult, error) {
	return nil, nil
}

// IsInterfaceNil -
func (tas *TransactionAPIHandlerStub) IsInterfaceNil() bool {
	return tas == nil
}
