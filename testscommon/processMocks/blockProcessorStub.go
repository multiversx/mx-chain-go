package processMocks

import "github.com/multiversx/mx-chain-core-go/data"

// BlockProcessorStub -
type BlockProcessorStub struct {
	ProcessBlockProposalCalled     func(handler data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error)
	CommitBlockProposalStateCalled func(headerHandler data.HeaderHandler) error
	RevertBlockProposalStateCalled func()
	PruneTrieAsyncHeaderCalled     func()
}

// ProcessBlockProposal -
func (bp *BlockProcessorStub) ProcessBlockProposal(header data.HeaderHandler, headerHash []byte, body data.BodyHandler) (data.BaseExecutionResultHandler, error) {
	if bp.ProcessBlockProposalCalled != nil {
		return bp.ProcessBlockProposalCalled(header, headerHash, body)
	}

	return nil, nil
}

// CommitBlockProposalState -
func (bp *BlockProcessorStub) CommitBlockProposalState(headerHandler data.HeaderHandler) error {
	if bp.CommitBlockProposalStateCalled != nil {
		return bp.CommitBlockProposalStateCalled(headerHandler)
	}

	return nil
}

// RevertBlockProposalState -
func (bp *BlockProcessorStub) RevertBlockProposalState() {
	if bp.RevertBlockProposalStateCalled != nil {
		bp.RevertBlockProposalStateCalled()
	}
}

// PruneTrieAsyncHeader -
func (bp *BlockProcessorStub) PruneTrieAsyncHeader() {
	if bp.PruneTrieAsyncHeaderCalled != nil {
		bp.PruneTrieAsyncHeaderCalled()
	}
}

// IsInterfaceNil -
func (bp *BlockProcessorStub) IsInterfaceNil() bool {
	return bp == nil
}
