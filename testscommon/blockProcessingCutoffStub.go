package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// BlockProcessingCutoffStub -
type BlockProcessingCutoffStub struct {
	HandleProcessErrorCutoffCalled func(header data.HeaderHandler) error
	HandlePauseCutoffCalled        func(header data.HeaderHandler)
}

// HandleProcessErrorCutoff -
func (b *BlockProcessingCutoffStub) HandleProcessErrorCutoff(header data.HeaderHandler) error {
	if b.HandleProcessErrorCutoffCalled != nil {
		return b.HandleProcessErrorCutoffCalled(header)
	}

	return nil
}

// HandlePauseCutoff -
func (b *BlockProcessingCutoffStub) HandlePauseCutoff(header data.HeaderHandler) {
	if b.HandlePauseCutoffCalled != nil {
		b.HandlePauseCutoffCalled(header)
	}
}

// IsInterfaceNil -
func (b *BlockProcessingCutoffStub) IsInterfaceNil() bool {
	return b == nil
}
