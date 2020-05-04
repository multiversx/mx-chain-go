package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TransactionLogProcessorDatabase = (*txLogProcessor)(nil)

type txLogProcessor struct {
}

// NewNilTxLogsProcessor -
func NewNilTxLogsProcessor() *txLogProcessor {
	return new(txLogProcessor)
}

// GetLogFromCache -
func (t *txLogProcessor) GetLogFromCache(_ []byte) (data.LogHandler, bool) {
	return nil, false
}

// EnableLogToBeSavedInCache -
func (t *txLogProcessor) EnableLogToBeSavedInCache() {
}

// Clean -
func (t *txLogProcessor) Clean() {
}

// IsInterfaceNil -
func (t *txLogProcessor) IsInterfaceNil() bool {
	return t == nil
}
