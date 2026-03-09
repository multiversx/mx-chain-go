package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
)

// ReceiptsRepositoryStub -
type ReceiptsRepositoryStub struct {
	SaveReceiptsCalled              func(holder common.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error
	SaveReceiptsForExecResultCalled func(holder common.ReceiptsHolder, execResult data.BaseExecutionResultHandler) error
	LoadReceiptsCalled              func(header data.HeaderHandler, headerHash []byte) (common.ReceiptsHolder, error)
}

// SaveReceipts -
func (stub *ReceiptsRepositoryStub) SaveReceipts(holder common.ReceiptsHolder, header data.HeaderHandler, headerHash []byte) error {
	if stub.SaveReceiptsCalled != nil {
		return stub.SaveReceiptsCalled(holder, header, headerHash)
	}

	return nil
}

// SaveReceiptsForExecResult -
func (stub *ReceiptsRepositoryStub) SaveReceiptsForExecResult(holder common.ReceiptsHolder, execResult data.BaseExecutionResultHandler) error {
	if stub.SaveReceiptsForExecResultCalled != nil {
		return stub.SaveReceiptsForExecResultCalled(holder, execResult)
	}

	return nil
}

// LoadReceipts -
func (stub *ReceiptsRepositoryStub) LoadReceipts(header data.HeaderHandler, headerHash []byte) (common.ReceiptsHolder, error) {
	if stub.LoadReceiptsCalled != nil {
		return stub.LoadReceiptsCalled(header, headerHash)
	}

	return holders.NewReceiptsHolder(nil), nil
}

// IsInterfaceNil -
func (stub *ReceiptsRepositoryStub) IsInterfaceNil() bool {
	return stub == nil
}
