package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxProcessorMock -
type TxProcessorMock struct {
	ProcessTransactionCalled         func(transaction *transaction.Transaction) error
	SetBalancesToTrieCalled          func(accBalance map[string]*big.Int) (rootHash []byte, err error)
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) error
}

// ProcessTransaction -
func (etm *TxProcessorMock) ProcessTransaction(transaction *transaction.Transaction) error {
	if etm.ProcessTransactionCalled != nil {
		return etm.ProcessTransactionCalled(transaction)
	}

	return nil
}

// SetBalancesToTrie -
func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	if etm.SetBalancesToTrieCalled != nil {
		return etm.SetBalancesToTrieCalled(accBalance)
	}
	return nil, nil
}

// ProcessSmartContractResult -
func (etm *TxProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (ReturnCode, error) {
	if etm.ProcessSmartContractResultCalled != nil {
		return 0, etm.ProcessSmartContractResultCalled(scr)
	}
	return 0, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (etm *TxProcessorMock) IsInterfaceNil() bool {
	return etm == nil
}
