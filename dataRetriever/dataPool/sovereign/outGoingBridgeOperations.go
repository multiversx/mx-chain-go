package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type outGoingBridgeOperations struct {
}

// NewOutGoingBridgeOperationsHandler creates a new sovereign outgoing bridge operations handler
func NewOutGoingBridgeOperationsHandler() (*outGoingBridgeOperations, error) {
	return &outGoingBridgeOperations{}, nil
}

// Send should be able to send bridge data operations from sovereign
func (op *outGoingBridgeOperations) Send(ctx context.Context, data *sovereign.BridgeOperations) error {
	_ = ctx
	_ = data

	log.Debug("outGoingBridgeOperations.Send called")

	return nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (op *outGoingBridgeOperations) IsInterfaceNil() bool {
	return op == nil
}
