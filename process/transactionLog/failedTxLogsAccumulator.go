package transactionLog

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type logData struct {
	tx   data.TransactionHandler
	logs []*vmcommon.LogEntry
}

type failedTxLogsAccumulator struct {
	mut     sync.RWMutex
	logsMap map[string]*logData
}

// NewFailedTxLogsAccumulator returns a new instance of failedTxLogsAccumulator
func NewFailedTxLogsAccumulator() *failedTxLogsAccumulator {
	return &failedTxLogsAccumulator{
		logsMap: make(map[string]*logData),
	}
}

// GetLogs returns the accumulated logs for the provided txHash
func (accumulator *failedTxLogsAccumulator) GetLogs(txHash []byte) (data.TransactionHandler, []*vmcommon.LogEntry, bool) {
	if len(txHash) == 0 {
		return nil, nil, false
	}

	logsData, found := accumulator.getLogDataCopy(txHash)

	if !found {
		return nil, nil, found
	}

	return logsData.tx, logsData.logs, found
}

func (accumulator *failedTxLogsAccumulator) getLogDataCopy(txHash []byte) (logData, bool) {
	accumulator.mut.RLock()
	defer accumulator.mut.RUnlock()

	logsData, found := accumulator.logsMap[string(txHash)]
	if !found {
		return logData{}, found
	}

	logsDataCopy := logData{
		tx: logsData.tx,
	}

	logsDataCopy.logs = append(logsDataCopy.logs, logsData.logs...)

	return logsDataCopy, found
}

// SaveLogs saves the logs into the internal map
func (accumulator *failedTxLogsAccumulator) SaveLogs(txHash []byte, tx data.TransactionHandler, logs []*vmcommon.LogEntry) error {
	if len(txHash) == 0 {
		return process.ErrNilTxHash
	}

	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	if len(logs) == 0 {
		return nil
	}

	accumulator.mut.Lock()
	defer accumulator.mut.Unlock()

	_, found := accumulator.logsMap[string(txHash)]
	if !found {
		accumulator.logsMap[string(txHash)] = &logData{
			tx:   tx,
			logs: logs,
		}

		return nil
	}

	accumulator.logsMap[string(txHash)].logs = append(accumulator.logsMap[string(txHash)].logs, logs...)

	return nil
}

// Remove removes the accumulated logs for the provided txHash
func (accumulator *failedTxLogsAccumulator) Remove(txHash []byte) {
	if len(txHash) == 0 {
		return
	}

	accumulator.mut.Lock()
	defer accumulator.mut.Unlock()

	delete(accumulator.logsMap, string(txHash))
}

// IsInterfaceNil returns true if there is no value under the interface
func (accumulator *failedTxLogsAccumulator) IsInterfaceNil() bool {
	return accumulator == nil
}
