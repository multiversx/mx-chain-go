package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
)

// ExecutionOrderHandlerStub -
type ExecutionOrderHandlerStub struct {
}

// PutExecutionOrderInTransactionPool -
func (e *ExecutionOrderHandlerStub) PutExecutionOrderInTransactionPool(
	_ *outport.Pool,
	_ data.HeaderHandler,
	_ data.BodyHandler,
) error {
	return nil
}

// PutExecutionOrderInAPIMiniblocks --
func (e *ExecutionOrderHandlerStub) PutExecutionOrderInAPIMiniblocks(
	_ []*api.MiniBlock,
	_ uint32,
	_ []byte,
) {
	return
}
