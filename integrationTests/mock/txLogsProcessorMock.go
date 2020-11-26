package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TransactionLogProcessorDatabase = (*TxLogProcessorMock)(nil)

// TxLogProcessorMock -
type TxLogProcessorMock struct {
}

// GetLogFromCache -
func (t *TxLogProcessorMock) GetLogFromCache(_ []byte) (data.LogHandler, bool) {
	return nil, false
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
