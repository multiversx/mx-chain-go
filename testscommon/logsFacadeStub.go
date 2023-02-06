package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

// LogsFacadeStub -
type LogsFacadeStub struct {
	GetLogCalled                    func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error)
	IncludeLogsInTransactionsCalled func(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error
}

// GetLog -
func (stub *LogsFacadeStub) GetLog(logKey []byte, epoch uint32) (*transaction.ApiLogs, error) {
	if stub.GetLogCalled != nil {
		return stub.GetLogCalled(logKey, epoch)
	}

	return nil, nil
}

// IncludeLogsInTransactions -
func (stub *LogsFacadeStub) IncludeLogsInTransactions(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error {
	if stub.IncludeLogsInTransactionsCalled != nil {
		return stub.IncludeLogsInTransactionsCalled(txs, logsKeys, epoch)
	}

	return nil
}

// IsInterfaceNil -
func (stub *LogsFacadeStub) IsInterfaceNil() bool {
	return stub == nil
}
