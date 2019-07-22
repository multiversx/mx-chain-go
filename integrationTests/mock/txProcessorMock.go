package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"math/big"
)

type TxProcessorMock struct {
	ProcessTransactionCalled         func(transaction data.TransactionHandler, round uint32) error
	SetBalancesToTrieCalled          func(accBalance map[string]*big.Int) (rootHash []byte, err error)
	ProcessSmartContractResultCalled func(scr *smartContractResult.SmartContractResult) error
}

func (etm *TxProcessorMock) ProcessTransaction(transaction data.TransactionHandler, round uint32) error {
	return etm.ProcessTransactionCalled(transaction, round)
}

func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	return etm.SetBalancesToTrieCalled(accBalance)
}

func (etm *TxProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	return etm.ProcessSmartContractResultCalled(scr)
}
