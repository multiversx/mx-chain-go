package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// BlockProcessingCutoffStub -
type BlockProcessingCutoffStub struct {
	HandleProcessErrorCutoffCalled func(header data.HeaderHandler) error
	HandlePauseCutoffCalled        func(header data.HeaderHandler)
	HandleGracefulStopCutoffCalled func(header data.HeaderHandler, beforeStop func())
	CloseCalled                    func()
}

// HandleGracefulStopCutoff -
func (b *BlockProcessingCutoffStub) HandleGracefulStopCutoff(header data.HeaderHandler, beforeStop func()) {
	if b.HandleGracefulStopCutoffCalled != nil {
		b.HandleGracefulStopCutoffCalled(header, beforeStop)
	}
}

// Close -
func (b *BlockProcessingCutoffStub) Close() {
	if b.CloseCalled != nil {
		b.CloseCalled()
	}
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
