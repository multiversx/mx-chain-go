package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

type failedTxLogsAccumulator struct {
}

// NewFailedTxLogsAccumulator returns a new instance of disabled failedTxLogsAccumulator
func NewFailedTxLogsAccumulator() *failedTxLogsAccumulator {
	return &failedTxLogsAccumulator{}
}

// GetLogs returns false as it is disabled
func (accumulator *failedTxLogsAccumulator) GetLogs(_ []byte) (data.TransactionHandler, []*vmcommon.LogEntry, bool) {
	return nil, nil, false
}

// SaveLogs returns nil as it is disabled
func (accumulator *failedTxLogsAccumulator) SaveLogs(_ []byte, _ data.TransactionHandler, _ []*vmcommon.LogEntry) error {
	return nil
}

// Remove does nothing as it is disabled
func (accumulator *failedTxLogsAccumulator) Remove(_ []byte) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (accumulator *failedTxLogsAccumulator) IsInterfaceNil() bool {
	return accumulator == nil
}
