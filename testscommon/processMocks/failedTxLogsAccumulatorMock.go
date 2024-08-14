package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// FailedTxLogsAccumulatorMock -
type FailedTxLogsAccumulatorMock struct {
	GetLogsCalled  func(txHash []byte) (data.TransactionHandler, []*vmcommon.LogEntry, bool)
	SaveLogsCalled func(txHash []byte, tx data.TransactionHandler, logs []*vmcommon.LogEntry) error
	RemoveCalled   func(txHash []byte)
}

// GetLogs -
func (mock *FailedTxLogsAccumulatorMock) GetLogs(txHash []byte) (data.TransactionHandler, []*vmcommon.LogEntry, bool) {
	if mock.GetLogsCalled != nil {
		return mock.GetLogsCalled(txHash)
	}
	return nil, nil, false
}

// SaveLogs -
func (mock *FailedTxLogsAccumulatorMock) SaveLogs(txHash []byte, tx data.TransactionHandler, logs []*vmcommon.LogEntry) error {
	if mock.SaveLogsCalled != nil {
		return mock.SaveLogsCalled(txHash, tx, logs)
	}
	return nil
}

// Remove -
func (mock *FailedTxLogsAccumulatorMock) Remove(txHash []byte) {
	if mock.RemoveCalled != nil {
		mock.RemoveCalled(txHash)
	}
}

// IsInterfaceNil -
func (mock *FailedTxLogsAccumulatorMock) IsInterfaceNil() bool {
	return mock == nil
}
