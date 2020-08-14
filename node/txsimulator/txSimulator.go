package txsimulator

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

type transactionSimulator struct {
	txProcessor     TransactionProcessor
	accountsAdapter state.AccountsAdapter
}

// New returns a new instance of a transactionSimulator
func New(txProcessor TransactionProcessor, accountsAdapter state.AccountsAdapter) (*transactionSimulator, error) {
	if check.IfNil(txProcessor) {
		return nil, node.ErrNilTxSimulatorProcessor
	}
	if check.IfNil(accountsAdapter) {
		return nil, node.ErrNilAccountsAdapter
	}

	txProcessor.SetAccountsAdapter(accountsAdapter)

	return &transactionSimulator{
		txProcessor:     txProcessor,
		accountsAdapter: accountsAdapter,
	}, nil
}

// ProcessTx will process the transaction in a special environment, where state-writing is not allowed
func (ts *transactionSimulator) ProcessTx(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
	txStatus := core.TxStatusReceived
	failReason := ""
	retCode, err := ts.txProcessor.ProcessTransaction(tx)
	if err != nil {
		failReason = err.Error()
		txStatus = core.TxStatusNotExecuted
	}

	smartContractResultsToRet := make(map[string]*smartContractResult.SmartContractResult)
	smartContractResults := ts.txProcessor.GetSmartContractResults()

	for hash, scr := range smartContractResults {
		scrWithType, ok := scr.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		smartContractResultsToRet[hash] = scrWithType
	}

	if retCode == vmcommon.Ok {
		txStatus = core.TxStatusExecuted
	}

	return &transaction.SimulationResults{
		Status:     txStatus,
		FailReason: failReason,
		ScResults:  smartContractResultsToRet,
	}, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *transactionSimulator) IsInterfaceNil() bool {
	return ts == nil
}
