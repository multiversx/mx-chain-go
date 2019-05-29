package smartContract

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type scProcessor struct {
	accounts state.AccountsAdapter
	adrConv  state.AddressConverter
}

func NewSmartContractProcessor() (*scProcessor, error) {
	return &scProcessor{}, nil
}

func (sc *scProcessor) ComputeTransactionType(
	transaction *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
) (process.TransactionType, error) {
	//TODO: add all kind of tests here
	if transaction == nil {
		return 0, process.ErrNilTransaction
	}

	if len(transaction.RcvAddr) == 0 {
		if len(transaction.Data) > 0 {
			return process.SCDeployment, nil
		}
		return 0, process.ErrWrongTransaction
	}

	if acntDst.IsInterfaceNil() {
		return process.MoveBalance, nil
	}

	if len(acntDst.GetCode()) > 0 {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

func (sc *scProcessor) ExecuteSmartContractTransaction(
	transaction *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
) error {
	return process.ErrNoVM
}
