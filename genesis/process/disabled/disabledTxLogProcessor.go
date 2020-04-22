package disabled

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/data"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// DisabledTxLogProcessor represents a disabled transaction log processor that does nothing
type DisabledTxLogProcessor struct {
}

// GetLog will return a not found error
func (dtlp *DisabledTxLogProcessor) GetLog(_ []byte) (data.LogHandler, error) {
	return nil, fmt.Errorf("log not found for provided tx hash")
}

// SaveLog will do nothing
func (dtlp *DisabledTxLogProcessor) SaveLog(_ []byte, _ data.TransactionHandler, _ []*vmcommon.LogEntry) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dtlp *DisabledTxLogProcessor) IsInterfaceNil() bool {
	return dtlp == nil
}
