package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

// LogsRepositoryStub -
type LogsRepositoryStub struct {
	GetLogCalled                    func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error)
	GetLogsCalled                   func(txHashes [][]byte, epoch uint32) ([]*transaction.ApiLogs, error)
	IncludeLogsInTransactionsCalled func(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error
}

// GetLog -
func (stub *LogsRepositoryStub) GetLog(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
	if stub.GetLogCalled != nil {
		return stub.GetLogCalled(txHash, epoch)
	}

	return nil, nil
}

// GetLogs -
func (stub *LogsRepositoryStub) GetLogs(txHashes [][]byte, epoch uint32) ([]*transaction.ApiLogs, error) {
	if stub.GetLogsCalled != nil {
		return stub.GetLogsCalled(txHashes, epoch)
	}

	return nil, nil
}

// IncludeLogsInTransactions
func (stub *LogsRepositoryStub) IncludeLogsInTransactions(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error {
	if stub.IncludeLogsInTransactionsCalled != nil {
		return stub.IncludeLogsInTransactionsCalled(txs, logsKeys, epoch)
	}

	return nil
}

// IsInterfaceNil -
func (stub *LogsRepositoryStub) IsInterfaceNil() bool {
	return stub == nil
}
