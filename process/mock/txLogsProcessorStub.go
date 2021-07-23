package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// TxLogsProcessorStub -
type TxLogsProcessorStub struct {
	GetLogCalled            func(txHash []byte) (data.LogHandler, error)
	SaveLogCalled           func(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error
	GetAllCurrentLogsCalled func() map[string]data.LogHandler
}

// GetLog -
func (txls *TxLogsProcessorStub) GetLog(txHash []byte) (data.LogHandler, error) {
	if txls.GetLogCalled != nil {
		return txls.GetLogCalled(txHash)
	}

	return nil, nil
}

// Clean -
func (txls *TxLogsProcessorStub) Clean() {
}

// SaveLog -
func (txls *TxLogsProcessorStub) SaveLog(txHash []byte, tx data.TransactionHandler, vmLogs []*vmcommon.LogEntry) error {
	if txls.SaveLogCalled != nil {
		return txls.SaveLogCalled(txHash, tx, vmLogs)
	}

	return nil
}

// GetAllCurrentLogs -
func (txls *TxLogsProcessorStub) GetAllCurrentLogs() map[string]data.LogHandler {
	if txls.GetAllCurrentLogsCalled != nil {
		return txls.GetAllCurrentLogsCalled()
	}

	return nil
}

// IsInterfaceNil -
func (txls *TxLogsProcessorStub) IsInterfaceNil() bool {
	return txls == nil
}
