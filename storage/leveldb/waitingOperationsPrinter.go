package leveldb

import (
	"fmt"
	"sync"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
)

const maxTimeOperationForHighPrio = time.Millisecond * 100

type waitingOperationsPrinter struct {
	counters      *waitingOperationsCounters
	logger        logger.Logger
	mutOperation  sync.Mutex
	problemsFound bool
}

func newWaitingOperationsPrinter() *waitingOperationsPrinter {
	return &waitingOperationsPrinter{
		counters: newWaitingOperationsCounters(),
		logger:   log,
	}
}

func (printer *waitingOperationsPrinter) startExecuting(priority common.StorageAccessType) {
	printer.mutOperation.Lock()
	defer printer.mutOperation.Unlock()

	printer.counters.increment(priority)
}

func (printer *waitingOperationsPrinter) endExecuting(priority common.StorageAccessType, operation string, startTime time.Time) {
	printer.mutOperation.Lock()
	defer printer.mutOperation.Unlock()

	printer.counters.decrement(priority)

	if priority != common.HighPriority {
		return
	}

	measuredDuration := time.Since(startTime)
	if measuredDuration > maxTimeOperationForHighPrio && !printer.problemsFound {
		printer.problemsFound = true

		printer.logger.Warn(fmt.Sprintf("high priority disk operation took > %v", maxTimeOperationForHighPrio),
			"measured duration", measuredDuration,
			"operation", operation,
			"pending operations", printer.counters.snapshotString())
	}

	printer.checkAndPrintProblemsEnded()
}

func (printer *waitingOperationsPrinter) checkAndPrintProblemsEnded() {
	if printer.counters.numNonZeroCounters > 0 {
		return
	}
	if !printer.problemsFound {
		return
	}

	printer.problemsFound = false

	printer.logger.Debug("storage problems ended, all pending operations finished")
}
