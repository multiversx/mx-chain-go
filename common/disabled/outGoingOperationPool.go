package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type outGoingOperationsPool struct {
}

// NewDisabledOutGoingOperationPool -
func NewDisabledOutGoingOperationPool() *outGoingOperationsPool {
	return &outGoingOperationsPool{}
}

// Add -
func (op *outGoingOperationsPool) Add(_ *sovereign.BridgeOutGoingData) {}

// Get -
func (op *outGoingOperationsPool) Get(_ []byte) *sovereign.BridgeOutGoingData {
	return &sovereign.BridgeOutGoingData{}
}

// Delete -
func (op *outGoingOperationsPool) Delete(_ []byte) {}

// ConfirmOperation -
func (op *outGoingOperationsPool) ConfirmOperation(_ []byte, _ []byte) error {
	return nil
}

// GetUnconfirmedOperations -
func (op *outGoingOperationsPool) GetUnconfirmedOperations() []*sovereign.BridgeOutGoingData {
	return make([]*sovereign.BridgeOutGoingData, 0)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outGoingOperationsPool) IsInterfaceNil() bool {
	return op == nil
}
