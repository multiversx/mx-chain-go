package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/update"
)

// PendingTransactionProcessorStub -
type PendingTransactionProcessorStub struct {
	ProcessTransactionsDstMeCalled func(txsInfo []*update.TxInfo) (block.MiniBlockSlice, error)
	RootHashCalled                 func() ([]byte, error)
}

// ProcessTransactionsDstMe -
func (ptps *PendingTransactionProcessorStub) ProcessTransactionsDstMe(txsInfo []*update.TxInfo) (block.MiniBlockSlice, error) {
	if ptps.ProcessTransactionsDstMeCalled != nil {
		return ptps.ProcessTransactionsDstMeCalled(txsInfo)
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
