package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
)

// LogsFacadeStub -
type LogsFacadeStub struct {
	GetLogCalled                    func(txHash []byte, epoch uint32) (*transaction.ApiLogs, error)
	GetLogsCalled                   func(txHashes [][]byte, epoch uint32) ([]*transaction.ApiLogs, error)
	IncludeLogsInTransactionsCalled func(txs []*transaction.ApiTransactionResult, logsKeys [][]byte, epoch uint32) error
}

// GetLog -
func (stub *LogsFacadeStub) GetLog(txHash []byte, epoch uint32) (*transaction.ApiLogs, error) {
	if stub.GetLogCalled != nil {
		return stub.GetLogCalled(txHash, epoch)
	}

	return nil, nil
}

// GetLogs -
func (stub *LogsFacadeStub) GetLogs(txHashes [][]byte, epoch uint32) ([]*transaction.ApiLogs, error) {
	if stub.GetLogsCalled != nil {
		return stub.GetLogsCalled(txHashes, epoch)
	}

	return nil, nil
}

// IncludeLogsInTransactions
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
