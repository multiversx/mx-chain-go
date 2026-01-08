package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

// BlockProcessorStub -
type BlockProcessorStub struct {
	ProcessBlockProposalCalled func(handler data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error)
}

// ProcessBlockProposal -
func (bp *BlockProcessorStub) ProcessBlockProposal(header data.HeaderHandler, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
	if bp.ProcessBlockProposalCalled != nil {
		return bp.ProcessBlockProposalCalled(header, body)
	}

	return nil, nil
}

// IsInterfaceNil -
func (bp *BlockProcessorStub) IsInterfaceNil() bool {
	return bp == nil
}
