package mock

import "github.com/multiversx/mx-chain-core-go/data/outport"

// TransactionsFeeHandlerMock -
type TransactionsFeeHandlerMock struct {
	PutFeeAndGasUsedCalled func(pool *outport.TransactionPool, epoch uint32) error
}

// PutFeeAndGasUsed -
func (tfh *TransactionsFeeHandlerMock) PutFeeAndGasUsed(pool *outport.TransactionPool, epoch uint32) error {
	if tfh.PutFeeAndGasUsedCalled != nil {
		return tfh.PutFeeAndGasUsedCalled(pool, epoch)
	}

	return nil
}

// IsInterfaceNil -
func (tfh *TransactionsFeeHandlerMock) IsInterfaceNil() bool {
	return tfh == nil
}
