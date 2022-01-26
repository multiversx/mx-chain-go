package mock

import "github.com/ElrondNetwork/elrond-go-core/data/transaction"

type UnmarshalTxStub struct {
	UnmarshalTransactionCalled func(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error)
}

func (uts *UnmarshalTxStub) UnmarshalTransaction(txBytes []byte, txType transaction.TxType) (*transaction.ApiTransactionResult, error) {
	if uts.UnmarshalTransactionCalled != nil {
		return uts.UnmarshalTransactionCalled(txBytes, txType)
	}

	return nil, nil
}
