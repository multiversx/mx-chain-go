package preprocess

import (
	"errors"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type scheduledTxsExecution struct {
	txProcessor     process.TransactionProcessor
	accounts        state.AccountsAdapter
	mapScheduledTxs map[string]data.TransactionHandler
	scheduledTxs    []data.TransactionHandler
	mutScheduledTxs sync.RWMutex
}

// NewScheduledTxsExecution creates a new object which handles the execution of scheduled transactions
func NewScheduledTxsExecution(
	txProcessor process.TransactionProcessor,
	accounts state.AccountsAdapter,
) (*scheduledTxsExecution, error) {

	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(accounts) {
		return nil, process.ErrNilAccountsAdapter
	}

	ste := &scheduledTxsExecution{
		txProcessor:     txProcessor,
		accounts:        accounts,
		mapScheduledTxs: make(map[string]data.TransactionHandler),
		scheduledTxs:    make([]data.TransactionHandler, 0),
	}

	return ste, nil
}

// Init method removes all the scheduled transactions
func (ste *scheduledTxsExecution) Init() {
	ste.mutScheduledTxs.Lock()
	ste.mapScheduledTxs = make(map[string]data.TransactionHandler)
	ste.scheduledTxs = make([]data.TransactionHandler, 0)
	ste.mutScheduledTxs.Unlock()
}

// Add method adds a scheduled transaction to be executed
func (ste *scheduledTxsExecution) Add(txHash []byte, tx data.TransactionHandler) bool {
	ste.mutScheduledTxs.Lock()
	defer ste.mutScheduledTxs.Unlock()

	_, exist := ste.mapScheduledTxs[string(txHash)]
	if exist {
		return false
	}

	ste.mapScheduledTxs[string(txHash)] = tx
	ste.scheduledTxs = append(ste.scheduledTxs, tx)

	return true
}

// Execute method executes the given scheduled transaction
func (ste *scheduledTxsExecution) Execute(txHash []byte) error {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	txHandler, exist := ste.mapScheduledTxs[string(txHash)]
	if !exist {
		return fmt.Errorf("%w: in scheduledTxsExecution.Execute", process.ErrMissingTransaction)
	}

	return ste.execute(txHandler)
}

// ExecuteAll method executes all the scheduled transactions
func (ste *scheduledTxsExecution) ExecuteAll() error {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	for _, txHandler := range ste.scheduledTxs {
		err := ste.execute(txHandler)
		if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
			return err
		}
	}

	return nil
}

func (ste *scheduledTxsExecution) execute(txHandler data.TransactionHandler) error {
	tx, ok := txHandler.(*transaction.Transaction)
	if !ok {
		return fmt.Errorf("%w: in scheduledTxsExecution.execute", process.ErrWrongTypeAssertion)
	}

	snapshot := ste.accounts.JournalLen()

	_, err := ste.txProcessor.ProcessTransaction(tx)
	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		errRevert := ste.accounts.RevertToSnapshot(snapshot)
		if errRevert != nil {
			log.Error("scheduledTxsExecution.execute: revert to snapshot", "error", errRevert.Error())
		}
	}

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ste *scheduledTxsExecution) IsInterfaceNil() bool {
	return ste == nil
}
