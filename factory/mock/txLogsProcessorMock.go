package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TransactionLogProcessorDatabase = (*TxLogProcessorMock)(nil)

// TxLogProcessorMock -
type TxLogProcessorMock struct {
}

// GetLogFromCache -
func (t *TxLogProcessorMock) GetLogFromCache(_ []byte) (*indexer.LogData, bool) {
	return &indexer.LogData{}, false
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
