package mock

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TxLogProcessorMock -
type TxLogProcessorMock struct {
}

// GetLog -
func (tlpm *TxLogProcessorMock) GetLog(_ []byte) (data.LogHandler, error) {
	return nil, fmt.Errorf("log not found for provided tx hash")
}

// SaveLog -
func (tlpm *TxLogProcessorMock) SaveLog(_ []byte, _ data.TransactionHandler, _ []*vmcommon.LogEntry) error {
	return nil
}

// IsInterfaceNil -
func (tlpm *TxLogProcessorMock) IsInterfaceNil() bool {
	return tlpm == nil
}
