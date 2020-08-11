package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TxLogsProcessorMock -
type TxLogsProcessorMock struct {
	txLogs map[string]*transaction.Log
}

// NewTxLogsProcessorMock -
func NewTxLogsProcessorMock() *TxLogsProcessorMock {
	return &TxLogsProcessorMock{
		txLogs: make(map[string]*transaction.Log),
	}
}

// GetLog -
func (m *TxLogsProcessorMock) GetLog(txHash []byte) (data.LogHandler, error) {
	txLog, ok := m.txLogs[string(txHash)]
	if !ok {
		return nil, process.ErrLogNotFound
	}

	return txLog, nil
}

// SaveLog -
func (m *TxLogsProcessorMock) SaveLog(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
	txLog := &transaction.Log{
		Address: m.getLogAddressByTx(tx),
	}

	for _, logEntry := range vmLogs {
		txLog.Events = append(txLog.Events, &transaction.Event{
			Identifier: logEntry.Identifier,
			Address:    logEntry.Address,
			Topics:     logEntry.Topics,
			Data:       logEntry.Data,
		})
	}

	m.txLogs[string(txHash)] = txLog
	return nil
}

func (m *TxLogsProcessorMock) getLogAddressByTx(tx data.TransactionHandler) []byte {
	if core.IsEmptyAddress(tx.GetRcvAddr()) {
		return tx.GetSndAddr()
	}
	return tx.GetRcvAddr()
}

// IsInterfaceNil -
func (m *TxLogsProcessorMock) IsInterfaceNil() bool {
	return m == nil
}
