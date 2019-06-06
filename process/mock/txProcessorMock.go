package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type TxProcessorMock struct {
	ProcessTransactionCalled func(transaction *transaction.Transaction, round uint32) error
	SetBalancesToTrieCalled  func(accBalance map[string]*big.Int) (rootHash []byte, err error)
}

func (etm *TxProcessorMock) ProcessTransaction(transaction *transaction.Transaction, round uint32) error {
	return etm.ProcessTransactionCalled(transaction, round)
}

func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	return etm.SetBalancesToTrieCalled(accBalance)
}
