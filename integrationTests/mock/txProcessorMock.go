package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

type TxProcessorMock struct {
	ProcessTransactionCalled         func(transaction *transaction.Transaction, round uint64) error
	SetBalancesToTrieCalled          func(accBalance map[string]*big.Int) (rootHash []byte, err error)
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) error
}

func (etm *TxProcessorMock) ProcessTransaction(transaction *transaction.Transaction, round uint64) error {
	return etm.ProcessTransactionCalled(transaction, round)
}

func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	return etm.SetBalancesToTrieCalled(accBalance)
}

func (etm *TxProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	return etm.ProcessSmartContractResultCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (etm *TxProcessorMock) IsInterfaceNil() bool {
	if etm == nil {
		return true
	}
	return false
}
