package preprocess

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type scrInfo struct {
	txHash    []byte
	txHandler data.TransactionHandler
}

type scheduledTxsExecution struct {
	txProcessor     process.TransactionProcessor
	mapScheduledTxs map[string]data.TransactionHandler
	scheduledSCRs   []data.TransactionHandler
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
		scheduledSCRs:   make([]data.TransactionHandler, 0),
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

	err := ste.execute(txHandler)
	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		return err
	}

	return nil
}

// ExecuteAll method executes all the scheduled transactions
func (ste *scheduledTxsExecution) ExecuteAll(
	haveTime func() time.Duration,
	txCoordinator process.TransactionCoordinator,
) error {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}
	if check.IfNil(txCoordinator) {
		return process.ErrNilTransactionCoordinator
	}

	allTxsBeforeScheduledExecution := txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)

	for _, txHandler := range ste.scheduledTxs {
		if haveTime() < 0 {
			return process.ErrTimeIsOut
		}

		err := ste.execute(txHandler)
		if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
			return err
		}
	}

	allTxsAfterScheduledExecution := txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	ste.computeScheduledSCRs(allTxsBeforeScheduledExecution, allTxsAfterScheduledExecution)

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

func (ste *scheduledTxsExecution) computeScheduledSCRs(
	allTxsBeforeScheduledExecution map[string]data.TransactionHandler,
	allTxsAfterScheduledExecution map[string]data.TransactionHandler,
) {
	scrsInfo := make([]*scrInfo, 0)
	for txHash, txHandler := range allTxsAfterScheduledExecution {
		_, txExists := allTxsBeforeScheduledExecution[txHash]
		if txExists {
			continue
		}

		scrsInfo = append(scrsInfo, &scrInfo{
			txHash:    []byte(txHash),
			txHandler: txHandler,
		})
	}

	sort.Slice(scrsInfo, func(a, b int) bool {
		return bytes.Compare(scrsInfo[a].txHash, scrsInfo[b].txHash) < 0
	})

	ste.scheduledSCRs = make([]data.TransactionHandler, len(scrsInfo))
	for scrIndex, scrInfo := range scrsInfo {
		ste.scheduledSCRs[scrIndex] = scrInfo.txHandler
	}
}

// GetScheduledSCRs returns all the scheduled SCRs
func (ste *scheduledTxsExecution) GetScheduledSCRs() []data.TransactionHandler {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	scheduledSCRs := make([]data.TransactionHandler, len(ste.scheduledSCRs))
	for scrIndex, txHandler := range ste.scheduledSCRs {
		scheduledSCRs[scrIndex] = txHandler
	}

	return scheduledSCRs
}

// SetScheduledSCRs sets the given scheduled SCRs
func (ste *scheduledTxsExecution) SetScheduledSCRs(scheduledSCRs []data.TransactionHandler) {
	ste.mutScheduledTxs.Lock()
	defer ste.mutScheduledTxs.Unlock()

	ste.scheduledSCRs = make([]data.TransactionHandler, len(scheduledSCRs))
	for scrIndex, txHandler := range scheduledSCRs {
		ste.scheduledSCRs[scrIndex] = txHandler
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ste *scheduledTxsExecution) IsInterfaceNil() bool {
	return ste == nil
}
