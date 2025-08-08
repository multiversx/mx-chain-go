package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

type BlockProcessorStub struct {
	ProcessBlockCalled func(handler data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error)
}

// ProcessBlock -
func (bp *BlockProcessorStub) ProcessBlock(header data.HeaderHandler, body data.BodyHandler) (data.ExecutionResultHandler, error) {
	if bp.ProcessBlockCalled != nil {
		return bp.ProcessBlockCalled(header, body)
	}

	return nil, nil
}

// IsInterfaceNil -
func (bp *BlockProcessorStub) IsInterfaceNil() bool {
	return bp == nil
}
