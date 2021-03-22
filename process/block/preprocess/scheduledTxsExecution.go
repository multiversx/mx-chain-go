package preprocess

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type scheduledTxsExecution struct {
	txProcessor     process.TransactionProcessor
	mapScheduledTxs map[string]data.TransactionHandler
	scheduledTxs    []data.TransactionHandler
	mutScheduledTxs sync.RWMutex
}

// NewScheduledTxsExecution creates a new object which handles the execution of scheduled transactions
func NewScheduledTxsExecution(
	txProcessor process.TransactionProcessor,
) (*scheduledTxsExecution, error) {

	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}

	ste := &scheduledTxsExecution{
		txProcessor:     txProcessor,
		mapScheduledTxs: make(map[string]data.TransactionHandler),
		scheduledTxs:    make([]data.TransactionHandler, 0),
	}

	return ste, nil
}

// Init method removes all the scheduled transactions
func (ste *scheduledTxsExecution) Init() {
	ste.mutScheduledTxs.Lock()
	//TODO: Remove this log
	log.Debug("scheduledTxsExecution.Init", "num scheduled txs", len(ste.scheduledTxs))
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

	err := ste.execute(txHandler)
	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		return err
	}

	return nil
}

// ExecuteAll method executes all the scheduled transactions
func (ste *scheduledTxsExecution) ExecuteAll(haveTime func() time.Duration) error {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	for _, txHandler := range ste.scheduledTxs {
		if haveTime() < 0 {
			return process.ErrTimeIsOut
		}

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

	_, err := ste.txProcessor.ProcessTransaction(tx)
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (ste *scheduledTxsExecution) IsInterfaceNil() bool {
	return ste == nil
}
