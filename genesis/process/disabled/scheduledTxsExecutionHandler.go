package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

// ScheduledTxsExecutionHandler implements ScheduledTxsExecutionHandler interface but does nothing as it is a disabled component
type ScheduledTxsExecutionHandler struct {
}

// Init does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Init() {
}

// Add does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Add(_ []byte, _ data.TransactionHandler) bool {
	return true
}

// Execute does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) Execute(_ []byte) error {
	return nil
}

// ExecuteAll does nothing as it is a disabled component
func (steh *ScheduledTxsExecutionHandler) ExecuteAll(_ func() time.Duration) error {
	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (steh *ScheduledTxsExecutionHandler) IsInterfaceNil() bool {
	return steh == nil
}
