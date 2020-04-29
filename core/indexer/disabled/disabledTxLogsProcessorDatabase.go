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

// GetLogFromRAM -
func (t *txLogProcessor) GetLogFromRAM(_ []byte) (data.LogHandler, bool) {
	return nil, false
}

// SaveLogsAlsoInRAM -
func (t *txLogProcessor) SaveLogsAlsoInRAM() {

}

// RemoveLogsFromRAM -
func (t *txLogProcessor) RemoveLogsFromRAM() {

}
