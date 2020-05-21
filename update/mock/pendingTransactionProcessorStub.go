package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// PendingTransactionProcessorStub -
type PendingTransactionProcessorStub struct {
	ProcessTransactionsDstMeCalled func(mapTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error)
	RootHashCalled                 func() ([]byte, error)
}

// ProcessTransactionsDstMe -
func (ptps *PendingTransactionProcessorStub) ProcessTransactionsDstMe(mapTxs map[string]data.TransactionHandler) (block.MiniBlockSlice, error) {
	if ptps.ProcessTransactionsDstMeCalled != nil {
		return ptps.ProcessTransactionsDstMeCalled(mapTxs)
	}
	return nil, nil
}

// RootHash --
func (ptps *PendingTransactionProcessorStub) RootHash() ([]byte, error) {
	if ptps.RootHashCalled != nil {
		return ptps.RootHashCalled()
	}
	return nil, nil
}

// IsInterfaceNil --
func (ptps *PendingTransactionProcessorStub) IsInterfaceNil() bool {
	return ptps == nil
}
