package sovereign

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/sovereign"
)

type outGoingBridgeOperations struct {
}

func NewOutGoingBridgeOperationsHandler() (*outGoingBridgeOperations, error) {
	return &outGoingBridgeOperations{}, nil
}

func (op *outGoingBridgeOperations) Send(ctx context.Context, data *sovereign.BridgeOperations) error {
	_ = ctx
	_ = data

	log.Debug("outGoingBridgeOperations.Send called")

	return nil
}

func (op *outGoingBridgeOperations) IsInterfaceNil() bool {
	return op == nil
}
