package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.TransactionLogProcessorDatabase = (*TxLogProcessorMock)(nil)

// TxLogProcessorMock -
type TxLogProcessorMock struct {
}

// GetLogFromCache -
func (t *TxLogProcessorMock) GetLogFromCache(_ []byte) (data.LogDataHandler, bool) {
	return &transaction.LogData{}, false
}

// EnableLogToBeSavedInCache -
func (t *TxLogProcessorMock) EnableLogToBeSavedInCache() {
}

// Clean -
func (t *TxLogProcessorMock) Clean() {
}

// IsInterfaceNil -
func (t *TxLogProcessorMock) IsInterfaceNil() bool {
	return t == nil
}
