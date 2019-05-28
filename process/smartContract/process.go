package smartContract

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// TransactionType specifies the type of the transaction
type TransactionType int

const (
	// MoveBalance defines ID of a payment transaction - moving balances
	MoveBalance TransactionType = iota
	// SCDeployment defines ID of a transaction to store a smart contract
	SCDeployment
	// BHProposed defines ID of a transaction of type smart contract call
	SCInvoking
)

type scProcessor struct {
	accounts         state.AccountsAdapter
	adrConv          state.AddressConverter
}

func NewSmartContractProcessor() *scProcessor {
	return &scProcessor{}
}

func (sc *scProcessor) ComputeTransactionType(transaction *transaction.Transaction) (TransactionType, error) {
	if transaction == nil {
		return 0, process.ErrNilTransaction
	}

	if len(transaction.RcvAddr) == 0 {
		sc.accounts.GetExistingAccount()
		if acco {
			return SCDeployment, nil
		}
		return 0, process.ErrWrongTransaction
	}

	if len(transaction.RcvAddr) >= 0 {
		if len(transaction.d)
	}

	return 0, process.ErrWrongTransaction
}